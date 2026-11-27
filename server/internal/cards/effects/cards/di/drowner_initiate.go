package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drowner Initiate — {U} Creature — Merfolk Wizard 1/1 (Shadowmoor, S82 older sets):
//
//	"Whenever a player casts a blue spell, you may pay {1}. If you do, target
//	 player mills two cards."
//
// Kraken's Eye's "a player casts a blue spell" trigger (any player's spell,
// judged on its colour as it is cast) with Ruthless Sniper's pay-and-target
// shape: the target player is chosen as the trigger goes on the stack, the {1} is
// offered as it resolves, and the mill happens only if it was paid.
//
// No simplification.
func init() {
	t := effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfColor("U")),
		"Drowner Initiate — you may pay {1}; if you do, target player mills two cards",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayPay{
				Chooser:  item.Controller,
				Cost:     "{1}",
				Question: "Drowner Initiate — pay {1} to have target player mill two cards?",
				Then:     drownerInitiatePaid,
			}.Apply(effects.NewContext(g, item))
		})
	t.Targets = effects.TargetPlayer("target player")
	effects.Register(effects.Spec{
		OracleID:     "19a4010f-f5bf-48be-a1d7-eacc3bbeeb59",
		Name:         "Drowner Initiate",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

var drownerInitiatePaid = effects.PayUnlessBranches("pay-unless/drowner-initiate-mill", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	target, ok := ctx.ClauseTarget(0)
	if !ok || target.Kind != game.TargetPlayer {
		return nil
	}
	return effects.MillCards{Player: target.ID, N: 2}.Apply(ctx)
}, nil)
