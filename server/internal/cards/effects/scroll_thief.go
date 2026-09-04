package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scroll Thief — 1/3 Creature — Merfolk Rogue for {2}{U}:
//
//	"Whenever Scroll Thief deals combat damage to a player, draw a
//	card."
//
// S19 sub-PR 7: the self-only shape — the source of the damage must
// be the Thief itself (ev.Source == source.InstanceID), not just any
// creature its controller has. Mandatory, no prompt.
func init() {
	Register(Spec{
		OracleID: "637c5583-4683-4ae4-8b4e-f5da42a772c7",
		Name:     "Scroll Thief",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Scroll Thief — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
