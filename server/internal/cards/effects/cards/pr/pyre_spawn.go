package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyre Spawn — Creature — Elemental {4}{R}{R}, 6/4 (Innistrad: Crimson
// Vow):
//
//	"When this creature dies, it deals 3 damage to any target."
//
// Dynamite Diver's dies trigger for 3: dealt by the Spawn as it last
// existed (CR 608.2h), target chosen as the trigger goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13b0a3f3-04a5-465f-898b-13a28f795a16",
		Name:         "Pyre Spawn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisDies("Pyre Spawn — 3 damage to any target", effects.SourceDealsDamageToEachLegalTarget(3)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 3))),
		},
	})
}
