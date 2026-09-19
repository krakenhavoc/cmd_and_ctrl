package game

import (
	"sort"

	"github.com/google/uuid"
)

// counter_cost.go — #625, then #789, then #943: counters as part of
// paying a cost (CR 602.2b / 118.3).
//
// #625 made "remove N counters" a component of an ACTIVATED
// ability's cost, in three printed shapes. #789 and #943 finish the
// seam, and the shapes they add are not new components — they are
// more answers to the one question the component already asks,
// "which counters come off, and from where":
//
//	self       "Remove a gold counter from this artifact"   From == nil
//	other      "remove a loyalty counter from a planeswalker you control"
//	any kind   "Remove a counter from a creature you control"  Counter == ""
//	variable   "Remove any number of storage counters from this land"  Variable
//	among      "Remove two +1/+1 counters from among artifacts you control"  Among
//	any-kind
//	among      "Remove three counters from among other artifacts,
//	           creatures, and planeswalkers you control"  Counter == "" && Among
//
// The last one is #943's, and it is the seam's final printed shape:
// Tekuthal, Inquiry Dominus removes three counters of WHATEVER kinds
// happen to be there, so the payment names a kind per permanent as
// well as a count. It is the among form with the kind question asked
// once per part instead of once per payment — not a sixth component.
//
// One type, one validator, one candidate walk, one client picker and
// one enumerator arm cover all six. The alternative — a sibling
// type per shape — was rejected in the ADR 0020 addendum: four
// validators that must agree about "you control it and it is not
// targeted" is four chances to disagree, and #544's lesson is that
// the enumerator and the engine disagreeing is the expensive bug.
//
// #789 also adds the other direction, a cost that ADDS a counter
// (Devoted Druid's "Put a -1/-1 counter on this creature"), as
// AbilityCost.AddCounter. Adding is genuinely a different operation
// — nothing is chosen, and CR 118.3's "you can't pay what you can't
// pay" bites on a prohibition rather than on a supply — so it is its
// own small type rather than a sixth answer here.
//
// Both halves share one rule with the loyalty cost that predates
// them: paying a cost is NOT an effect (CR 121.1), so neither the
// removal nor the placement is a replaceable event. Doubling Season
// does nothing to a Devoted Druid's -1/-1, and Vorinclex does
// nothing to a Vivid land's charge counter.

// CounterRemovalCost is the "remove N counters" component of an
// ability's cost. It is declared on an activated ability
// (AbilityCost.RemoveCounters) and, since #789, on a mana ability
// (ManaAbilityShape.RemoveCounters) — ONE type with two owners, so
// that Vivid Creek and Heart of Kiran are validated, enumerated,
// priced and rendered by the same code.
type CounterRemovalCost struct {
	// Counter is the kind removed ("loyalty", "+1/+1", "charge",
	// "storage"). Empty means "a counter" of ANY kind, and the
	// activator names the kind at announce alongside the permanent
	// (Fain, the Broker).
	//
	// An any-kind cost from ONE permanent removes N counters of the
	// one kind chosen, and effects.Register refuses N > 1 for it: a
	// single kind choice could not say how a mixed payment was made.
	// With Among set the kinds are asked per permanent instead
	// (#943), so "remove three counters from among …" of any kind is
	// a shape — Tekuthal, Inquiry Dominus prints it, and a loyalty
	// counter off a planeswalker plus two +1/+1 counters off a
	// creature is a legal payment.
	Counter string

	// N is how many counters are removed — or, when Variable is set,
	// the FLOOR on how many the activator may announce. Always at
	// least 1 for a fixed cost; effects.Register panics at boot
	// otherwise. A variable cost may print a floor of 0 ("any
	// number", Mage-Ring Network).
	N int

	// From is the "from a planeswalker you control" clause: a
	// predicate over permanents, reused from TargetSpec exactly as
	// SacrificeOther reuses it. It does NOT target — the permanent is
	// chosen as a cost, so hexproof, shroud and protection never
	// apply — and the controller restriction ("you control") is
	// enforced by the engine rather than asked of the spec.
	//
	// Nil means the source itself ("Remove a +1/+1 counter from
	// Mikaeus").
	From *TargetSpec

	// Among is the "from AMONG artifacts you control" form (#789):
	// the N counters may be split across any number of the
	// permanents From matches, in whatever per-permanent amounts the
	// activator names, as long as they total exactly N. Iron Spider,
	// Stark Upgrade's "Remove two +1/+1 counters from among
	// artifacts you control" is the shape; so is Hopeful Initiate's.
	//
	// Meaningless without From (there is only one source), and
	// meaningless with Variable (no printed card splits a variable
	// count); Register refuses both.
	//
	// The payment is validated as a SET, the way a crew payment is:
	// every named (permanent, kind) must be distinct, the permanent
	// controlled by the activator, matched by From and holding at
	// least the count named against it, and the counts must sum to
	// exactly N. Any failure refuses the whole activation with no
	// counter removed.
	//
	// The set's key is (permanent, KIND) rather than the permanent
	// alone because of #943's any-kind form: one creature carrying a
	// +1/+1 and a shield counter may pay both, and those are two
	// parts of one payment. For a printed kind the two keys are the
	// same thing, so a fixed-kind among payment that names one
	// permanent twice is refused exactly as it was before.
	Among bool

	// Variable is "Remove X counters" / "Remove any number of
	// counters" (#789): the count is announced at activation, the
	// way MinX announces an {X} in a mana component, and N is the
	// floor rather than the amount. Crucible of the Spirit Dragon
	// prints the X form and Mage-Ring Network the any-number form;
	// they are the same thing, because what a cost's X can be is
	// bounded by what the payer can actually pay and by nothing
	// else.
	//
	// The paid count reaches the effect through the ONE paid-cost
	// record (PaidCost.CountersRemoved) — on the stack item for an
	// activated ability, and handed to ProducedForPaid for a mana
	// ability, which is what lets "Add {C} for each storage counter
	// removed this way" be a fact about the announcement rather than
	// a re-read of a board that no longer holds those counters.
	Variable bool
}

// CounterAddCost is a cost that PUTS a counter on the source —
// Devoted Druid's "Put a -1/-1 counter on this creature" (#789). It
// is always the source: no printed card pays a cost by putting a
// counter on something else, and inventing the clause would mean
// inventing a picker for it.
type CounterAddCost struct {
	// Counter is the kind placed ("-1/-1", "+1/+1"). Required;
	// effects.Register panics on an empty kind.
	Counter string

	// N is how many go on. At least 1, and 1 on every printed card
	// that has this cost today.
	N int
}

// CounterCostKind is one counter kind a permanent holds that could pay
// a CounterRemovalCost, with how many of that kind it has.
type CounterCostKind struct {
	Kind  string
	Count int
}

// CounterCostOption is one permanent that could pay a
// CounterRemovalCost right now, and the kinds on it that could. For a
// fixed-kind cost Kinds has exactly one entry; for an any-kind cost it
// lists every kind with at least N counters, most counters first.
type CounterCostOption struct {
	CardID uuid.UUID
	Kinds  []CounterCostKind
}

// most is the largest count across the option's kinds, the ordering
// key both the protocol view and the move enumerator use.
func (o CounterCostOption) most() int {
	m := 0
	for _, k := range o.Kinds {
		if k.Count > m {
			m = k.Count
		}
	}
	return m
}

// counterFloor is how many counters ONE permanent must hold to be
// worth offering for this cost: N for a fixed single-permanent cost,
// and 1 for the among and variable forms, where a permanent that
// holds a single counter is a legal part of a payment.
func counterFloor(rc *CounterRemovalCost) int {
	if rc.Among || rc.Variable {
		if rc.N <= 0 {
			return 1
		}
		return 1
	}
	return rc.N
}

// CounterCostOptionsForEffect is the set of (permanent, kind) choices
// that could contribute to paying `rc` for an activation of an
// ability on `sourceID` by `playerID`, ordered most counters first
// (ties keep battlefield order). Empty when the cost cannot be paid
// at all.
//
// ONE walk shared by the protocol view (the client's picker), the
// legal-move enumerator (the bots) and — through the same predicates
// — validateCounterRemovalLocked, so the three cannot disagree about
// which permanent pays. It is the NON-targeting candidate walk
// (SpecCandidatesForEffect): choosing a permanent to pay a cost does
// not target it, so a hexproof planeswalker still pays Heart of
// Kiran's crew.
//
// For the among form it lists every permanent that holds at least one
// counter of the kind; whether they add up to N is the PAYMENT's
// question, answered by CounterCostPayable and by the validator, not
// by this walk — a list that pre-filtered to "sets that total N"
// would have to enumerate subsets, which is the combinatorial
// expansion #544 warns about.
//
// Caller must hold g.mu (read or write).
func (g *Game) CounterCostOptionsForEffect(playerID, sourceID uuid.UUID, rc *CounterRemovalCost) []CounterCostOption {
	if rc == nil || (rc.N <= 0 && !rc.Variable) {
		return nil
	}
	var ids []uuid.UUID
	if rc.From == nil {
		ids = []uuid.UUID{sourceID}
	} else {
		ids = g.specCandidatesLocked(playerID, rc.From).Cards
	}
	floor := counterFloor(rc)
	var out []CounterCostOption
	for _, id := range ids {
		c := findBattlefieldCard(g, id)
		if c == nil || c.Controller != playerID {
			continue
		}
		kinds := counterKindsPaying(c, rc.Counter, floor)
		if len(kinds) == 0 {
			continue
		}
		out = append(out, CounterCostOption{CardID: id, Kinds: kinds})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].most() > out[j].most() })
	return out
}

// CounterCostPayable reports whether `rc` could be paid at all right
// now — the question the auto-tapper, the legal enumerator and the
// client's greyed menu row all ask before offering an activation.
//
// For the single-permanent forms that is "some option exists". For
// the among form it is the extra question the option list does not
// answer: do the counters that could pay, across the matched
// permanents, total at least N? For a printed kind that is one kind's
// total; for #943's any-kind among (Tekuthal) every kind counts,
// because every kind can pay.
//
// Caller must hold g.mu (read or write).
func (g *Game) CounterCostPayable(playerID, sourceID uuid.UUID, rc *CounterRemovalCost) bool {
	if rc == nil {
		return true
	}
	opts := g.CounterCostOptionsForEffect(playerID, sourceID, rc)
	if len(opts) == 0 {
		return rc.Variable && rc.N <= 0
	}
	if !rc.Among {
		return true
	}
	total := 0
	for _, o := range opts {
		for _, k := range o.Kinds {
			if rc.Counter == "" || k.Kind == rc.Counter {
				total += k.Count
			}
		}
	}
	return total >= rc.N
}

// counterKindsPaying lists the kinds on c that could contribute at
// least `floor` counters, most counters first and then by name so
// the order is deterministic. An empty `kind` is the any-kind form.
func counterKindsPaying(c *Card, kind string, floor int) []CounterCostKind {
	if floor < 1 {
		floor = 1
	}
	if kind != "" {
		if n := c.Counters[kind]; n >= floor {
			return []CounterCostKind{{Kind: kind, Count: n}}
		}
		return nil
	}
	var out []CounterCostKind
	for k, n := range c.Counters {
		if k == "" || n < floor {
			continue
		}
		out = append(out, CounterCostKind{Kind: k, Count: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}

// counterPaymentPart is one permanent's share of a counter-removal
// payment: which kind comes off it, and how many.
//
// The kind lives on the PART rather than on the payment because of
// #943: an any-kind among cost is paid in whatever kinds the
// permanents hold. Every other shape simply repeats one kind across
// its parts, which is what makes this one field rather than two
// payment types.
type counterPaymentPart struct {
	cardID uuid.UUID
	kind   string
	n      int
}

// counterPayment is a validated CounterRemovalCost: the per-part
// split, each part naming its permanent, kind and count. The zero
// value (no parts) means the ability has no counter component.
type counterPayment struct {
	parts []counterPaymentPart
	total int
}

// counterPartKey is the identity of one part of a payment. A
// permanent may appear in two parts only with two different kinds
// (#943); naming the same (permanent, kind) twice is refused, so a
// fixed-kind payment still cannot name one permanent twice.
type counterPartKey struct {
	cardID uuid.UUID
	kind   string
}

// CounterCostPayment is the announce-time payment for a counter
// component, as it arrives from the wire. It is ONE shape for all
// five printed forms, and that is deliberate: the single-permanent
// forms are the one-part case of the among form, and a second wire
// shape for them would be a second thing to keep in step.
//
//   - SourceIDs names the permanents paid from. Empty is legal only
//     for the self form, where the source pays.
//   - Counts is the per-permanent split, parallel to SourceIDs.
//     Empty means "the printed count, from the one named permanent"
//     — the shape every #625 client already sends.
//   - Kind is the counter kind, for the any-kind form: ONE kind for
//     the whole payment.
//   - Kinds is #943's parallel array, one kind per part, for the
//     any-kind AMONG form where the parts may differ. It is an
//     EXTENSION of the triple above rather than a rival encoding of
//     it: a payment whose parts all share a kind still sends Kind and
//     no Kinds, so every client that already speaks this payload
//     keeps working untouched, and the array appears only on the one
//     printed cost that needs it.
type CounterCostPayment struct {
	SourceIDs []uuid.UUID
	Counts    []int
	Kind      string
	Kinds     []string
}

// empty reports a payment the client did not send at all.
func (p CounterCostPayment) empty() bool {
	return len(p.SourceIDs) == 0 && len(p.Counts) == 0 && p.Kind == "" && len(p.Kinds) == 0
}

// validateCounterRemovalLocked resolves a RemoveCounters component
// into the concrete removal to perform, without removing anything —
// the validate half of ADR 0020 §3's "validate everything, then pay
// everything". Shared by the activated-ability path and the mana-
// ability path, which is what makes "one type, two owners" true
// rather than aspirational.
//
// What it enforces, in the order the errors matter:
//
//   - A payment sent for an ability with no counter component is
//     rejected rather than ignored, as crew_ids and sacrifice_ids
//     are: a client that sends one is confused about which ability
//     it is firing.
//   - The self form takes no permanent, or the source's own ID.
//     The other and among forms name permanents, which must be on
//     the battlefield, controlled by the activator
//     (ErrCardCallerMismatch) and matched by From WITHOUT the
//     targeting gate (ErrIllegalTarget) — a cost does not target
//     (CR 601.2h).
//   - No (permanent, KIND) may be named twice. Naming one twice
//     would let a single artifact pay both halves of "remove two
//     +1/+1 counters from among artifacts you control". The kind is
//     part of the key only because #943's any-kind among cost lets
//     one permanent pay in two kinds; for a printed kind the rule is
//     the old one, one part per permanent.
//   - The counts must total exactly N — or at least N, for a
//     variable cost, where the announced total IS the payment.
//   - Every permanent must hold at least the count named against it,
//     else ErrInsufficientCounters. For the any-kind form that is
//     also the answer when the named kind simply is not on it.
//
// Caller must hold g.mu.
func (g *Game) validateCounterRemovalLocked(playerID, sourceID uuid.UUID, rc *CounterRemovalCost, pay CounterCostPayment) (counterPayment, error) {
	if rc == nil {
		if !pay.empty() {
			return counterPayment{}, ErrInvalidParam
		}
		return counterPayment{}, nil
	}
	if rc.N < 0 || (rc.N == 0 && !rc.Variable) {
		// Register refuses this at boot; an intrinsic ability built in
		// a test or a token template could still carry one.
		return counterPayment{}, ErrInvalidParam
	}

	// --- the permanents ------------------------------------------
	ids := pay.SourceIDs
	if rc.From == nil {
		switch {
		case len(ids) == 0:
			ids = []uuid.UUID{sourceID}
		case len(ids) == 1 && ids[0] == sourceID:
			// The client named the source explicitly; same thing.
		default:
			return counterPayment{}, ErrInvalidParam
		}
	} else if len(ids) == 0 {
		return counterPayment{}, ErrInvalidParam
	} else if !rc.Among && len(ids) != 1 {
		// Only an among cost is paid by more than one permanent.
		return counterPayment{}, ErrInvalidParam
	}

	// --- the split -----------------------------------------------
	counts := pay.Counts
	switch {
	case len(counts) == 0:
		// The #625 shape: one permanent, the printed count. A
		// variable cost has no printed count, so it must say.
		if rc.Variable || len(ids) != 1 {
			return counterPayment{}, ErrInvalidParam
		}
		counts = []int{rc.N}
	case len(counts) != len(ids):
		return counterPayment{}, ErrInvalidParam
	}

	// --- the kinds -----------------------------------------------
	kinds, err := counterPaymentKinds(rc, pay, len(ids))
	if err != nil {
		return counterPayment{}, err
	}

	total := 0
	seen := make(map[counterPartKey]bool, len(ids))
	parts := make([]counterPaymentPart, 0, len(ids))
	for i, id := range ids {
		n, kind := counts[i], kinds[i]
		if n < 1 {
			// A part that removes nothing is not a part. Refusing it
			// keeps the total honest and stops a client padding the
			// list with permanents it never paid from.
			return counterPayment{}, ErrInvalidParam
		}
		key := counterPartKey{cardID: id, kind: kind}
		if seen[key] {
			return counterPayment{}, ErrInvalidParam
		}
		seen[key] = true
		c := findBattlefieldCard(g, id)
		if c == nil {
			return counterPayment{}, ErrCardNotFound
		}
		// "a planeswalker you control": a cost is paid with your own
		// permanents. The source is already known to be the
		// activator's.
		if c.Controller != playerID {
			return counterPayment{}, ErrCardCallerMismatch
		}
		// specMatchLocked(..., false), not targetLegalLocked: choosing
		// a permanent to pay a cost does not target it (CR 601.2h /
		// 602.2b), so the CR 702 keyword gate must not apply.
		if rc.From != nil && !g.specMatchLocked(playerID, rc.From, TargetRef{Kind: TargetCard, ID: id}, false) {
			return counterPayment{}, ErrIllegalTarget
		}
		if c.Counters[kind] < n {
			return counterPayment{}, ErrInsufficientCounters
		}
		total += n
		parts = append(parts, counterPaymentPart{cardID: id, kind: kind, n: n})
	}

	// --- the total -----------------------------------------------
	if rc.Variable {
		// "Remove X counters" / "any number": the announcement is the
		// payment, bounded below by the printed floor and above by
		// what the permanents actually hold — which the per-part
		// check above has already enforced. There is deliberately no
		// ceiling of our own, for the same reason AbilityCost has no
		// MaxX.
		if total < rc.N {
			return counterPayment{}, ErrInvalidParam
		}
	} else if total != rc.N {
		return counterPayment{}, ErrInvalidParam
	}
	return counterPayment{parts: parts, total: total}, nil
}

// counterPaymentKinds resolves the counter kind each part of a
// payment removes — one per named permanent, in the same order
// (#943).
//
//	printed kind        rc.Counter, for every part.
//	any kind, one kind  pay.Kind, for every part — the #625 shape,
//	                    and still what a payment sends whenever its
//	                    parts happen to share a kind.
//	any kind, mixed     pay.Kinds, one per part. Only an any-kind
//	                    AMONG cost can produce this, because only it
//	                    asks the kind question more than once.
//
// A payment may send both fields; they must then agree, so the two
// can never describe different payments. An empty kind in the array
// is refused rather than defaulted: "a counter of no kind" is not a
// thing a permanent holds, and letting it fall through to
// Counters[""] would read a map entry no counter ever writes.
func counterPaymentKinds(rc *CounterRemovalCost, pay CounterCostPayment, n int) ([]string, error) {
	if len(pay.Kinds) > 0 {
		if len(pay.Kinds) != n {
			return nil, ErrInvalidParam
		}
		out := make([]string, n)
		for i, k := range pay.Kinds {
			if k == "" {
				return nil, ErrInvalidParam
			}
			if rc.Counter != "" && k != rc.Counter {
				return nil, ErrInvalidParam
			}
			if pay.Kind != "" && k != pay.Kind {
				return nil, ErrInvalidParam
			}
			out[i] = k
		}
		return out, nil
	}
	kind := rc.Counter
	if kind == "" {
		if pay.Kind == "" {
			return nil, ErrInvalidParam
		}
		kind = pay.Kind
	} else if pay.Kind != "" && pay.Kind != kind {
		return nil, ErrInvalidParam
	}
	out := make([]string, n)
	for i := range out {
		out[i] = kind
	}
	return out, nil
}

// validateCounterRemovalCostLocked is validateCounterRemovalLocked
// for an activated ability's whole AbilityCost. Kept as the named
// entry point the activation path reads, so the cost-component list
// in activated.go stays one call per component.
//
// Caller must hold g.mu.
func (g *Game) validateCounterRemovalCostLocked(playerID, sourceID uuid.UUID, cost AbilityCost, pay CounterCostPayment) (counterPayment, error) {
	return g.validateCounterRemovalLocked(playerID, sourceID, cost.RemoveCounters, pay)
}

// payCounterRemovalLocked removes the validated counters.
//
// applyCounterLocked, NOT AddCounterForEffect, and for the same reason
// the loyalty cost uses it: removing counters to pay a cost is not an
// effect and not a replaceable event (CR 614.1 / 118.3), so nothing
// that modifies counter placement or removal gets a say — the printed
// number comes off, no more and no fewer.
//
// It never touches LoyaltyActivatedThisTurn. Removing a loyalty counter
// from a planeswalker to pay a DIFFERENT permanent's cost is not
// activating a loyalty ability (CR 606.3 is about the walker's own
// abilities), so the walker can still activate one this turn. A walker
// left at 0 loyalty is put into its owner's graveyard by the CR 704.5i
// state-based action the activation runs on its way out — after the
// ability is on the stack.
//
// The parts come off in the order the activator named them, each in
// its own kind (#943: an any-kind among payment mixes kinds). They
// are all part of one payment, so nothing can happen between them.
//
// Caller must hold g.mu.
func (g *Game) payCounterRemovalLocked(pay counterPayment) error {
	for _, part := range pay.parts {
		if part.n <= 0 {
			continue
		}
		if err := g.applyCounterLocked(part.cardID, part.kind, -part.n); err != nil {
			return err
		}
	}
	return nil
}

// canPlaceCounterLocked reports CR 118.3 for an AddCounter cost: can
// this permanent have that counter put on it right now? A cost you
// cannot pay is a cost that stops the activation, so the answer is
// checked in the validate block and never at payment time.
//
// Today there is exactly one way to answer no — the permanent is not
// on the battlefield under the payer's control any more — because the
// engine models no "counters can't be put on …" prohibition
// (Solemnity, Melira). That is stated here rather than hidden: this
// predicate is the ONE place such a prohibition plugs in, and the
// refusal it drives is already wired through the validator, the legal
// enumerator, the protocol view and the client's greyed row. See the
// #789 addendum to ADR 0020.
//
// Note what this is NOT asking. A counter-placement REPLACEMENT
// (Doubling Season, Hardened Scales, a Solemnity-style cancel
// registered as one) is irrelevant here, because paying a cost is not
// an effect and the placement never enters the CR 614 pipeline — see
// payCounterAddLocked.
//
// Caller must hold g.mu (read or write).
func (g *Game) canPlaceCounterLocked(playerID, cardID uuid.UUID, ac *CounterAddCost) bool {
	if ac == nil {
		return true
	}
	if ac.Counter == "" || ac.N < 1 {
		return false
	}
	c := findBattlefieldCard(g, cardID)
	return c != nil && c.Controller == playerID
}

// CanPlaceCounterForEffect is canPlaceCounterLocked on the *ForEffect
// surface, for the readers that have to agree with the activation
// path without being it: the legal-move enumerator (so a bot is never
// offered an activation the engine refuses — #544) and the protocol
// view (so the client greys the row for the same reason the server
// would bounce it).
//
// Caller must hold g.mu (read or write).
func (g *Game) CanPlaceCounterForEffect(playerID, cardID uuid.UUID, ac *CounterAddCost) bool {
	return g.canPlaceCounterLocked(playerID, cardID, ac)
}

// payCounterAddLocked puts the cost's counter on the source.
//
// applyCounterLocked, not AddCounterForEffect, for the third time in
// this file and for the same rule: a cost is not an effect (CR
// 121.1), so a Doubling Season does NOT double the -1/-1 counter a
// Devoted Druid puts on itself to untap. Routing it through the CR
// 614 pipeline would silently make it — and it would make the
// untapper cost twice as much under a card that is supposed to be
// pure upside, which is the kind of wrong that takes a week to find.
//
// Caller must hold g.mu.
func (g *Game) payCounterAddLocked(cardID uuid.UUID, ac *CounterAddCost) error {
	if ac == nil || ac.N < 1 {
		return nil
	}
	return g.applyCounterLocked(cardID, ac.Counter, ac.N)
}
