package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ramroller — Artifact Creature — Juggernaut {3}, 2/3 (Magic Origins, S82 older
// sets):
//
//	"This creature attacks each combat if able.
//	 This creature gets +2/+0 as long as you control another artifact."
//
// Barricade Breaker's attack requirement and Court Homunculus's condition
// (another artifact you control, the Ramroller itself not counted) at +2/+0,
// applied in layer 7c (CR 613.4c) and read continuously.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d670d937-aaa7-4a96-9c6a-916ff7a6b935",
		Name:         "Ramroller",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.AttacksEachCombat(),
			{
				Label:     "This creature gets +2/+0 as long as you control another artifact.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.MagitekInfantryWhileAnotherArtifact,
				Apply:     effects.PumpBy(2, 0),
			},
		},
	})
}
