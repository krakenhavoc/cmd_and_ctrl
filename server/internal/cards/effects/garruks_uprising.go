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
// SANDBOX SIMPLIFICATION — the intervening-if (CR 603.4) is checked
// only when the trigger is put on the stack, not again on
// resolution. In paper, a board where the power-4 creature dies in
// response makes the trigger do nothing; here the draw still
// happens. That is the STRONGER direction, and it is worth being
// explicit about rather than quiet: it needs a re-check hook on the
// built StackItem that no catalog trigger has today, and it is the
// same posture every intervening-if card in the catalog takes.
// It is a corner case — it needs an opponent to respond to the
// trigger by removing the creature that enabled it — but it is a
// real divergence.
func init() {
	Register(Spec{
		OracleID:     "3127ae9b-a7a7-43ec-89d7-688f8445b33d",
		Name:         "Garruk's Uprising",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The enter-the-battlefield draw still happens even if your power-4 creature is removed in response; the condition isn't rechecked on resolution."},
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
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.CardID == source.InstanceID &&
						youControlPowerFourOrGreater(g, source.Controller)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Garruk's Uprising — draw a card",
						func(g *game.Game, item *game.StackItem) error {
							return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.CardID == source.InstanceID {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsCreature() &&
						c.Controller == source.Controller &&
						c.CurrentPower() >= 4
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Garruk's Uprising — draw a card",
						func(g *game.Game, item *game.StackItem) error {
							return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
