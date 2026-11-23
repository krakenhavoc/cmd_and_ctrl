package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stromkirk Captain — Creature — Vampire Soldier {1}{B}{R}, 2/2
// (EDHREC rank 3031):
//
//	"First strike
//	 Other Vampire creatures you control get +1/+1 and have first
//	 strike."
//
// The Vampire lord. First strike rides PrintedKeywords; the two
// grants are TribalAnthem and TribalKeywordGrant over OTHER Vampires
// the controller controls (Others and YoursOnly are both printed
// here), effective subtypes so a changeling counts.
//
// No simplification.
func init() {
	vampires := effects.TribeFilter{Tribes: []string{"Vampire"}, Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        "599c3fd6-3309-4b5d-adec-9c4062848ad5",
		Name:            "Stromkirk Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(vampires, 1, 1),
			effects.TribalKeywordGrant(vampires, "first strike"),
		},
	})
}
