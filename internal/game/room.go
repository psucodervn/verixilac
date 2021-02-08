package game

import (
	"bytes"
	"fmt"
	"sync"
)

type Room struct {
	id          string
	players     []*Player
	currentGame *Game

	mu sync.RWMutex
}

func NewRoom(id string, players ...*Player) *Room {
	r := &Room{
		id:      id,
		players: players,
	}
	if len(r.id) == 0 {
		r.id = generateRoomID()
	}
	return r
}

func (r *Room) ID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.id
}

func (r *Room) CurrentGame() *Game {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentGame
}

func (r *Room) SetCurrentGame(g *Game) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.currentGame = g
}

func (r *Room) JoinPlayer(p *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.findPlayerByID(p.ID()) != nil {
		return ErrPlayerAlreadyInRoom
	}
	r.players = append(r.players, p)
	return nil
}

func (r *Room) Players() []*Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.players
}

func (r *Room) findPlayerByID(id string) *Player {
	for i := range r.players {
		if r.players[i].id == id {
			return r.players[i]
		}
	}
	return nil
}

func (r *Room) removePlayerByID(id string) *Player {
	toRemove := -1
	for i := 0; i < len(r.players); i++ {
		if r.players[i].id == id {
			toRemove = i
			break
		}
	}
	if toRemove < 0 {
		return nil
	}
	p := r.players[toRemove]
	if toRemove < len(r.players)-1 {
		copy(r.players[toRemove:], r.players[toRemove+1:])
	}
	r.players = r.players[:len(r.players)-1]
	return p
}

func (r *Room) RemovePlayer(p *Player) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removePlayerByID(p.ID())
}

func (r *Room) Info() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bf := bytes.NewBuffer(nil)
	bf.WriteString("Phòng hiện tại: " + r.id)
	bf.WriteString("\nThành viên:\n")
	for _, p := range r.players {
		bf.WriteString(fmt.Sprintf(" - %s: %+dk\n", p.Name(), p.Balance()))
	}
	return bf.String()
}
