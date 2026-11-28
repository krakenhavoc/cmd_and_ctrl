package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pauper's Cage — Artifact {3} (Mirage, #5633):
//
//	"At the beginning of each opponent's upkeep, if that player has two or
//	 fewer cards in hand, this artifact deals 2 damage to that player."
//
// An intervening "if" (CR 603.4): checked as the upkeep begins and again as the
// trigger resolves, against that player's hand each time.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38ec5b13-b143-4818-9243-b1715dab3bef",
		Name:         "Paupers' Cage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "that player has two or fewer cards in hand",
				Test: func(ctx *effects.Context) bool {
					return effects.CardsInHand(ctx.Game, ctx.Trigger().Event.Actor) <= 2
				},
			}, effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ByAnOpponent(ev, source, lki, g) && effects.CardsInHand(g, ev.Actor) <= 2
			}, "Paupers' Cage — 2 damage to that player", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.DealDamage{Source: item.SourceCardID, Target: ctx.Trigger().Event.Actor, Amount: 2}.Apply(ctx)
			})),
		},
	})
}
