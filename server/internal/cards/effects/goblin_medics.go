package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Medics — Creature — Goblin Shaman {2}{R}, 1/1:
//
//	"Whenever this creature becomes tapped, it deals 1 damage to any
//	 target."
//
// "Becomes tapped" is the tap event and the attack that taps it
// (ThisBecameTapped), as Attentive Sunscribe reads it. The target is
// chosen as the trigger goes on the stack and the damage is Goblin
// Medics' own, dealt only while the target is still legal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "457bee4c-2332-440c-9336-36f6a69614f9",
		Name:         "Goblin Medics",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
					"Goblin Medics — 1 damage to any target", sourceDealsDamageToEachLegalTarget(1)),
				TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}
