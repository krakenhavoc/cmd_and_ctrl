package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f4aa8391-34d5-4ad2-a981-9aa08f8229db",
		Name:         "Absolute Law",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b29AllCreatures, "protection from red"),
		},
	})
}
