package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vedalken Infiltrator — Creature — Vedalken Rogue {1}{U}, 1/3 (Modern
// Horizons 2, #5501):
//
//	"This creature can't be blocked.
//	 Metalcraft — This creature gets +1/+0 as long as you control three
//	 or more artifacts."
//
// Slither Blade's unblockability (RestrictSelf) and Chrome Steed's
// metalcraft self pump (somB1MetalcraftPump), here +1/+0. The Infiltrator
// is not an artifact, so it never counts towards its own three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d6aa6ab-7cfc-463f-a8ba-007058933ebf",
		Name:         "Vedalken Infiltrator",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.RestrictSelf(game.CantBeBlocked),
			effects.SomB1MetalcraftPump("Metalcraft — This creature gets +1/+0 as long as you control three or more artifacts.", 1, 0),
		},
	})
}
