package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra, Fire Artisan — Legendary Planeswalker — Chandra {2}{R}{R}, loyalty 4:
//
//	"Whenever one or more loyalty counters are removed from Chandra, she deals
//	 that much damage to target opponent or planeswalker.
//	 +1: Exile the top card of your library. You may play it this turn.
//	 −7: Exile the top seven cards of your library. You may play them this
//	 turn."
//
// The trigger is one instance per removal EVENT (not per counter), and "that
// much" is the number the event took off, read from the event log so it is
// still right once Chandra has left with her last counter (CR 608.2h). A
// loyalty cost, damage that removes loyalty and any other removal all count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2867d539-6a63-4fae-aae2-a7fdd4dd9f28",
		Name:            "Chandra, Fire Artisan",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCounterPlaced, chandraFireArtisanLoyaltyRemoved,
				"Chandra, Fire Artisan — damage equal to the loyalty counters removed to target opponent or planeswalker",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ev := ctx.Trigger().Event
					n := g.CounterCountBeforeForEffect(ev) - ev.Amount
					for _, t := range ctx.LegalTargets() {
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
					}
					return nil
				}), effects.TargetOpponentOrPlaneswalker()),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Exile the top card of your library. You may play it this turn.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label: "−7: Exile the top seven cards of your library. You may play them this turn.",
				Cost:  effects.LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 7}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

func chandraFireArtisanLoyaltyRemoved(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventCounterPlaced && source != nil && ev.Target == source.InstanceID &&
		ev.Label == game.CounterLoyalty && ev.Amount < g.CounterCountBeforeForEffect(ev)
}
