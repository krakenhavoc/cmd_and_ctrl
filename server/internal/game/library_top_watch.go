package game

import "github.com/google/uuid"

// library_top_watch.go — StaticAbility.DependsOnLibraryTop's invalidation
// (ADR 0155 §8, #3202): Conspicuous Snoop and Skill Borrower have the
// activated abilities of the top card of their controller's library, so
// the cached layer resolution goes stale the moment any library's top
// card changes.
//
// # Why a watch and not an event arm
//
// The other DependsOn* flags gate a bump in the layer listener on the
// events that move their input. A library's top moves on a draw, a mill,
// a cast from the top and a card put on top, which are zone moves, but
// also on a scry, a surveil, "put them back in any order" and a shuffle,
// which move nothing across a zone boundary and emit nothing that names
// the new top. An arm per spelling would miss the next one, which is how
// the hand-size flag was once wrong for a sprint (layer_listener.go).
//
// So the flag arms a WATCH. A layer pass that finds a static declaring
// it on the battlefield records each library's top card; the staleness
// check (RecomputeLayersIfStaleLocked and ReadSnapshot) compares that
// record with the libraries now and bumps the layer version when one
// differs. A table with no such static records nothing and compares
// nothing, which is the flags' opt-in contract.

// libraryTopMark is one library's top card when the last pass ran:
// uuid.Nil for an empty library.
type libraryTopMark struct {
	owner uuid.UUID
	top   uuid.UUID
}

// libraryTopStaticIsLiveLocked is handSizeStaticIsLiveLocked for the top
// of a library.
func libraryTopStaticIsLiveLocked(g *Game) bool {
	return staticOnBattlefieldLocked(g, func(ab StaticAbility) bool { return ab.DependsOnLibraryTop })
}

// libraryTopsLocked lists every seat's library top now.
func (g *Game) libraryTopsLocked() []libraryTopMark {
	out := make([]libraryTopMark, 0, len(g.Seats))
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		m := libraryTopMark{owner: p.ID}
		if p.Library != nil && len(p.Library.Cards) > 0 {
			m.top = p.Library.Cards[len(p.Library.Cards)-1].InstanceID
		}
		out = append(out, m)
	}
	return out
}

// armLibraryTopWatchLocked records the library tops for the pass that
// just ran, or clears the watch when nothing on the battlefield reads
// them. Called at the end of every recompute.
//
// Caller must hold g.mu in write mode.
func (g *Game) armLibraryTopWatchLocked() {
	if !libraryTopStaticIsLiveLocked(g) {
		g.libraryTopWatch = nil
		return
	}
	g.libraryTopWatch = g.libraryTopsLocked()
}

// libraryTopMovedLocked reports whether a watched library's top card has
// changed since the last pass, and bumps the layer version when it has,
// so every reader keyed on the version sees the change too. False, after
// one nil test, at a table with no watch.
//
// Safe under the read lock: it reads the libraries, and the bump is an
// atomic add.
func (g *Game) libraryTopMovedLocked() bool {
	if g.libraryTopWatch == nil {
		return false
	}
	now := g.libraryTopsLocked()
	moved := len(now) != len(g.libraryTopWatch)
	for i := 0; !moved && i < len(now); i++ {
		moved = now[i] != g.libraryTopWatch[i]
	}
	if moved && g.layerVersion.Load() == g.lastResolvedVersion.Load() {
		g.layerVersion.Add(1)
	}
	return moved
}
