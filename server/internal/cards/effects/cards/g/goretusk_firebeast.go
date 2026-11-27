package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goretusk Firebeast — Creature — Elemental Boar Beast {5}{R}, 2/2 (Judgment,
// #5611):
//
//	"When this creature enters, it deals 4 damage to target player or
//	 planeswalker."
//
// Viashino Pyromancer's trigger at four damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "853ce64f-50b4-49a6-9aa3-f07b11cd8b69",
		Name:         "Goretusk Firebeast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Goretusk Firebeast — 4 damage to target player or planeswalker", effects.SourceDealsDamageToEachLegalTarget(4)),
				effects.TargetPlayerOrPlaneswalker()), effects.ForTargets(effects.DamageToTarget(0, 4))),
		},
	})
}
