package game

import (
	"sort"

	"github.com/google/uuid"
)

// counter_cost.go — #625: removing counters as part of paying an
// activated ability's cost (CR 602.2b / 118.3).
//
// Before this, AbilityCost could ADD or REMOVE loyalty on the source
// (a loyalty ability's +N / −N) and nothing else touched a counter.
// Every card whose printed cost says "Remove a counter" either left
// the ability off (Dragon's Hoard, Mikaeus, Benevolent Hydra, Fain,
// Iron Spider) or left out the alternative that costs one (Heart of
// Kiran). Deferring the removal to resolution was never an option: a
// proliferate or a second activation in response would bank a counter
// the printed card had already spent, which is the stronger-than-
// printed direction #259 forbids.
//
// One component covers three printed shapes:
//
//	self       "Remove a gold counter from this artifact"   From == nil
//	other      "remove a loyalty counter from a planeswalker you control"
//	any kind   "Remove a counter from a creature you control"  Counter == ""
//
// What it deliberately does not cover: a removal SPLIT across several
// permanents ("Remove two +1/+1 counters from among artifacts you
// control" — Iron Spider, Stark Upgrade), and a cost that ADDS a
// counter (Devoted Druid's "Put a -1/-1 counter on this creature").
// Both are still open seams; docs/engine-seams.md says so.

// CounterRemovalCost is the "remove N counters" component of an
// activated ability's cost. See AbilityCost.RemoveCounters.
type CounterRemovalCost struct {
	// Counter is the kind removed ("loyalty", "+1/+1", "gold"). Empty
	// means "a counter" of ANY kind, and the activator names the kind
	// at announce alongside the permanent (Fain, the Broker). An
	// any-kind cost removes N counters of the ONE kind chosen; a
	// printed "remove two counters" that may mix kinds has no shape,
	// and effects.Register refuses N > 1 with an empty Counter so a
	// card file cannot claim one.
	Counter string

	// N is how many counters are removed. Always at least 1;
	// effects.Register panics at boot otherwise.
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

// CounterCostOptionsForEffect is the set of (permanent, kind) choices
// that would pay `rc` for an activation of an ability on `sourceID` by
// `playerID`, ordered most counters first (ties keep battlefield
// order). Empty when the cost cannot be paid.
//
// ONE walk shared by the protocol view (the client's picker), the
// legal-move enumerator (the bots) and — through the same predicates —
// validateCounterRemovalCostLocked, so the three cannot disagree about
// which permanent pays. It is the NON-targeting candidate walk
// (SpecCandidatesForEffect): choosing a permanent to pay a cost does
// not target it, so a hexproof planeswalker still pays Heart of
// Kiran's crew.
//
// Caller must hold g.mu (read or write).
func (g *Game) CounterCostOptionsForEffect(playerID, sourceID uuid.UUID, rc *CounterRemovalCost) []CounterCostOption {
	if rc == nil || rc.N <= 0 {
		return nil
	}
	var ids []uuid.UUID
	if rc.From == nil {
		ids = []uuid.UUID{sourceID}
	} else {
		ids = g.specCandidatesLocked(playerID, rc.From).Cards
	}
	var out []CounterCostOption
	for _, id := range ids {
		c := findBattlefieldCard(g, id)
		if c == nil || c.Controller != playerID {
			continue
		}
		kinds := counterKindsPaying(c, rc)
		if len(kinds) == 0 {
			continue
		}
		out = append(out, CounterCostOption{CardID: id, Kinds: kinds})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].most() > out[j].most() })
	return out
}

// counterKindsPaying lists the kinds on c that could pay rc, most
// counters first and then by name so the order is deterministic.
func counterKindsPaying(c *Card, rc *CounterRemovalCost) []CounterCostKind {
	if rc.Counter != "" {
		if n := c.Counters[rc.Counter]; n >= rc.N {
			return []CounterCostKind{{Kind: rc.Counter, Count: n}}
		}
		return nil
	}
	var out []CounterCostKind
	for kind, n := range c.Counters {
		if kind == "" || n < rc.N {
			continue
		}
		out = append(out, CounterCostKind{Kind: kind, Count: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}

// counterPayment is a validated CounterRemovalCost: the permanent and
// kind to remove from and how many. The zero value (N == 0) means the
// ability has no counter component.
type counterPayment struct {
	cardID uuid.UUID
	kind   string
	n      int
}

// validateCounterRemovalCostLocked resolves a RemoveCounters
// component into the concrete removal to perform, without removing
// anything — the validate half of ADR 0020 §3's "validate everything,
// then pay everything".
//
// The announce-time choices it reads:
//
//   - `chosen` (ActivateAbilityParams.CounterSourceIDs). For the
//     other-permanent form exactly one ID, which must be on the
//     battlefield, controlled by the activator (ErrCardCallerMismatch)
//     and matched by From without the targeting gate
//     (ErrIllegalTarget). For the self form it may be omitted, or name
//     the source; anything else is a confused client.
//   - `kind` (ActivateAbilityParams.CounterKind). Required for an
//     any-kind cost. For a fixed-kind cost it may be omitted or repeat
//     the printed kind.
//
// The permanent must hold at least N counters of the kind, else
// ErrInsufficientCounters — which, for the any-kind form, is also the
// answer when the named kind simply is not on it.
//
// A payload that sends either field for an ability with no counter
// component is rejected rather than ignored, as crew_ids and
// sacrifice_ids are.
//
// Caller must hold g.mu.
func (g *Game) validateCounterRemovalCostLocked(playerID, sourceID uuid.UUID, cost AbilityCost, chosen []uuid.UUID, kind string) (counterPayment, error) {
	rc := cost.RemoveCounters
	if rc == nil {
		if len(chosen) > 0 || kind != "" {
			return counterPayment{}, ErrInvalidParam
		}
		return counterPayment{}, nil
	}
	if rc.N <= 0 {
		// Register refuses this at boot; an intrinsic ability built in
		// a test or a token template could still carry one.
		return counterPayment{}, ErrInvalidParam
	}

	var id uuid.UUID
	if rc.From == nil {
		switch {
		case len(chosen) == 0:
			id = sourceID
		case len(chosen) == 1 && chosen[0] == sourceID:
			id = sourceID
		default:
			return counterPayment{}, ErrInvalidParam
		}
	} else {
		if len(chosen) != 1 {
			return counterPayment{}, ErrInvalidParam
		}
		id = chosen[0]
	}
	c := findBattlefieldCard(g, id)
	if c == nil {
		return counterPayment{}, ErrCardNotFound
	}
	// "a planeswalker you control": a cost is paid with your own
	// permanents. The source is already known to be the activator's.
	if c.Controller != playerID {
		return counterPayment{}, ErrCardCallerMismatch
	}
	// specMatchLocked(..., false), not targetLegalLocked: choosing a
	// permanent to pay a cost does not target it (CR 601.2h / 602.2b),
	// so the CR 702 keyword gate must not apply.
	if rc.From != nil && !g.specMatchLocked(playerID, rc.From, TargetRef{Kind: TargetCard, ID: id}, false) {
		return counterPayment{}, ErrIllegalTarget
	}

	k := rc.Counter
	if k == "" {
		if kind == "" {
			return counterPayment{}, ErrInvalidParam
		}
		k = kind
	} else if kind != "" && kind != k {
		return counterPayment{}, ErrInvalidParam
	}
	if c.Counters[k] < rc.N {
		return counterPayment{}, ErrInsufficientCounters
	}
	return counterPayment{cardID: id, kind: k, n: rc.N}, nil
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
// state-based action ActivateCatalogAbility runs at the end of the
// activation — after the ability is on the stack.
//
// Caller must hold g.mu.
func (g *Game) payCounterRemovalLocked(pay counterPayment) error {
	if pay.n <= 0 {
		return nil
	}
	return g.applyCounterLocked(pay.cardID, pay.kind, -pay.n)
}
