package telegram

import (
	"fmt"
	"os"
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
	if p == nil {
		return fmt.Errorf("player not found")
	}

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
	case "rename":
		if len(ss) < 3 {
			h.sendMessage(m.Chat, "Cú pháp: /admin rename player_id new_name")
			return nil
		}
		h.doRename(m, ss[1], strings.Join(ss[2:], " "))
	case "cancel":
		h.doAdminCancel(m)
	case "pause":
		h.doAdminPause(m)
	case "resume":
		h.doAdminResume(m)
	case "deposit":
		h.doDeposit(m, p, ss[1:])
	case "reset":
		h.doResetBalance(m, p, ss[1:])
	case "restart":
		os.Exit(1)
	}
	return nil
}

func (h *Handler) doResetBalance(m *telebot.Message, operator *model.Player, ss []string) {
	if len(ss) != 1 {
		h.sendMessage(m.Chat, "Cú pháp: /reset new_balance")
		return
	}

	balance, err := strconv.ParseInt(ss[0], 10, 64)
	if err != nil {
		h.sendMessage(m.Chat, "Cú pháp: /reset new_balance")
		return
	}

	if err := h.store.ResetBalance(h.ctx(m), balance); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
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

	msg := fmt.Sprintf("💰`%s` đã bơm vào %s.", p.Name, stringer.FormatCurrency(amount))
	if amount < 0 {
		msg = fmt.Sprintf("💸 `%s` đã rút ra %s.", p.Name, stringer.FormatCurrency(-amount))
	}
	h.broadcast(h.game.AllPlayers(nil), msg, false)
}

func (h *Handler) doAdminCancel(m *telebot.Message) {
	ctx := h.ctx(m)
	if err := h.game.CancelGame(ctx); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	h.broadcast(h.game.AllPlayers(ctx), "🚫 Ván chơi hiện tại đã bị huỷ, bạn có thể tạo ván mới!", false)
}
