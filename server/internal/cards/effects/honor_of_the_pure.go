package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honor of the Pure — Enchantment {1}{W} (Magic 2010, #5585):
//
//	"White creatures you control get +1/+1."
//
// Kaysa's colour anthem with white: a layer 7c static over the white
// creatures you control, read from their current colours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb4188bb-62df-4309-bde1-f66318fe2f05",
		Name:         "Honor of the Pure",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			creaturesOfColorYouControlGet("W", 1, 1, nil, "White creatures you control get +1/+1."),
		},
	})
}
