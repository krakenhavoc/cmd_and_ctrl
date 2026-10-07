package game

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// own_ability_removal.go is the engine half of "this permanent loses ONE
// of its own abilities until end of turn" (#1859, CR 613.1f, CR 611.2a;
// ADR 0106 amendment of 2026-10-07). Glittering Lion and Glittering Lynx
// print "{3}: Until end of turn, this creature loses 'Prevent all damage
// that would be dealt to this creature.' Any player may activate this
// ability."
//
// The vocabulary already had two removals: removeKeywords (a keyword
// token) and loseAllAbilities (every own ability, AbilitiesRemoved). A
// catalog ability that is not a keyword is a ROW of the card's
// definition, read through a Catalog* hook keyed on the whole object, so
// switching one off means naming the row. The loseOwnAbility mod does
// that: a slot ("replacement", "triggered" or "activated") and the row's
// index in the full declared list of that slot. It is an ordinary layer-6
// operation: it takes its record's timestamp, it ends with the record's
// duration, and it writes Characteristic.RemovedOwnAbilities, which is
// not copiable (CR 707.2) and is rebuilt from scratch by every pass.
//
// # Where it is honoured
//
// loseAllAbilities is honoured by emptying the OWN half of the ability
// key (ownAbilityKey). The per-row removal rides the same accessor:
//
//   - replacement and triggered rows: catalogAbilityKeyOf appends one
//     "-<slot>:<row>" part per removal to the composite key, and
//     mergedCatalogDef, the one place a composite key becomes a
//     definition, drops those rows from the BASE definition's list. So
//     every reader that already asks for the layered key (the
//     replacement gather, the trigger harvest, the view's ability rows,
//     the last-known-information harvest of a dies trigger) honours it
//     with no change of its own.
//   - activated rows: activatedAbilityRows filters its own half by the
//     removed indexes, keeping the survivors' original refs so a stack
//     item or a stale client index still resolves (ADR 0093 Decision 5).
//
// A layer-6 GRANT of the same text is a different row, in
// GrantedAbilities, and is untouched: a grant with a later timestamp puts
// the ability back (CR 613.6) and the removal never reaches it. A copy of
// the creature copies the printed definition (PrintedValues), never the
// layered result, and a permanent that flickers is a new object the
// record no longer follows (CR 400.7).
//
// # What it deliberately does not do
//
// Static and mana rows. A static is gathered by the layer pass BEFORE
// any effect applies, from the printed key (staticAbilitiesOf), so a
// removal computed inside the pass cannot reach it without the silencing
// machinery loseAllAbilities uses per bucket; a mana row's ref also
// numbers through the intrinsic land half. Neither has a card waiting.
// Registration refuses both slots, so a card cannot ship a removal the
// engine does not honour (the closed-table rule); the follow-up is named
// in docs/engine-seams.md.

// AbilitySlotReplacement is the replacement-effect list of a CardDef, a
// slot loseOwnAbility can name beside AbilitySlotActivated and
// AbilitySlotTriggered.
const AbilitySlotReplacement = "replacement"

// OwnAbilityRemoval names one row of the object's own definition: a slot
// and the row's index in the full declared list of that slot.
type OwnAbilityRemoval struct {
	Slot string `json:"slot"`
	Row  int    `json:"row"`
}

// KnownOwnAbilitySlot reports whether loseOwnAbility can switch off a
// row of the slot.
func KnownOwnAbilitySlot(slot string) bool {
	switch slot {
	case AbilitySlotReplacement, AbilitySlotTriggered, AbilitySlotActivated:
		return true
	}
	return false
}

// LoseOwnAbilityMod is "<affected> loses '<one of its own printed
// abilities>'" (layer 6, #1859): the row at index `row` of the slot's
// full declared list on the object's definition. Reads Slot and Row.
func LoseOwnAbilityMod(slot string, row int) Mod {
	return Mod{Kind: ModLoseOwnAbility, Slot: slot, Row: row}
}

// loseOwnAbilityModProblem is the registration and restore check: a
// known slot and a non-negative row on loseOwnAbility, and neither field
// on any other kind (a newer binary's shape). "" when the mod is sound.
func loseOwnAbilityModProblem(m Mod) string {
	if m.Kind != ModLoseOwnAbility {
		if m.Slot != "" || m.Row != 0 {
			return fmt.Sprintf("mod %q names an ability row, which only loseOwnAbility reads", m.Kind)
		}
		return ""
	}
	if !KnownOwnAbilitySlot(m.Slot) {
		return fmt.Sprintf("a loseOwnAbility mod names unknown ability slot %q", m.Slot)
	}
	if m.Row < 0 {
		return fmt.Sprintf("a loseOwnAbility mod names negative row %d", m.Row)
	}
	return ""
}

// loseOwnAbilityApply is the mod's layer-6 Apply.
func loseOwnAbilityApply(m Mod) func(*Characteristic, *Card, *Game, *Card) {
	r := OwnAbilityRemoval{Slot: m.Slot, Row: m.Row}
	return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
		if !slices.Contains(ch.RemovedOwnAbilities, r) {
			ch.RemovedOwnAbilities = append(ch.RemovedOwnAbilities, r)
		}
	}
}

// removedRowKeyPrefix starts a composite-key part that is a removal and
// not a grant. A grant part starts with "grant:" and a card's with an
// oracle ID, so "-" is unambiguous.
const removedRowKeyPrefix = "-"

// withRemovedRows appends the key parts for the removals the composite
// key honours (replacement and triggered rows) to an own key. An empty
// own key stays empty: an object with no definition has no row to lose.
// The parts are sorted and de-duplicated so the same set of removals is
// always the same key (the merged-definition memo is keyed on it).
func withRemovedRows(own string, removed []OwnAbilityRemoval) string {
	if own == "" || len(removed) == 0 {
		return own
	}
	var parts []string
	for _, r := range removed {
		if r.Slot == AbilitySlotActivated {
			continue
		}
		parts = append(parts, removedRowKeyPrefix+r.Slot+":"+strconv.Itoa(r.Row))
	}
	if len(parts) == 0 {
		return own
	}
	sort.Strings(parts)
	parts = slices.Compact(parts)
	return own + grantKeySeparator + strings.Join(parts, grantKeySeparator)
}

// splitRemovedRows separates the removal parts of a composite key's tail
// from its grant parts.
func splitRemovedRows(tail []string) (grants []string, removed []OwnAbilityRemoval) {
	for _, p := range tail {
		if !strings.HasPrefix(p, removedRowKeyPrefix) {
			grants = append(grants, p)
			continue
		}
		slot, rowText, ok := strings.Cut(strings.TrimPrefix(p, removedRowKeyPrefix), ":")
		row, err := strconv.Atoi(rowText)
		if !ok || err != nil {
			continue
		}
		removed = append(removed, OwnAbilityRemoval{Slot: slot, Row: row})
	}
	return grants, removed
}

// removedRows reports whether row i of `slot` is among the removals.
func removedRowSet(removed []OwnAbilityRemoval, slot string) map[int]bool {
	var set map[int]bool
	for _, r := range removed {
		if r.Slot != slot {
			continue
		}
		if set == nil {
			set = make(map[int]bool, len(removed))
		}
		set[r.Row] = true
	}
	return set
}

// dropRows is rows without the indexes in `drop` that fall inside the
// first `own` entries (the base definition's rows; later entries are a
// grant's and are never touched). Returns a new slice.
func dropRows[T any](rows []T, own int, drop map[int]bool) []T {
	if len(drop) == 0 {
		return rows
	}
	out := make([]T, 0, len(rows))
	for i := range rows {
		if i < own && drop[i] {
			continue
		}
		out = append(out, rows[i])
	}
	return out
}

// dropRemovedRows returns d with the removed rows of the BASE definition
// (`base`, parts[0]) taken out of its replacement and triggered lists. d
// is the merged definition: base rows first, then each grant's, so row i
// of the base is row i of d's list. The lists are rebuilt, never written
// in place, since a merge may share a slice with the shared catalog
// definition.
func dropRemovedRows(d, base *CardDef, removed []OwnAbilityRemoval) *CardDef {
	if d == nil || base == nil || len(removed) == 0 {
		return d
	}
	if set := removedRowSet(removed, AbilitySlotReplacement); set != nil {
		d.Replacements = dropRows(d.Replacements, len(base.Replacements), set)
	}
	if set := removedRowSet(removed, AbilitySlotTriggered); set != nil {
		d.Triggered = dropRows(d.Triggered, len(base.Triggered), set)
	}
	return d
}

// ownActivatedRemovals reports whether the object has an activated row
// switched off, so activatedAbilityRows knows to ask for indexes.
func ownActivatedRemovals(c *Card) bool {
	if c.effective == nil {
		return false
	}
	for _, r := range c.effective.RemovedOwnAbilities {
		if r.Slot == AbilitySlotActivated {
			return true
		}
	}
	return false
}

// dropRemovedActivated filters an own activated list by the removals.
// `idx` is each row's index in the full declared list (nil means 1:1);
// the survivors' indexes are returned so their refs stay those of the
// declared list.
func dropRemovedActivated(c *Card, rows []ActivatedAbilityShape, idx []int) ([]ActivatedAbilityShape, []int) {
	set := removedRowSet(c.effective.RemovedOwnAbilities, AbilitySlotActivated)
	if set == nil {
		return rows, idx
	}
	out := make([]ActivatedAbilityShape, 0, len(rows))
	outIdx := make([]int, 0, len(rows))
	for i := range rows {
		k := i
		if idx != nil {
			k = idx[i]
		}
		if set[k] {
			continue
		}
		out = append(out, rows[i])
		outIdx = append(outIdx, k)
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, outIdx
}
