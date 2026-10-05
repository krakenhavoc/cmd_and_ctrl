package users

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// The users store's half of ADR 0123 §3: CountUsers, LastPlayed and
// the cmdctrl_users_played windows over real rows, and
// cmdctrl_users_created_total.

type noLive []string

func (n noLive) MetricsLiveUsers() []string { return n }

// seedUser inserts a users row and returns its id. last_seen_at is
// "now" for every one of them, on purpose: it must not decide who
// played.
func seedUser(t *testing.T, d *db.DB, now time.Time) string {
	t.Helper()
	id := uuid.NewString()
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, ?, ?)`,
		id, "p", now.UnixMilli(), now.UnixMilli()); err != nil {
		t.Fatal(err)
	}
	return id
}

// seedGame inserts a games row with the given lifecycle times (nil:
// NULL) and seats userIDs at it ("" is a guest seat).
func seedGame(t *testing.T, d *db.DB, state string, started, ended, archived *time.Time, userIDs ...string) {
	t.Helper()
	ms := func(at *time.Time) any {
		if at == nil {
			return nil
		}
		return at.UnixMilli()
	}
	id := uuid.NewString()
	if _, err := d.Exec(`INSERT INTO games (id, name, state, created_at, started_at, ended_at, archived_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, "g", state, int64(1), ms(started), ms(ended), ms(archived)); err != nil {
		t.Fatal(err)
	}
	for seat, u := range userIDs {
		var user any
		if u != "" {
			user = u
		}
		if _, err := d.Exec(`INSERT INTO seats (game_id, seat, player_id, user_id) VALUES (?, ?, ?, ?)`,
			id, seat, uuid.NewString(), user); err != nil {
			t.Fatal(err)
		}
	}
}

func ago(now time.Time, d time.Duration) *time.Time {
	at := now.Add(-d)
	return &at
}

const day = 24 * time.Hour

// ADR 0123 §3, §11: users_played windowing over seeded rows. An account
// counts in a window when a started table it sat at ended (or was
// closed by archiving) inside it, or when it sits at a live table now.
func TestUsersPlayedWindows(t *testing.T) {
	s, d := openStore(t, nil)
	now := time.Now().UTC()

	recent := seedUser(t, d, now)   // ended 2 hours ago: 1d, 7d, 30d
	lastWeek := seedUser(t, d, now) // ended 3 days ago: 7d, 30d
	closed := seedUser(t, d, now)   // archived while running 10 days ago: 30d
	old := seedUser(t, d, now)      // ended 40 days ago: none
	stuck := seedUser(t, d, now)    // a row still "active" with no end, not live: none
	waiting := seedUser(t, d, now)  // seated at a table that never started: none
	live := seedUser(t, d, now)     // at a live table now: every window
	both := seedUser(t, d, now)     // 40 days ago and live now: every window
	seedUser(t, d, now)             // signed in, never seated: none

	seedGame(t, d, "ended", ago(now, day), ago(now, 2*time.Hour), nil, recent, "")
	seedGame(t, d, "ended", ago(now, 4*day), ago(now, 3*day), nil, lastWeek)
	seedGame(t, d, "active", ago(now, 11*day), nil, ago(now, 10*day), closed)
	seedGame(t, d, "ended", ago(now, 41*day), ago(now, 40*day), ago(now, 39*day), old, both)
	seedGame(t, d, "active", ago(now, 50*day), nil, nil, stuck)
	seedGame(t, d, "lobby", nil, nil, nil, waiting)

	if n, err := s.CountUsers(context.Background()); err != nil || n != 9 {
		t.Fatalf("CountUsers = %d, %v; want 9", n, err)
	}
	ends, err := s.LastPlayed(context.Background(), now.Add(-30*day))
	if err != nil {
		t.Fatal(err)
	}
	if len(ends) != 3 || !ends[recent].Equal(ago(now, 2*time.Hour).Truncate(time.Millisecond)) {
		t.Errorf("LastPlayed = %v, want recent, lastWeek and closed, recent's latest end 2h ago", ends)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(metrics.NewUsersCollector(s, noLive{live, both, live}, nil))
	if n, err := testutil.GatherAndCount(reg, "cmdctrl_users_played"); err != nil || n != 3 {
		t.Fatalf("cmdctrl_users_played series = %d, %v; want 3", n, err)
	}
	// 1d: recent, live, both. 7d: and lastWeek. 30d: and closed.
	for window, want := range map[string]float64{"1d": 3, "7d": 4, "30d": 5} {
		if got := playedIn(t, reg, window); got != want {
			t.Errorf("cmdctrl_users_played{window=%q} = %v, want %v", window, got, want)
		}
	}
	if err := metrics.CheckClosedLabels(reg); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
}

func playedIn(t *testing.T, reg *prometheus.Registry, window string) float64 {
	t.Helper()
	families, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range families {
		if f.GetName() != "cmdctrl_users_played" {
			continue
		}
		for _, m := range f.GetMetric() {
			for _, lp := range m.GetLabel() {
				if lp.GetName() == "window" && lp.GetValue() == window {
					return m.GetGauge().GetValue()
				}
			}
		}
	}
	t.Fatalf("no cmdctrl_users_played{window=%q}", window)
	return 0
}

// A first sign-in is one new account; signing in again is none.
func TestUsersCreatedCountsFirstSignInsOnly(t *testing.T) {
	s, _ := openStore(t, nil)
	created := func() float64 {
		families, err := metrics.Registry.Gather()
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range families {
			if f.GetName() == "cmdctrl_users_created_total" {
				return f.GetMetric()[0].GetCounter().GetValue()
			}
		}
		t.Fatal("no cmdctrl_users_created_total")
		return 0
	}
	before := created()
	ctx := context.Background()
	if _, err := s.UpsertFromDiscord(ctx, alice, "", "identify"); err != nil {
		t.Fatal(err)
	}
	if got := created() - before; got != 1 {
		t.Errorf("first sign-in moved the counter by %v, want 1", got)
	}
	if _, err := s.UpsertFromDiscord(ctx, alice, "", "identify"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpsertFromDiscord(ctx, discord.User{ID: "222", Username: "bob"}, "", "identify"); err != nil {
		t.Fatal(err)
	}
	if got := created() - before; got != 2 {
		t.Errorf("a second sign-in and a new person moved the counter by %v, want 2", got)
	}
}
