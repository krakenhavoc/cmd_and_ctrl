package metrics

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// Tests for the games, players and WebSocket metrics (ADR 0123 §3):
// the collectors over fake sources, the event helpers' closed sets,
// and the label guard over all of them.

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

type fakeTables []Table

func (f fakeTables) MetricsTables() []Table { return f }

type fakeSockets []Socket

func (f fakeSockets) MetricsSockets() []Socket { return f }

func key(b byte) Key { return Key{b} }

// gathered is one registry's series as "name{l=v,…}" → value, labels
// sorted. Histograms report their sample count.
func gathered(t *testing.T, g prometheus.Gatherer) map[string]float64 {
	t.Helper()
	families, err := g.Gather()
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]float64{}
	for _, f := range families {
		for _, m := range f.GetMetric() {
			var ls []string
			for _, lp := range m.GetLabel() {
				ls = append(ls, lp.GetName()+"="+lp.GetValue())
			}
			sort.Strings(ls)
			var v float64
			switch {
			case m.GetCounter() != nil:
				v = m.GetCounter().GetValue()
			case m.GetGauge() != nil:
				v = m.GetGauge().GetValue()
			case m.GetHistogram() != nil:
				v = float64(m.GetHistogram().GetSampleCount())
			}
			out[f.GetName()+"{"+strings.Join(ls, ",")+"}"] = v
		}
	}
	return out
}

func expect(t *testing.T, got map[string]float64, want map[string]float64) {
	t.Helper()
	for k, v := range want {
		g, ok := got[k]
		if !ok {
			t.Errorf("%s: no such series", k)
			continue
		}
		if g != v {
			t.Errorf("%s = %v, want %v", k, g, v)
		}
	}
}

// The collector over a lobby with every case in it: a running table
// with a connected guest, an unconnected signed-in person, a bot and
// an agent; an archived running table; a table in the lobby; an ended
// one; and a practice table. Sockets: the guest twice (two tabs), the
// agent, a spectator, a seatless admin, and a socket for a seat at the
// archived table.
func TestTablesCollector(t *testing.T) {
	running := Table{Game: key(1), State: "active", Seats: []Seat{
		{Player: key(11), Kind: SeatHuman},
		{Player: key(12), Kind: SeatHuman, SignedIn: true},
		{Player: key(13), Kind: SeatBot},
		{Player: key(14), Kind: SeatAgent},
	}}
	archived := Table{Game: key(2), State: "active", Archived: true, Seats: []Seat{
		{Player: key(21), Kind: SeatHuman, SignedIn: true},
	}}
	waiting := Table{Game: key(3), State: "lobby", Seats: []Seat{{Player: key(31), Kind: SeatHuman}}}
	ended := Table{Game: key(4), State: "ended", Seats: []Seat{{Player: key(41), Kind: SeatHuman}}}
	practice := Table{Game: key(5), State: "active", Practice: true, Seats: []Seat{
		{Player: key(51), Kind: SeatHuman, SignedIn: true}, {Player: key(52), Kind: SeatBot},
	}}
	sockets := fakeSockets{
		{Game: key(1), Player: key(11), Role: RoleSeat},
		{Game: key(1), Player: key(11), Role: RoleSeat},
		{Game: key(1), Player: key(14), Role: RoleSeat},
		{Game: key(1), Role: RoleSpectator},
		{Game: key(1), Role: RoleAdmin},
		{Game: key(2), Player: key(21), Role: RoleSeat},
		{Game: key(5), Player: key(51), Role: RoleSeat},
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewTablesCollector(fakeTables{running, archived, waiting, ended, practice}, sockets))

	expect(t, gathered(t, reg), map[string]float64{
		"cmdctrl_games{archived=false,state=active}": 1,
		"cmdctrl_games{archived=true,state=active}":  1,
		"cmdctrl_games{archived=false,state=lobby}":  1,
		"cmdctrl_games{archived=false,state=ended}":  1,
		"cmdctrl_games{archived=true,state=lobby}":   0,
		"cmdctrl_games{archived=true,state=ended}":   0,
		"cmdctrl_practice_games{}":                   1,
		// Only the running, unarchived, non-practice table's seats.
		"cmdctrl_seats{kind=human}": 2,
		"cmdctrl_seats{kind=bot}":   1,
		"cmdctrl_seats{kind=agent}": 1,
		// The guest's two tabs are one seat; the signed-in seat has no
		// socket; the bot never does.
		"cmdctrl_seats_connected{account=guest,kind=human}":     1,
		"cmdctrl_seats_connected{account=signed_in,kind=human}": 0,
		"cmdctrl_seats_connected{account=guest,kind=agent}":     1,
		"cmdctrl_seats_connected{account=signed_in,kind=agent}": 0,
		"cmdctrl_spectators_connected{}":                        1,
		// Every live socket, whatever its table.
		"cmdctrl_ws_connections{role=seat}":      5,
		"cmdctrl_ws_connections{role=spectator}": 1,
		"cmdctrl_ws_connections{role=admin}":     1,
	})
	if err := CheckClosedLabels(reg); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
}

// A state outside the closed set is not a series.
func TestTablesCollectorDropsAnUnknownState(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewTablesCollector(fakeTables{{Game: key(1), State: "paused"}}, fakeSockets{}))
	if err := CheckClosedLabels(reg); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
	for k, v := range gathered(t, reg) {
		if strings.HasPrefix(k, "cmdctrl_games{") && v != 0 {
			t.Errorf("%s = %v, want 0", k, v)
		}
	}
}

type fakeUsers struct {
	mu    sync.Mutex
	calls int
	total int
	ends  map[string]time.Time
	err   error
	since []time.Time
}

func (f *fakeUsers) CountUsers(context.Context) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.total, f.err
}

func (f *fakeUsers) LastPlayed(_ context.Context, since time.Time) (map[string]time.Time, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.since = append(f.since, since)
	out := map[string]time.Time{}
	for u, e := range f.ends {
		if !e.Before(since) {
			out[u] = e
		}
	}
	return out, f.err
}

type fakeLive []string

func (f fakeLive) MetricsLiveUsers() []string { return f }

// users_played counts an account in every window its latest end falls
// in, plus every account at a live table in all three; the database is
// read once a minute at most, and a failed read keeps the last one.
func TestUsersCollectorWindowsAndCache(t *testing.T) {
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	db := &fakeUsers{total: 9, ends: map[string]time.Time{
		"a": now.Add(-2 * time.Hour),       // 1d
		"b": now.Add(-3 * 24 * time.Hour),  // 7d
		"c": now.Add(-20 * 24 * time.Hour), // 30d
		"d": now.Add(-40 * 24 * time.Hour), // none
		"e": now.Add(-10 * 24 * time.Hour), // 30d, and live now
	}}
	c := NewUsersCollector(db, fakeLive{"e", "f", "f"}, quiet()).(*usersCollector)
	c.now = func() time.Time { return now }
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	want := map[string]float64{
		"cmdctrl_users{}":                  9,
		"cmdctrl_users_played{window=1d}":  3, // a, e, f
		"cmdctrl_users_played{window=7d}":  4, // a, b, e, f
		"cmdctrl_users_played{window=30d}": 5, // a, b, c, e, f
	}
	expect(t, gathered(t, reg), want)
	if got := db.since[0]; !got.Equal(now.Add(-30 * 24 * time.Hour)) {
		t.Errorf("LastPlayed since %v, want 30 days back", got)
	}

	// Inside the TTL: no second read.
	now = now.Add(UsersCacheTTL - time.Second)
	db.total = 10
	gathered(t, reg)
	if db.calls != 1 {
		t.Fatalf("database read %d times inside the TTL, want 1", db.calls)
	}

	// Past it: read again.
	now = now.Add(2 * time.Second)
	expect(t, gathered(t, reg), map[string]float64{"cmdctrl_users{}": 10})
	if db.calls != 2 {
		t.Fatalf("database read %d times past the TTL, want 2", db.calls)
	}

	// A failed read keeps the last good one, and is tried again.
	now = now.Add(UsersCacheTTL)
	db.err = errors.New("disk on fire")
	expect(t, gathered(t, reg), map[string]float64{"cmdctrl_users{}": 10})
	now = now.Add(time.Second)
	gathered(t, reg)
	if db.calls != 4 {
		t.Errorf("failed reads retried %d times, want every scrape", db.calls-2)
	}
	if err := CheckClosedLabels(reg); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
}

// With no good read ever, the users families are absent, not zero.
func TestUsersCollectorWithNoReadReportsNothing(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewUsersCollector(&fakeUsers{err: errors.New("no")}, fakeLive{"x"}, quiet()))
	for k := range gathered(t, reg) {
		t.Errorf("series %s reported with no good read", k)
	}
}

type fakeDB struct {
	size int64
	err  error
	last time.Time
}

func (f fakeDB) SizeBytes() (int64, error) { return f.size, f.err }
func (f fakeDB) LastBackup() time.Time     { return f.last }

func TestDBCollector(t *testing.T) {
	last := time.UnixMilli(1_790_000_000_500)
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewDBCollector(fakeDB{size: 4096, last: last}, quiet()))
	expect(t, gathered(t, reg), map[string]float64{
		"cmdctrl_db_size_bytes{}":                            4096,
		"cmdctrl_db_backup_last_success_timestamp_seconds{}": 1_790_000_000.5,
	})

	// No backup yet is 0; an unreadable size is absent.
	reg = prometheus.NewRegistry()
	reg.MustRegister(NewDBCollector(fakeDB{err: errors.New("gone")}, quiet()))
	got := gathered(t, reg)
	if _, ok := got["cmdctrl_db_size_bytes{}"]; ok {
		t.Error("an unreadable size was reported")
	}
	expect(t, got, map[string]float64{"cmdctrl_db_backup_last_success_timestamp_seconds{}": 0})
}

// Each event helper keeps its label in the closed set, whatever it is
// handed.
func TestEventHelpersKeepTheirSetsClosed(t *testing.T) {
	cases := []struct {
		name string
		do   func()
		c    prometheus.Counter
	}{
		{"seats below the set", func() { GameStarted(0) }, gamesStarted.WithLabelValues("1")},
		{"seats above the set", func() { GameStarted(11) }, gamesStarted.WithLabelValues("8")},
		{"seats in the set", func() { GameStarted(3) }, gamesStarted.WithLabelValues("3")},
		{"no outcome", func() { GameEnded("", -1) }, gamesEnded.WithLabelValues(OutcomeClosed)},
		{"unknown outcome", func() { GameEnded("abandoned", -1) }, gamesEnded.WithLabelValues(OutcomeClosed)},
		{"a win", func() { GameEnded(OutcomeWin, time.Minute) }, gamesEnded.WithLabelValues(OutcomeWin)},
		{"unknown frame kind", func() { WSFrame(FrameIn, "<script>") }, wsFrames.WithLabelValues(FrameIn, FrameOther)},
		{"known frame kind", func() { WSFrame(FrameOut, "legal_moves") }, wsFrames.WithLabelValues(FrameOut, "legal_moves")},
		{"unknown rejection", func() { WSUpgradeRejected("teapot") }, wsRejections.WithLabelValues(RejectRejected)},
		{"known rejection", func() { WSUpgradeRejected(RejectGameNotFound) }, wsRejections.WithLabelValues(RejectGameNotFound)},
	}
	for _, c := range cases {
		before := testutil.ToFloat64(c.c)
		c.do()
		if got := testutil.ToFloat64(c.c) - before; got != 1 {
			t.Errorf("%s: moved by %v, want 1", c.name, got)
		}
	}

	// A start with no known length is counted and not observed.
	obs := histogramCount(t)
	GameEnded(OutcomeDraw, -1)
	if histogramCount(t) != obs {
		t.Error("an end with no known start was observed")
	}
	GameEnded(OutcomeDraw, time.Hour)
	if histogramCount(t) != obs+1 {
		t.Error("an end with a known start was not observed")
	}

	if err := CheckClosedLabels(Registry); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
}

func histogramCount(t *testing.T) uint64 {
	t.Helper()
	return uint64(gathered(t, Registry)["cmdctrl_game_duration_seconds{}"])
}

func TestWSRole(t *testing.T) {
	cases := []struct {
		seated, readOnly bool
		want             string
	}{
		{true, false, RoleSeat},
		{true, true, RoleSeat},
		{false, true, RoleSpectator},
		{false, false, RoleAdmin},
	}
	for _, c := range cases {
		if got := WSRole(c.seated, c.readOnly); got != c.want {
			t.Errorf("WSRole(%v, %v) = %q, want %q", c.seated, c.readOnly, got, c.want)
		}
	}
}

// Every family this change adds, registered on one registry with the
// event metrics, keeps the label rule; and a label name two families
// share with different sets is checked against each family's own.
func TestStateFamiliesKeepLabelsClosed(t *testing.T) {
	reg := NewRegistry()
	reg.MustRegister(
		NewTablesCollector(fakeTables{{Game: key(1), State: "active", Seats: []Seat{{Player: key(2), Kind: SeatHuman}}}},
			fakeSockets{{Game: key(1), Player: key(2), Role: RoleSeat}}),
		NewUsersCollector(&fakeUsers{total: 1}, fakeLive{}, quiet()),
		NewDBCollector(fakeDB{size: 1}, quiet()),
	)
	if err := CheckClosedLabels(reg); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
	names := map[string]bool{}
	for k := range gathered(t, reg) {
		names[k[:strings.IndexByte(k, '{')]] = true
	}
	for _, want := range []string{
		"cmdctrl_games", "cmdctrl_practice_games", "cmdctrl_seats", "cmdctrl_seats_connected",
		"cmdctrl_spectators_connected", "cmdctrl_ws_connections", "cmdctrl_users", "cmdctrl_users_played",
		"cmdctrl_db_size_bytes", "cmdctrl_db_backup_last_success_timestamp_seconds",
		"cmdctrl_games_created_total", "cmdctrl_games_started_total", "cmdctrl_games_ended_total",
		"cmdctrl_game_duration_seconds", "cmdctrl_users_created_total", "cmdctrl_ws_connects_total",
		"cmdctrl_ws_disconnects_total", "cmdctrl_ws_upgrade_rejections_total", "cmdctrl_ws_frames_total",
		"cmdctrl_ws_broadcast_seconds",
	} {
		if !names[want] {
			t.Errorf("registry has no %s", want)
		}
	}

	// The family rows are the family's own: a frame direction is not a
	// valid outcome, and an outcome is not a valid frame direction.
	for _, c := range []struct {
		family, label, value, want string
	}{
		{"cmdctrl_games_ended_total", "outcome", "in", `outcome="in" is outside its set`},
		{"cmdctrl_ws_frames_total", "direction", "win", `direction="win" is outside its set`},
		{"cmdctrl_ws_frames_total", "type", "cast_spell", `type="cast_spell" is outside its set`},
		{"cmdctrl_t_other", "outcome", "win", `label "outcome" is not in the label table`},
	} {
		r := prometheus.NewRegistry()
		r.MustRegister(gaugeWith(c.family, c.label, c.value))
		err := CheckClosedLabels(r)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s{%s=%q}: got %v, want an error containing %q", c.family, c.label, c.value, err, c.want)
		}
	}
}
