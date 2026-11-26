package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sibsig Icebreakers — Creature — Zombie {2}{B}, 2/3 (Dragons of Tarkir, #5551):
//
//	"When this creature enters, each player discards a card."
//
// Rotting Rats' entry trigger, each player choosing their own discard.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "4880c976-b417-48e4-98a2-9e5a8da0d758",
		Name:         "Sibsig Icebreakers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sibsig Icebreakers — each player discards a card", func(g *game.Game, item *game.StackItem) error {
				g.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
					Source: item.SourceCardID, N: 1, Question: "Sibsig Icebreakers — discard a card",
				})
				return nil
			}),
		},
	})
}
