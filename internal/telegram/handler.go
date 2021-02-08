package telegram

import (
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"gopkg.in/tucnak/telebot.v2"

	"github.com/psucodervn/verixilac/internal/game"
	"github.com/psucodervn/verixilac/internal/stringer"
)

type Handler struct {
	game *game.Manager
	bot  *telebot.Bot

	gameMessages sync.Map

	mu sync.RWMutex
}

func NewHandler(manager *game.Manager, bot *telebot.Bot) *Handler {
	return &Handler{
		game: manager,
		bot:  bot,
	}
}

func (h *Handler) onCallback(q *telebot.Callback) {
	// log.Info().Interface("data", q.Data).Interface("text", q.Message.Text).Msg("on callback")
	ar := strings.SplitN(q.Data, " ", 2)
	if len(ar) > 1 {
		q.Message.Payload = ar[1]
	}
	switch ar[0] {
	case "/join":
		h.doJoinRoom(q.Message, true)
	case "/bet":
		h.doBet(q.Message, true)
	case "/deal":
		// dealer deal cards
		h.doDeal(q.Message, true)
	case "/cancel":
		// dealer cancel game
		h.doCancel(q.Message, true)
	case "/hit":
		h.doHit(q.Message, true)
	case "/stand":
		h.doStand(q.Message, true)
	case "/endgame":
		h.doEndGame(q.Message, true)
	case "/compare":
		h.doCompare(q.Message, true)
	case "/newgame":
		h.doNewGame(q.Message, true)
	default:
		log.Warn().Str("cmd", ar[0]).Msg("unknown query command")
	}
}

func (h *Handler) doBet(m *telebot.Message, onQuery bool) {
	ar := strings.Split(m.Payload, " ")
	if len(ar) != 2 {
		h.sendMessage(m.Chat, "Sai cú pháp")
		return
	}

	p := h.joinServer(m)
	ctx := h.ctx(m)
	gameID := strings.TrimSpace(ar[0])
	g := h.game.FindGame(ctx, gameID)
	if g == nil {
		h.sendMessage(m.Chat, "Không có thông tin ván "+gameID)
		return
	}

	amount := cast.ToUint64(ar[1])
	if amount < 0 {
		h.sendMessage(m.Chat, "Số tiền cược không hợp lệ")
		return
	}
	if err := h.game.PlayerBet(ctx, g, p, amount); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
}

func (h *Handler) doDeal(m *telebot.Message, onQuery bool) {
	ctx := h.ctx(m)
	gameID := strings.TrimSpace(m.Payload)
	g := h.game.FindGame(ctx, gameID)
	if g == nil {
		h.sendMessage(m.Chat, "Không tìm thấy ván "+gameID)
		return
	}

	if err := h.game.Deal(ctx, g); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}

	h.broadcast(append(g.Players(), g.Dealer()), "Chốt deal:\n\n"+g.PreparingBoard(), true)

	// send cards
	for _, pg := range g.Players() {
		h.sendMessage(ToTelebotChat(pg.ID()), "Bài của bạn: "+pg.Cards().String(false))
	}
	h.sendMessage(ToTelebotChat(g.Dealer().ID()), "Bài của bạn: "+g.Dealer().Cards().String(false, true))

	// start game
	if err := h.game.Start(ctx, g); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}

	// FIXME: fake
	// m1 := &telebot.Message{ID: 123, Payload: g.ID(), Chat: &telebot.Chat{ID: 123, Username: "Test 1"}}
	// for {
	//   ok := h.doStand(m1, true)
	//   if ok {
	//     break
	//   }
	//   ok = h.doHit(m1, true)
	// }
	//
	// m2 := &telebot.Message{ID: 456, Payload: g.ID(), Chat: &telebot.Chat{ID: 456, Username: "Test 2"}}
	// for {
	//   ok := h.doStand(m2, true)
	//   if ok {
	//     break
	//   }
	//   ok = h.doHit(m2, true)
	// }
}

func (h *Handler) doCancel(m *telebot.Message, onQuery bool) {
	gameID := strings.TrimSpace(m.Payload)
	ctx := h.ctx(m)
	g := h.game.FindGame(ctx, gameID)
	if g == nil {
		h.sendMessage(m.Chat, "Không tìm thấy ván "+gameID)
		return
	}
}

func (h *Handler) onNewRoom(r *game.Room, creator *game.Player) {
	// send to creator
	h.sendMessage(ToTelebotChat(creator.ID()), "Bạn đã tạo phòng "+r.ID(), MakeNewlyCreatedRoomButtons(r)...)

	// send to other players in bot
	players := FilterPlayers(h.game.Players(), creator.ID())
	msg := creator.Name() + " đã tạo phòng " + r.ID()
	buttons := []InlineButton{{Text: "Vào phòng", Data: "/join " + r.ID()}}
	h.broadcast(players, msg, false, buttons...)
}

func (h *Handler) onNewGame(r *game.Room, g *game.Game) {
	msg := "Bắt đầu ván mới, hãy tham gia ngay!\n\n" + g.PreparingBoard()

	// send to dealer
	d := g.Dealer()
	h.broadcast(d.Player, msg, false, MakeDealerPrepareButtons(g)...)

	// send to members
	players := FilterPlayers(r.Players(), d.ID())
	h.broadcast(players, msg, false, MakeBetButtons(g)...)
}

func (h *Handler) onPlayerJoinRoom(r *game.Room, p *game.Player) {
	players := FilterPlayers(r.Players(), p.ID())
	h.broadcast(players, p.Name()+" vừa vào phòng "+r.ID(), false)
}

func (h *Handler) onPlayerBet(g *game.Game, p *game.PlayerInGame) {
	msg := "Bắt đầu ván mới, hãy tham gia ngay!\n\n" + g.PreparingBoard()
	dealer := g.Dealer()
	h.broadcast(dealer.Player, msg, true, MakeDealerPrepareButtons(g)...)

	r := g.Room()
	players := FilterPlayers(r.Players(), dealer.ID())
	h.broadcast(players, msg, true, MakeBetButtons(g)...)
}

func (h *Handler) doJoinRoom(m *telebot.Message, onQuery bool) {
	p := h.joinServer(m)
	roomID := strings.TrimSpace(m.Payload)
	r := h.game.FindRoom(h.ctx(m), roomID)
	if r == nil {
		h.sendMessage(m.Chat, stringer.Capitalize("Không tìm thấy phòng "+roomID))
		return
	}

	if err := h.game.JoinRoom(h.ctx(m), p, r); err != nil {
		if err == game.ErrPlayerAlreadyInRoom {
			err = game.ErrYouAlreadyInRoom
		}
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	if onQuery {
		h.editMessage(m, "Bạn đã vào phòng "+roomID)
	} else {
		h.sendMessage(m.Chat, "Bạn đã vào phòng "+roomID)
	}
}

func (h *Handler) doPlay(m *telebot.Message, onQuery bool, hit bool) bool {
	p := h.joinServer(m)
	gameID := strings.TrimSpace(m.Payload)
	g, pg := h.findPlayerInGame(m, gameID, p.ID())
	if g == nil {
		h.sendMessage(m.Chat, "Không tìm thấy ván "+gameID)
		return false
	}
	if pg == nil {
		h.sendMessage(m.Chat, "Bạn không có trong ván "+gameID)
		return false
	}

	ctx := h.ctx(m)
	var err error
	if hit {
		err = h.game.PlayerHit(ctx, g, pg)
	} else {
		err = h.game.PlayerStand(ctx, g, pg)
	}
	if err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return false
	}
	return true
}

func (h *Handler) doEndGame(m *telebot.Message, onQuery bool) bool {
	p := h.joinServer(m)
	gameID := strings.TrimSpace(m.Payload)
	if len(gameID) == 0 {
		if p.CurrentGame() == nil {
			h.sendMessage(m.Chat, "Bạn chưa vào ván")
			return false
		}
		gameID = p.CurrentGame().ID()
	}

	g, pg := h.findPlayerInGame(m, gameID, p.ID())
	if g == nil {
		h.sendMessage(m.Chat, "Không tìm thấy ván "+gameID)
		return false
	}
	if pg == nil || !pg.IsDealer() {
		h.sendMessage(m.Chat, "Bạn không phải nhà cái")
		return false
	}

	if err := h.game.FinishGame(h.ctx(m), g); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return false
	}
	return true
}

func (h *Handler) doStand(m *telebot.Message, onQuery bool) bool {
	return h.doPlay(m, onQuery, false)
}

func (h *Handler) doHit(m *telebot.Message, onQuery bool) bool {
	return h.doPlay(m, onQuery, true)
}

// joinServer check and register user
func (h *Handler) joinServer(m *telebot.Message) *game.Player {
	id := cast.ToString(m.Chat.ID)
	name := GetUsername(m.Chat)
	return h.game.PlayerRegister(h.ctx(m), id, name)
}

func (h *Handler) onPlayerStand(g *game.Game, pg *game.PlayerInGame) {
	h.broadcast(g.Dealer(), pg.Name()+" đã úp bài", false, MakeDealerPlayingButtons(g, pg)...)
	h.broadcast(g.Players(), pg.Name()+" đã úp bài", false)
}

func (h *Handler) onPlayerHit(g *game.Game, pg *game.PlayerInGame) {
	players := FilterInGamePlayers(append(g.Players(), g.Dealer()), pg.ID())
	h.broadcast(players, pg.Name()+" vừa rút thêm 1 lá", false)
	h.broadcast(pg, "Bài của bạn: "+pg.Cards().String(false, pg.IsDealer()), true, MakePlayerButton(g, pg)...)
}

func (h *Handler) doCompare(m *telebot.Message, onQuery bool) {
	ar := strings.Split(m.Payload, " ")
	if len(ar) != 2 {
		h.sendMessage(m.Chat, "Sai cú pháp")
		return
	}
	p := h.joinServer(m)
	g, dealer := h.findPlayerInGame(m, ar[0], p.ID())
	if g == nil || dealer == nil || !dealer.IsDealer() {
		h.sendMessage(m.Chat, "Sai thông tin")
		return
	}
	to := g.FindPlayer(ar[1])
	if to == nil {
		h.sendMessage(m.Chat, "Sai thông tin")
		return
	}
	if st := to.Status(); st != game.PlayerStood {
		if st < game.PlayerStood {
			h.sendMessage(m.Chat, "Người chơi chưa rút xong")
		} else {
			h.sendMessage(m.Chat, "Đã tính rồi")
		}
		return
	}

	g.Done(to)
	h.game.CheckIfFinish(h.ctx(m), g)
}

func (h *Handler) onGameFinish(g *game.Game) {
	_ = h.game.SaveToStorage()
	msg := "Kết quả ván chơi!\n\n" + g.ResultBoard()
	h.broadcast(g.Dealer(), msg, false, MakeResultButtons(g)...)
	h.broadcast(g.Players(), msg, false, MakeResultButtons(g)...)
}

func (h *Handler) onPlayerPlay(g *game.Game, pg *game.PlayerInGame) {
	if pg.IsDealer() {
		msg := "Thống kê ván hiện tại:\n" + g.PlayerBoard()
		h.broadcast(g.Dealer(), msg, false)
	}
	all := append(g.Players(), g.Dealer())
	h.broadcast(pg, "Tới lượt bạn: "+pg.Cards().String(false, pg.IsDealer()), false, MakePlayerButton(g, pg)...)
	h.broadcast(FilterInGamePlayers(all, pg.ID()), "Tới lượt "+pg.Name(), false)
}
