package game

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// trigger_mana_view.go is #4770, owner decision 2 of ADR 0074's
// amendment of 2026-10-30: the VIEW names the mana tapping a permanent
// for one of its mana abilities would add through triggered mana
// abilities (Crypt Ghast's {B} on a Swamp, Mirari's Wake and Zendikar
// Resurgent's "one mana of any type that land produced", Mana Flare,
// Wild Growth), so a reader without a game handle (the heuristic bot's
// turn plan, ADR 0136 §2) can count it.
//
// A mana ability's own `produced` stays what CR 106.7 says the ability
// "could produce", which never includes a trigger's mana. This is a
// second answer beside it, from the auto-tap planner's own judge
// (withTriggerManaLocked, which asks manaTriggerFiringsLocked about the
// production the executor would make), so the view, the planner and the
// payment cannot disagree about what a trigger adds.
//
// It carries only the case in which every combination of the ability's
// picks gets the same extra mana (the issue's sub-decision 1): a Swamp
// under Crypt Ghast, a basic land under Mirari's Wake. When the extra
// mana depends on the pick (Mirari's Wake on a dual), or the picks are
// above the planner's cap, the answer is "", the weaker direction the
// planner itself takes above the cap. A unit-mask reader could not carry
// "the same colour as the land's" anyway (#2558's problem).

// TapTriggerMana is the view's reader of what a tap adds through
// triggered mana abilities. One per view build: it holds the board-wide
// answers the planner asks once per gather.
type TapTriggerMana struct {
	g *Game
	// priceProduction is producesManaReplacementsExistLocked: the
	// trigger's mana is priced through the CR 106.12b window as not a
	// tap, as the planner prices it.
	priceProduction bool
	identity        map[uuid.UUID]commanderIdentity
}

// TapTriggerManaLocked returns the reader, and false when no triggered
// mana ability exists on this board (manaTriggersExistLocked), which is
// nearly every board: then nothing needs asking.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) TapTriggerManaLocked() (*TapTriggerMana, bool) {
	if !g.manaTriggersExistLocked() {
		return nil, false
	}
	return &TapTriggerMana{
		g:               g,
		priceProduction: g.producesManaReplacementsExistLocked(),
		identity:        map[uuid.UUID]commanderIdentity{},
	}, true
}

// AddsLocked is the mana `controller` tapping `c` for `ab` now would add
// through triggered mana abilities, in the produced-mana grammar
// ("{B}", "{G}{G}"), or "" for none. "" too when the ability does not
// tap (CR 106.12a: only a tap for mana fires one, as in the planner),
// when it adds no mana, and when the extra mana is not the same for
// every pick (see the file comment).
//
// It asks only what the trigger adds: whether the ability may be
// activated now is the row's other fields' question, so a tapped land
// still says what its next tap would add.
//
// Caller must hold g.mu (a read lock is enough).
func (t *TapTriggerMana) AddsLocked(controller uuid.UUID, c *Card, ab *ManaAbilityShape) string {
	if t == nil || t.g == nil || c == nil || ab == nil || !ab.TapCost {
		return ""
	}
	g := t.g
	identity, ok := t.identity[controller]
	if !ok {
		identity = commanderIdentityFor(g, g.playerByIDLocked(controller))
		t.identity[controller] = identity
	}
	slots, printed := manaPlannableSlotsShapeLocked(g, controller, *c, ab, identity, t.priceProduction, true)
	if len(slots) == 0 {
		return ""
	}
	// The permanent as the executor copies it: after the tap.
	tapped := *c
	tapped.Tapped = true
	var cands []tapSource
	for _, cand := range appendTapSource(nil, tapSource{CardID: c.InstanceID}, slots) {
		cands = g.withTriggerManaLocked(cands, cand, tapped, controller, t.priceProduction, printed)
	}
	if len(cands) == 0 {
		return ""
	}
	first := cands[0].Slots[cands[0].ownSlots():]
	for _, cand := range cands[1:] {
		if !sameSlotShapes(first, cand.Slots[cand.ownSlots():]) {
			return ""
		}
	}
	return formatProducedSlots(first)
}

// formatProducedSlots writes slots back in the ParseProducedMana
// grammar: one brace per slot, its options piped, an "N of one colour"
// amount after its letter and a different-colours count after a colon.
// ParseProducedMana reads the result back to the same slots.
func formatProducedSlots(slots []ProducedManaEntry) string {
	var b strings.Builder
	for _, s := range slots {
		if len(s.Options) == 0 {
			continue
		}
		b.WriteByte('{')
		for i, opt := range s.Options {
			if i > 0 {
				b.WriteByte('|')
			}
			b.WriteString(opt)
			if n, ok := s.Amounts[opt]; ok && n != 1 {
				b.WriteString(strconv.Itoa(n))
			}
		}
		if s.Distinct >= 2 {
			b.WriteByte(':')
			b.WriteString(strconv.Itoa(s.Distinct))
		}
		b.WriteByte('}')
	}
	return b.String()
}
