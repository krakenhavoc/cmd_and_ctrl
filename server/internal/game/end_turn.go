package game

import "github.com/google/uuid"

// end_turn.go — CR 724.1, "end the turn" (#2165, ADR 0059 amendment
// 2026-10-05). Sundial of the Infinite, Time Stop, Glorious End, Ultima,
// and Obeka, Brute Chronologist's "the player whose turn it is may end
// the turn".
//
// The rule, verbatim (September 25, 2026):
//
//	724.1   Some cards end the turn. When an effect ends the turn,
//	        follow these steps in order, as they differ from the normal
//	        process for resolving spells and abilities …
//	724.1a  If there are any triggered abilities that triggered before
//	        this process began but haven't been put onto the stack yet,
//	        those abilities cease to exist. They won't be put onto the
//	        stack. This rule does not apply to abilities that trigger
//	        during this process (see rule 724.1f).
//	724.1b  Exile every object on the stack, including the object
//	        that's resolving. All objects not on the battlefield or in
//	        the command zone that aren't represented by cards will cease
//	        to exist the next time state-based actions are checked …
//	724.1c  Check state-based actions. No player gets priority, and no
//	        triggered abilities are put onto the stack.
//	724.1d  The current phase and/or step ends. If this happens during
//	        combat, remove all creatures and planeswalkers from combat.
//	        The game skips straight to the cleanup step; skip any phases
//	        or steps between this phase or step and the cleanup step. If
//	        an effect ends the turn during the cleanup step, a new
//	        cleanup step begins.
//	724.1e  Even though the turn ends, "at the beginning of the end
//	        step" triggered abilities don't trigger because the end step
//	        is skipped.
//	724.1f  No player gets priority during this process, so triggered
//	        abilities are not put onto the stack. If any triggered
//	        abilities have triggered since this process began, those
//	        abilities are put onto the stack during the cleanup step,
//	        then the active player gets priority and players can cast
//	        spells and activate abilities. Then there will be another
//	        cleanup step before the turn finally ends. If no triggered
//	        abilities have triggered during this process, no player gets
//	        priority during the cleanup step. See rule 514, "Cleanup
//	        Step."
//
// # How each step maps
//
//   - 724.1a: dropUnstackedTriggersLocked empties PendingTriggers and
//     withdraws the prompts still putting one on the stack (its order,
//     its "you may", its mode, its target). It runs FIRST, so a trigger
//     the steps below cause lands on PendingTriggers afresh and
//     survives, which is the rule's last sentence.
//   - 724.1b: exileEntireStackLocked, end_combat.go's, unchanged. Not
//     countering: no EventCounterSpell, and "can't be countered" does
//     not stop it. A copy ceases to exist at once.
//   - 724.1c: stateChecksLocked(false) — the SBA loop with the trigger
//     drain switched off. What the actions trigger waits on
//     PendingTriggers.
//   - 724.1d: clearCombatLocked (which also ends a block declaration
//     still parked, #1501, and the announcements of the combat it
//     wipes), then skipToCleanupStepLocked: the turn plan
//     (turn_plan.go) is cut to the cleanup step alone and the cursor
//     pops it through advanceCursorLocked, so nothing between — the end
//     step included, and any phase an effect added — begins. From a
//     cleanup step, that pops a new one.
//   - 724.1e is the skip: the end step is never entered, so nothing
//     announces it and fireDelayedTriggersLocked never runs for it. A
//     delayed "next end step" trigger stays queued for the next end step
//     that does begin, which is next turn's (Glorious End). One bound to
//     an extra turn's end step (Final Fortune) is swept as that turn ends
//     without reaching it (ADR 0059 Decision 8).
//   - 724.1f is cleanup.go, unchanged: the cleanup step begins through
//     runStepEntryHooksLocked like any other step, its hand-size discard
//     and CR 514.2 sweep run, and its one exit grants the active player
//     priority when an SBA fired or a trigger is waiting — which puts
//     the 724.1c triggers on the stack there — and otherwise ends the
//     turn. Another cleanup step follows a window (CR 514.3a).
//
// # Two halves
//
// The process begins inside a resolution — the effect that says "end
// the turn" is resolving — and a turn must never end there (rotation.go,
// ADR 0059 Decision 6): beginning the next turn untaps, announces steps
// and queues upkeep triggers, and none of that may run under a spell.
// The CR 704.3 check is also held for the whole of a resolution
// (resolution_pause.go), so neither 724.1c nor the cleanup step's
// CR 514.3a check could read anything there.
//
// So EndTheTurnForEffect does 724.1a and 724.1b, parks priority (724.1f:
// nobody gets it during the process) and sets Game.TurnEndPending. The
// resolution's CR 704.3 boundary — runStateChecksLocked once nothing
// holds it — consumes the flag in finishEndingTheTurnLocked: 724.1c,
// 724.1d, and the cleanup step. The resolution frame does not hand
// priority back afterwards (passPriorityLocked's cursorMovedSince test),
// because the step the cursor moved to has decided who holds it.
//
// When the boundary waits — the resolution paused on its own prompt,
// or "may end the turn" was a prompt the active player answered later
// (Obeka) — the flag waits with it, in the snapshot, and the answer's
// boundary finishes the turn. SettleResolution backs that up after
// every action.

// EndTheTurnForEffect ends the turn (CR 724.1) from inside a resolving
// spell or ability, or from the answer to a prompt that resolution
// asked. `source` is the card whose effect ends it, for the log.
//
// It must be the LAST thing the resolving effect does: it exiles the
// object that is resolving, so any instruction after it would be run
// by a spell that is not on the stack. A no-op outside an active game.
//
// Caller must hold g.mu.
func (g *Game) EndTheTurnForEffect(source uuid.UUID) {
	if g.State != StateActive || len(g.Seats) == 0 {
		return
	}
	g.EmitEvent(Event{
		Kind:   EventTurnEnded,
		Actor:  g.activePlayerIDLocked(),
		Source: source,
		Amount: g.Turn.Seq,
	})
	// 724.1a.
	g.dropUnstackedTriggersLocked()
	// 724.1b.
	g.exileEntireStackLocked()
	// 724.1f: no player gets priority during this process.
	g.grantPriorityLocked(NoPriority)
	// 724.1c onwards: at the resolution's boundary.
	g.TurnEndPending = true
}

// dropUnstackedTriggersLocked is CR 724.1a (and 724.2a): every
// triggered ability that has triggered but is not on the stack yet
// ceases to exist. That is PendingTriggers, and the trigger-announcing
// prompts that hold one back from it — the CR 603.3b order, the
// CR 603.5 "you may", the CR 603.3c mode pick and the CR 603.3d target
// pick (triggerAnnouncementOpenLocked's set). Each is withdrawn through
// dropChoiceLocked, whose action for these kinds is nothing: the
// trigger they belong to is gone.
//
// Caller must hold g.mu.
func (g *Game) dropUnstackedTriggersLocked() {
	g.PendingTriggers = nil
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c == nil {
			continue
		}
		if c.Kind == PendingChoiceTriggerOrder ||
			c.triggerResume != nil || c.pickTargetResume != nil || c.modePickResume != nil {
			g.dropChoiceLocked(i)
		}
	}
}

// skipToCleanupStepLocked is the cursor half of CR 724.1d: the current
// step and phase end and the cursor moves straight to the cleanup step,
// through advanceCursorLocked like every step transition, WITHOUT
// running the entry hooks of anything in between. The plan is cut to
// the cleanup step alone, so the turn ends when it does: a phase an
// effect added after the ending phase is skipped with everything else.
//
// The cleanup step keeps the ending phase's id: the next one the plan
// lists, or the current phase's when the cursor is already in the
// ending phase with no cleanup left to come (a CR 514.3a priority
// window: "a new cleanup step begins").
//
// The cleanup step's own entry hooks do NOT run here; the caller runs
// them.
//
// Caller must hold g.mu.
func (g *Game) skipToCleanupStepLocked() {
	g.ensureTurnPlanLocked()
	phaseID := templatePhaseID(StepCleanup)
	if PhaseOf(g.Turn.Step) == PhaseEnding {
		phaseID = g.Turn.PhaseID
	}
	for _, ps := range g.TurnPlan {
		if ps.Step == StepCleanup {
			phaseID = ps.PhaseID
			break
		}
	}
	// A fresh slice, never an in-place edit: an undo clone may share
	// the old backing array (splicePlan's rule).
	g.TurnPlan = []PlannedStep{{Step: StepCleanup, PhaseID: phaseID}}
	g.advanceCursorLocked()
}

// finishEndingTheTurnLocked is the half of CR 724.1 that waits for the
// resolution's CR 704.3 boundary (see the file header). Consumes
// TurnEndPending. Called from runStateChecksLocked once nothing holds
// it, and nowhere else.
//
// Reports whether any state-based action was performed, as
// runStateChecksLocked does.
//
// Caller must hold g.mu.
func (g *Game) finishEndingTheTurnLocked() bool {
	g.TurnEndPending = false
	seq := g.Turn.Seq
	// 724.1c: state-based actions, with nobody receiving priority and no
	// triggered ability put on the stack. What they trigger waits on
	// PendingTriggers for the cleanup step (724.1f).
	fired := g.stateChecksLocked(false)
	if g.State != StateActive {
		return fired
	}
	if g.Turn.Seq != seq {
		// The check took the active player out of the game, and its turn
		// already ended through the rotation seam
		// (advancePastEliminatedLocked, ADR 0059 Decision 6). The new
		// turn's boundary is what is owed now.
		return g.stateChecksLocked(true) || fired
	}
	// 724.1d: combat ends, and the cursor skips to the cleanup step,
	// which pops with nobody holding priority (CR 514.3).
	g.clearCombatLocked()
	g.skipToCleanupStepLocked()
	// 724.1f: the cleanup step begins, and cleanup.go decides whether
	// anyone gets priority in it.
	g.runStepEntryHooksLocked()
	// The boundary every step entry gets from its caller (AdvanceStep's
	// closing check): the cleanup step either ended the turn, and the
	// next one's upkeep has triggers to drain, or it is waiting on a
	// discard or a priority window that has already drained its own.
	return g.stateChecksLocked(true) || fired
}

// cursorMovedSince reports whether the turn cursor has left the step
// `before` names — a different turn, a different step, or the same
// step begun again (a cleanup step after another) — or is about to (an
// ended turn whose cleanup step has not begun).
//
// The resolution frame asks it before handing priority back to the
// active player (CR 117.3b): an effect that ended the combat phase or
// the turn has moved the cursor itself, and the step it moved to has
// already decided who holds priority — nobody, in a cleanup step
// waiting on a discard.
//
// Caller must hold g.mu.
func (g *Game) cursorMovedSince(before Turn) bool {
	return g.TurnEndPending ||
		g.Turn.Seq != before.Seq ||
		g.Turn.Step != before.Step ||
		g.Turn.StepOrdinal != before.StepOrdinal
}
