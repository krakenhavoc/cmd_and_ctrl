package game

import (
	"slices"

	"github.com/google/uuid"
)

// prevention_then_pause.go — a prevention effect's additional effect
// that asks a question is answered AT ONCE (#2040, ADR 0108 amendment
// of 2026-10-16).
//
// THE RULES.
//
//   - CR 615.5: the prevention "takes place at the time the original
//     event would have happened; the rest of the effect takes place
//     immediately afterward."
//   - The Khârn the Betrayer ruling: "An opponent gaining control of it
//     is part of the replacement effect and happens immediately, even if
//     you are in the middle of resolving the spell or ability and it has
//     more instructions to perform after taking damage." And: a spell
//     that deals it damage several times changes its controller each
//     time.
//   - CR 704.3 / 510.2: state-based actions are checked only when a
//     player would receive priority, and combat damage is dealt all at
//     once. So a combat damage step's follow-ups, choices included, are
//     made after all of the step's damage and before the state-based
//     actions (decision 3 below).
//
// WHAT WAS WRONG. A follow-up (PreventionFollowUp) runs as its instance
// of damage ends. A body that has to ask — Immortal Coil's which cards,
// Khârn's which opponent — can only QUEUE a prompt, and the code after
// the damage instruction went on without the answer: the damage
// instruction's own continuation (a `...ThenForEffect` caller, the rest
// of an Each walk's caller), the next owed follow-up (whose chooser may
// be the player the answer hands Khârn to), and in combat the
// state-based actions, which ran before the question was answered.
//
// THE SHAPE. It is #1289's resolution pause, applied to the follow-up:
//
//  1. A follow-up body runs with the resolution OPEN (runFollowUpAsking-
//     Locked), so a blocking prompt it queues is stamped midResolution —
//     also in a combat damage step, which has no resolution of its own.
//     That alone holds the CR 704.3 boundary (state-based actions and
//     the trigger drain) until the answer, parks a modal spell's later
//     bullets (mode_walk.go) and its exit (resolution_tail.go).
//  2. While any such prompt is open, no other follow-up runs
//     (followUpsHeldLocked): the owed entries wait in the queue, in
//     order, and are run as the prompt is answered. So the second time a
//     spell damages Khârn, "an opponent of your choice" is asked of the
//     player who controls it after the first answer.
//  3. The instruction whose instance asked parks its continuation
//     (followUpParks): runDamageTailLocked's `then` and an Each walk's
//     total. When the last prompt is answered, the boundary's hold check
//     (holdForOpenResolutionLocked) first runs the instance's remaining
//     follow-ups, then the parked continuation — before a modal spell's
//     next bullet, because the continuation is the earlier instruction.
//  4. The catch-all flush that runs before the state-based actions
//     (stateChecksLocked) stops when a follow-up asks, and the boundary
//     returns: the answer's own boundary runs the rest and then the
//     state-based actions.
//
// SNAPSHOTS. The owed follow-ups, the open resolution and the prompt's
// stamp are all plain data the snapshot already carries, and a follow-up
// that asks with a KEYED prompt (an option pick or a card-set pick with
// a registered continuation) leaves a table that is still a restore
// point. A parked continuation is a closure, counted by the census like
// the CR 616 resume it sits beside, so a table holding one keeps the
// restore point from before the damage — exactly what a damage event
// paused on a CR 616 ordering prompt does.
//
// A catalog effect written as a plain sequence of steps
// (Do(DealDamage{…}, DrawCards{…})) waits too since #4012: the steps
// after the one that asked are parked in Game.pausedStepWalk and resumed
// after the damage instruction's own continuation and the follow-ups
// still owed, announcements included (step_walk.go). A Do
// run from inside a hand-written closure still runs its later steps on
// the Go stack, as it does after a CR 616 ordering prompt.

// followUpPark is the rest of one damage instruction, parked behind a
// prompt its instance's follow-up asked.
type followUpPark struct {
	inst DamageInstance
	run  func(g *Game) error

	// keyed is `run` as data (#4253, damage_keyed.go): a damage event's
	// registered tail, told target and dealt. Never set together with run.
	keyed  *KeyedTail
	target uuid.UUID
	dealt  int
}

// followUpsHeldLocked reports whether a follow-up must wait: the
// resolution is paused on one of its own prompts (#1289), a follow-up's
// or an earlier instruction's.
//
// Caller must hold g.mu.
func (g *Game) followUpsHeldLocked() bool {
	return g.resolutionOpen && g.pausedResolutionChoiceLocked() != nil
}

// runFollowUpAskingLocked runs one follow-up with the resolution open
// and reports whether it left the resolution paused on a prompt. When it
// did not, the resolution is put back as it was; when it did, it stays
// open, so the boundary is held until the answer (decision 1).
//
// Caller must hold g.mu (write).
func (g *Game) runFollowUpAskingLocked(f PreventionFollowUp) bool {
	wasOpen := g.resolutionOpen
	g.resolutionOpen = true
	g.runPreventionFollowUpLocked(f)
	if g.pausedResolutionChoiceLocked() != nil {
		return true
	}
	g.resolutionOpen = wasOpen
	return false
}

// maxFollowUpRuns bounds one flush. A follow-up's own damage can meet
// another prevention effect with a follow-up, and that one's damage
// another; ADR 0107 bounded the same regress at eight rounds.
const maxFollowUpRuns = 256

// runOwedFollowUpsLocked runs the owed follow-ups `due` admits, one at a
// time in the order they were first owed, and stops as soon as one asks
// a question. Reports whether the flush is paused: one of them asked, or
// one is held behind an open prompt (followUpsHeldLocked) and has not
// run. The entries not run stay owed.
//
// Caller must hold g.mu (write).
func (g *Game) runOwedFollowUpsLocked(due func(PreventionFollowUp) bool) bool {
	for n := 0; n < maxFollowUpRuns; n++ {
		i := slices.IndexFunc(g.preventionFollowUps, due)
		if i < 0 {
			return false
		}
		if g.followUpsHeldLocked() {
			return true
		}
		f := g.preventionFollowUps[i]
		// Copy on write: an undo clone shares the slice.
		g.preventionFollowUps = slices.Delete(slices.Clone(g.preventionFollowUps), i, i+1)
		if len(g.preventionFollowUps) == 0 {
			g.preventionFollowUps = nil
		}
		if g.runFollowUpAskingLocked(f) {
			return true
		}
	}
	return false
}

// parkFollowUpContinuationLocked parks the rest of a damage instruction
// behind its instance's follow-up prompt (decision 3).
//
// Caller must hold g.mu (write).
func (g *Game) parkFollowUpContinuationLocked(inst DamageInstance, run func(g *Game) error) {
	if run == nil {
		return
	}
	g.followUpParks = append(slices.Clone(g.followUpParks), followUpPark{inst: inst, run: run})
}

// resumeFollowUpParksLocked runs what the follow-up prompts held, now
// that none is open: each parked instruction's remaining follow-ups,
// then the instruction's continuation, oldest first. Reports whether it
// is paused again — a follow-up or a continuation asked something new.
//
// The entries run before a continuation are its own instance's and every
// EARLIER instance's still owed (instances are numbered in the order
// they begin): an announcement (#2047, CR 615.13) or a follow-up an
// earlier instruction owed behind the same question happened before
// this instruction's damage, so it is emitted or run before this
// instruction goes on. A later instance's entries wait for their own
// park or the catch-all flush.
//
// Called from holdForOpenResolutionLocked, before a modal item's next
// bullet. Caller must hold g.mu (write).
func (g *Game) resumeFollowUpParksLocked() bool {
	for len(g.followUpParks) > 0 {
		p := g.followUpParks[0]
		if g.runOwedFollowUpsLocked(func(f PreventionFollowUp) bool { return f.Instance != 0 && f.Instance <= p.inst }) {
			return true
		}
		g.followUpParks = slices.Clone(g.followUpParks[1:])
		if len(g.followUpParks) == 0 {
			g.followUpParks = nil
		}
		if err := g.runFollowUpParkLocked(p); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: "damage continuation failed: " + err.Error()})
		}
		if g.pausedResolutionChoiceLocked() != nil {
			return true
		}
	}
	return false
}
