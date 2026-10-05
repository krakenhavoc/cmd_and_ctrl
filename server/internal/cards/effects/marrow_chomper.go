package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marrow Chomper — "Devour 2. When this creature enters, you gain 2 life for each creature it devoured."
//
// The count is read when the trigger is BUILT, off the permanent's
// Card.Devoured (CR 702.82b), and carried on the item as Params.Amount: the
// trigger resolves after the creature may have left, and a closure
// must not hold the *Card. No simplifications.
func init() {
	Register(Spec{
		OracleID:     "96313465-6a97-4c0e-ae9a-91e95f247f5b",
		Name:         "Marrow Chomper",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Devour("Marrow Chomper", 2)},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Marrow Chomper — gain 2 life for each creature it devoured",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Marrow Chomper — gain 2 life for each creature it devoured")
				item.Params.Amount = source.Devoured
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: 2 * item.Params.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
