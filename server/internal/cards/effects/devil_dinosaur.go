package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	dinos := TribeFilter{Tribes: []string{"Dinosaur"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:        "08de884f-1930-43d6-8c23-04353a136b4e",
		Name:            "Devil Dinosaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			TribalAnthem(dinos, 1, 1),
			TribalKeywordGrant(dinos, "hexproof"),
		},
	})
}
