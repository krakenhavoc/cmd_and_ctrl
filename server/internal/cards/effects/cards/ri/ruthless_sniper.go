package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruthless Sniper — Creature — Human Archer {B}, 1/2 (Amonkhet, S82 OLD batch 242, #5535):
//
//	"Whenever you cycle or discard a card, you may pay {1}. If you do, put a
//	 -1/-1 counter on target creature."
//
// Drake Haven's trigger with a target. Cycling is discarding the card as a cost
// (CR 702.29a), so the one discard event covers both words and a cycled card
// triggers once. The target is chosen as the ability goes on the stack; the
// {1} is asked as it resolves, and a target that has gone gets no counter.
//
// No simplification.
func init() {
	t := effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return effects.DiscardedByYou(ev, source)
	}, "Ruthless Sniper — you may pay {1}; if you do, put a -1/-1 counter on target creature",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayPay{
				Chooser:  item.Controller,
				Cost:     "{1}",
				Question: "Ruthless Sniper — pay {1} to put a -1/-1 counter on target creature?",
				Then:     ruthlessSniperPaid,
			}.Apply(effects.NewContext(g, item))
		})
	t.Targets = effects.TargetCreature("target creature")
	effects.Register(effects.Spec{
		OracleID:     "02a506bd-203f-421c-b72e-e2ba77e922aa",
		Name:         "Ruthless Sniper",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

var ruthlessSniperPaid = effects.PayUnlessBranches("pay-unless/ruthless-sniper-counter", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	target, ok := ctx.ClauseTarget(0)
	if !ok || target.Kind != game.TargetCard {
		return nil
	}
	return effects.AddCounter{Target: target.ID, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
}, nil)
