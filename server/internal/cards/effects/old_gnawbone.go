package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Old Gnawbone — Legendary Creature — Dragon, {5}{G}{G}, 7/7
// (EDHREC rank 924):
//
//	"Flying
//	 Whenever a creature you control deals combat damage to a player,
//	 create that many Treasure tokens."
//
// The green Treasure dragon. One trigger PER CREATURE, as printed —
// this is "a creature", not "one or more creatures", so three
// attackers connecting are three triggers and no dedup is wanted.
// The count is the damage actually dealt, read off the event after
// prevention and replacement have had their say, and captured in
// Build by value.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dff3f4c7-f792-4ca3-9b0a-0738e70664d9",
		Name:            "Old Gnawbone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				n := ev.Amount
				return game.NewTriggeredItem(source, "Old Gnawbone — create that many Treasures",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   TreasureToken(),
							N:          n,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
