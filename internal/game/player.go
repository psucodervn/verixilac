package game

import (
	"sync"

	"go.uber.org/atomic"
)

type Player struct {
	id          string
	name        string
	balance     atomic.Int64
	isAdmin     atomic.Bool
	ruleID      atomic.String
	currentRoom *Room
	currentGame *Game

	mu sync.RWMutex
}

func NewPlayer(id string, name string, initialBalance int64) *Player {
	p := &Player{id: id, name: name}
	p.balance.Store(initialBalance)
	return p
}

func (p *Player) ID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id
}

func (p *Player) Name() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.name
}

func (p *Player) IsAdmin() bool {
	return p.isAdmin.Load()
}

func (p *Player) SetIsAdmin(isAdmin bool) {
	p.isAdmin.Store(isAdmin)
}

func (p *Player) CurrentRoom() *Room {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentRoom
}

func (p *Player) CurrentGame() *Game {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentGame
}

func (p *Player) SetCurrentRoom(r *Room) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.currentRoom = r
}

func (p *Player) SetCurrentGame(g *Game) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.currentGame = g
}

func (p *Player) AddBalance(amount int64) {
	p.balance.Add(amount)
}

func (p *Player) Balance() int64 {
	return p.balance.Load()
}

func (p *Player) Rule() *Rule {
	if r, ok := DefaultRules[p.ruleID.Load()]; ok {
		return &r
	}
	r := DefaultRules[DefaultRuleID]
	return &r
}

func (p *Player) SetRule(ruleID string) {
	p.ruleID.Store(ruleID)
}
