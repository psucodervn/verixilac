package telegram

import (
	"context"
	"reflect"
	"sync"

	"github.com/rs/zerolog/log"
	"gopkg.in/tucnak/telebot.v2"

	"github.com/psucodervn/verixilac/internal/game"
)

func (h *Handler) ctx(m *telebot.Message) context.Context {
	l := log.Logger.With().
		Int64("id", m.Chat.ID).
		Str("user", GetUsername(m.Chat)).
		Logger()
	return l.WithContext(context.Background())
}

func (h *Handler) sendMessage(chat *telebot.Chat, msg string, buttons ...InlineButton) *telebot.Message {
	options := &telebot.SendOptions{}
	if len(buttons) > 0 {
		options.ReplyMarkup = &telebot.ReplyMarkup{
			InlineKeyboard: ToTelebotInlineButtons(buttons),
		}
	}
	m, err := h.bot.Send(chat, msg, options)
	if err != nil {
		log.Err(err).Str("msg", msg).Str("receiver", GetUsername(chat)).Msg("send message failed")
		// TODO: deal with nil
	}
	return m
}

func (h *Handler) editMessage(m *telebot.Message, msg string, buttons ...InlineButton) *telebot.Message {
	options := &telebot.SendOptions{}
	if len(buttons) > 0 {
		options.ReplyMarkup = &telebot.ReplyMarkup{
			InlineKeyboard: ToTelebotInlineButtons(buttons),
		}
	}
	m, err := h.bot.Edit(m, msg, options)
	if err != nil {
		l := log.Err(err)
		if m != nil {
			l = l.Str("receiver", GetUsername(m.Chat))
		}
		l.Msg("edit message failed")
		// TODO: deal with nil
	}
	return m
}

func (h *Handler) broadcast(receivers interface{}, msg string, edit bool, buttons ...InlineButton) {
	var recvs []*game.Player
	switch receivers.(type) {
	case []*game.Player:
		recvs = receivers.([]*game.Player)
	case *game.Player:
		recvs = append(recvs, receivers.(*game.Player))
	case []*game.PlayerInGame:
		tmp := receivers.([]*game.PlayerInGame)
		for i := range tmp {
			recvs = append(recvs, tmp[i].Player)
		}
	case *game.PlayerInGame:
		recvs = append(recvs, receivers.(*game.PlayerInGame).Player)
	default:
		log.Error().Str("type", reflect.TypeOf(receivers).String()).Msg("invalid receivers type")
		return
	}

	wg := sync.WaitGroup{}
	for _, p := range recvs {
		wg.Add(1)
		p := p

		go func() {
			defer wg.Done()

			// log.Debug().Str("id", p.ID()).Msg("will send to")
			var m *telebot.Message
			var err error
			options := &telebot.SendOptions{
				ReplyMarkup: &telebot.ReplyMarkup{
					InlineKeyboard: ToTelebotInlineButtons(buttons),
				},
			}
			pm, ok := h.gameMessages.Load(p.ID())
			if edit && ok && pm != nil {
				m, err = h.bot.Edit(pm.(*telebot.Message), msg, options)
			} else {
				m, err = h.bot.Send(ToTelebotChat(p.ID()), msg, options)
			}
			if err != nil {
				log.Err(err).Str("receiver", p.Name()).Str("msg", msg).Msg("send message failed")
			} else {
				h.gameMessages.Store(p.ID(), m)
			}
		}()
	}

	wg.Wait()
}

func (h *Handler) broadcastDeal(players []*game.Player, msg string, edit bool, buttons ...InlineButton) {
	options := &telebot.SendOptions{
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: ToTelebotInlineButtons(buttons),
		},
	}

	wg := sync.WaitGroup{}
	for _, p := range players {
		wg.Add(1)
		p := p

		go func() {
			defer wg.Done()

			var m *telebot.Message
			var err error
			pm, ok := h.dealMessages.Load(p.ID())
			if edit && ok && pm != nil {
				m, err = h.bot.Edit(pm.(*telebot.Message), msg, options)
			} else {
				m, err = h.bot.Send(ToTelebotChat(p.ID()), msg, options)
			}
			if err != nil {
				log.Err(err).Str("receiver", p.Name()).Str("msg", msg).Msg("send message failed")
			} else {
				h.dealMessages.Store(p.ID(), m)
			}
		}()
	}

	wg.Wait()
}

func (h *Handler) findPlayerInGame(m *telebot.Message, gameID string, playerID string) (*game.Game, *game.PlayerInGame) {
	g := h.game.FindGame(h.ctx(m), gameID)
	if g == nil {
		h.sendMessage(m.Chat, "Không tìm thấy ván "+gameID)
		return nil, nil
	}
	pg := g.FindPlayer(playerID)
	if pg == nil {
		h.sendMessage(m.Chat, "Không tìm thấy người chơi "+playerID)
		return g, nil
	}
	return g, pg
}

type Playable interface {
	ID() string
}
