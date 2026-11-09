package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orochi Colony — Plane — Kamigawa (Commander Masters, #4870):
//
//	"Whenever a creature you control deals combat damage to a player, you
//	 may search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle.
//	 Whenever chaos ensues, target creature can't be blocked this turn."
//
// A plane (ADR 0216 §6). The first ability triggers for each of the planar
// controller's creatures that deals combat damage to a player (not once per
// batch); the search is optional and the land enters tapped, then the
// library is shuffled (CR 701.23). The chaos ability (CR 311.7) targets any
// creature and it can't be blocked this turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6e008fdc-48f2-40ef-95bd-80d014535ca3",
		Name:         "Orochi Colony",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Orochi Colony — you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return b07SearchBasicOntoBattlefield(g, item, item.Controller, true, true,
						"Orochi Colony — a basic land card, onto the battlefield tapped")
				}),
			Targeting(WheneverChaosEnsues("Orochi Colony — target creature can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return RestrictUntilEOT{Target: t, Restrictions: game.CantBeBlocked, Label: "Orochi Colony — can't be blocked"}.Apply(ctx)
				}), TargetCreature("target creature")),
		},
	})
}
