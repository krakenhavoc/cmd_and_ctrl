package draft

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// fakeTime is a clock a test moves by hand, and the room's timer: the
// one pending wake-up fires when the clock reaches it.
type fakeTime struct {
	mu   sync.Mutex
	now  time.Time
	at   time.Time
	f    func()
	sets int
}

func (c *fakeTime) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeTime) After(wait time.Duration, f func()) func() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.at, c.f = c.now.Add(wait), f
	c.sets++
	mine := c.sets
	return func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.sets == mine {
			c.f = nil
		}
	}
}

// Advance moves the clock, firing each wake-up that comes due on the
// way, at its own time.
func (c *fakeTime) Advance(d time.Duration) {
	c.mu.Lock()
	end := c.now.Add(d)
	for c.f != nil && !c.at.After(end) {
		f := c.f
		c.f = nil
		c.now = c.at
		c.mu.Unlock()
		f()
		c.mu.Lock()
	}
	c.now = end
	c.mu.Unlock()
}

// sink is a socket: it keeps every view it was sent.
type sink struct {
	mu     sync.Mutex
	views  []View
	closed string
}

func (s *sink) Send(v View) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.views = append(s.views, v)
}

func (s *sink) Close(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = reason
}

func (s *sink) last(t *testing.T) View {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.views) == 0 {
		t.Fatal("the socket was sent nothing")
	}
	return s.views[len(s.views)-1]
}

func (s *sink) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.views)
}

func newRoom(t *testing.T, mode Mode, seats, humans int, seed uint64) (*Room, *fakeTime, []uuid.UUID) {
	t.Helper()
	clock := &fakeTime{now: t0}
	d, err := New(Config{Name: "Room test", Set: "syn", Mode: mode, Seats: seats, Seed: seed}, t0)
	if err != nil {
		t.Fatal(err)
	}
	r := NewRoom(d, testSource(t))
	r.SetClock(clock.Now, clock.After)
	var ids []uuid.UUID
	for range humans {
		id, err := r.Join("Person", uuid.New())
		if err != nil {
			t.Fatalf("Join: %v", err)
		}
		ids = append(ids, id)
	}
	return r, clock, ids
}

func attach(t *testing.T, r *Room, seat uuid.UUID) (*sink, func()) {
	t.Helper()
	s := &sink{}
	detach, err := r.Attach(seat, s)
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	return s, detach
}

func pickFirst(t *testing.T, r *Room, seat uuid.UUID) {
	t.Helper()
	v, ok := r.View(seat)
	if !ok || len(v.Pack) < v.Take || v.Take == 0 {
		t.Fatalf("seat has nothing to pick: take %d of %d", v.Take, len(v.Pack))
	}
	var ids []string
	for _, c := range v.Pack[:v.Take] {
		ids = append(ids, c.ID)
	}
	if err := r.Pick(seat, ids); err != nil {
		t.Fatalf("Pick: %v", err)
	}
}

// The first frame a socket gets is its seat's view, and every change
// after sends a new one with a higher seq.
func TestASocketGetsItsViewFirstAndEveryChangeAfter(t *testing.T) {
	r, _, seats := newRoom(t, PickOne, 4, 2, 1)
	a, _ := attach(t, r, seats[0])
	b, _ := attach(t, r, seats[1])
	if v := a.last(t); v.State != StateLobby || v.You != 0 {
		t.Fatalf("first frame: %+v", v)
	}
	if err := r.Start(seats[0]); err != nil {
		t.Fatalf("Start: %v", err)
	}
	va, vb := a.last(t), b.last(t)
	if va.State != StatePicking || len(va.Pack) != 14 || vb.You != 1 || len(vb.Pack) != 14 {
		t.Fatalf("after start: a %s/%d b you=%d/%d", va.State, len(va.Pack), vb.You, len(vb.Pack))
	}
	if va.Pack[0].ID == vb.Pack[0].ID {
		t.Fatal("two seats were sent the same pack")
	}
	before := a.last(t).Seq
	pickFirst(t, r, seats[0])
	if v := a.last(t); v.Seq <= before || v.Pack != nil || len(v.Picks) != 1 {
		t.Fatalf("after a's pick: seq %d (was %d), pack %d, picks %d", v.Seq, before, len(v.Pack), len(v.Picks))
	}
	// b is told a has picked, and still holds its own pack.
	if v := b.last(t); v.Players[0].Picking || !v.Players[1].Picking || len(v.Pack) != 14 {
		t.Fatalf("b after a's pick: %+v", v.Players)
	}
	seqs := map[uint64]bool{}
	var prev uint64
	a.mu.Lock()
	for _, v := range a.views {
		if v.Seq < prev {
			t.Fatalf("seq went backwards: %d after %d", v.Seq, prev)
		}
		prev = v.Seq
		seqs[v.Seq] = true
	}
	a.mu.Unlock()
}

// The timer: a connected seat that stalls is picked for when the timer
// runs out, not before, and the draft moves on to the next pick.
func TestTheTimerPicksForAStalledSeat(t *testing.T) {
	r, clock, seats := newRoom(t, PickOne, 4, 2, 2)
	a, _ := attach(t, r, seats[0])
	b, _ := attach(t, r, seats[1])
	if err := r.Start(seats[0]); err != nil {
		t.Fatal(err)
	}
	pickFirst(t, r, seats[0])
	clock.Advance(DefaultPickTimeout - time.Second)
	if v := b.last(t); v.Pick != 1 || len(v.Picks) != 0 {
		t.Fatalf("picked for before the timer ran out: pick %d, %d picks", v.Pick, len(v.Picks))
	}
	held := map[string]bool{}
	for _, c := range b.last(t).Pack {
		held[c.ID] = true
	}
	clock.Advance(time.Second)
	v := b.last(t)
	if len(v.Picks) != 1 || !held[v.Picks[0].ID] {
		t.Fatalf("the timer's pick: %+v", v.Picks)
	}
	if v.Pick != 2 || len(v.Pack) != 13 {
		t.Fatalf("after the timer: pick %d with a pack of %d, want pick 2 with 13", v.Pick, len(v.Pack))
	}
	if got := a.last(t); got.Pick != 2 || got.Deadline == nil || !got.Deadline.Equal(clock.Now().Add(DefaultPickTimeout)) {
		t.Fatalf("a after the timer: pick %d deadline %v", got.Pick, got.Deadline)
	}
	// Left alone, the timer drafts the whole thing for both.
	clock.Advance(time.Duration(3*14) * DefaultPickTimeout)
	if v := a.last(t); v.State != StateDone || len(v.Pool) != 42 {
		t.Fatalf("left to the timer: state %s, pool %d", v.State, len(v.Pool))
	}
}

// A person who disconnects is picked for the same way, AwayGrace into
// each pick, and comes back to their seat and their picks.
func TestADisconnectedSeatIsPickedForAndCanComeBack(t *testing.T) {
	r, clock, seats := newRoom(t, PickOne, 4, 2, 3)
	a, _ := attach(t, r, seats[0])
	b, detachB := attach(t, r, seats[1])
	if err := r.Start(seats[0]); err != nil {
		t.Fatal(err)
	}
	pickFirst(t, r, seats[0])
	pickFirst(t, r, seats[1])
	mine := b.last(t).Picks[0].ID

	clock.Advance(2 * time.Second)
	detachB()
	if v := a.last(t); v.Players[1].Connected {
		t.Fatal("a still sees b as connected")
	}
	// a picks; b is away, so only b is waited on, for AwayGrace from
	// when it went.
	pickFirst(t, r, seats[0])
	clock.Advance(AwayGrace - time.Second)
	if v := a.last(t); v.Pick != 2 {
		t.Fatalf("b was picked for before the grace ran out (pick %d)", v.Pick)
	}
	clock.Advance(time.Second)
	if v := a.last(t); v.Pick != 3 || v.Players[1].Picked != 2 {
		t.Fatalf("after the grace: pick %d, b has %d", v.Pick, v.Players[1].Picked)
	}
	// The next pick: b is away from its start, so it is picked for
	// AwayGrace in, without a waiting for the whole timer.
	pickFirst(t, r, seats[0])
	clock.Advance(AwayGrace)
	if v := a.last(t); v.Pick != 4 || v.Players[1].Picked != 3 {
		t.Fatalf("second away pick: pick %d, b has %d", v.Pick, v.Players[1].Picked)
	}

	// b comes back: same seat, the pick it made itself first, then the
	// two made for it, and the pack it holds now.
	b2, _ := attach(t, r, seats[1])
	v := b2.last(t)
	if v.You != 1 || len(v.Picks) != 3 || v.Picks[0].ID != mine || len(v.Pack) != 11 || v.Take != 1 {
		t.Fatalf("b on return: you %d, %d picks, first %s (want %s), pack %d", v.You, len(v.Picks), v.Picks[0].ID, mine, len(v.Pack))
	}
	if got := a.last(t); !got.Players[1].Connected {
		t.Fatal("a does not see b come back")
	}
	// Back, b has the whole timer again.
	pickFirst(t, r, seats[0])
	clock.Advance(AwayGrace + time.Second)
	if v := b2.last(t); len(v.Picks) != 3 {
		t.Fatal("a connected seat was picked for at the away grace")
	}
	pickFirst(t, r, seats[1])
	if v := b2.last(t); len(v.Picks) != 4 || v.Pick != 5 {
		t.Fatalf("b's own pick after returning: %d picks, pick %d", len(v.Picks), v.Pick)
	}
}

// A seat that never connected (the host started and someone had not
// opened the page) is away from the first pick.
func TestASeatThatNeverConnectedIsAway(t *testing.T) {
	r, clock, seats := newRoom(t, PickOne, 2, 2, 4)
	a, _ := attach(t, r, seats[0])
	if err := r.Start(seats[0]); err != nil {
		t.Fatal(err)
	}
	pickFirst(t, r, seats[0])
	clock.Advance(AwayGrace)
	if v := a.last(t); v.Pick != 2 {
		t.Fatalf("pick %d, want 2", v.Pick)
	}
}

// Two tabs on one seat: both get every frame, and the seat is away only
// when the last one goes.
func TestTwoSocketsOnOneSeat(t *testing.T) {
	r, _, seats := newRoom(t, PickOne, 2, 2, 5)
	one, detachOne := attach(t, r, seats[0])
	two, detachTwo := attach(t, r, seats[0])
	other, _ := attach(t, r, seats[1])
	if err := r.Start(seats[0]); err != nil {
		t.Fatal(err)
	}
	if one.last(t).Seq != two.last(t).Seq || len(two.last(t).Pack) != 14 {
		t.Fatal("the second tab is not sent what the first is")
	}
	detachOne()
	detachOne() // twice is harmless
	if !other.last(t).Players[0].Connected {
		t.Fatal("the seat reads as away with one tab still open")
	}
	detachTwo()
	if other.last(t).Players[0].Connected {
		t.Fatal("the seat reads as connected with no tab open")
	}
}

func TestLeavingClosesTheSeatsSockets(t *testing.T) {
	r, _, seats := newRoom(t, PickOne, 4, 2, 6)
	a, _ := attach(t, r, seats[0])
	b, _ := attach(t, r, seats[1])
	humans, err := r.Leave(seats[1])
	if err != nil || humans != 1 {
		t.Fatalf("Leave = %d, %v", humans, err)
	}
	if b.closed != ClosedReason {
		t.Fatalf("the leaver's socket was closed with %q", b.closed)
	}
	if v := a.last(t); len(v.Players) != 1 {
		t.Fatalf("a sees %d players after b left the lobby", len(v.Players))
	}
	if _, err := r.Attach(seats[1], &sink{}); !errors.Is(err, ErrNoSuchSeat) {
		t.Fatalf("attaching to a seat that left = %v", err)
	}
	if r.HasSeat(seats[1]) || !r.HasSeat(seats[0]) {
		t.Fatal("HasSeat")
	}
}

// A person who leaves during the picks: their seat is a bot from then
// on, their socket closes, and nobody waits on it.
func TestLeavingDuringThePicks(t *testing.T) {
	r, _, seats := newRoom(t, PickOne, 4, 2, 7)
	a, _ := attach(t, r, seats[0])
	b, _ := attach(t, r, seats[1])
	if err := r.Start(seats[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Leave(seats[1]); err != nil {
		t.Fatal(err)
	}
	if b.closed != ClosedReason {
		t.Fatal("the leaver's socket stayed open")
	}
	if v := a.last(t); !v.Players[1].Bot || v.Players[1].Picking {
		t.Fatalf("the seat that was left: %+v", v.Players[1])
	}
	for a.last(t).State == StatePicking {
		pickFirst(t, r, seats[0])
	}
	pools, done := r.Pools()
	if !done || len(pools) != 4 || len(pools[1]) != 42 {
		t.Fatalf("pools: done %v, %d seats, seat 1 has %d", done, len(pools), len(pools[1]))
	}
}

func TestClosingARoom(t *testing.T) {
	r, clock, seats := newRoom(t, PickOne, 2, 1, 8)
	a, _ := attach(t, r, seats[0])
	if err := r.Start(seats[0]); err != nil {
		t.Fatal(err)
	}
	n := a.count()
	r.Close()
	r.Close()
	if a.closed != ClosedReason {
		t.Fatalf("closed with %q", a.closed)
	}
	clock.Advance(time.Hour)
	r.Tick()
	if a.count() != n {
		t.Fatal("a closed room sent a view")
	}
	if err := r.Pick(seats[0], []string{"c1"}); err == nil {
		t.Fatal("a closed room took a pick")
	}
	if _, err := r.Attach(seats[0], &sink{}); err == nil {
		t.Fatal("a closed room took a socket")
	}
	if st := r.Status(); !st.Closed {
		t.Fatal("Status does not say closed")
	}
}

func TestStatusCountsPeopleAndRemembersThem(t *testing.T) {
	r, clock, seats := newRoom(t, PickOne, 8, 2, 9)
	if st := r.Status(); st.State != StateLobby || st.Humans != 2 || st.Connected != 0 {
		t.Fatalf("status: %+v", st)
	}
	clock.Advance(time.Minute)
	_, detach := attach(t, r, seats[0])
	if st := r.Status(); st.Connected != 1 || !st.LastHuman.Equal(clock.Now()) {
		t.Fatalf("status after a socket: %+v", st)
	}
	clock.Advance(time.Minute)
	detach()
	if st := r.Status(); st.Connected != 0 || !st.LastHuman.Equal(clock.Now()) {
		t.Fatalf("status after it went: %+v", st)
	}
	if got := r.SeatOfUser(uuid.New()); got != uuid.Nil {
		t.Fatal("a stranger holds a seat")
	}
}

// A room is away when nobody has had a socket for AwayGrace: a seat
// that never connected is away from the start, and one whose socket has
// just dropped is not, until the grace has run (#5698).
func TestARoomIsAwayOnlyAfterTheGrace(t *testing.T) {
	r, clock, seats := newRoom(t, PickOne, 8, 2, 9)
	if !r.Status().Away {
		t.Fatal("a room nobody ever connected to is not away")
	}
	_, detachA := attach(t, r, seats[0])
	_, detachB := attach(t, r, seats[1])
	if r.Status().Away {
		t.Fatal("a room with sockets is away")
	}
	detachA()
	clock.Advance(AwayGrace)
	detachB()
	if st := r.Status(); st.Connected != 0 || st.Away {
		t.Fatalf("a socket that has just gone: %+v, want not away", st)
	}
	clock.Advance(AwayGrace - time.Second)
	if r.Status().Away {
		t.Fatal("away before the grace has run")
	}
	clock.Advance(time.Second)
	if !r.Status().Away {
		t.Fatal("not away once the grace has run")
	}
	_, detach := attach(t, r, seats[1])
	if r.Status().Away {
		t.Fatal("away with a socket back")
	}
	detach()
}

// The real timer, once: a draft of one person who never picks finishes
// on its own.
func TestTheRealTimerDrivesADraft(t *testing.T) {
	d, err := New(Config{Mode: PickTwo, Seats: 2, Seed: 10, PickTimeout: time.Millisecond}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	r := NewRoom(d, testSource(t))
	seat, err := r.Join("Host", uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := attach(t, r, seat)
	if err := r.Start(seat); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, done := r.Pools(); done {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	pools, done := r.Pools()
	if !done || len(pools[0]) != 42 {
		t.Fatalf("done %v, %d cards", done, len(pools[0]))
	}
	if v := s.last(t); v.State != StateDone {
		t.Fatalf("the socket's last view is %s", v.State)
	}
}
