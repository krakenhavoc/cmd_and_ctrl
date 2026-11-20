package legal_test

import (
	"testing"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Slaughter
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// optional_life_cost_test.go — the enumerator half of ADR 0100's
// 2026-11-20 amendment (#5782): an optional cost that pays a fixed
// amount of life (Slaughter's "Buyback—Pay 4 life") is offered only
// while the seat can pay it (CR 119.4) and survive it, and the life it
// pays rides the move's cost, so the policy does not price it as free.

const oracleSlaughter = "37de575d-15d4-42a2-a3d2-3855fb6ab110"

func TestSlaughterBuybackIsOfferedOnlyWhenTheLifeCanBePaid(t *testing.T) {
	for _, tc := range []struct {
		life       int
		wantBought bool
	}{
		{life: 40, wantBought: true},
		{life: 5, wantBought: true},
		// CR 119.4 lets a seat pay exactly its life total, and the next
		// state-based check would kill it (CR 704.5a): not offered.
		{life: 4},
		{life: 3},
	} {
		g := newTable(t)
		active := g.Seats[g.Turn.ActiveSeat]
		clearHand(active)
		active.Life = tc.life
		card := handCard(active, game.Card{Name: "Slaughter", TypeLine: "Instant", ManaCost: "{2}{B}{B}", OracleID: oracleSlaughter})
		lands(g, active, "Swamp", "Swamp", 4)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		battlefieldCard(g, opp, creature("Grizzly Bears", "{1}{G}", 2, 2))
		advanceTo(t, g, game.StepPrecombatMain)

		moves := legal.EnumerateFor(g, active.ID)
		var bought, plain int
		for _, m := range castMovesFor(moves, card) {
			if !kickedOf(t, m) {
				plain++
				if m.Cost != nil && m.Cost.Life != 0 {
					t.Errorf("life %d: the plain cast %q charges %d life", tc.life, m.Label, m.Cost.Life)
				}
				continue
			}
			bought++
			if m.Cost == nil || m.Cost.Life != 4 {
				t.Errorf("life %d: the bought-back cast %q has cost %+v, want 4 life", tc.life, m.Label, m.Cost)
			}
		}
		if plain == 0 {
			t.Errorf("life %d: the cast without buyback was not offered", tc.life)
		}
		if (bought > 0) != tc.wantBought {
			t.Errorf("life %d: offered %d bought-back casts, want any = %v", tc.life, bought, tc.wantBought)
		}
		dispatchAll(t, g, active.ID, moves)
	}
}
