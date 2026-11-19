package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Compulsion — Enchantment {1}{U} (Enchanting Tales):
//
//	"{1}{U}, Discard a card: Draw a card.
//	 {1}{U}, Sacrifice this enchantment: Draw a card."
//
// Two activated abilities; the mana and the discard or the sacrifice are
// costs paid at announcement, and the draw is Axiom Engraver's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3fafb6b2-5cae-45b6-8550-3ff8daa02802",
		Name:         "Compulsion",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{U}, Discard a card: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}{U}"), DiscardACard()),
				Effect:  ebDrawACard,
			},
			{
				Label:   "{1}{U}, Sacrifice this enchantment: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}{U}"), SacrificeThis()),
				Effect:  ebDrawACard,
			},
		},
	})
}
