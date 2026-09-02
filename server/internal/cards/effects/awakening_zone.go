package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Awakening Zone — Enchantment for {2}{G}:
//
//	"At the beginning of your upkeep, create a 0/1 colorless Eldrazi
//	Spawn creature token. It has 'Sacrifice this creature: Add {C}.'"
//
// S19 sub-PR 5: a pure token-generating "your upkeep" trigger. The
// token's sac-for-mana activated ability is deferred (cost model
// follow-up); here it's a vanilla 0/1. Mandatory; the token arrives
// when the trigger resolves.
func init() {
	Register(Spec{
		OracleID: "f955bc96-d602-4142-a9a2-87009cc7028c",
		Name:     "Awakening Zone",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Awakening Zone — create an Eldrazi Spawn",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   EldraziSpawnToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
