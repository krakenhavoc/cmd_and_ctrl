package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "21bdba6e-3f9d-4ead-8212-0cbb0ce7f8cc",
		Name:         "Inalla, Archmage Ritualist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			EminenceTrigger(On(game.EventETB, AnotherNontokenCreatureOfTypeEnteredUnderYourControl("Wizard"),
				"Inalla, Archmage Ritualist — you may pay {1} to create a hasty token copy of that Wizard",
				inallaMayPayToCopy)),
		},
		Activated: []ActivatedAbility{{
			Label: "Tap five untapped Wizards you control: Target player loses 7 life.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  5,
				Filter: TargetPermanent("five untapped Wizards you control", OfCreatureType("Wizard")),
				Label:  "five untapped Wizards you control",
			}},
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b43TargetPlayerLoses(g, item, 7)
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
	wizard, controller := item.Trigger.Event.CardID, item.Controller
	return MayPay{
		Chooser:  controller,
		Cost:     "{1}",
		Question: "Inalla, Archmage Ritualist — pay {1} to create a token that's a copy of that Wizard?",
		OnPay: func(ctx *Context) error {
			return inallaHastyCopy(ctx, controller, wizard)
		},
	}.Apply(NewContext(g, item))
}

// inallaHastyCopy creates the token copy with haste and schedules its
// exile at the beginning of the next end step.
func inallaHastyCopy(ctx *Context, controller, wizard uuid.UUID) error {
	tmpl, ok := TokenCopyTemplate(ctx.Game, wizard)
	if !ok {
		return nil
	}
	made, err := ctx.Game.CreateTokensForEffect(controller, tmpl, 1, game.TokenEntryOptions{
		Keywords: []string{"haste"},
	})
	if err != nil || len(made) == 0 {
		return err
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: controller,
		Label:      "Inalla, Archmage Ritualist — exile the copy",
		Cards:      made,
		Body:       exileListedCardsBody,
	}.Apply(ctx)
}
