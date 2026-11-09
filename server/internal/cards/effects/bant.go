package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bant — Plane — Alara (#4870):
//
//	"All creatures have exalted. (Whenever a creature attacks alone, it gets
//	 +1/+1 until end of turn for each instance of exalted among permanents its
//	 controller controls.)
//	 Whenever chaos ensues, put a divinity counter on target green, white, or
//	 blue creature. That creature has indestructible for as long as it has a
//	 divinity counter on it."
//
// The static gives every creature one instance of exalted (a cumulative
// keyword, CR 113.2c, so each creature's controller's instances add up as the
// rulings say). The chaos ability's indestructible is the ability's own
// continuous effect (CR 611.2), not the plane's static: it lasts while the
// target has a divinity counter on it, even after Bant stops being face up
// (game.WhilePinnedHasCounter), and moves with nothing if the counters move.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82e67c57-6d7a-44cc-ba63-77ce887ab075",
		Name:         "Bant",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			Label:     "All creatures have exalted.",
			AppliesTo: allCreatures,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, game.KeywordExalted)
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Bant — put a divinity counter on target green, white, or blue creature; it has indestructible for as long as it has a divinity counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return CounterThenWhileItHasIt(ctx, id, bantDivinity, "Bant — indestructible while it has a divinity counter",
						game.AddKeywordsMod("indestructible"))
				}),
				TargetCreature("target green, white, or blue creature", Or(OfColor("G"), OfColor("W"), OfColor("U")))),
		},
	})
}

// bantDivinity is the counter's name.
const bantDivinity = "divinity"
