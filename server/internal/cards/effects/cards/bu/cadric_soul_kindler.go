package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cadric, Soul Kindler — Legendary Creature — Dwarf Wizard {2}{R}{W}, 4/3
// (Dominaria United Commander):
//
//	"The "legend rule" doesn't apply to tokens you control.
//	 Whenever another nontoken legendary permanent you control enters, you may
//	 pay {1}. If you do, create a token that's a copy of it. That token gains
//	 haste. Sacrifice it at the beginning of the next end step."
//
// The static is Spec.LegendRuleExemptions over a new "tokens you control"
// value (CR 704.5j): a legendary token of yours is never put into the
// graveyard by the legend rule, while a legendary card of yours, and anything
// an opponent controls, still is. The trigger is Arni Metalbrow's MayPay: {1}
// is paid as it resolves and "if you do" is the paid branch, which copies the
// permanent that entered, from last-known information if it has left already
// (CR 608.2h), with the haste and the end-step sacrifice that Mishra, Eminent
// One's Warform has (CR 603.7: the sacrifice is a delayed trigger naming the
// token, so a token that left does nothing). A token entering, or Cadric
// entering, does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:             "739bbf44-9169-4749-abea-21c20823915f",
		Name:                 "Cadric, Soul Kindler",
		Completeness:         effects.CompletenessFull,
		LegendRuleExemptions: effects.LegendRuleDoesntApplyToYourTokens(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, cadricAnotherNontokenLegendEntered,
				"Cadric, Soul Kindler — you may pay {1}; if you do, create a token that's a copy of it, with haste, sacrificed at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Cadric, Soul Kindler — pay {1} to create a hasty token copy of that legendary permanent?",
						Then:     cadricPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// cadricAnotherNontokenLegendEntered is the trigger's condition: another
// nontoken legendary permanent entered under the trigger's controller.
func cadricAnotherNontokenLegendEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken() && c.HasSupertype("Legendary") && c.Controller == source.Controller
}

var cadricPaid = effects.PayUnlessBranches("cadric-soul-kindler/paid", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	if ctx.Item == nil || ctx.Item.Trigger == nil {
		return nil
	}
	g := ctx.Game
	cursor := effects.B25LastEventSeq(g)
	if err := (effects.CreateTokenCopy{
		Controller: ctx.Controller(),
		Copy:       ctx.Item.Trigger.Event.CardID,
		N:          1,
	}).Apply(ctx); err != nil {
		return err
	}
	tokens := effects.B27TokensCreatedByAfter(g, ctx.Controller(), cursor)
	if len(tokens) == 0 {
		return nil
	}
	for _, token := range tokens {
		if err := (effects.ScopedEffectFor{
			Target:   token,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: effects.DurationUntilEndOfTurn(ctx),
			Label:    "Cadric, Soul Kindler — the token gains haste",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.ScheduleDelayedTrigger{
		Label: "Cadric, Soul Kindler — sacrifice the token",
		Cards: tokens,
		Body:  effects.SacrificeListedCardsBody,
	}.Apply(ctx)
}, nil)
