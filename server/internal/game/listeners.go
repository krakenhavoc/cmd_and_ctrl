package game

// listeners.go holds the per-game event listener registry. Pull-
// based dispatch: the engine walks Game.Listeners synchronously
// after every EmitEvent, and each listener decides whether the
// event is interesting to it. Contrast with push-based dispatch,
// where each mutation would have to know about every card that
// cares — that coupling is what makes the XMage-style per-card
// implementation so heavy.
//
// S14 ships the registry infrastructure with ZERO production
// listeners. The interface + call site + Register API are stable
// so S19 can register the first real listener — the triggered-
// ability harvester that watches events and enqueues matching
// triggers onto PendingTriggers — without plumbing work.
//
// Listeners run under the game write lock (same lock held by the
// mutation that emitted the event). A listener that needs to
// mutate must call the *Locked helpers; calling public locking
// mutators will deadlock. This constraint is documented per-call
// on Listener.OnEvent and enforced by code review — Go can't
// encode "locked context" at the type level.

// Listener is the interface every event consumer implements. The
// engine invokes OnEvent synchronously after each EmitEvent, in
// registration order. Listeners are process-lifetime singletons —
// they carry no per-game state (game state lives on g), so Clone /
// RestoreFrom shallow-copies the slice.
//
// OnEvent runs under g.mu held in write mode. A listener that
// needs to mutate the game calls *Locked helpers. One that only
// reads can use the public accessors (they'll re-acquire the read
// lock path and will NOT deadlock against the already-held write
// lock because Go's sync.RWMutex permits write-lock holders to
// call read-lock paths via the *Locked helpers — but the cleanest
// pattern is to read Game fields directly within OnEvent).
type Listener interface {
	OnEvent(g *Game, ev Event)
}

// RegisterListener appends l to the game's listener list. Safe to
// call at any time; listeners registered mid-game see only events
// emitted after their registration. Takes the write lock briefly
// to protect the slice append — listeners are typically registered
// at Start() time, not in hot paths.
func (g *Game) RegisterListener(l Listener) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Listeners = append(g.Listeners, l)
}

// notifyListenersLocked walks the listener slice and invokes each
// listener's OnEvent. Caller MUST hold g.mu (the EmitEvent path
// already does). Errors in listeners are NOT swallowed — a panic
// propagates back to the caller and is expected to crash the room
// goroutine, not the whole server. S19 listeners need to be
// defensive about their own correctness.
func (g *Game) notifyListenersLocked(ev Event) {
	for _, l := range g.Listeners {
		l.OnEvent(g, ev)
	}
}
