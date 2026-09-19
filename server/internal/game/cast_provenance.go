package game

import "github.com/google/uuid"

// cast_provenance.go — S42, #653: a permanent remembers HOW its spell
// was cast (CR 400.7d).
//
// CR 400.7 is the rule that a card changing zones becomes a new object
// with no memory of the old one, and this engine leans on it hard —
// Card.ObjectEpoch, the counters and damage MoveCard wipes, the
// permission that ends because the object it named is gone. CR 400.7d
// is the one clause that reaches BACKWARDS through it: "an ability of
// a permanent can reference information about the spell that became
// that permanent as it resolved, including what costs were paid".
//
// That is not a general exception. It is a narrow, named one, and the
// cards that use it say so in the same words every time:
//
//	"When Phlage enters, sacrifice it unless it ESCAPED."   CR 702.138b
//	"When it enters THIS WAY, each opponent sacrifices…"    Pharika's Spawn
//	"This creature escapes WITH [an ability]."              CR 702.138d
//	"When this creature enters, if it was KICKED, …"        CR 702.33b
//
// The fact existed while the spell was on the stack (StackItem.AltCost,
// StackItem.Paid) and died there. fireETBHookLocked's signature is
// (cardID, oracleID) and a TriggeredAbility.Build never sees a stack
// item at all, so by the time "sacrifice it unless it escaped" ran
// there was nothing left to ask.
//
// THE SHAPE, in two decisions.
//
//  1. IT IS A VALUE ON THE PERMANENT, not a lookup. Card.NamedTribe is
//     the precedent (S26) and Card.Solved and Card.ChosenColor are the
//     same pattern: a per-entry fact written as the permanent enters,
//     cleared when it leaves, carried by the snapshot. The alternative
//     — keeping the stack item alive in a side map — would have to
//     answer "for how long", and the answer is exactly "as long as
//     this permanent", which is what a field on the permanent means.
//
//  2. IT IS STAMPED AT ONE PLACE, because there is one place. #919's
//     executeEntryToBattlefieldLocked is the finisher every battlefield
//     entry runs through, resumed or not, and stack resolution reaches
//     it with the StackItem on the event. A spell becoming a permanent
//     is that finisher plus ev.stackItem != nil, so the stamp is one
//     line next to queueAltCostEntryTriggerLocked, which is already the
//     "last moment the cost that was paid is still in hand".
//
// WHAT IT DOES NOT CARRY, and why. Not the mana (StackItem.Paid.Mana):
// nothing reads it after entry, because the clauses that care —
// sunburst, converge, "escapes with a +1/+1 counter" — all land their
// counters through the CR 614 entry pipeline while the item is still
// there. Not the targets, which are not costs. Adding a fact here
// means a card needs it AFTER the permanent has landed; #664's kicker
// will need exactly that ("if it was kicked", checked by an ETB
// trigger) and the field it wants is a sibling of AltCost on this
// struct rather than a second record.

// AltCostKeyEscape is the CR 702.138 keyword's alternative-cost key —
// the one "escaped" reads (CR 702.138b). It is a constant here because
// three layers now spell it: the catalog's Escape() constructor, the
// permission Underworld Breach grants, and the reader below.
const AltCostKeyEscape = "escape"

// CastProvenance is what a permanent remembers about the spell that
// became it (CR 400.7d). The zero value is a permanent that did not
// come from a spell at all — a token, a reanimated creature, a land —
// or one cast for its printed mana cost.
//
// Pure data. One reference-typed field, OptionalCosts, so cloneCard
// and the snapshot reallocate exactly that one slice and everything
// else rides the value copy they already make.
type CastProvenance struct {
	// AltCost is the Key of the alternative cost that paid for the
	// spell (CR 118.9) — "escape", "flashback", "evoke" — or empty
	// when the caster paid the printed mana cost.
	//
	// The same key StackItem.AltCost carried, copied rather than
	// re-derived: the catalog cannot answer for it, because a cost
	// can be GRANTED (Underworld Breach's escape, Snapcaster's
	// flashback) and the permission that granted it is usually gone.
	AltCost string `json:"altCost,omitempty"`

	// FromZone is the zone the spell was cast from — the other half
	// of CR 400.7d's "information about the spell", and the one a
	// printed clause asks about second most often ("if you cast it
	// from your graveyard"). Empty for a permanent that did not come
	// from a cast.
	FromZone ZoneKind `json:"fromZone,omitempty"`

	// OptionalCosts are the optional additional costs paid for the
	// spell (CR 601.2b, ADR 0073 §5), as positions in the card's
	// OptionalCosts slice, one entry per payment — so multikicker's
	// "the number of times it was kicked" is a length rather than a
	// count the caller has to keep.
	//
	// It was Card.PaidOptionalCosts when #664 shipped it, a week
	// before #653 shipped this record. Both were per-entry, cleared in
	// the same two places, carried by the same clone and snapshot, and
	// stamped three lines apart in the same finisher — two fields for
	// one question. #719's rebase folded the first into the second;
	// the readers (CardKickedTimes, CardPaidOptionalCost) did not
	// change shape, only where they look.
	OptionalCosts []int `json:"optionalCosts,omitempty"`
}

// Any reports whether this record says anything at all.
func (p CastProvenance) Any() bool {
	return p.AltCost != "" || p.FromZone != "" || len(p.OptionalCosts) > 0
}

// Clone deep-copies the record. The one reference-typed field is
// OptionalCosts, so this is one reallocation; everything else is a
// string the value copy already carried.
func (p CastProvenance) Clone() CastProvenance {
	if len(p.OptionalCosts) > 0 {
		p.OptionalCosts = append([]int(nil), p.OptionalCosts...)
	}
	return p
}

// Escaped reports CR 702.138b for this permanent: "escaped" means it
// was cast for its escape cost and is still the permanent that spell
// became.
//
// False for a permanent that came from anywhere else, which is the
// weaker-than-printed answer and the right one: Phlage reanimated,
// Phlage hard-cast and Phlage put onto the battlefield are all
// sacrificed, and only the escaped one stays.
func (p CastProvenance) Escaped() bool {
	return p.AltCost == AltCostKeyEscape
}

// Escaped is the same question asked of a card. On Card rather than
// only on the record so a catalog reader can say `c.Escaped()`.
func (c Card) Escaped() bool { return c.Provenance.Escaped() }

// CastProvenanceForEffect returns what the permanent with this ID
// remembers about the spell it came from, or the zero record when the
// card is not on the battlefield.
//
// The battlefield and nowhere else, deliberately: CR 400.7d is written
// about a PERMANENT, and a card that has left has already been cleared
// by MoveCard. A caller that finds nothing is looking at either a
// permanent that was not cast or one that is gone, and both answers
// are "no".
//
// Caller must hold g.mu (read or write).
func (g *Game) CastProvenanceForEffect(cardID uuid.UUID) CastProvenance {
	if g.Battlefield == nil {
		return CastProvenance{}
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			return g.Battlefield.Cards[i].Provenance
		}
	}
	return CastProvenance{}
}

// stampCastProvenanceLocked records how the spell that just became
// this permanent was cast (CR 400.7d).
//
// Called from executeEntryToBattlefieldLocked — the ONE place a spell
// becomes a permanent — after the permanent has landed and before
// EventETB fires, so the card's own entry trigger already sees it.
// That ordering is the whole requirement: "When Phlage enters,
// sacrifice it unless it escaped" is harvested off that event.
//
// A nil item is every entry that is not a resolving spell: a land
// play, a search, a reanimation, a blink, a token. Those permanents
// were not cast and their record stays zero, which is what CR 400.7d
// says about them.
//
// Caller must hold g.mu (write).
func (g *Game) stampCastProvenanceLocked(cardID uuid.UUID, item *StackItem) {
	if item == nil || g.Battlefield == nil {
		return
	}
	prov := CastProvenance{
		AltCost:  item.AltCost,
		FromZone: item.CastFromZone,
		// Its own backing array: the item's slice outlives this call
		// on the undo stack, and a permanent that shared it would see
		// a rewind edit its own record.
		OptionalCosts: append([]int(nil), item.Paid.OptionalCosts...),
	}
	if !prov.Any() {
		// Nothing to say. Left zero rather than written, so a
		// permanent hard-cast from hand is byte-identical to one that
		// was never a spell — the two answer the same to every reader
		// this record has.
		return
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			g.Battlefield.Cards[i].Provenance = prov
			return
		}
	}
}
