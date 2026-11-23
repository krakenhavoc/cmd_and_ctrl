package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Misty Knight, Hero for Hire — Legendary Creature — Human Detective Hero
// {1}{R}, 3/1 (Marvel Super Heroes, #4455):
//
//	"{2}, {T}, Discard a card: Draw a card for each card you've
//	 discarded this turn."
//
// The discard is a cost, so it is in the turn's tally by the time the
// ability resolves, and the count is read then (Change of Fortune's
// CardsDiscarded tally): the card paid for it counts, as do cards
// discarded earlier in the turn by any means, and a discard in response
// counts too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b090ba3-0ecb-4585-b5b7-9d6028c32487",
		Name:         "Misty Knight, Hero for Hire",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Discard a card: Draw a card for each card you've discarded this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: g.TurnTallyFor(item.Controller).CardsDiscarded}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
