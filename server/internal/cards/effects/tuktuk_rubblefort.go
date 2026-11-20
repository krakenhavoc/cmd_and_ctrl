package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tuktuk Rubblefort — Creature — Wall {2}{R}, 0/3 (Zendikar Rising,
// #5508):
//
//	"Defender, reach
//	 Creatures you control have haste."
//
// Defender and reach are printed keywords; the grant is Urabrask the
// Hidden's layer 6 haste over every creature its controller controls,
// itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "48dab180-4ffd-4b39-9f5a-6bdc017a1eae",
		Name:            "Tuktuk Rubblefort",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "haste"),
		},
	})
}
