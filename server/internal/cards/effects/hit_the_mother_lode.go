package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hit the Mother Lode — Sorcery {4}{R}{R}{R}:
//
//	"Discover 10. If the discovered card's mana value is less than 10,
//	 create a number of tapped Treasure tokens equal to the difference."
//
// The Treasures read CR 701.57c's discovered card, which the discover's
// continuation hands over once the discoverer has answered — whether
// they cast it or put it into their hand. An empty walk discovers no
// card, so it makes no Treasures (ADR 0099, owner decision 6).
func init() {
	Register(Spec{
		OracleID:     "a49900b3-cc34-428f-806d-71861dc8059d",
		Name:         "Hit the Mother Lode",
		Completeness: CompletenessFull,
		Discovers:    true,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := item.Controller
			return Discover{N: 10, Then: func(ctx *Context, r game.DiscoverResult) error {
				if r.Discovered == uuid.Nil || r.ManaValue >= 10 {
					return nil
				}
				return b13CreateTappedTreasures(ctx, controller, 10-r.ManaValue)
			}}.Apply(ctx)
		},
	})
}
