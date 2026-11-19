package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Insolent Neonate — Creature — Vampire {R}, 1/1 (Foundations Jumpstart,
// BNS batch 32):
//
//	"Menace
//	 Discard a card, Sacrifice this creature: Draw a card."
//
// Falkenrath Pit Fighter's discard cost with a sacrifice of the Neonate
// itself, both paid as the ability is activated: a player with no card to
// discard can't activate it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a9a1d92d-150b-48e1-b41b-dd746b7c1626",
		Name:            "Insolent Neonate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:   "Discard a card, Sacrifice this creature: Draw a card.",
			Cost:    Plus(DiscardACard(), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
