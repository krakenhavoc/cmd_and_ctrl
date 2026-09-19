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
//     epoch stamped on the stack item at announce is what tells the two
//     apart; see StackItem.SourceEpoch.

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
// (AttachSourceForEffect) is the whole of that set today. An ability
// that merely mentions its source (damage, a counter, a draw) resolves
// from last known information under CR 608.2 and must NOT consult
// this.
//
// A nil item, or one with no source, reads as gone: there is no
// permanent to act on either way.
//
// The CR 400.7 half is asked only of an ACTIVATED item, because those
// are exactly the items whose announce paths stamp
// StackItem.SourceEpoch — both of them, with no third. Gating on the
// kind rather than on "is the epoch zero" matters: zero is a real
// epoch (a token created straight onto the battlefield has never
// moved), so a sentinel reading of it would quietly disable the check
// for tokens and enable a wrong one for everything else. A trigger or
// a spell gets the plain "is it still on the battlefield" test, which
// is what every caller had before this existed.
//
// Caller must hold g.mu (read or write).
func (g *Game) AbilitySourceGoneForEffect(item *StackItem) bool {
	if item == nil || item.SourceCardID == uuid.Nil {
		return true
	}
	if findCardOnBattlefield(g, item.SourceCardID) < 0 {
		return true
	}
	if item.Kind != StackItemActivated {
		return false
	}
	return g.cardObjectEpochLocked(item.SourceCardID) != item.SourceEpoch
}
