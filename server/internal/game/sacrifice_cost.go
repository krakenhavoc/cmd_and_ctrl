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
// pays: the spec's Max. Nil-safe: a nil clause pays nothing.
//
// It answers for a fixed clause only. A VARIABLE one (#1213 —
// "sacrifice one or more", "sacrifice X"; ADR 0100 §3 — "sacrifice any
// number of") has no single number to give, and reads here as its
// floor, which is the weaker answer and never the one a payment should
// be built from. Ask SacrificeCostBounds instead, and
// SacrificeCostVariable first if you need to know which kind you have.
//
// Min 0 with Max 0 is "any number" (ADR 0100 §3, SacrificeAnyNumber),
// whose floor is zero. Before ADR 0100 that pair read as 1 — a
// hand-built fixture that never set a count — and effects.Register
// still refuses it everywhere but a cast's additional cost, where zero
// is printed.
func SacrificeCostCount(spec *TargetSpec) int {
	if spec == nil {
		return 0
	}
	if spec.CountFromX || SacrificeAnyNumber(spec) {
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

// SacrificeAnyNumber reports the "sacrifice any number of …" form
// (ADR 0100 §3): Min 0 and Max 0, not CountFromX. Its bounds are 0 and
// "no ceiling", so naming none is a legal payment. Only a CAST's
// additional cost may carry it — Vicious Betrayal's "sacrifice any
// number of creatures", Torgaar's "you may sacrifice any number of
// creatures", Plumb the Forbidden's "you may sacrifice one or more":
// for these cards "you may" and "any number" mean the same thing,
// because sacrificing none is not paying. effects.Register refuses it
// on an ability, where a cost that can be paid with nothing is free.
// Nil-safe (false).
func SacrificeAnyNumber(spec *TargetSpec) bool {
	return spec != nil && !spec.CountFromX && spec.Min == 0 && spec.Max == 0
}

// SacrificeCostVariable reports a clause whose count the ACTIVATOR
// announces rather than one the card prints (#1213):
//
//   - "Sacrifice one or more artifacts" (Radiant Lotus) — a floor of
//     one with no printed ceiling, Min ≥ 1 and Max 0.
//   - "Sacrifice X Treasures" (Grim Hireling) — CountFromX, the count
//     announced at CR 602.2b with the rest of the activation.
//   - "Sacrifice any number of creatures" (Vicious Betrayal, ADR 0100
//     §3) — Min 0 and Max 0, a cast's additional cost only.
//
// Nil-safe (false). The readers that care are the announce path (how
// many IDs is this payment allowed to name), the protocol view (what
// min / max does the picker enforce) and the legal enumerator (how
// many payments does it offer).
func SacrificeCostVariable(spec *TargetSpec) bool {
	if spec == nil {
		return false
	}
	return spec.CountFromX || spec.Max != spec.Min || SacrificeAnyNumber(spec)
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
//	any number   Min == Max == 0        lo = 0,   hi = 0   (open, ADR 0100)
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
	if SacrificeAnyNumber(spec) {
		// "Any number": zero is a payment, and nothing but the board
		// bounds the count (ADR 0100 §3).
		return 0, 0
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

// sacrificeRefsLocked names the permanents a sacrifice payment is about
// to take as the battlefield OBJECTS they are (ADR 0113 §1, #2072):
// one ObjectRef per ID, in the order named, for
// PaidCost.SacrificedObjects. Read BEFORE payCostSacrificesLocked moves
// them, the way a PaidTap is read before the tap, because the epoch is
// a fact about the object on the battlefield.
//
// One ref per ID always, so the list names exactly the set
// PaidCost.Sacrificed counts. Every ID was validated as a permanent on
// the battlefield, so the fallback (the card's most recent battlefield
// object this turn, else the bare ID at epoch zero) is never taken by a
// payment that passed validation; it keeps the two fields in step if
// one ever were. Nil for no IDs.
//
// Caller must hold g.mu.
func (g *Game) sacrificeRefsLocked(ids []uuid.UUID) []ObjectRef {
	if len(ids) == 0 {
		return nil
	}
	out := make([]ObjectRef, 0, len(ids))
	for _, id := range ids {
		ref, ok := g.PermanentRefForEffect(id)
		if !ok {
			ref = ObjectRef{ID: id}
		}
		out = append(out, ref)
	}
	return out
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

// --- #2526: a sacrifice clause that names different kinds ------------
//
// "Sacrifice a Swamp and a Forest" (Jarad, Golgari Lich Lord) is two
// permanents with DIFFERENT predicates in one clause. TargetSpec.EachOf
// carries the entries; the head spec's own predicate is their union, so
// every per-permanent walk keeps working unchanged and the three
// functions below are the only places that know a set rule exists.

// SacrificeSetGroup is one entry of a sacrifice clause's EachOf rule
// with the candidates that could fill it — what the protocol view ships
// and the client's picker reads.
type SacrificeSetGroup struct {
	// Label is the entry as printed ("a Swamp").
	Label string
	// Candidates are the permanents that pass this entry's predicate, in
	// the order the caller passed them (payment order from the view).
	Candidates []uuid.UUID
}

// SacrificeKind is one entry of a sacrifice clause's EachOf rule: a
// KIND of permanent, as data. A permanent is of the kind when it has any
// of the named subtypes or any of the named card types, read after
// continuous effects (a land something turned into a Swamp counts, and a
// changeling is every creature type — Card.HasSubtype). It carries no
// func, so the clause that holds it adds no route from Game to a closure.
type SacrificeKind struct {
	// Label is the entry as printed ("a Swamp"), for the picker.
	Label string
	// Subtypes, any of: "Swamp", "Treasure". Case as printed.
	Subtypes []string
	// CardTypes, any of, lower-case as Card.HasCardType reads them:
	// "creature", "land", "artifact".
	CardTypes []string
}

// Matches reports whether c is of this kind.
func (k SacrificeKind) Matches(c Card) bool {
	for _, st := range k.Subtypes {
		if c.HasSubtype(st) {
			return true
		}
	}
	for _, ct := range k.CardTypes {
		if c.HasCardType(ct) {
			return true
		}
	}
	return false
}

// SacrificeSetKinds is the EachOf entries of a sacrifice clause, or nil
// when the clause has no set rule. Nil-safe.
func SacrificeSetKinds(spec *TargetSpec) []SacrificeKind {
	if spec == nil {
		return nil
	}
	return spec.EachOf
}

// sacrificeSetFitsLocked answers, for each entry and each candidate,
// whether the candidate is of the entry's kind. A candidate that is not
// on the battlefield fits nothing. Caller must hold g.mu.
func (g *Game) sacrificeSetFitsLocked(entries []SacrificeKind, ids []uuid.UUID) [][]bool {
	fits := make([][]bool, len(entries))
	for i := range entries {
		fits[i] = make([]bool, len(ids))
		for j, id := range ids {
			if c := findBattlefieldCard(g, id); c != nil {
				fits[i][j] = entries[i].Matches(*c)
			}
		}
	}
	return fits
}

// assignSacrificeSet finds one pick per entry such that no pick is used
// twice, preferring earlier candidates for earlier entries. `order` is
// the order candidates are tried in (indexes into fits' columns). It
// returns the chosen column per entry, or nil when no assignment
// exists. A backtracking search: the entry count is the card's printed
// "a X and a Y", two in every card that exists.
func assignSacrificeSet(fits [][]bool, order []int) []int {
	out := make([]int, len(fits))
	taken := map[int]bool{}
	var walk func(i int) bool
	walk = func(i int) bool {
		if i == len(fits) {
			return true
		}
		for _, j := range order {
			if taken[j] || !fits[i][j] {
				continue
			}
			taken[j] = true
			out[i] = j
			if walk(i + 1) {
				return true
			}
			delete(taken, j)
		}
		return false
	}
	if !walk(0) {
		return nil
	}
	return out
}

// sacrificeSetSatisfiedLocked reports whether `picks` can fill every
// EachOf entry of `spec` one-to-one. A clause with no set rule is
// satisfied by anything (its per-permanent check already ran). The
// count is checked by the caller (SacrificeCountLegal); this judges the
// kinds. Caller must hold g.mu.
func (g *Game) sacrificeSetSatisfiedLocked(spec *TargetSpec, picks []uuid.UUID) bool {
	entries := SacrificeSetKinds(spec)
	if len(entries) == 0 {
		return true
	}
	if len(picks) != len(entries) {
		return false
	}
	order := make([]int, len(picks))
	for i := range order {
		order[i] = i
	}
	return assignSacrificeSet(g.sacrificeSetFitsLocked(entries, picks), order) != nil
}

// SacrificeSetSatisfiedForEffect is sacrificeSetSatisfiedLocked for a
// reader outside the package (the legal enumerator's tests, the view).
// Caller must hold g.mu (read or write).
func (g *Game) SacrificeSetSatisfiedForEffect(spec *TargetSpec, picks []uuid.UUID) bool {
	return g.sacrificeSetSatisfiedLocked(spec, picks)
}

// SacrificeSetPaymentForEffect searches `candidates` (already in
// payment order) for ONE set that fills every entry of spec's EachOf
// rule, or nil when the board cannot pay it. It is what replaces "the
// first N of the payment order" for a clause with a set rule: the first
// N may be two Swamps.
//
// A candidate that fits only some entries is tried before one that fits
// them all, so a seat holding a Swamp, a Forest and an Overgrown Tomb
// pays the basics and keeps the dual — payment order alone would have
// eaten whichever came first. Ties keep the order given.
//
// The result is in entry order. Caller must hold g.mu (read or write).
func (g *Game) SacrificeSetPaymentForEffect(spec *TargetSpec, candidates []uuid.UUID) []uuid.UUID {
	entries := SacrificeSetKinds(spec)
	if len(entries) == 0 {
		return nil
	}
	fits := g.sacrificeSetFitsLocked(entries, candidates)
	versatility := make([]int, len(candidates))
	for j := range candidates {
		for i := range entries {
			if fits[i][j] {
				versatility[j]++
			}
		}
	}
	order := make([]int, len(candidates))
	for j := range order {
		order[j] = j
	}
	sort.SliceStable(order, func(a, b int) bool { return versatility[order[a]] < versatility[order[b]] })
	cols := assignSacrificeSet(fits, order)
	if cols == nil {
		return nil
	}
	out := make([]uuid.UUID, len(cols))
	for i, j := range cols {
		out[i] = candidates[j]
	}
	return out
}

// SacrificeSetGroupsForEffect splits `candidates` by the EachOf entry
// each could fill, for the protocol view. A candidate that fits two
// entries (a Swamp Forest) appears under both. Nil without a set rule.
// Caller must hold g.mu (read or write).
func (g *Game) SacrificeSetGroupsForEffect(spec *TargetSpec, candidates []uuid.UUID) []SacrificeSetGroup {
	entries := SacrificeSetKinds(spec)
	if len(entries) == 0 {
		return nil
	}
	fits := g.sacrificeSetFitsLocked(entries, candidates)
	out := make([]SacrificeSetGroup, len(entries))
	for i := range entries {
		out[i].Label = entries[i].Label
		for j, id := range candidates {
			if fits[i][j] {
				out[i].Candidates = append(out[i].Candidates, id)
			}
		}
	}
	return out
}
