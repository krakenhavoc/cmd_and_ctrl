package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani's Chosen — Creature — Cat Soldier {2}{W}{W}, 3/3:
//
//	"Whenever an enchantment you control enters, create a 2/2 white Cat
//	 creature token. If that enchantment is an Aura, you may attach it to
//	 the token."
//
// Archon of Sun's Grace's constellation trigger, one per enchantment. The
// Cat is made with a keyed tail that knows the token, and the tail asks
// "you may attach it" only when the enchantment is still the object that
// entered (CR 400.7), is an Aura as it is now, and could legally enchant
// the Cat (CR 701.3b: an Aura with "enchant creature you don't control"
// stays where it is, so nothing is asked). The answer re-checks the Aura
// and the token, so one that left in the meantime is not moved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c77eb0a-543a-41db-ba1f-aa0163837ae0",
		Name:         "Ajani's Chosen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsEnchantment()
			}, "Ajani's Chosen — create a 2/2 white Cat; you may attach an Aura to it", ajanisChosenCat),
		},
	})
}

// ajanisChosenCat makes the Cat; the offer to attach rides its tail.
func ajanisChosenCat(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(TokenCard("2/2 white Cat")),
			Count:    1,
		}},
	}, ctx.ItemTail(ajanisChosenOfferAttach))
}

// ajanisChosenAura is the enchantment that entered when it is still that
// object on the battlefield and an Aura, and could enchant `token`.
func ajanisChosenAura(ctx *Context, token uuid.UUID) (uuid.UUID, bool) {
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Left || !slices.Contains(info.Characteristic.Subtypes, "Aura") {
		return uuid.Nil, false
	}
	aura := ctx.Trigger().Event.CardID
	return aura, ctx.Game.AttachLegalForEffect(aura, token)
}

// ajanisChosenOfferAttach is "if that enchantment is an Aura, you may
// attach it to the token", keyed (#4253). An on-disk key.
var ajanisChosenOfferAttach = CardTail("ajanis-chosen/offer-attach", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	token := o.IDs[0]
	if _, ok := ajanisChosenAura(ctx, token); !ok {
		return nil
	}
	return MayChoice{
		Question: "Ajani's Chosen — attach the Aura to the Cat token?",
		Then:     ajanisChosenAttach,
		Carry:    []uuid.UUID{token},
	}.Apply(ctx)
})

// ajanisChosenAttach is the answer: on a yes the Aura moves onto the Cat,
// both re-checked. An on-disk key.
var ajanisChosenAttach = MayYes("confirm/ajanis-chosen-attach", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	aura, ok := ajanisChosenAura(ctx, r.Carry[0])
	if !ok {
		return nil
	}
	return ctx.Game.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]})
})
