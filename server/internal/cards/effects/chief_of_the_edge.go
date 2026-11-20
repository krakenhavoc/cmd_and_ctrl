package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chief of the Edge — Creature — Human Warrior {W}{B}, 3/2 (Khans of
// Tarkir, #5554):
//
//	"Other Warrior creatures you control get +1/+0."
//
// Drogskol Captain's TribalAnthem over the other Warriors you control,
// read through the layered view (layer 7c), so a changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "636a9364-54a6-4f67-ad96-731646e8bde7",
		Name:         "Chief of the Edge",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Warrior"}, Others: true, YoursOnly: true}, 1, 0),
		},
	})
}
