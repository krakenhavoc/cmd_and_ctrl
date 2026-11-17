package draft

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// AwayGrace is how long into a pick a seat with no socket is waited
// for before the Picker picks for it. A person who is away is picked
// for the same way as one whose timer ran out (ADR 0222 §5), only
// sooner, so one closed laptop does not cost the others the whole
// timer on every pick.
const AwayGrace = 10 * time.Second

// Subscriber is one socket watching a draft as one seat.
type Subscriber interface {
	// Send delivers the seat's view. It must not block: it is called
	// with the room's lock held, which is what keeps views in order.
	Send(View)
	// Close ends the socket: the draft closed, or the seat left it.
	Close(reason string)
}

// ClosedReason is the close reason a socket gets when its draft closes
// or its seat leaves (docs/protocol.md).
const ClosedReason = "draft closed"

// Room is a Draft with a lock, the pick timer and the sockets watching
// it. Every method is safe for concurrent use.
//
// Lock order: the lobby may hold its own lock and call a Room; a Room
// never calls the lobby, and calls a Subscriber only through Send and
// Close.
type Room struct {
	mu  sync.Mutex
	d   *Draft
	src Source

	subs map[uuid.UUID]map[Subscriber]struct{}
	// leftAt is when a seat's last socket went away. A seat that never
	// connected has no entry and counts as away from the pick's start.
	leftAt map[uuid.UUID]time.Time

	clock func() time.Time
	// after schedules the timer; a test replaces it to drive time by
	// hand.
	after func(time.Duration, func()) (stop func())
	stop  func()
	// gen invalidates a timer that fires after it was replaced.
	gen    uint64
	closed bool

	// lastHuman is the last time a person did something here: took a
	// seat, picked, opened or closed a socket. The lobby's idle sweep
	// reads it.
	lastHuman time.Time
	// decorate lets the lobby add what the draft does not know to a
	// view (the invite, for the host).
	decorate func(seat uuid.UUID, v *View)
	// wake is signalled after every change, for the lobby's watcher: it
	// opens the tables the rounds need (Wake).
	wake chan struct{}
}

// NewRoom wraps a draft. src is what Start draws the packs from.
func NewRoom(d *Draft, src Source) *Room {
	r := &Room{
		d:      d,
		src:    src,
		subs:   map[uuid.UUID]map[Subscriber]struct{}{},
		leftAt: map[uuid.UUID]time.Time{},
		clock:  time.Now,
		wake:   make(chan struct{}, 1),
	}
	r.after = func(wait time.Duration, f func()) func() {
		t := time.AfterFunc(wait, f)
		return func() { t.Stop() }
	}
	r.lastHuman = r.clock()
	return r
}

// SetClock replaces the room's clock and timer, for tests. Call it
// before the room is shared.
func (r *Room) SetClock(now func() time.Time, after func(time.Duration, func()) (stop func())) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if now != nil {
		r.clock = now
		r.lastHuman = now()
	}
	if after != nil {
		r.after = after
	}
}

// SetDecorator sets the hook that adds the lobby's fields to a view.
// It is called with the room's lock held and must not call the room.
func (r *Room) SetDecorator(f func(seat uuid.UUID, v *View)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.decorate = f
}

// ID is the draft's id.
func (r *Room) ID() uuid.UUID { return r.d.ID }

// Status is what the lobby reads about a room for the load cap and the
// idle sweep.
type Status struct {
	State     State
	Humans    int
	Connected int
	// Away is true when nobody is on the draft's page by the pick
	// timer's measure: no seat has a socket, and none has had one for
	// AwayGrace. A socket that dropped and is redialling is not away
	// (#5698).
	Away bool
	Solo bool
	// Phase is where the draft is after its picks (limited.go).
	Phase Phase
	// LastHuman is the last time a person did something here: took a
	// seat, picked, or opened or closed a socket.
	LastHuman time.Time
	Closed    bool
}

// Status reads the room.
func (r *Room) Status() Status {
	r.mu.Lock()
	defer r.mu.Unlock()
	away := len(r.subs) == 0
	if away {
		now := r.clock()
		for _, left := range r.leftAt {
			if now.Before(left.Add(AwayGrace)) {
				away = false
			}
		}
	}
	return Status{
		State:     r.d.State(),
		Humans:    r.d.Humans(),
		Connected: len(r.subs),
		Away:      away,
		Solo:      r.d.Config().Solo,
		Phase:     r.d.Phase(),
		LastHuman: r.lastHuman,
		Closed:    r.closed,
	}
}

// With runs f on the draft under the room's lock, for a read the room
// has no method for. f must not keep the draft.
func (r *Room) With(f func(d *Draft)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f(r.d)
}

func (r *Room) connectedLocked(seat uuid.UUID) bool { return len(r.subs[seat]) > 0 }

func (r *Room) viewLocked(seat uuid.UUID) (View, bool) {
	v, ok := r.d.ViewFor(seat, r.clock(), r.connectedLocked)
	if ok && r.decorate != nil {
		r.decorate(seat, &v)
	}
	return v, ok
}

// View is one seat's view now.
func (r *Room) View(seat uuid.UUID) (View, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.viewLocked(seat)
}

// Preview is the invite page's summary.
func (r *Room) Preview() Preview {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.d.Preview()
}

// changedLocked is called after every change: it sends each watching
// seat its new view and sets the timer for whatever comes next.
func (r *Room) changedLocked() {
	for seat, set := range r.subs {
		v, ok := r.viewLocked(seat)
		for s := range set {
			if !ok {
				// The seat is gone (it left the lobby).
				s.Close(ClosedReason)
				continue
			}
			s.Send(v)
		}
		if !ok {
			delete(r.subs, seat)
		}
	}
	r.armLocked()
	select {
	case r.wake <- struct{}{}:
	default:
	}
}

// Wake is signalled after every change to the draft. It is how the
// lobby learns that a round has begun and needs its tables, without
// the room ever calling the lobby. One receiver.
func (r *Room) Wake() <-chan struct{} { return r.wake }

// dueLocked is when the Picker should pick for a seat that has not
// picked: at the deadline, or AwayGrace into the pick (or after its
// socket went, whichever is later) for a seat that is away.
func (r *Room) dueLocked(s *Seat) time.Time {
	due := r.d.Deadline()
	if r.connectedLocked(s.ID) {
		return due
	}
	from := r.d.StepAt()
	if left, ok := r.leftAt[s.ID]; ok && left.After(from) {
		from = left
	}
	if away := from.Add(AwayGrace); away.Before(due) {
		return away
	}
	return due
}

// buildDueLocked is when an untimed build stops waiting for a person
// who has not submitted: UntimedBuildAwayGrace after the build began,
// or after their last socket went if that was later. away is false for
// a person with a socket open, who is never built for.
func (r *Room) buildDueLocked(s *Seat) (due time.Time, away bool) {
	if r.connectedLocked(s.ID) {
		return time.Time{}, false
	}
	from := r.d.BuildAt()
	if left, ok := r.leftAt[s.ID]; ok && left.After(from) {
		from = left
	}
	return from.Add(UntimedBuildAwayGrace), true
}

// armLocked sets the timer for the earliest pending pick.
func (r *Room) armLocked() {
	if r.stop != nil {
		r.stop()
		r.stop = nil
	}
	r.gen++
	if r.closed {
		return
	}
	if deadline := r.d.BuildDeadline(); !deadline.IsZero() {
		// The build timer (limited.go).
		gen := r.gen
		r.stop = r.after(max(deadline.Sub(r.clock()), 0), func() { r.fire(gen) })
		return
	}
	if r.d.Phase() == PhaseBuilding && r.d.BuildUntimed() {
		// An untimed build has no deadline, only the grace of a person
		// who is away (buildDueLocked).
		next, ok := time.Time{}, false
		for _, s := range r.d.Unbuilt() {
			if due, away := r.buildDueLocked(s); away && (!ok || due.Before(next)) {
				next, ok = due, true
			}
		}
		if ok {
			gen := r.gen
			r.stop = r.after(max(next.Sub(r.clock()), 0), func() { r.fire(gen) })
		}
		return
	}
	pending := r.d.Pending()
	if len(pending) == 0 {
		return
	}
	next := r.dueLocked(pending[0])
	for _, s := range pending[1:] {
		if due := r.dueLocked(s); due.Before(next) {
			next = due
		}
	}
	gen := r.gen
	wait := max(next.Sub(r.clock()), 0)
	r.stop = r.after(wait, func() { r.fire(gen) })
}

func (r *Room) fire(gen uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if gen != r.gen {
		return
	}
	r.tickLocked()
}

// Tick picks for every seat whose time is up. The timer calls it; a
// test with a hand-driven clock calls it directly.
func (r *Room) Tick() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tickLocked()
}

func (r *Room) tickLocked() {
	if r.closed {
		return
	}
	now := r.clock()
	r.d.ExpireBuild(now)
	if r.d.Phase() == PhaseBuilding && r.d.BuildUntimed() {
		// BuildFor can end the build, so the seats are read first.
		for _, s := range r.d.Unbuilt() {
			if due, away := r.buildDueLocked(s); away && !due.After(now) {
				r.d.BuildFor(s.ID)
			}
		}
	}
	// Each AutoPick can pass the packs and start a new pick, so the
	// pending list is read again after every one. A seat is picked for
	// at most once per pick, and a pick that begins at now is not yet
	// due, so this ends.
	for again := true; again; {
		again = false
		for _, s := range r.d.Pending() {
			if !r.dueLocked(s).After(now) {
				r.d.AutoPick(s.ID, now)
				again = true
				break
			}
		}
	}
	r.changedLocked()
}

// Join seats a person in the lobby.
func (r *Room) Join(name string, userID uuid.UUID) (uuid.UUID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return uuid.Nil, ErrNoSuchSeat
	}
	s, err := r.d.AddSeat(name, userID)
	if err != nil {
		return uuid.Nil, err
	}
	r.lastHuman = r.clock()
	r.changedLocked()
	return s.ID, nil
}

// SeatOfUser is the seat a signed-in person already holds, or uuid.Nil.
func (r *Room) SeatOfUser(userID uuid.UUID) uuid.UUID {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.d.SeatOfUser(userID)
}

// HasSeat reports whether a person holds this seat.
func (r *Room) HasSeat(seat uuid.UUID) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.d.Seats() {
		if s.ID == seat && !s.Bot {
			return true
		}
	}
	return false
}

// Start starts the draft for the host.
func (r *Room) Start(by uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrNoSuchSeat
	}
	if err := r.d.Start(by, r.src, r.clock()); err != nil {
		return err
	}
	r.lastHuman = r.clock()
	r.changedLocked()
	return nil
}

// Pick makes a seat's pick.
func (r *Room) Pick(seat uuid.UUID, ids []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrNotPicking
	}
	if err := r.d.Pick(seat, ids, r.clock()); err != nil {
		return err
	}
	r.lastHuman = r.clock()
	r.changedLocked()
	return nil
}

// Leave takes a person out of their seat and closes their sockets. It
// reports how many people are still seated.
func (r *Room) Leave(seat uuid.UUID) (humans int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.d.Leave(seat, r.clock()); err != nil {
		return r.d.Humans(), err
	}
	for s := range r.subs[seat] {
		s.Close(ClosedReason)
	}
	delete(r.subs, seat)
	delete(r.leftAt, seat)
	r.changedLocked()
	return r.d.Humans(), nil
}

// Attach connects a socket to a seat. The seat's view is sent through
// it before Attach returns, so it is the socket's first frame. detach
// must be called when the socket goes.
func (r *Room) Attach(seat uuid.UUID, sub Subscriber) (detach func(), err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil, ErrNoSuchSeat
	}
	i, s := r.d.seat(seat)
	if i < 0 || s.Bot {
		return nil, ErrNoSuchSeat
	}
	set := r.subs[seat]
	first := len(set) == 0
	if set == nil {
		set = map[Subscriber]struct{}{}
		r.subs[seat] = set
	}
	set[sub] = struct{}{}
	delete(r.leftAt, seat)
	r.lastHuman = r.clock()
	if first {
		// Everyone sees the seat come back, and its timer is the full
		// one again.
		r.changedLocked()
	} else if v, ok := r.viewLocked(seat); ok {
		sub.Send(v)
	}
	var once sync.Once
	return func() { once.Do(func() { r.detach(seat, sub) }) }, nil
}

func (r *Room) detach(seat uuid.UUID, sub Subscriber) {
	r.mu.Lock()
	defer r.mu.Unlock()
	set, ok := r.subs[seat]
	if !ok {
		return
	}
	if _, ok := set[sub]; !ok {
		return
	}
	delete(set, sub)
	if len(set) > 0 {
		return
	}
	delete(r.subs, seat)
	if r.closed {
		return
	}
	r.leftAt[seat] = r.clock()
	r.lastHuman = r.clock()
	r.changedLocked()
}

// Close ends the draft: the timer stops and every socket is closed. A
// closed room refuses everything.
func (r *Room) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	r.closed = true
	if r.stop != nil {
		r.stop()
		r.stop = nil
	}
	r.gen++
	for seat, set := range r.subs {
		for s := range set {
			s.Close(ClosedReason)
		}
		delete(r.subs, seat)
	}
}

// Pools is every seat's card pool by seat number, and whether the draft
// is done. It is how slice 6 (#5440) reads a finished draft.
func (r *Room) Pools() (pools [][]Card, done bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.d.Pools(), r.d.State() == StateDone
}

// SubmitDeck submits a person's deck for the rounds.
func (r *Room) SubmitDeck(seat uuid.UUID, built LimitedDeck) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrNotBuilding
	}
	if err := r.d.SubmitDeck(seat, built, r.clock()); err != nil {
		return err
	}
	r.lastHuman = r.clock()
	r.changedLocked()
	return nil
}

// SetTable records the table the lobby opened for a match.
func (r *Room) SetTable(round, index int, table uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrNoSuchMatch
	}
	if err := r.d.SetTable(round, index, table); err != nil {
		return err
	}
	r.changedLocked()
	return nil
}

// ReportMatch records a played match's result.
func (r *Room) ReportMatch(round, index, winner, winsA, winsB int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrNoSuchMatch
	}
	if err := r.d.ReportMatch(round, index, winner, winsA, winsB, r.clock()); err != nil {
		return err
	}
	// A result is a person's doing: the table it came from was played.
	r.lastHuman = r.clock()
	r.changedLocked()
	return nil
}

// Touch notes that a person is still here, for the idle sweep: they
// are playing a match at one of the draft's tables, which the draft's
// own sockets do not see.
func (r *Room) Touch() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastHuman = r.clock()
}
