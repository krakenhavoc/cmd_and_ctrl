package cardsco

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cradle of Vitality — Enchantment {3}{W} (Shards of Alara; Jumpstart):
//
//	"Whenever you gain life, you may pay {1}{W}. If you do, put a +1/+1
//	 counter on target creature for each 1 life you gained."
//
// Dawn of Hope's trigger (once per life-gain event, lifelink included),
// with a target picked as it goes on the stack: any creature. The
// payment is offered as it resolves, and the paid branch reads the
// amount off the triggering event, so two separate gains are two
// prompts, each for its own amount. A target that left in response
// leaves the trigger with nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "956250da-532a-4457-8696-73915be56943",
		Name:         "Cradle of Vitality",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouGainLife("Cradle of Vitality — pay {1}{W} to put a +1/+1 counter on target creature for each 1 life you gained",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := ctx.Trigger().Event.Amount
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{W}",
						Question: fmt.Sprintf("Cradle of Vitality — pay {1}{W} to put %d +1/+1 counter(s) on the target creature?", n),
						Then:     cradleOfVitalityPaid,
					}.Apply(ctx)
				}),
				effects.TargetCreature("target creature")),
		},
	})
}

// cradleOfVitalityPaid is the paid half, keyed: the asking trigger, its
// target and its event ride the prompt.
var cradleOfVitalityPaid = effects.PayUnlessBranches("pay-unless/cradle-of-vitality-counters",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		n := ctx.Trigger().Event.Amount
		if n <= 0 {
			return nil
		}
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
			}
		}
		return nil
	}, nil)
