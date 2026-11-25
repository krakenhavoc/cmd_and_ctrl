package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hunted Wumpus — Creature — Beast {3}{G}, 6/6 (Battlebond, #5529):
//
//	When this creature enters, each other player may put a creature card from their hand onto the battlefield.
//
// Each other player in turn order is offered a creature card from their hand, as Braids puts one onto the battlefield (PutFromHandOntoBattlefield); they enter under that player's control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "698c9882-4117-4919-8a19-b1cbffc16692",
		Name:         "Hunted Wumpus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hunted Wumpus — each other player may put a creature card from their hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, who := range effects.SeatsFromController(g, item.Controller) {
						if who == item.Controller {
							continue
						}
						if err := (effects.PutFromHandOntoBattlefield{
							Player:   who,
							Match:    effects.Creature(),
							Optional: true,
							Label:    "Hunted Wumpus — you may put a creature card from your hand onto the battlefield",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
