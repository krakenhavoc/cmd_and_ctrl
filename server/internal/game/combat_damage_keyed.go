package game

import (
	"slices"

	"github.com/google/uuid"
)

// combat_damage_keyed.go — #4362: a game paused in a combat damage step
// is written to its restore point.
//
// A combat damage step pauses in two ways, and before this neither was a
// restore point:
//
//   - THE HELD STEP (#3665, combat_damage_hold.go). A step that asks a
//     CR 510.1c or 510.1d damage_assignment prompt deals none of its
//     damage until the last prompt is answered (CR 510.2). The damage it
//     decided by itself and each answer so far were held on
//     Game.heldCombatDamage, which the snapshot did not carry, so the
//     census counted it. Now the hold is written as data
//     (GameSnapshot.HeldCombatDamage): the step's tag, every assignment
//     still to be dealt (source, recipient, amount, and whether it goes
//     to what the attacker is attacking) and every answered prompt with
//     its frame, its division and its trample excess. Every open
//     damage_assignment prompt then says HeldCombatDamageKey in PickThen,
//     so a binary from before this refuses the file rather than reading
//     the prompt as one with no hold behind it, which would deal its
//     answer at once and drop the rest of the step.
//
//   - A CR 616 PROMPT OVER COMBAT DAMAGE. A combat damage event that two
//     replacement or prevention effects want to modify asks its affected
//     player for their order (CR 616.1) and waits, while the rest of the
//     step's events are dealt. Its frame was kept off the disk
//     (damageEventIsData): the event's IsCombatDamage and its tail's
//     combat flag and step tag were not in the mirror. They are now
//     (replacementEventSnapshot.IsCombatDamage, damageTailSnapshot.Combat
//     and CombatStep), and the frame is written under a key of its own,
//     ReplacementResumeCombatDamage. An older binary knows the plain
//     damage key and would resume the event as noncombat damage (no
//     "deals combat damage" trigger, no commander damage); under the new
//     key it refuses the file.
//
// WHAT IS NOT WRITTEN. Damage already dealt is on the board and in the
// event log, as it always is: nothing about it is replayed. The step's
// shared damage instance (Game.combatDamageInstance) stays transient
// (ADR 0108 owner decision 4): a held step is dealt after its restore
// under one new instance, as all of it is dealt at once; a paused CR 616
// event carries its own instance on the event. A combat damage event
// with redirected parts still keeps its frame off the disk.
//
// A prevention follow-up that asks during the step (#2040) needed
// nothing new: it is owed as data (PreventionFollowUps) and a combat
// event has no continuation to park, so it is a restore point when its
// own prompt is keyed.

// HeldCombatDamageKey marks a damage_assignment prompt written with its
// step's held damage (GameSnapshot.HeldCombatDamage). An on-disk
// identity: never renamed, never reused.
const HeldCombatDamageKey = "damage-assignment/held-step"

// ReplacementResumeCombatDamage is the frame key of a combat damage
// event (#4362): its own key, because a binary that keys only noncombat
// damage would drop the combat fields. An on-disk identity: never
// renamed, never reused.
const ReplacementResumeCombatDamage = "replacement-resume/combat-damage"

// damageEventIsCombat reports whether a damage event is combat damage by
// any of the three marks it can carry.
func damageEventIsCombat(ev *ReplacementEvent) bool {
	if ev.IsCombatDamage {
		return true
	}
	t := ev.damageTail
	return t != nil && (t.combat || t.combatStep != "")
}

// snapshotDamageIsCombat is damageEventIsCombat over the mirror.
func snapshotDamageIsCombat(ev *replacementEventSnapshot) bool {
	if ev.IsCombatDamage {
		return true
	}
	t := ev.DamageTail
	return t != nil && (t.Combat || t.CombatStep != "")
}

// --- the held step on disk ------------------------------------------------

// heldCombatDamageSnapshot is a heldCombatDamage on disk.
type heldCombatDamageSnapshot struct {
	// Step is the step's Event.CombatStep tag.
	Step string `json:"step,omitempty"`
	// Marks are the assignments the step decided without a prompt, in the
	// order it decided them: still to be dealt.
	Marks []heldCombatMarkSnapshot `json:"marks,omitempty"`
	// Answers are the answered prompts, in answer order: also still to be
	// dealt, once the last prompt is answered.
	Answers []heldDamageAssignmentSnapshot `json:"answers,omitempty"`
}

// heldCombatMarkSnapshot is a heldCombatMark on disk.
type heldCombatMarkSnapshot struct {
	Source         uuid.UUID `json:"source"`
	Target         uuid.UUID `json:"target"`
	Amount         int       `json:"amount"`
	ToAttackTarget bool      `json:"toAttackTarget,omitempty"`
}

// heldDamageAssignmentSnapshot is an answered damage_assignment prompt on
// disk: its frame as the prompt carried it, the division and the trample
// excess.
type heldDamageAssignmentSnapshot struct {
	Frame   *DamageAssignmentFrame  `json:"frame"`
	Ordered []DamageAssignmentEntry `json:"ordered,omitempty"`
	Trample int                     `json:"trample,omitempty"`
}

// copyDamageAssignmentFrame copies a frame with its own slices. Nil is
// nil.
func copyDamageAssignmentFrame(f *DamageAssignmentFrame) *DamageAssignmentFrame {
	if f == nil {
		return nil
	}
	out := *f
	out.BlockerIDs = copyUUIDs(f.BlockerIDs)
	out.SourceLKI = copyCharacteristic(f.SourceLKI)
	return &out
}

// snapshotHeldCombatDamage writes the held step, nil for none.
func snapshotHeldCombatDamage(h *heldCombatDamage) *heldCombatDamageSnapshot {
	if h == nil {
		return nil
	}
	out := &heldCombatDamageSnapshot{Step: h.step}
	for _, m := range h.marks {
		out.Marks = append(out.Marks, heldCombatMarkSnapshot{Source: m.source, Target: m.target, Amount: m.amount, ToAttackTarget: m.toAttackTarget})
	}
	for _, a := range h.answers {
		out.Answers = append(out.Answers, heldDamageAssignmentSnapshot{
			Frame: copyDamageAssignmentFrame(a.frame), Ordered: slices.Clone(a.ordered), Trample: a.trample,
		})
	}
	return out
}

// restoreHeldCombatDamage is snapshotHeldCombatDamage's inverse.
func restoreHeldCombatDamage(s *heldCombatDamageSnapshot) *heldCombatDamage {
	if s == nil {
		return nil
	}
	out := &heldCombatDamage{step: s.Step}
	for _, m := range s.Marks {
		out.marks = append(out.marks, heldCombatMark{source: m.Source, target: m.Target, amount: m.Amount, toAttackTarget: m.ToAttackTarget})
	}
	for _, a := range s.Answers {
		out.answers = append(out.answers, heldDamageAssignment{
			frame: copyDamageAssignmentFrame(a.Frame), ordered: slices.Clone(a.Ordered), trample: a.Trample,
		})
	}
	return out
}

// unknownHeldCombatDamageKeys is checkEffectKeys' line for a
// damage_assignment prompt: one that names the held step must have it
// written beside it, one that names anything else is not this binary's,
// and while a step is held every one of its prompts names it.
func unknownHeldCombatDamageKeys(c *pendingChoiceSnapshot, held *heldCombatDamageSnapshot) []string {
	switch {
	case c.PickThen != "" && c.PickThen != HeldCombatDamageKey:
		return []string{"damage-assignment continuation " + c.PickThen}
	case c.PickThen == HeldCombatDamageKey && held == nil:
		return []string{"damage-assignment " + HeldCombatDamageKey + " with no held step"}
	case c.PickThen == "" && held != nil:
		return []string{"damage-assignment prompt outside its held step"}
	}
	return nil
}

// unknownHeldAnswerFrames is checkEffectKeys' line for the held step
// itself: an answer with no frame cannot be dealt.
func unknownHeldAnswerFrames(held *heldCombatDamageSnapshot) []string {
	if held == nil {
		return nil
	}
	for _, a := range held.Answers {
		if a.Frame == nil {
			return []string{"held combat damage answer with no frame"}
		}
	}
	return nil
}
