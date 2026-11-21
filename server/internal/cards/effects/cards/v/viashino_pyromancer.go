package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Pyromancer — Creature — Lizard Wizard {1}{R}, 2/1
// (Foundations, #4465):
//
//	"When this creature enters, it deals 2 damage to target player or
//	 planeswalker."
//
// Keldon Champion's entry trigger with 2 damage: the target is a player
// or a planeswalker, chosen as the trigger goes on the stack, and the
// creature deals the damage if the target is still legal as it resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e78e6e0-60f0-4c3c-b8dd-bd673f5152b8",
		Name:         "Viashino Pyromancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Viashino Pyromancer — 2 damage to target player or planeswalker", effects.SourceDealsDamageToEachLegalTarget(2)),
				effects.TargetPlayerOrPlaneswalker()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
