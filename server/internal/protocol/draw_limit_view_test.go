package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draw_limit_view_test.go — #4016: a seat under a per-turn draw limit
// (CR 121.2b) carries it on the wire, with the cap, the draws left and
// the permanents imposing it, so the client can badge it.

const viewDrawLimiter = "view-fixture-draw-limiter"

func stubViewDrawLimiter(t *testing.T) {
	t.Helper()
	prevLookup := game.CatalogLookup
	game.CatalogLookup = func(key string) *game.CardDef {
		if key == viewDrawLimiter {
			return &game.CardDef{}
		}
		if prevLookup == nil {
			return nil
		}
		return prevLookup(key)
	}
	prev := game.CatalogDrawLimits
	game.CatalogDrawLimits = func(key string) []game.DrawLimitStatic {
		if key == viewDrawLimiter {
			return []game.DrawLimitStatic{{Whose: game.DrawLimitOpponents, Max: 1}}
		}
		if prev == nil {
			return nil
		}
		return prev(key)
	}
	t.Cleanup(func() {
		game.CatalogLookup = prevLookup
		game.CatalogDrawLimits = prev
	})
}

func seatDrawLimit(g *game.Game, player uuid.UUID) *DrawLimitView {
	for _, s := range ViewOfGame(g).Seats {
		if s.ID == player.String() {
			return s.DrawLimit
		}
	}
	return nil
}

func TestDrawLimitReachesTheSeatsItCovers(t *testing.T) {
	stubViewDrawLimiter(t)
	g := buildThreeSeatGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if seatDrawLimit(g, opp.ID) != nil {
		t.Fatal("a draw limit before anything imposes one")
	}

	narset := game.NewCard("Narset, Parter of Veils", me.ID)
	narset.OracleID = viewDrawLimiter
	narset.TypeLine = "Legendary Planeswalker — Narset"
	g.WithWriteLock(func() { g.Battlefield.PushTop(narset) })

	got := seatDrawLimit(g, opp.ID)
	if got == nil {
		t.Fatal("the opponent's seat carries no draw limit")
	}
	if got.Max != 1 || got.Left != 1 || !slices.Equal(got.Sources, []string{"Narset, Parter of Veils"}) {
		t.Errorf("draw_limit = %+v, want max 1, left 1, from Narset", *got)
	}
	if seatDrawLimit(g, me.ID) != nil {
		t.Error("Narset's controller is not limited, but their seat says so")
	}

	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	if got := seatDrawLimit(g, opp.ID); got == nil || got.Left != 0 {
		t.Errorf("after one draw, draw_limit = %+v, want no draws left", got)
	}
}
