package game

import (
	"strings"

	"github.com/google/uuid"
)

// autotap_trigger_mana.go is ADR 0074's amendment of 2026-10-30 (#4716):
// the auto-tap PLANNER counts the mana triggered mana abilities add when
// it taps a source (CR 605.1b, 605.4a), and judges each one by that
// trigger's own condition.
//
// Before it, a Swamp under Crypt Ghast was planned as one {B}. In the
// reported game the planner found five Swamps one mana short of
// Veinripper's {3}{B}{B}{B}, tapped all five, cracked a Treasure, and the
// executor then fired the real trigger and floated the surplus. Three
// Swamps paid it. And a cast only the trigger mana makes payable had no
// plan at all, so strict payment refused it, the move list did not offer
// it, and the ready ring and the bots never saw it.
//
// # One judge, two readers
//
// The planner asks the judge the executor fires with,
// manaTriggerFiringsLocked, about the production the executor will make
// when it taps the source: the permanent as a copy, marked tapped
// (`Source`), the payer (`Controller`), and the colours the source's
// priced slots make (`Colors`, after the CR 106.12b window). Whatever the
// judge says fires becomes TRIGGER SLOTS appended to the candidate
// (tapSource.TriggerSlots), priced through the CR 614 window as not a tap
// (fromTap false, CR 106.12b: Mana Reflection does not double Wild
// Growth's {G}). The solver, the generic recruit, the tiers and the hand
// demand all read slots, so they see the trigger's with no change of
// their own.
//
// # A pick the trigger reads
//
// Mirari's Wake adds "one mana of any type that land produced", so on a
// {W|U} dual what it adds depends on the pick. Modelling it as a second
// {W|U} slot would let one dual pay {W}{U}, which is stronger than
// printed. So each combination of the source's picks is judged
// (maxTriggerPickCombos bounds it):
//
//   - every combination gets the same extra: the trigger slots are
//     appended to the one candidate (Crypt Ghast on a Watery Grave);
//   - they differ: the source becomes one candidate per combination, its
//     picks fixed, mutually exclusive under tapPlan.hasCard (#779's
//     expansion). One pick is carried as OneColor, which the executor
//     finds with firstSlotOffering; two or more as PinnedSlots, the
//     per-slot colours of plannedTap.Slots;
//   - above the cap, or a shape whose picks the plan cannot carry: the
//     source gets no trigger mana at all, the weaker direction.
//
// # Booked, then made
//
// A trigger slot with a pick (Fertile Ground's "any color") is booked by
// the solver like any other, and the colour it booked travels on the plan
// (plannedTap.TriggerColors) to the trigger's adder, which makes it
// rather than a greedy pick (stampTriggerColors, fireManaTriggersLocked).
//
// # And checked
//
// The planner judges every source against the board before the payment;
// the executor judges each one at its own tap. They differ only when an
// earlier step of the same payment changes a trigger's condition (a
// sacrifice in the plan that removes the trigger's permanent). The clone
// check catches that, and autoTapFundedLocked plans once more without
// trigger mana, which is exactly the planner before this file.

// maxTriggerPickCombos bounds how many combinations of one source's
// picks are judged (ADR 0074 amendment, owner decision 1): every source
// with two picks fits (a filter land is 4, two five-colour picks 25).
// Above it the source is planned with no trigger mana.
const maxTriggerPickCombos = 36

// triggerMana is what the triggered mana abilities add to one
// production, as the planner books it.
type triggerMana struct {
	// slots is the mana, parsed from each firing's output and priced
	// through the CR 614 window as not a tap, in firing order.
	slots []ProducedManaEntry
	// from is the firing's source for each slot.
	from []uuid.UUID
	// key names the firings, so two productions the judge answers alike
	// compare equal.
	key string
}

// triggerManaLocked asks the judge about `prod` and returns what its
// firings add, as slots. A firing whose output does not parse adds
// nothing, as it adds nothing in the executor (addTriggeredManaLocked
// logs the error and mints none).
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) triggerManaLocked(prod ManaProduced, priceProduction bool) triggerMana {
	var tm triggerMana
	var key strings.Builder
	for _, f := range g.manaTriggerFiringsLocked(prod) {
		key.WriteString(f.source.String())
		key.WriteByte('=')
		key.WriteString(f.produced)
		key.WriteByte(';')
		slots, err := ParseProducedMana(f.produced)
		if err != nil {
			continue
		}
		// CR 106.12b: the trigger's own mana is not a tap for mana, so a
		// tap-keyed doubler does not see it (Mana Reflection's ruling); a
		// replacement that does apply to it is counted. Priced against the
		// trigger's permanent, which is the source its token names.
		if priceProduction {
			slots = g.priceProducedSlotsLocked(prod.Controller, f.source, slots, false)
		}
		for _, s := range slots {
			if len(s.Options) == 0 {
				continue
			}
			tm.slots = append(tm.slots, s)
			tm.from = append(tm.from, f.source)
		}
	}
	tm.key = key.String()
	return tm
}

// appendTriggeredTapSourceLocked is appendTapSource for a source the
// payment taps while a triggered mana ability is on the board: each
// candidate appendTapSource makes, with the trigger mana its tap would
// add (withTriggerManaLocked).
//
// `c` is the permanent as the gather read it, `printed` whether its
// slots are the ones the activation prints (manaPlannableSlotsShapeLocked).
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) appendTriggeredTapSourceLocked(
	out []tapSource,
	proto tapSource,
	slots []ProducedManaEntry,
	c Card,
	controller uuid.UUID,
	priceProduction, printed bool,
) []tapSource {
	// The permanent as the executor copies it (tappedForMana): after the
	// tap, which the payment makes before it produces.
	tapped := c
	tapped.Tapped = true
	for _, cand := range appendTapSource(nil, proto, slots) {
		out = g.withTriggerManaLocked(out, cand, tapped, controller, priceProduction, printed)
	}
	return out
}

// withTriggerManaLocked appends `cand` with the trigger mana tapping it
// adds, expanded per pick when that mana depends on the pick (see the
// file comment for the three outcomes).
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) withTriggerManaLocked(
	out []tapSource,
	cand tapSource,
	tapped Card,
	controller uuid.UUID,
	priceProduction, printed bool,
) []tapSource {
	var picks []int
	combos := 1
	for i, slot := range cand.Slots {
		if len(slot.Options) < 2 {
			continue
		}
		picks = append(picks, i)
		combos *= len(slot.Options)
		if combos > maxTriggerPickCombos {
			// Owner decision 1: above the cap the source gets no trigger
			// mana, and is planned as it was before the amendment.
			return append(out, cand)
		}
	}
	type outcome struct {
		choice []string
		mana   triggerMana
	}
	outcomes := make([]outcome, 0, combos)
	at := make([]int, len(picks))
	for {
		// One mana per slot: appendTapSource has already flattened every
		// "N of one color" and "different colors" slot into fixed ones,
		// so what is left is a fixed colour or a plain pick.
		colors := make([]string, len(cand.Slots))
		for k, slot := range cand.Slots {
			colors[k] = slot.Options[0]
		}
		choice := make([]string, len(picks))
		for j, i := range picks {
			choice[j] = cand.Slots[i].Options[at[j]]
			colors[i] = choice[j]
		}
		outcomes = append(outcomes, outcome{
			choice: choice,
			mana:   g.triggerManaLocked(ManaProduced{Source: tapped, Controller: controller, Colors: colors}, priceProduction),
		})
		j := len(picks) - 1
		for ; j >= 0; j-- {
			at[j]++
			if at[j] < len(cand.Slots[picks[j]].Options) {
				break
			}
			at[j] = 0
		}
		if j < 0 {
			break
		}
	}
	same := true
	for _, o := range outcomes[1:] {
		if o.mana.key != outcomes[0].mana.key {
			same = false
			break
		}
	}
	if same {
		return append(out, cand.withTriggerMana(outcomes[0].mana))
	}
	switch {
	case cand.OneColor != "":
		// A "N mana of one color" candidate with a further pick is a
		// shape no printed card has; its slots are no longer the printed
		// ones, so the picks cannot be carried.
	case len(picks) == 1:
		for _, o := range outcomes {
			c := cand.withPicksFixed(picks, o.choice)
			c.OneColor = o.choice[0]
			out = append(out, c.withTriggerMana(o.mana))
		}
		return out
	case printed:
		for _, o := range outcomes {
			c := cand.withPicksFixed(picks, o.choice)
			c.PinnedSlots = make([]string, len(cand.Slots))
			for j, i := range picks {
				c.PinnedSlots[i] = o.choice[j]
			}
			out = append(out, c.withTriggerMana(o.mana))
		}
		return out
	}
	return append(out, cand)
}

// withPicksFixed is the candidate with each pick slot narrowed to the
// colour `choice` names for it.
func (s tapSource) withPicksFixed(picks []int, choice []string) tapSource {
	slots := append([]ProducedManaEntry(nil), s.Slots...)
	for j, i := range picks {
		slots[i] = ProducedManaEntry{Options: []string{choice[j]}}
	}
	s.Slots = slots
	return s
}

// withTriggerMana is the candidate with `tm`'s slots appended as its
// trigger slots.
func (s tapSource) withTriggerMana(tm triggerMana) tapSource {
	if len(tm.slots) == 0 {
		return s
	}
	slots := make([]ProducedManaEntry, 0, len(s.Slots)+len(tm.slots))
	slots = append(slots, s.Slots...)
	s.Slots = append(slots, tm.slots...)
	s.TriggerSlots = len(tm.slots)
	s.TriggerFrom = append([]uuid.UUID(nil), tm.from...)
	return s
}

// stampTriggerColors copies what the solve booked each planned source's
// trigger slots for onto its plan entry (plannedTap.TriggerColors), ""
// for a slot left to generic. A permanent has one candidate in a plan
// (tapPlan.hasCard), so the entry is the one with its ID.
func stampTriggerColors(plan tapPlan, sources []tapSource, used []bool, booked [][]string) {
	for i := range sources {
		s := &sources[i]
		if !used[i] || s.TriggerSlots == 0 {
			continue
		}
		for p := range plan {
			if plan[p].CardID == s.CardID {
				plan[p].TriggerColors = append([]string(nil), booked[i][s.ownSlots():]...)
				break
			}
		}
	}
}

// countsTriggerMana reports whether any entry of the plan counted trigger
// mana. The payers' fallback asks the gate instead (autoTapFundedLocked),
// which is cheaper to be wrong about; the tests ask this.
func (p tapPlan) countsTriggerMana() bool {
	for _, t := range p {
		if len(t.TriggerColors) > 0 {
			return true
		}
	}
	return false
}
