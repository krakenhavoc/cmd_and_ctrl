package game

import (
	"strings"

	"github.com/google/uuid"
)

// saturate.go — the table's ceilings on counts (#3040).
//
// Nothing in the rules bounds a +1/+1 counter, a power, or the amount a
// mana ability adds, and the engine used to take that literally. Twenty
// Ouroboroids double a board every combat: 2^20 power is one combat, 2^40
// is the next, 2^60 the one after, and the combat after that wraps `int`
// negative. Long before the wrap, Marwyn, the Nurturer tapping at 2^40
// power builds strings.Repeat("{G}", 2^40) — a single 3.3 TB string. Go's
// out-of-memory is a fatal error `recover()` cannot catch, so one table's
// doubling board would take every room on the box down with it.
//
// Two ceilings, because the two quantities are stored differently:
//
//   - MaxCount bounds counters, power and toughness. They are plain ints
//     in a map or a struct field, so the only thing a big one costs is
//     the arithmetic that touches it. 2^30 is far past any real board
//     (MTGO caps at 2^31 - 1, and no paper table has ever written down a
//     number that large) and leaves headroom: two clamped values added or
//     two clamped values multiplied still fit a 64-bit int, so the
//     helpers below saturate instead of ever wrapping.
//
//   - MaxProducedMana bounds what one mana-producing instruction adds.
//     The pool stores ONE ENTRY PER MANA, and every action snapshots the
//     game for undo, so a floating pool costs memory per mana per action
//     (about 860 MB for the restore point alone at 2^20 mana). 1024 is
//     past anything a real board makes — Gaea's Cradle under a table of
//     tokens is tens, a pathological Marwyn is hundreds — and keeps the
//     worst case at a few hundred KB. It is a PER-PRODUCTION ceiling and
//     says nothing about a pool that several taps fill; the pool's own
//     representation is the companion issue, after which this number can
//     rise.
//
// The clamp lives at the WRITERS — counter placement, the P/T read, the
// mana parse — rather than in the readers, so the stored state is never
// out of range and a reader added next year is safe by construction.
// A clamp that actually bites is announced (EventCountCapped), so a
// capped table is told it is capped rather than left to wonder why the
// number stopped growing.
const (
	// MaxCount is the ceiling on a counter total, a power or a toughness.
	MaxCount = 1 << 30
	// MaxProducedMana is the ceiling on the mana one production adds.
	MaxProducedMana = 1 << 10
)

// ClampCount pins n to [-MaxCount, MaxCount].
func ClampCount(n int) int {
	switch {
	case n > MaxCount:
		return MaxCount
	case n < -MaxCount:
		return -MaxCount
	}
	return n
}

// SaturatingAdd is a + b pinned to [-MaxCount, MaxCount], computed so
// that no input, however large, can wrap.
func SaturatingAdd(a, b int) int {
	a, b = ClampCount(a), ClampCount(b)
	return ClampCount(a + b)
}

// SaturatingMul is a * b pinned to [-MaxCount, MaxCount]. Operands are
// pinned first, and two pinned values multiply to at most 2^60, which an
// int holds, so the product cannot wrap before it is pinned again.
func SaturatingMul(a, b int) int {
	a, b = ClampCount(a), ClampCount(b)
	return ClampCount(a * b)
}

// RepeatMana is strings.Repeat(slot, n) with the count bounded. It is
// what every card that builds a produced-mana string from a board
// quantity (a power, a counter total, a creature count) must use: the
// string is allocated BEFORE the engine ever sees it, so the engine
// cannot clamp it after the fact. A non-positive n yields "".
//
// An n past the ceiling yields ONE MORE than MaxProducedMana copies, on
// purpose: the engine's own cut (parseProducedMana) is the single place
// that announces a clamp, and it can only announce one it can see. The
// extra slot is the evidence, and is itself dropped by that cut.
func RepeatMana(slot string, n int) string {
	if n <= 0 || slot == "" {
		return ""
	}
	if n > MaxProducedMana {
		n = MaxProducedMana + 1
	}
	return strings.Repeat(slot, n)
}

// capProducedMana pins a parsed produced-mana declaration to
// MaxProducedMana entries and reports whether it had to. The parser
// already caps each "{G3}" amount; this is the sum across slots.
func capProducedMana(slots []ProducedManaEntry) ([]ProducedManaEntry, bool) {
	if len(slots) <= MaxProducedMana {
		return slots, false
	}
	return slots[:MaxProducedMana], true
}

// emitCountCappedLocked announces that a clamp bit. `what` names the
// quantity ("+1/+1 counters", "mana"), `card` is the object whose count
// was capped (uuid.Nil when there is none), and `ceiling` the value it
// was pinned to. Caller must hold g.mu.
func (g *Game) emitCountCappedLocked(what string, card, actor uuid.UUID, ceiling int) {
	g.EmitEvent(Event{
		Kind:   EventCountCapped,
		Actor:  actor,
		Source: card,
		Label:  what,
		Amount: ceiling,
	})
}

// ManaAmount is a count-driven mana output as data (#3042): N mana in
// the slot whose colour options are Options. The zero value, a
// non-positive N or an empty Options adds nothing.
//
//   - one option: N mana of that colour (Gaea's Cradle, Marwyn);
//   - several options, OneColor false: N independent picks, one mana
//     each (Selvala's "three mana in any combination of colors");
//   - several options, OneColor true: ONE pick, and N mana of the
//     picked colour ("N mana of any one color", Mona Lisa).
//
// It is the smaller of the two shapes #3042 sketched: a (colour, n)
// pair cannot say "any one color", a run list is #3041's pool shape and
// no helper here mixes colours in one result. Options is shared, not
// copied, by the slots built from it; callers keep it immutable.
type ManaAmount struct {
	Options  []string
	N        int
	OneColor bool
}

// slotsFromAmount is the structured twin of parseProducedMana for a
// ProducedAmount ability: the slots the engine would have parsed, built
// straight from the count with the MaxProducedMana ceiling applied to N
// BEFORE anything is allocated, and capped reporting whether it bit.
// No text exists at any point.
func slotsFromAmount(a ManaAmount) (slots []ProducedManaEntry, capped bool) {
	if a.N <= 0 || len(a.Options) == 0 {
		return nil, false
	}
	n := a.N
	if n > MaxProducedMana {
		n, capped = MaxProducedMana, true
	}
	if a.OneColor && len(a.Options) > 1 {
		entry := ProducedManaEntry{Options: a.Options}
		if n != 1 {
			entry.Amounts = make(map[string]int, len(a.Options))
			for _, c := range a.Options {
				entry.Amounts[c] = n
			}
		}
		return []ProducedManaEntry{entry}, capped
	}
	slots = make([]ProducedManaEntry, n)
	for i := range slots {
		slots[i] = ProducedManaEntry{Options: a.Options}
	}
	return slots, capped
}

// manaAbilitySlotsLocked is the produced-mana slots one mana ability
// would add, given what its cost paid: manaAbilityProducedLocked's
// precedence, but a ProducedAmount ability never goes through text.
// The result is bounded by the ceiling; the pool-adding site, which
// announces a cut, reads slotsFromAmount itself. Read-only.
// Caller must hold g.mu.
func manaAbilitySlotsLocked(g *Game, playerID, cardID uuid.UUID, ab *ManaAbilityShape, paid PaidCost) ([]ProducedManaEntry, error) {
	if ab != nil && ab.ProducedAmount != nil && ab.ProducedForPaid == nil && ab.DerivedMatch == nil {
		slots, _ := slotsFromAmount(ab.ProducedAmount(g, playerID, cardID))
		return slots, nil
	}
	return ParseProducedMana(manaAbilityProducedLocked(g, playerID, cardID, ab, paid))
}
