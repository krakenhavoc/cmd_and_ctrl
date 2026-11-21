package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inalla, Archmage Ritualist — Legendary Creature — Human Wizard
// {2}{U}{B}{R}, 4/5 (EDHREC rank 8494):
//
//	"Eminence — Whenever another nontoken Wizard you control enters, if
//	 Inalla is in the command zone or on the battlefield, you may pay
//	 {1}. If you do, create a token that's a copy of that Wizard. The
//	 token gains haste. Exile it at the beginning of the next end step.
//	 Tap five untapped Wizards you control: Target player loses 7
//	 life."
//
// The eminence line works from the command zone (#2802,
// EminenceTrigger). The payment is offered as the trigger resolves,
// and the copy is made from the Wizard that entered, read off the
// trigger's own event, so a Wizard that died in response is still
// copied from its last-known values (Miirym's shape). Haste rides the
// token's entry and the exile is a delayed trigger on that token alone
// (Saheeli Rai's shape). Being a token, the copy does not trigger
// Inalla again.
//
// The activated ability taps five Wizards the activator controls,
// Inalla among them if they like; tapping for it is not the {T}
// symbol, so Wizards that arrived this turn may pay (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21bdba6e-3f9d-4ead-8212-0cbb0ce7f8cc",
		Name:         "Inalla, Archmage Ritualist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.EminenceTrigger(effects.On(game.EventETB, effects.AnotherNontokenCreatureOfTypeEnteredUnderYourControl("Wizard"),
				"Inalla, Archmage Ritualist — you may pay {1} to create a hasty token copy of that Wizard",
				inallaMayPayToCopy)),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "Tap five untapped Wizards you control: Target player loses 7 life.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  5,
				Filter: effects.TargetPermanent("five untapped Wizards you control", effects.OfCreatureType("Wizard")),
				Label:  "five untapped Wizards you control",
			}},
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B43TargetPlayerLoses(g, item, 7)
			},
		}},
	})
}

// inallaMayPayToCopy is the eminence trigger's body: the offer, then
// the copy if it was paid.
func inallaMayPayToCopy(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Inalla, Archmage Ritualist — pay {1} to create a token that's a copy of that Wizard?",
		Then:     inallaPaid,
	}.Apply(effects.NewContext(g, item))
}

// inallaHastyCopy creates the token copy with haste and schedules its
// exile at the beginning of the next end step.
func inallaHastyCopy(ctx *effects.Context, controller, wizard uuid.UUID) error {
	tmpl, ok := effects.TokenCopyTemplate(ctx.Game, wizard)
	if !ok {
		return nil
	}
	made, err := ctx.Game.CreateTokensForEffect(controller, effects.CopiedToken(tmpl), 1, game.TokenEntryOptions{
		Keywords: []string{"haste"},
	})
	if err != nil || len(made) == 0 {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: controller,
		Label:      "Inalla, Archmage Ritualist — exile the copy",
		Cards:      made,
		Body:       effects.ExileListedCardsBody,
	}.Apply(ctx)
}

// inallaPaid is the paid half, keyed (#4264): the Wizard is the asking trigger's event.
var inallaPaid = effects.PayUnlessBranches("pay-unless/inalla-archmage-ritualist-copy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return inallaHastyCopy(ctx, ctx.Controller(), ctx.Trigger().Event.CardID)
	}, nil)
