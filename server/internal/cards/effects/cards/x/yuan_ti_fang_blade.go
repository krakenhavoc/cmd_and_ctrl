package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yuan-Ti Fang-Blade — Creature — Snake Rogue {2}{B}, 2/2:
//
//	"Deathtouch
//	 Whenever this creature deals combat damage to a player, venture
//	 into the dungeon."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ef06396-861d-4599-9976-6e5c04575456",
		Name:            "Yuan-Ti Fang-Blade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(effects.VentureLabel("Yuan-Ti Fang-Blade"), effects.Venture),
		},
	})
}
