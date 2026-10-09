package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Garruk's Uprising — Enchantment {2}{G}:
//
//	"When this enchantment enters, if you control a creature with
//	 power 4 or greater, draw a card.
//	 Creatures you control have trample.
//	 Whenever a creature you control with power 4 or greater enters,
//	 draw a card."
//
// Three clauses, three existing mechanisms: an ETB trigger with an
// intervening-if, a Layer 6 keyword grant, and an ongoing ETB
// watcher. Green's stompy decks play it as a repeatable cantrip that
// also turns off chump blocking.
//
// The two triggers are separate abilities, not one — the ETB clause
// fires exactly once, on the Uprising's own entry, and the ongoing
// clause explicitly excludes the source so an Uprising that is
// somehow a creature could not double-dip on its own arrival.
//
// The ETB clause is an intervening if (CR 603.4): it is checked when
// the trigger would go on the stack and again as it resolves
// (drawIfYouControlPowerFourOrGreater), so a power-4 creature removed
// in response leaves no card. The ongoing clause is not an intervening
// if: its power test is what the entering creature triggers on.
func init() {
	Register(Spec{
		OracleID:     "3127ae9b-a7a7-43ec-89d7-688f8445b33d",
		Name:         "Garruk's Uprising",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, k := range c.Abilities {
					if k == "trample" {
						return
					}
				}
				c.Abilities = append(c.Abilities, "trample")
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID &&
					youControlPowerFourOrGreater(g, source.Controller)
			}, "Garruk's Uprising — draw a card", drawIfYouControlPowerFourOrGreater),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() &&
					c.Controller == source.Controller &&
					c.CurrentPower() >= 4
			}, "Garruk's Uprising — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
