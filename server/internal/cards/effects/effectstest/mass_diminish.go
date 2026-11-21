package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mass_diminish_test.go pins the catalog half of the CR 611.2
// duration model (#756's siblings live in control_cards_test.go):
// "until your next turn" really does last a full round in a
// four-player game, and the affected set is locked in at resolution.

const MassDiminishOracle = "3a3fd911-b71c-49eb-aadb-cfa80b9e3a3a"

// ctrlPushCreature seeds a creature on the battlefield with the
// zone-move event fired, so the layer listener stamps
// EnteredBattlefieldAt (the key every affected set is pinned on).
func CtrlPushCreature(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// ctrlEffectivePower reads a card's post-layer power.
func CtrlEffectivePower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return EffectivePower(t, g, id)
}

// advancePastCleanupForTest walks the cursor until the active seat
// changes — the marker that this seat's cleanup step ran (Turn.Round
// counts rounds, not seat-turns, so the seat is the reliable signal).
func AdvancePastCleanupForTest(t *testing.T, g *game.Game) {
	t.Helper()
	start := g.Turn.ActiveSeat
	for i := 0; i < 40; i++ {
		if g.Turn.ActiveSeat != start {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("cursor never left seat %d's turn", start)
}
