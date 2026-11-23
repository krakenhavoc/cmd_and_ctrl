package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "32e4de02-8437-4ffc-acc9-a3736652d282",
		Name:         "Chief of the Scale",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Warrior"}, Others: true, YoursOnly: true}, 0, 1),
		},
	})
}
