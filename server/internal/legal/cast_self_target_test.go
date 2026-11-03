package legal_test

import (
	"fmt"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cast_self_target_test.go — #4869, CR 601.2a. Cling to Dust ("Exile
// target card from a graveyard", escape) cast from the graveyard is on
// the stack before its target is chosen, so the enumerator never offers
// the bot a cast that names the card itself. Every move it does offer is
// accepted by the engine.
func TestEnumeratorNeverOffersAGraveyardCastNamingItself(t *testing.T) {
	const oracleClingToDust = "9c44e556-4c7a-48bf-a108-c584f69c2cfa"
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 4, "Swamp")

	cling := graveyardCard(seat, game.Card{
		Name: "Cling to Dust", TypeLine: "Instant", ManaCost: "{B}",
		OracleID: oracleClingToDust,
	})
	for i := 0; i < 5; i++ {
		graveyardCard(seat, game.Card{Name: fmt.Sprintf("Fodder %d", i), TypeLine: "Sorcery"})
	}

	moves := legal.EnumerateFor(g, seat.ID)
	casts := castPayloadsOf(t, moves, cling)
	if len(casts) == 0 {
		t.Fatalf("no escape cast offered for Cling to Dust: %v", labels(moves))
	}
	for _, p := range casts {
		for _, tr := range p.Targets {
			if tr.ID == cling.String() {
				t.Errorf("the enumerator offered Cling to Dust targeting itself: %+v", p)
			}
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}
