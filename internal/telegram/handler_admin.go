package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/tucnak/telebot.v2"

	"github.com/psucodervn/verixilac/internal/stringer"
)

func (h *Handler) CmdAdmin(m *telebot.Message) {
	p := h.joinServer(m)
	if !p.IsAdmin() {
		h.sendMessage(m.Chat, "Bạn không có quyền admin")
		return
	}
	ss := strings.Split(strings.TrimSpace(m.Payload), " ")
	if len(ss) == 0 {
		return
	}

	cmd := ss[0]
	switch cmd {
	case "pause":
		h.doAdminPause(m)
	case "resume":
		h.doAdminResume(m)
	case "deposit":
		h.doDeposit(m, ss[1:])
	}
}

func (h *Handler) doAdminPause(m *telebot.Message) {
	if err := h.game.Pause(h.ctx(m)); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	h.sendMessage(m.Chat, "Server đã pause")
}

func (h *Handler) doAdminResume(m *telebot.Message) {
	if err := h.game.Resume(h.ctx(m)); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	h.sendMessage(m.Chat, "Server đã resume")
}

func (h *Handler) doDeposit(m *telebot.Message, ss []string) {
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

	if err := h.game.Deposit(h.ctx(m), id, amount); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	h.sendMessage(m.Chat, fmt.Sprintf("Đã nạp %dK cho người chơi %s.", amount, id))
}
