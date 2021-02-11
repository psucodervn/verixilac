package telegram

import (
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
	cmd := strings.TrimSpace(m.Payload)
	switch cmd {
	case "pause":
		h.doAdminPause(m)
	case "resume":
		h.doAdminResume(m)
	}
}

func (h *Handler) doAdminPause(m *telebot.Message) {
	if err := h.game.Pause(h.ctx(m)); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
	}
	h.sendMessage(m.Chat, "Server đã pause")
	return
}

func (h *Handler) doAdminResume(m *telebot.Message) {
	if err := h.game.Resume(h.ctx(m)); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
	}
	h.sendMessage(m.Chat, "Server đã resume")
	return
}
