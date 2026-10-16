package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flatman — Legendary Creature — Mutant Hero {3}{G}, 1/10:
//
//	"Origami-Fu — {2}{G}: Switch Flatman's power and toughness until end
//	 of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7).
// Origami-Fu is an ability word (CR 207.2c): no rules meaning.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a7e3423-4df3-4e10-9861-ac2f403bc49d",
		Name:         "Flatman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Origami-Fu — {2}{G}: Switch Flatman's power and toughness until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{2}{G}"),
				Effect:  switchThisPT("Flatman — switch its power and toughness"),
			},
		},
	})
}
