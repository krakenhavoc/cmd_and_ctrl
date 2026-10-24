package protocol

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_view_test.go — ADR 0208 §6: a seat's sheets are public (CR
// 123.2c) with each sticker marked when it is on an object, the stickers
// on a public object ride its card, tickets ride the player counters, and
// a placement is one log line.

func TestStickersOnTheWire(t *testing.T) {
	g := attractionViewGame(t)
	me := g.Seats[0]
	sheet := game.StickerSheet{
		OracleID: uuid.New().String(),
		Name:     "Night Brushwagg Ringmaster",
		Stickers: []game.StickerPrint{
			{Kind: game.StickerName, Text: "Night"},
			{Kind: game.StickerPT, Text: "2/3", TicketCost: 2, Power: 2, Toughness: 3},
		},
	}
	bear := game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() {
		me.StickerSheets = []game.StickerSheet{sheet}
		g.Battlefield.PushTop(bear)
		if err := g.AddPlayerCounterByForEffect(me.ID, me.ID, game.CounterTicket, 3); err != nil {
			t.Fatal(err)
		}
	})
	if !g.PlaceStickerForTest(me.ID, bear.InstanceID, game.StickerID(sheet.OracleID, 1), 0) {
		t.Fatal("the sticker did not go on")
	}
	for _, viewer := range []string{me.ID.String(), g.Seats[1].ID.String(), SpectatorViewerID} {
		v := ViewOfGameFor(g, viewer)
		seat := v.Seats[0]
		if seat.Counters[game.CounterTicket] != 1 {
			t.Errorf("viewer %q: tickets %d, want 1", viewer, seat.Counters[game.CounterTicket])
		}
		if len(seat.StickerSheets) != 1 || len(seat.StickerSheets[0].Stickers) != 2 {
			t.Fatalf("viewer %q: sheets %+v", viewer, seat.StickerSheets)
		}
		if s := seat.StickerSheets[0].Stickers; s[0].InUse || !s[1].InUse {
			t.Errorf("viewer %q: in_use %v %v, want false true", viewer, s[0].InUse, s[1].InUse)
		}
		if v.Seats[1].StickerSheets != nil {
			t.Errorf("viewer %q: a seat with no sheets shows %+v", viewer, v.Seats[1].StickerSheets)
		}
	}
	card := cardViewByID(t, g, bear.InstanceID)
	if len(card.Stickers) != 1 || card.Stickers[0].Text != "2/3" || card.Power != 2 || card.Toughness != 3 {
		t.Errorf("the bear's stickers %+v, P/T %d/%d", card.Stickers, card.Power, card.Toughness)
	}
	entry := findLog(t, ViewOfGame(g).Log, LogStickerPlaced)
	if want := fmt.Sprintf("%s put the 2/3 sticker on Grizzly Bears (2 tickets)", me.Name); entry.Text != want {
		t.Errorf("log %q, want %q", entry.Text, want)
	}
}
