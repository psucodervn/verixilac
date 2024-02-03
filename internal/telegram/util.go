package telegram

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"
	"gopkg.in/tucnak/telebot.v2"

	"github.com/psucodervn/verixilac/internal/game"
	"github.com/psucodervn/verixilac/internal/model"
)

func ToTelebotChats(ids ...string) []*telebot.Chat {
	cs := make([]*telebot.Chat, len(ids))
	for i, id := range ids {
		cs[i] = &telebot.Chat{ID: cast.ToInt64(id)}
	}
	return cs
}

func ToTelebotChat(id string) *telebot.Chat {
	return &telebot.Chat{ID: cast.ToInt64(id)}
}

func GetUsername(chat *telebot.Chat) string {
	name := strings.TrimSpace(chat.FirstName + " " + chat.LastName)
	if len(name) == 0 {
		name = strings.TrimSpace(chat.Username)
	}
	if len(name) == 0 {
		name = fmt.Sprintf("%v", chat.ID)
	}
	return name
}

func FilterPlayers(players []model.Player, ids ...string) []model.Player {
	m := make(map[string]struct{})
	for _, id := range ids {
		m[id] = struct{}{}
	}
	var ps []model.Player
	for i := 0; i < len(players); i++ {
		if _, exists := m[players[i].ID]; !exists {
			ps = append(ps, players[i])
		}
	}
	return ps
}

func FilterInGamePlayers(players []*game.PlayerInGame, ids ...string) []*game.PlayerInGame {
	m := make(map[string]struct{})
	for _, id := range ids {
		m[id] = struct{}{}
	}
	var ps []*game.PlayerInGame
	for i := 0; i < len(players); i++ {
		if _, exists := m[players[i].ID]; !exists {
			ps = append(ps, players[i])
		}
	}
	return ps
}
