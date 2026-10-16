package game

// step_walk.go — an effect's later steps wait for an earlier step's
// prompt (#4012, CR 608.2c).
//
// # The bug this closes
//
// CR 608.2c: "The controller of the spell or ability follows its
// instructions in the order written." A catalog ability whose effect is
// a plain sequence of steps (effects.Do) ran them back to back on the Go
// stack. A step that ASKS something — a scry, a CR 616 ordering prompt
// on its damage or its counters, a prevention follow-up's question
// (#2040) — only queues the prompt and returns, so the next step ran
// before the answer. Vivi Ornitier dealt its damage while its +1/+1
// counter was still waiting for the Doubling Season / Hardened Scales
// order; a "scry 2, then draw" drew the old top card.
//
// #2789 fixed the same thing for a modal item's bullets (mode_walk.go),
// and this file is that walk one level down, for the steps of one
// effect.
//
// # The rule
//
// Before each step, the sequence asks whether the resolution is paused
// on one of its own prompts (ParkEffectStepsForEffect). If it is, the
// position of the next step is parked in Game.pausedStepWalk and the
// sequence returns. The resolution stays open (#1289), so the CR 704.3
// boundary is held. When the last of those prompts is answered, the
// boundary's hold check (holdForOpenResolutionLocked) resumes the steps
// through ResumeEffectSteps — after the damage instruction's own parked
// continuation (#2040), which is the earlier instruction, and before a
// modal item's next bullet and the resolution's tail, which are later
// ones. A resumed step may pause again.
//
// # Only the item's own effect parks
//
// The parked position is DATA — the stack item and a path of step
// indexes — and resuming it has to find the same steps again from that
// data. The effects package can do that only for a sequence that IS the
// item's Effect: a stamped ability's Effect is re-derived from its
// catalog row on restore (ability_ref.go), and the sequence answers a
// probe with its steps. A sequence called from inside a hand-written
// closure cannot be found again, so it does not park and keeps running
// on the Go stack, as before. The path has one index per level, so a
// step that is itself a conditional with its own steps (effects.If, as
// every intervening-if row wraps its steps) parks inside them too; the
// condition is not checked again on resume, because it was already
// checked as the step began.
//
// # What it costs: nothing, for restore points
//
// No closure is stored. Clone and the snapshot carry the walk
// (GameSnapshot.PausedStepWalk) the way they carry the mode walk, so a
// restore or an undo into the open prompt still runs the later steps
// after the answer. An item whose Effect is unkeyed is already counted
// by the census when the walk's item is snapshotted.

// stepWalk is the rest of the resolving item's effect steps, parked
// behind a prompt an earlier step queued.
type stepWalk struct {
	// item is the resolving stack item, whose Effect holds the steps.
	item *StackItem
	// path is the position of the next step to run: path[0] indexes
	// the Effect's steps, and each further index the steps of the
	// step before it.
	path []int
}

// ResumeEffectSteps runs the steps of item's Effect from the position
// path names, parking again if a step pauses. The effects package wires
// it at init (effects.Do); nil leaves a parked walk with nothing to
// run.
var ResumeEffectSteps func(g *Game, item *StackItem, path []int) error

// ParkEffectStepsForEffect parks the rest of item's effect steps at
// path when the resolution is paused on one of its own prompts, and
// reports whether it did. The caller returns at once when it did.
//
// Called by effects.Do, before each step of the sequence that is the
// item's own Effect. Caller holds g.mu (write).
func (g *Game) ParkEffectStepsForEffect(item *StackItem, path []int) bool {
	if item == nil || !g.resolutionOpen || g.pausedResolutionChoiceLocked() == nil {
		return false
	}
	if g.pausedStepWalk != nil {
		// One item resolves at a time and its walk is resumed before
		// anything else of it runs, so this cannot happen; running on
		// is the old behaviour, and losing the parked walk would not be.
		return false
	}
	g.pausedStepWalk = &stepWalk{item: item, path: copyInts(path)}
	return true
}

// resumeStepWalkLocked runs the parked steps, if any, inside a
// resolution function so the boundary stays held while they run.
// Reports whether there was a walk to resume.
//
// Caller must hold g.mu in write mode.
func (g *Game) resumeStepWalkLocked() bool {
	w := g.pausedStepWalk
	if w == nil {
		return false
	}
	g.pausedStepWalk = nil
	if g.State != StateActive || ResumeEffectSteps == nil {
		return true
	}
	g.beginResolutionLocked()
	defer g.endResolutionLocked()
	if err := ResumeEffectSteps(g, w.item, w.path); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    w.item.Controller,
			Source:   w.item.SourceCardID,
			ErrorMsg: err.Error(),
		})
	}
	return true
}

// cloneStepWalk copies the parked walk for Clone.
func cloneStepWalk(w *stepWalk) *stepWalk {
	if w == nil {
		return nil
	}
	return &stepWalk{item: cloneStackItem(w.item), path: copyInts(w.path)}
}

// stepWalkSnapshot is Game.pausedStepWalk on disk.
type stepWalkSnapshot struct {
	Item stackItemSnapshot `json:"item"`
	Path []int             `json:"path"`
}

func snapshotStepWalk(g *Game, w *stepWalk, cen *ContinuationCensus) *stepWalkSnapshot {
	if w == nil || w.item == nil {
		return nil
	}
	return &stepWalkSnapshot{Item: snapshotStackItem(g, w.item, cen), Path: copyInts(w.path)}
}

func restoreStepWalk(s *stepWalkSnapshot) *stepWalk {
	if s == nil {
		return nil
	}
	item, _ := restoreStackItem(&s.Item)
	return &stepWalk{item: item, path: copyInts(s.Path)}
}
