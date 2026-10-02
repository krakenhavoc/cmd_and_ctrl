package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keldon Champion — Creature — Human Barbarian, {2}{R}{R}, 3/2:
//
//	"Haste
//	 Echo {2}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, it deals 3 damage to target player or planeswalker."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1600b361-9298-49f0-8925-4c1aad363e1b",
		Name:            "Keldon Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Echo("Keldon Champion", "{2}{R}{R}"),
			Targeting(WhenThisEnters("Keldon Champion — 3 damage to target player or planeswalker", sourceDealsDamageToEachLegalTarget(3)),
				targetPlayerOrPlaneswalker()),
		},
	})
}
