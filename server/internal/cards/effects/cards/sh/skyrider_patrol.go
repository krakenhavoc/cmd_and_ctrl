package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyrider Patrol — Creature — Elf Scout {2}{G}{U}, 2/3 (Core Set 2019, S82 OLD batch 199, #5527):
//
//	"Flying
//	 At the beginning of combat on your turn, you may pay {G}{U}. When you do,
//	 put a +1/+1 counter on another target creature you control, and that
//	 creature gains flying until end of turn."
//
// Spined Tyrranax's may-pay at the beginning of your combat with a reflexive
// trigger on payment (CR 603.12), whose target is another creature you
// control, chosen as the reflexive trigger goes on the stack. It puts the
// counter on, then grants flying until end of turn, on the first target still
// legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "104386d9-93b4-4a18-86d5-68718b474f4a",
		Name:            "Skyrider Patrol",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat(skyriderPatrolLabel, skyriderPatrolOffer),
		},
	})
}

const skyriderPatrolLabel = "Skyrider Patrol — you may pay {G}{U}; when you do, +1/+1 counter and flying on another target creature you control"

func skyriderPatrolOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{G}{U}",
		Question: skyriderPatrolLabel + "?",
		Then:     skyriderPatrolPaid,
	}.Apply(effects.NewContext(g, item))
}

var skyriderPatrolPaid = effects.PayUnlessBranches("pay-unless/s82-old-b199-skyrider-patrol",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Skyrider Patrol — +1/+1 counter and flying on another target creature you control",
			Body:  skyriderPatrolBody,
		}.Apply(ctx)
	}, nil)

var skyriderPatrolBody = game.ReflexiveBody("s82-old-b199/skyrider-patrol-counter-and-flying",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Skyrider Patrol — flying until end of turn"}.Apply(ctx)
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))
	}))
