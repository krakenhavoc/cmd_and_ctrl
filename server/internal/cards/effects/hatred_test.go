package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const hatredOracle = "40601061-1b25-4db8-8b99-33324ce945cf"

// TestHatredPaysXLifeAndPumpsUntilEndOfTurn pins all three halves of
// the card in one pass: the life leaves at CAST (CR 601.2f, so it is
// gone before the spell resolves), the pump is exactly +X/+0, and it
// is a turn-scoped continuous effect rather than a permanent change.
//
// X = 0 is a legal announcement and is included because it is the
// boundary the additional cost and the pump share: nothing is paid
// and nothing is pumped.
func TestHatredPaysXLifeAndPumpsUntilEndOfTurn(t *testing.T) {
	for _, tc := range []struct {
		name string
		x    int
	}{
		{"X=0", 0},
		{"X=5", 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := seedCreature(g, "Grizzly Bears", me.ID)
			before := me.Life

			castWipe(t, g, "Hatred", "Instant", hatredOracle, game.CastSpellParams{
				XValue:  tc.x,
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
			})

			// CR 601.2f: the additional cost is paid at cast, with the
			// spell still on the stack.
			if want := before - tc.x; me.Life != want {
				t.Fatalf("life at cast %d -> %d, want %d (X life is a cast cost)", before, me.Life, want)
			}

			passPriorityAroundTable(t, g)

			if got, want := effectivePower(t, g, bear), 2+tc.x; got != want {
				t.Errorf("power after Hatred: got %d, want %d", got, want)
			}
			// +X/+0 — the toughness is deliberately untouched, which is
			// what makes the card lethal rather than resilient.
			if got := effectiveToughness(t, g, bear); got != 2 {
				t.Errorf("toughness after Hatred: got %d, want 2", got)
			}

			// "Until end of turn" expires at cleanup, so the next
			// player's upkeep sees a plain 2/2 again.
			advanceToUpkeepOf(t, g, 1)
			if got := effectivePower(t, g, bear); got != 2 {
				t.Errorf("power after the turn ended: got %d, want 2", got)
			}
		})
	}
}

// TestHatredRejectsMoreLifeThanYouHave is CR 119.4 on the additional
// cost: an unpayable X is a refused cast, not a player at -1.
func TestHatredRejectsMoreLifeThanYouHave(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := seedCreature(g, "Grizzly Bears", me.ID)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Hatred", TypeLine: "Instant",
		OracleID: hatredOracle, Owner: me.ID, Controller: me.ID,
	})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		XValue:  me.Life + 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err == nil {
		t.Fatalf("casting Hatred for more life than you have was allowed")
	}
	if !me.Hand.Contains(id) {
		t.Errorf("a rejected cast still moved the card off the hand")
	}
}
