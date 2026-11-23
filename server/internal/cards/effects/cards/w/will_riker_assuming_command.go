package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Will Riker, Assuming Command — Legendary Creature — Human Officer {4}{W},
// 4/5 (Star Trek Commander, #4974):
//
//	"When Will Riker enters, put a +1/+1 counter on another target
//	 creature you control."
//
// Necklace of Girion's counter on the one target, chosen as the trigger
// goes on the stack and re-checked at resolution (CR 608.2b). "Another"
// keeps Riker himself out; with no other creature the trigger is not put
// on the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a479b8b6-79de-4810-9165-992a6a10dc1d",
		Name:         "Will Riker, Assuming Command",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Will Riker, Assuming Command — put a +1/+1 counter on another target creature you control",
					effects.PutPlusOneCounterOnEachLegalTarget),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
