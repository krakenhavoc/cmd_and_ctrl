package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rummaging Goblin — Creature — Goblin Rogue {2}{R}, 1/1 (Magic 2013, #5568):
//
//	"{T}, Discard a card: Draw a card."
//
// Goblin Picker's looter without the mana: the discard is a cost (CR
// 118.8), paid before the draw, so a player with no card to discard
// cannot activate it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2055eb91-ee36-4752-a6dc-581eaef335c8",
		Name:         "Rummaging Goblin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Discard a card: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
