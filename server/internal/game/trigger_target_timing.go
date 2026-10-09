package game

// trigger_target_timing.go holds one rule (#809): a targeted
// triggered ability's legal-target set belongs to the moment the
// ability goes on the stack, not to the moment the event fired.
//
// The two are not the same moment. triggerHarvester runs inside the
// resolving object's own EmitEvent calls, so an ETB trigger raised by
// Persist returning Dualcaster Mage is dispatched while Persist is
// STILL on the stack — CR 608.2n does not put a resolving instant or
// sorcery into its owner's graveyard until the final step of its
// resolution. queuePickTargetLocked froze that stack into the prompt,
// and the player was then asked to pick a spell that had ceased to be
// one by the time they could answer: ResolvePickTargets re-validates
// against the current board (as it should) and returns
// ErrIllegalTarget, nothing withdrew the prompt, and since #791 an
// unanswered prompt gates the whole table. That is a permanent wedge.
//
// The rules:
//
//	CR 603.3   a triggered ability is put on the stack the next time
//	           a player would receive priority — not while a spell is
//	           still resolving.
//	CR 603.3d  its targets are chosen as it is put on the stack; if
//	           it has no legal targets then, it is removed from the
//	           stack and does nothing.
//	CR 608.2n  a resolving instant or sorcery goes to its owner's
//	           graveyard as the final step of its resolution.
//
// runStateChecksLocked IS that priority-grant boundary — it is the
// paired CR 704.3 / 603.3b sweep every path runs before anyone gets
// priority again. So the frozen set is re-read there, against the
// board as it stands, and a prompt with nothing left to offer is
// withdrawn rather than left for a player who cannot answer it.
//
// Widening is as much a part of this as emptying: with a Lightning
// Bolt still under the spell that resolved, the prompt must offer the
// Bolt and not the spell that has gone to the graveyard.

// refreshTargetChoicesLocked re-reads every outstanding pick_target
// prompt's legal set against the current board and withdraws the ones
// that no longer have a legal option.
//
// Two prompt shapes carry a target spec:
//
//   - a targeted TRIGGER (pickTargetResume). Empty set ⇒ CR 603.3d:
//     the ability is removed and does nothing, so the prompt is
//     dropped and no item reaches the stack.
//   - the CR 707.10c "you may choose new targets for the copy"
//     re-target (copyResume), for a spell copy or an ability copy.
//     Its offer is the open step's legal new targets plus the step's
//     original ones (#2622). Nothing left to change to ⇒ the step keeps
//     its targets, which is the same state offerCopyTargetsLocked
//     handles at queue time, and the walk goes on to the next step or
//     creates the copy. It is NOT dropped — the copy exists either way,
//     and losing it would be a second bug in place of the first.
//
// Both are engine withdrawals of a prompt nobody answered, so they go
// through dropChoiceLocked rather than dequeueChoiceLocked: no player
// made a decision here.
//
// Caller must hold g.mu.
func (g *Game) refreshTargetChoicesLocked() {
	if len(g.PendingChoices) == 0 {
		return
	}
	// Deferred so the copy lands after the scan: createSpellCopyLocked
	// emits EventBecomesTarget, which can harvest further triggers and
	// queue further prompts, and a harvest running inside this walk
	// would be mutating the slice it is iterating. The same goes for a
	// walk's next step, which queues a prompt.
	var keepStepTargets []*copyFrame
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c == nil || c.Kind != PendingChoicePickTarget {
			continue
		}
		if cf := c.copyResume; cf != nil {
			if pos := cf.stepRefs(); len(pos) > 0 {
				if lt, fresh := g.copyStepCandidatesLocked(cf, pos); fresh {
					c.PickTargetPlayers, c.PickTargetCards = lt.Players, lt.Cards
					continue
				}
			}
			keepStepTargets = append(keepStepTargets, cf)
			g.dropChoiceLocked(i)
			continue
		}
		f := c.pickTargetResume
		if f == nil {
			continue
		}
		// #764: the clause the OPEN step is asking about, not the
		// ability's first — a multi-clause trigger re-reads whichever
		// one the prompt belongs to.
		spec := f.currentClause()
		if spec == nil {
			continue
		}
		// #662: the SOURCE, not just the chooser. The frame keeps a
		// value copy of the object whose ability is picking, which is
		// what CR 702.16b tests the quality against.
		lt := g.legalTargetsLocked(SourceObject(f.source.Controller, &f.source), spec)
		if len(lt.Players) > 0 || len(lt.Cards) > 0 {
			c.PickTargetPlayers, c.PickTargetCards = lt.Players, lt.Cards
			continue
		}
		g.dropChoiceLocked(i)
	}
	for _, cf := range keepStepTargets {
		// The step keeps its targets; the walk asks the next one or
		// creates the copy through createCopyLocked — not the spell
		// builder: the frame may be an ABILITY copy's (#1223), whose
		// `src` is the source permanent, and the spell builder would put
		// a spell copy of that card on the stack (#1449).
		cf.step++
		g.queueCopyTargetStepLocked(cf)
	}
}
