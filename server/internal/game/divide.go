package game

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// divide.go — #1563, CR 601.2d / CR 700.2i: "N damage divided as you
// choose among …" (ADR 0065's 2026-09-27 amendment on divided effects).
//
// CR 601.2d: "If the spell requires the player to divide or distribute
// an effect (such as damage or counters) as he or she casts it, the
// player announces the division. Each of these targets must receive at
// least one of whatever is being divided." CR 603.3d and CR 602.2b
// apply the same rule to a triggered ability as it is put on the stack
// and to an activated ability as it is activated.
//
// The division is an ANNOUNCE-TIME choice, like X and the targets. It
// rides StackItem.Distribution (keyed by target id), is carried by
// clone and snapshot, remapped with its slot by a CR 115.7 retarget,
// and copied with the rest of the announcement by CR 707.10. At
// resolution each target that is still legal takes exactly its share;
// a target that became illegal takes nothing and its share is NOT
// redistributed (CR 608.2b — the illegal target is simply not
// affected).
//
// The clause says the division exists and how big it is. The card's
// resolution decides what is divided — damage today (DealDividedDamage
// in the effects package), counters when a card needs it — which is why
// DivideSpec carries only the amount.

// DivideSpec is a clause's "divided as you choose" amount. It is DATA,
// not a func: a trigger's clause is reachable from Game through the
// pick_target resume frame, and a func-typed field there would be a
// new closure route (ADR 0041 phase 3's ratchet). The three shapes
// every printed divided card uses are all expressible:
//
//	Total 4                      Fury, Pyrokinesis
//	FromX                        Rolling Thunder, Fire Covenant
//	FromX, DoubleFromX 6         Shatterskull Smashing — "If X is 6
//	                             or more, … deals twice X damage
//	                             divided as you choose among them
//	                             instead."
type DivideSpec struct {
	// Total is the fixed amount divided. Ignored when FromX.
	Total int

	// FromX makes the amount the X announced at CR 601.2b / 602.2b.
	// Only an announcement with an X may declare it: a spell's or an
	// activated ability's clause (effects.Register refuses it on a
	// trigger).
	FromX bool

	// DoubleFromX, with FromX, is the X at and above which the amount
	// is twice X. 0 means never.
	DoubleFromX int

	// AmountKey names a registered amount RULE (#1657) — "X is the
	// number of lands you control" (Ureni, the Song Unending), "damage
	// equal to its power" (Orca, Siege Demon), "up to that many" where
	// that many is the life you gained this turn (Lathiel), "2, or X
	// if this spell's madness cost was paid" (Avacyn's Judgment). When
	// set it REPLACES Total / FromX / DoubleFromX: the rule is read
	// once, at announce (CR 601.2d — the division is announced with
	// the targets, so its amount must be known then), and answers
	// with one of the data shapes above, which the rest of the engine
	// already reads. See divideAmountLocked.
	//
	// A key, not a func, for the reason ModeCountCondition and
	// LifeCostCount are keys: a trigger's clause is reachable from
	// Game through the pick_target resume frame, and ADR 0041 phase
	// 3's closure ratchet admits no new func-typed route.
	AmountKey DivideAmount

	// UpTo is "distribute UP TO that many" (Lathiel, the Bounteous
	// Dawn): the shares may add up to LESS than the amount, never
	// more. Each chosen target still receives at least 1 — CR 601.2d,
	// and Lathiel's own ruling: "Each target must receive at least one
	// +1/+1 counter." Whether zero targets is legal is the clause's
	// Min, as it is for every clause.
	UpTo bool
}

// DivideAmount names a registered divided-amount rule (#1657). The
// key is unexported, so the only way to hold a non-zero one is
// RegisterDivideAmount — a func literal on a DivideSpec does not
// compile.
type DivideAmount struct{ key string }

// Key is the rule's registry key.
func (a DivideAmount) Key() string { return a.key }

// IsZero reports whether no rule is named.
func (a DivideAmount) IsZero() bool { return a.key == "" }

// DivideAmountArgs is everything an amount rule may read at announce.
type DivideAmountArgs struct {
	// Controller is the player announcing: the caster, the activator,
	// or the controller of the trigger being put on the stack — "you"
	// in "the number of lands you control".
	Controller uuid.UUID

	// Source is the spell's or ability's source object.
	Source uuid.UUID

	// SourceLKI is a trigger's source as it last existed when the
	// ability triggered (CR 603.10) — "its power" on a dies trigger
	// (Orca). Nil for a cast or an activation, whose source is live.
	SourceLKI *Characteristic

	// AltCost is the alternative cost claimed at announce (CR 118.9),
	// the key that lands on StackItem.AltCost — "if this spell's
	// madness cost was paid". Empty for a cast that paid its mana
	// cost, and always for an ability.
	AltCost string
}

// DivideAmountFunc computes a divided amount. It answers in the DATA
// shapes — a Total, or FromX (with DoubleFromX) — so the announced X
// is still applied by TotalFor, and the client applies it to the X it
// collected; AmountKey and UpTo on the answer are ignored (UpTo is the
// clause's). It is read-only and runs under g.mu, so it must not call
// a public locking accessor. A negative Total is read as 0.
type DivideAmountFunc func(g *Game, a DivideAmountArgs) DivideSpec

var divideAmounts = struct {
	sync.RWMutex
	byKey map[string]DivideAmountFunc
}{byKey: map[string]DivideAmountFunc{}}

// RegisterDivideAmount registers a divided-amount rule under `key` and
// returns its name. Call it once, from a package-level var. Panics on
// an empty key, a nil function or a duplicate — each a card-file bug
// that would otherwise ship a division the engine cannot size.
func RegisterDivideAmount(key string, fn DivideAmountFunc) DivideAmount {
	if key == "" {
		panic("game: RegisterDivideAmount with an empty key")
	}
	if fn == nil {
		panic(fmt.Sprintf("game: divide amount %q has no function", key))
	}
	divideAmounts.Lock()
	defer divideAmounts.Unlock()
	if _, dup := divideAmounts.byKey[key]; dup {
		panic(fmt.Sprintf("game: divide amount %q registered twice", key))
	}
	divideAmounts.byKey[key] = fn
	return DivideAmount{key: key}
}

// divideAmountLocked is the clause's division with its amount rule
// read NOW — the one place AmountKey is evaluated (#1657). A spec
// without a rule is returned as is; one with a rule comes back as a
// fresh spec in the data shape the rule answered with, carrying the
// clause's UpTo and no key, so every later reader (the gate, the even
// split, the view, the enumerator's cap) sees a plain amount. A key
// that is not registered sizes the division at 0 — no target can be
// chosen — rather than guessing, the weaker reading.
//
// Caller must hold g.mu.
func (g *Game) divideAmountLocked(d *DivideSpec, a DivideAmountArgs) *DivideSpec {
	if d == nil || d.AmountKey.IsZero() {
		return d
	}
	divideAmounts.RLock()
	fn, ok := divideAmounts.byKey[d.AmountKey.key]
	divideAmounts.RUnlock()
	out := DivideSpec{UpTo: d.UpTo}
	if ok {
		r := fn(g, a)
		out.Total, out.FromX, out.DoubleFromX = r.Total, r.FromX, r.DoubleFromX
		if out.Total < 0 {
			out.Total = 0
		}
	}
	return &out
}

// DivideAmountForEffect is divideAmountLocked on the *ForEffect
// surface, for the protocol projection: the amount the view quotes is
// the one the gate would fix if the announcement were made now.
func (g *Game) DivideAmountForEffect(d *DivideSpec, a DivideAmountArgs) *DivideSpec {
	return g.divideAmountLocked(d, a)
}

// bindDivideAmountsLocked fixes every divided step's amount rule at
// announce (#1657, CR 601.2d): each step whose clause names an
// AmountKey has its Divide replaced by the rule's answer. The steps
// hold clause COPIES (AnnouncedClauses) and the replacement is a new
// spec, so the catalog's shared clause is never touched. Called at the
// three announce points — a cast, an activation, and the moment a
// trigger's target walk opens — before anything reads the amount, and
// never again: a land that leaves, or life gained in response, changes
// nothing about a division already announced (Ureni's and Lathiel's
// rulings).
//
// Caller must hold g.mu.
func (g *Game) bindDivideAmountsLocked(steps []AnnouncedClause, a DivideAmountArgs) {
	for i := range steps {
		if d := steps[i].Clause.Divide; d != nil && !d.AmountKey.IsZero() {
			steps[i].Clause.Divide = g.divideAmountLocked(d, a)
		}
	}
}

// BindDivideAmountsForEffect is bindDivideAmountsLocked for the bot's
// move enumerator, which must size a division exactly as the gate it
// is about to be judged by does (#544).
func (g *Game) BindDivideAmountsForEffect(steps []AnnouncedClause, a DivideAmountArgs) {
	g.bindDivideAmountsLocked(steps, a)
}

// TotalFor is the amount divided under an announced X. Nil-safe (0).
func (d *DivideSpec) TotalFor(x int) int {
	if d == nil {
		return 0
	}
	if !d.FromX {
		return d.Total
	}
	if x < 0 {
		x = 0
	}
	if d.DoubleFromX > 0 && x >= d.DoubleFromX {
		return 2 * x
	}
	return x
}

// Dividing marks the clause "divided as you choose among" its targets.
// Mutates and returns the receiver, as WithCount does.
func (s *TargetSpec) Dividing(d DivideSpec) *TargetSpec {
	s.Divide = &d
	return s
}

// errDivision is the announce-time refusal for a bad division. It
// wraps ErrInvalidParam, so the wire reports bad_request with the
// sentence after the colon.
func errDivision(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidParam, fmt.Sprintf(format, args...))
}

// dividedStepIDs lists the target ids that answer one step, in announce
// order, skipping the Self / None placeholders.
func dividedStepIDs(step AnnouncedClause, targets []TargetRef) []uuid.UUID {
	var ids []uuid.UUID
	for _, t := range targets {
		if t.Kind == TargetSelf || t.Kind == TargetNone {
			continue
		}
		if t.Mode == step.Mode && t.Slot == step.Slot {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

// settleDistribution is the CR 601.2d gate: it checks an announcement's
// division against its divided clauses and returns the division to
// store on the stack item.
//
// For every step whose clause divides:
//
//   - each target of that step is assigned at least 1;
//   - the shares sum to exactly the clause's amount under the
//     announced X — or, on an UpTo clause ("distribute up to that
//     many", #1657), to at most that amount;
//   - a step with more targets than the amount cannot be announced at
//     all, since some target would get 0;
//   - a step with ONE target and no share named for it is given the
//     whole amount — there is nothing to choose, so a client (or a
//     test, or gamecli) need not say it.
//
// A share naming anything that is not a target of a divided step is
// refused, as is a division on an announcement with no divided clause
// at all: both are a client confused about what it is dividing, and
// quietly dropping the share would resolve something the player did
// not announce. The same object in two divided steps is refused too,
// because the division is keyed by id and could not tell them apart
// (no printed card divides across two clauses).
//
// An announcement with NO structured clause list — a card the catalog
// does not know, on the S13.1 free-form path — keeps whatever division
// the client sent, unjudged, exactly as that path keeps its targets:
// the sandbox records the choice for the table to resolve by hand.
//
// `targets` must already carry their (Mode, Slot) — call it after
// assignAnnouncedSlots and validateAnnouncedTargetsLocked — and every
// step's amount rule must already be fixed by bindDivideAmountsLocked
// (#1657). Nil in, nil out, for the announcement that divides nothing.
func settleDistribution(steps []AnnouncedClause, targets []TargetRef, dist map[uuid.UUID]int, x int) (map[uuid.UUID]int, error) {
	if len(steps) == 0 {
		return dist, nil
	}
	var out map[uuid.UUID]int
	owned := make(map[uuid.UUID]bool)
	for i := range steps {
		d := steps[i].Clause.Divide
		if d == nil {
			continue
		}
		ids := dividedStepIDs(steps[i], targets)
		if len(ids) == 0 {
			continue
		}
		total := d.TotalFor(x)
		if len(ids) > total {
			return nil, errDivision("%d targets cannot divide %d — each target must be assigned at least 1", len(ids), total)
		}
		if out == nil {
			out = make(map[uuid.UUID]int, len(ids))
		}
		if len(ids) == 1 {
			if _, named := dist[ids[0]]; !named {
				if owned[ids[0]] {
					return nil, errDivision("the same target is in two divided clauses")
				}
				owned[ids[0]] = true
				out[ids[0]] = total
				continue
			}
		}
		sum := 0
		for _, id := range ids {
			if owned[id] {
				return nil, errDivision("the same target is in two divided clauses")
			}
			owned[id] = true
			v := dist[id]
			if v < 1 {
				return nil, errDivision("each target must be assigned at least 1 of the %d", total)
			}
			sum += v
			out[id] = v
		}
		if d.UpTo {
			if sum > total {
				return nil, errDivision("the division must add up to at most %d, not %d", total, sum)
			}
		} else if sum != total {
			return nil, errDivision("the division must add up to %d, not %d", total, sum)
		}
	}
	for id := range dist {
		if !owned[id] {
			return nil, errDivision("a share was assigned to something that is not a target of the division")
		}
	}
	return out, nil
}

// EvenDistribution is the legal DEFAULT division for an announcement:
// each divided step's amount split as evenly as possible across its
// targets in announce order, the remainder going one point at a time to
// the earliest. It is what the bot's enumerator announces, so every
// divided move it offers passes settleDistribution (#544).
//
// ok is false when some divided step has more targets than its amount
// — an announcement the gate refuses, which the enumerator must not
// offer. A nil map with ok true is an announcement that divides
// nothing (no divided clause, or a divided clause with no targets).
func EvenDistribution(steps []AnnouncedClause, targets []TargetRef, x int) (map[uuid.UUID]int, bool) {
	var out map[uuid.UUID]int
	for i := range steps {
		d := steps[i].Clause.Divide
		if d == nil {
			continue
		}
		ids := dividedStepIDs(steps[i], targets)
		if len(ids) == 0 {
			continue
		}
		total := d.TotalFor(x)
		if len(ids) > total {
			return nil, false
		}
		if out == nil {
			out = make(map[uuid.UUID]int, len(ids))
		}
		share, extra := total/len(ids), total%len(ids)
		for j, id := range ids {
			if _, dup := out[id]; dup {
				return nil, false
			}
			out[id] = share
			if j < extra {
				out[id]++
			}
		}
	}
	return out, true
}

// StepsDivide reports whether any step of an announcement divides — the
// enumerator's cheap test before it computes a default.
func StepsDivide(steps []AnnouncedClause) bool {
	for i := range steps {
		if steps[i].Clause.Divide != nil {
			return true
		}
	}
	return false
}

// PickTargetDivideForEffect is the divided amount of the clause a
// pick_target prompt is asking about, for the view and the enumerator:
// the clause's DivideSpec, or nil when it divides nothing (or the
// prompt is not a trigger's target walk). A trigger announces no X, so
// the amount is TotalFor(0).
func PickTargetDivideForEffect(c *PendingChoice) *DivideSpec {
	if c == nil || c.Kind != PendingChoicePickTarget || c.pickTargetResume == nil {
		return nil
	}
	clause := c.pickTargetResume.currentClause()
	if clause == nil {
		return nil
	}
	return clause.Divide
}

// PickTargetDefaultDistribution is EvenDistribution for one answer to a
// trigger's pick_target prompt: the refs answer the prompt's current
// step. Nil when the step divides nothing.
func PickTargetDefaultDistribution(c *PendingChoice, targets []TargetRef) (map[uuid.UUID]int, bool) {
	if PickTargetDivideForEffect(c) == nil {
		return nil, true
	}
	f := c.pickTargetResume
	cur := f.steps[f.step]
	stamped := make([]TargetRef, 0, len(targets))
	for _, t := range targets {
		t.Mode, t.Slot = cur.Mode, cur.Slot
		stamped = append(stamped, t)
	}
	return EvenDistribution(f.steps[f.step:f.step+1], stamped, 0)
}
