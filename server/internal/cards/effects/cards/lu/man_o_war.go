package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Man-o'-War — Creature — Jellyfish {2}{U}, 2/2 (Visions, #5631):
//
//	"When this creature enters, return target creature to its owner's
//	 hand."
//
// The trigger targets as it goes on the stack (CR 603.3d), any creature,
// the Man-o'-War itself included. A target that has left does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67a3541c-8408-40c8-b44f-90035b860f57",
		Name:         "Man-o'-War",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Man-o'-War — return target creature to its owner's hand",
				effects.BounceTheTargetToHand), effects.TargetCreature("target creature")),
		},
	})
}
