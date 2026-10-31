package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exile_any_number_cost_test.go — the enumerator half of #4455: Baron
// Helmut Zemo's boast, whose cost exiles any number of black cards from the
// graveyard holding fifteen or more black mana symbols among their mana
// costs. #544's invariant: every activation offered is one the engine
// accepts, and none is offered that it would refuse.

const oracleBaronHelmutZemo = "0fe6c042-1182-4392-8938-fee36149d74f"

// zemoBoastBoard is Zemo having attacked, with `n` four-symbol black cards
// in his controller's graveyard.
func zemoBoastBoard(t *testing.T, n int) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	zemo := battlefieldCard(g, seat, game.Card{
		Name: "Baron Helmut Zemo", TypeLine: "Legendary Creature — Human Noble Villain", OracleID: oracleBaronHelmutZemo,
		ManaCost: "{B}{B}{B}", Power: 3, Toughness: 3,
	})
	var fuel []uuid.UUID
	for i := 0; i < n; i++ {
		fuel = append(fuel, graveyardCard(seat, game.Card{Name: "Black Fuel", TypeLine: "Creature — Test", ManaCost: "{B}{B}{B}{B}", Power: 1, Toughness: 1}))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(zemo, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	return g, seat, zemo, fuel
}

// Three four-symbol cards are twelve symbols: no activation. Four are
// sixteen: one activation, naming enough cards to reach fifteen, and no
// more than it needs.
func TestZemoBoastIsEnumeratedOnlyWithFifteenBlackSymbolsInTheGraveyard(t *testing.T) {
	g, seat, zemo, _ := zemoBoastBoard(t, 3)
	if got := movesFrom(legal.EnumerateFor(g, seat.ID), zemo, legal.KindActivate); len(got) != 0 {
		t.Fatalf("twelve black symbols offered %v: the cost needs fifteen", labels(got))
	}

	g, seat, zemo, fuel := zemoBoastBoard(t, 4)
	moves := legal.EnumerateFor(g, seat.ID)
	got := movesFrom(moves, zemo, legal.KindActivate)
	if len(got) == 0 {
		t.Fatalf("sixteen black symbols, but no boast offered: %v", labels(moves))
	}
	for _, m := range got {
		p := exileParamsOf(t, m)
		if len(p.ExileIDs) != 4 {
			t.Errorf("params = %s, want four exile_ids (three cards are only twelve symbols)", string(m.Params))
		}
		want := map[string]bool{}
		for _, id := range fuel {
			want[id.String()] = true
		}
		for _, id := range p.ExileIDs {
			if !want[id] {
				t.Errorf("exile_ids %v name a card outside the graveyard fuel", p.ExileIDs)
			}
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}
