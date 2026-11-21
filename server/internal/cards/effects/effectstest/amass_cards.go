package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// amass_cards_test.go — #1236, the catalog half of CR 701.47. The
// engine half (find-or-create, the multi-Army prompt, the CR 614
// windows, the layer-4 subtype) is in game/amass_test.go; everything
// here is a real card, cast or triggered through the stack.
//
// Four proofs, chosen for what each one adds:
//
//	Orcish Bowmasters    the row's named card — amass off a trigger,
//	                     twice in a turn, on the same Army
//	Dreadhorde Invasion  find-or-create across turns: one token, then
//	                     counters
//	Eternal Skylord      the SUBTYPE reaching another card's static
//	Widespread Brutality "the Army you amassed" (CR 701.47c) and the
//	                     "non-Army" read side

const (
	OrcishBowmastersOracle   = "ea5103f5-27e0-4eb1-902c-7f34652d6bf3"
	DreadhordeInvasionOracle = "01deabbb-af6a-4998-99a7-35b7cfa9ef77"
	EternalSkylordOracle     = "324a30cc-bddd-463c-9f28-c94c93a26780"
	WidespreadBrutalityOracl = "3a06f4c1-2af1-495f-914f-1ac4e26f87d4"
)

// armyOf returns the single Army creature `controller` controls, or
// fails. Every test below expects exactly one, which is the point of
// find-or-create.
func ArmyOf(t *testing.T, g *game.Game, controller uuid.UUID) *game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	var found *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && game.IsArmy(*c) {
			if found != nil {
				t.Fatalf("two Armies on the battlefield; amass must find the one already there")
			}
			found = c
		}
	}
	if found == nil {
		t.Fatal("no Army on the battlefield — the amass created nothing")
	}
	return found
}

// advanceToDrawStepOfSeat walks to the named seat's draw step.
func AdvanceToDrawStepOfSeat(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if g.Turn.Step == game.StepDraw && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached seat %d's draw step", seat)
}

// advanceToUpkeepOfSeat walks to the named seat's NEXT upkeep.
func AdvanceToUpkeepOfSeat(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat {
			return
		}
	}
	t.Fatalf("never reached seat %d's upkeep", seat)
}
