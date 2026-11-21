package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "12172602-89db-4c8b-b752-4f8380896589",
		Name:            "Marauding Dreadship",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Marauding Dreadship — incubate 2", effects.Do(effects.Incubate{N: 2})),
		},
	})
}
