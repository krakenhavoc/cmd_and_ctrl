package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Cellarspawn — Enchantment Creature — Horror {1}{B}{B}, 3/3
// (Duskmourn Commander, #5368):
//
//	"Each spell you cast that's a Demon, Horror, or Nightmare costs {1}
//	 less to cast.
//	 Whenever you cast a spell, if the amount of mana spent to cast it was
//	 less than its mana value, target opponent loses life equal to the
//	 difference."
//
// The discount is one reducer over the three creature types (CR 601.2f,
// generic mana only). The trigger is Tokka & Rahzar's comparison — the
// spell's payment record against its mana value on the stack, X included
// — with the difference taken as the spell is cast and carried on the
// item (the spell may be gone by resolution). The intervening "if"
// (CR 603.4) is the shortfall being above zero when it triggers; the
// number cannot change afterwards (CR 601.2h). The Cellarspawn's own
// discount makes a Horror it discounts trigger it, as printed. Like
// Roiling Vortex it declares ReadsYourCasts so a waived payment asks what
// was paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "b8510b34-0e5f-4639-9db1-1ba40d84b800",
		Name:           "Ancient Cellarspawn",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Each spell you cast that's a Demon, Horror, or Nightmare costs {1} less to cast.",
				YourSpell(), cmdr2B05SpellOfAnyType("Demon", "Horror", "Nightmare")),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && ev.Actor == source.Controller && cmdr2S50ManaSpentShortfall(g, ev.CardID) > 0
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Key:     "Ancient Cellarspawn — target opponent loses life equal to the difference",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Ancient Cellarspawn — target opponent loses life equal to the difference")
				item.Params.Amount = cmdr2S50ManaSpentShortfall(g, ev.CardID)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer && item.Params.Amount > 0 {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -item.Params.Amount)
					}
				}
				return nil
			},
		}},
	})
}
