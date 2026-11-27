package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seed the Land — Enchantment {2}{G}{G} (Saviors of Kamigawa, #5601):
//
//	"Whenever a land enters, its controller creates a 1/1 green Snake
//	 creature token."
//
// Any land entering, whoever's, triggers it, Seed the Land's controller's
// lands and opponents' alike; the Snake goes to the player who controls that
// land as it entered (Zo-Zu's capture of the controller, read as the trigger is
// built, so a land that has since left or changed hands still credits the right
// player). Several lands entering at once trigger it once each.
//
// No simplification.
func init() {
	const label = "Seed the Land — its controller creates a 1/1 green Snake creature token"
	effects.Register(effects.Spec{
		OracleID:     "75b39a67-6720-482f-935a-c142ca0f7891",
		Name:         "Seed the Land",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.B39LandEnteredController(ev, g)
				return ok
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Player, _ = effects.B39LandEnteredController(ev, g)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Params.Player == uuid.Nil {
					return nil
				}
				return effects.CreateToken{Controller: item.Params.Player, Template: effects.TokenCard("1/1 green Snake"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
