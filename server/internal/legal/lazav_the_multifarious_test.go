package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

const oracleLazavTheMultifarious = "c14bcef2-6d49-4430-86d0-e5a87ba442d9"

// lazav_the_multifarious_test.go — #1723's enumerator half. "With
// mana value X" is NOT monotonic in X the way "X or less" is: a
// larger X can have FEWER legal targets than a smaller one, so the
// single "largest affordable X" every other {X} ability in this
// package offers (x_abilities_test.go) would routinely offer nothing
// here. The enumerator instead tries every affordable X and keeps
// only the ones with an actual legal target — #544's rule, over the
// pair (X, target) rather than over target alone.
func TestLazavTheMultifariousEnumeratesOnlyLegalXTargetPairs(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	lazav := battlefieldCard(g, active, game.Card{
		Name: "Lazav, the Multifarious", TypeLine: "Legendary Creature — Shapeshifter",
		OracleID: oracleLazavTheMultifarious, Power: 1, Toughness: 3,
	})
	// Creature cards at mana value 0, 1 and 3 — deliberately NONE at
	// 2, and none above 3, even though the mana below affords X up to
	// 5. If the enumerator picked one X the way the ordinary {X}
	// ladder does, it would pick X=5 and offer NOTHING, or bind every
	// candidate to the wrong X and offer an illegal move.
	mv0 := graveyardCard(active, game.Card{Name: "MV0", TypeLine: "Creature — Test", ManaCost: "{0}"})
	mv1 := graveyardCard(active, game.Card{Name: "MV1", TypeLine: "Creature — Test", ManaCost: "{B}"})
	mv3 := graveyardCard(active, game.Card{Name: "MV3", TypeLine: "Creature — Test", ManaCost: "{3}"})
	mana(g, active, 5)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, lazav)

	want := map[int]uuid.UUID{0: mv0, 1: mv1, 3: mv3}
	if len(acts) != len(want) {
		t.Fatalf("got %d Lazav activations, want %d (one per X with a legal target): %v",
			len(acts), len(want), labels(moves))
	}
	seen := make(map[int]bool, len(acts))
	for _, m := range acts {
		x := xValueOf(t, m)
		var p struct {
			Targets []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params %s: %v", string(m.Params), err)
		}
		wantID, ok := want[x]
		if !ok {
			t.Errorf("offered an activation at X=%d, which has no creature card of that mana value", x)
			continue
		}
		if seen[x] {
			t.Errorf("X=%d offered more than once", x)
		}
		seen[x] = true
		if len(p.Targets) != 1 || p.Targets[0].ID != wantID.String() {
			t.Errorf("X=%d targets %v, want exactly [%s]", x, p.Targets, wantID)
		}
	}

	// The soundness invariant every enumerator test in this package
	// carries: the engine accepts every move offered, exactly as
	// announced.
	dispatchAll(t, g, active.ID, moves)
}
