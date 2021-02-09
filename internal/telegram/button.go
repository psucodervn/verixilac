package telegram

import (
	"gopkg.in/tucnak/telebot.v2"

	"github.com/psucodervn/verixilac/internal/game"
)

type InlineButton struct {
	Text string
	Data string
	Row  int
}

func ToTelebotInlineButtons(bs []InlineButton) [][]telebot.InlineButton {
	mr := 0
	for _, b := range bs {
		if mr < b.Row {
			mr = b.Row
		}
	}
	ar := make([][]telebot.InlineButton, mr+1)
	for _, b := range bs {
		if len(b.Data) == 0 {
			b.Data = b.Text
		}
		ar[b.Row] = append(ar[b.Row], telebot.InlineButton{
			Text: b.Text,
			Data: b.Data,
		})
	}
	return ar
}

func MakeBetButtons(g *game.Game) []InlineButton {
	return []InlineButton{
		{Text: "10k", Data: "/bet " + g.ID() + " 10"},
		{Text: "20k", Data: "/bet " + g.ID() + " 20"},
		{Text: "50k", Data: "/bet " + g.ID() + " 50"},
		{Text: "100k", Data: "/bet " + g.ID() + " 100", Row: 1},
		{Text: "200k", Data: "/bet " + g.ID() + " 200", Row: 1},
		{Text: "Rút lui", Data: "/bet " + g.ID() + " 0", Row: 1},
	}
}

func MakeDealerPrepareButtons(g *game.Game) []InlineButton {
	return []InlineButton{
		{Text: "Chia bài", Data: "/deal " + g.ID()},
		{Text: "Huỷ", Data: "/cancel " + g.ID()},
	}
}

func MakeDealerPlayingButtons(g *game.Game, pg *game.PlayerInGame) []InlineButton {
	return []InlineButton{
		{Text: "Lật bài của " + pg.Name(), Data: "/compare " + g.ID() + " " + pg.ID()},
	}
}

// func MakeDealerPlayingButtons(g *game.Game) []InlineButton {
//   var ar []InlineButton
//   for _, p := range g.Players() {
//     if !p.IsDone() {
//       ar = append(ar, InlineButton{Text: "Lật bài của " + p.Name(), Data: "/compare " + g.ID() + " " + p.ID()})
//     }
//   }
//   return ar
// }

func MakePlayerButton(g *game.Game, pg *game.PlayerInGame) []InlineButton {
	var ar []InlineButton
	if pg.CanHit() {
		ar = append(ar, InlineButton{Text: "Rút thêm", Data: "/hit " + g.ID()})
	}
	if pg.CanStand() {
		if pg.IsDealer() {
			ar = append(ar, InlineButton{Text: "Thôi", Data: "/endgame " + g.ID()})
		} else {
			ar = append(ar, InlineButton{Text: "Thôi", Data: "/stand " + g.ID()})
		}
	}
	return ar
}

func MakeResultButtons(g *game.Game) []InlineButton {
	return []InlineButton{
		{Text: "Tạo ván mới", Data: "/newgame"},
	}
}

func MakeNewlyCreatedRoomButtons(r *game.Room) []InlineButton {
	return []InlineButton{
		{Text: "Tạo ván mới", Data: "/newgame"},
	}
}
