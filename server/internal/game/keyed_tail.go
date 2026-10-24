package game

import (
	"sort"
	"sync"

	"github.com/google/uuid"
)

// keyed_tail.go — #4253 slice 3: a paused event's continuation written as
// data, so a CR 616 prompt whose event carries one is a restore point.
//
// A paused CR 614 window resumes through its event's tail: an entry's
// entryTail, a simultaneous entry's batch, a token creation's tokenTail,
// a counter placement's counterTail, a life change's lifeTail. Each of
// those held the rest of the effect as a Go closure, which a restore
// point cannot carry, so a frame whose event had one stayed off the disk
// (replacementFrameKeyable) and the table was not a restore point.
//
// A tail now has a second, closure-free form beside the closure: a
// KeyedTail, a registered body named by a key plus the plain data it
// reads. The engine's own sequencers (a library search's remaining
// takes and its shuffle, a token batch's remaining tokens) are data of
// their own on the tail (searchEntryRun, tokenEntryRun). A caller whose
// continuation is a card closure still passes the closure, and that
// frame is still counted by the census exactly as before.

// TailRef names a registered TailFunc. The key is unexported, so a func
// never reaches an event and a ref is only made by RegisterTail.
type TailRef struct{ key string }

// Key is the registered key, or "" for the zero ref.
func (r TailRef) Key() string { return r.key }

// KeyedTail is a registered continuation and the scalars it reads:
// plain data, copied by value into a resume frame and written with it.
// The zero value is "no continuation".
type KeyedTail struct {
	Ref TailRef
	// Player and Source are what a card-side body most often reads: the
	// effect's controller and its source card.
	Player uuid.UUID
	Source uuid.UUID
	// Carry is anything else the body reads, as IDs: a target, the
	// objects read as the effect resolved (CR 608.2b).
	Carry []uuid.UUID
	// N is a number the body reads.
	N int
	// Asked is the asking stack item as data (NewAskedItem, #4253 slice 4),
	// for a body that reads more of the item than its controller and
	// source: its targets, its X, its text, the window its source object
	// had. A body runs against a copy rebuilt from it (AskedItem.StackItem),
	// the way a keyed pay-unless answer does. Nil when the card put none
	// there.
	Asked *AskedItem
}

// IsZero reports whether no continuation is recorded.
func (t KeyedTail) IsZero() bool { return t.Ref.key == "" }

// TailOutcome is what a keyed continuation is told when the event it
// follows has settled: the tail it was recorded with, and the outcome.
// IDs is the permanents that entered or the tokens that were created
// (in order), Amount the counters placed or the life that moved.
type TailOutcome struct {
	Tail   KeyedTail
	IDs    []uuid.UUID
	Amount int
}

// TailFunc is the body of a keyed continuation. It runs with g.mu held,
// takes the live *Game, and may start the next event or queue the next
// prompt itself.
type TailFunc func(g *Game, o TailOutcome) error

var (
	tailMu     sync.RWMutex
	tailBodies = map[string]TailFunc{}
)

// RegisterTail registers a continuation body under key, an on-disk
// identity: never renamed, never reused, ledgered "tail <key>". Call it
// once, from a package-level var or an init. The body must capture
// nothing that lives in a game.
func RegisterTail(key string, body TailFunc) TailRef {
	if key == "" || body == nil {
		panic("game.RegisterTail: a key and a body are required")
	}
	tailMu.Lock()
	tailBodies[key] = body
	tailMu.Unlock()
	return TailRef{key: key}
}

// KnownTail reports whether a body is registered under key.
func KnownTail(key string) bool {
	tailMu.RLock()
	defer tailMu.RUnlock()
	_, ok := tailBodies[key]
	return ok
}

func registeredTailKeys() []string {
	tailMu.RLock()
	defer tailMu.RUnlock()
	out := make([]string, 0, len(tailBodies))
	for k := range tailBodies {
		out = append(out, "tail "+k)
	}
	sort.Strings(out)
	return out
}

// runKeyedTailLocked runs a recorded continuation, if any. A key the
// running binary does not register runs nothing (a restore refuses such
// a file first, checkEffectKeys).
//
// Caller must hold g.mu.
func (g *Game) runKeyedTailLocked(t KeyedTail, ids []uuid.UUID, amount int) error {
	if t.IsZero() {
		return nil
	}
	tailMu.RLock()
	body := tailBodies[t.Ref.key]
	tailMu.RUnlock()
	if body == nil {
		return nil
	}
	return body(g, TailOutcome{Tail: copyKeyedTail(t), IDs: copyUUIDs(ids), Amount: amount})
}

// Run runs the tail at once, told ids: for a caller whose event never
// happened (nothing to exile, nothing to place) but whose continuation is
// still owed.
//
// Caller must hold g.mu.
func (t KeyedTail) Run(g *Game, ids []uuid.UUID) error {
	return g.runKeyedTailLocked(t, ids, 0)
}

// isScryOrSurveil reports whether a prompt kind can carry a keyed look-at
// continuation (PendingChoice.scryThen).
func isScryOrSurveil(k PendingChoiceKind) bool {
	return k == PendingChoiceScry || k == PendingChoiceSurveil
}

// copyKeyedTail gives a tail its own Carry.
func copyKeyedTail(t KeyedTail) KeyedTail {
	t.Carry = copyUUIDs(t.Carry)
	t.Asked = copyAskedItem(t.Asked)
	return t
}

// keyedTailSnapshot is a KeyedTail on disk.
type keyedTailSnapshot struct {
	Key    string      `json:"key"`
	Player *uuid.UUID  `json:"player,omitempty"`
	Source *uuid.UUID  `json:"source,omitempty"`
	Carry  []uuid.UUID `json:"carry,omitempty"`
	N      int         `json:"n,omitempty"`
	Asked  *AskedItem  `json:"asked,omitempty"`
}

func snapshotKeyedTail(t KeyedTail) *keyedTailSnapshot {
	if t.IsZero() {
		return nil
	}
	return &keyedTailSnapshot{
		Key: t.Ref.key, Player: uuidPtrOrNil(t.Player), Source: uuidPtrOrNil(t.Source),
		Carry: copyUUIDs(t.Carry), N: t.N, Asked: copyAskedItem(t.Asked),
	}
}

func restoreKeyedTail(s *keyedTailSnapshot) KeyedTail {
	if s == nil {
		return KeyedTail{}
	}
	return KeyedTail{
		Ref: TailRef{key: s.Key}, Player: uuidOrNil(s.Player), Source: uuidOrNil(s.Source),
		Carry: copyUUIDs(s.Carry), N: s.N, Asked: copyAskedItem(s.Asked),
	}
}

// unknownKeyedTail is checkEffectKeys' line for a tail the running
// binary does not register.
func unknownKeyedTail(s *keyedTailSnapshot) []string {
	if s == nil || KnownTail(s.Key) {
		return nil
	}
	return []string{"tail " + s.Key}
}
