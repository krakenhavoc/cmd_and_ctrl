package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "457bee4c-2332-440c-9336-36f6a69614f9",
		Name:         "Goblin Medics",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
					"Goblin Medics — 1 damage to any target", effects.SourceDealsDamageToEachLegalTarget(1)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
