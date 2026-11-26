package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rage Thrower — {5}{R} Creature — Human Shaman 4/2 (Innistrad, S82 OLD b448,
// #5572):
//
//	"Whenever another creature dies, this creature deals 2 damage to target
//	 player or planeswalker."
//
// Spiteful Prankster's trigger with 2 damage: any other creature dying, its
// controller's or not, aims the damage as the trigger goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9697a59-46e2-4612-b30b-d041a38fdd30",
		Name:         "Rage Thrower",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.On(game.EventLTB, effects.AnotherCreatureDied, "Rage Thrower — 2 damage to target player or planeswalker",
					effects.SourceDealsDamageToEachLegalTarget(2)),
				effects.TargetPlayerOrPlaneswalker()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
