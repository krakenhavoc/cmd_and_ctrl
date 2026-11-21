package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragonborn Looter — Creature — Dragon Rogue {1}{U}, 1/2 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"{1}, {T}: Draw a card, then discard a card."
//
// Bag of Holding's loot (B16DrawThenDiscard), paid with {1} and tapping this
// creature, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f74e53e1-e575-4386-a128-5c87124987d4",
		Name:         "Dragonborn Looter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 1)
			},
		}},
	})
}
