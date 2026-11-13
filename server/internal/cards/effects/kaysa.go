package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kaysa — Legendary Creature — Elf Druid {3}{G}{G}, 2/3 (Alliances,
// #5634):
//
//	"Green creatures you control get +1/+1."
//
// A layer 7c anthem (CR 613.4c) over the green creatures its controller
// controls. It does not say "other", and Kaysa is green, so it pumps
// itself too. Built by creaturesOfColorYouControlGet (color_anthems.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0fd18a9b-c112-4e48-8fd7-e53fe4500943",
		Name:         "Kaysa",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			creaturesOfColorYouControlGet("G", 1, 1, nil, "Green creatures you control get +1/+1."),
		},
	})
}
