package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// hideaway_moves_test.go — the enumerator half of ADR 0091 (#1331). A
// card hidden face down by hideaway is a cast surface only under the
// free-play grant its linked ability writes, and only for the player
// that grant names; the enumerator reads the same CastPermissionForLocked
// the cast path does, and dispatchAll proves the offered play is one
// the engine accepts — out of a face-down exile, for {0}.
func TestEnumeratorOffersAHiddenCardOnlyUnderItsGrant(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)

	src := game.Card{InstanceID: uuid.New(), Name: "Enum Heights", TypeLine: "Land", Owner: seat.ID, Controller: seat.ID}
	g.Battlefield.PushTop(src)
	hidden := game.NewCard("Enum Hidden Sorcery", seat.ID)
	hidden.TypeLine = "Sorcery"
	hidden.ManaCost = "{7}"
	seat.Library.PushTop(hidden)
	g.WithWriteLock(func() {
		if _, err := g.ExileHiddenForEffect(game.ObjectRefOf(src), seat.ID, hidden.InstanceID); err != nil {
			t.Fatalf("ExileHiddenForEffect: %v", err)
		}
	})

	if got := castMovesFor(legal.EnumerateFor(g, seat.ID), hidden.InstanceID); len(got) != 0 {
		t.Fatalf("a hidden card with no grant was offered %d casts", len(got))
	}

	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(hidden.InstanceID, game.CastPermission{
			Player: seat.ID, Zone: game.ZoneExile, Cost: "{0}", Timing: game.TimingFlash,
		})
	})
	moves := legal.EnumerateFor(g, seat.ID)
	if got := castMovesFor(moves, hidden.InstanceID); len(got) != 1 {
		t.Fatalf("hidden card under its grant offered %d casts, want 1: %v", len(got), labels(moves))
	}
	for _, p := range g.Seats {
		if p.ID == seat.ID {
			continue
		}
		if got := castMovesFor(legal.EnumerateFor(g, p.ID), hidden.InstanceID); len(got) != 0 {
			t.Errorf("seat %s was offered the hidden card", p.Name)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}
