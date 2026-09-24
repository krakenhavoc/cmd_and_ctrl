package game

import "github.com/google/uuid"

// paid_cost.go — #789 / #761: ONE record of what an announcement
// actually paid.
//
// Two unrelated-looking asks turned out to be the same one. A
// variable counter cost ("Remove any number of storage counters:
// Add {C} for each storage counter removed this way", Mage-Ring
// Network) has to tell the effect how many came off; converge,
// sunburst and adamant have to tell a resolving spell which mana
// paid for it. Both are the question "what did this announcement
// cost?", asked a moment after the cost was paid, and both were
// unanswerable because the payment threw its own facts away.
//
// So there is one answer, and it is DATA rather than a closure: a
// PaidCost hangs off the thing the payment announced. For a spell
// and an activated ability that is the stack item (StackItem.Paid,
// classified `carried` in snapshot_drift_test.go). For a mana
// ability there is no stack item — CR 605.3b — so the record lives
// for the length of the activation and is handed to the one
// callback that needs it (ManaAbilityShape.ProducedForPaid).
//
// What it is NOT: a log. It records what the engine charged, not
// the story of how. A payment that was waived (permissive mode) says
// so with OnPaper rather than by leaving the record empty, because
// "empty" and "nothing was paid" are different facts and CR 118.3's
// readers — "if no mana was spent to cast it" — need to tell them
// apart.

// PaidCost is what one announcement paid: the mana that left the
// pool, the counters that came off or went on, the life that was
// paid. Zero value means "nothing was paid, and that is known" — a
// free cast, a copy, a cost with no components.
//
// Every field is a FACT about the payment, never a re-derivation: an
// effect that reads CountersRemoved is reading the number the engine
// took, not a number it could recompute from the board (the counters
// are gone by then, which is the whole point).
type PaidCost struct {
	// Mana is the tokens that actually left the payer's pool, in the
	// order the solver spent them, each still carrying the Source
	// that produced it and the Restrictions it was minted with
	// (#761). Nil when no mana component was paid.
	//
	// Readers: converge (CR 702.86) counts distinct colours,
	// sunburst (CR 702.44) counts them at entry, adamant counts one
	// colour, and "if no mana was spent" asks whether this is empty
	// — but only together with OnPaper, which is the difference
	// between "nothing" and "we did not charge you".
	Mana []ManaToken

	// OnPaper marks a payment the engine did NOT take: permissive
	// mode (the human default, client/src/lib/settings.ts) and the
	// strict-mode ForceCast override both let the cast through with
	// the pool untouched and an EventCostWarning in the log. The
	// player paid on paper; the engine has no record of WHAT.
	//
	// It exists so "no mana was spent to cast it" is never ambiguous.
	// Without it an unrecorded payment and a genuinely free cast look
	// identical, and every reader of the clause would silently fire
	// on half the casts at a permissive table. With it, the rule is
	// one line and it is the #259 direction: an OnPaper record
	// answers "unknown", and every reader treats unknown as the
	// weaker-than-printed answer — see NoManaSpent and ColorsSpent.
	OnPaper bool

	// CountersRemoved is how many counters a RemoveCounters
	// component actually took off, across every permanent it took
	// them from (#789). The number the ACTIVATOR announced for a
	// variable cost, and the printed number for a fixed one.
	CountersRemoved int

	// CountersAdded is how many counters an AddCounter component put
	// on (Devoted Druid's -1/-1). Almost always 1.
	CountersAdded int

	// LifePaid is the life a Life component cost, and the life a
	// Phyrexian symbol was paid with (CR 107.4). Zero for a cost
	// with neither.
	LifePaid int

	// Sacrificed is how many permanents a sacrifice component
	// actually took — the source when the cost sacrificed it, and the
	// permanents the clause named (#1213).
	//
	// It exists for the VARIABLE count: Radiant Lotus's "three mana of
	// the chosen color for each artifact sacrificed this way" is read
	// at resolution, by which time the artifacts are in graveyards and
	// nothing on the board could count them. Exactly the argument
	// CountersRemoved was added under (#789), which is why it is the
	// neighbouring field of the same record rather than a second
	// mechanism.
	//
	// Recorded for a FIXED cost too, where it is simply the printed
	// number. "The engine charged N" and "the card prints N" are
	// different facts, and a record that only spoke up for the
	// interesting case would make every reader ask which it was
	// looking at.
	Sacrificed int

	// ReturnedAttacking is what the permanent a ReturnToHand component
	// returned was ATTACKING when it was returned — the player,
	// planeswalker or battle, in Card.AttackingTarget's own overloaded
	// domain. uuid.Nil when the cost returned nothing, or returned a
	// permanent that was not attacking, which is every printed return
	// cost but one.
	//
	// The one is ninjutsu (CR 702.49a, #1227): "Return an unblocked
	// attacker you control to hand: Put this card onto the battlefield
	// from your hand tapped and attacking" — attacking THE SAME player
	// or planeswalker the returned creature was attacking. Nothing at
	// resolution could recompute that: the exit clears
	// Card.AttackingTarget (zone.go) and LKI carries no combat state,
	// so the fact has to be read BEFORE the bounce and carried here.
	// Exactly the argument CountersRemoved (#789) and Sacrificed
	// (#1213) were added under, which is why it is the neighbouring
	// field of the same record rather than a second mechanism.
	//
	// For a clause that returns more than one permanent — none is
	// printed — this is the FIRST returned permanent that was
	// attacking, in the order the activator named them. Read through
	// Context.ReturnedAttacking().
	ReturnedAttacking uuid.UUID

	// Exiled is the cards an ExileCards component exiled (#1297), in
	// the order the activator named them — "the card exiled this way"
	// (Holistic Wisdom: "if it shares a card type with the card exiled
	// this way"; Dread Defiler: "the exiled card's power"). Nil when
	// the cost exiled nothing, which is almost every payment.
	//
	// A LIST of instance IDs rather than a count, because every reader
	// printed so far asks about the card itself, and the card is still
	// findable: it is in exile, by the same instance ID, because the
	// cost put it there. A count would be Sacrificed's shape answering
	// a question nobody on this component asks.
	Exiled []uuid.UUID `json:"exiled,omitempty"`

	// OptionalCosts is which of the card's optional additional costs
	// the caster chose to pay (CR 601.2b), as positions in the card's
	// OptionalCosts slice, ascending. A cost paid N times appears N
	// times, which is how multikicker records its count (CR 702.33d)
	// without a second field. Nil for an unkicked cast — and for a
	// card that offers nothing, which is nearly every card.
	//
	// "Was it kicked" is a fact about the PAYMENT, so it belongs in
	// this record rather than beside AltCost: it is the same kind of
	// thing as "which mana paid" and it is unrecomputable for the
	// same reason — by resolution the mana is gone, the sacrificed
	// creature is in a graveyard, and the catalog cannot say whether
	// a choice was taken.
	//
	// A COPY of a kicked spell IS kicked (CR 707.10 copies the
	// choices made when it was cast), which is the one place this
	// field differs from Mana above: CopySpellForEffect carries it
	// and clears the rest. Added in ADR 0073 (#664).
	OptionalCosts []int `json:"optionalCosts,omitempty"`

	// GiftOpponent is the opponent the caster chose while paying a
	// gift cost (CR 702.174a) — the player the gift was PROMISED to —
	// or uuid.Nil when no gift was promised. Non-nil is exactly CR
	// 702.174k's "that spell's gift was promised", read through
	// GiftPromised().
	//
	// A fact about the announcement for the reason OptionalCosts is:
	// "choose an opponent" is how the gift cost is paid, and by
	// resolution nothing on the board records the choice. It is the
	// same kind of fact as "was it kicked" and travels the same way —
	// a CR 707.10 copy keeps it, and CR 400.7d carries it onto the
	// permanent as CastProvenance.GiftOpponent. ADR 0089 §2.
	GiftOpponent uuid.UUID `json:"giftOpponent,omitempty"`

	// TappedOthers is what a TapOthers component tapped (#759): one
	// entry per permanent the activator named, in the order named.
	// Nil for every cost without the component, which is nearly all
	// of them.
	//
	// It exists for station (CR 702.184a), whose effect puts "charge
	// counters equal to the tapped creature's power" on the source —
	// the first printed effect that reads a fact about a permanent
	// tapped to PAY for it. Exactly the argument Sacrificed (#1213)
	// and ReturnedAttacking (#1227) were added under, which is why it
	// is a field of this record and not a second mechanism. Read
	// through PaidTapPowerForEffect, never off Power directly — see
	// PaidTap for why the number here is a fallback and not the
	// answer. ADR 0071, addendum 2026-09-23.
	TappedOthers []PaidTap `json:"tappedOthers,omitempty"`
}

// PaidTap is one permanent a TapOthers cost tapped, as the payment
// record keeps it (#759).
//
// WHICH power a station ability reads is CR 608.2h's question, not
// the payment's: the effect "uses the current information of that
// object if it's in the public zone it was expected to be in", and
// its last-known information if it is not. The Edge of Eternities
// release notes say the same thing of station in so many words — the
// power is read AS THE ABILITY RESOLVES, and a creature that left
// the battlefield is read as it last existed there. So a creature
// pumped in response puts more counters on, and one shrunk in
// response puts fewer.
//
// That makes this a record of IDENTITY first and a number second:
//
//   - ID and Epoch name the object (CR 400.7). A creature that
//     leaves and comes back under the same instance ID is a new
//     object, and the epoch is what says so.
//   - Power is the power the object had when it was tapped, and it
//     is rewritten ONCE, to the power it had as it left the
//     battlefield, by the exit choke point (battlefieldExitLocked).
//     Left marks that the rewrite happened, and from then on Power
//     is the object's last-known power and final.
//
// The engine's own CR 603.10 snapshot (lastKnownBattlefield) cannot
// serve here: it lives for one mutation and is gone by the time the
// ability resolves, so the record keeps its own.
type PaidTap struct {
	ID    uuid.UUID `json:"id"`
	Epoch int       `json:"epoch,omitempty"`
	// Power is PowerForComparison — layers AND +1/+1 / -1/-1
	// counters, NOT clamped at zero, because a negative power is a
	// real answer (the release notes: "no charge counters are put
	// onto or removed from" the permanent) and clamping belongs to
	// the reader.
	Power int  `json:"power"`
	Left  bool `json:"left,omitempty"`
}

// PaidOptionalCost reports whether the optional cost at `index` was
// paid at least once.
func (p PaidCost) PaidOptionalCost(index int) bool {
	return p.OptionalCostTimes(index) > 0
}

// OptionalCostTimes is how many times the optional cost at `index`
// was paid — 0 when it was declined, and the multikicker count when
// it was taken more than once (CR 702.33d).
func (p PaidCost) OptionalCostTimes(index int) int {
	n := 0
	for _, i := range p.OptionalCosts {
		if i == index {
			n++
		}
	}
	return n
}

// ManaTokens is the tokens that paid, or nil. The accessor rather
// than the field so a caller cannot append into the record.
//
// Was ManaSpent() until #1212 gave that name to the VIEW below; a
// caller that wants the tokens still gets them, and one that wants a
// question answered asks Spent().
func (p PaidCost) ManaTokens() []ManaToken {
	return ManaSpent{tokens: p.Mana}.Tokens()
}

// ManaSpentCount is how many mana paid — adamant's "at least three"
// and Memory Deluge's "the amount spent" both count from here. Zero
// for an OnPaper payment, which is the weaker answer.
func (p PaidCost) ManaSpentCount() int {
	return p.Spent().Total()
}

// ColorsSpent is the distinct COLOURS the payment spent, in WUBRG
// order. Colourless mana is not a colour (CR 105.1), so {C} never
// appears here — which is exactly what converge and sunburst count,
// and why an Etched Oracle cast for four generic off Sol Rings
// enters as a 0/0.
//
// Empty for an OnPaper payment: an unrecorded payment claims no
// colours, so converge draws nothing and sunburst adds no counters
// rather than guessing five.
func (p PaidCost) ColorsSpent() []string {
	return p.Spent().Colors()
}

// ColorsSpentCount is len(ColorsSpent) without the allocation —
// converge's X and sunburst's counter count.
func (p PaidCost) ColorsSpentCount() int {
	return p.Spent().ColorCount()
}

// SpentOfColor is how many mana of one colour paid: adamant's "at
// least three red mana was spent" is SpentOfColor("R") >= 3. Zero
// for an OnPaper payment.
func (p PaidCost) SpentOfColor(color string) int {
	return p.Spent().Count(color)
}

// NoManaSpent reports the clause "if no mana was spent to cast it"
// (Vexing Bauble, Satoru, the Infiltrator).
//
// It is true only when the engine KNOWS nothing was spent: a cascade
// or "without paying its mana cost" free cast, a {0} alternative
// cost, a copy of a spell (CR 707.10 — mana is not an object, so
// nothing was spent to cast the copy). An OnPaper payment answers
// FALSE, because the player did pay something the engine did not
// see, and a punisher that fired on it would counter half the spells
// cast at a permissive table.
func (p PaidCost) NoManaSpent() bool {
	return p.Spent().None()
}

// Known reports whether the mana half of this record is a fact. The
// escape hatch for a reader that wants to say "unknown" out loud
// rather than fold it into the weaker answer; nothing in the catalog
// needs it yet, and the protocol view uses it to grey the pill.
func (p PaidCost) Known() bool { return !p.OnPaper }

// IsZero reports a record with nothing in it at all — no payment
// was made and none was waived. Used by the clone and snapshot
// paths to keep the common case sparse.
func (p PaidCost) IsZero() bool {
	return len(p.Mana) == 0 && !p.OnPaper &&
		p.CountersRemoved == 0 && p.CountersAdded == 0 && p.LifePaid == 0 &&
		p.Sacrificed == 0 && p.ReturnedAttacking == uuid.Nil && len(p.OptionalCosts) == 0 &&
		len(p.TappedOthers) == 0 && len(p.Exiled) == 0
}

// clonePaidCost deep-copies the record. The ManaToken slice is
// reallocated and each token's Restrictions with it: an undo
// snapshot that aliased the live backing array would let a restore
// mutate the game it was taken from.
func clonePaidCost(p PaidCost) PaidCost {
	out := p
	if len(p.Mana) > 0 {
		out.Mana = make([]ManaToken, len(p.Mana))
		for i, t := range p.Mana {
			out.Mana[i] = t
			if len(t.Restrictions) > 0 {
				out.Mana[i].Restrictions = append([]string(nil), t.Restrictions...)
			}
		}
	}
	if len(p.OptionalCosts) > 0 {
		out.OptionalCosts = append([]int(nil), p.OptionalCosts...)
	}
	if len(p.TappedOthers) > 0 {
		out.TappedOthers = append([]PaidTap(nil), p.TappedOthers...)
	}
	if len(p.Exiled) > 0 {
		out.Exiled = append([]uuid.UUID(nil), p.Exiled...)
	}
	return out
}

// isColorSymbol reports whether a ManaToken's Color is one of the
// five colours. "C" is colourless and every other value is a bug.
func isColorSymbol(c string) bool {
	switch c {
	case "W", "U", "B", "R", "G":
		return true
	}
	return false
}
