package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orcish Hellraiser — Creature — Orc Warrior, {1}{R}, 3/2:
//
//	"Echo {R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature dies, it deals 2 damage to target player or planeswalker."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd21cbaa-7537-414b-b720-3c7d6fc7c93d",
		Name:         "Orcish Hellraiser",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Orcish Hellraiser", "{R}"),
			Targeting(WhenThisDies("Orcish Hellraiser — 2 damage to target player or planeswalker", sourceDealsDamageToEachLegalTarget(2)),
				targetPlayerOrPlaneswalker()),
		},
	})
}
