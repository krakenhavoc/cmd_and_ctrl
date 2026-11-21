package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quantum Riddler — Creature — Sphinx {3}{U}{U}, 4/6 (Edge of Eternities,
// #4461):
//
//	"Flying
//	 When this creature enters, draw a card.
//	 As long as you have one or fewer cards in hand, if you would draw one
//	 or more cards, you draw that many cards plus one instead.
//	 Warp {1}{U}"
//
// The replacement is the engine's whole-instruction draw seam
// (game.ReplacementEffect.DrawInstructionExtra, draw_instruction.go): it
// rewrites "draw N" to N+1 before the single draws open their windows, as
// the rulings require, and applies once however many draws follow. The
// hand is counted before any card is drawn. The enters trigger's own draw
// is one instruction too, so with a hand of one or fewer it draws two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "58582fac-4c42-4a1c-9ab1-4a892b500da4",
		Name:             "Quantum Riddler",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{1}{U}")},
		Replacements: []game.ReplacementEffect{
			effects.YouDrawThatManyPlusOneWhileHandAtMost(1, "Quantum Riddler: draw that many cards plus one instead"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Quantum Riddler — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
