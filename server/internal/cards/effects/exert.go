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
//
// The row is stamped game.ExertRowLinked, so the bot knows to price an
// exert by its declared Purpose (ADR 0130's amendment of 2026-10-07).
func WhenExerted(label string, effect Effect) game.TriggeredAbility {
	t := On(game.EventExert, exertedAsItAttacked, label, effect)
	t.Exert = game.ExertRowLinked
	return t
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
//
// The row is stamped game.ExertRowPayoff, so the bot adds its declared
// Purpose to every exert it prices (ADR 0130's amendment of
// 2026-10-07).
func WheneverYouExert(label string, effect Effect) game.TriggeredAbility {
	t := On(game.EventExert, youExertedACreature, label, effect)
	t.Exert = game.ExertRowPayoff
	return t
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

// ExertedGets is the linked "When you do, it gets +P/+T [and gains
// <keywords>] until end of turn" of the self-pump exert cards (Gust
// Walker, Bitterblade Warrior, Hooded Brawler …), with the pump
// declared as the row's Purpose so the bot prices the exert by it (ADR
// 0130's amendment of 2026-10-07). Power and toughness are the printed
// numbers, and keywords are lowercase tokens.
//
// "It" is this creature as the exert made it an object: a creature
// that left the battlefield and came back before the trigger resolved
// is a new object and gets nothing (CR 400.7).
func ExertedGets(label string, power, toughness int, keywords ...string) game.TriggeredAbility {
	effect := func(g *game.Game, item *game.StackItem) error {
		var mods []game.Mod
		if power != 0 || toughness != 0 {
			mods = append(mods, game.ModifyPTMod(power, toughness))
		}
		if len(keywords) > 0 {
			mods = append(mods, game.AddKeywordsMod(keywords...))
		}
		ctx := NewContext(g, item)
		return ScopedEffectFor{
			Target:   item.SourceCardID,
			Mods:     mods,
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    label,
		}.Apply(ctx)
	}
	return TriggerWithPurpose(WhenExerted(label, effect),
		game.Purpose{Pump: &game.Pump{Power: power, Toughness: toughness, Keywords: keywords}})
}

// whenYouExertBuild is the Build of a "whenever you exert a creature"
// payoff that refers to "that creature" (Rohirrim Chargers): it records
// the exerted object, instance and epoch, in Params.Object, so the
// effect can tell whether it is still the same permanent (CR 400.7).
func whenYouExertBuild(label string) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		ref := game.ObjectRef{ID: ev.CardID}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok {
			ref.Epoch = c.ObjectEpoch
		}
		item.Params.Object = ref
		return item
	}
}
