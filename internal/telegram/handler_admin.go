package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/telebot.v3"

	"github.com/psucodervn/verixilac/internal/model"
	"github.com/psucodervn/verixilac/internal/stringer"
)

func (h *Handler) CmdAdmin(ctx telebot.Context) error {
	m := ctx.Message()
	p := h.getPlayer(m)
	if !p.IsAdmin() {
		h.sendMessage(m.Chat, "Bạn không có quyền admin")
		return nil
	}
	ss := strings.Split(strings.TrimSpace(m.Payload), " ")
	if len(ss) == 0 {
		return nil
	}

	cmd := ss[0]
	switch cmd {
	case "pause":
		h.doAdminPause(m)
	case "resume":
		h.doAdminResume(m)
	case "deposit":
		h.doDeposit(m, p, ss[1:])
	}
	return nil
}

func (h *Handler) doAdminPause(m *telebot.Message) {
	if err := h.game.Pause(h.ctx(m)); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	h.broadcast(h.game.AllPlayers(h.ctx(m)), "‼️Server is Under Maintenance. Please wait!", false)
}

func (h *Handler) doAdminResume(m *telebot.Message) {
	if err := h.game.Resume(h.ctx(m)); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	h.broadcast(h.game.AllPlayers(nil), "✅ Server is Live now. Enjoy!", false)
}

func (h *Handler) doDeposit(m *telebot.Message, operator *model.Player, ss []string) {
	if len(ss) != 2 {
		h.sendMessage(m.Chat, "Cú pháp: /deposit player_id amount")
		return
	}

	id := ss[0]
	amount, err := strconv.ParseInt(ss[1], 10, 64)
	if err != nil {
		h.sendMessage(m.Chat, "Cú pháp: /deposit player_id amount")
		return
	}

	p, err := h.game.Deposit(h.ctx(m), id, amount)
	if err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}

	log.Info().Str("operator", operator.Name).
		Str("operator_id", operator.ID).
		Str("recipient", p.Name).
		Str("recipient_id", p.ID).
		Int64("amount", amount).Msg("deposit")

	msg := fmt.Sprintf("💰`%s` đã bơm vào %dk.", p.Name, amount)
	if amount < 0 {
		msg = fmt.Sprintf("💸 `%s` đã rút ra %dk.", p.Name, -amount)
	}
	h.broadcast(h.game.AllPlayers(nil), msg, false)
}
