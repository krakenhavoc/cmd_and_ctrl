package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stkPrinted is a sheet by its printed name, as the deck importer reads it.
func stkPrinted(t *testing.T, name string) game.StickerSheet {
	t.Helper()
	for _, s := range StkPrintedSheets {
		if s.Name == name {
			return deck.StickerSheetOf(cards.Card{
				OracleID: uuid.MustParse(s.Oracle), Name: s.Name, TypeLine: StickerSheetTypeLine, OracleText: s.Text,
			})
		}
	}
	t.Fatalf("no printed sheet %q", name)
	return game.StickerSheet{}
}

// stkTableWith is a catalog game whose seat 0 plays with the named printed
// sheets and has `tickets` ticket counters.
func StkTableWith(t *testing.T, tickets int, sheets ...string) (*game.Game, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for _, s := range sheets {
			me.StickerSheets = append(me.StickerSheets, stkPrinted(t, s))
		}
		if tickets > 0 {
			if err := g.AddPlayerCounterByForEffect(me.ID, me.ID, game.CounterTicket, tickets); err != nil {
				t.Fatalf("tickets: %v", err)
			}
		}
	})
	return g, me
}
