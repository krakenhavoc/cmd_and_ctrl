package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Attunement — {2}{U} Enchantment:
//
//	"Return this enchantment to its owner's hand: Draw three cards, then
//	 discard four cards."
//
// #2028: the return is the cost (ReturnThis), paid at announce
// (CR 602.2b), so Attunement is in its owner's hand before the draw and
// may be one of the four discards. The discard prompt opens after the
// three draws (drawThenDiscard).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ea46ccf-5e36-4889-bf64-00d4f327e9b2",
		Name:         "Attunement",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Return this enchantment to its owner's hand: Draw three cards, then discard four cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ReturnThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 3, 4, "")
			},
		}},
	})
}
