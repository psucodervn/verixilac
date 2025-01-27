package game

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/rs/xid"
	"go.uber.org/atomic"

	"github.com/psucodervn/verixilac/internal/model"
	"github.com/psucodervn/verixilac/internal/stringer"
)

type Game struct {
	id         string
	dealer     *PlayerInGame
	rule       *Rule
	players    []*PlayerInGame
	table      []Card
	status     atomic.Uint32
	doneCnt    atomic.Uint32
	maxBet     atomic.Uint64
	timeout    atomic.Duration
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

func NewGame(dealer *model.Player, rule *Rule, maxBet uint64, timeout time.Duration) *Game {
	return &Game{
		id:         xid.New().String(),
		dealer:     NewPlayerInGame(dealer, 0, true),
		rule:       rule,
		currentIdx: -1,
		maxBet:     *atomic.NewUint64(maxBet),
		timeout:    *atomic.NewDuration(timeout),
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
	g.mu.Lock()
	if Status(g.status.Load()) != Betting {
		g.mu.Unlock()
		return ErrGameAlreadyStarted
	}

	if len(g.players) == 0 {
		g.mu.Unlock()
		return ErrEmptyGame
	}

	g.table = make(Cards, 52)
	for i := 0; i < 52; i++ {
		g.table[i] = Card{id: i}
	}

	// shuffle cards
	for i := 52 - 1; i > 0; i-- {
		bj, _ := rand.Int(rand.Reader, big.NewInt(int64(i)))
		j := bj.Int64()
		g.table[i], g.table[j] = g.table[j], g.table[i]
	}

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

func (g *Game) PlayerBet(p *model.Player, betAmount uint64) (*PlayerInGame, error) {
	if Status(g.status.Load()) != Betting {
		return nil, ErrGameAlreadyStarted
	}

	if p.Balance < int64(betAmount) {
		return nil, fmt.Errorf("bạn không đủ số dư để bet %s", stringer.FormatCurrency(betAmount))
	}
	if betAmount > g.maxBet.Load() {
		return nil, fmt.Errorf("bạn chỉ được bet tối đa %s", stringer.FormatCurrency(g.maxBet.Load()))
	}

	g.mu.Lock()
	pg := g.findPlayer(p.ID)
	if pg == nil {
		pg = NewPlayerInGame(p, int64(betAmount), false)
		g.players = append(g.players, pg)
	} else {
		pg.SetBet(betAmount)
	}
	g.mu.Unlock()

	return pg, nil
}

func (g *Game) PreparingBoard() string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	bf := bytes.NewBuffer(nil)
	bf.WriteString(fmt.Sprintf("Nhà cái: `%s`\n", g.dealer.Name))
	bf.WriteString(fmt.Sprintf("Người chơi (%d - %s):", len(g.players), stringer.FormatCurrency(g.totalBetAmount())))
	if len(g.players) == 0 {
		bf.WriteString("\n(chưa có ai)")
	} else {
		for _, p := range g.players {
			bf.WriteString(fmt.Sprintf("\n  - `%s`: %s", p.Name, stringer.FormatCurrency(p.BetAmount())))
		}
	}
	return bf.String()
}

func (g *Game) CurrentBoard() string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	bf := bytes.NewBuffer(nil)
	bf.WriteString(fmt.Sprintf("Nhà cái : %s\n", g.dealer.CardsString()))
	bf.WriteString(fmt.Sprintf("Người chơi (%d - %s):", len(g.players), stringer.FormatCurrency(g.totalBetAmount())))
	if len(g.players) == 0 {
		bf.WriteString("\n(chưa có ai)")
	} else {
		for _, p := range g.players {
			bf.WriteString(fmt.Sprintf("\n  - %s: %s", p.Name, p.CardsString()))
		}
	}
	return bf.String()
}

func (g *Game) PlayerBoard() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	bf := bytes.NewBuffer(nil)
	for _, p := range g.players {
		bf.WriteString(fmt.Sprintf(" - %s: %s\n", p.Name, p.CardsString()))
	}
	return bf.String()
}

func (g *Game) ResultBoard() string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	bf := bytes.NewBuffer(nil)
	bf.WriteString(fmt.Sprintf("Nhà cái: %s\n", g.dealer.Cards().String(false, true)))
	bf.WriteString(fmt.Sprintf("Người chơi (%d - %s):", len(g.players), stringer.FormatCurrency(g.totalBetAmount())))
	for _, p := range g.players {
		bf.WriteString(fmt.Sprintf("\n  - `%s`: %s", p.Name, p.Cards().String(false, false)))
	}

	bf.WriteString(fmt.Sprintf("\n\nThưởng:\n\nNhà cái (`%s`): %s (%s)\n",
		g.dealer.Name,
		stringer.FormatCurrency(g.dealer.Reward()),
		stringer.FormatCurrency(g.dealer.Balance+g.dealer.Reward())))

	bf.WriteString(fmt.Sprintf("Người chơi: "))
	for _, p := range g.players {
		bf.WriteString(fmt.Sprintf("\n  - `%s`: %s (%s)", p.Name, stringer.FormatCurrency(p.Reward()), stringer.FormatCurrency(p.Balance+p.Reward())))
	}
	return bf.String()
}

type ResultMapItem struct {
	PlayerID   string
	Reward     int64
	ResultType model.ResultType
	Value      int
	IsDealer   bool
}

func (g *Game) ResultMap() []ResultMapItem {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]ResultMapItem, len(g.players)+1)
	for i, p := range g.players {
		result[i] = ResultMapItem{
			PlayerID:   p.ID,
			Reward:     p.Reward(),
			ResultType: p.ResultType(),
			Value:      p.Cards().Value(),
			IsDealer:   false,
		}
	}
	result[len(g.players)] = ResultMapItem{
		PlayerID:   g.dealer.ID,
		Reward:     g.dealer.Reward(),
		ResultType: g.dealer.ResultType(),
		Value:      g.dealer.Cards().Value(),
		IsDealer:   true,
	}
	return result
}

func (g *Game) FindPlayer(id string) *PlayerInGame {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.findPlayer(id)
}

func (g *Game) RemovePlayer(id string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if Status(g.status.Load()) != Betting {
		return ErrGameAlreadyStarted
	}
	for i := range g.players {
		if g.players[i].ID != id {
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

func (g *Game) PlayersInGame() []*PlayerInGame {
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
		if pg.ID != g.dealer.ID {
			return ErrYouNotPlaying
		}
		err = g.dealer.Stand()
	} else {
		if pg.ID != g.players[g.currentIdx].ID {
			return ErrYouNotPlaying
		}
		err = g.players[g.currentIdx].Stand()
	}
	return err
}

func (g *Game) PlayerNext() (*PlayerInGame, error) {
	g.mu.Lock()
	var playPG *PlayerInGame
	for {
		g.currentIdx++
		if g.currentIdx < len(g.players) {
			if g.players[g.currentIdx].IsDone() {
				continue
			}
			if err := g.players[g.currentIdx].Play(); err != nil {
				g.mu.Unlock()
				return nil, err
			}
			playPG = g.players[g.currentIdx]
			break
		} else {
			g.status.Store(uint32(DealerPlaying))
			if err := g.dealer.Play(); err != nil {
				g.mu.Unlock()
				return nil, err
			}
			playPG = g.dealer
			break
		}
	}
	f := g.onPlayerPlayFunc
	g.mu.Unlock()

	if f != nil && playPG != nil {
		f(playPG)
	}
	return playPG, nil
}

func (g *Game) Done(pg *PlayerInGame, force bool) (int64, error) {
	if pg.IsDone() {
		if pg.IsDealer() {
			return pg.Reward(), nil
		} else {
			return -pg.Reward(), nil
		}
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

	reward := GetReward(g.rule, g.dealer, pg)
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
	if g.dealer.ID == id {
		return g.dealer
	}
	for i := range g.players {
		if g.players[i].ID == id {
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

func (g *Game) CurrentPlaying() *PlayerInGame {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.currentIdx < 0 {
		return nil
	} else if g.currentIdx < len(g.players) {
		return g.players[g.currentIdx]
	} else {
		return g.dealer
	}
}

func (g *Game) Pass(pg *PlayerInGame) error {
	passed := time.Duration(time.Now().Unix()-pg.LastHit()) * time.Second
	need := g.timeout.Load()
	if pg.IsDealer() {
		need = need * 5
	}
	if passed < need {
		return ErrNotTimeout
	}
	pg.SetStatus(PlayerStood)
	if pg.IsDealer() {
		g.status.Store(uint32(Finished))
		return nil
	}
	_, err := g.PlayerNext()
	return err
}

func (g *Game) Rule() *Rule {
	return g.rule
}

func (g *Game) Players() []model.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	res := make([]model.Player, len(g.players)+1)
	res[0] = *g.dealer.Player
	for i, p := range g.players {
		res[i+1] = *p.Player
	}
	return res
}

func (g *Game) totalBetAmount() uint64 {
	res := uint64(0)
	for _, p := range g.players {
		res += p.BetAmount()
	}
	return res
}

func Compare(a, b *PlayerInGame) Result {
	rta := a.Cards().Type(a.IsDealer())
	rtb := b.Cards().Type(b.IsDealer())
	if rta < rtb {
		return Win
	} else if rta > rtb {
		return Lose
	}
	if rta == model.TypeTooHigh || rta == model.TypeBusted || rta == model.TypeTooLow {
		return Draw
	}
	res := compareScore(a.Cards().Value(), b.Cards().Value())
	if rta == model.TypeHighFive {
		res = reverseResult(res)
	}
	return res
}

func GetReward(rule *Rule, dealer, participant *PlayerInGame) int64 {
	cp := Compare(dealer, participant)
	if cp == Draw {
		return 0
	}
	rtDealer := dealer.Cards().Type(true)
	rtb := participant.Cards().Type(false)

	bm := int64(participant.BetAmount())
	var coff int64
	if cp == Win {
		// dealer win
		if v, ok := rule.Multipliers[Dealer][rtDealer]; ok {
			coff = v
		} else {
			coff = 1
		}
	} else {
		// participant win
		if v, ok := rule.Multipliers[Participant][rtb]; ok {
			coff = -v
		} else {
			coff = -1
		}
	}
	return bm * coff
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
