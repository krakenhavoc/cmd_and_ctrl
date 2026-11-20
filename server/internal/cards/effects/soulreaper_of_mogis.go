package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soulreaper of Mogis — Enchantment Creature — Minotaur Shaman {2}{B}, 2/3 (Theros Beyond Death, #5514):
//
//	"{2}{B}, Sacrifice a creature: Draw a card."
//
// Grim Backwoods' sacrifice outlet without the tap: any creature its
// controller controls, itself included, pays the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "508ce3a9-9050-4494-9adf-7f9cd6cc2f9f",
		Name:         "Soulreaper of Mogis",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice a creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{2}{B}"), SacrificeACreature()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
