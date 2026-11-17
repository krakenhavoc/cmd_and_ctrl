package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gibbering Barricade — Creature — Nightmare Wall {2}{B}, 2/4 (Dominaria
// United):
//
//	"Defender
//	 {2}{B}, Sacrifice a creature: You gain 1 life and draw a card."
//
// A sacrifice outlet (Viscera Seer's cost; it may sacrifice itself). The
// life and the card are one effect, in that order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c6106021-08ae-49c8-b32c-fdc7397522e7",
		Name:            "Gibbering Barricade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice a creature: You gain 1 life and draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{2}{B}"), SacrificeACreature()),
			Effect:  Do(GainLife{Amount: 1}, DrawCards{N: 1}),
		}},
	})
}
