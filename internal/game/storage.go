package game

import (
	"encoding/json"
	"io/ioutil"
)

const (
	storageFile = "data/data.json"
)

type (
	stRoom struct {
		ID string `json:"id"`
	}
	stPlayer struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Balance int64  `json:"balance"`
		RoomID  string `json:"roomId,omitempty"`
		IsAdmin bool   `json:"isAdmin"`
	}
	stData struct {
		Rooms   map[string]stRoom   `json:"rooms"`
		Players map[string]stPlayer `json:"players"`
	}
)

func (m *Manager) LoadFromStorage() error {
	b, err := ioutil.ReadFile(storageFile)
	if err != nil {
		return err
	}
	var data stData
	if err = json.Unmarshal(b, &data); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, stR := range data.Rooms {
		r := NewRoom(stR.ID)
		m.rooms.Store(r.ID(), r)
	}
	for _, stP := range data.Players {
		p := NewPlayer(stP.ID, stP.Name)
		p.SetIsAdmin(stP.IsAdmin)
		p.AddBalance(stP.Balance)
		m.players.Store(p.ID(), p)
		if len(stP.RoomID) > 0 {
			rr, ok := m.rooms.Load(stP.RoomID)
			if !ok || rr == nil {
				continue
			}
			r := rr.(*Room)
			if r.JoinPlayer(p) != nil {
				continue
			}
			p.SetCurrentRoom(r)
		}
	}

	return nil
}

func (m *Manager) SaveToStorage() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data := stData{
		Rooms:   make(map[string]stRoom),
		Players: make(map[string]stPlayer),
	}
	m.rooms.Range(func(id, rr interface{}) bool {
		r := rr.(*Room)
		data.Rooms[r.ID()] = stRoom{
			ID: r.ID(),
		}
		return true
	})
	m.players.Range(func(id, pp interface{}) bool {
		p := pp.(*Player)
		stP := stPlayer{
			ID:      p.ID(),
			Name:    p.Name(),
			Balance: p.Balance(),
			IsAdmin: p.IsAdmin(),
		}
		if p.CurrentRoom() != nil {
			stP.RoomID = p.CurrentRoom().ID()
		}
		data.Players[p.ID()] = stP
		return true
	})
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(storageFile, b, 0666)
}
