package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honor of the Pure — Enchantment {1}{W} (Magic 2010, #5585):
//
//	"White creatures you control get +1/+1."
//
// Kaysa's colour anthem with white: a layer 7c static over the white
// creatures you control, read from their current colours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb4188bb-62df-4309-bde1-f66318fe2f05",
		Name:         "Honor of the Pure",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CreaturesOfColorYouControlGet("W", 1, 1, nil, "White creatures you control get +1/+1."),
		},
	})
}
