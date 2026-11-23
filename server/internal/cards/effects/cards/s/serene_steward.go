package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serene Steward — Creature — Human Cleric Ally {1}{W}, 2/2 (Jumpstart 2022):
//
//	"Whenever you gain life, you may pay {W}. If you do, put a +1/+1 counter
//	 on target creature."
//
// Cradle of Vitality's shape with a fixed counter: the target is chosen as
// the trigger goes on the stack (CR 603.3d), the {W} is an optional cost paid
// as it resolves, once per life-gain event. A target gone by resolution gets
// no counter (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "559463f8-396f-40e0-9515-a1be6397ca31",
		Name:         "Serene Steward",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouGainLife("Serene Steward — pay {W} to put a +1/+1 counter on target creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{W}",
						Question: "Serene Steward — pay {W} to put a +1/+1 counter on the target creature?",
						Then:     sereneStewardPaid,
					}.Apply(effects.NewContext(g, item))
				}),
				effects.TargetCreature("target creature")),
		},
	})
}

// sereneStewardPaid is the paid half, keyed: the asking trigger and its
// target ride the prompt. An on-disk key: never renamed, never reused.
var sereneStewardPaid = effects.PayUnlessBranches("pay-unless/serene-steward-counter",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}
		}
		return nil
	}, nil)
