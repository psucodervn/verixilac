package telegram

import (
	"bytes"
	"fmt"
	"os"
	"strings"

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
			Text:        "newroom",
			Description: "Tạo phòng mới",
		},
		{
			Text:        "join",
			Description: "Tham gia vào phòng. Cú pháp: /join room_id",
		},
		{
			Text:        "leave",
			Description: "Rời phòng",
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
			Text:        "rooms",
			Description: "Xem danh sách phòng",
		},
		{
			Text:        "rules",
			Description: "Xem danh sách rule",
		},
		{
			Text:        "setrule",
			Description: "Thay đổi rule. Cú pháp: /setrule rule_id",
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

	h.game.OnNewRoom(h.onNewRoom)
	h.game.OnNewGame(h.onNewGame)
	h.game.OnPlayerJoinRoom(h.onPlayerJoinRoom)
	h.game.OnPlayerBet(h.onPlayerBet)
	h.game.OnPlayerStand(h.onPlayerStand)
	h.game.OnPlayerHit(h.onPlayerHit)
	h.game.OnPlayerPlay(h.onPlayerPlay)
	h.game.OnGameFinish(h.onGameFinish)

	h.bot.Handle("/start", h.CmdStart)
	h.bot.Handle("/newroom", h.CmdNewRoom)
	h.bot.Handle("/newgame", h.CmdNewGame)
	h.bot.Handle("/join", h.CmdJoinRoom)
	h.bot.Handle("/leave", h.CmdLeaveRoom)
	h.bot.Handle("/endgame", h.CmdEndGame)
	h.bot.Handle("/save", h.CmdSave)
	h.bot.Handle("/room", h.CmdRoomInfo)
	h.bot.Handle("/rooms", h.CmdListRoom)
	h.bot.Handle("/pass", h.CmdPass)
	h.bot.Handle("/status", h.CmdStatus)
	h.bot.Handle("/rules", h.CmdListRules)
	h.bot.Handle("/setrule", h.CmdSetRule)
	h.bot.Handle("/admin", h.CmdAdmin)

	h.bot.Handle(telebot.OnQuery, func(q *telebot.Query) {
		log.Info().Interface("q", q).Msg("on query")
	})

	h.bot.Handle(telebot.OnCallback, h.onCallback)

	h.bot.Handle(telebot.OnText, func(m *telebot.Message) {
		log.Info().Msg(m.Text + " " + GetUsername(m.Chat))
		p := h.joinServer(m)
		if r := p.CurrentRoom(); r != nil {
			ps := FilterPlayers(r.Players(), p.ID())
			h.sendChat(ps, "🗣 "+GetUsername(m.Chat)+": "+m.Text)
		}
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
	p := h.joinServer(m)
	ruleID := strings.TrimSpace(m.Payload)
	r, ok := game.DefaultRules[ruleID]
	if !ok {
		h.sendMessage(m.Chat, "Không tìm thấy rule: "+ruleID)
		return
	}
	p.SetRule(ruleID)
	_ = h.game.SaveToStorage()
	h.sendMessage(m.Chat, "Đã thay đổi rule của bạn thành: "+r.Name+". Tạo game mới để cảm nhận!")
}

func (h *Handler) CmdStatus(m *telebot.Message) {
	p := h.joinServer(m)
	r := p.Rule()
	msg := fmt.Sprintf(`Thông tin của bạn:
- ID: %s
- Name: %s
- Balance: %dk
- Rule: %s (%s)
`, p.ID(), p.Name(), p.Balance(), r.ID, r.Name)
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

// CmdNewRoom creates a new room and assign called user as dealer
func (h *Handler) CmdNewRoom(m *telebot.Message) {
	p := h.joinServer(m)
	_, err := h.game.NewRoom(h.ctx(m), p)
	if err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
}

func (h *Handler) CmdNewGame(m *telebot.Message) {
	h.doNewGame(m, false)
}

func (h *Handler) doNewGame(m *telebot.Message, onQuery bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	p := h.joinServer(m)
	ctx := h.ctx(m)
	r := p.CurrentRoom()
	if r == nil {
		h.sendMessage(m.Chat, "Bạn chưa vào phòng")
		return
	}
	g, err := h.game.NewGame(r, p)
	if err != nil {
		h.sendMessage(m.Chat, "Không thể tạo ván mới: "+err.Error())
		return
	}

	if len(os.Getenv("TEST_ACCOUNT")) > 0 {
		_ = g
		_ = ctx
		botP1 := game.NewPlayer("123", "Test 1", 1000)
		_ = h.game.PlayerBet(ctx, g, botP1, 50)
		botP2 := game.NewPlayer("456", "Test 2", 2000)
		_ = h.game.PlayerBet(ctx, g, botP2, 100)
	}
}

// CmdJoinRoom joins user to room
func (h *Handler) CmdJoinRoom(m *telebot.Message) {
	h.doJoinRoom(m, false)
}

func (h *Handler) CmdLeaveRoom(m *telebot.Message) {
	p := h.joinServer(m)
	if r, err := h.game.LeaveRoom(h.ctx(m), p); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
	} else {
		h.sendMessage(m.Chat, "Bạn đã rời khỏi phòng "+r.ID())
	}
}

func (h *Handler) CmdRoomInfo(m *telebot.Message) {
	p := h.joinServer(m)
	r := p.CurrentRoom()
	if r == nil {
		h.sendMessage(m.Chat, "Bạn chưa vào phòng")
		return
	}
	h.sendMessage(m.Chat, r.Info())
}

func (h *Handler) CmdListRoom(m *telebot.Message) {
	rooms, err := h.game.Rooms(h.ctx(m))
	if err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	bf := bytes.NewBuffer(nil)
	for _, r := range rooms {
		bf.WriteString(fmt.Sprintf("Phòng %s:\n", r.ID()))
		for _, p := range r.Players() {
			bf.WriteString(fmt.Sprintf(" - %s (%+dk)\n", p.Name(), p.Balance()))
		}
	}
	h.sendMessage(m.Chat, bf.String())
}

func (h *Handler) CmdPass(m *telebot.Message) {
	p := h.joinServer(m)
	g := p.CurrentGame()
	if g == nil {
		h.sendMessage(m.Chat, "Không có thông tin ván")
		return
	}
	pg, err := h.game.PlayerPass(h.ctx(m), g)
	if err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	// h.broadcast(g.AllPlayers(), pg.Name() + " đã bị qua lượt", false)
	log.Info().Str("game_id", g.ID()).Str("user_id", pg.ID()).Msg(pg.Name() + " đã bị qua lượt")
}
