package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shield_families_recipient.go — ADR 0108 Delivery PR 7 (#1904): the
// recipient and "dealt to and dealt by" families of the not-one-use
// shield (PreventDamageFromSource, game.ModPreventFromSource).
//
// Append-only, mechanic-named. Writing a card:
//
//	PreventDamageFromSource{Protect: ShieldTheTarget}
//	  — Indestructible Aura's "Prevent all damage that would be dealt
//	    to target creature this turn"
//	PreventDamageFromSource{Protect: ShieldTheTargetPermanents}
//	  — Redeem's "to up to two target creatures", one record for both
//	PreventDamageFromSource{Protect: ShieldYouAndPermanentsYouControl}
//	  — Endure's "to you and permanents you control"
//	toAndByShield(ShieldTheTarget, true)
//	  — Maze of Ith's "all combat damage that would be dealt to and
//	    dealt by that creature this turn", one record (Mod.AndDealtBy)

// ShieldTheTargetPermanents protects every permanent the item targets,
// in one record: "up to two target creatures" (Redeem), "those
// creatures" (Energy Arc). Only the not-one-use shield reads it.
var ShieldTheTargetPermanents = ShieldTarget{kind: shieldTheTargetPermanents}

// ShieldObjects protects the named permanents, in one record ("those
// permanents", Mutational Advantage). Only the not-one-use shield reads
// it.
func ShieldObjects(ids ...uuid.UUID) ShieldTarget {
	return ShieldTarget{kind: shieldObjects, ids: append([]uuid.UUID(nil), ids...)}
}

// ShieldYouAndYourPermanents is "to you and <type> permanents you
// control": "you and planeswalkers you control" (Take the Bait).
func ShieldYouAndYourPermanents(types ...string) ShieldTarget {
	return ShieldTarget{kind: shieldYouAndYourPermanentsOf, permTypes: append([]string(nil), types...)}
}

// ShieldYouAndPermanentsYouControl is "to you and permanents you
// control" (Endure): every permanent has one of these card types
// (CR 110.4).
var ShieldYouAndPermanentsYouControl = ShieldYouAndYourPermanents(
	"artifact", "battle", "creature", "enchantment", "land", "planeswalker")

// protectMany is the permanents a several-permanent ShieldTarget names
// as the shield is made: the item's legal permanent targets, or the
// named objects still on the battlefield as the objects they were. A
// target gone by resolution is not protected (CR 608.2b).
func (t ShieldTarget) protectMany(ctx *Context) []uuid.UUID {
	var ids []uuid.UUID
	switch t.kind {
	case shieldTheTargetPermanents:
		for _, tr := range ctx.LegalTargets() {
			if tr.Kind == game.TargetCard {
				ids = append(ids, tr.ID)
			}
		}
	case shieldObjects:
		ids = t.ids
	}
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil || ctx.isNewSourceObject(id) || !onBattlefield(ctx.Game, id) || slices.Contains(out, id) {
			continue
		}
		out = append(out, id)
	}
	return out
}

// toAndByShield is "Prevent all [combat] damage that would be dealt to
// and dealt by <protect> this turn": one record whose protected
// permanents are also its sources (game.Mod.AndDealtBy).
func toAndByShield(protect ShieldTarget, combatOnly bool) PreventDamageFromSource {
	return PreventDamageFromSource{Protect: protect, AndDealtBy: true, CombatOnly: combatOnly}
}

// untapAttackerToAndByRow is the Maze of Ith family's ability: "<cost>:
// Untap target attacking creature [<filter>]. Prevent all combat damage
// that would be dealt to and dealt by that creature this turn." A target
// gone by resolution leaves the ability nothing to do (CR 608.2b).
func untapAttackerToAndByRow(label string, cost game.AbilityCost, targets *game.TargetSpec) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Cost:    cost,
		Targets: targets,
		Effect: func(g *game.Game, item *game.StackItem) error {
			return untapThenToAndBy(NewContext(g, item))
		},
	}
}

// untapThenToAndBy untaps the item's first legal creature target and
// shields it, to and by, from combat damage this turn.
func untapThenToAndBy(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return toAndByShield(ShieldObject(t.ID), true).Apply(ctx)
	}
	return nil
}

// toAndByThisRow is "<cost>: Prevent all combat damage that would be
// dealt to and dealt by this creature this turn" (Moonlight Geist,
// Urborg Phantom, Deftblade Elite). A source that has left and come back
// is a new object, and is not shielded (CR 400.7).
func toAndByThisRow(label string, cost game.AbilityCost) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Cost:    cost,
		Purpose: preventsDamageUnlessTargeted(nil),
		Effect: func(g *game.Game, item *game.StackItem) error {
			return shieldThis(g, item, toAndByShield(ShieldTarget{}, true))
		},
	}
}

// untilYourNextTurnToAndByRow is "<cost>: Until your next turn, prevent
// all damage that would be dealt to and dealt by target permanent an
// opponent controls" (Dovin, Hand of Control; Kiora, the Crashing Wave):
// one to-and-by record, all damage, until your next turn (CR 611.2b).
func untilYourNextTurnToAndByRow(label string, cost game.AbilityCost) ActivatedAbility {
	shield := toAndByShield(ShieldTheTarget, false)
	shield.UntilYourNextTurn = true
	return sourceShieldRow(label, cost, TargetPermanent("target permanent an opponent controls", OpponentControls()), shield)
}

// shieldThis is `shield` pinned to the ability's own source, "this
// creature" (Trained Pronghorn, Favored Hoplite). A source that has left
// and come back is a new object, and is not shielded (CR 400.7).
func shieldThis(g *game.Game, item *game.StackItem, shield PreventDamageFromSource) error {
	if sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	shield.Protect = ShieldObject(item.SourceCardID)
	return shield.Apply(NewContext(g, item))
}

// shieldTargetRow is "<cost>: Prevent all damage that would be dealt to
// target <X> this turn" (Kitsune Healer, Oriss, Godtoucher).
func shieldTargetRow(label string, cost game.AbilityCost, targets *game.TargetSpec) ActivatedAbility {
	return sourceShieldRow(label, cost, targets, PreventDamageFromSource{Protect: ShieldTheTarget})
}

// shieldTargetSpell is an instant's OnResolve whose whole effect is
// "Prevent all damage that would be dealt to target creature this turn"
// (Indestructible Aura, Shielded Passage).
func shieldTargetSpell() func(*game.StackItem, *Context) error {
	return sourceShieldSpell(PreventDamageFromSource{Protect: ShieldTheTarget})
}

// thenShieldTheTarget runs `first` on the item's first legal creature
// target, then shields that creature from all damage this turn: "Untap
// target creature. Prevent all damage that would be dealt to it this
// turn" (Djeru's Resolve), "Target creature gains flying until end of
// turn. Prevent all damage that would be dealt to that creature this
// turn" (Leap of Faith). A target gone by resolution leaves nothing to
// do (CR 608.2b).
func thenShieldTheTarget(ctx *Context, combatOnly bool, first func(id uuid.UUID) error) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if first != nil {
			if err := first(t.ID); err != nil {
				return err
			}
		}
		return PreventDamageFromSource{Protect: ShieldObject(t.ID), CombatOnly: combatOnly}.Apply(ctx)
	}
	return nil
}

// activePlayerIDOf is whose turn it is, read without taking the game's
// lock (a cast condition, a target predicate and an effect all run
// under it). uuid.Nil before the game has a turn.
func activePlayerIDOf(g *game.Game) uuid.UUID {
	if g == nil || g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) || g.Seats[g.Turn.ActiveSeat] == nil {
		return uuid.Nil
	}
	return g.Seats[g.Turn.ActiveSeat].ID
}

// heroic is "Heroic — Whenever you cast a spell that targets this
// creature, <effect>" (CR 207.2c ability word): the spell's targets are
// chosen before it becomes cast (CR 601.2c), so they are on its stack
// item as EventCast is emitted. A spell with this creature as two of
// its targets triggers it once; a copy is not cast (CR 707.10) and an
// ability is not a spell, so neither triggers it.
func heroic(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Actor != source.Controller {
			return false
		}
		item := g.StackItemForEffect(ev.CardID)
		if item == nil || item.Kind != game.StackItemSpell {
			return false
		}
		for _, t := range item.Targets {
			if t.Kind == game.TargetCard && t.ID == source.InstanceID {
				return true
			}
		}
		return false
	}, label, effect)
}

// "At the beginning of the next end step, return those cards to the
// battlefield tapped under their owners' control" (Morningtide's Light):
// flicker's return body, with the cards entering tapped, together.
var returnExiledToOwnersTappedBody = game.SimpleDelayedBody("flicker/return-exiled-to-owners-tapped", returnExiledCardsToOwnersTapped)

func returnExiledCardsToOwnersTapped(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			ids = append(ids, t.ID)
		}
	}
	return ReturnFromExileTogether{Targets: ids, Tapped: true}.Apply(NewContext(g, item))
}
