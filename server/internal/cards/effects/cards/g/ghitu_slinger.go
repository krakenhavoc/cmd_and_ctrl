package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghitu Slinger — Creature — Human Nomad, {2}{R}, 2/2:
//
//	"Echo {2}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, it deals 2 damage to any target."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e2b5f90-e405-4b24-a71a-0988c37cf67a",
		Name:         "Ghitu Slinger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Ghitu Slinger", "{2}{R}"),
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Ghitu Slinger — 2 damage to any target", effects.SourceDealsDamageToEachLegalTarget(2)), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
