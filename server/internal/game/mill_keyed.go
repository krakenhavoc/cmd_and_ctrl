package game

import (
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
)

// mill_keyed.go — #4253 slice 9: a mill whose continuation is data.
//
// A mill with a continuation (MillToZoneThenForEffect) routes its planned
// cards through routeAllThenLocked, so a leg paused on a CR 616 prompt
// (an ordering prompt between Rest in Peace and Leyline of the Void over
// one milled card) held the rest of the mill and the caller's "then" in a
// closure. The keyed doors put the batch on the keyed route batch
// (route_batch_keyed.go), told the caller's tail what landed.
//
// A "mills a card, then repeats this process until …" run (millRun) held
// two closures: the caller's "then" and the stop clause. Its keyed twin,
// millUntilRun, is plain data riding the batch of the repetition in
// flight (routeBatchRun.Mill): the clause is a registered predicate
// (RegisterMillUntil) and the number it reads, and the run's view of the
// cards it has landed is their pre-move copies, written through the card
// mirror. Every other field is a scalar or an ID list.
//
// What stays a closure: the mill-AMOUNT window (RepEventMill, Bruvac the
// Grandiloquent's doubling) carries its tail as a gate field, so an
// ordering prompt between two mill-amount replacements is not a restore
// point yet. The per-card legs are.

// MillUntilFunc is a registered stop clause: asked, after each repetition,
// with the cards that have LANDED in the destination so far (pre-move
// copies, in arrival order) and the number the clause was given. It must
// be pure: an undo replays it over the same list.
type MillUntilFunc func(landed []Card, n int) bool

// MillUntilRef names a registered stop clause. The zero ref is "no clause".
type MillUntilRef struct{ key string }

// Key is the registered key.
func (r MillUntilRef) Key() string { return r.key }

// MillUntil is a stop clause with the number it reads (a count bound, a
// mana-value threshold).
type MillUntil struct {
	Ref MillUntilRef
	N   int
}

// IsZero reports whether no clause is set.
func (u MillUntil) IsZero() bool { return u.Ref.key == "" }

// Pred is the clause as the closure form's predicate, for a caller that
// still runs a mill through MillToZoneThenForEffect. Nil for the zero
// clause.
func (u MillUntil) Pred() func([]Card) bool {
	if u.IsZero() {
		return nil
	}
	fn, n := lookupMillUntil(u.Ref.key), u.N
	if fn == nil {
		return nil
	}
	return func(landed []Card) bool { return fn(landed, n) }
}

var (
	millUntilMu  sync.RWMutex
	millUntilFns = map[string]MillUntilFunc{}
)

// RegisterMillUntil registers a stop clause under key, an on-disk
// identity: never renamed, never reused, ledgered "milluntil <key>". Call
// it once, from a package-level var.
func RegisterMillUntil(key string, fn MillUntilFunc) MillUntilRef {
	if key == "" || fn == nil {
		panic("game.RegisterMillUntil: a key and a function are required")
	}
	millUntilMu.Lock()
	defer millUntilMu.Unlock()
	if _, dup := millUntilFns[key]; dup {
		panic(fmt.Sprintf("game.RegisterMillUntil: %q registered twice", key))
	}
	millUntilFns[key] = fn
	return MillUntilRef{key: key}
}

// KnownMillUntil reports whether a clause is registered under key.
func KnownMillUntil(key string) bool {
	millUntilMu.RLock()
	defer millUntilMu.RUnlock()
	_, ok := millUntilFns[key]
	return ok
}

func registeredMillUntilKeys() []string {
	millUntilMu.RLock()
	defer millUntilMu.RUnlock()
	out := make([]string, 0, len(millUntilFns))
	for k := range millUntilFns {
		out = append(out, "milluntil "+k)
	}
	sort.Strings(out)
	return out
}

func lookupMillUntil(key string) MillUntilFunc {
	millUntilMu.RLock()
	defer millUntilMu.RUnlock()
	return millUntilFns[key]
}

// millUntilRun is millRun as data. Never written after it is built: each
// repetition's continuation builds a new one.
type millUntilRun struct {
	Player      uuid.UUID `json:"player"`
	Dest        ZoneKind  `json:"dest"`
	Until       string    `json:"until,omitempty"`
	UntilN      int       `json:"untilN,omitempty"`
	Repetitions int       `json:"repetitions,omitempty"`
	Done        int       `json:"done,omitempty"`
	// Depth is the library's depth before the repetition in flight
	// (millRun.stalled).
	Depth  int         `json:"depth,omitempty"`
	Landed []uuid.UUID `json:"landed,omitempty"`
	// Seen is the pre-move copy of each landed card, the run's view of
	// what it has milled (millRun.before), so the clause reads the cards
	// as they were in the library.
	Seen []cardSnapshot `json:"seen,omitempty"`
	// Planned is the repetition in flight's planned cards as they were in
	// the library, from which its landed cards' copies are taken.
	Planned []cardSnapshot     `json:"planned,omitempty"`
	Then    *keyedTailSnapshot `json:"then,omitempty"`
}

func cloneMillUntilRun(r *millUntilRun) *millUntilRun {
	if r == nil {
		return nil
	}
	out := *r
	out.Landed = copyUUIDs(r.Landed)
	out.Seen = append([]cardSnapshot(nil), r.Seen...)
	out.Planned = append([]cardSnapshot(nil), r.Planned...)
	if r.Then != nil {
		out.Then = snapshotKeyedTail(restoreKeyedTail(r.Then))
	}
	return &out
}

func unknownMillUntilKeys(r *millUntilRun) []string {
	if r == nil {
		return nil
	}
	out := unknownKeyedTail(r.Then)
	if r.Until != "" && !KnownMillUntil(r.Until) {
		out = append(out, "milluntil "+r.Until)
	}
	return out
}

// stops asks the clause about everything that has landed.
func (r *millUntilRun) stops() bool {
	if r.Until == "" || len(r.Landed) == 0 {
		return false
	}
	fn := lookupMillUntil(r.Until)
	if fn == nil {
		return false
	}
	cards := make([]Card, 0, len(r.Seen))
	for i := range r.Seen {
		cards = append(cards, restoreCard(&r.Seen[i]))
	}
	return len(cards) > 0 && fn(cards, r.UntilN)
}

// millUntilRunKeyedLocked runs one repetition of a keyed run, with the
// run riding that repetition's mill: millUntilRunLocked's keyed twin.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) millUntilRunKeyedLocked(run *millUntilRun) error {
	p := g.playerByIDLocked(run.Player)
	if p == nil {
		return ErrPlayerNotFound
	}
	if (run.Repetitions > 0 && run.Done >= run.Repetitions) || len(p.Library.Cards) == 0 {
		return g.runKeyedTailLocked(restoreKeyedTail(run.Then), run.Landed, 0)
	}
	rep := cloneMillUntilRun(run)
	rep.Depth = len(p.Library.Cards)
	_, err := g.millThroughReplacementsTailLocked(run.Player, 1, run.Dest, nil, &millTail{dest: run.Dest, run: rep})
	return err
}

// continueMillUntilRunKeyedLocked is a repetition's end: count what it
// milled, ask the clause, and finish the run or start the next one.
//
// Caller must hold g.mu in write mode.
func (g *Game) continueMillUntilRunKeyedLocked(run *millUntilRun, milled []uuid.UUID) error {
	next := cloneMillUntilRun(run)
	next.Done++
	next.Planned = nil
	for _, id := range milled {
		next.Landed = append(next.Landed, id)
		for _, c := range run.Planned {
			if c.InstanceID == id {
				next.Seen = append(next.Seen, c)
				break
			}
		}
	}
	stalled := true
	if p := g.playerByIDLocked(run.Player); p != nil {
		stalled = len(p.Library.Cards) >= run.Depth
	}
	if next.stops() || stalled {
		return g.runKeyedTailLocked(restoreKeyedTail(next.Then), next.Landed, 0)
	}
	return g.millUntilRunKeyedLocked(next)
}

// MillToZoneKeyedForEffect is MillToZoneThenForEffect with the
// continuation as data (#4253 slice 9): `then` is a registered tail told
// the cards that landed in `dest`, and `until`, when set, is a registered
// stop clause that makes this a run (millUntilRunLocked's rules). A zero
// tail is "no continuation"; the mill still sequences, so a leg's prompt
// is a restore point.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) MillToZoneKeyedForEffect(playerID uuid.UUID, n int, dest ZoneKind, until MillUntil, then KeyedTail) error {
	if g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	switch dest {
	case ZoneGraveyard, ZoneExile:
	default:
		return ErrInvalidParam
	}
	if !until.IsZero() {
		return g.millUntilRunKeyedLocked(&millUntilRun{
			Player: playerID, Dest: dest, Until: until.Ref.key, UntilN: until.N,
			Repetitions: n, Then: snapshotKeyedTail(then),
		})
	}
	t := then
	_, err := g.millThroughReplacementsTailLocked(playerID, n, dest, nil, &millTail{dest: dest, keyed: &t})
	return err
}

// MillToZoneKeyedMovedForEffect is MillToZoneThenMovedForEffect with the
// continuation as data: `then` is told every card the mill took out of the
// library, wherever it went (CR 701.17c).
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) MillToZoneKeyedMovedForEffect(playerID uuid.UUID, n int, dest ZoneKind, then KeyedTail) error {
	t := then
	_, err := g.millThroughReplacementsTailLocked(playerID, n, dest, nil, &millTail{dest: dest, keyed: &t, allMoved: true})
	return err
}
