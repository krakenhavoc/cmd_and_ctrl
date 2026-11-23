package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Absolute Law — Enchantment {1}{W} (Urza's Saga):
//
//	"All creatures have protection from red."
//
// Absolute Grace's twin: a layer 6 grant of the printed "protection from
// red" token to every creature on the battlefield, every player's
// (ADR 0072).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4aa8391-34d5-4ad2-a981-9aa08f8229db",
		Name:         "Absolute Law",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.B29AllCreatures, "protection from red"),
		},
	})
}
