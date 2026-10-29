package game

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// trigger_frame_keyed.go — #4361 S14a: a triggered ability paused while
// it is being put on the stack, written as data, so a table waiting on
// that prompt is a restore point.
//
// Three prompts pause a trigger's announcement (CR 603.3), and each
// holds the trigger in a resume frame:
//
//   - the CR 603.5 "you may" asked before the trigger goes on the stack
//     (PendingChoiceTriggerPrompt, triggerResumeFrame);
//   - the CR 603.3c mode choice (PendingChoiceModePick, modePickFrame);
//   - the CR 603.3d target walk (PendingChoicePickTarget,
//     pickTargetFrame), written by S14b as a targetWalkSnapshot: see
//     "The target walk" below.
//
// Every one of them holds the same five things, triggerFrameParts: the
// triggering event, a value copy of the source, its last-known
// characteristics, the doubler that made this instance, and the
// TriggeredAbility itself. The first four are data. The ability is a
// catalog declaration full of funcs (Build, Effect, Targets, Modes,
// HasLegalTarget, OptionalPrompt.Chooser, ...), so it is written as the
// catalog row it came from (ADR 0041 P9's AbilityRef, the same ref a
// harvested stack item carries) and rebuilt from the running catalog on
// restore. The answer reads the row again (OptionalPrompt, Modes,
// Targets, Build or Effect) exactly as it would have on the live table.
//
// On disk the frame is pendingChoices[].triggerFrame, and PickThen says
// TriggerFrameKey, so a binary from before this slice — which reads a
// PickThen it does not know as a revealed-hand pick's — refuses the
// file instead of restoring a prompt whose "yes" or mode does nothing.
//
// # Which rows can be named
//
// The frame is named by catalogTriggerRowRef, not TriggeredAbilityRef.
// TriggeredAbilityRef also refuses a row with no declared Effect and a
// row whose TargetsFrom reads the board, and both refusals are about the
// ITEM: a hand-written Build runs at trigger time and its item may
// capture what it saw, and a board-reading clause cannot be re-derived
// once the board has moved. A frame has built nothing yet. Answering it
// calls the row's Build (or installs its Effect) and runs its TargetsFrom
// for the first time, on the live table and on a restored one alike, so
// the row alone is the frame's ability. That matters: the one
// before-the-stack "you may" in production, miracle's reveal
// (MiracleTrigger), is a Build-only row whose Build makes a keyed item.
//
// # A trigger that cannot be named (owner decision D5, 2026-10-27)
//
// catalogTriggerRowRef returns nil for a trigger with no catalog row:
// one the engine derives (prowess, cipher, a face-down ward), a test
// stub, or a row the running catalog does not hand back under its ref.
// Its frame is not written; it stays the closure it was, and it is
// counted by the nearest existing runtime counter,
// ContinuationCensus.ChoiceResumeFrames ("pending choice <kind> holds
// triggerResume" / "modePickResume"), which is the counter that already
// counted every such pause. Not IntrinsicAbilityCards: that counter is
// about an ability carried on a card instance, and this is a pause in
// the announcement of an ability that is nobody's instance list. Owner
// decision D8 keeps the ChoiceResumeFrames runtime counter for exactly
// this kind of blocker.
//
// # A row the running catalog no longer has (Q3)
//
// As for a stamped stack item (ability_ref.go): the game is restored
// anyway. The prompt comes back without its frame, so answering it does
// nothing (the trigger is lost, never the table), the source is flagged
// AbilitiesLostOnRestore, and GameSnapshot.LostStackAbilities lists it
// so the boot path logs it.
//
// # API for the other trigger frames (S14b, S15)
//
//	parts := frame.parts()                        // a live frame's five things
//	snap := snapshotTriggerFrame(parts, cen)      // nil: not nameable (D5)
//	parts, ok := snap.rebuild()                   // !ok: the row is lost (Q3)
//
// A frame that holds more than the five writes the rest beside them
// and rebuilds its frame from parts plus that, as the target walk does.
//
// # The target walk (#4361 S14b)
//
// A pick_target prompt's frame is the five things plus the walk: the
// clause statement (spec, or the modal ability's modeSpec), the steps
// the statement announced, the CR 603.3c modes, the cursor, the targets
// picked so far, the division announced so far (CR 601.2d via CR
// 603.3d), the opponents named for "of an opponent's choice" steps
// (#2055) and each divided step's amount as it was fixed (#1657). On
// disk it is pendingChoices[].targetWalk, and PickThen says
// TargetWalkKey.
//
// The clauses are catalog declarations (predicates), so they are not
// written: restore asks the rebuilt row for its statement again, exactly
// as the walk's start did (triggerTargetsLocked with the frame's own
// event and source copy), and lists its steps with the written modes.
// What the walk added on top of the row is data and is carried: the
// picks, the division, the named opponents, and the divided amounts.
// An amount rule reads the board ("X is the number of lands you
// control"), and CR 601.2d fixes it as the ability is put on the stack,
// which is the start of this walk, so the restored walk uses the amount
// the live one fixed, never the restored board's.
//
// A row whose TargetsFrom reads the board (TargetsFromReadsBoard) is
// not written, and its frame stays a closure counted as D5 counts one.
// Its statement was fixed when the walk began (CR 601.2c with CR
// 603.3d: the number and kind of targets are part of the announcement),
// and the board can move while the prompt is open, so calling
// TargetsFrom again on the restored board could ask a different
// question from the one the live table was asking. Carrying the result
// instead is not possible: it is a clause of predicates. No catalog row
// sets the flag today.
//
// The "which opponent chooses" option pick of a walk's "of an
// opponent's choice" step holds the same frame as its pickTargetResume,
// and its keyed answer (triggerOpponentChooserThen) reads it back. It is
// written the same way, beside its own PickThen.

// TriggerFrameKey is what a written trigger frame's PickThen says. An
// on-disk identity: never renamed, never reused.
const TriggerFrameKey = "trigger/frame"

// triggerFrameParts is what every frame of a trigger's announcement
// holds. A plain value, never stored on the Game: the frames keep their
// own fields, and this is how they are written and rebuilt.
type triggerFrameParts struct {
	tc        TriggerContext
	source    Card
	lki       Characteristic
	ability   TriggeredAbility
	doubledBy doublerRef
}

// parts is the "you may" frame's five things.
func (f *triggerResumeFrame) parts() triggerFrameParts {
	return triggerFrameParts{tc: f.tc, source: f.source, lki: f.lki, ability: f.ability, doubledBy: f.doubledBy}
}

// parts is the mode pick frame's five things.
func (f *modePickFrame) parts() triggerFrameParts {
	return triggerFrameParts{tc: f.tc, source: f.source, lki: f.lki, ability: f.ability, doubledBy: f.doubledBy}
}

// parts is the target walk's five things.
func (f *pickTargetFrame) parts() triggerFrameParts {
	return triggerFrameParts{tc: f.tc, source: f.source, lki: f.lki, ability: f.ability, doubledBy: f.doubledBy}
}

// triggerResume is a "you may" frame holding p.
func (p triggerFrameParts) triggerResume() *triggerResumeFrame {
	return &triggerResumeFrame{tc: p.tc, source: p.source, lki: p.lki, ability: p.ability, doubledBy: p.doubledBy}
}

// modePick is a mode pick frame holding p.
func (p triggerFrameParts) modePick() *modePickFrame {
	return &modePickFrame{tc: p.tc, source: p.source, lki: p.lki, ability: p.ability, doubledBy: p.doubledBy}
}

// triggerFrameSnapshot is a trigger frame's five things on disk.
type triggerFrameSnapshot struct {
	// Trigger is the triggering event (#1223), carried whole as a
	// harvested stack item's is.
	Trigger TriggerContext `json:"trigger"`
	// Source is the dispatch's value copy of the source. A card value,
	// so a card whose instance abilities restore cannot re-derive is
	// counted as snapshotCard counts one (owner decision D1:
	// census:IntrinsicAbilityCards).
	Source cardSnapshot `json:"source"`
	// LKI is the source's last-known characteristics.
	LKI Characteristic `json:"lki"`
	// Ability names the catalog row (ADR 0041 P9).
	Ability AbilityRef `json:"ability"`
	// DoubledBy and DoubledByName are the CR 603.2d doubler that made
	// this instance, if any.
	DoubledBy     uuid.UUID `json:"doubledBy,omitempty"`
	DoubledByName string    `json:"doubledByName,omitempty"`
}

// snapshotTriggerFrame writes p, or returns nil when its ability cannot
// be named (D5): the caller then counts the frame as a closure. Caller
// holds g.mu.
func snapshotTriggerFrame(p triggerFrameParts, cen *ContinuationCensus) *triggerFrameSnapshot {
	ref := catalogTriggerRowRef(p.ability)
	if ref == nil {
		return nil
	}
	return &triggerFrameSnapshot{
		Trigger:       *cloneTriggerContext(&p.tc),
		Source:        snapshotCard(p.source, cen),
		LKI:           *copyCharacteristic(&p.lki),
		Ability:       *ref,
		DoubledBy:     p.doubledBy.id,
		DoubledByName: p.doubledBy.name,
	}
}

// rebuild is the restore half: the ability from the running catalog,
// the rest from the file. ok is false when no row answers the ref (Q3).
func (s *triggerFrameSnapshot) rebuild() (triggerFrameParts, bool) {
	ability, ok := triggeredAbilityFromRef(s.Ability)
	if !ok {
		return triggerFrameParts{}, false
	}
	source := restoreCard(&s.Source)
	// Derived and never written, as restoreZone builds it.
	source.stampPrinted()
	return triggerFrameParts{
		tc:        *cloneTriggerContext(&s.Trigger),
		source:    source,
		lki:       *copyCharacteristic(&s.LKI),
		ability:   ability,
		doubledBy: doublerRef{id: s.DoubledBy, name: s.DoubledByName},
	}, true
}

// triggeredAbilityFromRef is the catalog row a ref names, carrying the
// catalog identity the harvest's copy carried (the key and index the
// registry stamped, and a granted bundle's instance number), so the row
// is the same row to sameRow and to TriggeredAbilityRef as before the
// restart.
func triggeredAbilityFromRef(ref AbilityRef) (TriggeredAbility, bool) {
	row, at, outcome := resolveTriggeredAbilityRef(ref)
	if outcome == abilityRefLost {
		return TriggeredAbility{}, false
	}
	if strings.HasPrefix(at.Ref, abilityRefGrant) {
		if _, _, n, ok := parseGrantAbilityRef(at.Ref); ok {
			row.row.occurrence = n
		}
	}
	return row, true
}

// carriesTriggerFrame reports whether a prompt kind's frame is a
// trigger frame this binary writes.
func carriesTriggerFrame(kind PendingChoiceKind) bool {
	return kind == PendingChoiceTriggerPrompt || kind == PendingChoiceModePick
}

// triggerFrameOf is the live trigger frame a prompt holds, as parts.
func triggerFrameOf(c *PendingChoice) (triggerFrameParts, bool) {
	switch {
	case c == nil:
	case c.Kind == PendingChoiceTriggerPrompt && c.triggerResume != nil:
		return c.triggerResume.parts(), true
	case c.Kind == PendingChoiceModePick && c.modePickResume != nil:
		return c.modePickResume.parts(), true
	}
	return triggerFrameParts{}, false
}

// snapshotTriggerFrameOf writes a prompt's trigger frame, or nil when it
// holds none or its ability cannot be named (D5).
func snapshotTriggerFrameOf(c *PendingChoice, cen *ContinuationCensus) *triggerFrameSnapshot {
	p, ok := triggerFrameOf(c)
	if !ok {
		return nil
	}
	return snapshotTriggerFrame(p, cen)
}

// restoreTriggerFrameLocked puts a written trigger frame back on its
// prompt. A lost row (Q3) leaves the prompt with no frame and flags the
// source. Caller owns g.
func (g *Game) restoreTriggerFrameLocked(c *PendingChoice, s *triggerFrameSnapshot) {
	c.PickThen = ""
	p, ok := s.rebuild()
	if !ok {
		g.flagAbilitiesLostLocked(s.Source.InstanceID)
		return
	}
	switch c.Kind {
	case PendingChoiceTriggerPrompt:
		c.triggerResume = p.triggerResume()
	case PendingChoiceModePick:
		c.modePickResume = p.modePick()
	}
}

// unknownTriggerFrameKeys is checkEffectKeys' half for a trigger frame:
// the marker, and a ref this binary can read at all. A well-formed ref
// whose row is gone is Q3, restored without the frame, not refused.
func unknownTriggerFrameKeys(c *pendingChoiceSnapshot) []string {
	if c.TriggerFrame == nil || c.PickThen != TriggerFrameKey {
		return []string{fmt.Sprintf("%s continuation %q", c.Kind, c.PickThen)}
	}
	if ref := c.TriggerFrame.Ability; ref.Slot != AbilitySlotTriggered || !wellFormedAbilityRef(ref) {
		return []string{fmt.Sprintf("trigger frame ability ref %+v", ref)}
	}
	return nil
}

// lostTriggerFrames is LostStackAbilities' half for a trigger paused on
// a prompt whose row the running catalog no longer has.
func (s *GameSnapshot) lostTriggerFrames() []LostStackAbility {
	var out []LostStackAbility
	for _, c := range s.PendingChoices {
		var f *triggerFrameSnapshot
		switch {
		case c.TriggerFrame != nil && carriesTriggerFrame(c.Kind):
			f = c.TriggerFrame
		case c.TargetWalk != nil && carriesTargetWalk(c.Kind):
			// #4361 S14b: a target walk, or its "which opponent" question.
			f = &c.TargetWalk.Frame
		}
		if f == nil || !abilityRefLostInCatalog(f.Ability) {
			continue
		}
		out = append(out, LostStackAbility{
			ItemID: c.ID,
			CardID: f.Source.InstanceID,
			Card:   f.Source.Name,
			Label:  f.Ability.Name,
			Ref:    f.Ability,
		})
	}
	return out
}

// TargetWalkKey is what a written target walk's pick_target PickThen
// says (#4361 S14b). The walk's "which opponent chooses" option pick
// keeps its own key there (triggerOpponentChooserThen). An on-disk
// identity: never renamed, never reused.
const TargetWalkKey = "trigger/target-walk"

// targetWalkSnapshot is a trigger's CR 603.3d target walk on disk: its
// five things, and what the walk added to the row's statement.
type targetWalkSnapshot struct {
	// Frame is the walk's five things, the ability by its catalog row.
	Frame triggerFrameSnapshot `json:"frame"`
	// Modes are the CR 603.3c occurrences chosen, nil for a non-modal
	// ability. The steps are listed again from these and the row.
	Modes []int `json:"modes,omitempty"`
	// Step is the cursor: the step the open prompt asks about.
	Step int `json:"step"`
	// Picked are the targets answered so far, each stamped with its
	// step's Mode and Slot.
	Picked []TargetRef `json:"picked,omitempty"`
	// Dist is the division announced so far (#1563).
	Dist map[uuid.UUID]int `json:"dist,omitempty"`
	// Choosers are the opponents named for "of an opponent's choice"
	// steps so far (#2055).
	Choosers []TargetChooser `json:"choosers,omitempty"`
	// Divides are the divided steps' amounts as the walk's start fixed
	// them (#1657, CR 601.2d): one per step whose clause divides.
	Divides []boundDivideSnapshot `json:"divides,omitempty"`
}

// boundDivideSnapshot is one divided step's amount, as fixed.
type boundDivideSnapshot struct {
	Step        int  `json:"step"`
	Total       int  `json:"total"`
	FromX       bool `json:"fromX,omitempty"`
	DoubleFromX int  `json:"doubleFromX,omitempty"`
	UpTo        bool `json:"upTo,omitempty"`
}

// carriesTargetWalk reports whether a prompt kind can hold a trigger's
// target walk: its pick_target, or the option pick that asks which
// opponent chooses a step.
func carriesTargetWalk(kind PendingChoiceKind) bool {
	return kind == PendingChoicePickTarget || kind == PendingChoiceOptionPick
}

// targetWalkMarker is the PickThen a written walk on a prompt of this
// kind says.
func targetWalkMarker(kind PendingChoiceKind) string {
	switch kind {
	case PendingChoicePickTarget:
		return TargetWalkKey
	case PendingChoiceOptionPick:
		return triggerOpponentChooserThen.key
	}
	return ""
}

// snapshotTargetWalkOf writes a prompt's target walk, or nil when it
// holds none, its ability cannot be named (D5), or its statement reads
// the board (TargetsFromReadsBoard, see the file comment). Caller holds
// g.mu.
func snapshotTargetWalkOf(c *PendingChoice, cen *ContinuationCensus) *targetWalkSnapshot {
	if c == nil || !carriesTargetWalk(c.Kind) || c.pickTargetResume == nil {
		return nil
	}
	f := c.pickTargetResume
	if f.ability.TargetsFromReadsBoard {
		return nil
	}
	frame := snapshotTriggerFrame(f.parts(), cen)
	if frame == nil {
		return nil
	}
	out := &targetWalkSnapshot{
		Frame:    *frame,
		Modes:    copyInts(f.modes),
		Step:     f.step,
		Picked:   append([]TargetRef(nil), f.picked...),
		Dist:     cloneDistributionLocked(f.dist),
		Choosers: copyTargetChoosers(f.choosers),
	}
	for i := range f.steps {
		if d := f.steps[i].Clause.Divide; d != nil {
			out.Divides = append(out.Divides, boundDivideSnapshot{
				Step: i, Total: d.Total, FromX: d.FromX, DoubleFromX: d.DoubleFromX, UpTo: d.UpTo,
			})
		}
	}
	return out
}

// rebuildTargetWalkLocked is the restore half: the row from the running
// catalog, its statement and steps listed again exactly as the walk's
// start listed them, and the walk's own answers and fixed amounts from
// the file. ok is false when the row is gone (Q3) or no longer makes
// the statement the walk was answering (another step count, or a step
// that divides where the file says none did). Caller owns g.
func (g *Game) rebuildTargetWalkLocked(s *targetWalkSnapshot) (*pickTargetFrame, bool) {
	p, ok := s.Frame.rebuild()
	if !ok || p.ability.TargetsFromReadsBoard {
		return nil, false
	}
	spec := g.triggerTargetsLocked(p.ability, p.tc, &p.source)
	steps := AnnouncedClauses(spec, p.ability.Modes, s.Modes)
	if s.Step < 0 || s.Step >= len(steps) {
		return nil, false
	}
	bound := make(map[int]boundDivideSnapshot, len(s.Divides))
	for _, d := range s.Divides {
		bound[d.Step] = d
	}
	dividing := 0
	for i := range steps {
		if steps[i].Clause.Divide == nil {
			continue
		}
		b, carried := bound[i]
		if !carried {
			return nil, false
		}
		dividing++
		// The steps hold clause copies (AnnouncedClauses), so this never
		// touches the catalog's clause, as bindDivideAmountsLocked.
		steps[i].Clause.Divide = &DivideSpec{Total: b.Total, FromX: b.FromX, DoubleFromX: b.DoubleFromX, UpTo: b.UpTo}
	}
	if dividing != len(bound) {
		return nil, false
	}
	return &pickTargetFrame{
		ability:   p.ability,
		tc:        p.tc,
		source:    p.source,
		lki:       p.lki,
		spec:      spec,
		modeSpec:  p.ability.Modes,
		modes:     copyInts(s.Modes),
		steps:     steps,
		step:      s.Step,
		picked:    append([]TargetRef(nil), s.Picked...),
		dist:      cloneDistributionLocked(s.Dist),
		choosers:  copyTargetChoosers(s.Choosers),
		doubledBy: p.doubledBy,
	}, true
}

// restoreTargetWalkLocked puts a written walk back on its prompt. A
// pick_target's marker is cleared; an option pick keeps its key, which
// is its answer. A walk that cannot be rebuilt (Q3) leaves the prompt
// with no walk, so answering it does nothing, and flags the source.
// Caller owns g.
func (g *Game) restoreTargetWalkLocked(c *PendingChoice, s *targetWalkSnapshot) {
	if c.Kind == PendingChoicePickTarget {
		c.PickThen = ""
	}
	f, ok := g.rebuildTargetWalkLocked(s)
	if !ok {
		g.flagAbilitiesLostLocked(s.Frame.Source.InstanceID)
		return
	}
	c.pickTargetResume = f
}

// unknownTargetWalkKeys is checkEffectKeys' half for a target walk: the
// prompt's marker, and a ref this binary can read at all. A well-formed
// ref whose row is gone is Q3, restored without the walk, not refused.
func unknownTargetWalkKeys(c *pendingChoiceSnapshot) []string {
	w := c.TargetWalk
	if w == nil {
		// #4361 S13: a parked announcement's pick_target carries its own
		// marker, checked by unknownParkedAnnouncementKeys; (S15) a copy
		// walk's is unknownCopyWalkKeys'.
		if c.Kind == PendingChoicePickTarget && c.PickThen != "" && c.Parked == nil && (c.PickThen != CopyWalkKey || c.CopyWalk == nil) {
			return []string{fmt.Sprintf("%s continuation %q", c.Kind, c.PickThen)}
		}
		return nil
	}
	if !carriesTargetWalk(c.Kind) || c.PickThen != targetWalkMarker(c.Kind) {
		return []string{fmt.Sprintf("target walk on a %s prompt with tail %q", c.Kind, c.PickThen)}
	}
	if ref := w.Frame.Ability; ref.Slot != AbilitySlotTriggered || !wellFormedAbilityRef(ref) {
		return []string{fmt.Sprintf("target walk ability ref %+v", ref)}
	}
	return nil
}
