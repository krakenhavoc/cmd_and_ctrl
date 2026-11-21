package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// Same per-card batching as the Mako. Cycling arrived with
// #660: an activated ability from hand (CR 702.29a), not a cast.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3a46d85b-ce1a-4842-a342-92a5bddb1053",
		Name:            "Scrounging Skyray",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Scrounging Skyray — +1/+1 counter", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.AddCounter{Target: ctx.Source(), Kind: "+1/+1", N: 1}.Apply(ctx)
			}), game.Purpose{DiscardPayoff: &game.DiscardPayoff{Any: true, Counters: 1}}),
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
