package game

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

// duration_copy.go — "<permanent> becomes a copy of <object> until end
// of turn" (#1593, ADR 0043's amendment of 2026-09-28).
//
// # What is different from an entry copy
//
// An entry copy (copy.go, copy_choice.go) is settled ONCE, as the
// permanent enters, and never changes again while it is on the
// battlefield. A duration copy (Mirage Mirror, Cytoshape, Mirrorweave,
// Shifting Woodland's delirium, Unstable Shapeshifter, Lazav) is a
// CR 707.2 copy effect created by a resolving spell or ability
// (CR 611.2) on a permanent that is ALREADY on the battlefield. It has
// a timestamp of its own (CR 613.7), it may be one of several layer-1
// effects on the same object, and it ENDS — at which point the
// permanent is whatever it was underneath: its entry copy if it has
// one, otherwise itself.
//
// # The record
//
// The effect is a ScopedEffect (ADR 0041's data record) carrying one
// ModBecomeCopy mod, and that mod carries the copied PrintedValues BY
// VALUE. Never a reference to the copied card: Shifting Woodland and
// Lazav copy a card in a graveyard, which can be exiled, shuffled away
// or returned to hand while the copy lasts, and the copy does not
// care. CR 707.2's copiable values were read when the effect began —
// Cytoshape's ruling says so in as many words — so they are stored
// then. The record's Duration, its pin and its sweep are the ones
// every other scoped effect has.
//
// # Where it is applied
//
// Not in the layer pass. ADR 0043 Decision 1 is why: a copy has to
// change the oracle ID every catalog hook keys on, and the printed P/T
// the CR 704.5f check reads between recomputes, and neither lives on a
// Characteristic. So a duration copy is MATERIALISED onto the flat
// printed fields, exactly as an entry copy is, by
// materialiseDurationCopiesLocked, which runs
//
//   - at the top of every layer recompute, before the pass (so layers
//     2-7 run on its result, CR 613.1a);
//   - whenever the sweep drops a record (so the cleanup step's revert
//     is visible to the next reader of Card.OracleID, not only to the
//     next recompute);
//   - when a record is registered (so the resolving ability's own
//     continuation already sees the copy).
//
// # The two baselines
//
// `Card.PrintedSelf` is still the card's OWN printed values, what
// CR 400.7 puts back when the permanent leaves the battlefield.
// `Card.DurationCopyBase` is new: the layer-1 baseline BEFORE any
// duration copy — the entry copy's values for a Clone, the card's own
// for anything else. It is stashed by the first duration copy and put
// back when the last one ends. It is the one piece of state that
// cannot be re-derived, because once a Cytoshape has overwritten a
// Clone's printed fields nothing else on the board remembers what the
// Clone had copied.

// BecomeCopyMod is "becomes a copy of <values>" (#1593). `v` is the
// copied object's copiable values with the card's except clause
// already applied; it is cloned, so the caller may keep editing its
// own copy.
func BecomeCopyMod(v PrintedValues) Mod {
	return Mod{Kind: ModBecomeCopy, Copy: []PrintedValues{v.Clone()}}
}

// copyModProblem is the registration check for the copy kind: exactly
// one set of values on a becomeCopy mod, and none on any other. ""
// when the mod is sound.
func copyModProblem(m Mod) string {
	if m.Kind == ModBecomeCopy {
		if len(m.Copy) != 1 {
			return fmt.Sprintf("a becomeCopy mod carries exactly one set of copied values, got %d", len(m.Copy))
		}
		return ""
	}
	if len(m.Copy) != 0 {
		return fmt.Sprintf("mod %q carries copied values, which only becomeCopy reads", m.Kind)
	}
	return ""
}

// clonePrintedValuesSlice deep-copies a list of PrintedValues,
// preserving nil.
func clonePrintedValuesSlice(in []PrintedValues) []PrintedValues {
	if len(in) == 0 {
		return nil
	}
	out := make([]PrintedValues, len(in))
	for i := range in {
		out[i] = in[i].Clone()
	}
	return out
}

// CopiableValuesForEffect returns the copiable values (CR 707.2) of the
// object `cardID` names, wherever it is — a permanent on the
// battlefield, a card in a graveyard or in exile, a spell on the
// stack. The second result is false when no such object exists.
//
// Caller must hold g.mu.
func (g *Game) CopiableValuesForEffect(cardID uuid.UUID) (PrintedValues, bool) {
	c, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return PrintedValues{}, false
	}
	return CopiableValuesOf(c), true
}

// BecomeCopyForEffect makes each permanent in `targets` a copy of `v`
// for duration `d` (CR 707.2, CR 611.2) — the duration copy effect.
// `copied` names the object the values were taken from, for the event
// log only; the values themselves are what the copy uses, so it may be
// uuid.Nil or an object that has since left its zone.
//
// Targets not on the battlefield are skipped. A single target also
// pins the duration to that object (Duration.Pinned), which is the
// garbage collection an indefinite copy (Unstable Shapeshifter, Lazav)
// needs: the record is dropped when the permanent leaves, as it has
// nothing left to do (CR 400.7).
//
// An INDEFINITE copy of a single object also drops every older copy
// record that names exactly that object. It is invisible for as long
// as the new one lasts, which is as long as the object does, so
// keeping it would only grow the snapshot by one PrintedValues per
// creature an Unstable Shapeshifter ever saw enter.
//
// Reports false, having done nothing, when no target is on the
// battlefield. Caller must hold g.mu (write); effects call it from the
// resolution frame, which does.
func (g *Game) BecomeCopyForEffect(sourceID, copied uuid.UUID, targets []uuid.UUID, v PrintedValues, d Duration, label string) bool {
	affected := g.PinnedObjectsLocked(targets...)
	if len(affected) == 0 {
		return false
	}
	if len(affected) == 1 && d.Pinned == uuid.Nil {
		d = g.PinnedTo(d, affected[0].ID)
	}
	if d.Kind == Indefinite && len(affected) == 1 {
		g.dropSupersededCopiesLocked(affected[0])
	}
	if !g.appendScopedEffectLocked(sourceID, affected, ScopeNone, uuid.Nil, []Mod{BecomeCopyMod(v)}, d, label, timeNowUnixNano()) {
		return false
	}
	g.materialiseDurationCopiesLocked()
	for _, a := range affected {
		g.EmitEvent(Event{Kind: EventCopyApplied, CardID: a.ID, Source: copied})
	}
	return true
}

// dropSupersededCopiesLocked removes every copy-only record whose
// affected set is exactly `only`. Same fresh-slice rule as the sweep:
// the backing array is shared with every undo snapshot.
//
// Caller must hold g.mu (write).
func (g *Game) dropSupersededCopiesLocked(only AffectedObject) {
	kept := make([]ScopedEffect, 0, len(g.ScopedEffects))
	for _, e := range g.ScopedEffects {
		if isCopyOnlyRecord(&e) && len(e.Affected) == 1 && e.Affected[0] == only {
			continue
		}
		kept = append(kept, e)
	}
	if len(kept) == len(g.ScopedEffects) {
		return
	}
	if len(kept) == 0 {
		kept = nil
	}
	g.ScopedEffects = kept
}

// isCopyOnlyRecord reports whether every mod of e is a becomeCopy.
func isCopyOnlyRecord(e *ScopedEffect) bool {
	if len(e.Mods) == 0 {
		return false
	}
	for _, m := range e.Mods {
		if m.Kind != ModBecomeCopy {
			return false
		}
	}
	return true
}

// durationCopyFor returns the copied values the duration copies say
// `c` should carry right now, and false when none applies to it.
//
// CR 613.1a applies the layer-1 effects in timestamp order (CR 613.7),
// the entry copy first — it is already the baseline — and then each
// duration copy. Every copy effect sets ALL of the copiable values
// (CR 707.2), with its own except clause already folded into the
// values it carries, so applying them in order leaves exactly the
// values of the LAST one; that is what this returns. Ties keep
// registration order, which is how the layer pass sorts ties too.
func durationCopyFor(c *Card, g *Game, records []*ScopedEffect) (PrintedValues, bool) {
	var (
		best  *PrintedValues
		bestT int64
	)
	for _, e := range records {
		if !affectedPredicate(e.Affected)(c, g, nil) {
			continue
		}
		for i := range e.Mods {
			m := &e.Mods[i]
			if m.Kind != ModBecomeCopy || len(m.Copy) != 1 {
				continue
			}
			if best == nil || e.Timestamp >= bestT {
				best, bestT = &m.Copy[0], e.Timestamp
			}
		}
	}
	if best == nil {
		return PrintedValues{}, false
	}
	return *best, true
}

// materialiseDurationCopiesLocked brings every battlefield permanent's
// flat printed fields in line with the duration copies that apply to
// it — layer 1 of CR 613.1a for this effect class. Idempotent: a
// permanent already carrying the right values is not touched, so a
// recompute that changes nothing writes nothing.
//
// For a permanent some duration copy applies to, it stashes the
// pre-copy baseline in DurationCopyBase (and the card's own values in
// PrintedSelf, if no entry copy put them there already) and writes the
// winning copy's values. For a permanent none applies to any longer, it
// puts DurationCopyBase back — the entry copy's values for a Clone,
// the card's own for anything else — and drops PrintedSelf again when
// that baseline IS the card's own, so IsCopy is false once more.
//
// It bumps the layer version when it writes, because the printed
// baseline under the layer cache just changed. It emits no event: it
// runs inside the recompute, where a listener would read a half-built
// board (the #930 hazard). The registration emits EventCopyApplied.
//
// Caller must hold g.mu (write).
func (g *Game) materialiseDurationCopiesLocked() {
	if g.Battlefield == nil {
		return
	}
	var records []*ScopedEffect
	for i := range g.ScopedEffects {
		if scopedEffectHasMod(&g.ScopedEffects[i], ModBecomeCopy) {
			records = append(records, &g.ScopedEffects[i])
		}
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if len(records) == 0 && c.DurationCopyBase == nil {
			continue
		}
		want, ok := durationCopyFor(c, g, records)
		if !ok {
			if c.DurationCopyBase != nil {
				c.revertDurationCopy()
				g.layerVersion.Add(1)
			}
			continue
		}
		if c.DurationCopyBase == nil {
			base := printedValuesOf(*c)
			c.DurationCopyBase = &base
			if c.PrintedSelf == nil {
				own := base.Clone()
				c.PrintedSelf = &own
			}
		}
		if reflect.DeepEqual(printedValuesOf(*c), want) {
			continue
		}
		c.setPrintedValues(want)
		c.clearCarriedAbilitySlices()
		g.layerVersion.Add(1)
	}
}

// revertDurationCopy ends every duration copy on the permanent: the
// flat printed fields go back to DurationCopyBase. When that baseline
// is the card's own printed values there is no copy effect left on it
// at all, so PrintedSelf goes too.
func (c *Card) revertDurationCopy() {
	base := *c.DurationCopyBase
	c.setPrintedValues(base)
	c.clearCarriedAbilitySlices()
	if c.PrintedSelf != nil && reflect.DeepEqual(*c.PrintedSelf, base) {
		c.PrintedSelf = nil
	}
	c.DurationCopyBase = nil
}

// clearCarriedAbilitySlices drops the card-carried mana and activated
// ability slices when a duration copy changes the permanent's
// identity. Both are catalog-rebuildable (snapshot.go re-derives them
// from the key), and the reader falls back to the catalog under the
// NEW key when they are empty — which is the point: a Treasure token
// that becomes a copy of a creature no longer sacrifices for mana, and
// a creature that becomes a copy of a Treasure token does, through
// the token key the copied values carry (#521).
func (c *Card) clearCarriedAbilitySlices() {
	c.ManaAbilities = nil
	c.ActivatedAbilities = nil
}

// settleTimedEntryCopyLocked turns an "as this enters, it becomes a
// copy … until end of turn" entry copy (Cursed Mirror) into a duration
// copy, once the permanent has its entry stamp.
//
// The copy itself landed in applyEntersAsCopyLocked as an ordinary
// entry copy, BEFORE any event, so nothing ever saw the permanent as
// its own printed self and its ETB triggers are the copied card's —
// which is what "as this enters" means. What an entry copy lacks is an
// END. The record that gives it one has to name the object by its
// entry stamp (CR 400.7), and the stamp is written by the zone-move
// listener, after the copy. So this runs between that event and
// EventETB: it re-files the copy as a duration copy whose baseline is
// the card's own values, registered pinned to the stamped object.
// Nothing in between can revert it, because until the record exists
// the permanent has no DurationCopyBase and looks like any Clone.
//
// Caller must hold g.mu (write).
func (g *Game) settleTimedEntryCopyLocked(ev *ReplacementEvent, cardID uuid.UUID) {
	if ev == nil || ev.EntersAsCopyOf == nil || !ev.copyUntilEndOfTurn {
		return
	}
	c, ok := g.battlefieldCardLocked(cardID)
	if !ok || c.PrintedSelf == nil {
		return
	}
	values := printedValuesOf(*c)
	own := c.PrintedSelf.Clone()
	c.DurationCopyBase = &own
	label := fmt.Sprintf("%s — a copy until end of turn", own.Name)
	affected := g.PinnedObjectsLocked(cardID)
	d := g.PinnedTo(g.UntilEndOfTurnDuration(), cardID)
	if !g.appendScopedEffectLocked(cardID, affected, ScopeNone, uuid.Nil, []Mod{BecomeCopyMod(values)}, d, label, timeNowUnixNano()) {
		// Unreachable — the card is on the battlefield — but a copy
		// with no end must not be left looking like one that has.
		c.DurationCopyBase = nil
	}
}
