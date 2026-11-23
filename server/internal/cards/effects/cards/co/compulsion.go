package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3fafb6b2-5cae-45b6-8550-3ff8daa02802",
		Name:         "Compulsion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{U}, Discard a card: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.DiscardACard()),
				Effect:  effects.EbDrawACard,
			},
			{
				Label:   "{1}{U}, Sacrifice this enchantment: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.SacrificeThis()),
				Effect:  effects.EbDrawACard,
			},
		},
	})
}
