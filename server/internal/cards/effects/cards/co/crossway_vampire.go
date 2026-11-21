package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crossway Vampire — Creature — Vampire {1}{R}{R}, 3/2 (Innistrad, #5572):
//
//	"When this creature enters, target creature can't block this turn."
//
// Summit Intimidator's entry: the creature is targeted as the trigger
// goes on the stack, and can't block for the rest of the turn.
//
// No simplification.
func init() {
	const label = "Crossway Vampire — target creature can't block this turn"
	effects.Register(effects.Spec{
		OracleID:     "543071da-fc89-4274-a2ae-d5d755d1f2a9",
		Name:         "Crossway Vampire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(effects.WhenThisEnters(label, effects.TargetCreatureCantBlockThisTurn(label)),
					effects.TargetCreature("target creature")),
				effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0))),
		},
	})
}
