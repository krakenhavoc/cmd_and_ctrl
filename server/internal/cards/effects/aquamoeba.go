package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aquamoeba — Creature — Elemental Beast {1}{U}, 1/3:
//
//	"Discard a card: Switch this creature's power and toughness until end
//	 of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55de434c-0fb9-4127-81b6-47c1c0b088fe",
		Name:         "Aquamoeba",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Discard a card: Switch this creature's power and toughness until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    DiscardACard(),
				Effect:  switchThisPT("Aquamoeba — switch its power and toughness"),
			},
		},
	})
}
