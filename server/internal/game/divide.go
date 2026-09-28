package game

import (
	"fmt"

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
//     announced X;
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
// assignAnnouncedSlots and validateAnnouncedTargetsLocked. Nil in, nil
// out, for the announcement that divides nothing.
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
		if sum != total {
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
