package schedule

// Tests for the plans and their lists (ADR 0215 §3, §10, §12 and the
// privacy section): the limits, the order of the list with its
// waitlist, moves, cancels, removal, the holds after doors-open driven
// by the clock, the sweep and account deletion.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/moderation"
)

var ctx = context.Background()

// t0 is the fixture's "now"; plans start a day later.
var t0 = time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

type fixture struct {
	s   *Store
	d   *db.DB
	now time.Time
}

func newFixture(t *testing.T, people ...string) *fixture {
	t.Helper()
	d, err := db.Open(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	f := &fixture{s: NewStore(d), d: d, now: t0}
	for _, p := range people {
		f.user(t, p)
	}
	return f
}

// user adds an account with a Discord identity: its id is p, its
// Discord id "d-"+p and its name p capitalised.
func (f *fixture) user(t *testing.T, p string) {
	t.Helper()
	name := strings.ToUpper(p[:1]) + p[1:]
	if _, err := f.d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`, p, name); err != nil {
		t.Fatal(err)
	}
	if _, err := f.d.Exec(`INSERT INTO identities (provider, subject, user_id, display_name, scopes, linked_at)
		VALUES ('discord', ?, ?, ?, 'identify', 1)`, "d-"+p, p, name); err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) game(t *testing.T) string {
	t.Helper()
	id := uuid.NewString()
	if _, err := f.d.Exec(`INSERT INTO games (id, name, state, created_at) VALUES (?, 'T', 'lobby', 1)`, id); err != nil {
		t.Fatal(err)
	}
	return id
}

func draft(start time.Time, players int) Draft {
	return Draft{Name: "Friday pod", StartsAt: start, Zone: "America/New_York", Players: players}
}

func (f *fixture) plan(t *testing.T, host string, players int) Plan {
	t.Helper()
	p, err := f.s.Create(ctx, f.now, host, draft(f.now.Add(24*time.Hour), players), nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	return p
}

func (f *fixture) reserve(t *testing.T, plan, who string) {
	t.Helper()
	f.now = f.now.Add(time.Second)
	if _, _, err := f.s.Reserve(ctx, f.now, plan, who); err != nil {
		t.Fatalf("Reserve %s: %v", who, err)
	}
}

// states is the list in order, as "who:state".
func (f *fixture) states(t *testing.T, plan string) string {
	t.Helper()
	_, rs, err := f.s.Get(ctx, plan)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, r := range rs {
		out = append(out, strings.TrimPrefix(r.DiscordID, "d-")+":"+string(r.State))
	}
	return strings.Join(out, " ")
}

func TestCreateHoldsTheHostsSeatAndChecksTheShape(t *testing.T) {
	f := newFixture(t, "ann")
	p := f.plan(t, "ann", 0)
	if p.Players != DefaultPlayers || p.LengthMinutes != DefaultLengthMinutes || p.Status != StatusScheduled {
		t.Errorf("defaults: players %d length %d status %s", p.Players, p.LengthMinutes, p.Status)
	}
	if p.HostName != "Ann" || p.HostDiscordID != "d-ann" {
		t.Errorf("host = %q %q", p.HostName, p.HostDiscordID)
	}
	if got := f.states(t, p.ID); got != "ann:seat" {
		t.Errorf("list = %q, want the host holding seat 1", got)
	}
	start := f.now.Add(24 * time.Hour)
	for name, tc := range map[string]struct {
		d    Draft
		want error
	}{
		"too soon":   {draft(f.now.Add(29*time.Minute), 4), ErrTooSoon},
		"too far":    {draft(f.now.Add(MaxAhead+time.Minute), 4), ErrTooFar},
		"one player": {draft(start, 1), ErrBadPlayers},
		"five":       {draft(start, 5), ErrBadPlayers},
		"length":     {Draft{Name: "x", StartsAt: start, Zone: "UTC", LengthMinutes: 90}, ErrBadLength},
		"zone":       {Draft{Name: "x", StartsAt: start, Zone: "Mars/Olympus"}, ErrBadZone},
		"no zone":    {Draft{Name: "x", StartsAt: start}, ErrBadZone},
		"setup":      {Draft{Name: "x", StartsAt: start, Zone: "UTC", Options: Options{Setup: "first"}}, ErrBadSetup},
		"blank":      {Draft{Name: "  ", StartsAt: start, Zone: "UTC"}, ErrNameInvalid},
		"two lines":  {Draft{Name: "a\nb", StartsAt: start, Zone: "UTC"}, ErrNameInvalid},
	} {
		if _, err := f.s.Create(ctx, f.now, "ann", tc.d, nil); !errors.Is(err, tc.want) {
			t.Errorf("%s: %v, want %v", name, err, tc.want)
		}
	}
	if _, err := f.s.Create(ctx, f.now, "ann", Draft{Name: "retarded", StartsAt: start, Zone: "UTC"}, nil); !errors.Is(err, moderation.ErrNameNotAllowed) {
		t.Errorf("a filtered name: %v", err)
	}
	long := strings.Repeat("x", 100)
	if got, err := f.s.Create(ctx, f.now, "ann", Draft{Name: long, StartsAt: start, Zone: "UTC"}, nil); err != nil || len(got.Name) != MaxNameLen {
		t.Errorf("a long name: %q %v, want it cut to %d", got.Name, err, MaxNameLen)
	}
}

func TestCreateCapsAndTheCreateBucket(t *testing.T) {
	f := newFixture(t, "ann", "bea")
	for i := 0; i < MaxUpcomingPerHost; i++ {
		f.plan(t, "ann", 4)
	}
	if _, err := f.s.Create(ctx, f.now, "ann", draft(f.now.Add(time.Hour), 4), nil); !errors.Is(err, ErrTooManyPlans) {
		t.Errorf("a fourth upcoming plan: %v, want ErrTooManyPlans", err)
	}
	// The bucket is spent only by a create that passed every other check.
	spent := 0
	allow := func() bool { spent++; return false }
	if _, err := f.s.Create(ctx, f.now, "ann", draft(f.now.Add(time.Hour), 4), allow); !errors.Is(err, ErrTooManyPlans) || spent != 0 {
		t.Errorf("refused by the cap: %v, bucket spent %d times", err, spent)
	}
	if _, err := f.s.Create(ctx, f.now, "bea", draft(f.now.Add(time.Hour), 4), allow); !errors.Is(err, ErrRateLimited) || spent != 1 {
		t.Errorf("past the bucket: %v, spent %d", err, spent)
	}
	// The server-wide cap, with plans of many hosts.
	if _, err := f.d.Exec(`UPDATE scheduled_tables SET status = 'scheduled'`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < MaxUpcoming-MaxUpcomingPerHost; i++ {
		host := fmt.Sprintf("h%02d", i)
		f.user(t, host)
		f.plan(t, host, 4)
	}
	if _, err := f.s.Create(ctx, f.now, "bea", draft(f.now.Add(time.Hour), 4), nil); !errors.Is(err, ErrCalendarFull) {
		t.Errorf("the 51st upcoming plan: %v, want ErrCalendarFull", err)
	}
}

func TestTheListIsSeatsInOrderThenTheWaitlist(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee", "eve")
	p := f.plan(t, "ann", 3)
	for _, who := range []string{"bea", "cal", "dee", "eve"} {
		f.reserve(t, p.ID, who)
	}
	if got, want := f.states(t, p.ID), "ann:seat bea:seat cal:seat dee:waiting eve:waiting"; got != want {
		t.Fatalf("list = %q, want %q", got, want)
	}
	// Reserving again keeps the place.
	f.reserve(t, p.ID, "bea")
	// Leaving moves the first on the waitlist up at once.
	if _, _, gone, err := f.s.Leave(ctx, f.now, p.ID, "bea"); err != nil || gone.State != StateSeat {
		t.Fatalf("Leave: %v %+v", err, gone)
	}
	if got, want := f.states(t, p.ID), "ann:seat cal:seat dee:seat eve:waiting"; got != want {
		t.Errorf("after a leave: %q, want %q", got, want)
	}
	// Coming back is at the back.
	f.reserve(t, p.ID, "bea")
	if got, want := f.states(t, p.ID), "ann:seat cal:seat dee:seat eve:waiting bea:waiting"; got != want {
		t.Errorf("after coming back: %q, want %q", got, want)
	}
	// The host removes Dee from this occurrence: Eve moves up, and Dee
	// cannot reserve again.
	if _, _, _, err := f.s.Remove(ctx, f.now, p.ID, "bea", "d-dee"); !errors.Is(err, ErrNotHost) {
		t.Errorf("a removal by someone else: %v", err)
	}
	if _, _, _, err := f.s.Remove(ctx, f.now, p.ID, "ann", "d-dee"); err != nil {
		t.Fatal(err)
	}
	if got, want := f.states(t, p.ID), "ann:seat cal:seat dee:removed eve:seat bea:waiting"; got != want {
		t.Errorf("after a removal: %q, want %q", got, want)
	}
	if _, _, err := f.s.Reserve(ctx, f.now, p.ID, "dee"); !errors.Is(err, ErrRemoved) {
		t.Errorf("a removed person reserving: %v, want ErrRemoved", err)
	}
	if _, _, _, err := f.s.Leave(ctx, f.now, p.ID, "dee"); !errors.Is(err, ErrNotOnList) {
		t.Errorf("a removed person leaving: %v, want ErrNotOnList", err)
	}
	// The host holds seat 1 and does neither.
	if _, _, err := f.s.Reserve(ctx, f.now, p.ID, "ann"); !errors.Is(err, ErrIsHost) {
		t.Errorf("the host reserving: %v", err)
	}
	if _, _, _, err := f.s.Leave(ctx, f.now, p.ID, "ann"); !errors.Is(err, ErrIsHost) {
		t.Errorf("the host leaving: %v", err)
	}
	if _, _, _, err := f.s.Remove(ctx, f.now, p.ID, "ann", "d-ann"); !errors.Is(err, ErrIsHost) {
		t.Errorf("the host removing themselves: %v", err)
	}
	// Fewer players: the latest seats are cut, to the front of the
	// waitlist.
	two := 2
	if _, _, _, err := f.s.Update(ctx, f.now, p.ID, "ann", Patch{Players: &two}, nil); err != nil {
		t.Fatal(err)
	}
	if got, want := f.states(t, p.ID), "ann:seat cal:seat dee:removed eve:waiting bea:waiting"; got != want {
		t.Errorf("after fewer players: %q, want %q", got, want)
	}
}

func TestMovesCancelsAndTheDoors(t *testing.T) {
	f := newFixture(t, "ann", "bea")
	p := f.plan(t, "ann", 4)
	at := func(d time.Duration) *time.Time { t := f.now.Add(d); return &t }
	if _, _, _, err := f.s.Update(ctx, f.now, p.ID, "bea", Patch{StartsAt: at(48 * time.Hour)}, nil); !errors.Is(err, ErrNotHost) {
		t.Errorf("a move by someone else: %v", err)
	}
	if _, _, _, err := f.s.Update(ctx, f.now, p.ID, "ann", Patch{StartsAt: at(10 * time.Minute)}, nil); !errors.Is(err, ErrTooSoon) {
		t.Errorf("a move too soon: %v", err)
	}
	if _, _, _, err := f.s.Update(ctx, f.now, p.ID, "ann", Patch{StartsAt: at(48 * time.Hour)}, func() bool { return false }); !errors.Is(err, ErrRateLimited) {
		t.Errorf("a move past the bucket: %v", err)
	}
	for i := 1; i <= MaxMoves; i++ {
		got, _, moved, err := f.s.Update(ctx, f.now, p.ID, "ann", Patch{StartsAt: at(time.Duration(24+i) * time.Hour)}, nil)
		if err != nil || !moved || got.Moves != i {
			t.Fatalf("move %d: %v moved %v moves %d", i, err, moved, got.Moves)
		}
	}
	if _, _, _, err := f.s.Update(ctx, f.now, p.ID, "ann", Patch{StartsAt: at(72 * time.Hour)}, nil); !errors.Is(err, ErrTooManyMoves) {
		t.Errorf("a fourth move: %v", err)
	}
	// The same start again is not a move.
	if _, _, moved, err := f.s.Update(ctx, f.now, p.ID, "ann", Patch{StartsAt: at(27 * time.Hour)}, nil); err != nil || moved {
		t.Errorf("the same start: %v moved %v", err, moved)
	}
	// Doors open 15 minutes before the start: no change after that.
	late := p.StartsAt.Add(3*time.Hour - 10*time.Minute)
	name := "Renamed"
	if _, _, _, err := f.s.Update(ctx, late, p.ID, "ann", Patch{Name: &name}, nil); !errors.Is(err, ErrDoorsOpen) {
		t.Errorf("a change at doors-open: %v", err)
	}
	// Cancel: someone else no, an admin yes.
	if _, _, err := f.s.Cancel(ctx, f.now, p.ID, "bea", false); !errors.Is(err, ErrNotHost) {
		t.Errorf("a cancel by someone else: %v", err)
	}
	got, _, err := f.s.Cancel(ctx, f.now, p.ID, "", true)
	if err != nil || got.Status != StatusCancelled {
		t.Fatalf("an admin's cancel: %v %s", err, got.Status)
	}
	if _, _, err := f.s.Cancel(ctx, f.now, p.ID, "ann", false); !errors.Is(err, ErrClosed) {
		t.Errorf("a second cancel: %v", err)
	}
	if _, _, err := f.s.Reserve(ctx, f.now, p.ID, "bea"); !errors.Is(err, ErrClosed) {
		t.Errorf("reserving a cancelled plan: %v", err)
	}
}

// openPlan opens p's doors at their time onto a fresh games row.
func (f *fixture) openPlan(t *testing.T, p Plan) Plan {
	t.Helper()
	f.now = p.DoorsOpenAt()
	got, _, err := f.s.Open(ctx, f.now, p.ID, f.game(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return got
}

func (f *fixture) advance(t *testing.T, p Plan, tb Table) Outcome {
	t.Helper()
	_, _, out, err := f.s.Advance(ctx, f.now, p.ID, tb)
	if err != nil {
		t.Fatalf("Advance: %v", err)
	}
	return out
}

func table(taken int, seated ...string) Table {
	tb := Table{Live: true, Taken: taken, Users: map[string]bool{}, Discord: map[string]bool{}}
	for _, s := range seated {
		tb.Users[s] = true
	}
	return tb
}

func TestTheHoldsAfterDoorsOpen(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee", "eve")
	p := f.plan(t, "ann", 3)
	for _, who := range []string{"bea", "cal", "dee", "eve"} {
		f.reserve(t, p.ID, who)
	}
	p = f.openPlan(t, p)
	if p.Status != StatusOpen || p.GameID == "" {
		t.Fatalf("opened: %s %q", p.Status, p.GameID)
	}
	if got, want := f.states(t, p.ID), "ann:held bea:held cal:held dee:waiting eve:waiting"; got != want {
		t.Fatalf("at doors-open: %q, want %q", got, want)
	}
	_, rs, _ := f.s.Get(ctx, p.ID)
	if !rs[0].HeldUntil.Equal(p.GraceEndsAt()) {
		t.Errorf("a hold runs until %v, want the grace's end %v", rs[0].HeldUntil, p.GraceEndsAt())
	}

	// Bea sits down; nothing else changes before the grace ends.
	f.now = p.StartsAt
	if out := f.advance(t, p, table(1, "bea")); len(out.Seated) != 1 || out.Settled {
		t.Errorf("Bea sat down: %+v", out)
	}
	if got, want := f.states(t, p.ID), "ann:held bea:seated cal:held dee:waiting eve:waiting"; got != want {
		t.Errorf("after Bea sat: %q, want %q", got, want)
	}

	// Cal leaves after doors-open: the seat goes down the waitlist at
	// once, held until the grace ends (10 more minutes is sooner).
	if _, _, _, err := f.s.Leave(ctx, f.now, p.ID, "cal"); err != nil {
		t.Fatal(err)
	}
	out := f.advance(t, p, table(1, "bea"))
	if len(out.Promoted) != 1 || out.Promoted[0].DiscordID != "d-dee" || !out.Promoted[0].HeldUntil.Equal(p.GraceEndsAt()) {
		t.Errorf("after Cal left: %+v", out.Promoted)
	}

	// The grace ends: Ann's and Dee's holds lapse; Eve gets 10 minutes,
	// and there is one more free seat than people waiting.
	f.now = p.GraceEndsAt()
	out = f.advance(t, p, table(1, "bea"))
	if len(out.Lapsed) != 2 || len(out.Promoted) != 1 || out.Promoted[0].DiscordID != "d-eve" ||
		!out.Promoted[0].HeldUntil.Equal(f.now.Add(WaitlistHold)) || out.Settled {
		t.Errorf("at the grace's end: %+v", out)
	}
	if got, want := f.states(t, p.ID), "ann:lapsed bea:seated dee:lapsed eve:held"; got != want {
		t.Errorf("at the grace's end: %q, want %q", got, want)
	}

	// A lapsed person who reserves again goes to the back of the waitlist.
	f.reserve(t, p.ID, "dee")
	if got, want := f.states(t, p.ID), "ann:lapsed bea:seated eve:held dee:waiting"; got != want {
		t.Errorf("Dee again: %q, want %q", got, want)
	}

	// Eve's hold lapses too; Dee's comes and goes; then the plan settles.
	f.now = f.now.Add(WaitlistHold)
	out = f.advance(t, p, table(1, "bea"))
	if len(out.Promoted) != 1 || out.Promoted[0].DiscordID != "d-dee" || out.Settled {
		t.Errorf("Eve lapsed: %+v", out)
	}
	f.now = f.now.Add(WaitlistHold)
	out = f.advance(t, p, table(1, "bea"))
	if !out.Settled || out.FillBots != 0 {
		t.Errorf("the last hold lapsed: %+v, want settled with no bots (not asked)", out)
	}
	got, _, _ := f.s.Get(ctx, p.ID)
	if got.Status != StatusSettled {
		t.Errorf("status %s, want settled", got.Status)
	}
	// Settled is the end: a later advance changes nothing.
	if out := f.advance(t, got, table(1, "bea")); out.Settled || len(out.Lapsed)+len(out.Promoted) != 0 {
		t.Errorf("an advance after settling: %+v", out)
	}
}

func TestFillBotsAndAStartedTable(t *testing.T) {
	f := newFixture(t, "ann", "bea")
	p, err := f.s.Create(ctx, f.now, "ann", Draft{Name: "Bots", StartsAt: f.now.Add(time.Hour), Zone: "UTC",
		Players: 4, Options: Options{FillBots: true}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f.reserve(t, p.ID, "bea")
	p = f.openPlan(t, p)
	f.now = p.GraceEndsAt()
	out := f.advance(t, p, table(1, "ann"))
	if !out.Settled || out.FillBots != 3 {
		t.Errorf("settled with Ann seated: %+v, want 3 bots", out)
	}

	// A table that started before the grace ended settles at once, and
	// every hold lapses.
	q, err := f.s.Create(ctx, f.now, "ann", Draft{Name: "Early", StartsAt: f.now.Add(time.Hour), Zone: "UTC",
		Players: 4, Options: Options{FillBots: true}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f.reserve(t, q.ID, "bea")
	q = f.openPlan(t, q)
	out = f.advance(t, q, Table{Live: false})
	if !out.Settled || len(out.Lapsed) != 2 || out.FillBots != 0 {
		t.Errorf("a started table: %+v", out)
	}
}

func TestReservingAnOpenPlanJoinsTheWaitlist(t *testing.T) {
	f := newFixture(t, "ann", "bea")
	p := f.plan(t, "ann", 2)
	p = f.openPlan(t, p)
	f.reserve(t, p.ID, "bea")
	if got, want := f.states(t, p.ID), "ann:held bea:waiting"; got != want {
		t.Errorf("a reservation after doors-open: %q, want %q", got, want)
	}
	// Bea joined by the invite link while she waited: she is seated.
	out := f.advance(t, p, table(1, "bea"))
	if len(out.Seated) != 1 || out.Seated[0].DiscordID != "d-bea" {
		t.Errorf("Bea seated by the link: %+v", out)
	}
}

func TestMissedAndActive(t *testing.T) {
	f := newFixture(t, "ann")
	p := f.plan(t, "ann", 4)
	if got, err := f.s.Active(ctx, f.now); err != nil || len(got) != 0 {
		t.Errorf("before doors-open: %d active, %v", len(got), err)
	}
	if got, _ := f.s.Active(ctx, p.DoorsOpenAt()); len(got) != 1 {
		t.Errorf("at doors-open: %d active, want 1", len(got))
	}
	got, _, err := f.s.Miss(ctx, p.StartsAt, p.ID)
	if err != nil || got.Status != StatusMissed {
		t.Fatalf("Miss: %v %s", err, got.Status)
	}
	if got, _ := f.s.Active(ctx, p.StartsAt); len(got) != 0 {
		t.Errorf("a missed plan is still active")
	}
	if all, _ := f.s.List(ctx); len(all) != 1 || all[0].Plan.Status != StatusMissed {
		t.Errorf("a missed plan is listed until it is swept: %+v", all)
	}
}

func TestSweepKeepsAPlanADayAfterItsStartOrCancel(t *testing.T) {
	f := newFixture(t, "ann", "bea")
	kept := f.plan(t, "ann", 4)
	f.reserve(t, kept.ID, "bea")
	cancelled := f.plan(t, "ann", 4)
	cancelAt := f.now.Add(time.Hour)
	if _, _, err := f.s.Cancel(ctx, cancelAt, cancelled.ID, "ann", false); err != nil {
		t.Fatal(err)
	}
	count := func(q string) int {
		var n int
		if err := f.d.QueryRow(q).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if n, err := f.s.Sweep(ctx, cancelAt.Add(Keep-time.Minute)); err != nil || n != 0 {
		t.Errorf("before a day: swept %d, %v", n, err)
	}
	if n, _ := f.s.Sweep(ctx, cancelAt.Add(Keep)); n != 1 {
		t.Errorf("a day after the cancel: swept %d, want the cancelled plan", n)
	}
	if n, _ := f.s.Sweep(ctx, kept.StartsAt.Add(Keep-time.Minute)); n != 0 {
		t.Errorf("before a day after the start: swept %d", n)
	}
	if n, _ := f.s.Sweep(ctx, kept.StartsAt.Add(Keep)); n != 1 {
		t.Errorf("a day after the start: swept %d, want 1", n)
	}
	if n := count(`SELECT COUNT(*) FROM scheduled_rsvps`); n != 0 {
		t.Errorf("%d RSVPs left after their plans were swept", n)
	}
}

func TestForgetAccountDeletesHostedPlansAndEveryRSVP(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee")
	hosted := f.plan(t, "ann", 4)
	f.reserve(t, hosted.ID, "bea")
	open := f.openPlan(t, f.plan(t, "ann", 4))
	f.now = t0
	theirs := f.plan(t, "bea", 2)
	f.reserve(t, theirs.ID, "ann")
	f.reserve(t, theirs.ID, "cal")
	// A row with Ann's Discord id and no account (slice 5's Interested).
	other := f.plan(t, "dee", 4)
	if _, err := f.d.Exec(`INSERT INTO scheduled_rsvps (schedule_id, discord_id, source, state, rsvp_at, updated_at)
		VALUES (?, 'd-ann', 'discord', 'seat', 99, 99)`, other.ID); err != nil {
		t.Fatal(err)
	}

	var c AccountCounts
	err := func() error {
		tx, err := f.d.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		if c, err = ForgetAccount(ctx, tx, "ann", "d-ann", f.now); err != nil {
			return err
		}
		return tx.Commit()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if c.PlansDeleted != 2 || c.PlansCancelled != 1 || c.RSVPs != 2 || len(c.Games) != 1 || c.Games[0] != open.GameID {
		t.Errorf("counts = %+v", c)
	}
	for _, id := range []string{hosted.ID, open.ID} {
		if _, _, err := f.s.Get(ctx, id); !errors.Is(err, ErrNotFound) {
			t.Errorf("a hosted plan is still there: %v", err)
		}
	}
	// Cal moves up from the waitlist of Bea's two-seat table.
	if got, want := f.states(t, theirs.ID), "bea:seat cal:seat"; got != want {
		t.Errorf("Bea's table: %q, want %q", got, want)
	}
	if got, want := f.states(t, other.ID), "dee:seat"; got != want {
		t.Errorf("Dee's table: %q, want %q (the row by Discord id went)", got, want)
	}
}

func TestAdvanceIsPure(t *testing.T) {
	// advance never touches a plan that is not open.
	p := &Plan{Status: StatusScheduled, Players: 4, StartsAt: t0}
	rs := []RSVP{{DiscordID: "a", State: StateSeat}}
	if out := advance(p, rs, Table{}, t0); out.Settled || rs[0].State != StateSeat {
		t.Errorf("advance changed a scheduled plan: %+v %+v", out, rs)
	}
}
