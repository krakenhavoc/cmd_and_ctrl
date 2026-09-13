package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temur Ascendancy — Enchantment {G}{U}{R} (EDHREC rank 886):
//
//	"Creatures you control have haste.
//	 Whenever a creature you control with power 4 or greater enters,
//	 you may draw a card."
//
// The big-creature deck's haste anthem and card engine. Haste is a
// Layer 6 grant (Garruk's Uprising's trample, with the keyword the
// summoning-sickness gate honours); the draw is Garruk's Uprising's
// second trigger with a "you may" — CurrentPower, so an anthem or
// counters that lift a 3-power creature to 4 as it lands count, as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e68dc47c-692f-4420-9799-eee104017273",
		Name:         "Temur Ascendancy",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !eotHasAbility(c.Abilities, "haste") {
					c.Abilities = append(c.Abilities, "haste")
				}
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.CurrentPower() >= 4
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Temur Ascendancy — draw a card?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Temur Ascendancy — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
