package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c1177f22-a1cf-4da3-a68d-ff954e878403",
		Name:         "Goblin Arsonist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Optional(Targeting(
				WhenThisDies("Goblin Arsonist — 1 damage to any target", sourceDealsDamageToEachLegalTarget(1)),
				TargetAny()), "Goblin Arsonist — have it deal 1 damage to any target?"),
				ForTargets(DamageToTarget(0, 1))),
		},
	})
}
