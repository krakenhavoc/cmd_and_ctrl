package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urborg Drake — Creature — Drake {1}{U}{B}, 2/3 (Battlebond):
//
//	"Flying
//	 This creature attacks each combat if able."
//
// Rubblebelt Recluse's attack requirement (CR 508.1d) with a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "57d93683-8739-49ef-856a-5a70f92d8763",
		Name:            "Urborg Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
