package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b580_helpers.go — shared pieces of Ravnica: City of Guilds batch 580
// (#5600).

// OldB580PreventDamageFromThisToRedCreatures is Indentured Oaf's "Prevent all
// damage that this creature would deal to red creatures" (CR 615.1a): a
// prevention replacement that applies to damage whose source is the Oaf and whose
// recipient is a red creature on the battlefield, read through the layers as the
// damage would be dealt. Damage to a player, a planeswalker or a creature of
// another color is untouched.
func OldB580PreventDamageFromThisToRedCreatures(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource != src.InstanceID {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.DamageTarget)
			return ok && g.Battlefield.Contains(ev.DamageTarget) && c.IsCreature() &&
				OfColor("R")(g, src.Controller, game.ThroughTextOf(src, c))
		},
		Replace:    preventThatDamage,
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      label,
		RowLabel:   "Prevent all damage that this creature would deal to red creatures.",
	}
}

// OldB580EachOpponentLosesLifeForEachCreatureTheyControl is Netherborn
// Phalanx's "each opponent loses 1 life for each creature they control": every
// other player, in turn order, loses as many life as they control creatures as
// the ability resolves (read once per opponent, through the layers).
func OldB580EachOpponentLosesLifeForEachCreatureTheyControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range TablePlayers(ctx) {
		if id == item.Controller {
			continue
		}
		n := len(PermanentsControlledByMatching(g, id, Creature()))
		if n <= 0 {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, id, -n); err != nil {
			return err
		}
	}
	return nil
}

// OldB580CounterTheSpellUnlessDecline is Perplex's "or else": its controller
// declined to discard their hand, so the spell the prompt carries (Carry[0], its
// stack ID) is countered if it is still on the stack. Keyed: on disk, never
// renamed.
var OldB580CounterTheSpellUnlessDecline = PayUnlessElse("pay-unless/old-b580-perplex-counter",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 || ctx.Game.StackItemForEffect(r.Carry[0]) == nil {
			return nil
		}
		return CounterTarget{StackID: r.Carry[0]}.Apply(ctx)
	})

// OldB580CounterUnlessItsControllerDiscardsTheirHand is Perplex's resolution:
// the target spell's controller is asked whether to discard their hand, and a
// "no" counters it. The payment is "discard N cards" with N the hand as it stands
// when asked (the payer names every card), and with an empty hand discarding it
// is no payment at all (CR 118.12a), so the spell stands. A spell that has left
// the stack asks nothing.
func OldB580CounterUnlessItsControllerDiscardsTheirHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		stacked := g.StackItemForEffect(t.ID)
		if stacked == nil {
			return nil
		}
		payer := stacked.Controller
		p := g.PlayerByIDForEffect(payer)
		if p == nil {
			return nil
		}
		hand := 0
		if p.Hand != nil {
			hand = p.Hand.Size()
		}
		pay := NoPayment()
		if hand > 0 {
			pay = DiscardPayment(hand)
		}
		_, action := pay.Prompt()
		words := "Discard your hand"
		var a *game.PayAction
		if action != nil {
			c := *action
			a = &c
		}
		return UpkeepPayUnless{
			Chooser:  payer,
			Cost:     words,
			Action:   a,
			Question: "Perplex — discard your hand, or the spell is countered?",
			Else:     OldB580CounterTheSpellUnlessDecline,
			Carry:    []uuid.UUID{t.ID},
		}.Apply(ctx)
	}
	return nil
}

// OpponentsCantCastSpellsNamedLikeTheCardsExiledWithThis is "Your opponents can't
// cast spells with the same name as a card exiled with <this>" (Circu, Dimir
// Lobotomist): Ashiok's Erasure's static, read live off the cards exiled with the
// permanent (CR 607.2a), under the label this card prints.
func OpponentsCantCastSpellsNamedLikeTheCardsExiledWithThis() game.CastRestriction {
	r := OpponentsCantCastSpellsNamedLikeTheExiledCard()
	r.Label = "Your opponents can't cast spells with the same name as a card exiled with this creature."
	return r
}
