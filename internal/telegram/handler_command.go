package telegram

import (
	"bytes"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"gopkg.in/telebot.v3"

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

	h.bot.Handle(telebot.OnQuery, func(ctx telebot.Context) error {
		q := ctx.Query()
		log.Info().Interface("q", q).Msg("on query")
		return nil
	})

	h.bot.Handle(telebot.OnCallback, h.onCallback)

	h.bot.Handle(telebot.OnText, func(ctx telebot.Context) error {
		m := ctx.Message()
		log.Info().Msg(m.Text + " " + GetUsername(m.Chat))
		p := h.getPlayer(m)
		if p == nil {
			p = h.joinServer(m)
		}

		ps := FilterPlayers(h.game.AllPlayers(h.ctx(m)), p.ID)
		h.sendChat(ps, "📣 `"+p.Name+":` "+m.Text)
		return nil
	})

	h.bot.Start()
	return
}

func (h *Handler) CmdStart(ctx telebot.Context) error {
	h.joinServer(ctx.Message())
	return nil
}

func (h *Handler) CmdJoin(ctx telebot.Context) error {
	h.joinServer(ctx.Message())
	return nil
}

func (h *Handler) CmdLeave(ctx telebot.Context) error {
	m := ctx.Message()
	id := cast.ToString(m.Chat.ID)
	if err := h.game.PlayerLeave(h.ctx(m), id); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
	}
	return nil
}

func (h *Handler) CmdRoom(ctx telebot.Context) error {
	m := ctx.Message()
	ps, err := h.store.ListPlayers(h.ctx(m))
	if err != nil {
		h.sendMessage(m.Chat, "Lỗi: "+err.Error())
		return nil
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
	return nil
}

func (h *Handler) CmdListRules(ctx telebot.Context) error {
	m := ctx.Message()
	h.sendMessage(m.Chat, game.RuleListText)
	return nil
}

func (h *Handler) CmdSetRule(ctx telebot.Context) error {
	m := ctx.Message()
	h.sendMessage(m.Chat, "Chức năng tạm thời bị vô hiệu hóa")
	return nil
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

func (h *Handler) CmdHistory(ctx telebot.Context) error {
	m := ctx.Message()
	h.sendMessage(m.Chat, "Chức năng tạm thời bị vô hiệu hóa")
	return nil
}

func (h *Handler) CmdStatus(ctx telebot.Context) error {
	m := ctx.Message()
	p := h.getPlayer(m)
	if p == nil {
		p = h.joinServer(m)
	}

	msg := fmt.Sprintf("Thông tin của bạn:\n"+
		"- ID: `%s`\n"+
		"- Name: %s\n"+
		"- Balance: `%s`\n"+
		"- Status: %s\n",
		p.ID, p.Name, stringer.FormatCurrency(p.Balance), p.UserStatus)
	h.sendMessage(m.Chat, msg)

	return nil
}

func (h *Handler) CmdEndGame(ctx telebot.Context) error {
	h.doEndGame(ctx.Message(), false)
	return nil
}

func (h *Handler) CmdNewGame(ctx telebot.Context) error {
	h.doNewGame(ctx.Message(), false)
	return nil
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

func (h *Handler) CmdPass(ctx telebot.Context) error {
	// p := h.getPlayer(m)
	m := ctx.Message()
	pg, err := h.game.PlayerPass(h.ctx(m))
	if err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return err
	}
	log.Info().Str("user_id", pg.ID).Msg(pg.Name + " đã bị qua lượt")
	return nil
}
