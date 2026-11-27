package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dega Sanctuary — Enchantment {2}{W} (Apocalypse, #5614):
//
//	"At the beginning of your upkeep, if you control a black or red
//	 permanent, you gain 2 life. If you control a black permanent and a red
//	 permanent, you gain 4 life instead."
//
// One triggered ability with an intervening if (CR 603.4), checked as it
// triggers and again on resolution; the amount is read as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75c626fb-9dfc-4a94-b59f-f0e45c7b2f56",
		Name:         "Dega Sanctuary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB634Sanctuary("B", "R", "Dega Sanctuary — you gain 2 life (4 with a black and a red permanent)",
				nil, effects.S82OldB634YouGainLife),
		},
	})
}
