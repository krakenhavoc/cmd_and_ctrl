package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ivory Tower — Artifact {1}:
//
//	"At the beginning of your upkeep, you gain X life, where X is the number
//	 of cards in your hand minus 4."
//
// X is read as the trigger resolves; a hand of four or fewer gains nothing
// (a negative X is never a life loss).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d416a4ed-9f16-4a8b-8aee-03c630e1eb6c",
		Name:         "Ivory Tower",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Ivory Tower — you gain X life, where X is the number of cards in your hand minus 4", func(g *game.Game, item *game.StackItem) error {
				p := g.PlayerByIDForEffect(item.Controller)
				if p == nil || p.Hand == nil {
					return nil
				}
				x := len(p.Hand.Cards) - 4
				if x <= 0 {
					return nil
				}
				return effects.GainLife{Player: item.Controller, Amount: x}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
