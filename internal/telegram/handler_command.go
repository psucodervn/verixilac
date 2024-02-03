package telegram

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
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

	if err = h.game.LoadFromStorage(); err != nil {
		log.Err(err).Msg("load data failed")
	}

	h.game.OnNewGame(h.onNewGame)
	h.game.OnPlayerJoin(h.onPlayerJoin)
	h.game.OnPlayerBet(h.onPlayerBet)
	h.game.OnPlayerStand(h.onPlayerStand)
	h.game.OnPlayerHit(h.onPlayerHit)
	h.game.OnPlayerPlay(h.onPlayerPlay)
	h.game.OnGameFinish(h.onGameFinish)

	h.bot.Handle("/start", h.CmdStart)
	h.bot.Handle("/newgame", h.CmdNewGame)
	// h.bot.Handle("/join", h.CmdJoin)
	// h.bot.Handle("/leave", h.CmdLeave)
	h.bot.Handle("/endgame", h.CmdEndGame)
	h.bot.Handle("/save", h.CmdSave)
	h.bot.Handle("/pass", h.CmdPass)
	h.bot.Handle("/status", h.CmdStatus)
	h.bot.Handle("/rules", h.CmdListRules)
	h.bot.Handle("/setrule", h.CmdSetRule)
	h.bot.Handle("/history", h.CmdHistory)
	h.bot.Handle("/admin", h.CmdAdmin)

	h.bot.Handle(telebot.OnQuery, func(q *telebot.Query) {
		log.Info().Interface("q", q).Msg("on query")
	})

	h.bot.Handle(telebot.OnCallback, h.onCallback)

	h.bot.Handle(telebot.OnText, func(m *telebot.Message) {
		log.Info().Msg(m.Text + " " + GetUsername(m.Chat))
		p := h.joinServer(m)
		ps := FilterPlayers(h.game.Players(), p.ID)
		h.sendChat(ps, "🗣 "+GetUsername(m.Chat)+": "+m.Text)
	})

	h.bot.Start()
	return
}

func (h *Handler) CmdStart(m *telebot.Message) {
	h.joinServer(m)
}

func (h *Handler) CmdListRules(m *telebot.Message) {
	h.sendMessage(m.Chat, game.RuleListText)
}

func (h *Handler) CmdSetRule(m *telebot.Message) {
	h.sendMessage(m.Chat, "Chức năng tạm thời bị vô hiệu hóa")
	// p := h.joinServer(m)
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
	_ = h.joinServer(m)
	h.sendMessage(m.Chat, h.game.ListPlayers(context.Background()))
}

func (h *Handler) CmdStatus(m *telebot.Message) {
	p := h.joinServer(m)
	r := game.DefaultRule
	msg := fmt.Sprintf("Thông tin của bạn:\n"+
		"- ID: `%s`\n"+
		"- Name: %s\n"+
		"- Balance: `%s` k\n"+
		"- Rule: %s (%s)\n",
		p.ID, p.Name, stringer.FormatCurrency(p.Balance), r.ID, r.Name)
	h.sendMessage(m.Chat, msg)
}

func (h *Handler) CmdSave(m *telebot.Message) {
	if err := h.game.SaveToStorage(); err != nil {
		h.sendMessage(m.Chat, "Save failed: "+err.Error())
	}
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

	p := h.joinServer(m)

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
	// p := h.joinServer(m)
	pg, err := h.game.PlayerPass(h.ctx(m))
	if err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	log.Info().Str("user_id", pg.ID).Msg(pg.Name + " đã bị qua lượt")
}
