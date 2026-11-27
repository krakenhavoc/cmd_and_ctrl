package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spined Sliver — Creature — Sliver {R}{G}, 2/2 (Stronghold, S82 OLD b750,
// #5627):
//
//	"Whenever a Sliver becomes blocked, that Sliver gets +1/+1 until end of turn
//	 for each creature blocking it."
//
// Berserk Murlodont's trigger over every Sliver (anyone's, itself included): the
// event names the attacker as it becomes blocked, and the pump counts the
// creatures blocking it now, as the ability resolves (CR 608.2h). The Sliver is
// read as the object that was blocked, so one that left and came back gets
// nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eed9c168-1c78-43c6-a814-9b80054d1373",
		Name:         "Spined Sliver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, spinedSliverASliverBecameBlocked,
				"Spined Sliver — that Sliver gets +1/+1 until end of turn for each creature blocking it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					obj := ctx.Trigger().Object
					if obj == nil {
						return nil
					}
					if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
						return nil
					}
					id := obj.Ref().ID
					n := g.CreaturesBlockingForEffect(id)
					return effects.BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Spined Sliver — +1/+1 for each creature blocking it"}.Apply(ctx)
				}),
		},
	})
}

// spinedSliverASliverBecameBlocked is "whenever a Sliver becomes blocked": the
// attacker the event names is a Sliver creature as it becomes blocked.
func spinedSliverASliverBecameBlocked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesBlocked {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && effects.OfCreatureType("Sliver")(g, source.Controller, game.ThroughTextOf(source, c))
}
