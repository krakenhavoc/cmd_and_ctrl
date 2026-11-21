package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Absolute Grace — Enchantment {1}{W} (Urza's Saga):
//
//	"All creatures have protection from black."
//
// A layer 6 keyword grant over every creature on the battlefield, every
// player's, as printed (Mass Hysteria's shape). The granted token is the
// one a printed "protection from black" carries, so damage, blocking,
// targeting and attaching all read it (ADR 0072).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41785f7b-3d02-4b7f-86ff-b381f626142e",
		Name:         "Absolute Grace",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.B29AllCreatures, "protection from black"),
		},
	})
}
