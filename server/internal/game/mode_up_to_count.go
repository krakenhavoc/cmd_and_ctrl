package game

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// mode_up_to_count.go — "choose up to X —" where X is a count of the
// board (Bumi, King of Three Trials: "When Bumi enters, choose up to X,
// where X is the number of Lesson cards in your graveyard —").
//
// The sibling of ModeSpec.UpToCastX (#4617). That one reads the X paid
// for the spell that became the source (CR 107.3m); this one reads a
// number off the game as the trigger is put on the stack, which is when
// a triggered ability's modes are chosen (CR 603.3c). The bound is that
// number capped at the printed bullets, because no bullet may be chosen
// twice (CR 700.2d), and at 0 no bullet can be chosen and the trigger is
// removed (CR 700.2b) — the same two rules UpToCastX follows, through the
// same lines of queueModePickLocked.
//
// The count is a registered KEY, not a func: a ModeSpec is reachable
// from Game through a trigger's paused mode_pick frame, and the closure
// ratchet (testdata/closure_fields.txt) allows no new func-typed route.
// It is registered once at init by RegisterModeCount, the same way
// ModeCondition registers a conditional mode count.

// ModeUpToCount names a registered count for ModeSpec.UpToCount. The zero
// value names none.
type ModeUpToCount struct{ key string }

// Key is the count's registry key.
func (c ModeUpToCount) Key() string { return c.key }

// IsZero reports whether no count is named.
func (c ModeUpToCount) IsZero() bool { return c.key == "" }

// modeCountFunc is a registered count. ev is the event that fired the
// trigger whose modes are being chosen (TriggerContext.Event); a count of
// the board ignores it.
type modeCountFunc func(g *Game, chooser uuid.UUID, ev Event) int

var modeCounts = struct {
	sync.RWMutex
	byKey map[string]modeCountFunc
}{byKey: map[string]modeCountFunc{}}

// RegisterModeCount registers the count "choose up to X" reads under
// `key` and returns its name. Call it once, from a package-level var in
// the file that uses it. Panics on an empty key, a nil count or a
// duplicate, each a card-file bug.
//
// The count is read-only and runs under g.mu, so it must not call a
// public locking accessor; `chooser` is the player choosing the modes,
// the "you" of "Lesson cards in your graveyard".
func RegisterModeCount(key string, fn func(g *Game, chooser uuid.UUID) int) ModeUpToCount {
	if fn == nil {
		panic(fmt.Sprintf("game: mode count %q has no count", key))
	}
	return registerModeCount(key, func(g *Game, chooser uuid.UUID, _ Event) int { return fn(g, chooser) })
}

// RegisterModeCountOfEvent is RegisterModeCount for an X read off the
// event that fired the trigger (#4468): Riku of Many Paths' "Whenever
// you cast a modal spell, choose up to X, where X is the number of times
// you chose a mode for that spell". `ev` is the triggering event as the
// trigger is put on the stack (CR 603.3c), so "that spell" is ev.CardID.
// The same rules as RegisterModeCount: once, at init, read-only, under
// g.mu.
func RegisterModeCountOfEvent(key string, fn func(g *Game, chooser uuid.UUID, ev Event) int) ModeUpToCount {
	if fn == nil {
		panic(fmt.Sprintf("game: mode count %q has no count", key))
	}
	return registerModeCount(key, fn)
}

func registerModeCount(key string, fn modeCountFunc) ModeUpToCount {
	if key == "" {
		panic("game: RegisterModeCount with an empty key")
	}
	modeCounts.Lock()
	defer modeCounts.Unlock()
	if _, dup := modeCounts.byKey[key]; dup {
		panic(fmt.Sprintf("game: mode count %q registered twice", key))
	}
	modeCounts.byKey[key] = fn
	return ModeUpToCount{key: key}
}

// count evaluates the named count for `chooser`, for the trigger fired by
// ev. An unknown or zero key counts 0, which chooses no bullet: the
// weaker reading.
func (c ModeUpToCount) count(g *Game, chooser uuid.UUID, ev Event) int {
	if c.key == "" {
		return 0
	}
	modeCounts.RLock()
	fn, ok := modeCounts.byKey[c.key]
	modeCounts.RUnlock()
	if !ok {
		return 0
	}
	return fn(g, chooser, ev)
}

// upToX reports whether the spec's upper bound is an X read as the
// trigger is put on the stack: the cast X (UpToCastX) or a board count
// (UpToCount).
func (ms *ModeSpec) upToX() bool {
	return ms != nil && (ms.UpToCastX || !ms.UpToCount.IsZero())
}
