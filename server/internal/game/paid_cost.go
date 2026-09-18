package game

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
}

// ManaSpent is the tokens that paid, or nil. The accessor rather
// than the field so a caller cannot append into the record.
func (p PaidCost) ManaSpent() []ManaToken {
	if len(p.Mana) == 0 {
		return nil
	}
	return append([]ManaToken(nil), p.Mana...)
}

// ManaSpentCount is how many mana paid — adamant's "at least three"
// and Memory Deluge's "the amount spent" both count from here. Zero
// for an OnPaper payment, which is the weaker answer.
func (p PaidCost) ManaSpentCount() int {
	return len(p.Mana)
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
	if p.OnPaper {
		return nil
	}
	seen := make(map[string]bool, 5)
	for _, t := range p.Mana {
		if isColorSymbol(t.Color) {
			seen[t.Color] = true
		}
	}
	var out []string
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		if seen[c] {
			out = append(out, c)
		}
	}
	return out
}

// ColorsSpentCount is len(ColorsSpent) without the allocation —
// converge's X and sunburst's counter count.
func (p PaidCost) ColorsSpentCount() int {
	return len(p.ColorsSpent())
}

// SpentOfColor is how many mana of one colour paid: adamant's "at
// least three red mana was spent" is SpentOfColor("R") >= 3. Zero
// for an OnPaper payment.
func (p PaidCost) SpentOfColor(color string) int {
	if p.OnPaper {
		return 0
	}
	n := 0
	for _, t := range p.Mana {
		if t.Color == color {
			n++
		}
	}
	return n
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
	return !p.OnPaper && len(p.Mana) == 0
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
		p.CountersRemoved == 0 && p.CountersAdded == 0 && p.LifePaid == 0
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
