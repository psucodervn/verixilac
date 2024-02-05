package telegram

import (
	"context"
	"reflect"
	"sync"

	"github.com/rs/zerolog/log"
	"gopkg.in/tucnak/telebot.v2"

	"github.com/psucodervn/verixilac/internal/game"
	"github.com/psucodervn/verixilac/internal/model"
)

func (h *Handler) ctx(m *telebot.Message) context.Context {
	l := log.Logger.With().
		Int64("id", m.Chat.ID).
		Str("user", GetUsername(m.Chat)).
		Logger()
	return l.WithContext(context.Background())
}

func (h *Handler) sendMessage(chat *telebot.Chat, msg string, buttons ...InlineButton) *telebot.Message {
	// for testing purpose
	if chat.ID >= 0 && chat.ID <= 10 {
		return nil
	}

	options := &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
	}
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
	var rcvIDs []string
	switch receivers.(type) {
	case []model.Player:
		tmp := receivers.([]model.Player)
		for i := range tmp {
			if !tmp[i].IsBot() {
				rcvIDs = append(rcvIDs, tmp[i].TelegramID)
			}
		}
	case []*model.Player:
		tmp := receivers.([]*model.Player)
		for i := range tmp {
			if !tmp[i].IsBot() {
				rcvIDs = append(rcvIDs, tmp[i].TelegramID)
			}
		}
	case *model.Player:
		if !receivers.(*model.Player).IsBot() {
			rcvIDs = append(rcvIDs, receivers.(*model.Player).TelegramID)
		}
	case []*game.PlayerInGame:
		tmp := receivers.([]*game.PlayerInGame)
		for i := range tmp {
			if !tmp[i].Player.IsBot() {
				rcvIDs = append(rcvIDs, tmp[i].Player.TelegramID)
			}
		}
	case *game.PlayerInGame:
		if !receivers.(*game.PlayerInGame).Player.IsBot() {
			rcvIDs = append(rcvIDs, receivers.(*game.PlayerInGame).Player.TelegramID)
		}
	default:
		log.Error().Str("type", reflect.TypeOf(receivers).String()).Msg("invalid receivers type")
		return
	}

	options := &telebot.SendOptions{
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: ToTelebotInlineButtons(buttons),
		},
		ParseMode: telebot.ModeMarkdown,
	}

	wg := sync.WaitGroup{}
	for _, id := range rcvIDs {
		wg.Add(1)
		id := id

		go func() {
			defer wg.Done()

			var m *telebot.Message
			var err error
			pm, ok := h.gameMessages.Load(id)
			if edit && ok && pm != nil {
				m, err = h.bot.Edit(pm.(*telebot.Message), msg, options)
			} else {
				m, err = h.bot.Send(ToTelebotChat(id), msg, options)
			}
			if err != nil {
				log.Err(err).Str("receiver_id", id).Str("msg", msg).Msg("send message failed")
			} else {
				h.gameMessages.Store(id, m)
			}
		}()
	}

	wg.Wait()
}

func (h *Handler) findPlayerInGame(m *telebot.Message, gameID string, playerID string) (*game.Game, *game.PlayerInGame) {
	g, pg := h.game.FindPlayerInGame(gameID, playerID)
	if g == nil {
		h.sendMessage(m.Chat, "Không tìm thấy ván "+gameID)
		return nil, nil
	}
	if pg == nil {
		h.sendMessage(m.Chat, "Không tìm thấy người chơi "+playerID)
		return g, nil
	}
	return g, pg
}
