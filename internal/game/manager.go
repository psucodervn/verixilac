package game

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.uber.org/atomic"

	"github.com/psucodervn/verixilac/internal/model"
)

var initialBalance int

func init() {
	initialBalance, _ = strconv.Atoi(os.Getenv("INITIAL_BALANCE"))
}

type Manager struct {
	maxBet  atomic.Uint64
	minDeal atomic.Uint64
	timeout atomic.Duration

	players       sync.Map
	canCreateGame atomic.Bool
	currentGame   *Game

	store Storage

	mu                sync.RWMutex
	onNewGameFunc     OnNewGameFunc
	onPlayerJoinFunc  OnPlayerJoinFunc
	onPlayerLeaveFunc OnPlayerLeaveFunc
	onPlayerBetFunc   OnPlayerBetFunc
	onPlayerStandFunc OnPlayerStandFunc
	onPlayerHitFunc   OnPlayerHitFunc
	onGameFinishFunc  OnGameFinishFunc
	onPlayerPlayFunc  OnPlayerPlayFunc
}

type OnNewGameFunc func(g *Game)
type OnPlayerJoinFunc func(p *model.Player)
type OnPlayerLeaveFunc func(p *model.Player)
type OnPlayerBetFunc func(g *Game, p *PlayerInGame)
type OnPlayerStandFunc func(g *Game, p *PlayerInGame)
type OnPlayerHitFunc func(g *Game, p *PlayerInGame)
type OnGameFinishFunc func(g *Game)
type OnPlayerPlayFunc func(g *Game, pg *PlayerInGame)

func NewManager(store Storage, maxBet uint64, minDeal uint64, timeout time.Duration) *Manager {
	m := &Manager{
		maxBet:        *atomic.NewUint64(maxBet),
		minDeal:       *atomic.NewUint64(minDeal),
		timeout:       *atomic.NewDuration(timeout),
		canCreateGame: *atomic.NewBool(true),
		store:         store,
	}
	return m
}

func (m *Manager) PlayerRegister(ctx context.Context, id string, name string, role model.UserRole) (p *model.Player, joined bool) {
	p, err := m.store.GetPlayerByID(ctx, id)
	if err != nil {
		if !model.IsNotFound(err) {
			log.Ctx(ctx).Err(err).Str("id", id).Msg("get player failed")
			return nil, false
		}

		// create new player
		p = &model.Player{
			ID:         id,
			TelegramID: id,
			Name:       name,
			UserRole:   role,
			Balance:    int64(initialBalance),
		}
		if err := m.store.SavePlayer(ctx, p); err != nil {
			log.Ctx(ctx).Err(err).Str("id", id).Msg("save player failed")
			return nil, false
		}
		return p, false
	}

	if !p.IsActive() {
		if _, err := m.store.UpdatePlayerStatus(ctx, p.ID, model.UserStatusActive); err != nil {
			log.Ctx(ctx).Err(err).Str("id", id).Msg("update player status failed")
			return nil, false
		}
		return p, false
	}

	return p, true
}

func (m *Manager) findPlayer(ctx context.Context, id string) *model.Player {
	p, err := m.store.GetPlayerByID(ctx, id)
	if err == nil {
		return p
	} else if model.IsNotFound(err) {
		log.Ctx(ctx).Debug().Str("id", id).Msg("player not found")
	} else {
		log.Ctx(ctx).Err(err).Str("id", id).Msg("get player failed")
	}
	return nil
}

func (m *Manager) ActivePlayers(ctx context.Context) []model.Player {
	ps, err := m.store.ListActivePlayers(ctx)
	if err != nil {
		log.Err(err).Msg("list players failed")
		return nil
	}
	return ps
}

func (m *Manager) AllPlayers(ctx context.Context) []model.Player {
	ps, err := m.store.ListPlayers(ctx)
	if err != nil {
		log.Err(err).Msg("list players failed")
		return nil
	}
	return ps
}

func (m *Manager) Join(ctx context.Context, p *model.Player) error {
	m.mu.RLock()
	f := m.onPlayerJoinFunc
	m.mu.RUnlock()
	if f != nil {
		f(p)
	}
	log.Ctx(ctx).Debug().Str("player_id", p.ID).Msg("player joined room")
	return nil
}

func (m *Manager) Leave(ctx context.Context, p *model.Player) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentGame != nil {
		return ErrYouAlreadyInGame
	}

	// TODO: leave
	log.Ctx(ctx).Debug().Str("player_id", p.ID).Msg("player left room")
	return nil
}

func (m *Manager) OnNewGame(f OnNewGameFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onNewGameFunc = f
}

func (m *Manager) OnPlayerJoin(f OnPlayerJoinFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onPlayerJoinFunc = f
}

func (m *Manager) OnPlayerLeave(f OnPlayerLeaveFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onPlayerLeaveFunc = f
}

func (m *Manager) OnPlayerBet(f OnPlayerBetFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onPlayerBetFunc = f
}

func (m *Manager) OnPlayerStand(f OnPlayerStandFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onPlayerStandFunc = f
}

func (m *Manager) OnGameFinish(f OnGameFinishFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onGameFinishFunc = f
}

func (m *Manager) OnPlayerHit(f OnPlayerHitFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onPlayerHitFunc = f
}

func (m *Manager) OnPlayerPlay(f OnPlayerPlayFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onPlayerPlayFunc = f
}

func (m *Manager) NewGame(dealer *model.Player) (*Game, error) {
	if !m.canCreateGame.Load() {
		return nil, ErrServerMaintenance
	}
	if dealer.Balance < int64(m.minDeal.Load()) {
		return nil, fmt.Errorf("kiếm thêm tiền đi bạn ơi, tối thiểu %dk", m.minDeal.Load())
	}

	m.mu.Lock()
	if m.currentGame != nil {
		m.mu.Unlock()
		return nil, ErrGameIsExisted
	}

	g := NewGame(dealer, &DefaultRule, m.maxBet.Load(), m.timeout.Load())
	m.currentGame = g
	f := m.onNewGameFunc
	m.mu.Unlock()

	if f != nil {
		f(g)
	}
	return g, nil
}

func (m *Manager) PlayerBet(ctx context.Context, gameID string, p *model.Player, amount uint64) (err error) {
	m.mu.RLock()
	g := m.currentGame
	f := m.onPlayerBetFunc
	m.mu.RUnlock()

	if g == nil || g.ID() != gameID {
		return ErrGameNotFound
	}

	var pg *PlayerInGame
	if amount == 0 {
		if err = g.RemovePlayer(p.ID); err != nil {
			return err
		}
	} else {
		pg, err = g.PlayerBet(p, amount)
		if err != nil {
			return err
		}
	}

	if f != nil {
		f(g, pg)
	}

	return nil
}

func (m *Manager) PlayerStand(ctx context.Context, g *Game, pg *PlayerInGame) error {
	if pg.IsDone() {
		return nil
	}
	if !pg.CanStand() {
		return ErrYouCannotStand
	}
	if err := g.PlayerStand(pg); err != nil {
		log.Ctx(ctx).Err(err).Str("cards", pg.Cards().String(false)).Msg("player stand failed")
		return err
	}

	m.mu.RLock()
	f := m.onPlayerStandFunc
	m.mu.RUnlock()

	if f != nil {
		f(g, pg)
	}

	if _, err := g.PlayerNext(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) PlayerHit(ctx context.Context, g *Game, pg *PlayerInGame) error {
	if !pg.CanHit() {
		return ErrYouCannotHit
	}

	c, err := g.RemoveCard()
	if err != nil {
		return err
	}
	pg.AddCard(c)
	pg.SetLastHit(time.Now().Unix())

	m.mu.RLock()
	f := m.onPlayerHitFunc
	m.mu.RUnlock()

	if f != nil {
		f(g, pg)
	}
	return nil
}

func (m *Manager) CheckIfFinish(ctx context.Context, g *Game) bool {
	if !g.Finished() {
		return false
	}
	err := m.FinishGame(ctx, g, false)
	return err == nil
}

func (m *Manager) Deal(ctx context.Context, gameID string) (*Game, error) {
	m.mu.RLock()
	g := m.currentGame
	m.mu.RUnlock()

	if g == nil || g.ID() != gameID {
		return nil, ErrGameNotFound
	}

	g.OnPlayerPlay(func(pg *PlayerInGame) {
		m.mu.RLock()
		f := m.onPlayerPlayFunc
		m.mu.RUnlock()
		if f != nil {
			f(g, pg)
		}
	})
	if err := g.Deal(); err != nil {
		return nil, err
	}
	return g, nil
}

func (m *Manager) Start(ctx context.Context, g *Game) error {
	// check for early win
	gt := g.Dealer().ResultType()
	if gt == TypeDoubleBlackJack || gt == TypeBlackJack {
		return m.FinishGame(ctx, g, true)
	}

	cnt := 0
	for _, p := range g.PlayersInGame() {
		pt := p.ResultType()
		if pt == TypeDoubleBlackJack || pt == TypeBlackJack {
			_, _ = g.Done(p, true)
			cnt++
		}
	}

	if cnt == len(g.PlayersInGame()) {
		return m.FinishGame(ctx, g, true)
	}

	if _, err := g.PlayerNext(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) FinishGame(ctx context.Context, g *Game, force bool) error {
	if err := m.store.SaveRecord(ctx, &model.Record{
		GameID: g.ID(),
		Data:   g.ResultBoard(),
	}); err != nil {
		return err
	}

	for _, pg := range g.PlayersInGame() {
		if _, err := g.Done(pg, force); err != nil {
			return err
		}
	}

	m.mu.Lock()
	f := m.onGameFinishFunc
	m.currentGame = nil
	m.mu.Unlock()

	if f != nil {
		f(g)
	}
	return nil
}

func (m *Manager) CancelGame(ctx context.Context) error {
	m.mu.Lock()
	m.currentGame = nil
	m.mu.Unlock()

	return nil
}

func (m *Manager) SetMaxBet(maxBet uint64) uint64 {
	m.maxBet.Store(maxBet)
	return maxBet
}

func (m *Manager) PlayerPass(ctx context.Context) (*PlayerInGame, error) {
	m.mu.RLock()
	g := m.currentGame
	m.mu.RUnlock()

	if g == nil {
		return nil, ErrGameNotFound
	}

	pg := g.CurrentPlaying()
	if pg == nil {
		return nil, ErrPlayerNotFound
	}
	if err := g.Pass(pg); err != nil {
		return nil, err
	}
	m.CheckIfFinish(ctx, g)
	return pg, nil
}

func (m *Manager) Pause(ctx context.Context) error {
	m.canCreateGame.Store(false)
	return nil
}

func (m *Manager) Resume(ctx context.Context) error {
	m.canCreateGame.Store(true)
	return nil
}

func (m *Manager) Deposit(ctx context.Context, id string, amount int64) (*model.Player, error) {
	p := m.findPlayer(ctx, id)
	if p == nil {
		return nil, ErrPlayerNotFound
	}

	return m.store.AddPlayerBalance(ctx, p.ID, amount)
}

func (m *Manager) PlayerHistory(ctx context.Context, p *model.Player) string {
	records, err := m.store.ListRecords(ctx, p.ID)
	if err != nil {
		return err.Error()
	}

	bf := bytes.NewBuffer(nil)
	for _, r := range records {
		bf.WriteString(fmt.Sprintf("%s\n", r.GameID))
	}
	return bf.String()
}

func (m *Manager) ListPlayers(ctx context.Context) string {
	players, err := m.store.ListPlayers(ctx)
	if err != nil {
		return err.Error()
	}
	res, _ := json.Marshal(players)
	return string(res)
}

func (m *Manager) FindPlayerInGame(gameID string, playerID string) (*Game, *PlayerInGame) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	g := m.currentGame
	if g == nil || g.ID() != gameID {
		return nil, nil
	}
	return g, g.FindPlayer(playerID)
}

func (m *Manager) CurrentGame() *Game {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentGame
}

func (m *Manager) PlayerLeave(ctx context.Context, id string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, _ := m.store.GetPlayerByID(ctx, id); p == nil {
		return ErrPlayerNotFound
	} else if !p.IsActive() {
		return nil
	}

	if m.currentGame != nil {
		if _, pg := m.FindPlayerInGame(m.currentGame.ID(), id); pg != nil {
			return ErrYouAlreadyInGame
		}
	}

	p, err := m.store.UpdatePlayerStatus(ctx, id, model.UserStatusInactive)
	if err != nil {
		return err
	}

	if m.onPlayerLeaveFunc != nil {
		m.onPlayerLeaveFunc(p)
	}

	return nil
}
