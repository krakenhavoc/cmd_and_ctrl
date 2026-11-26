package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian Shield-Bearer — Creature — Human Soldier {1}{G}, 2/1 (Dragons of
// Tarkir, S82 old batch 331):
//
//	"Megamorph {3}{G} (You may cast this card face down as a 2/2 creature for
//	 {3}. Turn it face up any time for its megamorph cost and put a +1/+1
//	 counter on it.)
//	 When this creature is turned face up, put a +1/+1 counter on another
//	 target creature you control."
//
// Megamorph is the shared constructor. The face-up trigger is Keen-Eyed Raven's
// "another target creature you control": Another is object identity, so the
// Shield-Bearer never offers itself, and with no other creature the trigger is
// removed for want of a target (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2bd8a6a4-d0ac-4da3-be1d-5f404ae4b614",
		Name:             "Guardian Shield-Bearer",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{3}{G}")},
		Triggered: []game.TriggeredAbility{
			guardianShieldBearerTurnedFaceUp(),
		},
	})
}

func guardianShieldBearerTurnedFaceUp() game.TriggeredAbility {
	t := effects.WhenThisIsTurnedFaceUp("Guardian Shield-Bearer — put a +1/+1 counter on another target creature you control",
		effects.PlusOneCounterOnChosenTarget)
	t.Targets = effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))
	return t
}
