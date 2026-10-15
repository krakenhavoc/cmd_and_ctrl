package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// tiered_test.go — #3777, CR 702.183a: the enumerator half of tiered.
// A tiered spell is one cast per bullet, never two bullets, and only
// the tiers the seat can pay for.

const oracleFireMagic = "c87db803-3490-41bc-af7a-78f99bcd77c8"

// Fire Magic ({R}; Fire {0}, Fira {2}, Firaga {5}) with three
// Mountains: Fire and Fira are affordable, Firaga is not, and no two
// bullets are ever offered together, however many the seat could pay
// for.
func TestTieredOffersOneAffordableTierAtATime(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	spell := handCard(active, game.Card{
		Name: "Fire Magic", TypeLine: "Instant",
		OracleID: oracleFireMagic, ManaCost: "{R}",
	})
	lands(g, active, "Mountain", "Mountain", 3)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	combos := explosiveDerailmentModes(castMovesFor(moves, spell), spell)
	if !combos["0"] || !combos["1"] {
		t.Errorf("Fire ({R}) and Fira ({2}{R}) fit three Mountains and should be offered; combos: %v", combos)
	}
	if combos["2"] {
		t.Errorf("Firaga costs {5}{R}, more than three Mountains; combos: %v", combos)
	}
	for k := range combos {
		if len(k) != 1 {
			t.Errorf("tiered offered %q — exactly one bullet is chosen", k)
		}
	}

	// #544: every offer this seat sees, the engine accepts.
	dispatchAll(t, g, active.ID, castMovesFor(moves, spell))
}
