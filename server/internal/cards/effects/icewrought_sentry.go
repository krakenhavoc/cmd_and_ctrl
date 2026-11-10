package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// ability is the "you tap" seam (Event.TappedBy, std53YouTappedOpponentCreature):
// it counts a tap the controller's own effect instructed, including this
// Sentry's own reflexive tap, and never a tap an attack, a cost or an
// opponent's instruction made (the ruling). It is a trigger of its own, so
// each such tap gives +2/+1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0828d891-f7c3-4e51-b42e-6ce88d48ac8b",
		Name:            "Icewrought Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Icewrought Sentry — you may pay {1}{U}", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}{U}",
					Question: "Icewrought Sentry — pay {1}{U} to tap target creature an opponent controls?",
					Then:     icewroughtSentryPaid,
				}.Apply(NewContext(g, item))
			}),
			On(game.EventTapCard, std53YouTappedOpponentCreature,
				"Icewrought Sentry — this creature gets +2/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !b15OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 1,
						Label: "Icewrought Sentry — +2/+1 until end of turn"}.Apply(NewContext(g, item))
				}),
		},
	})
}

var icewroughtSentryPaid = PayUnlessBranches("pay-unless/icewrought-sentry-tap",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return WhenYouDo("Icewrought Sentry — tap target creature an opponent controls",
			icewroughtSentryTapBody).Apply(ctx)
	}, nil)

var icewroughtSentryTapBody = game.ReflexiveBody("icewrought-sentry/tap",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return TapTarget{Target: id}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("target creature an opponent controls", OpponentControls())
	}))
