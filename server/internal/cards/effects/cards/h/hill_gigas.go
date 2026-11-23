package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hill Gigas — Creature — Giant {4}{R}{R}, 5/4 (Final Fantasy, #4462):
//
//	"Trample, haste
//	 Mountaincycling {2} ({2}, Discard this card: Search your library
//	 for a Mountain card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// Mongoose Lizard's Mountaincycling (typecycling, CR 702.29e) on two
// printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "137594f7-2422-461b-a3ae-7ce3c135e0c7",
		Name:            "Hill Gigas",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Mountaincycling", "{2}", "a Mountain card", effects.IsLandWithSubtype("Mountain")),
		},
	})
}
