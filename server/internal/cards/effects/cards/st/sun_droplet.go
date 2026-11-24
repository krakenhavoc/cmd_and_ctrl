package cardsst

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sun Droplet — Artifact {2} (Commander 2021, #5504):
//
//	"Whenever you're dealt damage, put that many charge counters on this
//	 artifact.
//	 At the beginning of each upkeep, you may remove a charge counter from
//	 this artifact. If you do, you gain 1 life."
//
// "Dealt damage" is damage, not life loss, and "that many" is the damage
// actually applied (Darien, King of Kjeldor's reading): one trigger per
// damage event, with the amount fixed as it fires. The upkeep ability
// triggers for every player's upkeep, is asked of the Droplet's controller
// as it resolves, and the life comes only if a counter really came off
// this same object; with no counter there is nothing to ask.
//
// No simplification.
func init() {
	dealt := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			_, ok := effects.B43YouWereDealtDamage(ev, source, g)
			return ok
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			n, _ := effects.B43YouWereDealtDamage(ev, source, g)
			return game.NewTriggeredItem(source, "Sun Droplet — put "+strconv.Itoa(n)+" charge counters on it")
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "1820af5c-9cc2-4b77-b4ca-86084442f087",
		Name:         "Sun Droplet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			dealt,
			effects.AtEachUpkeep("Sun Droplet — you may remove a charge counter; if you do, you gain 1 life", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.SourceIsStillThisPermanent(g, item) || sunDropletCharge(g, item) == 0 {
					return nil
				}
				return effects.MayChoice{
					Player:   item.Controller,
					Question: "Sun Droplet — remove a charge counter to gain 1 life?",
					YesLabel: "Remove it",
					NoLabel:  "Keep it",
					Then:     sunDropletRemove,
				}.Apply(ctx)
			}),
		},
	})
}

// sunDropletCharge is the charge counters on this artifact as it is now.
func sunDropletCharge(g *game.Game, item *game.StackItem) int {
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return 0
	}
	return c.Counters[game.CounterCharge]
}

// sunDropletRemove is the "yes": remove a counter, and the life only if it
// came off. Keyed.
var sunDropletRemove = effects.MayYes("confirm/sun-droplet-remove", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) || sunDropletCharge(ctx.Game, ctx.Item) == 0 {
		return nil
	}
	if err := ctx.Game.AddCounterForEffect(ctx.Item.SourceCardID, game.CounterCharge, -1); err != nil {
		return err
	}
	return effects.GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
})
