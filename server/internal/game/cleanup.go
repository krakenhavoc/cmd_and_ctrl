package game

// cleanup.go — the cleanup step (CR 514), and the extra one CR 514.3a
// asks for.
//
// Cleanup is one of the two steps that grant nobody priority
// (CR 514.3): the active player discards down to their maximum hand
// size (CR 514.1), marked damage and "until end of turn" effects go
// away (CR 514.2), and the turn ends. CR 514.3a is the exception, and
// it is the one that matters — it is how a trigger watching the
// hand-size discard reaches the stack in the turn the discard
// happened in:
//
//	"However, if any state-based actions are performed as a result of
//	the state-based action check, or if any triggered abilities are
//	waiting to be put onto the stack, those state-based actions are
//	performed, then those triggered abilities are put on the stack,
//	then the active player gets priority. Players may cast spells and
//	activate abilities. Once the stack is empty and all players pass
//	in succession, another cleanup step begins."
//
// So the cleanup step has two exits and this file is both of them:
// the turn ends (exitCleanupStepLocked's ordinary path), or the
// cursor holds here with the active player on priority and the table
// comes back for another cleanup step once it has passed around
// (repeatCleanupStepLocked, from PassPriority's wrap).
//
// Before #661 there was no second exit: the cleanup hook advanced out
// of the turn as soon as nobody owed a discard, and a trigger queued
// by that discard sat on PendingTriggers until the next player's
// upkeep — the next priority boundary, because untap grants none
// either. Everything the trigger then asked about "this turn" read
// the wrong turn.

// exitCleanupStepLocked ends the cleanup step's turn-based actions,
// and is the one place that decides what happens next: the turn ends
// (CR 514.3), or the active player gets priority right here and
// another cleanup step follows (CR 514.3a).
//
// Two callers, which are the two ways one cleanup step's turn-based
// actions can finish:
//
//   - the StepCleanup case of the step-entry hook (game.go), and
//   - DiscardSelection (mutations.go), one CR 514.1 discard later,
//     once the pending map has drained.
//
// One function, because those two sites used to make the decision
// separately — the hook advanced the cursor, and the discard re-ran
// the whole step entry to reach the same line. The hand-size discard
// is the commonest way into CR 514.3a, so its resume has to make
// exactly the decision the hook makes.
//
// Caller must hold g.mu.
func (g *Game) exitCleanupStepLocked() {
	// CR 514.1 comes first: the cursor waits at cleanup, with nobody
	// holding priority, until the active player has discarded down to
	// their maximum hand size. DiscardSelection calls back here.
	if len(g.DiscardPending) > 0 {
		return
	}
	grant := g.cleanupGrantsPriorityLocked()
	// The check can take the active player out of the game, and a
	// departed active player's turn ends inside it through the
	// rotation seam (advancePastEliminatedLocked). Play has already
	// moved on; there is no cleanup step left to exit.
	if g.State != StateActive || g.Turn.Step != StepCleanup {
		return
	}
	if grant {
		// CR 514.3a: the active player gets priority IN the cleanup
		// step. The table passing it around brings it back to
		// repeatCleanupStepLocked below.
		if as := g.Turn.ActiveSeat; as >= 0 && as < len(g.Seats) &&
			g.Seats[as] != nil && !g.Seats[as].Eliminated {
			g.Turn.PriorityHolder = as
			return
		}
		// No active seat to hand it to — fall through and end the
		// turn rather than park the cursor on a window nobody holds.
	}
	// CR 514.3: nobody gets priority and the turn ends.
	g.advanceCursorLocked()
	g.runStepEntryHooksLocked()
}

// cleanupGrantsPriorityLocked performs the CR 514.3a check and
// reports whether the active player gets priority in this cleanup
// step.
//
// The condition is the rule's, both halves of it: "if any state-based
// actions are performed as a result of the state-based action check,
// OR if any triggered abilities are waiting to be put onto the
// stack". Either alone is enough. An SBA that puts nothing on the
// stack — an Aura falling off the creature an "until end of turn"
// animation stopped animating a line earlier, in the CR 514.2 sweep —
// still earns the table a priority window and a second cleanup step,
// and a waiting trigger earns one whether or not any SBA fired.
//
// A trigger held behind a prompt counts as waiting. An OPTIONAL
// trigger queues its CR 603.5 "you may" and reaches PendingTriggers
// only when the answer arrives (dispatchTriggerInstanceLocked), and a
// targeted one queues its pick the same way — so a cleanup that read
// PendingTriggers alone would walk the turn past Sangromancer's own
// prompt, which is the shape the #661 repro found. The engine's one
// "does this prompt stop the table" predicate answers it
// (blockingChoiceLocked, #730 / #794), and the prompt then holds the
// window open until it is answered, exactly as it does everywhere
// else.
//
// The work the rule asks for happens here whatever the answer turns
// out to be: the state-based actions are performed and the waiting
// triggers are put on the stack. runStateChecksLocked is the engine's
// one pairing of those two (CR 704.3 + 603.3b) and reports whether
// any state-based action fired; the trigger half is read off
// PendingTriggers before that drain empties it.
//
// Caller must hold g.mu.
func (g *Game) cleanupGrantsPriorityLocked() bool {
	waiting := len(g.PendingTriggers) > 0
	fired := g.runStateChecksLocked()
	if waiting || fired || g.blockingChoiceLocked() != nil {
		return true
	}
	// The stack is the belt to those braces. Nothing puts an item on
	// it during cleanup today except the drain above, whose triggers
	// `waiting` already counted — but a turn must never end with
	// something still on the stack, and the cost of asking is one map
	// length.
	return g.stackHasItemsLocked()
}

// repeatCleanupStepLocked begins ANOTHER cleanup step (CR 514.3a) —
// what "once the stack is empty and all players pass in succession"
// buys. Called from PassPriority's wrap in place of the step advance
// every other step's wrap gets.
//
// A fresh cleanup step, not a resumed one. Hand size is checked again
// (a trigger that drew cards can owe a second discard), the CR 514.2
// sweep runs again — which is what makes an "until end of turn"
// effect created by a trigger during cleanup expire in this turn
// instead of leaking into the next one (ADR 0035 §3) — and
// EventStepBegan is emitted again, because a step really did begin.
//
// The cursor does not move, so the event batch is opened here rather
// than in advanceCursorLocked: a step beginning is one of the two
// points where play moves on (#829), and this is the one step the
// rules begin again without the cursor moving. See event_batch.go,
// which names this as its third and only other caller.
//
// Caller must hold g.mu.
func (g *Game) repeatCleanupStepLocked() {
	g.beginEventBatchLocked()
	// The new step grants nobody priority until its own CR 514.3a
	// check says otherwise, exactly as the first one did.
	g.Turn.PriorityHolder = NoPriority
	g.runStepEntryHooksLocked()
}
