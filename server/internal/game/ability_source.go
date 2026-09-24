package game

import "github.com/google/uuid"

// ability_source.go — an ability on the stack and the OBJECT it came
// from (CR 400.7), which is not the same question as the card it came
// from (#812).
//
// CR 608.2 resolves an ability whether or not its source is still
// around, and most abilities do not care: "{T}: this deals 2 damage to
// any target" deals its damage from a graveyard, reading last known
// information. A minority cannot be performed at all without the
// source as a PERMANENT — equip is the printed example, "attach this
// Equipment to target creature", and CR 301.5c says only a permanent
// on the battlefield can be attached to anything. For those, and only
// for those, "the source is gone" is an answer rather than a failure.
//
// Gone has two spellings and both are here, because both produce the
// same outcome and a primitive that checked only the first would be
// wrong in the more interesting case:
//
//  1. the source has LEFT the battlefield — sacrificed, destroyed,
//     exiled, bounced;
//  2. the source has left and COME BACK. Card.InstanceID survives a
//     zone change, so a Loxodon Warhammer bounced and replayed while
//     its equip sits on the stack is findable under the same ID and is
//     a different object with no memory of the ability (CR 400.7). The
//     object stamped on the stack item as it was put there is what
//     tells the two apart; see StackItem.SourceObject (#1418), and
//     StackItem.SourceEpoch for items older than it.

// cardObjectEpochLocked reads Card.ObjectEpoch off whichever zone
// holds the card, or -1 when no zone does. The sentinel is negative
// because 0 is a real epoch — a token created straight onto the
// battlefield has never moved — and a "card not found" that compared
// equal to a live reading would make a vanished source look present.
//
// Caller must hold g.mu (read or write).
func (g *Game) cardObjectEpochLocked(cardID uuid.UUID) int {
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return -1
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == cardID {
			return z.Cards[i].ObjectEpoch
		}
	}
	return -1
}

// AbilitySourceGoneForEffect reports whether the permanent whose
// ability `item` is has stopped being that permanent since the item
// was put on the stack: it is no longer on the battlefield, or the
// card with that ID is back on the battlefield as a NEW OBJECT
// (CR 400.7).
//
// It is the ONE read behind "source gone, so the ability does
// nothing". Use it only in an ability whose effect genuinely cannot
// happen without its source on the battlefield — attaching it
// (AttachSourceForEffect), and station's "put charge counters on THIS
// permanent" (#759, effects.Station), whose counters would otherwise
// land on a card in a graveyard or on the new object a replay made.
// An ability
// that merely mentions its source (damage, a counter, a draw) resolves
// from last known information under CR 608.2 and must NOT consult
// this.
//
// A nil item, or one with no source, reads as gone: there is no
// permanent to act on either way.
//
// The CR 400.7 half reads StackItem.SourceObject, which every ability
// item carries since #1418 — a TRIGGER as much as an activation, so
// a Batterskull flickered in response to its own living-weapon
// trigger attaches nothing, exactly as a replayed Warhammer does not
// finish its equip. The ID is the "stamped" bit, never the epoch:
// zero is a real epoch (a token created straight onto the battlefield
// has never moved), so a sentinel reading of it would quietly disable
// the check for tokens.
//
// An item with no SourceObject — restored from a snapshot written
// before #1418 — keeps the rule it was put on the stack under: an
// ACTIVATED item compares StackItem.SourceEpoch, and a trigger or a
// spell gets the plain "is it still on the battlefield" test.
//
// Caller must hold g.mu (read or write).
func (g *Game) AbilitySourceGoneForEffect(item *StackItem) bool {
	if item == nil || item.SourceCardID == uuid.Nil {
		return true
	}
	if findCardOnBattlefield(g, item.SourceCardID) < 0 {
		return true
	}
	if ref := item.SourceObject; ref.ID == item.SourceCardID {
		return g.cardObjectEpochLocked(item.SourceCardID) != ref.Epoch
	}
	if item.Kind != StackItemActivated {
		return false
	}
	return g.cardObjectEpochLocked(item.SourceCardID) != item.SourceEpoch
}

// AbilitySourceIsNewObjectForEffect reports whether the card `item`'s
// ability came from is on the battlefield right now as a DIFFERENT
// object from the one the ability names (#1432, CR 400.7): it left and
// came back while the ability waited — flickered in response, most
// often — so "put a +1/+1 counter on this creature", "sacrifice it" and
// "untap it" have nothing to act on, and must not act on the permanent
// the card has become.
//
// It is the ONE read behind every effects primitive that acts on its
// own source (effects.Context.isNewSourceObject), and it answers a
// narrower question than AbilitySourceGoneForEffect on purpose:
//
//   - It is about the BATTLEFIELD. A source that is not on the
//     battlefield reads false, and the primitive keeps its own
//     behaviour for that zone. That is the rule that keeps the card-
//     following text working: "when this dies, return it to its
//     owner's hand" names the card in the zone it went to, and "return
//     this card from your graveyard" names a card that was never a
//     permanent. Neither is a permanent that left and came back.
//   - A move this very resolution made is not "left and came back".
//     CR 400.7's exception lets the rest of an effect find the object
//     it moved: unearth returns the card and then gives IT haste, a
//     self-flicker returns it and then puts a counter on it. The
//     resolving slot's epoch (resolving_item.go) tells the two apart —
//     nothing else can move a card while an item resolves, so a source
//     whose epoch changed since the resolution began was moved by it.
//
// False for a nil item, a spell (a spell's "this" is the spell), an
// item with no source, and an item with no stamp — restored from a
// snapshot written before #1418 — which keeps today's behaviour.
//
// Caller must hold g.mu (read or write).
func (g *Game) AbilitySourceIsNewObjectForEffect(item *StackItem) bool {
	if item == nil || item.Kind == StackItemSpell || item.SourceCardID == uuid.Nil {
		return false
	}
	ref := item.SourceObject
	if ref.ID != item.SourceCardID {
		return false
	}
	c := findBattlefieldCard(g, item.SourceCardID)
	if c == nil || c.ObjectEpoch == ref.Epoch {
		return false
	}
	if start, ok := g.resolvingSourceEpochLocked(item); ok && start != c.ObjectEpoch {
		return false
	}
	return true
}

// followedSourceObjectLocked is the object `item`'s source is as
// `item`'s own resolution sees it (#1432): its stamp, unless this very
// resolution has moved the card ONTO THE BATTLEFIELD — then the
// permanent the move made, which CR 400.7 lets the rest of the effect
// find.
//
// Only a move onto the battlefield is followed, because that is the
// only case AbilitySourceIsNewObjectForEffect would otherwise misjudge.
// A source the resolution moved anywhere else keeps its stamp, so a
// "when you do" after "exile this" still reads the permanent it was
// as last-known information through SourcePermanent.
//
// The delayed and reflexive triggers a resolution creates inherit it
// (CR 603.7d, CR 603.12). Unearth is why: it is activated from the
// graveyard, so its stamp is the graveyard card, and the "exile it at
// the beginning of the next end step" it schedules is about the
// creature it has just returned. Inheriting the stamp would make the
// returned creature a stranger to its own delayed trigger, and it
// would never be exiled.
//
// Caller must hold g.mu.
func (g *Game) followedSourceObjectLocked(item *StackItem) ObjectRef {
	ref := item.SourceObject
	start, ok := g.resolvingSourceEpochLocked(item)
	if !ok || ref.ID != item.SourceCardID {
		return ref
	}
	if c := findBattlefieldCard(g, item.SourceCardID); c != nil && c.ObjectEpoch != start {
		return ObjectRef{ID: item.SourceCardID, Epoch: c.ObjectEpoch}
	}
	return ref
}

// sourceObjectRefLocked names the object card `cardID` is at this
// moment, for an ability that is being put on the stack from it
// (#1418):
//
//   - on the battlefield, the live permanent;
//   - leaving the battlefield in the event being dispatched right now
//     (its CR 603.10 snapshot is still in lastKnownBattlefield), the
//     permanent it was — the object a "when this dies" trigger is
//     about, not the graveyard card it has just become;
//   - in any other zone, the card there (a trigger from a graveyard,
//     an ability activated from a hand);
//   - a token that has ceased to exist, the last permanent it was.
//
// The zero ref when the card is nowhere and never left the
// battlefield this turn: nothing to name.
//
// Caller must hold g.mu.
func (g *Game) sourceObjectRefLocked(cardID uuid.UUID) ObjectRef {
	if cardID == uuid.Nil {
		return ObjectRef{}
	}
	if c := findBattlefieldCard(g, cardID); c != nil {
		return ObjectRef{ID: cardID, Epoch: c.ObjectEpoch}
	}
	recs := g.lastKnownPermanents[cardID]
	if _, leaving := g.lastKnownBattlefield[cardID]; leaving && len(recs) > 0 {
		return ObjectRef{ID: cardID, Epoch: recs[len(recs)-1].Epoch}
	}
	if epoch := g.cardObjectEpochLocked(cardID); epoch >= 0 {
		return ObjectRef{ID: cardID, Epoch: epoch}
	}
	if len(recs) > 0 {
		return ObjectRef{ID: cardID, Epoch: recs[len(recs)-1].Epoch}
	}
	return ObjectRef{}
}

// stampSourceObjectLocked gives `item` a SourceObject when it has
// none, naming the object its source card is right now. The fallback
// stamp for every path that builds an ability item without the
// dispatch's better knowledge; a stamp already made is never
// overwritten. Caller must hold g.mu.
func (g *Game) stampSourceObjectLocked(item *StackItem) {
	if item == nil || item.SourceObject.ID != uuid.Nil || item.Kind == StackItemSpell {
		return
	}
	item.SourceObject = g.sourceObjectRefLocked(item.SourceCardID)
}

// SourceObjectForEffect names the object `item`'s ability came from
// (#1418, CR 400.7): StackItem.SourceObject when it was stamped. Hand
// it to PermanentForEffect for "this permanent" as it is now, or as it
// last existed once that object has gone — never a new object the
// card has become since.
//
// An ability item with no stamp (restored from a snapshot written
// before #1418) falls back to the object the card most recently was,
// PermanentRefForEffect, which is what every reader had before. False
// for a spell, for an item with no source, and when the fallback
// finds nothing.
//
// Caller must hold g.mu.
func (g *Game) SourceObjectForEffect(item *StackItem) (ObjectRef, bool) {
	if item == nil || item.SourceCardID == uuid.Nil || item.Kind == StackItemSpell {
		return ObjectRef{}, false
	}
	if item.SourceObject.ID == item.SourceCardID {
		return item.SourceObject, true
	}
	return g.PermanentRefForEffect(item.SourceCardID)
}
