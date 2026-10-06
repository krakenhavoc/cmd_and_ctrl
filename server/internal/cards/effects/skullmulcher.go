package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skullmulcher — "Devour 1. When this creature enters, draw a card for each creature it devoured."
//
// The count is read when the trigger is BUILT, off the permanent's
// Card.Devoured (CR 702.82b), and carried on the item as Params.Amount: the
// trigger resolves after the creature may have left, and a closure
// must not hold the *Card. No simplifications.
func init() {
	Register(Spec{
		OracleID:     "66cdbd34-a864-4c39-acac-9fb26ef4adf9",
		Name:         "Skullmulcher",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{DevourPaying("Skullmulcher", 1, 1, 0)},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Skullmulcher — draw a card for each creature it devoured",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Skullmulcher — draw a card for each creature it devoured")
				item.Params.Amount = source.Devoured
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
