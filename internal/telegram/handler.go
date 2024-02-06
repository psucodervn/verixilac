package telegram

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"gopkg.in/telebot.v3"

	"github.com/psucodervn/verixilac/internal/game"
	"github.com/psucodervn/verixilac/internal/model"
	"github.com/psucodervn/verixilac/internal/stringer"
)

type Handler struct {
	game  *game.Manager
	bot   *telebot.Bot
	store game.Storage

	gameMessages sync.Map
	dealMessages sync.Map

	mu sync.RWMutex
}

func NewHandler(manager *game.Manager, bot *telebot.Bot, store game.Storage) *Handler {
	return &Handler{
		game:  manager,
		bot:   bot,
		store: store,
	}
}

func (h *Handler) onCallback(ctx telebot.Context) error {
	q := ctx.Callback()

	// log.Info().Interface("data", q.Data).Interface("text", q.Message.Text).Msg("on callback")
	ar := strings.SplitN(q.Data, " ", 2)
	if len(ar) > 1 {
		q.Message.Payload = ar[1]
	}
	switch ar[0] {
	case "/join":
		h.doJoin(q.Message, true)
	case "/bet":
		h.doBet(q.Message, true)
	case "/deal":
		// dealer deal cards
		h.doDeal(q.Message, true)
	case "/cancel":
		// dealer cancel game
		h.doCancel(q.Message, true)
	case "/hit":
		h.doHit(q.Message, false)
	case "/stand":
		h.doStand(q.Message, true, false)
	case "/endgame":
		h.doEndGame(q.Message, true)
	case "/compare":
		h.doCompare(q.Message, true)
	case "/newgame":
		h.doNewGame(q.Message, true)
	default:
		log.Warn().Str("cmd", ar[0]).Msg("unknown query command")
	}

	return nil
}

func (h *Handler) doJoin(m *telebot.Message, onQuery bool) {
	_ = h.getPlayer(m)
	h.sendMessage(m.Chat, "Bạn đã vào sòng")
}

func (h *Handler) doBet(m *telebot.Message, onQuery bool) {
	ar := strings.Split(m.Payload, " ")
	if len(ar) != 2 {
		h.sendMessage(m.Chat, "Sai cú pháp")
		return
	}

	p := h.getPlayer(m)
	ctx := h.ctx(m)
	gameID := strings.TrimSpace(ar[0])

	amount := cast.ToUint64(ar[1])
	if amount < 0 {
		h.sendMessage(m.Chat, "Số tiền cược không hợp lệ")
		return
	}
	if err := h.game.PlayerBet(ctx, gameID, p, amount); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
}

func (h *Handler) doDeal(m *telebot.Message, onQuery bool) {
	ctx := h.ctx(m)
	gameID := strings.TrimSpace(m.Payload)

	g, err := h.game.Deal(ctx, gameID)
	if err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}

	h.broadcast(g.Players(), "Chốt deal:\n\n"+g.PreparingBoard(), true)

	// send cards
	for _, pg := range g.PlayersInGame() {
		if !pg.IsDone() && !pg.IsBot() {
			h.sendMessage(ToTelebotChat(pg.ID), "Bài của bạn: "+pg.Cards().String(false))
		}
	}
	h.sendMessage(ToTelebotChat(g.Dealer().ID), "Bài của bạn: "+g.Dealer().Cards().String(false, true))

	// start game
	if err := h.game.Start(ctx, g); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}

	if !g.Finished() {
		for _, pg := range g.PlayersInGame() {
			if !pg.IsDone() {
				continue
			}
			msg := fmt.Sprintf("Bài của %s: %s\n%s đã thắng %dk",
				pg.Name, pg.Cards().String(false, false),
				pg.Name, pg.Reward())
			h.broadcast(g.AllPlayers(), msg, false)
		}
	}

	// auto bot play for testing purpose
	if autoBotCount > 0 {
		fakePlay(h, g, autoBotCount)
	}
}

func (h *Handler) doCancel(m *telebot.Message, onQuery bool) {
	p := h.getPlayer(m)
	gameID := strings.TrimSpace(m.Payload)
	ctx := h.ctx(m)
	g, pg := h.findPlayerInGame(m, gameID, p.ID)
	if g == nil || pg == nil {
		return
	}
	if !pg.IsDealer() {
		h.sendMessage(m.Chat, "Bạn không phải nhà cái")
		return
	}
	if err := h.game.CancelGame(ctx, g); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	h.broadcast(h.game.ActivePlayers(ctx), pg.Name+" đã huỷ ván này", true, InlineButton{
		Text: "Tạo ván mới", Data: "/newgame",
	})
}

func (h *Handler) onNewGame(g *game.Game) {
	msg := "Bắt đầu ván mới, hãy tham gia ngay!\n\n" + g.PreparingBoard()

	// send to dealer
	d := g.Dealer()
	h.broadcast(d.Player, msg, false, MakeDealerPrepareButtons(g)...)

	// send to members
	players := FilterPlayers(h.game.ActivePlayers(context.TODO()), d.ID)
	h.broadcast(players, msg, false, MakeBetButtons(g)...)
}

func (h *Handler) onPlayerJoin(p *model.Player) {
	h.broadcast(h.game.AllPlayers(context.TODO()), p.Name+" vừa vào sòng", false)
}

func (h *Handler) onPlayerLeave(p *model.Player) {
	h.broadcast(h.game.AllPlayers(context.TODO()), p.Name+" vừa ra khỏi sòng", false)
}

func (h *Handler) onPlayerBet(g *game.Game, p *game.PlayerInGame) {
	msg := "Bắt đầu ván mới, hãy tham gia ngay!\n\n" + g.PreparingBoard()
	dealer := g.Dealer()
	h.broadcast(dealer.Player, msg, true, MakeDealerPrepareButtons(g)...)

	players := FilterPlayers(h.game.ActivePlayers(context.TODO()), dealer.ID)
	h.broadcast(players, msg, true, MakeBetButtons(g)...)
}

func (h *Handler) doEndGame(m *telebot.Message, onQuery bool) bool {
	p := h.getPlayer(m)
	gameID := strings.TrimSpace(m.Payload)
	if len(gameID) == 0 {
		g := h.game.CurrentGame()
		if g == nil {
			h.sendMessage(m.Chat, "Bạn chưa vào ván")
			return false
		}
		gameID = g.ID()
	}

	g, pg := h.findPlayerInGame(m, gameID, p.ID)
	if g == nil || pg == nil {
		return false
	}
	if !pg.IsDealer() {
		h.sendMessage(m.Chat, "Bạn không phải nhà cái")
		return false
	}

	if err := h.game.FinishGame(h.ctx(m), g, false); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return false
	}
	if onQuery {
		if _, err := h.bot.EditReplyMarkup(m, nil); err != nil {
			log.Err(err).Msg("edit reply markup failed")
		}

	}
	return true
}

func (h *Handler) doStand(m *telebot.Message, onQuery bool, isBot bool) bool {
	p := h.getPlayer(m, isBot)
	gameID := strings.TrimSpace(m.Payload)
	g, pg := h.findPlayerInGame(m, gameID, p.ID)
	if g == nil || pg == nil {
		return false
	}

	if err := h.game.PlayerStand(h.ctx(m), g, pg); err != nil {
		if !pg.IsBot() {
			h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		}
		return false
	}
	if onQuery {
		_, _ = h.bot.EditReplyMarkup(m, nil)
	}
	return true
}

func (h *Handler) doHit(m *telebot.Message, isBot bool) bool {
	p := h.getPlayer(m, isBot)
	force := false
	ar := strings.Split(strings.TrimSpace(m.Payload), " ")
	if len(ar) >= 2 {
		force = true
	}
	gameID := ar[0]
	g, pg := h.findPlayerInGame(m, gameID, p.ID)
	if g == nil || pg == nil {
		return false
	}

	if !force && pg.CanHit() && pg.Cards().Value() >= 18 {
		h.editMessage(m, "Bài của bạn: "+pg.Cards().String(false, pg.IsDealer())+"\nBạn chắc chắn muốn rút thêm?", MakePlayerButton(g, pg, true)...)
		return false
	}

	if err := h.game.PlayerHit(h.ctx(m), g, pg); err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return false
	}
	return true
}

func (h *Handler) getPlayer(m *telebot.Message, isBot ...bool) *model.Player {
	id := cast.ToString(m.Chat.ID)
	p, err := h.store.GetPlayerByID(h.ctx(m), id)
	if err != nil {
		log.Err(err).Str("user_id", id).Msg("get player failed")
		return nil
	}
	return p
}

// joinServer check and register user
func (h *Handler) joinServer(m *telebot.Message, isBot ...bool) *model.Player {
	id := cast.ToString(m.Chat.ID)
	name := GetUsername(m.Chat)
	role := model.UserRoleNormal
	if len(isBot) > 0 && isBot[0] {
		name = "Bot #" + id
		role = model.UserRoleBot
	}

	p, existed := h.game.PlayerRegister(h.ctx(m), id, name, role)
	if !existed {
		h.onPlayerJoin(p)
	}
	return p
}

func (h *Handler) onPlayerStand(g *game.Game, pg *game.PlayerInGame) {
	// h.broadcast(g.Dealer(), pg.Name+" đã úp bài", false)
	// h.broadcast(g.AllPlayers(), pg.Name+" đã úp bài", false)
}

func (h *Handler) onPlayerHit(g *game.Game, pg *game.PlayerInGame) {
	players := FilterInGamePlayers(g.AllPlayers(), pg.ID)
	h.broadcast(players, "`"+pg.Name+"` vừa rút thêm 1 lá", false)
	h.broadcast(pg, "Bài của bạn: "+pg.Cards().String(false, pg.IsDealer()), true, MakePlayerButton(g, pg, false)...)
}

func (h *Handler) doCompare(m *telebot.Message, onQuery bool) {
	ar := strings.Split(m.Payload, " ")
	if len(ar) != 2 {
		h.sendMessage(m.Chat, "Sai cú pháp")
		return
	}

	p := h.getPlayer(m)
	g, dealer := h.findPlayerInGame(m, ar[0], p.ID)
	if g == nil || dealer == nil {
		return
	}
	if !dealer.IsDealer() {
		h.sendMessage(m.Chat, "Bạn không phải nhà cái")
		return
	}
	if !dealer.CanStand() {
		h.sendMessage(m.Chat, "Bạn chưa đủ tẩy")
		return
	}
	to := g.FindPlayer(ar[1])
	if to == nil {
		h.sendMessage(m.Chat, "Sai thông tin")
		return
	}

	reward, err := g.Done(to, false)
	if err != nil {
		h.sendMessage(m.Chat, stringer.Capitalize(err.Error()))
		return
	}
	if h.game.CheckIfFinish(h.ctx(m), g) {
		return
	}

	msgDealer := fmt.Sprintf("Bài của %s: %s",
		to.Name, to.Cards().String(false, false),
	)

	var msgPlayer string
	if reward < 0 {
		msgDealer += fmt.Sprintf("\n%s thắng và được cộng %dk", to.Name, -reward)
		msgPlayer = fmt.Sprintf("🤑 Cái lật bài bạn và thua. Bạn được cộng %dk", -reward)
	} else if reward > 0 {
		msgDealer += fmt.Sprintf("\n%s thua và bị trừ %dk", to.Name, reward)
		msgPlayer = fmt.Sprintf("🔻 Cái lật bài bạn và thắng. Bạn bị trừ %dk", reward)
	} else {
		msgDealer += fmt.Sprintf("\n%s và cái hoà nhau", to.Name)
		msgPlayer = fmt.Sprintf("🤝 Cái lật bài bạn và hoà. Bạn không bị mất tiền")
	}
	msgPlayer += fmt.Sprintf("\nBài của cái: %s",
		dealer.Cards().String(false, true),
	)

	if onQuery {
		h.editMessage(m, msgDealer)
	} else {
		h.sendMessage(ToTelebotChat(dealer.ID), msgDealer)
	}
	h.sendMessage(ToTelebotChat(to.ID), msgPlayer)
}

func (h *Handler) onGameFinish(g *game.Game) {
	ctx := context.TODO()

	for _, p := range g.PlayersInGame() {
		_, _ = h.store.AddPlayerBalance(ctx, p.Player.ID, p.Reward())
	}
	_, _ = h.store.AddPlayerBalance(ctx, g.Dealer().Player.ID, g.Dealer().Reward())

	// _ = h.game.SaveToStorage()
	msg := "Kết quả ván chơi!\n\n" + g.ResultBoard()
	h.broadcast(g.AllPlayers(), msg, false, MakeResultButtons(g)...)
}

func (h *Handler) onPlayerPlay(g *game.Game, pg *game.PlayerInGame) {
	if pg.IsDealer() {
		for _, p := range g.PlayersInGame() {
			if p.IsDone() {
				continue
			}
			msg := fmt.Sprintf("%s đang cầm %d lá", p.Name, len(p.Cards()))
			h.broadcast(g.Dealer(), msg, false, MakeDealerPlayingButtons(g, p)...)
		}
	}
	h.broadcast(pg, "Tới lượt bạn: "+pg.Cards().String(false, pg.IsDealer()), false, MakePlayerButton(g, pg, false)...)
	h.broadcast(FilterInGamePlayers(g.AllPlayers(), pg.ID), "Tới lượt `"+pg.Name+"`", false)
}

func (h *Handler) sendChat(receivers []model.Player, msg string) {
	wg := sync.WaitGroup{}
	for _, p := range receivers {
		if p.IsBot() {
			continue
		}

		wg.Add(1)
		p := p

		go func() {
			defer wg.Done()
			_, err := h.bot.Send(ToTelebotChat(p.TelegramID), msg)
			if err != nil {
				log.Err(err).Str("receiver", p.Name).Str("msg", msg).Msg("send message failed")
			}
		}()
	}

	wg.Wait()
}
