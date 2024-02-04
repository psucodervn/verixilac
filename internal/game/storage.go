package game

import (
	"context"

	"github.com/psucodervn/verixilac/internal/model"
)

const (
	storageFile = "data/data.json"
)

type Storage interface {
	SaveRecord(ctx context.Context, r *model.Record) error
	ListRecords(ctx context.Context, playerID string) ([]model.Record, error)
	GetPlayerByID(ctx context.Context, id string) (*model.Player, error)
	SavePlayer(ctx context.Context, p *model.Player) error
	ListPlayers(ctx context.Context) ([]model.Player, error)
	ListActivePlayers(ctx context.Context) ([]model.Player, error)
	AddPlayerBalance(ctx context.Context, id string, amount int64) (*model.Player, error)
	UpdatePlayerStatus(ctx context.Context, id string, status model.UserStatus) (*model.Player, error)
}

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
		RuleID  string `json:"ruleId,omitempty"`
	}
	stData struct {
		Rooms   map[string]stRoom   `json:"rooms"`
		Players map[string]stPlayer `json:"players"`
	}
)

func (m *Manager) LoadFromStorage() error {
	// b, err := ioutil.ReadFile(storageFile)
	// if err != nil {
	// 	return err
	// }
	// var data stData
	// if err = json.Unmarshal(b, &data); err != nil {
	// 	return err
	// }
	//
	// m.mu.Lock()
	// defer m.mu.Unlock()

	// for _, stP := range data.PlayersInGame {
	// 	p := NewPlayer(stP.ID, stP.Name, stP.Balance)
	// 	p.SetIsAdmin(stP.IsAdmin)
	// 	p.SetRule(stP.RuleID)
	// 	m.players.Store(p.ID, p)
	// 	if len(stP.RoomID) > 0 {
	// 		rr, ok := m.rooms.Load(stP.RoomID)
	// 		if !ok || rr == nil {
	// 			continue
	// 		}
	// 		r := rr.(*Room)
	// 		if r.JoinPlayer(p) != nil {
	// 			continue
	// 		}
	// 		p.SetCurrentRoom(r)
	// 	}
	// }

	return nil
}

func (m *Manager) SaveToStorage() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// data := stData{
	// 	Rooms:   make(map[string]stRoom),
	// 	PlayersInGame: make(map[string]stPlayer),
	// }

	// m.players.Range(func(id, pp interface{}) bool {
	// 	p := pp.(*Player)
	// 	stP := stPlayer{
	// 		ID:      p.ID,
	// 		Name:    p.Name,
	// 		Balance: p.Balance(),
	// 		IsAdmin: p.IsAdmin(),
	// 		RuleID:  p.Rule().ID,
	// 	}
	// 	if p.CurrentRoom() != nil {
	// 		stP.RoomID = p.CurrentRoom().ID
	// 	}
	// 	data.PlayersInGame[p.ID] = stP
	// 	return true
	// })
	// b, err := json.MarshalIndent(data, "", "  ")
	// if err != nil {
	// 	return err
	// }
	// return ioutil.WriteFile(storageFile, b, 0666)

	return nil
}
