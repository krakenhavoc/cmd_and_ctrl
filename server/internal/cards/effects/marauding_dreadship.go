package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marauding Dreadship — Artifact — Vehicle {2}{R}, 4/1 (March of the
// Machine):
//
//	"Haste
//	 When this Vehicle enters, incubate 2.
//	 Crew 2"
//
// Crew comes from the printed card data like every Vehicle's; haste is
// a printed keyword that matters once it is crewed. The enters trigger
// makes the Incubator (CR 701.53a, ADR 0185).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "12172602-89db-4c8b-b752-4f8380896589",
		Name:            "Marauding Dreadship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Marauding Dreadship — incubate 2", Do(Incubate{N: 2})),
		},
	})
}
