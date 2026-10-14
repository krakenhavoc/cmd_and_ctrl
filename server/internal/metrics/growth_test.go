package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type fakeGrowth struct {
	r     GrowthReading
	err   error
	calls int
	ctxOK bool
}

func (f *fakeGrowth) Growth(ctx context.Context, _ time.Time) (GrowthReading, error) {
	f.calls++
	_, f.ctxOK = ctx.Deadline()
	return f.r, f.err
}

// The growth collector reports a reading under closed labels, reads the
// database once per GrowthCacheTTL under a deadline, keeps the last
// good read on a failure, and reports nothing if it has never had one.
func TestGrowthCollector(t *testing.T) {
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	db := &fakeGrowth{r: GrowthReading{
		Windows: map[string]GrowthWindow{
			"7d": {NewPlayers: 3, ReturningPlayers: 5, GamesFinished: 4, Seats: 14, Turns: 80, Rounds: 20, FinishedPlayers: 6},
		},
		RetentionPrevious: 10, RetentionRetained: 4,
		Cohorts: map[string]GrowthCohort{"1w": {Accounts: 7, Active: 2}},
	}}
	c := NewGrowthCollector(db, quiet()).(*growthCollector)
	c.now = func() time.Time { return now }
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	expect(t, gathered(t, reg), map[string]float64{
		"cmdctrl_players_new{window=7d}":           3,
		"cmdctrl_players_returning{window=7d}":     5,
		"cmdctrl_players_new{window=1d}":           0,
		"cmdctrl_finished_games{window=7d}":        4,
		"cmdctrl_finished_game_seats{window=7d}":   14,
		"cmdctrl_finished_game_turns{window=7d}":   80,
		"cmdctrl_finished_game_rounds{window=7d}":  20,
		"cmdctrl_finished_game_players{window=7d}": 6,
		"cmdctrl_retention_previous_players{}":     10,
		"cmdctrl_retention_retained_players{}":     4,
		"cmdctrl_signup_cohort_accounts{age=1w}":   7,
		"cmdctrl_signup_cohort_active{age=1w}":     2,
		"cmdctrl_signup_cohort_accounts{age=8w}":   0,
	})
	if !db.ctxOK {
		t.Error("the database read has no deadline")
	}
	if err := CheckClosedLabels(reg); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}

	now = now.Add(GrowthCacheTTL - time.Second)
	gathered(t, reg)
	if db.calls != 1 {
		t.Fatalf("database read %d times inside the TTL, want 1", db.calls)
	}
	now = now.Add(2 * time.Second)
	db.err = errors.New("disk on fire")
	expect(t, gathered(t, reg), map[string]float64{"cmdctrl_players_new{window=7d}": 3})
	if db.calls != 2 {
		t.Fatalf("database read %d times past the TTL, want 2", db.calls)
	}
}

func TestGrowthCollectorWithNoReadReportsNothing(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewGrowthCollector(&fakeGrowth{err: errors.New("no")}, quiet()))
	for k := range gathered(t, reg) {
		t.Errorf("series %s reported with no good read", k)
	}
}

// The feature counters count, keep their labels closed, and clamp a
// reason they do not know into other.
func TestFeatureCounters(t *testing.T) {
	reg := NewRegistry()
	before := gathered(t, reg)
	TableStarted(true)
	TableStarted(false)
	TableStarted(false)
	OwnKeySeatAdded()
	OwnKeyFallback(OwnKeyQuota)
	OwnKeyFallback("some provider text with a key in it")
	StatsPageLoaded()
	RematchOffered()
	after := gathered(t, reg)
	for series, want := range map[string]float64{
		"cmdctrl_tables_started_total{kind=practice}":   1,
		"cmdctrl_tables_started_total{kind=real}":       2,
		"cmdctrl_own_key_seats_added_total{}":           1,
		"cmdctrl_own_key_fallbacks_total{reason=quota}": 1,
		"cmdctrl_own_key_fallbacks_total{reason=other}": 1,
		"cmdctrl_stats_page_loads_total{}":              1,
		"cmdctrl_rematches_offered_total{}":             1,
	} {
		got, ok := after[series]
		if !ok {
			t.Errorf("%s: no such series", series)
		} else if got-before[series] != want {
			t.Errorf("%s rose by %v, want %v", series, got-before[series], want)
		}
	}
	if err := CheckClosedLabels(reg); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
}
