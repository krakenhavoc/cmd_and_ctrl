package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// sacrifice_half_test.go — #5075: a cast whose additional cost is
// "sacrifice half the lands you control, rounded up" (Tectonic Split) is
// offered with exactly that many of the seat's lands, and every move is
// one the engine accepts (#544).

const oracleTectonicSplit = "b71c3c1e-f927-4192-ac3a-cf6a442ffd00"

func TestSacrificeHalfNamesHalfTheLandsRoundedUp(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	// Thirteen Forests: seven go, six pay {4}{G}{G}.
	basicLands(g, seat, 13, "Forest")
	basicLands(g, g.Seats[1], 3, "Forest")
	split := handCard(seat, game.Card{Name: "Tectonic Split", TypeLine: "Enchantment", ManaCost: "{4}{G}{G}", OracleID: oracleTectonicSplit})
	var mine []legal.Move
	all := legal.EnumerateFor(g, seat.ID)
	for _, m := range all {
		if m.Source == split && m.Kind == legal.KindCast {
			mine = append(mine, m)
		}
	}
	if len(mine) == 0 {
		t.Fatal("Tectonic Split is not offered")
	}
	for _, p := range varSacMovesOf(t, mine, split) {
		if len(p.SacrificeIDs) != 7 {
			t.Fatalf("sacrifice_ids = %v, want seven of thirteen lands", p.SacrificeIDs)
		}
		for _, id := range p.SacrificeIDs {
			c, ok := g.LookupCardForEffect(uuid.MustParse(id))
			if !ok || c.Controller != seat.ID || !c.IsLand() {
				t.Fatalf("sacrifice %s is not one of the seat's lands", id)
			}
		}
	}
	dispatchAll(t, g, seat.ID, all)
}
