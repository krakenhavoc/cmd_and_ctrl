package game

// mana_spent.go — #1212: the READ of what mana paid for something.
//
// #761 built the record (PaidCost.Mana, ADR 0068) and hung six
// accessors off PaidCost for the five mechanics that shipped with it.
// That worked while every reader was a spell reading its own stack
// item. It stops working the moment the reader is a PERMANENT — "when
// this creature enters, if mana from a Treasure was spent to cast it"
// — because by then the item is gone and the fact lives on
// Card.Provenance instead (CR 400.7d, #653).
//
// Two homes for one fact is the moment a vocabulary either gets
// written down or gets duplicated. So the accessors move onto a VIEW
// that either home can hand out, and PaidCost and CastProvenance both
// return it. There is one set of questions about spent mana in this
// engine, and this file is it.
//
// It is a view and not a record: unexported fields, no JSON tags,
// built on demand from whichever home holds the tokens. The RECORD is
// still the token slice, in exactly one place per announcement.

// ManaSpent is what one announcement's payment can be asked about.
//
// The zero value is "nothing was spent, and that is known" — a free
// cast, a copy of a spell (CR 707.10), a permanent that was never
// cast at all. `unknown` is the other absence: a payment the engine
// WAIVED in part (an insufficient permissive pool or ForceCast), where
// every question below answers in the weaker-than-printed direction
// because the engine genuinely does not know (ADR 0068 §3).
//
// Since #3844 (ADR 0068 amendment 2026-10-15) a waived payment still
// spends what the pool could pay, so an unknown record can carry real
// tokens: the mana that left the pool. Every counting question below
// reads them as a LOWER BOUND — what is certainly true of the payment —
// which is never stronger than printed for a reader that grows with
// the mana spent. None is the one reader that shrinks with it, and it
// answers false for an unknown record whatever it holds.
//
// Since #3855 an undeclared waived payment also carries the AMOUNT it
// paid outside the pool (`paper`), so Total is exact on every record;
// only the colours and sources of that share stay unknown. A payment a
// reader cares about is never undeclared: the caster is asked, and the
// answer is in the tokens.
type ManaSpent struct {
	tokens  []ManaToken
	paper   int
	unknown bool
}

// Spent is the readable view of the mana half of this record.
func (p PaidCost) Spent() ManaSpent {
	return ManaSpent{tokens: p.Mana, paper: p.PaperMana, unknown: p.OnPaper}
}

// Known reports whether this is a fact rather than a waived charge.
// The escape hatch for a reader that wants to say "unknown" out loud
// instead of folding it into the weaker answer.
func (m ManaSpent) Known() bool { return !m.unknown }

// Tokens is the tokens themselves, copied, in the order the solver
// spent them. Each is a RUN (#3041): read its amount with N. For an
// unknown payment, the tokens the pool did pay (#3844) — nil when it
// paid none. For a reader that needs a fact this file does not name.
func (m ManaSpent) Tokens() []ManaToken {
	if len(m.tokens) == 0 {
		return nil
	}
	return append([]ManaToken(nil), m.tokens...)
}

// Total is how many mana paid — "if five or more mana was spent to
// cast that spell" (Colorstorm Stallion), "at least seven mana"
// (Raggadragga), and the amount Memory Deluge draws off. For an
// unknown payment, the mana that left the pool: a lower bound, and
// zero when the pool paid nothing.
//
// Since #3855 that bound is the whole amount: an undeclared waived
// payment carries the shortfall it asserted paid on paper.
func (m ManaSpent) Total() int {
	return TotalMana(m.tokens) + m.paper
}

// Count is how many mana of one SYMBOL paid: `{W}` through `{G}` and
// `"C"` for colourless (CR 107.4c). Adamant's "at least three red
// mana was spent to cast this spell" is Count("R") >= 3, and
// Desecrate Reality's "at least three colorless mana" is Count("C")
// >= 3 — the reason this takes a symbol rather than a colour.
//
// For an unknown payment, the mana of that symbol that left the pool:
// a lower bound, so adamant turns on only when the pool alone paid
// three.
func (m ManaSpent) Count(symbol string) int {
	n := 0
	for _, t := range m.tokens {
		if t.Color == symbol {
			n += t.N()
		}
	}
	return n
}

// Colors is the distinct COLOURS spent, in WUBRG order. Colourless is
// not a colour (CR 105.1), so `{C}` never appears here — which is
// exactly what converge (CR 702.86) and sunburst (CR 702.44) count,
// and why an Etched Oracle paid for entirely out of Sol Rings enters
// as a 0/0.
//
// For an unknown payment, only the colours that left the pool: the
// part paid on paper claims no colours rather than guessing five.
func (m ManaSpent) Colors() []string {
	seen := make(map[string]bool, 5)
	for _, t := range m.tokens {
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

// ColorCount is len(Colors) without the allocation — converge's X and
// sunburst's counter count.
func (m ManaSpent) ColorCount() int { return len(m.Colors()) }

// None reports the clause "if no mana was spent to cast it" (Vexing
// Bauble, Satoru, the Infiltrator).
//
// True only when the engine KNOWS nothing was spent. An unknown
// payment answers FALSE — the player did pay something the engine did
// not see, and a punisher that fired on it would counter half the
// spells cast at a permissive table. The one reader that does not take
// an unknown record's tokens as a lower bound, because it is the one
// that shrinks as the mana spent grows.
func (m ManaSpent) None() bool { return !m.unknown && m.paper == 0 && len(m.tokens) == 0 }

// CountFrom is how many of the mana came from a source matching ANY
// of `kinds` — Marut's "create a Treasure token for each mana from a
// Treasure spent to cast it", and Inga and Esika's "three or more
// mana from creatures".
//
// The kinds were snapshotted when each mana was MADE (mana_source.go),
// so a Treasure that sacrificed itself to pay for this spell still
// counts. Zero for a zero `kinds`, which asks about nothing.
//
// For an unknown payment, the mana from those sources that left the
// pool (#3844). That is exact rather than a bound for any source the
// engine runs: its mana exists only as a token in the pool, so mana
// from it that did not leave the pool was not spent.
func (m ManaSpent) CountFrom(kinds ManaSourceKinds) int {
	if kinds == 0 {
		return 0
	}
	n := 0
	for _, t := range m.tokens {
		if t.SourceKinds.HasAny(kinds) {
			n += t.N()
		}
	}
	return n
}

// From reports "if mana from a <kind> was spent" — the clause
// eighteen printed cards spell with Treasure. For an unknown payment,
// true when such mana left the pool (#3844).
func (m ManaSpent) From(kinds ManaSourceKinds) bool {
	return m.CountFrom(kinds) > 0
}

// FromTreasure is From(ManaSourceTreasure), named because that is how
// every card that reads it is worded (Hired Hexblade, Jaded
// Sell-Sword, Devour Intellect, Rain of Riches).
func (m ManaSpent) FromTreasure() bool { return m.From(ManaSourceTreasure) }

// Snow reports whether any of the mana came from a snow source
// (CR 205.4h) — what a `{S}` symbol in a cost is paid with.
//
// Nothing in the catalog reads it: no printed card asks whether snow
// mana was spent, and `ParsedCost.HasSnow` is still informational
// (mana_cost.go:39). It is here because the snapshot records the bit
// anyway and the `{S}` work has nowhere else to read from.
func (m ManaSpent) Snow() bool { return m.From(ManaSourceSnow) }
