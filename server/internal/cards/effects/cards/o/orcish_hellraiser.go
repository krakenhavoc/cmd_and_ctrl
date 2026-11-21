package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orcish Hellraiser — Creature — Orc Warrior, {1}{R}, 3/2:
//
//	"Echo {R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature dies, it deals 2 damage to target player or planeswalker."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd21cbaa-7537-414b-b720-3c7d6fc7c93d",
		Name:         "Orcish Hellraiser",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Orcish Hellraiser", "{R}"),
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisDies("Orcish Hellraiser — 2 damage to target player or planeswalker", effects.SourceDealsDamageToEachLegalTarget(2)),
				effects.TargetPlayerOrPlaneswalker()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
