package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Regal Imperiosaur — Creature — Dinosaur {1}{G}{G}, 5/4 (Aetherdrift,
// #4464):
//
//	"Other Dinosaurs you control get +1/+1."
//
// Devil Dinosaur's tribal anthem without the hexproof: a layer 7c static
// (CR 613.4c) over OTHER Dinosaurs its controller controls, read off
// effective subtypes, so a changeling counts and the Imperiosaur itself
// does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83a76601-ef59-4888-af70-5088dd1f504c",
		Name:         "Regal Imperiosaur",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Dinosaur"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
