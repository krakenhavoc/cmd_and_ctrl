package game

import "github.com/google/uuid"

// enter_as_copy_static.go — #6262: a permanent's static "creatures you
// control enter as a copy of <a permanent>" (Essence of the Wild's "of
// this creature", Infinite Reflection's "of enchanted creature").
//
// CR 614.1c: "[This permanent] enters as . . ." is a replacement effect;
// here it is another permanent's, applied to each creature entering under
// its controller. The entering permanent wears the copiable values of the
// named permanent as they are when it enters (CR 707.2), with whatever
// that permanent gained as part of its own copying (CR 707.9a): a creature
// entering as a copy of Essence of the Wild is an Essence of the Wild, with
// Essence's static of its own.
//
// It is the static sibling of Mystic Reflection's one-shot record
// (scoped_enter_as_copies.go, #6160) and lands through the same field a
// Clone's own choice does (ReplacementEvent.EntersAsCopyOf), so the entry
// path, the look-ahead (CR 614.12) and the copy event are shared. The
// catalog declares the replacement (effects.CreaturesYouControlEnterAs
// CopyOfThis and its enchanted-creature sibling) and calls the method
// below from its Replace.

// EnterAsCopyOfPermanentForEffect makes the entering permanent enter as a
// copy of the permanent `id` (CR 614.1c, CR 707.2). It reports false and
// changes nothing when the event is not an entry onto the battlefield or
// `id` is not a permanent on the battlefield.
//
// Caller must hold g.mu (write) — every caller is a replacement's Replace.
func (ev *ReplacementEvent) EnterAsCopyOfPermanentForEffect(g *Game, id uuid.UUID) bool {
	if ev == nil || g == nil || ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || id == uuid.Nil {
		return false
	}
	src, ok := g.battlefieldCardLocked(id)
	if !ok {
		return false
	}
	values := SingleFacedCopiableValuesOf(*src)
	ev.EntersAsCopyOf = &values
	ev.copySourceID = id
	ev.copyUntilEndOfTurn = false
	return true
}

// EnteringAsCreatureForEffect reports whether the permanent this entry is
// moving onto the battlefield is a creature as it would exist there
// (CR 614.12): a copy it has already been told to enter as is what it is.
//
// Caller must hold g.mu.
func (ev *ReplacementEvent) EnteringAsCreatureForEffect(g *Game) bool {
	if ev == nil || g == nil || ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.Canceled {
		return false
	}
	if v := ev.EntersAsCopyOf; v != nil {
		return v.HasCardType("Creature")
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	return ok && entering.IsCreature()
}
