package telegram

import (
	"bytes"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"gopkg.in/tucnak/telebot.v2"

	"github.com/psucodervn/verixilac/internal/game"
	"github.com/psucodervn/verixilac/internal/stringer"
)

var (
	commands = []telebot.Command{
		{
			Text:        "status",
			Description: "Thông tin người chơi",
		},
		{
			Text:        "newgame",
			Description: "Tạo ván mới",
		},
		{
			Text:        "endgame",
			Description: "Kết thúc ván",
		},
		{
			Text:        "join",
			Description: "Tham gia vào phòng chờ",
		},
		{
			Text:        "leave",
			Description: "Rời phòng chờ",
		},
		{
			Text:        "pass",
			Description: "Cho qua lượt",
		},
		{
			Text:        "room",
			Description: "Xem thông tin phòng",
		},
		{
			Text:        "rules",
			Description: "Xem danh sách rule",
		},
		{
			Text:        "history",
			Description: "Xem lịch sử chơi. Cú pháp: /history",
		},
	}
)

func (h *Handler) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recover: %v", r)
		}
	}()

	if err = h.bot.SetCommands(commands); err != nil {
		return
	}

	h.game.OnNewGame(h.onNewGame)
	h.game.OnPlayerJoin(h.onPlayerJoin)
	h.game.OnPlayerLeave(h.onPlayerLeave)
	h.game.OnPlayerBet(h.onPlayerBet)
	h.game.OnPlayerStand(h.onPlayerStand)
	h.game.OnPlayerHit(h.onPlayerHit)
	h.game.OnPlayerPlay(h.onPlayerPlay)
	h.game.OnGameFinish(h.onGameFinish)

	h.bot.Handle("/start", h.CmdStart)
	h.bot.Handle("/newgame", h.CmdNewGame)
	h.bot.Handle("/join", h.CmdJoin)
	h.bot.Handle("/leave", h.CmdLeave)
	h.bot.Handle("/room", h.CmdRoom)
	h.bot.Handle("/endgame", h.CmdEndGame)
	h.bot.Handle("/pass", h.CmdPass)
	h.bot.Handle("/status", h.CmdStatus)
	h.bot.Handle("/rules", h.CmdListRules)
	h.bot.Handle("/history", h.CmdHistory)
	h.bot.Handle("/admin", h.CmdAdmin)

	h.bot.Handle(telebot.OnQuery, func(q *telebot.Query) {
		log.Info().Interface("q", q).Msg("on query")
	})

	h.bot.Handle(telebot.OnCallback, h.onCallback)

	h.bot.Handle(telebot.OnText, func(m *telebot.Message) {
		log.Info().Msg(m.Text + " " + GetUsername(m.Chat))
		p := h.getPlayer(m)
		ps := FilterPlayers(h.game.AllPlayers(h.ctx(m)), p.ID)
		h.sendChat(ps, "📣 "+GetUsername(m.Chat)+": "+m.Text)
	})

	h.bot.Start()
	return
}

func (h *Handler) CmdStart(m *telebot.Message) {
	h.joinServer(m)
}

func (h *Handler) CmdJoin(m *telebot.Message) {
	h.joinServer(m)
}

func (h *Handler) CmdLeave(m *telebot.Message) {
	id := cast.ToString(m.Chat.ID)
	if err := h.game.PlayerLeave(h.ctx(m), id); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
	}
}

func (h *Handler) CmdRoom(m *telebot.Message) {
	ps, err := h.store.ListPlayers(h.ctx(m))
	if err != nil {
		h.sendMessage(m.Chat, "Lỗi: "+err.Error())
		return
	}

	bf := bytes.NewBuffer(nil)
	bf.WriteString("Danh sách người chơi:\n")
	for _, p := range ps {
		bf.WriteString(fmt.Sprintf("- %s (`%s`): %sk", p.Name, p.ID, stringer.FormatCurrency(p.Balance)))
		if !p.IsActive() {
			bf.WriteString(" (offline)")
		}
		bf.WriteString("\n")
	}
	h.sendMessage(m.Chat, bf.String())
}

func (h *Handler) CmdListRules(m *telebot.Message) {
	h.sendMessage(m.Chat, game.RuleListText)
}

func (h *Handler) CmdSetRule(m *telebot.Message) {
	h.sendMessage(m.Chat, "Chức năng tạm thời bị vô hiệu hóa")
	// p := h.getPlayer(m)
	// ruleID := strings.TrimSpace(m.Payload)
	// r, ok := game.DefaultRules[ruleID]
	// if !ok {
	// 	h.sendMessage(m.Chat, "Không tìm thấy rule: "+ruleID)
	// 	return
	// }
	// p, err := h.game.SetRule(h.ctx(m), p, ruleID)
	// h.sendMessage(m.Chat, "Đã thay đổi rule của bạn thành: "+r.Name+". Tạo game mới để cảm nhận!")
}

func (h *Handler) CmdHistory(m *telebot.Message) {
	h.sendMessage(m.Chat, "Chức năng tạm thời bị vô hiệu hóa")
}

func (h *Handler) CmdStatus(m *telebot.Message) {
	p := h.getPlayer(m)
	if p == nil {
		p = h.joinServer(m)
	}

	r := game.DefaultRule
	msg := fmt.Sprintf("Thông tin của bạn:\n"+
		"- ID: `%s`\n"+
		"- Name: %s\n"+
		"- Balance: `%s` k\n"+
		"- Rule: %s (%s)\n"+
		"- Status: %s\n",
		p.ID, p.Name, stringer.FormatCurrency(p.Balance), r.ID, r.Name, p.UserStatus)
	h.sendMessage(m.Chat, msg)
}

func (h *Handler) CmdEndGame(m *telebot.Message) {
	h.doEndGame(m, false)
}

func (h *Handler) CmdNewGame(m *telebot.Message) {
	h.doNewGame(m, false)
}

func (h *Handler) doNewGame(m *telebot.Message, onQuery bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	p := h.getPlayer(m)
	if p == nil || !p.IsActive() {
		h.sendMessage(m.Chat, "Bạn chưa vào sòng")
		return
	}

	g, err := h.game.NewGame(p)
	if err != nil {
		h.sendMessage(m.Chat, "Không thể tạo ván mới: "+err.Error())
		return
	}

	if autoBotCount > 0 {
		fakeBet(h.ctx(m), h, g, autoBotCount)
	}
}

func (h *Handler) CmdPass(m *telebot.Message) {
	// p := h.getPlayer(m)
	pg, err := h.game.PlayerPass(h.ctx(m))
	if err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	log.Info().Str("user_id", pg.ID).Msg(pg.Name + " đã bị qua lượt")
}
