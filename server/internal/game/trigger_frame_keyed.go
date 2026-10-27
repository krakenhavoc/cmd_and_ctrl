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
//     pickTargetFrame). Not written yet: S14b builds on this file.
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
// A frame that holds more than the five (the target walk's cursor,
// picks and bound amounts) writes the rest beside the triggerFrame and
// rebuilds its frame from parts plus that. Note for the target walk: its
// frame holds a clause TargetsFrom already built, so unlike these two it
// cannot simply re-run TargetsFrom on restore when the row reads the
// board (TriggeredAbility.TargetsFromReadsBoard).

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
		f := c.TriggerFrame
		if f == nil || !carriesTriggerFrame(c.Kind) || !abilityRefLostInCatalog(f.Ability) {
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
