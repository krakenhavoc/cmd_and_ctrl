package draft

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Untimed deck building (#5700, ADR 0222 amendment of 2026-11-11): no
// deadline, nobody is built for on time, the rounds begin when the
// last person has submitted, and a person who has been away for
// UntimedBuildAwayGrace is built for.

// untimedDraft is limitedDraft with an untimed build.
func untimedDraft(t *testing.T, seats, humans int, seed uint64) (*Draft, []uuid.UUID) {
	t.Helper()
	d, err := New(Config{Name: "Untimed", Set: "syn", Mode: PickOne, Seats: seats, Seed: seed, BuildUntimed: true}, t0)
	if err != nil {
		t.Fatal(err)
	}
	var people []uuid.UUID
	for range humans {
		s, err := d.AddSeat("Person", uuid.Nil)
		if err != nil {
			t.Fatal(err)
		}
		people = append(people, s.ID)
	}
	if err := d.Start(people[0], testSource(t), t0); err != nil {
		t.Fatal(err)
	}
	finishPicks(t, d, people)
	if d.Phase() != PhaseBuilding {
		t.Fatalf("after the picks: phase %q", d.Phase())
	}
	return d, people
}

func TestAnUntimedBuildHasNoDeadlineAndWaitsForEveryPerson(t *testing.T) {
	d, people := untimedDraft(t, 4, 2, 31)
	if !d.BuildDeadline().IsZero() || !d.BuildUntimed() {
		t.Fatalf("deadline %v, untimed %v", d.BuildDeadline(), d.BuildUntimed())
	}
	v, _ := d.ViewFor(people[0], t0, nil)
	if v.Build == nil || !v.Build.Untimed || v.Build.Deadline != nil || v.Build.Seconds != 0 {
		t.Fatalf("untimed build view = %+v", v.Build)
	}
	// On the wire there is no deadline to count down to.
	raw, err := json.Marshal(v.Build)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "deadline") || !strings.Contains(string(raw), `"untimed":true`) {
		t.Fatalf("untimed build JSON = %s", raw)
	}
	// No time is long enough to build for anyone.
	d.ExpireBuild(t0.Add(1000 * time.Hour))
	if d.Phase() != PhaseBuilding {
		t.Fatalf("an untimed build ended by the clock: phase %s", d.Phase())
	}
	for _, id := range people {
		if _, built := d.Deck(seatNumber(t, d, id)); built {
			t.Fatal("a person was built for on time")
		}
	}

	// The rounds begin on the last submit, and not before.
	if err := d.SubmitDeck(people[0], legalDeck(d, people[0]), t0); err != nil {
		t.Fatal(err)
	}
	if d.Phase() != PhaseBuilding {
		t.Fatalf("phase %s with a person still building", d.Phase())
	}
	if err := d.SubmitDeck(people[1], legalDeck(d, people[1]), t0); err != nil {
		t.Fatal(err)
	}
	if d.Phase() != PhaseRounds || d.MatchRound() != 1 {
		t.Fatalf("after the last deck: phase %s, round %d", d.Phase(), d.MatchRound())
	}
	if v, _ := d.ViewFor(people[0], t0, nil); v.Build != nil {
		t.Fatalf("a build view in the rounds: %+v", v.Build)
	}
}

// A timed build still shows its deadline and its length, as before.
func TestATimedBuildViewIsUnchanged(t *testing.T) {
	d, people := limitedDraft(t, 4, 2, 32)
	v, _ := d.ViewFor(people[0], t0, nil)
	if v.Build == nil || v.Build.Untimed || v.Build.Deadline == nil ||
		!v.Build.Deadline.Equal(t0.Add(DefaultBuildTimeout)) || v.Build.Seconds != 900 {
		t.Fatalf("timed build view = %+v", v.Build)
	}
	raw, _ := json.Marshal(v.Build)
	if !strings.Contains(string(raw), `"deadline"`) || strings.Contains(string(raw), "untimed") {
		t.Fatalf("timed build JSON = %s", raw)
	}
	// An untimed flag overrides a length.
	u, err := New(Config{Name: "x", Set: "syn", Mode: PickOne, Seats: 2, BuildTimeout: 5 * time.Minute, BuildUntimed: true}, t0)
	if err != nil || !u.BuildUntimed() {
		t.Fatalf("New: %v", err)
	}
}

func seatNumber(t *testing.T, d *Draft, id uuid.UUID) int {
	t.Helper()
	i, s := d.seat(id)
	if s == nil {
		t.Fatal("no such seat")
	}
	return i
}

// BuildFor gives one person the DeckBuilder's deck from their own pool,
// leaves a deck that was submitted alone, and begins the rounds when it
// was the last one the build waited for.
func TestBuildForLetsGoOfOnePerson(t *testing.T) {
	d, people := untimedDraft(t, 4, 2, 33)
	if err := d.SubmitDeck(people[0], legalDeck(d, people[0]), t0); err != nil {
		t.Fatal(err)
	}
	before, _ := d.Deck(0)
	d.BuildFor(people[0]) // already has a deck: nothing changes
	if after, _ := d.Deck(0); after.Size() != before.Size() || len(after.Main) != len(before.Main) || after.Basics["Island"] != 9 {
		t.Fatalf("a submitted deck was rebuilt: %+v", after)
	}
	if d.Phase() != PhaseBuilding {
		t.Fatalf("phase %s", d.Phase())
	}
	d.BuildFor(d.Seats()[3].ID) // a bot: nothing to do
	if got := d.Unbuilt(); len(got) != 1 || got[0].ID != people[1] {
		t.Fatalf("unbuilt = %v", got)
	}
	d.BuildFor(people[1])
	if d.Phase() != PhaseRounds {
		t.Fatalf("phase %s after the last person was built for", d.Phase())
	}
	deck, ok := d.Deck(1)
	if !ok || deck.Size() < 40 {
		t.Fatalf("built deck = %+v, %v", deck, ok)
	}
	d.BuildFor(people[1]) // after the build: nothing happens
}

// A cube draft takes the same choice.
func TestAnUntimedCubeDraftBuild(t *testing.T) {
	src := cubeSource(t, 360, 15)
	d, err := New(Config{Name: "Cube", SetName: "My cube", Booster: "cube", Cube: true, Mode: PickOne, Seats: 4, Seed: 99, BuildUntimed: true}, t0)
	if err != nil {
		t.Fatal(err)
	}
	host, err := d.AddSeat("Host", uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Start(host.ID, src, t0); err != nil {
		t.Fatal(err)
	}
	finishPicks(t, d, []uuid.UUID{host.ID})
	v, _ := d.ViewFor(host.ID, t0, nil)
	if !v.Cube || v.Build == nil || !v.Build.Untimed || v.Build.Deadline != nil {
		t.Fatalf("cube build view: cube %v, build %+v", v.Cube, v.Build)
	}
	d.ExpireBuild(t0.Add(1000 * time.Hour))
	if d.Phase() != PhaseBuilding {
		t.Fatalf("phase %s after the clock", d.Phase())
	}
	if err := d.SubmitDeck(host.ID, legalDeck(d, host.ID), t0); err != nil {
		t.Fatal(err)
	}
	if d.Phase() != PhaseRounds {
		t.Fatalf("phase %s after the only person submitted", d.Phase())
	}
}

// untimedRoom is a room of an untimed draft with humans people, each
// with a socket, taken to the build.
func untimedRoom(t *testing.T, seats, humans int, seed uint64) (*Room, *fakeTime, []uuid.UUID, []*sink, []func()) {
	t.Helper()
	clock := &fakeTime{now: t0}
	d, err := New(Config{Name: "Untimed room", Set: "syn", Mode: PickOne, Seats: seats, Seed: seed, BuildUntimed: true}, t0)
	if err != nil {
		t.Fatal(err)
	}
	r := NewRoom(d, testSource(t))
	r.SetClock(clock.Now, clock.After)
	var ids []uuid.UUID
	var socks []*sink
	var detach []func()
	for range humans {
		id, err := r.Join("Person", uuid.New())
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := r.Start(ids[0]); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		s, off := attach(t, r, id)
		socks = append(socks, s)
		detach = append(detach, off)
	}
	for guard := 0; r.Status().Phase != PhaseBuilding; guard++ {
		if guard > 10000 {
			t.Fatal("the picks never ended")
		}
		for _, id := range ids {
			if v, _ := r.View(id); v.Take > 0 && len(v.Pack) >= v.Take {
				pickFirst(t, r, id)
			}
		}
	}
	return r, clock, ids, socks, detach
}

func roomDeck(t *testing.T, r *Room, id uuid.UUID) LimitedDeck {
	t.Helper()
	v, _ := r.View(id)
	out := LimitedDeck{Basics: map[string]int{"Island": 9, "Swamp": 8}}
	for _, c := range v.Pool[:23] {
		out.Main = append(out.Main, c.ID)
	}
	return out
}

// Someone with their tab open is never built for, however long it
// takes; a person who is away is, UntimedBuildAwayGrace after their
// socket went; and the rounds begin then, with nobody else waited for.
func TestAnUntimedBuildBuildsForAPersonWhoIsAway(t *testing.T) {
	r, clock, ids, socks, detach := untimedRoom(t, 4, 2, 41)
	defer detach[0]()

	clock.Advance(1000 * time.Hour)
	if got := r.Status().Phase; got != PhaseBuilding {
		t.Fatalf("phase %s: someone connected was built for", got)
	}
	if err := r.SubmitDeck(ids[0], roomDeck(t, r, ids[0])); err != nil {
		t.Fatal(err)
	}

	detach[1]()
	clock.Advance(UntimedBuildAwayGrace - time.Second)
	if got := r.Status().Phase; got != PhaseBuilding {
		t.Fatalf("phase %s inside the grace", got)
	}
	clock.Advance(2 * time.Second)
	if got := r.Status().Phase; got != PhaseRounds {
		t.Fatalf("phase %s after the grace", got)
	}
	v := socks[0].last(t)
	if v.Phase != PhaseRounds || v.Play == nil {
		t.Fatalf("the person who stayed was not told the rounds began: %+v", v.Phase)
	}
	var deck LimitedDeck
	var built bool
	r.With(func(d *Draft) { deck, built = d.Deck(seatNumber(t, d, ids[1])) })
	if !built || deck.Size() < 40 {
		t.Fatalf("the absent person's deck = %+v, %v", deck, built)
	}
}

// Coming back starts the grace again, and a person who submitted is
// never built for.
func TestComingBackRestartsTheAwayGrace(t *testing.T) {
	r, clock, ids, _, detach := untimedRoom(t, 2, 1, 42)
	detach[0]()
	clock.Advance(UntimedBuildAwayGrace - time.Minute)
	_, off := attach(t, r, ids[0])
	clock.Advance(time.Hour)
	if got := r.Status().Phase; got != PhaseBuilding {
		t.Fatalf("phase %s while the person is back", got)
	}
	off()
	clock.Advance(UntimedBuildAwayGrace - time.Second)
	if got := r.Status().Phase; got != PhaseBuilding {
		t.Fatalf("phase %s: the grace did not start again", got)
	}
	clock.Advance(2 * time.Second)
	if got := r.Status().Phase; got != PhaseRounds {
		t.Fatalf("phase %s after the second grace", got)
	}
}

// A person who never opened a socket is away from the build's start.
func TestAPersonWhoNeverConnectedIsBuiltForAfterTheGrace(t *testing.T) {
	clock := &fakeTime{now: t0}
	d, err := New(Config{Name: "Nobody home", Set: "syn", Mode: PickOne, Seats: 2, Seed: 43, BuildUntimed: true}, t0)
	if err != nil {
		t.Fatal(err)
	}
	r := NewRoom(d, testSource(t))
	r.SetClock(clock.Now, clock.After)
	id, err := r.Join("Person", uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Start(id); err != nil {
		t.Fatal(err)
	}
	// An absent person is picked for by the away rule until the picks
	// end.
	for guard := 0; r.Status().Phase != PhaseBuilding; guard++ {
		if guard > 10000 {
			t.Fatal("the picks never ended")
		}
		clock.Advance(AwayGrace)
	}
	clock.Advance(UntimedBuildAwayGrace - time.Second)
	if got := r.Status().Phase; got != PhaseBuilding {
		t.Fatalf("phase %s inside the grace", got)
	}
	clock.Advance(2 * time.Second)
	if got := r.Status().Phase; got != PhaseRounds && got != PhaseFinished {
		t.Fatalf("phase %s after the grace", got)
	}
}

// The timed build, through the room, is as it was: the deadline builds
// for a person who is connected the whole time.
func TestATimedRoomStillBuildsAtItsDeadline(t *testing.T) {
	r, clock, ids := newRoom(t, PickOne, 2, 1, 44)
	if err := r.Start(ids[0]); err != nil {
		t.Fatal(err)
	}
	_, off := attach(t, r, ids[0])
	defer off()
	for guard := 0; r.Status().Phase != PhaseBuilding; guard++ {
		if guard > 10000 {
			t.Fatal("the picks never ended")
		}
		if v, _ := r.View(ids[0]); v.Take > 0 && len(v.Pack) >= v.Take {
			pickFirst(t, r, ids[0])
		}
	}
	v, _ := r.View(ids[0])
	if v.Build == nil || v.Build.Untimed || v.Build.Deadline == nil {
		t.Fatalf("build view = %+v", v.Build)
	}
	clock.Advance(UntimedBuildAwayGrace + time.Minute)
	if got := r.Status().Phase; got != PhaseBuilding {
		t.Fatalf("the away grace of an untimed build applied to a timed one: phase %s", got)
	}
	clock.Advance(DefaultBuildTimeout)
	if got := r.Status().Phase; got != PhaseRounds {
		t.Fatalf("phase %s after the deadline", got)
	}
}
