package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yuan-Ti Fang-Blade — Creature — Snake Rogue {2}{B}, 2/2:
//
//	"Deathtouch
//	 Whenever this creature deals combat damage to a player, venture
//	 into the dungeon."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8ef06396-861d-4599-9976-6e5c04575456",
		Name:            "Yuan-Ti Fang-Blade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(ventureLabel("Yuan-Ti Fang-Blade"), Venture),
		},
	})
}
