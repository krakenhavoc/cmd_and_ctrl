package game

import (
	"fmt"

	"github.com/google/uuid"
)

// copy_walk_keyed.go — #4361 S15: the CR 707.10c "you may choose new
// targets for the copy" walk, written as data, so a table waiting on
// that pick_target is a restore point.
//
// The walk's frame (copyFrame, spell_copy.go) holds a value copy of the
// copied stack item, a value copy of the object its targets are judged
// from (the copied spell, or an ability's source permanent), the copy's
// controller, the copied announcement's steps, the original targets, the
// copy's targets so far and the cursor. Everything but the steps is data
// or a mirror the engine already writes:
//
//   - the item goes through the stack-item mirror (snapshotStackItemAs),
//     the way a paused mode walk's item does (modeWalk.item). Its target
//     and mode clauses come back from where the item says it came from:
//     a spell by oracle ID, an ability by its catalog row (ADR 0041 P9),
//     a reflexive trigger by its body. An item the mirror cannot write
//     whole (a closure Effect with no body, a clause it cannot re-derive)
//     keeps the frame off the disk, and the prompt is counted by the
//     runtime ChoiceResumeFrames counter as before;
//   - the source object goes through the card mirror (snapshotCard),
//     which counts a card whose instance abilities restore cannot
//     re-derive (owner decision D1: census:IntrinsicAbilityCards);
//   - the targets and the cursor are plain data.
//
// The steps are clauses, catalog predicates, so they are not written.
// Restore lists them again from the rebuilt item exactly as the offer
// did (copyStepsFor, with the same clause list the offer's caller
// passed), which is S14b's approach for a trigger's target walk
// (trigger_frame_keyed.go): ask the source of the statement again rather
// than carry it. Each step's identity (its mode occurrence, option and
// slot) IS written, and a rebuilt list that does not match it is a
// statement the walk was not answering. That walk, like a stamped item
// whose catalog row is gone (Q3), comes back without its frame: the
// prompt stays, answering it does nothing, the source is flagged
// AbilitiesLostOnRestore, and LostStackAbilities lists a lost row.
//
// X and the counters removed are carried on the item and bound again as
// the offer bound them (bindStepsBound). A divided step's amount is the
// item's own: a copy changes which objects are targeted, never the
// division's total (CR 115.7f), and the item's Distribution is written.
//
// On disk the walk is pendingChoices[].copyWalk, and PickThen says
// CopyWalkKey, so a binary from before this slice refuses the file
// (it reads an unknown pick_target marker as a newer shape) instead of
// restoring a prompt whose answer makes no copy.

// CopyWalkKey is what a written copy walk's pick_target PickThen says
// (#4361 S15). An on-disk identity: never renamed, never reused.
const CopyWalkKey = "copy/target-walk"

// copyWalkSnapshot is a copy's CR 707.10c target walk on disk.
type copyWalkSnapshot struct {
	// Src is the object the copy's targets are judged from: the copied
	// spell, or the ability's source permanent (copyFrame.src).
	Src cardSnapshot `json:"src"`
	// Item is the copied stack item, through the stack-item mirror.
	Item stackItemSnapshot `json:"item"`
	// Controller is the copy's controller (CR 707.10b).
	Controller uuid.UUID `json:"controller"`
	// Steps are the announced steps' identities, to check the rebuilt
	// list against; the clauses are listed again on restore.
	Steps []copyStepSnapshot `json:"steps"`
	// Orig is the original's targets stamped with their steps, Next the
	// copy's targets so far, and Step the cursor.
	Orig []TargetRef `json:"orig,omitempty"`
	Next []TargetRef `json:"next,omitempty"`
	Step int         `json:"step"`
}

// copyStepSnapshot is one announced step's identity.
type copyStepSnapshot struct {
	Mode   int `json:"mode"`
	Option int `json:"option"`
	Slot   int `json:"slot"`
}

// copyWalkOracleID is the oracle ID the walk's item is written under:
// the copied spell's own, which the frame's source carries. A copy of a
// copy is in no zone, so the zone lookup the stack's mirror makes would
// find nothing. An ability's item is read through its catalog row, not
// an oracle ID, and gets its source's.
func copyWalkOracleID(f *copyFrame) string {
	return f.src.OracleID
}

// snapshotCopyWalkOf writes a prompt's copy walk, or nil when it holds
// none or its item cannot be written whole (the frame is then counted
// as a closure). Caller holds g.mu.
func snapshotCopyWalkOf(c *PendingChoice, cen *ContinuationCensus) *copyWalkSnapshot {
	if c == nil || c.Kind != PendingChoicePickTarget || c.copyResume == nil {
		return nil
	}
	f := c.copyResume
	var scratch ContinuationCensus
	item := snapshotStackItemAs(&f.item, copyWalkOracleID(f), &scratch)
	if !scratch.Empty() {
		return nil
	}
	out := &copyWalkSnapshot{
		Src:        snapshotCard(f.src, cen),
		Item:       item,
		Controller: f.controller,
		Orig:       copyTargetRefs(f.orig),
		Next:       copyTargetRefs(f.next),
		Step:       f.step,
	}
	for _, st := range f.steps {
		out.Steps = append(out.Steps, copyStepSnapshot{Mode: st.Mode, Option: st.Option, Slot: st.Slot})
	}
	return out
}

// rebuildCopyWalkLocked is the restore half: the item from the mirror,
// its clauses from where it says it came from, the steps listed again
// as the offer listed them, and the walk's own answers from the file.
// ok is false when the item's row is gone (Q3) or the rebuilt steps are
// not the ones the walk was answering. Caller owns g; the zones are
// restored.
func (g *Game) rebuildCopyWalkLocked(s *copyWalkSnapshot) (*copyFrame, bool) {
	it, ok := restoreStackItem(&s.Item)
	if !ok {
		return nil, false
	}
	// A triggered item whose row builds its clause from the trigger
	// context gets it back the way a restored stack item does.
	g.rederiveTriggerClauseLocked(it, &s.Item)
	src := restoreCard(&s.Src)
	// Derived and never written, as restoreZone builds it.
	src.stampPrinted()
	// The clause list the offer's caller passed: the spell copy's
	// (copySpellFromLocked) or the ability copy's (CopyAbilityForEffect).
	spec := it.targetSpec
	if it.Kind == StackItemSpell {
		spec = castTargetSpecForItem(CatalogKey(src), it)
	}
	steps := copyStepsFor(src, it, spec)
	if len(steps) != len(s.Steps) || s.Step < 0 || s.Step >= len(steps) {
		return nil, false
	}
	for i, st := range steps {
		if w := s.Steps[i]; st.Mode != w.Mode || st.Option != w.Option || st.Slot != w.Slot {
			return nil, false
		}
	}
	return &copyFrame{
		src:        src,
		item:       *it,
		controller: s.Controller,
		steps:      steps,
		orig:       copyTargetRefs(s.Orig),
		next:       copyTargetRefs(s.Next),
		step:       s.Step,
	}, true
}

// restoreCopyWalkLocked puts a written copy walk back on its prompt and
// clears the marker. A walk that cannot be rebuilt leaves the prompt
// with no frame, so answering it does nothing, and flags the source.
// Caller owns g.
func (g *Game) restoreCopyWalkLocked(c *PendingChoice, s *copyWalkSnapshot) {
	c.PickThen = ""
	f, ok := g.rebuildCopyWalkLocked(s)
	if !ok {
		g.flagAbilitiesLostLocked(s.Src.InstanceID)
		return
	}
	c.copyResume = f
}

// unknownCopyWalkKeys is checkEffectKeys' half for a copy walk: the
// prompt kind and its marker, and the keys the item names, held to the
// rules a stack item is (a body this binary registers, a well-formed
// ability ref, a valid spell filter). A well-formed ref whose row is
// gone is Q3, restored without the walk, not refused.
func unknownCopyWalkKeys(c *pendingChoiceSnapshot) []string {
	w := c.CopyWalk
	if w == nil {
		return nil
	}
	if c.Kind != PendingChoicePickTarget || c.PickThen != CopyWalkKey || c.TargetWalk != nil {
		return []string{fmt.Sprintf("copy walk on a %s prompt with tail %q", c.Kind, c.PickThen)}
	}
	var out []string
	if w.Item.Body != "" && !KnownEffectBody(w.Item.Body) {
		out = append(out, "copy walk stack-item body "+w.Item.Body)
	}
	if w.Item.Params != nil && !w.Item.Params.Filter.Valid() {
		out = append(out, fmt.Sprintf("copy walk stack-item spell filter %v", w.Item.Params.Filter.Types))
	}
	return append(out, unknownAbilityRef(w.Item)...)
}

// lostCopyWalks is LostStackAbilities' half for a copy walk whose
// copied ability names a row the running catalog no longer has.
func (s *GameSnapshot) lostCopyWalks() []LostStackAbility {
	var out []LostStackAbility
	for _, c := range s.PendingChoices {
		w := c.CopyWalk
		if w == nil || c.Kind != PendingChoicePickTarget {
			continue
		}
		if _, ok := catalogBodySlot(w.Item.Body); !ok || w.Item.Params == nil || w.Item.Params.Ability == nil {
			continue
		}
		ref := *w.Item.Params.Ability
		if !abilityRefLostInCatalog(ref) {
			continue
		}
		out = append(out, LostStackAbility{
			ItemID: c.ID,
			CardID: w.Src.InstanceID,
			Card:   w.Src.Name,
			Label:  w.Item.Label,
			Ref:    ref,
		})
	}
	return out
}
