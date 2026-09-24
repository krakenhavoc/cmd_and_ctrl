package game

import (
	"sort"

	"github.com/google/uuid"
)

// sacrifice_cost.go — #747 (ADR 0020 addendum §12–§15, ADR 0021
// addendum): a sacrifice cost of N permanents. "Sacrifice two
// artifacts" (Sai), "Sacrifice three Foods" (Samwise Gamgee),
// "Sacrifice five Treasures" (Magda, Brazen Outlaw).
//
// There is no new cost component. The count lives on the sacrifice
// clause itself, as the TargetSpec's Min == Max, so the three cost
// sites that already carry the clause — AbilityCost.SacrificeOther,
// ManaAbilityShape.SacrificeOther and AdditionalCost.Sacrifice — get
// the count with no new field, and a cost merge (effects.Plus) cannot
// separate the count from the thing it counts.
//
// Three pieces live here because all three sites and both readers
// (the legal enumerator and the protocol view) share them:
//
//   - SacrificeCostCount reads the count off a clause.
//   - payCostSacrificesLocked pays one payment as one simultaneous
//     exit.
//   - SacrificePaymentOrderForEffect is the policy-neutral order the
//     enumerator takes its one payment from, and the order the view
//     ships the options in, so the client's "Choose for me" fills the
//     picker with the same permanents a bot would pay.

// SacrificeCostCount is how many permanents a FIXED sacrifice clause
// pays: the spec's Max. A hand-built fixture clause that never set a
// count is read as 1, the count every sacrifice cost had before #747.
// Nil-safe: a nil clause pays nothing.
//
// It answers for a fixed clause only. A VARIABLE one (#1213 —
// "sacrifice one or more", "sacrifice X") has no single number to
// give, and reads here as its floor, which is the weaker answer and
// never the one a payment should be built from. Ask
// SacrificeCostBounds instead, and SacrificeCostVariable first if you
// need to know which kind you have.
func SacrificeCostCount(spec *TargetSpec) int {
	if spec == nil {
		return 0
	}
	if spec.CountFromX {
		return 0
	}
	if spec.Max < 1 {
		if spec.Min > 1 {
			return spec.Min
		}
		return 1
	}
	return spec.Max
}

// SacrificeCostVariable reports a clause whose count the ACTIVATOR
// announces rather than one the card prints (#1213):
//
//   - "Sacrifice one or more artifacts" (Radiant Lotus) — a floor of
//     one with no printed ceiling, Min ≥ 1 and Max 0.
//   - "Sacrifice X Treasures" (Grim Hireling) — CountFromX, the count
//     announced at CR 602.2b with the rest of the activation.
//
// Nil-safe (false). The readers that care are the announce path (how
// many IDs is this payment allowed to name), the protocol view (what
// min / max does the picker enforce) and the legal enumerator (how
// many payments does it offer).
func SacrificeCostVariable(spec *TargetSpec) bool {
	if spec == nil {
		return false
	}
	return spec.CountFromX || spec.Max != spec.Min
}

// SacrificeCountFromX reports the "Sacrifice X …" form specifically —
// the one whose count is the announced X rather than a number the
// activator is free to pick. Nil-safe (false).
//
// It is what makes AbilityCost.DemandsX() true for Grim Hireling,
// whose mana component ({B}) has no {X} slot at all.
func SacrificeCountFromX(spec *TargetSpec) bool {
	return spec != nil && spec.CountFromX
}

// SacrificeCostBounds is how many permanents ONE payment of a
// sacrifice clause may name, given the X announced at CR 601.2b /
// 602.2b (#1213). `lo` is the floor and `hi` the ceiling; **hi == 0
// means there is no printed ceiling** and the activator's own board is
// the only bound.
//
// The three shapes, and the whole of the variable-count design:
//
//	fixed        Min == Max == N        lo = hi = N
//	one or more  Min >= 1, Max == 0     lo = Min, hi = 0   (open)
//	X            CountFromX             lo = hi = x
//
// A CountFromX clause with a negative announced X reads as zero rather
// than as a negative bound: the announce path rejects a negative X
// separately, and a bound that could go negative would make an empty
// payment look illegal for the wrong reason.
//
// Nil-safe: a nil clause pays nothing, 0 / 0 — which a caller must
// read as "no component", not as "an open count", so every caller
// checks the clause for nil before asking.
func SacrificeCostBounds(spec *TargetSpec, x int) (lo, hi int) {
	if spec == nil {
		return 0, 0
	}
	if spec.CountFromX {
		if x < 0 {
			x = 0
		}
		return x, x
	}
	n := SacrificeCostCount(spec)
	if spec.Max < 1 && spec.Min >= 1 {
		// "One or more": a floor the card prints and no ceiling.
		return spec.Min, 0
	}
	return n, n
}

// SacrificeCountLegal reports whether naming `named` permanents is a
// legal payment of `spec` at the announced X — the one predicate the
// validator, the protocol view and the legal enumerator share, so
// none of them can offer or accept a count the others refuse (#544).
//
// The CountFromX form is answered FIRST rather than through the
// bounds, because the two readings of `hi == 0` collide there: it is
// the "no printed ceiling" sentinel for an open count, and it is the
// real and only legal answer for an X announced as zero. A clause
// whose count IS the announced X has one legal payment and this says
// so directly.
//
// Nil-safe: a nil clause is paid by naming nothing.
func SacrificeCountLegal(spec *TargetSpec, x, named int) bool {
	if spec == nil {
		return named == 0
	}
	if spec.CountFromX {
		if x < 0 {
			x = 0
		}
		return named == x
	}
	lo, hi := SacrificeCostBounds(spec, x)
	if named < lo {
		return false
	}
	return hi == 0 || named <= hi
}

// payCostSacrificesLocked sacrifices every permanent of one cost
// payment — the source when the cost sacrifices it, and the N
// permanents of the clause — as ONE simultaneous battlefield exit.
//
// Each permanent still goes through sacrificePermanentLocked, so each
// emits its own EventSacrifice and takes its own route to the
// graveyard: "whenever you sacrifice a permanent" and dies triggers
// fire once per permanent. What the batch adds is what the watchers
// see. Leaves-the-battlefield and sacrifice triggers look back in time
// (CR 603.10a), so a Blood Artist sacrificed alongside a Goblin must
// see both deaths. Paid one at a time with no batch, the Artist would
// see the Goblin only when the Goblin happened to go first, and the
// trigger count would depend on the order of IDs on the wire.
//
// beginSimultaneousExitLocked has no indestructible filter (that lives
// in DestroyPermanentsForEffect, which this does not call), so an
// indestructible permanent is still sacrificed, as CR 701.21a says.
//
// Caller must hold g.mu in write mode, and must have validated the
// payment with validateSacrificeCostLocked.
func (g *Game) payCostSacrificesLocked(ids []uuid.UUID, answers map[uuid.UUID]bool) error {
	if len(ids) == 0 {
		return nil
	}
	defer g.beginSimultaneousExitLocked(ids)()
	for _, id := range ids {
		// #1397: `answers` are the CR 903.9 answers the commanders'
		// owners gave before the payment began; nil for a payer that
		// asks nobody (the auto-tapper), which is "unasked".
		if err := g.sacrificeAnsweredLocked(id, commanderAnswerFor(answers, id)); err != nil {
			return err
		}
	}
	return nil
}

// SacrificePaymentOrderForEffect returns `ids` — candidates for a
// sacrifice cost — in the order a payment takes them from (ADR 0020
// addendum §15):
//
//  1. tokens before nontokens;
//  2. then lower mana value;
//  3. then the ability's own source last, when it is a candidate
//     (pass uuid.Nil when there is no source, as for a spell's
//     additional cost);
//  4. then the order given, which callers pass in board order, so the
//     result is stable.
//
// The order is policy-neutral on purpose. The legal enumerator offers
// one sacrifice-N payment per move — the first N of this order —
// rather than every combination, and `legal` must not import a bot
// policy (#687). The protocol view ships its options in the same
// order, so the owner's "Choose for me" button (#747) fills the picker
// with exactly the permanents a bot would have paid.
//
// A new slice; `ids` is not modified. IDs not on the battlefield sort
// as nontokens of mana value 0, which no caller passes.
//
// Caller must hold g.mu (read or write).
func (g *Game) SacrificePaymentOrderForEffect(ids []uuid.UUID, sourceID uuid.UUID) []uuid.UUID {
	type key struct {
		id     uuid.UUID
		token  bool
		mv     int
		source bool
	}
	keys := make([]key, 0, len(ids))
	for _, id := range ids {
		k := key{id: id, source: sourceID != uuid.Nil && id == sourceID}
		if c := findBattlefieldCard(g, id); c != nil {
			k.token = typeLineHas(c.TypeLine, "token")
			k.mv = c.ManaValue()
		}
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.token != b.token {
			return a.token
		}
		if a.mv != b.mv {
			return a.mv < b.mv
		}
		if a.source != b.source {
			return b.source
		}
		return false
	})
	out := make([]uuid.UUID, len(keys))
	for i, k := range keys {
		out[i] = k.id
	}
	return out
}
