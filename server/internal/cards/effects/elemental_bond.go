package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elemental Bond — Enchantment {2}{G} (EDHREC rank 559):
//
//	"Whenever a creature you control with power 3 or greater enters,
//	 draw a card."
//
// Garruk's Uprising's ongoing half on its own: a card for every big
// creature. The power is read as the creature enters (CurrentPower:
// effective power plus counters, so a creature that arrives with
// +1/+1 counters counts them).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "d9a7e5a6-3e41-4fc6-987a-18fe1b9d67dd",
		Name:     "Elemental Bond",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.CurrentPower() >= 3
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Elemental Bond — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
