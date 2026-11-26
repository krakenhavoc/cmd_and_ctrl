package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abzan Beastmaster — {2}{G} Creature — Dog Shaman 2/1 (Fate Reforged):
//
//	"At the beginning of your upkeep, draw a card if you control the creature
//	 with the greatest toughness or tied for the greatest toughness."
//
// Not an intervening "if" (the condition ends the sentence): the trigger goes
// on the stack every upkeep and the board is read as it resolves, on the
// layered view (CR 613).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aab463bb-e790-403b-a88d-06ef81c7302b",
		Name:         "Abzan Beastmaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Abzan Beastmaster — draw a card if you control the creature with the greatest toughness or tied",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.S82OldB338ControlsGreatestToughness(g, item.Controller) {
						return nil
					}
					return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
