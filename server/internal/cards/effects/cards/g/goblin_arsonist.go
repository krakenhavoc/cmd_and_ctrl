package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Arsonist — Creature — Goblin Shaman {R}, 1/1 (Foundations
// Jumpstart):
//
//	"When this creature dies, you may have it deal 1 damage to any
//	 target."
//
// Dynamite Diver's dies-ping (the damage comes from the creature's
// last-known object, CR 603.10a) behind a CR 603.5 "you may".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1177f22-a1cf-4da3-a68d-ff954e878403",
		Name:         "Goblin Arsonist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Optional(effects.Targeting(
				effects.WhenThisDies("Goblin Arsonist — 1 damage to any target", effects.SourceDealsDamageToEachLegalTarget(1)),
				effects.TargetAny()), "Goblin Arsonist — have it deal 1 damage to any target?"),
				effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
