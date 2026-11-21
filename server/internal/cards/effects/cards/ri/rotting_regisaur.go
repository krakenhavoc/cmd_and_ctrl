package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rotting Regisaur — Creature — Zombie Dinosaur {2}{B}, 7/6 (Core Set 2020, #5518):
//
//	"At the beginning of your upkeep, discard a card."
//
// The discard is the controller's choice, queued as a prompt over their
// hand; an empty hand discards nothing (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acdbaa99-18c9-4aa1-8797-69837ca7d8f5",
		Name:         "Rotting Regisaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Rotting Regisaur — discard a card", func(g *game.Game, item *game.StackItem) error {
				g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: item.Controller, Source: item.SourceCardID, N: 1,
					Question: "Rotting Regisaur — discard a card"})
				return nil
			}),
		},
	})
}
