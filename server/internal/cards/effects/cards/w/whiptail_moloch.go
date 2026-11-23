package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whiptail Moloch — Creature — Lizard {4}{R}, 6/3 (Dissension, #5598):
//
//	"When this creature enters, it deals 3 damage to target creature you
//	 control."
//
// A forced enters trigger (CR 603.3d): the Moloch itself is a legal
// target when it is the only creature you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d57079ae-7760-4a81-895f-dc8990896ed7",
		Name:         "Whiptail Moloch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Whiptail Moloch — 3 damage to target creature you control", effects.SourceDealsDamageToEachLegalTarget(3)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
