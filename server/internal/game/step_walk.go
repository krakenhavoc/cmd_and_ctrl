package game

import "github.com/google/uuid"

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
// # Two more sequences the engine can find again (#4019)
//
// A resolving spell's OnResolve written as steps (effects.ResolveSteps)
// is found again through the catalog key the spell resolved under: the
// walk records the key (stepWalk.key) and resumes through that key's
// CardDef.ResumeResolve. It parks only while the engine runs that
// spell's catalog hook (stepSites.spell, runSpellResolverLocked), so the
// same steps run from anywhere else run on as before.
//
// A modal bullet written as steps (effects.ModeSteps) is found again by
// the mode walk, which already finds the bullet from its key and
// occurrence. Its steps park through ParkBulletStepsForEffect while the
// walk runs that occurrence (stepSites.bullet), and the mode walk parks
// itself with that bullet first and the position inside it
// (modeWalk.inner, mode_walk.go). The rest of the bullet then runs
// before the next bullet.
//
// The sites are plain data set around one call, and the parked position
// is still an item, a key and a path, so no closure is stored and the
// closure ratchet does not move.
//
// # What it costs: nothing, for restore points
//
// No closure is stored. Clone and the snapshot carry the walk
// (GameSnapshot.PausedStepWalk) the way they carry the mode walk, so a
// restore or an undo into the open prompt still runs the later steps
// after the answer. An item whose Effect is unkeyed is already counted
// by the census when the walk's item is snapshotted.
//
// # Continuations park too (#4162)
//
// A continuation (a primitive's Then, an OnYes or OnAccept, the
// callback a *ThenForEffect method runs) is a Go closure stored in a
// prompt, so the walk above could not find it again. A continuation
// written as steps registers them under an on-disk key
// (effects.RegisterThenSteps, ledger lines "steps <key>"), and parks as
// that key and a path (ParkThenStepsForEffect, stepWalk.then). The
// closure that calls it still rides its own prompt as before; what is
// new is that the rest of the continuation, once the closure has run
// and one of its steps has asked, is data.
//
// A continuation runs while another walk may be parked (the item's own
// Effect, waiting behind the very prompt whose answer ran the
// continuation), and a step may run a continuation that parks before
// the step's own walk does. So the parked walks are a chain
// (stepWalk.outer), resumed top first, and a walk that parks goes
// directly above the mark taken as its last step began
// (parkStepWalkLocked): a continuation is an earlier instruction than
// the rest of the steps it belongs to (CR 608.2c).
//
// What the continuation needs from before the park (the number chosen,
// the creature sacrificed) is written to the per-resolution record
// (step_record.go, #4163), not captured, so the rest of it reads the
// same values after a restore.

// stepWalk is the rest of the resolving item's effect steps, parked
// behind a prompt an earlier step queued.
type stepWalk struct {
	// item is the resolving stack item, whose Effect holds the steps.
	item *StackItem
	// path is the position of the next step to run: path[0] indexes
	// the Effect's steps, and each further index the steps of the
	// step before it.
	path []int
	// key is empty when the steps are item's own Effect. Otherwise it
	// is the catalog key of the resolving SPELL whose OnResolve is a
	// step sequence (#4019, effects.ResolveSteps), and the steps are
	// found again through that key's CardDef.ResumeResolve.
	key string
	// then is set instead when the steps are a continuation's (#4162,
	// effects.RegisterThenSteps): the registered key the steps are
	// found again by, through ResumeThenSteps. occ is the bullet
	// occurrence the continuation ran in, plus one (0: none), so a
	// resumed step reads the same ctx.Occurrence().
	then string
	occ  int
	// outer is the walk parked beneath this one, which resumes after
	// it (#4162). See parkStepWalkLocked for the order.
	outer *stepWalk
}

// StepWalkMark is the parked walks as they stood when a step began
// (#4162). A walk that parks after that step goes directly above the
// mark: beneath every walk the step itself parked, which are earlier
// instructions (a continuation the step's prompt ran, a sequence the
// step called), and above every walk parked before the step began,
// which are later ones (the rest of the effect whose prompt the
// continuation answers). The effects package takes one before each step
// and hands it to the park call. It means something only inside the
// one call that took it.
type StepWalkMark struct{ w *stepWalk }

// StepWalkMarkForEffect is the mark for a step about to run.
//
// Caller holds g.mu.
func (g *Game) StepWalkMarkForEffect() StepWalkMark {
	return StepWalkMark{w: g.pausedStepWalk}
}

// parkStepWalkLocked puts w into the parked walks directly above mark.
// The walks above the mark are copied rather than relinked, because a
// clone may share them (restoring from a clone adopts its pointers). A
// mark that is no longer parked puts w at the bottom; it cannot happen,
// because nothing resumes a walk while a step runs.
//
// Caller holds g.mu (write).
func (g *Game) parkStepWalkLocked(w *stepWalk, mark StepWalkMark) {
	var above []*stepWalk
	n := g.pausedStepWalk
	for n != nil && n != mark.w {
		above = append(above, n)
		n = n.outer
	}
	w.outer = n
	top := w
	for i := len(above) - 1; i >= 0; i-- {
		c := *above[i]
		c.outer = top
		top = &c
	}
	g.pausedStepWalk = top
}

// stepsMayParkLocked reports whether the resolution is paused on one of
// its own prompts, so a walk about to run its next step parks.
//
// Caller holds g.mu.
func (g *Game) stepsMayParkLocked(item *StackItem) bool {
	return item != nil && g.resolutionOpen && g.pausedResolutionChoiceLocked() != nil
}

// stepSites is the transient half of #4019: which step sequences, other
// than an item's own Effect, may park right now. The engine sets each
// around the one call that runs it, so a sequence run from anywhere else
// (a hand-written closure, a spliced text, a fused half) finds no site
// and runs on as before.
type stepSites struct {
	// spell is the resolving spell's item ID while its catalog OnResolve
	// runs (runSpellResolverLocked), and spellKey the catalog key that
	// ran it.
	spell    uuid.UUID
	spellKey string
	// bullet is the item ID whose chosen bullet the mode walk is
	// running and occ that bullet's occurrence. bulletPark is the
	// position inside it where the bullet's steps parked (nil when they
	// did not).
	bullet     uuid.UUID
	occ        int
	bulletPark []int
}

// ResumeEffectSteps runs the steps of item's Effect from the position
// path names, parking again if a step pauses. The effects package wires
// it at init (effects.Do); nil leaves a parked walk with nothing to
// run.
var ResumeEffectSteps func(g *Game, item *StackItem, path []int) error

// ParkEffectStepsForEffect parks the rest of item's effect steps at
// path when the resolution is paused on one of its own prompts, and
// reports whether it did. The caller returns at once when it did. mark
// is the parked walks as they stood when the step before path began
// (#4162): a continuation that step's prompt ran may have parked since,
// and it resumes first.
//
// Called by effects.Do, before each step of the sequence that is the
// item's own Effect. Caller holds g.mu (write).
func (g *Game) ParkEffectStepsForEffect(item *StackItem, path []int, mark StepWalkMark) bool {
	if !g.stepsMayParkLocked(item) {
		return false
	}
	g.parkStepWalkLocked(&stepWalk{item: item, path: copyInts(path)}, mark)
	return true
}

// ParkSpellStepsForEffect is ParkEffectStepsForEffect for the steps of
// a resolving spell's OnResolve (#4019, effects.ResolveSteps). It parks
// only while the engine runs that spell's catalog hook, and only when
// the key it ran can find the steps again (CardDef.ResumeResolve): a
// fused split spell's composite key cannot, so its halves run on as
// before.
//
// Caller holds g.mu (write).
func (g *Game) ParkSpellStepsForEffect(item *StackItem, path []int, mark StepWalkMark) bool {
	if item == nil || item.ID != g.stepSites.spell || g.stepSites.spellKey == "" {
		return false
	}
	if d := catalogDef(g.stepSites.spellKey); d == nil || d.ResumeResolve == nil {
		return false
	}
	if !g.stepsMayParkLocked(item) {
		return false
	}
	g.parkStepWalkLocked(&stepWalk{item: item, path: copyInts(path), key: g.stepSites.spellKey}, mark)
	return true
}

// ParkThenStepsForEffect parks the rest of a continuation written as
// steps (#4162, effects.RegisterThenSteps) when the resolution is paused
// on one of its own prompts, and reports whether it did. The steps are
// found again by their registered key, so the continuation may run from
// anywhere: an answered prompt's Then, a hand-written body, a delayed
// trigger's keyed body. occ is the bullet occurrence it ran in, plus one
// (0: none).
//
// Caller holds g.mu (write).
func (g *Game) ParkThenStepsForEffect(item *StackItem, key string, occ int, path []int, mark StepWalkMark) bool {
	if !KnownThenSteps(key) || !g.stepsMayParkLocked(item) {
		return false
	}
	g.parkStepWalkLocked(&stepWalk{item: item, path: copyInts(path), then: key, occ: occ}, mark)
	return true
}

// ResumeThenSteps runs a continuation's registered steps from the
// position path names, parking again if a step pauses. The effects
// package wires it at init (effects.RegisterThenSteps); nil leaves a
// parked continuation with nothing to run.
var ResumeThenSteps func(g *Game, item *StackItem, key string, occ int, path []int) error

// ParkBulletStepsForEffect parks the rest of one chosen bullet's steps
// (#4019, effects.ModeSteps) when the resolution is paused on one of its
// own prompts, and reports whether it did. It parks only while the mode
// walk runs that occurrence of item's bullets; the walk then parks
// itself with this occurrence first and the position inside it
// (modeWalk.inner), so the bullet's later steps run before the next
// bullet.
//
// Caller holds g.mu (write).
func (g *Game) ParkBulletStepsForEffect(item *StackItem, occurrence int, path []int) bool {
	s := &g.stepSites
	if item == nil || s.bullet == uuid.Nil || item.ID != s.bullet || occurrence != s.occ || s.bulletPark != nil {
		return false
	}
	if !g.resolutionOpen || g.pausedResolutionChoiceLocked() == nil {
		return false
	}
	s.bulletPark = copyInts(path)
	return true
}

// ResumeBulletSteps runs one chosen bullet's steps from the position
// path names, parking again if a step pauses. The effects package wires
// it at init (effects.ModeSteps) and finds the steps from the bullet's
// own body; nil leaves the bullet with nothing to run.
var ResumeBulletSteps func(g *Game, item *StackItem, bullet func(*Game, *StackItem, int) error, occurrence int, path []int) error

// runSpellResolverLocked runs a spell's catalog OnResolve with the
// spell's step site open, so a step sequence it runs may park
// (ParkSpellStepsForEffect). The previous site comes back after it,
// which is only ever the empty one.
//
// Caller holds g.mu (write).
func (g *Game) runSpellResolverLocked(item *StackItem, key string, run func() error) error {
	prevID, prevKey := g.stepSites.spell, g.stepSites.spellKey
	g.stepSites.spell, g.stepSites.spellKey = uuid.Nil, ""
	if item != nil {
		g.stepSites.spell, g.stepSites.spellKey = item.ID, key
	}
	defer func() { g.stepSites.spell, g.stepSites.spellKey = prevID, prevKey }()
	return run()
}

// resumeStepWalkLocked runs the parked steps on top, if any, inside a
// resolution function so the boundary stays held while they run.
// Reports whether there was a walk to resume. The walks beneath it stay
// parked (#4162); resumeStepWalksLocked runs them in turn.
//
// The prevention follow-ups still owed run first (#2040): a CR 615.13
// announcement (#2047) or a follow-up held behind the answered question
// was owed by an earlier step's damage, so it happens before the next
// step (CR 615.5, CR 608.2c). Every entry owed now is an earlier step's:
// the walk has not resumed, and an earlier resolution's entries were
// flushed at the priority boundary before this item resolved. If one of
// them asks, the walk stays parked behind the new question.
//
// Caller must hold g.mu in write mode.
func (g *Game) resumeStepWalkLocked() bool {
	w := g.pausedStepWalk
	if w == nil {
		return false
	}
	if g.flushPreventionFollowUpsLocked() {
		return true
	}
	g.pausedStepWalk = w.outer
	if g.State != StateActive {
		return true
	}
	g.beginResolutionLocked()
	defer g.endResolutionLocked()
	if err := g.resumeStepsLocked(w); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    w.item.Controller,
			Source:   w.item.SourceCardID,
			ErrorMsg: err.Error(),
		})
	}
	return true
}

// resumeStepWalksLocked resumes the parked walks top first, until one
// pauses again or none is left, and reports whether the resolution is
// paused on a prompt afterwards. A continuation's rest (#4162) resumes
// before the steps whose prompt it continued.
//
// Caller must hold g.mu in write mode.
func (g *Game) resumeStepWalksLocked() bool {
	for g.pausedStepWalk != nil {
		top := g.pausedStepWalk
		g.resumeStepWalkLocked()
		if g.pausedResolutionChoiceLocked() != nil {
			return true
		}
		if g.pausedStepWalk == top {
			// A follow-up asked something that does not pause the
			// resolution: the walk stays parked, as it did before.
			return false
		}
	}
	return false
}

// resumeStepsLocked finds the parked steps again from the walk's data
// and runs them from its path: an item's own Effect through
// ResumeEffectSteps, a spell's OnResolve through its key's
// CardDef.ResumeResolve, with the spell's step site open again so a
// resumed step may park, and a continuation through its registered key
// (ResumeThenSteps, #4162).
//
// Caller must hold g.mu in write mode.
func (g *Game) resumeStepsLocked(w *stepWalk) error {
	if w.then != "" {
		if ResumeThenSteps == nil {
			return nil
		}
		return ResumeThenSteps(g, w.item, w.then, w.occ, w.path)
	}
	if w.key == "" {
		if ResumeEffectSteps == nil {
			return nil
		}
		return ResumeEffectSteps(g, w.item, w.path)
	}
	d := catalogDef(w.key)
	if d == nil || d.ResumeResolve == nil {
		return nil
	}
	return g.runSpellResolverLocked(w.item, w.key, func() error {
		return d.ResumeResolve(g, w.item, w.path)
	})
}

// cloneStepWalk copies the parked walks for Clone.
func cloneStepWalk(w *stepWalk) *stepWalk {
	if w == nil {
		return nil
	}
	return &stepWalk{item: cloneStackItem(w.item), path: copyInts(w.path), key: w.key,
		then: w.then, occ: w.occ, outer: cloneStepWalk(w.outer)}
}

// stepWalkSnapshot is Game.pausedStepWalk on disk.
type stepWalkSnapshot struct {
	Item stackItemSnapshot `json:"item"`
	Path []int             `json:"path"`
	// Key is stepWalk.key (#4019): the catalog key of the spell whose
	// OnResolve steps are parked. Additive within v9: a file without it
	// parked an item's own Effect, the only kind there was.
	Key string `json:"key,omitempty"`
	// Then, Occ and Outer are stepWalk.then, .occ and .outer (#4162): a
	// continuation's registered key, the bullet occurrence it ran in,
	// and the walk parked beneath this one. Additive within v9: a file
	// without them parked one walk, and not a continuation's.
	Then  string            `json:"then,omitempty"`
	Occ   int               `json:"occ,omitempty"`
	Outer *stepWalkSnapshot `json:"outer,omitempty"`
}

func snapshotStepWalk(g *Game, w *stepWalk, cen *ContinuationCensus) *stepWalkSnapshot {
	if w == nil || w.item == nil {
		return nil
	}
	return &stepWalkSnapshot{Item: snapshotStackItem(g, w.item, cen), Path: copyInts(w.path), Key: w.key,
		Then: w.then, Occ: w.occ, Outer: snapshotStepWalk(g, w.outer, cen)}
}

func restoreStepWalk(s *stepWalkSnapshot) *stepWalk {
	if s == nil {
		return nil
	}
	item, _ := restoreStackItem(&s.Item)
	return &stepWalk{item: item, path: copyInts(s.Path), key: s.Key,
		then: s.Then, occ: s.Occ, outer: restoreStepWalk(s.Outer)}
}

// unknownThenSteps lists the continuation keys the parked walks name
// that this binary has not registered (#4162), for checkEffectKeys: a
// walk restored without its steps would run nothing after the answer.
func unknownThenSteps(s *stepWalkSnapshot) []string {
	var out []string
	for ; s != nil; s = s.Outer {
		if s.Then != "" && !KnownThenSteps(s.Then) {
			out = append(out, "continuation steps "+s.Then)
		}
	}
	return out
}
