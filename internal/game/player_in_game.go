package game

import (
	"sync"

	"go.uber.org/atomic"
)

type PlayerInGame struct {
	*Player
	cards     Cards
	betAmount atomic.Uint64
	isDealer  atomic.Bool
	status    atomic.Uint32
	reward    atomic.Int64
	result    atomic.Uint32

	mu sync.RWMutex
}

type PlayerInGameStatus uint32

const (
	PlayerWaiting PlayerInGameStatus = iota
	PlayerPlaying
	PlayerStood
	PlayerDone
)

func NewPlayerInGame(player *Player, betAmount int64, isDealer bool) *PlayerInGame {
	return &PlayerInGame{
		Player:    player,
		betAmount: *atomic.NewUint64(uint64(betAmount)),
		isDealer:  *atomic.NewBool(isDealer),
	}
}

func (p *PlayerInGame) IsDealer() bool {
	return p.isDealer.Load()
}

func (p *PlayerInGame) IsDone() bool {
	return PlayerInGameStatus(p.status.Load()) >= PlayerDone
}

func (p *PlayerInGame) BetAmount() uint64 {
	return p.betAmount.Load()
}

func (p *PlayerInGame) SetBet(betAmount uint64) {
	p.betAmount.Store(betAmount)
}

func (p *PlayerInGame) AddBet(betAmount uint64) {
	p.betAmount.Add(betAmount)
}

func (p *PlayerInGame) Cards() Cards {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cards
}

func (p *PlayerInGame) CardsString() string {
	var censor bool
	if p.isDealer.Load() {
		censor = PlayerInGameStatus(p.status.Load()) < PlayerPlaying
	} else {
		censor = PlayerInGameStatus(p.status.Load()) != PlayerDone
	}
	return p.cards.String(censor, p.isDealer.Load())
}

func (p *PlayerInGame) AddCard(card Card) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cards = append(p.cards, card)
}

func (p *PlayerInGame) Play() error {
	if PlayerInGameStatus(p.status.Load()) != PlayerWaiting {
		return ErrYouArePlayed
	}
	p.status.Store(uint32(PlayerPlaying))
	return nil
}

func (p *PlayerInGame) Stand() error {
	if PlayerInGameStatus(p.status.Load()) != PlayerPlaying {
		return ErrYouNotPlaying
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.cards.Type(p.isDealer.Load()) == TypeTooLow {
		return ErrTooLow
	}
	p.status.Store(uint32(PlayerStood))
	return nil
}

func (p *PlayerInGame) Status() PlayerInGameStatus {
	return PlayerInGameStatus(p.status.Load())
}

func (p *PlayerInGame) Done(reward int64) {
	p.reward.Store(reward)
	p.status.Store(uint32(PlayerDone))
}

func (p *PlayerInGame) AddReward(reward int64) int64 {
	return p.reward.Add(reward)
}

func (p *PlayerInGame) Reward() int64 {
	return p.reward.Load()
}

func (p *PlayerInGame) CanHit() bool {
	t := p.Cards().Type(p.isDealer.Load())
	return t == TypeTooLow || (t == TypeNormal && p.Cards().Value() < 21)
}

func (p *PlayerInGame) CanStand() bool {
	t := p.Cards().Type(p.isDealer.Load())
	return t != TypeTooLow
}

func ToPlayers(playersInGame []*PlayerInGame) []*Player {
	ps := make([]*Player, len(playersInGame))
	for i := range playersInGame {
		ps[i] = playersInGame[i].Player
	}
	return ps
}
