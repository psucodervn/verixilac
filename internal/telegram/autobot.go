package telegram

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
	"gopkg.in/tucnak/telebot.v2"

	"github.com/psucodervn/verixilac/internal/game"
	"github.com/psucodervn/verixilac/internal/model"
)

var (
	autoBotCount int
)

func init() {
	autoBotCount, _ = strconv.Atoi(os.Getenv("TEST_ACCOUNT"))
}

func fakePlay(h *Handler, g *game.Game, count int) {
	for i := 1; i <= count; i++ {
		m := &telebot.Message{ID: i, Payload: g.ID(), Chat: &telebot.Chat{ID: int64(i), Username: fmt.Sprint("Bot #", i)}}
		for {
			ok := h.doStand(m, true, true)
			if ok {
				break
			}
			if ok = h.doHit(m, true); !ok {
				break
			}
		}
	}
}

func fakeBet(ctx context.Context, h *Handler, g *game.Game, count int) {
	for i := 1; i <= count; i++ {
		botP1 := &model.Player{ID: fmt.Sprint(i), Name: fmt.Sprint("Bot #", i), Balance: 1000, UserRole: model.UserRoleBot}
		if p, _ := h.game.PlayerRegister(ctx, botP1.ID, botP1.Name, botP1.UserRole); p == nil {
			log.Error().Msg("cannot register bot")
		}
		if err := h.game.PlayerBet(ctx, g.ID(), botP1, uint64(20*i)); err != nil {
			log.Err(err).Msg("fakeBet")
		}
	}
}
