package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exert.go — the catalog side of exert (CR 701.43, ADR 0130).
//
// A card declares "You may exert this creature as it attacks" on
// Spec.ExertOnAttack and its linked "when you do" trigger in Triggered:
//
//	ExertOnAttack: ExertAsItAttacks(),
//	Triggered: []game.TriggeredAbility{
//	    WhenExerted("Oketra's Avenger — prevent all combat damage that would be dealt to it this turn", effect),
//	},
//
// The engine offers the choice with the attack, pays it as the
// declaration locks in, and emits game.EventExert; the triggers below
// watch that event. Append-only: shared exert helpers go here.

// ExertAsItAttacks is the static ability "You may exert this creature
// as it attacks" (CR 701.43d).
func ExertAsItAttacks() *game.ExertOnAttack {
	return &game.ExertOnAttack{}
}

// ExertAsItAttacksUnless is ExertAsItAttacks with a condition under
// which the creature may NOT be exerted: Combat Celebrant's "If this
// creature hasn't been exerted this turn, you may exert it as it
// attacks" is ExertAsItAttacksUnless(ExertedThisTurn).
func ExertAsItAttacksUnless(unless func(g *game.Game, source *game.Card) bool) *game.ExertOnAttack {
	return &game.ExertOnAttack{Unless: unless}
}

// ExertedThisTurn is "this creature has been exerted this turn", per
// object (CR 400.7): the Unless of Combat Celebrant.
func ExertedThisTurn(g *game.Game, source *game.Card) bool {
	return source != nil && g.ExertedThisTurn(source.InstanceID)
}

// WhenExerted is a "When you do" triggered ability linked to the card's
// "You may exert this creature as it attacks" (CR 701.43d, 607.2h): it
// triggers when THIS permanent is exerted as it attacks. An EventExert
// with a Target is an exert as it attacked, which is the only action
// the linked ability can refer to; an exert to pay a cost has none.
//
// A targeted one (Glorybringer) takes Targeting as usual; with no legal
// target it is removed from the stack and the creature stays exerted
// (CR 603.3d, the Glorybringer ruling).
func WhenExerted(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventExert, exertedAsItAttacked, label, effect)
}

// exertedAsItAttacked is WhenExerted's condition.
func exertedAsItAttacked(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.CardID == source.InstanceID && ev.Target != uuid.Nil
}

// WheneverYouExert is "Whenever you exert a creature" (Resolute
// Survivors, Battlefield Scavenger): any exert by the source's
// controller of a permanent that was a creature as it was exerted — the
// source itself, another creature, or a creature exerted to pay a cost
// (the Amonkhet ruling). Arena of Glory, a land exerted for mana, does
// not count.
//
// The harvester reads the event as it is emitted, while the exerted
// permanent is still on the battlefield, so "was a creature as it was
// exerted" is read off it now.
func WheneverYouExert(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventExert, youExertedACreature, label, effect)
}

// youExertedACreature is WheneverYouExert's condition.
func youExertedACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}

// untapAllOtherCreaturesYouControl is "untap all other creatures you
// control" (Combat Celebrant, Ahn-Crop Champion). "Other" is the
// ability's source object; if that permanent has left and come back it
// is a new object (CR 400.7), and is untapped like any other creature.
func untapAllOtherCreaturesYouControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	self := item.SourceCardID
	if sourceIsNewObject(g, item) {
		self = uuid.Nil
	}
	return b16UntapAllYouControlMatching(ctx, item.Controller, func(c game.Card) bool {
		return c.IsCreature() && c.InstanceID != self
	})
}
