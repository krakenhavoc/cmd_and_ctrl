package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5e78e6e0-60f0-4c3c-b8dd-bd673f5152b8",
		Name:         "Viashino Pyromancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Viashino Pyromancer — 2 damage to target player or planeswalker", sourceDealsDamageToEachLegalTarget(2)),
				targetPlayerOrPlaneswalker()), ForTargets(DamageToTarget(0, 2))),
		},
	})
}
