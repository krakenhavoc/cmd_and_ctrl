package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrounging Skyray — 1/2 Creature — Fish Pirate for {1}{U}:
//
//	"Flying
//	 Whenever you discard one or more cards, put that many +1/+1
//	 counters on this creature.
//	 Cycling {2}"
//
// Marauding Mako with evasion, which in a deck that loots every turn
// matters more than the extra mana: the counters actually connect.
//
// Same per-card batching as the Mako (CR 603.1), and cycling is an
// alternative cast path (S29) that isn't modelled.
func init() {
	Register(Spec{
		OracleID:        "3a46d85b-ce1a-4842-a342-92a5bddb1053",
		Name:            "Scrounging Skyray",
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Scrounging Skyray — +1/+1 counter",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return AddCounter{Target: ctx.Source(), Kind: "+1/+1", N: 1}.Apply(ctx)
					})
			},
		}},
	})
}
