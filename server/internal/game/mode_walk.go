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
	for i, occ := range occs {
		if g.resolutionOpen && g.pausedResolutionChoiceLocked() != nil {
			g.pausedModeWalk = &modeWalk{item: item, key: key, rest: append([]int(nil), occs[i:]...)}
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
		if err := fn(g, item, occ); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    item.Controller,
				Source:   item.SourceCardID,
				ErrorMsg: err.Error(),
			})
		}
	}
}

// resumeModeWalkLocked runs the parked bullets, if any, inside a
// resolution function so the boundary stays held while they run.
// Reports whether there was a walk to resume.
//
// Caller must hold g.mu in write mode.
func (g *Game) resumeModeWalkLocked() bool {
	w := g.pausedModeWalk
	if w == nil {
		return false
	}
	g.pausedModeWalk = nil
	if g.State != StateActive {
		return true
	}
	g.beginResolutionLocked()
	defer g.endResolutionLocked()
	if ms := w.spec(); ms != nil {
		g.runModeOccurrencesLocked(w.item, ms, w.key, w.rest)
	}
	return true
}

// cloneModeWalk copies the parked walk for Clone.
func cloneModeWalk(w *modeWalk) *modeWalk {
	if w == nil {
		return nil
	}
	return &modeWalk{item: cloneStackItem(w.item), key: w.key, rest: append([]int(nil), w.rest...)}
}

// modeWalkSnapshot is Game.pausedModeWalk on disk.
type modeWalkSnapshot struct {
	Item stackItemSnapshot `json:"item"`
	Key  string            `json:"key,omitempty"`
	Rest []int             `json:"rest"`
}

func snapshotModeWalk(g *Game, w *modeWalk, cen *ContinuationCensus) *modeWalkSnapshot {
	if w == nil || w.item == nil {
		return nil
	}
	return &modeWalkSnapshot{Item: snapshotStackItem(g, w.item, cen), Key: w.key, Rest: copyInts(w.rest)}
}

func restoreModeWalk(s *modeWalkSnapshot) *modeWalk {
	if s == nil {
		return nil
	}
	item, _ := restoreStackItem(&s.Item)
	return &modeWalk{item: item, key: s.Key, rest: copyInts(s.Rest)}
}
