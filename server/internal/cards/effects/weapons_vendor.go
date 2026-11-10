package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weapons Vendor — Creature — Human Artificer {3}{W}, 2/2 (Final
// Fantasy, #4462):
//
//	"When this creature enters, draw a card.
//	 At the beginning of combat on your turn, if you control an
//	 Equipment, you may pay {1}. When you do, attach target Equipment you
//	 control to target creature you control."
//
// Matt Murdock's pay-then-reflexive pair (CR 603.12): the two targets
// are chosen when the reflexive trigger goes on the stack, after the {1}
// is spent, as the ruling says. The "if you control an Equipment" is an
// intervening "if" (CR 603.4), checked as the trigger fires and again
// before the question. Attaching moves the Equipment, whoever it was
// on, to the creature (CR 701.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd0a7a86-a567-4145-b1d8-e0b72091db84",
		Name:         "Weapons Vendor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Weapons Vendor — draw a card", Do(DrawCards{N: 1})),
			InterveningIf(CondBoard{
				Name: "you control an Equipment",
				Test: func(ctx *Context) bool {
					return countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype("Equipment") }) > 0
				},
			}, AtBeginningOfYourCombat(weaponsVendorLabel, weaponsVendorOffer)),
		},
	})
}

const weaponsVendorLabel = "Weapons Vendor — pay {1} to attach target Equipment you control to target creature you control"

func weaponsVendorOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: weaponsVendorLabel + "?",
		Then:     weaponsVendorPaid,
	}.Apply(NewContext(g, item))
}

var weaponsVendorPaid = PayUnlessBranches("pay-unless/weapons-vendor-attach",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "Weapons Vendor — attach target Equipment you control to target creature you control",
			Body:  weaponsVendorAttachBody,
		}.Apply(ctx)
	}, nil)

// weaponsVendorAttachBody is "when you do, attach target Equipment you
// control to target creature you control": slot 0 the Equipment, slot 1
// the creature; either gone or illegal and nothing is attached.
var weaponsVendorAttachBody = game.ReflexiveBody("weapons-vendor/attach", simpleBody(weaponsVendorAttach),
	constTargets(func() *game.TargetSpec {
		return PermanentYouControl("target Equipment you control", HasSubtype("Equipment")).
			Then(PermanentYouControl("target creature you control", Creature()))
	}))

func weaponsVendorAttach(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	equip, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	creature, ok := ctx.ClauseTarget(1)
	if !ok {
		return nil
	}
	return g.AttachForEffect(equip.ID, game.TargetRef{Kind: game.TargetCard, ID: creature.ID})
}
