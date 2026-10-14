package game

// resolution_tail.go — a resolution that asks something finishes after
// the answer (#3613, CR 608.2c, CR 608.2n).
//
// # The bug this closes
//
// A spell or ability that asks a question as it resolves (fabricate's
// counters-or-Servos, Council's Judgment's vote, a library search) only
// QUEUES the prompt; the resolution function carries on and returns.
// Everything after the effect therefore ran before the answer: the
// spell went to its owner's graveyard with its vote still open, and
// EventResolve, which fires as the item begins to resolve, put
// "… resolved" in the log above the prompt. CR 608.2c follows the
// instructions in order and CR 608.2n puts an instant or sorcery into
// the graveyard as the FINAL step of its resolution, so both belong
// after the answer.
//
// # The rule, and where it is enforced
//
// One choke point per resolution function, and one resume. When
// resolveTopOfStackLocked or resolveTopAbilityLocked has run the
// item's effect and its bullets and the resolution is paused on one of
// its own prompts (pausedResolutionChoiceLocked, #1289, or a bullet walk
// parked behind one, #2789), holdResolutionTailLocked parks the rest:
//
//   - for a SPELL, everything the frame does after the effect — the
//     battlefield entry of a permanent spell, a copy ceasing to exist,
//     the graveyard (or flashback, buyback, rebound, Adventure) exit —
//     in finishSpellResolutionLocked. The card stays on the stack;
//   - for both, the log's "… resolved" line: the EventResolve already
//     emitted is marked LogDeferred, and EventResolutionFinished
//     carries the line when the tail runs.
//
// The boundary's hold check (holdForOpenResolutionLocked) runs the tail
// once the last stamped prompt is answered and any parked bullets have
// run, inside a resolution function so a prompt the tail asks (the
// CR 616.1 exit choice, a Clone's copy pick) is stamped and holds the
// boundary in turn.
//
// EventResolve itself does not move. Card code (batch11/12/13/27
// helpers) and the turn tally read it as "this item began resolving",
// before its effect's events; moving it would break them. Only the
// log line moves, which is the part a player reads.
//
// # The stack while it waits
//
// The spell's StackMeta entry is gone, as before: the object is
// resolving, not waiting to resolve, so nothing may target it or
// respond to it (and the prompt blocks the table anyway). The card is
// still in the Stack zone, and the view lists the held item in
// stack_items (HeldResolutionItemForView) so the table sees what is
// asking. A copy made from the resolving spell (CR 707.10) still finds
// it through the resolving slot (resolving_item.go), because
// stackSpellLocked falls back there when the card has no meta.
//
// # Restore points
//
// The tail is data — the stack item — and Clone and the snapshot carry
// it, like the parked bullet walk, so an undo or a restore into the
// open prompt still finishes the resolution after the answer.

// resolutionTail is the end of a resolution parked behind one of its
// own prompts.
type resolutionTail struct {
	// item is the resolving stack item. A spell's item ID is its card's
	// instance ID.
	item *StackItem
}

// isSpell reports whether the parked resolution is a spell's, which has
// a card on the stack to route away.
func (t *resolutionTail) isSpell() bool {
	return t.item.Kind != StackItemActivated && t.item.Kind != StackItemTriggered
}

// holdResolutionTailLocked parks the rest of item's resolution when it
// is paused on one of its own prompts, and reports whether it did.
// resolveSeq is the Seq of the item's EventResolve, whose log line moves
// to EventResolutionFinished.
//
// Caller must hold g.mu in write mode, inside the resolution function.
func (g *Game) holdResolutionTailLocked(item *StackItem, resolveSeq uint64) bool {
	if item == nil || !g.resolutionOpen {
		return false
	}
	if g.pausedModeWalk == nil && g.pausedResolutionChoiceLocked() == nil {
		return false
	}
	g.deferResolveLogLineLocked(resolveSeq)
	g.pausedResolutionTail = &resolutionTail{item: item}
	return true
}

// deferResolveLogLineLocked marks the EventResolve with Seq seq
// LogDeferred, so the public log writes no line for it.
//
// The one in-place edit of the event log outside a restore. The event
// was emitted under the write lock this resolution still holds, so no
// view has folded it yet; the generation is bumped anyway, so the log
// projection's cache (projection_cache.go) never has to rely on that
// and simply refolds. A paused resolution is a prompt, so the refold
// costs one fold per question asked, not one per event.
//
// Caller must hold g.mu in write mode.
func (g *Game) deferResolveLogLineLocked(seq uint64) {
	for i := len(g.Events) - 1; i >= 0 && g.Events[i].Seq >= seq; i-- {
		if g.Events[i].Seq == seq && g.Events[i].Kind == EventResolve {
			g.Events[i].LogDeferred = true
			g.eventLogGen++
			return
		}
	}
}

// finishResolutionTailLocked runs the parked tail, if any, and reports
// whether there was one. It waits for parked bullets: those are earlier
// instructions than anything the tail does.
//
// Caller must hold g.mu in write mode.
func (g *Game) finishResolutionTailLocked() bool {
	t := g.pausedResolutionTail
	if t == nil || g.pausedModeWalk != nil {
		return false
	}
	g.pausedResolutionTail = nil
	if g.State != StateActive {
		return true
	}
	g.beginResolutionLocked()
	defer g.endResolutionLocked()
	defer g.recomputeSplitSecondLocked()
	g.EmitEvent(resolutionFinishedEvent(t))
	if !t.isSpell() {
		return true
	}
	card, ok := g.stackCardLocked(t.item.ID)
	if !ok || card == nil {
		// #489: the spell moved itself while it waited, or left with a
		// departing owner. Nothing is left on the stack to route.
		return true
	}
	if err := g.finishSpellResolutionLocked(*card, t.item); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    t.item.Controller,
			Source:   t.item.ID,
			ErrorMsg: err.Error(),
		})
	}
	return true
}

// resolutionFinishedEvent mirrors the EventResolve the tail's item
// emitted as it began to resolve.
func resolutionFinishedEvent(t *resolutionTail) Event {
	it := t.item
	if t.isSpell() {
		return Event{Kind: EventResolutionFinished, Actor: it.Controller, Source: it.ID, CardID: it.ID}
	}
	return Event{
		Kind:                EventResolutionFinished,
		Actor:               it.Controller,
		Source:              it.SourceCardID,
		Label:               it.Label,
		ResolvedStackItemID: it.ID,
	}
}

// HeldResolutionItemForView is the stack item whose resolution is
// waiting on one of its own prompts, or nil (#3613). The view lists it
// on the stack: the object has not finished resolving. A spell that
// moved itself off the stack while it waited is not listed.
//
// Caller must hold g.mu (the view projection does).
func (g *Game) HeldResolutionItemForView() *StackItem {
	t := g.pausedResolutionTail
	if t == nil || t.item == nil {
		return nil
	}
	if t.isSpell() && (g.Stack == nil || !g.Stack.Contains(t.item.ID)) {
		return nil
	}
	return t.item
}

// cloneResolutionTail copies the parked tail for Clone.
func cloneResolutionTail(t *resolutionTail) *resolutionTail {
	if t == nil {
		return nil
	}
	return &resolutionTail{item: cloneStackItem(t.item)}
}

// resolutionTailSnapshot is Game.pausedResolutionTail on disk.
type resolutionTailSnapshot struct {
	Item stackItemSnapshot `json:"item"`
}

func snapshotResolutionTail(g *Game, t *resolutionTail, cen *ContinuationCensus) *resolutionTailSnapshot {
	if t == nil || t.item == nil {
		return nil
	}
	return &resolutionTailSnapshot{Item: snapshotStackItem(g, t.item, cen)}
}

func restoreResolutionTail(s *resolutionTailSnapshot) *resolutionTail {
	if s == nil {
		return nil
	}
	item, _ := restoreStackItem(&s.Item)
	if item == nil {
		return nil
	}
	return &resolutionTail{item: item}
}
