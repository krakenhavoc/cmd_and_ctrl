package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Miasmic Mummy — Creature — Zombie Jackal {1}{B}, 2/2 (Jumpstart, #5510):
//
//	"When this creature enters, each player discards a card."
//
// Rotting Rats' trigger: every player, the Mummy's controller included,
// chooses and discards their own card (CR 701.9a); a player with an empty
// hand discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5b049d4-242a-48ec-85d1-1ef66c8cfa4c",
		Name:         "Miasmic Mummy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Miasmic Mummy — each player discards a card", func(g *game.Game, item *game.StackItem) error {
				g.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
					Source: item.SourceCardID, N: 1, Question: "Miasmic Mummy — discard a card",
				})
				return nil
			}),
		},
	})
}
