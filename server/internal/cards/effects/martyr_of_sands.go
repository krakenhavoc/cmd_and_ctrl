package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Martyr of Sands — Creature — Human Cleric {W}, 1/1:
//
//	"{1}, Reveal X white cards from your hand, Sacrifice this creature:
//	 You gain three times X life."
//
// The reveal cost is effects.RevealX (#2598, ADR 0020's 2026-10-08
// amendment): its count is the X announced with the activation, the
// revealed cards stay in the hand, and the mana cost is {1} whatever X
// is. The life gain reads the same X off the stack item.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2048140-b691-4753-8239-1aa75059e389",
		Name:         "Martyr of Sands",
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Reveal X white cards from your hand, Sacrifice this creature: You gain three times X life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), RevealX("X white cards", "W"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return GainLife{Player: item.Controller, Amount: 3 * ctx.X()}.Apply(ctx)
			},
		}},
	})
}
