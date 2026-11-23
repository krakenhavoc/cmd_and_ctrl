package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ore-Scale Guardian — Creature — Dragon {5}{R}{R}, 4/4 (Modern Horizons, #5519):
//
//	"This spell costs {1} less to cast for each land card in your graveyard.
//	 Flying, haste"
//
// Rumbleweed's cost reduction: {1} off for each land card in your
// graveyard, read as the cost is paid. It never touches {R}{R}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c4224ae5-4807-45cb-bd4b-70f2c0d9d9f4",
		Name:            "Ore-Scale Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.Cmdr2S14LandCardsInYourGraveyard,
				"This spell costs {1} less to cast for each land card in your graveyard."),
		},
	})
}
