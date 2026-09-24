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
// WHAT IT DOES NOT CARRY, and why. ~~Not the mana (StackItem.Paid.Mana):
// nothing reads it after entry, because the clauses that care —
// sunburst, converge, "escapes with a +1/+1 counter" — all land their
// counters through the CR 614 entry pipeline while the item is still
// there.~~ Not the targets, which are not costs. Adding a fact here
// means a card needs it AFTER the permanent has landed; #664's kicker
// will need exactly that ("if it was kicked", checked by an ETB
// trigger) and the field it wants is a sibling of AltCost on this
// struct rather than a second record.
//
// THE MANA IS NOW CARRIED (#1212). The paragraph above was right about
// the clauses that existed when it was written and wrong about the
// family that is larger than all of them: a printed ETB trigger that
// reads the payment.
//
//	"When this creature enters, if mana from a Treasure was
//	 spent to cast it, you draw a card…"      Hired Hexblade
//	"…if {R} was spent to cast it, it gains haste…"  Gruul Scrapper
//	"…sacrifice it unless {U} was spent to cast it." Azorius Herald
//
// An entry REPLACEMENT can read StackItem.Paid, because the item is
// still on the event (entry_counters.go, #1002). An entry TRIGGER
// cannot: by the time it resolves the spell has finished resolving and
// the item is gone, which is the same wall "sacrifice it unless it
// escaped" hit and the same answer — the fact rides the permanent. It
// is the exact sibling of AltCost the paragraph above predicted, and
// it is the same tokens rather than a second record.

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

	// GiftOpponent is PaidCost.GiftOpponent carried across the entry
	// (CR 400.7d, ADR 0089 §2): the opponent a gift was promised to,
	// or uuid.Nil. What a gift PERMANENT's "when this enters, if the
	// gift was promised" reads, and who its gift trigger gives to.
	GiftOpponent uuid.UUID `json:"giftOpponent,omitempty"`

	// Mana is the tokens that paid for the spell, copied off
	// StackItem.Paid.Mana at the entry finisher (#1212) — each still
	// carrying the colour it was and the SourceKinds snapshot of the
	// permanent that made it (mana_source.go).
	//
	// The same tokens, not a summary: a summary would be a second
	// shape of the one record, and the questions a card asks about a
	// payment are already written down once, on ManaSpent. Read it
	// through Spent() and never by ranging this slice.
	Mana []ManaToken `json:"mana,omitempty"`

	// X is StackItem.XValue at the moment the spell became this
	// permanent (CR 107.3m, #1312): "if an object's enters-the-
	// battlefield triggered ability … refers to X, and the spell that
	// became that object … had a value of X chosen for any of its
	// costs, the value of X for that ability is the same as the value
	// of X for that spell." An entry REPLACEMENT can read
	// StackItem.XValue directly (EntryCountersFromCast / CastCounts.X
	// already do, for "enters with X counters"); an entry TRIGGER
	// cannot, because the item is gone by the time it resolves — the
	// same wall AltCost and Mana hit, and the same answer: it rides
	// the permanent.
	//
	// Zero for a permanent that came from a spell with no {X} in its
	// cost, and a real zero for one announced at X=0 (CR 107.3) — the
	// same documented ambiguity CastCounts.X already carries, and
	// harmless for the same reason: every printed clause that reads X
	// computes the same answer (draw zero, deal zero) whether X was
	// truly zero or absent.
	//
	// Placed before the bool below rather than after it: an int here
	// and a bool there is the layout TestCardHasNoInteriorPadding
	// wants, and the reverse order leaves 7 bytes of alignment padding
	// between them.
	X int `json:"x,omitempty"`

	// ManaOnPaper is PaidCost.OnPaper carried across the entry: the
	// engine WAIVED the charge (permissive mode, a strict-mode
	// ForceCast) and has no record of what was paid.
	//
	// It has to come too, or the entry-side readers lose the one
	// distinction ADR 0068 §3 was written to preserve. Without it a
	// waived payment and a genuinely free cast are the same empty
	// slice, and "if mana from a Treasure was spent" would answer the
	// same to both — which is fine here (both answer no, the weaker
	// direction) and wrong the moment an entry clause reads "if NO
	// mana was spent to cast it".
	ManaOnPaper bool `json:"manaOnPaper,omitempty"`
}

// Spent is what this permanent's cast paid, as the same view a
// resolving spell reads off its own stack item (CR 400.7d).
//
// ONE vocabulary for the question, two homes for the fact. A card
// that asks ctx.ManaSpent().FromTreasure() during resolution and one
// that asks it from an enters trigger are asking the same thing and
// get the same answer shape.
func (p CastProvenance) Spent() ManaSpent {
	return ManaSpent{tokens: p.Mana, unknown: p.ManaOnPaper}
}

// Any reports whether this record says anything at all.
func (p CastProvenance) Any() bool {
	return p.AltCost != "" || p.FromZone != "" || len(p.OptionalCosts) > 0 ||
		len(p.Mana) > 0 || p.ManaOnPaper || p.GiftOpponent != uuid.Nil || p.X != 0
}

// Clone deep-copies the record. Two reference-typed fields now —
// OptionalCosts and the token slice, whose per-token Restrictions get
// their own backing array for the reason clonePaidCost gives: an undo
// snapshot that aliased the live array would let a restore mutate the
// game it came from.
func (p CastProvenance) Clone() CastProvenance {
	if len(p.OptionalCosts) > 0 {
		p.OptionalCosts = append([]int(nil), p.OptionalCosts...)
	}
	if len(p.Mana) > 0 {
		mana := make([]ManaToken, len(p.Mana))
		for i, t := range p.Mana {
			mana[i] = t
			if len(t.Restrictions) > 0 {
				mana[i].Restrictions = append([]string(nil), t.Restrictions...)
			}
		}
		p.Mana = mana
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

// ManaSpentToCast is what paid for the spell that became this
// permanent (CR 400.7d), as the same view a resolving spell reads off
// its own stack item.
//
// On Card for the reason Escaped is: an intervening-if (CR 603.4) is
// handed the permanent and nothing else — "when this creature enters,
// IF mana from a Treasure was spent to cast it" is checked in a
// trigger's AppliesTo, where `source *Card` is all there is.
//
// The zero view for a permanent that was not cast, which answers every
// question the weaker way.
func (c Card) ManaSpentToCast() ManaSpent { return c.Provenance.Spent() }

// CastX is CR 107.3m's X for the spell that became this permanent —
// "put X +1/+1 counters on him. Then draw half X cards" (Wan Shi
// Tong, Librarian, #1312). On Card for the same reason Escaped and
// ManaSpentToCast are: a triggered ability's Build closure is handed
// `source *Card` and nothing else, and the effect it returns should
// capture the plain int this returns rather than the card itself.
//
// Zero for a permanent that was not cast, or cast with no {X} in its
// cost — see CastProvenance.X for the one documented ambiguity that
// leaves unclaimed.
func (c Card) CastX() int { return c.Provenance.X }

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
		// ADR 0089: and who the gift was promised to, the same
		// announcement fact one field over.
		GiftOpponent: item.Paid.GiftOpponent,
		// #1212: the mana, for the same reason and with the same
		// copy. Taken off the item here rather than looked up later,
		// because this is the last moment it exists — the item is
		// discarded the instant resolution finishes.
		ManaOnPaper: item.Paid.OnPaper,
		// #1312 (CR 107.3m): X, for the same reason and at the same
		// last moment.
		X: item.XValue,
	}
	if len(item.Paid.Mana) > 0 {
		prov.Mana = make([]ManaToken, len(item.Paid.Mana))
		for i, t := range item.Paid.Mana {
			prov.Mana[i] = t
			if len(t.Restrictions) > 0 {
				prov.Mana[i].Restrictions = append([]string(nil), t.Restrictions...)
			}
		}
	}
	if !prov.Any() {
		// Nothing to say. Left zero rather than written, so a
		// permanent whose cast recorded nothing at all is
		// byte-identical to one that was never a spell — the two
		// answer the same to every reader this record has.
		//
		// #1212 narrowed what reaches here: a cast that spent any
		// mana, or that was waived, now says so, and only a genuinely
		// free cast from hand (cascade, "without paying its mana
		// cost", a {0} alternative cost) still falls through. That is
		// the right narrowing — "no mana was spent to cast it" is a
		// fact a permanent may yet be asked about, and the zero
		// record is what answers it.
		return
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			g.Battlefield.Cards[i].Provenance = prov
			return
		}
	}
}
