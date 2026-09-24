package game

// resolution_pause.go — state-based actions wait for a paused
// resolution to finish (#1289, CR 704.3, CR 608.2).
//
// # The bug this closes
//
// CR 704.3: "Whenever a player would get priority, the game checks for
// any of the listed conditions for state-based actions." Nothing is
// checked in the middle of a resolving spell or ability. Here a
// resolution that asks a question returns before it is finished: the
// prompt is queued, resolveTopOfStackLocked carries on and returns,
// and passPriorityLocked runs the CR 704.3 boundary straight after it.
// The rest of the resolution runs later, from the answer.
//
// So the sweep ran inside the resolution. On a Doubling Season plus
// Hardened Scales board an earthbend makes the land a 0/0 creature and
// then its counter placement pauses on the CR 616 ordering prompt. The
// sweep killed the 0/0 land, the earthbend return brought it back as a
// new object, and the counters had nothing to land on. A fresh amass
// Army died the same way.
//
// # The rule, and the one place it is enforced
//
// A resolution is OPEN from the moment an item begins to resolve until
// the CR 704.3 boundary after it runs. While it is open:
//
//  1. a prompt queued is stamped PendingChoice.midResolution, unless it
//     belongs to putting a triggered ability on the stack or to the
//     CR 726 loop breaker (choiceBelongsToResolution);
//  2. runStateChecksLocked does nothing while a resolution function is
//     still on the Go stack, or while a stamped prompt that blocks the
//     table is open (holdForOpenResolutionLocked). It holds the whole
//     boundary, not only the sweep: CR 603.3 puts triggered abilities
//     on the stack "the next time a player would receive priority",
//     which is the same moment.
//
// When the last stamped prompt is answered, the answer's own
// runStateChecksLocked finds nothing to hold for, closes the
// resolution and runs the boundary. The actions layer backs that up:
// Dispatch calls SettleResolution after every action, so an answer
// path that does not end in runStateChecksLocked still gets the
// boundary before anyone acts again. Concede settles too, because a
// departure can drop the last prompt a resolution was waiting on.
//
// # Why this cannot wedge a table
//
// Only a prompt that blocks the table holds (ChoicePromptBlocksTable).
// A blocking prompt already stops every pass, so holding the boundary
// behind it stops nothing that was not already stopped. A pay_unless
// asked after its ability has left the stack does not block, so it
// does not hold, and play goes on around it as ADR 0018 §6 decided.
// Every blocking kind has an always-legal answer (legal.choiceMoves),
// and a seat that leaves has its prompts dropped or reassigned by the
// departure table (ADR 0018 §6, ADR 0060). A dropped prompt no longer
// holds anything, and a reassigned one holds until its new chooser
// answers.

// choiceBelongsToResolution reports whether a prompt queued while a
// resolution is open is part of that resolution.
//
// Two families are not, because the rules place them after the
// resolution has finished:
//
//   - Putting a triggered ability on the stack (CR 603.3): the ordering
//     prompt, the optional-trigger yes/no, the CR 603.3c mode pick and
//     the CR 603.3d target pick. They are queued when the trigger is
//     harvested, which can be mid-resolution, but they are answered at
//     the priority boundary this file is holding. Holding for them
//     would put the boundary after them, the wrong way round (#809).
//   - The CR 726 loop shortcut. It is proposed when a resolution is
//     noted, but it is a question about the loop, not a step of the
//     resolving item.
//
// A pick_target for a spell COPY (copyResume, CR 707.10c) is part of
// the resolution that made the copy, so it is stamped.
func choiceBelongsToResolution(c *PendingChoice) bool {
	switch c.Kind {
	case PendingChoiceTriggerOrder, PendingChoiceLoopShortcut:
		return false
	}
	if c.triggerResume != nil || c.pickTargetResume != nil || c.modePickResume != nil {
		return false
	}
	return true
}

// beginResolutionLocked opens a resolution: the item has begun to
// resolve and its CR 704.3 boundary is owed. It also enters the
// resolution function; the caller defers endResolutionLocked. Caller
// must hold g.mu.
func (g *Game) beginResolutionLocked() {
	g.resolutionOpen = true
	g.resolutionDepth++
}

// endResolutionLocked leaves the resolution function. The resolution
// stays open until the boundary runs. Caller must hold g.mu.
func (g *Game) endResolutionLocked() {
	if g.resolutionDepth > 0 {
		g.resolutionDepth--
	}
}

// holdForOpenResolutionLocked reports whether the CR 704.3 boundary
// must wait, and closes the resolution when it need not. Caller must
// hold g.mu.
func (g *Game) holdForOpenResolutionLocked() bool {
	if g.resolutionDepth > 0 {
		return true
	}
	if !g.resolutionOpen {
		return false
	}
	if g.pausedResolutionChoiceLocked() != nil {
		return true
	}
	g.resolutionOpen = false
	return false
}

// pausedResolutionChoiceLocked returns the first open prompt holding
// the resolution open, or nil. Caller must hold g.mu.
func (g *Game) pausedResolutionChoiceLocked() *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.midResolution && g.ChoicePromptBlocksTable(c) {
			return c
		}
	}
	return nil
}

// ResolutionPaused reports whether a resolution is waiting on one of
// its own prompts, so state-based actions and the trigger drain are
// being held (#1289).
//
// Caller must NOT hold g.mu.
func (g *Game) ResolutionPaused() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.resolutionOpen && g.pausedResolutionChoiceLocked() != nil
}

// SettleResolution runs the CR 704.3 boundary a finished resolution is
// still owed. It is a no-op unless a resolution is open and none of its
// prompts is, which is the state an answer leaves behind when its path
// did not end in runStateChecksLocked. The actions layer calls it after
// every action (#1289).
//
// Caller must NOT hold g.mu.
func (g *Game) SettleResolution() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.settleResolutionLocked()
}

// settleResolutionLocked is SettleResolution under the caller's lock.
// Concede calls it too, because a departure can drop the prompt a
// resolution was paused on. Caller must hold g.mu.
func (g *Game) settleResolutionLocked() {
	if !g.resolutionOpen || g.State != StateActive {
		return
	}
	g.runStateChecksLocked()
}
