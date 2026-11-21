package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pitchburn Devils — Creature — Devil {4}{R}, 3/3 (Innistrad, #5572):
//
//	"When this creature dies, it deals 3 damage to any target."
//
// Pyre Spawn's death trigger: the target is chosen as the trigger goes
// on the stack, and the creature, now in the graveyard, is the source of
// the damage (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86e2bbdd-a788-4f76-9ebb-ef48d9bc9dcf",
		Name:         "Pitchburn Devils",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisDies("Pitchburn Devils — 3 damage to any target", effects.SourceDealsDamageToEachLegalTarget(3)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 3))),
		},
	})
}
