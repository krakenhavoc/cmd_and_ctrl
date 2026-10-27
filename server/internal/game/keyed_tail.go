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
	// Word is a word the body reads (#4253 slice 6): a creature type, a
	// name. Empty when the body reads none.
	Word string
	// Next is the tail to run after this one, as data (#4253 slice 6): an
	// engine tail that finishes a keyword action (amass) and then hands
	// the card's own tail what it produced. Nil when there is none.
	Next *KeyedTail
	// Permission is a cast permission the body grants, as data (#4253
	// slice 10a): an exile's "you may cast it" (airbend, impulse exile,
	// the free-cast grants, the Oblivion Ring family's return record).
	// Nil when the body grants none.
	Permission *CastPermission
	// Nums and Words are more numbers and words the body reads (#4253
	// slice 10b): a run's position (proliferate's permanent, counter kind
	// and remaining asks) and the names it walks (the counter kinds).
	Nums  []int
	Words []string
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
	// Rolls is a dice roll's kept results, in the order rolled (#4307,
	// dice_roll.go): the rolls an ignored one was dropped from.
	Rolls []int
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

// TailRefByKey is the ref for a registered key, for a continuation that
// carries its tail as a word on a prompt (#4253 slice 10c). False when this
// binary registers no such body.
func TailRefByKey(key string) (TailRef, bool) {
	if !KnownTail(key) {
		return TailRef{}, false
	}
	return TailRef{key: key}, true
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

// runKeyedRollTailLocked runs a dice roll's recorded continuation, told
// the kept rolls (#4307).
//
// Caller must hold g.mu.
func (g *Game) runKeyedRollTailLocked(t KeyedTail, rolls []int) error {
	if t.IsZero() {
		return nil
	}
	tailMu.RLock()
	body := tailBodies[t.Ref.key]
	tailMu.RUnlock()
	if body == nil {
		return nil
	}
	return body(g, TailOutcome{Tail: copyKeyedTail(t), Amount: len(rolls), Rolls: append([]int(nil), rolls...)})
}

// Run runs the tail at once, told ids: for a caller whose event never
// happened (nothing to exile, nothing to place) but whose continuation is
// still owed.
//
// Caller must hold g.mu.
func (t KeyedTail) Run(g *Game, ids []uuid.UUID) error {
	return g.runKeyedTailLocked(t, ids, 0)
}

// RunTold is Run with the amount the event settled with (the counters placed,
// the life that moved), for a tail that hands what it was told on to the next
// one (#4253 slice 7).
//
// Caller must hold g.mu.
func (t KeyedTail) RunTold(g *Game, ids []uuid.UUID, amount int) error {
	return g.runKeyedTailLocked(t, ids, amount)
}

// carriesLookAtThen reports whether a prompt kind can carry a keyed look-at
// continuation (PendingChoice.scryThen): a scry, a surveil, and (#4301) a
// look at the top of the library to put the cards back in any order.
func carriesLookAtThen(k PendingChoiceKind) bool {
	return k == PendingChoiceScry || k == PendingChoiceSurveil || k == PendingChoiceLookAtTop
}

// copyKeyedTail gives a tail its own Carry.
func copyKeyedTail(t KeyedTail) KeyedTail {
	t.Carry = copyUUIDs(t.Carry)
	t.Asked = copyAskedItem(t.Asked)
	t.Nums = copyInts(t.Nums)
	t.Words = copyStrings(t.Words)
	if t.Next != nil {
		n := copyKeyedTail(*t.Next)
		t.Next = &n
	}
	t.Permission = copyCastPermissionPtr(t.Permission)
	return t
}

// copyCastPermissionPtr deep-copies a tail's permission.
func copyCastPermissionPtr(p *CastPermission) *CastPermission {
	if p == nil {
		return nil
	}
	out := cloneCastPermissions([]CastPermission{*p})[0]
	return &out
}

// keyedTailSnapshot is a KeyedTail on disk.
type keyedTailSnapshot struct {
	Key    string      `json:"key"`
	Player *uuid.UUID  `json:"player,omitempty"`
	Source *uuid.UUID  `json:"source,omitempty"`
	Carry  []uuid.UUID `json:"carry,omitempty"`
	N      int         `json:"n,omitempty"`
	Asked  *AskedItem  `json:"asked,omitempty"`
	// Word and Next (#4253 slice 6).
	Word string             `json:"word,omitempty"`
	Next *keyedTailSnapshot `json:"next,omitempty"`
	// Permission (#4253 slice 10a): a cast permission the body grants.
	Permission *CastPermission `json:"permission,omitempty"`
	// Nums and Words (#4253 slice 10b).
	Nums  []int    `json:"nums,omitempty"`
	Words []string `json:"words,omitempty"`
}

func snapshotKeyedTail(t KeyedTail) *keyedTailSnapshot {
	if t.IsZero() {
		return nil
	}
	return &keyedTailSnapshot{
		Key: t.Ref.key, Player: uuidPtrOrNil(t.Player), Source: uuidPtrOrNil(t.Source),
		Carry: copyUUIDs(t.Carry), N: t.N, Asked: copyAskedItem(t.Asked),
		Word: t.Word, Next: snapshotKeyedTail(nextTailValue(t.Next)),
		Permission: copyCastPermissionPtr(t.Permission),
		Nums:       copyInts(t.Nums), Words: copyStrings(t.Words),
	}
}

// nextTailValue is the tail a Next points at, or the zero tail.
func nextTailValue(n *KeyedTail) KeyedTail {
	if n == nil {
		return KeyedTail{}
	}
	return *n
}

func restoreKeyedTail(s *keyedTailSnapshot) KeyedTail {
	if s == nil {
		return KeyedTail{}
	}
	out := KeyedTail{
		Ref: TailRef{key: s.Key}, Player: uuidOrNil(s.Player), Source: uuidOrNil(s.Source),
		Carry: copyUUIDs(s.Carry), N: s.N, Asked: copyAskedItem(s.Asked), Word: s.Word,
		Permission: copyCastPermissionPtr(s.Permission),
		Nums:       copyInts(s.Nums), Words: copyStrings(s.Words),
	}
	if s.Next != nil {
		n := restoreKeyedTail(s.Next)
		out.Next = &n
	}
	return out
}

// unknownKeyedTail is checkEffectKeys' line for a tail the running
// binary does not register.
func unknownKeyedTail(s *keyedTailSnapshot) []string {
	if s == nil {
		return nil
	}
	var out []string
	if !KnownTail(s.Key) {
		out = append(out, "tail "+s.Key)
	}
	return append(out, unknownKeyedTail(s.Next)...)
}
