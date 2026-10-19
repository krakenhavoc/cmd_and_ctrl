package game

// mode_walk.go — a modal spell or ability's later bullets wait for an
// earlier bullet's prompt (#2789, CR 608.2c).
//
// # The bug this closes
//
// CR 608.2c: "The controller of the spell or ability follows its
// instructions in the order written." runChosenModeEffectsLocked has
// walked the chosen bullets in printed order since #1653, but a bullet
// that ASKS something — a library search, a scry, a discard — only
// queues the prompt and returns. The walk then ran the next bullet on
// the line below, before the answer. Titania's Command put its +1/+1
// counters on the board while its land search was still open, so a
// creature land it found missed them; Insatiable Avarice drew its three
// cards before the tutored card had been put on top.
//
// # The rule
//
// Before each bullet, the walk asks whether the resolution is paused
// on one of its own prompts (pausedResolutionChoiceLocked, #1289). If
// it is, the bullets not yet run are parked in Game.pausedModeWalk and
// the walk returns. The resolution stays OPEN — that is #1289's rule —
// so the CR 704.3 boundary is held, and a prompt the answer's
// continuation chains on (Cultivate's second search) is stamped as part
// of the resolution too and keeps the walk parked.
//
// When the last of those prompts is answered (or withdrawn), the
// boundary's hold check finds nothing to wait for and, instead of
// closing the resolution, resumes the walk (holdForOpenResolutionLocked).
// A resumed bullet may pause again; the boundary runs only once every
// bullet has.
//
// Checking BEFORE each bullet rather than after it means an OnResolve
// or a trigger body that paused before the bullets holds them too, which
// is the same rule: the earlier instruction is not finished.
//
// # What it costs: nothing, for restore points
//
// The parked walk is data: the stack item (through the ordinary stack
// item mirror, which re-derives a spell's ModeSpec from its oracle ID
// and a stamped ability's from its catalog row), the catalog key the
// spell's bullets were read from, and the occurrence indexes still to
// run. The snapshot carries it (GameSnapshot.PausedModeWalk) and Clone
// carries it, so neither a restore nor an undo into the open prompt
// loses the bullets after it. Whether the prompt itself is a restore
// point is the prompt's own business.

// modeWalk is the rest of a modal item's bullets, parked behind a
// prompt one of its earlier instructions queued.
type modeWalk struct {
	// item is the resolving stack item. Clone copies it; the bullets
	// read only its data (modes, targets, X, paid costs).
	item *StackItem
	// key is the catalog key a SPELL's bullets were read from
	// (ModeSpecFor(CatalogKey(card))), so a face or a split half reads
	// the same bullets after a restore. Empty for an ability, whose
	// bullets are its own item.modeSpec.
	key string
	// rest is the occurrence indexes still to run, in printed order.
	rest []int
	// splices is the spliced texts still to run after them (ADR 0202,
	// CR 702.47b), as indexes into item.Spliced, in order.
	splices []int
	// inner is the position inside rest[0]'s bullet where its steps
	// parked (#4019, effects.ModeSteps): that bullet has begun, and
	// resumes from there. Nil when rest[0] has not begun.
	inner []int
}

// spec is the ModeSpec the parked bullets come from.
func (w *modeWalk) spec() *ModeSpec {
	if w.key != "" {
		if ms := ModeSpecFor(w.key); ms != nil {
			return ms
		}
	}
	return w.item.modeSpec
}

// runModeOccurrencesLocked runs the occurrences in `occs` (printed
// order) of item's chosen bullets, read from `ms` (`key` names where a
// spell's ms came from, for the parked walk), parking the remainder in
// g.pausedModeWalk as soon as the resolution is paused on a prompt.
//
// Caller must hold g.mu in write mode.
func (g *Game) runModeOccurrencesLocked(item *StackItem, ms *ModeSpec, key string, occs []int) {
	g.runBodyWalkLocked(item, ms, key, occs, nil, nil)
}

// runSpellBodiesLocked runs a resolving spell's chosen bullets, then its
// spliced texts (ADR 0202, CR 702.47b: "The effects of the main spell
// must happen first"). `ownText` false skips the bullets: the spell's
// own text had targets and every one is illegal, so it does nothing
// (CR 608.2b) while a spliced text with a legal target still resolves.
//
// Caller must hold g.mu in write mode, inside the resolution function.
func (g *Game) runSpellBodiesLocked(item *StackItem, ms *ModeSpec, key string, ownText bool) {
	if item == nil {
		return
	}
	var occs []int
	if ms != nil && ownText {
		occs = PrintedModeOrder(item.Modes)
	}
	var splices []int
	for k := range item.Spliced {
		splices = append(splices, k)
	}
	g.runBodyWalkLocked(item, ms, key, occs, splices, nil)
}

// runBodyWalkLocked is the walk itself: the occurrences, then the
// spliced texts, parking whatever is left as soon as the resolution is
// paused on a prompt. `inner`, when set, is the position inside occs[0]'s
// bullet to resume its steps from (#4019).
//
// A bullet written as steps (effects.ModeSteps) parks inside itself
// (ParkBulletStepsForEffect): the walk then parks with that bullet
// first and the position inside it, so the bullet's later steps run
// before the next bullet (CR 608.2c).
//
// Caller must hold g.mu in write mode.
func (g *Game) runBodyWalkLocked(item *StackItem, ms *ModeSpec, key string, occs, splices, inner []int) {
	for i, occ := range occs {
		resume := i == 0 && inner != nil
		if !resume && g.resolutionOpen && g.pausedResolutionChoiceLocked() != nil {
			g.pausedModeWalk = &modeWalk{item: item, key: key, rest: append([]int(nil), occs[i:]...), splices: copyInts(splices)}
			return
		}
		opt := item.Modes[occ]
		if opt < 0 || opt >= len(ms.Options) {
			continue
		}
		fn := ms.Options[opt].Effect
		if fn == nil {
			continue
		}
		if parked := g.runBulletLocked(item, fn, occ, inner, resume); parked != nil {
			g.pausedModeWalk = &modeWalk{item: item, key: key, rest: append([]int(nil), occs[i:]...), splices: copyInts(splices), inner: parked}
			return
		}
	}
	for i, k := range splices {
		if g.resolutionOpen && g.pausedResolutionChoiceLocked() != nil {
			g.pausedModeWalk = &modeWalk{item: item, key: key, splices: append([]int(nil), splices[i:]...)}
			return
		}
		g.runSplicedTextLocked(item, k)
	}
}

// runBulletLocked runs one chosen bullet with its step site open
// (#4019): from its start, or, when `resume`, its steps from the
// position `inner` names (ResumeBulletSteps). It returns the position
// inside the bullet where its steps parked, or nil when they did not.
// An error is logged and the walk goes on, as it always has.
//
// Caller must hold g.mu in write mode.
func (g *Game) runBulletLocked(item *StackItem, fn func(*Game, *StackItem, int) error, occ int, inner []int, resume bool) []int {
	prev := g.stepSites
	g.stepSites.bullet, g.stepSites.occ, g.stepSites.bulletPark = item.ID, occ, nil
	var err error
	switch {
	case !resume:
		err = fn(g, item, occ)
	case ResumeBulletSteps != nil:
		err = ResumeBulletSteps(g, item, fn, occ, inner)
	}
	parked := g.stepSites.bulletPark
	g.stepSites.bullet, g.stepSites.occ, g.stepSites.bulletPark = prev.bullet, prev.occ, prev.bulletPark
	if err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    item.Controller,
			Source:   item.SourceCardID,
			ErrorMsg: err.Error(),
		})
	}
	return parked
}

// resumeModeWalkLocked runs the parked bullets, if any, inside a
// resolution function so the boundary stays held while they run.
// Reports whether there was a walk to resume.
//
// The prevention follow-ups still owed run first (#4070, ADR 0108): a
// CR 615.13 announcement (#2047, Selfless Squire) or a follow-up held
// behind the answered question was owed by an earlier bullet's damage,
// so it happens before the next bullet (CR 615.5, CR 608.2c), as
// resumeStepWalkLocked does for the next step. If one of them asks, the
// walk stays parked behind the new question.
//
// Caller must hold g.mu in write mode.
func (g *Game) resumeModeWalkLocked() bool {
	w := g.pausedModeWalk
	if w == nil {
		return false
	}
	if g.flushPreventionFollowUpsLocked() {
		return true
	}
	g.pausedModeWalk = nil
	if g.State != StateActive {
		return true
	}
	g.beginResolutionLocked()
	defer g.endResolutionLocked()
	ms := w.spec()
	if ms == nil {
		w.rest, w.inner = nil, nil
	}
	g.runBodyWalkLocked(w.item, ms, w.key, w.rest, w.splices, w.inner)
	return true
}

// cloneModeWalk copies the parked walk for Clone.
func cloneModeWalk(w *modeWalk) *modeWalk {
	if w == nil {
		return nil
	}
	return &modeWalk{item: cloneStackItem(w.item), key: w.key, rest: append([]int(nil), w.rest...), splices: copyInts(w.splices), inner: copyInts(w.inner)}
}

// modeWalkSnapshot is Game.pausedModeWalk on disk.
type modeWalkSnapshot struct {
	Item stackItemSnapshot `json:"item"`
	Key  string            `json:"key,omitempty"`
	Rest []int             `json:"rest"`
	// Splices is modeWalk.splices (ADR 0202): the spliced texts still to
	// run. Additive within v9; a file without it has none.
	Splices []int `json:"splices,omitempty"`
	// Inner is modeWalk.inner (#4019): the position inside the first
	// parked bullet where its steps parked. Additive within v9; a file
	// without it parked between bullets, the only kind there was.
	Inner []int `json:"inner,omitempty"`
}

func snapshotModeWalk(g *Game, w *modeWalk, cen *ContinuationCensus) *modeWalkSnapshot {
	if w == nil || w.item == nil {
		return nil
	}
	return &modeWalkSnapshot{Item: snapshotStackItem(g, w.item, cen), Key: w.key, Rest: copyInts(w.rest), Splices: copyInts(w.splices), Inner: copyInts(w.inner)}
}

func restoreModeWalk(s *modeWalkSnapshot) *modeWalk {
	if s == nil {
		return nil
	}
	item, _ := restoreStackItem(&s.Item)
	return &modeWalk{item: item, key: s.Key, rest: copyInts(s.Rest), splices: copyInts(s.Splices), inner: copyInts(s.Inner)}
}
