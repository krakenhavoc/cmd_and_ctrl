package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devil Dinosaur — Legendary Creature — Dinosaur {2}{G}{G}, 5/5
// (Marvel Super Heroes Commander, #4455):
//
//	"Trample
//	 Other Dinosaurs you control get +1/+1 and have hexproof."
//
// Dinosaurs on a Spaceship's lord shape: the tribal anthem and a tribal
// keyword grant over OTHER Dinosaurs the controller controls (effective
// subtypes, so a changeling counts; Devil Dinosaur himself gets neither).
//
// No simplification.
func init() {
	dinos := effects.TribeFilter{Tribes: []string{"Dinosaur"}, Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        "08de884f-1930-43d6-8c23-04353a136b4e",
		Name:            "Devil Dinosaur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(dinos, 1, 1),
			effects.TribalKeywordGrant(dinos, "hexproof"),
		},
	})
}
