package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Escape Protocol — Enchantment {1}{U} (Ikoria: Lair of Behemoths, #5512):
//
//	"Whenever you cycle a card, you may pay {1}. When you do, exile target
//	 artifact or creature you control, then return it to the battlefield
//	 under its owner's control."
//
// The cycle trigger is Prickly Marmoset's (a cycling ability emits EventCycle
// for its controller, CR 702.29b). The optional payment and the reflexive
// "when you do" are Weapons Vendor's (CR 603.12): the target is chosen when
// the reflexive trigger goes on the stack, after the mana is spent, and an
// opponent can respond between the payment and the blink. The blink is
// Soulherder's: the permanent returns as a new object under its owner's
// control, so it loses what was on it.
//
// No simplification.
func init() {
	const label = "Escape Protocol — pay {1} to exile target artifact or creature you control, then return it to the battlefield under its owner's control"
	effects.Register(effects.Spec{
		OracleID:     "e875b468-b475-4573-adf6-312ca4ab6994",
		Name:         "Escape Protocol",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCycle(label, func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: label + "?",
					Then:     escapeProtocolPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// escapeProtocolPaid is the paid half, keyed: "when you do" is raised only
// once the payment happened.
var escapeProtocolPaid = effects.PayUnlessBranches("pay-unless/escape-protocol-blink",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Escape Protocol — exile target artifact or creature you control, then return it to the battlefield under its owner's control",
			Body:  escapeProtocolBlinkBody,
		}.Apply(ctx)
	}, nil)

// escapeProtocolBlinkBody is "exile target artifact or creature you control,
// then return it to the battlefield under its owner's control".
var escapeProtocolBlinkBody = game.ReflexiveBody("escape-protocol/blink", effects.SimpleBody(effects.FlickerFirstLegalTargetUnderOwner),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.PermanentYouControl("target artifact or creature you control", effects.Or(effects.Artifact(), effects.Creature()))
	}))
