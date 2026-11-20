package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chief of the Scale — Creature — Human Warrior {W}{B}, 2/3 (Khans of
// Tarkir, #5554):
//
//	"Other Warrior creatures you control get +0/+1."
//
// Drogskol Captain's TribalAnthem over the other Warriors you control,
// read through the layered view (layer 7c), so a changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "32e4de02-8437-4ffc-acc9-a3736652d282",
		Name:         "Chief of the Scale",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Warrior"}, Others: true, YoursOnly: true}, 0, 1),
		},
	})
}
