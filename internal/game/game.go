package game

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/rs/xid"
	"go.uber.org/atomic"
)

type Game struct {
	id         string
	room       *Room
	dealer     *PlayerInGame
	players    []*PlayerInGame
	table      []Card
	status     atomic.Uint32
	doneCnt    atomic.Uint32
	maxBet     atomic.Uint64
	currentIdx int

	onPlayerPlayFunc func(pg *PlayerInGame)

	mu sync.RWMutex
}

type Status uint32

const (
	Betting Status = iota
	Playing
	DealerPlaying
	Finished
)

type Result uint8

const (
	Win Result = iota
	Draw
	Lose
)

func (r Result) String() string {
	if r == Win {
		return "Win"
	} else if r == Draw {
		return "Draw"
	} else {
		return "Lose"
	}
}

func NewGame(dealer *Player, room *Room, maxBet uint64) *Game {
	return &Game{
		id:         xid.New().String(),
		room:       room,
		dealer:     NewPlayerInGame(dealer, 0, true),
		currentIdx: -1,
		maxBet:     *atomic.NewUint64(maxBet),
	}
}

func (g *Game) ID() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.id
}

func (g *Game) Dealer() *PlayerInGame {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.dealer
}

func (g *Game) Deal() error {
	if Status(g.status.Load()) != Betting {
		return ErrGameAlreadyStarted
	}
	g.mu.Lock()

	if len(g.players) == 0 {
		g.mu.Unlock()
		return ErrEmptyGame
	}

	g.table = make(Cards, 52)
	for i := 0; i < 52; i++ {
		g.table[i] = Card{id: i}
	}

	// TODO: use crypto/rand
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(52, func(i, j int) {
		g.table[i], g.table[j] = g.table[j], g.table[i]
	})

	// split cards
	g.dealer.AddCard(g.table[0])
	g.dealer.AddCard(g.table[len(g.players)+1])
	// g.dealer.AddCard(Card{id: 0})
	// g.dealer.AddCard(Card{id: 10})
	for i := 0; i < len(g.players); i++ {
		// if i == 2 {
		//   g.players[i].AddCard(Card{id: 13})
		//   g.players[i].AddCard(Card{id: 26})
		// } else {
		g.players[i].AddCard(g.table[i+1])
		g.players[i].AddCard(g.table[i+len(g.players)+2])
		// }
	}
	g.table = g.table[len(g.players)*2+2:]
	g.doneCnt.Store(uint32(len(g.players)))
	g.status.Store(uint32(Playing))
	g.mu.Unlock()
	return nil
}

func (g *Game) PlayerBet(p *Player, betAmount uint64) (*PlayerInGame, error) {
	if Status(g.status.Load()) != Betting {
		return nil, ErrGameAlreadyStarted
	}

	if betAmount > g.maxBet.Load() {
		return nil, fmt.Errorf("bạn chỉ được bet tối đa %dk", g.maxBet.Load())
	}

	pg := g.FindPlayer(p.ID())
	if pg == nil {
		pg = NewPlayerInGame(p, 0, false)
		g.mu.Lock()
		g.players = append(g.players, pg)
		g.mu.Unlock()
	}

	if pg.BetAmount()+betAmount > g.maxBet.Load() {
		return nil, fmt.Errorf("bạn chỉ được bet tối đa %dk", g.maxBet.Load())
	}
	pg.AddBet(betAmount)
	return pg, nil
}

func (g *Game) PreparingBoard() string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	bf := bytes.NewBuffer(nil)
	bf.WriteString(fmt.Sprintf("Nhà cái: %s\n", g.dealer.Name()))
	bf.WriteString(fmt.Sprintf("Người chơi (%d):", len(g.players)))
	if len(g.players) == 0 {
		bf.WriteString("\n(chưa có ai)")
	} else {
		for _, p := range g.players {
			bf.WriteString(fmt.Sprintf("\n  - %s: %dk", p.Name(), p.BetAmount()))
		}
	}
	return bf.String()
}

func (g *Game) CurrentBoard() string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	bf := bytes.NewBuffer(nil)
	bf.WriteString(fmt.Sprintf("Nhà cái: %s\n", g.dealer.CardsString()))
	bf.WriteString(fmt.Sprintf("Người chơi (%d):", len(g.players)))
	if len(g.players) == 0 {
		bf.WriteString("\n(chưa có ai)")
	} else {
		for _, p := range g.players {
			bf.WriteString(fmt.Sprintf("\n  - %s: %s", p.Name(), p.CardsString()))
		}
	}
	return bf.String()
}

func (g *Game) PlayerBoard() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	bf := bytes.NewBuffer(nil)
	for _, p := range g.players {
		bf.WriteString(fmt.Sprintf(" - %s: %s\n", p.Name(), p.CardsString()))
	}
	return bf.String()
}

func (g *Game) ResultBoard() string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	bf := bytes.NewBuffer(nil)
	bf.WriteString(fmt.Sprintf("Nhà cái: %s\n", g.dealer.Cards().String(false, true)))
	bf.WriteString(fmt.Sprintf("Người chơi (%d):", len(g.players)))
	for _, p := range g.players {
		bf.WriteString(fmt.Sprintf("\n  - %s: %s", p.Name(), p.Cards().String(false, false)))
	}

	bf.WriteString(fmt.Sprintf("\n\nTiền thưởng:\n\nNhà cái (%s): %+dk\n", g.dealer.Name(), g.dealer.Reward()))
	bf.WriteString(fmt.Sprintf("Người chơi:"))
	for _, p := range g.players {
		bf.WriteString(fmt.Sprintf("\n  - %s: %+dk", p.Name(), p.Reward()))
	}
	return bf.String()
}

func (g *Game) FindPlayer(id string) *PlayerInGame {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.findPlayer(id)
}

func (g *Game) RemovePlayer(id string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i := range g.players {
		if g.players[i].ID() != id {
			continue
		}
		g.players = append(g.players[:i], g.players[i+1:]...)
		return nil
	}
	return ErrPlayerNotFound
}

func (g *Game) Playing() bool {
	return Status(g.status.Load()) == Playing
}

func (g *Game) Finished() bool {
	return Status(g.status.Load()) == Finished
}

func (g *Game) Status() Status {
	return Status(g.status.Load())
}

func (g *Game) Room() *Room {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.room
}

func (g *Game) Players() []*PlayerInGame {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.players
}

func (g *Game) RemoveCard() (Card, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	// TODO: check if available
	c := g.table[0]
	g.table = g.table[1:]
	return c, nil
}

func (g *Game) PlayerStand(pg *PlayerInGame) (err error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.currentIdx < 0 {
		return ErrPlayerNotFound
	}
	if g.currentIdx >= len(g.players) {
		err = g.dealer.Stand()
	} else {
		err = g.players[g.currentIdx].Stand()
	}
	return err
}

func (g *Game) PlayerNext() (*PlayerInGame, error) {
	g.mu.RLock()
	var playPG *PlayerInGame
	for {
		g.currentIdx++
		if g.currentIdx < len(g.players) {
			if g.players[g.currentIdx].IsDone() {
				continue
			}
			if err := g.players[g.currentIdx].Play(); err != nil {
				g.mu.RUnlock()
				return nil, err
			}
			playPG = g.players[g.currentIdx]
			break
		} else {
			g.status.Store(uint32(DealerPlaying))
			if err := g.dealer.Play(); err != nil {
				g.mu.RUnlock()
				return nil, err
			}
			playPG = g.dealer
			break
		}
	}
	f := g.onPlayerPlayFunc
	g.mu.RUnlock()

	if f != nil && playPG != nil {
		f(playPG)
	}
	return playPG, nil
}

func (g *Game) Done(pg *PlayerInGame, force bool) (int64, error) {
	if pg.IsDone() {
		return pg.Reward(), nil
	}
	if !force {
		if st := pg.Status(); st != PlayerStood {
			if st < PlayerStood {
				return 0, ErrPlayerNotStandYet
			} else {
				return 0, ErrPlayerIsDone
			}
		}
	}

	reward := GetReward(g.dealer, pg)
	g.dealer.AddReward(reward)
	g.doneCnt.Dec()
	pg.Done(-reward)
	if g.doneCnt.Load() == 0 {
		g.status.Store(uint32(Finished))
	}
	return reward, nil
}

func (g *Game) OnPlayerPlay(f func(pg *PlayerInGame)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.onPlayerPlayFunc = f
}

func (g *Game) findPlayer(id string) *PlayerInGame {
	if g.dealer.ID() == id {
		return g.dealer
	}
	for i := range g.players {
		if g.players[i].ID() == id {
			return g.players[i]
		}
	}
	return nil
}

// AllPlayers includes dealer
func (g *Game) AllPlayers() []*PlayerInGame {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return append(g.players, g.dealer)
}

func Compare(a, b *PlayerInGame) Result {
	rta := a.Cards().Type(a.IsDealer())
	rtb := b.Cards().Type(b.IsDealer())
	if rta < rtb {
		return Win
	} else if rta > rtb {
		return Lose
	}
	if rta == TypeTooHigh || rta == TypeBusted || rta == TypeTooLow {
		return Draw
	}
	res := compareScore(a.Cards().Value(), b.Cards().Value())
	if rta == TypeHighFive {
		res = reverseResult(res)
	}
	return res
}

func GetReward(a, b *PlayerInGame) int64 {
	cp := Compare(a, b)
	if cp == Draw {
		return 0
	}
	rta := a.Cards().Type(a.IsDealer())
	rtb := b.Cards().Type(b.IsDealer())
	if rta == TypeDoubleBlackJack {
		return int64(b.BetAmount() * 2)
	} else if rtb == TypeDoubleBlackJack {
		return -int64(b.BetAmount() * 2)
	} else if cp == Win {
		return int64(b.BetAmount())
	} else {
		return -int64(b.BetAmount())
	}
}

func compareScore(a, b int) Result {
	if a < b {
		return Lose
	} else if a > b {
		return Win
	} else {
		return Draw
	}
}

func reverseResult(s Result) Result {
	return 2 - s
}
