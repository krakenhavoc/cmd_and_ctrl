package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Penumbra Kavu — Creature — Kavu {4}{G}, 3/3 (Apocalypse, #5614):
//
//	"When this creature dies, create a 3/3 black Kavu creature token."
//
// A dies trigger (CR 700.4) that makes one token under its controller's
// control, whether or not the Kavu is still findable. The token has no
// printed card (only a black Kavu in print is the all-colours one), so its
// template follows the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0abcf934-2552-4885-bbe7-38d93ce0e642",
		Name:         "Penumbra Kavu",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisDies("Penumbra Kavu — create a 3/3 black Kavu creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 black Kavu"), N: 1})),
				game.Purpose{Tokens: 1}),
		},
	})
}
