package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icewrought Sentry — Creature — Elemental Soldier {2}{U}, 2/3 (Wilds of
// Eldraine, #4471):
//
//	"Vigilance
//	 Whenever this creature attacks, you may pay {1}{U}. When you do, tap
//	 target creature an opponent controls.
//	 Whenever you tap an untapped creature an opponent controls, this
//	 creature gets +2/+1 until end of turn."
//
// The attack trigger is Itzquinth's "you may pay. When you do" shape:
// MayPay asks as the trigger resolves and only a payment queues the
// reflexive trigger (CR 603.12), which then picks its target. The last
// ability is the "you tap" seam (Event.TappedBy, YouTappedAnOpponentsCreature):
// it counts a tap the controller's own effect instructed, including this
// Sentry's own reflexive tap, and never a tap an attack, a cost or an
// opponent's instruction made (the ruling). It is a trigger of its own, so
// each such tap gives +2/+1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0828d891-f7c3-4e51-b42e-6ce88d48ac8b",
		Name:            "Icewrought Sentry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Icewrought Sentry — you may pay {1}{U}", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}{U}",
					Question: "Icewrought Sentry — pay {1}{U} to tap target creature an opponent controls?",
					Then:     icewroughtSentryPaid,
				}.Apply(effects.NewContext(g, item))
			}),
			effects.On(game.EventTapCard, effects.YouTappedAnOpponentsCreature,
				"Icewrought Sentry — this creature gets +2/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.B15OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 1,
						Label: "Icewrought Sentry — +2/+1 until end of turn"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var icewroughtSentryPaid = effects.PayUnlessBranches("pay-unless/icewrought-sentry-tap",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.WhenYouDo("Icewrought Sentry — tap target creature an opponent controls",
			icewroughtSentryTapBody).Apply(ctx)
	}, nil)

var icewroughtSentryTapBody = game.ReflexiveBody("icewrought-sentry/tap",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.TapTarget{Target: id}.Apply(ctx)
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	}))
