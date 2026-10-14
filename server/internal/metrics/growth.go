package metrics

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// The growth collector (ADR 0123 amendment, #3557): how many accounts
// are new or returning, how many of last week's players came back, how
// signup cohorts hold on, and how much the finished games are played.
// Aggregate only: every number is a count over accounts or games, and
// the only labels are closed sets (window, age). The database computes
// them (GrowthSource), once per GrowthCacheTTL, exactly as users.go
// does; main.go registers it only when there is a database, so with
// none the families are absent rather than zero.
//
// "Played" means what cmdctrl_users_played means from the rows: seated
// at a started table that ended (or was archived) inside the window. A
// table still running is not counted here, only once it ends.
// "First game" is an account's earliest such end, so a new player in a
// window is one whose first finished game ended inside it.

// GrowthCacheTTL is how long the growth collector reuses its last read.
// Longer than the users collector's: the queries group every seat of
// every finished table, and weekly views do not need a fresher number.
const GrowthCacheTTL = 5 * time.Minute

// growthQueryTimeout bounds one refresh of the cache.
const growthQueryTimeout = 10 * time.Second

// Ages of cmdctrl_signup_cohort_*: whole weeks since the accounts were
// created. An account is in cohort "1w" when it was created between 14
// and 7 days ago, "4w" between 35 and 28 days ago, "8w" between 63 and
// 56 days ago.
var signupCohortAges = []struct {
	label string
	weeks int
}{
	{"1w", 1},
	{"4w", 4},
	{"8w", 8},
}

// SignupCohortWeeks is the age in weeks of the cohort labelled label,
// and false for a label the gauges do not have.
func SignupCohortWeeks(label string) (int, bool) {
	for _, a := range signupCohortAges {
		if a.label == label {
			return a.weeks, true
		}
	}
	return 0, false
}

// SignupCohortLabels are the age labels, in order.
func SignupCohortLabels() []string { return append([]string(nil), ageLabels...) }

var ageLabels = func() []string {
	out := make([]string, len(signupCohortAges))
	for i, a := range signupCohortAges {
		out[i] = a.label
	}
	return out
}()

// GrowthWindow is one window's reading.
type GrowthWindow struct {
	// NewPlayers played in the window and had no finished game before it.
	NewPlayers int
	// ReturningPlayers played in the window and had one before it.
	ReturningPlayers int
	// GamesFinished are the games with recorded statistics that ended in
	// the window; Seats, Turns and Rounds are their sums.
	GamesFinished, Seats, Turns, Rounds int
	// FinishedPlayers are the distinct accounts seated at those games.
	FinishedPlayers int
}

// GrowthCohort is one signup cohort's reading.
type GrowthCohort struct {
	// Accounts were created in the cohort's week.
	Accounts int
	// Active of them played in the last 7 days.
	Active int
}

// GrowthReading is everything the growth collector reports.
type GrowthReading struct {
	// Windows is keyed by the window label (1d, 7d, 30d).
	Windows map[string]GrowthWindow
	// RetentionPrevious is the accounts that played in the 7 days before
	// the last 7; RetentionRetained is those of them that also played in
	// the last 7.
	RetentionPrevious, RetentionRetained int
	// Cohorts is keyed by the age label (1w, 4w, 8w).
	Cohorts map[string]GrowthCohort
}

// GrowthSource is the database half of the growth collector.
type GrowthSource interface {
	// Growth reads the numbers as of now.
	Growth(ctx context.Context, now time.Time) (GrowthReading, error)
}

var (
	playersNewDesc = prometheus.NewDesc("cmdctrl_players_new",
		"Accounts whose first finished game ended inside the window. Read from the database, cached 5 min.",
		[]string{"window"}, nil)
	playersReturningDesc = prometheus.NewDesc("cmdctrl_players_returning",
		"Accounts that played in the window and had a finished game before it. Read from the database, cached 5 min.",
		[]string{"window"}, nil)
	retentionPreviousDesc = prometheus.NewDesc("cmdctrl_retention_previous_players",
		"Accounts that played in the 7 days before the last 7. Divide cmdctrl_retention_retained_players by it for week-over-week retention.", nil, nil)
	retentionRetainedDesc = prometheus.NewDesc("cmdctrl_retention_retained_players",
		"Of the accounts that played in the 7 days before the last 7, those that played in the last 7 as well.", nil, nil)
	cohortAccountsDesc = prometheus.NewDesc("cmdctrl_signup_cohort_accounts",
		"Accounts created in the 7-day band that ended age weeks ago (1w: 14 to 7 days ago).",
		[]string{"age"}, nil)
	cohortActiveDesc = prometheus.NewDesc("cmdctrl_signup_cohort_active",
		"Of the accounts in cmdctrl_signup_cohort_accounts{age}, those that played in the last 7 days.",
		[]string{"age"}, nil)
	finishedGamesDesc = prometheus.NewDesc("cmdctrl_finished_games",
		"Games with recorded statistics that ended in the window.", []string{"window"}, nil)
	finishedSeatsDesc = prometheus.NewDesc("cmdctrl_finished_game_seats",
		"Seats (people and bots) at the games counted in cmdctrl_finished_games. Divide by that for seats per game.",
		[]string{"window"}, nil)
	finishedTurnsDesc = prometheus.NewDesc("cmdctrl_finished_game_turns",
		"Turns played in the games counted in cmdctrl_finished_games. Divide by that for turns per game.",
		[]string{"window"}, nil)
	finishedRoundsDesc = prometheus.NewDesc("cmdctrl_finished_game_rounds",
		"Rounds played in the games counted in cmdctrl_finished_games. Divide by that for rounds per game.",
		[]string{"window"}, nil)
	finishedPlayersDesc = prometheus.NewDesc("cmdctrl_finished_game_players",
		"Distinct signed-in accounts seated at the games counted in cmdctrl_finished_games. Divide cmdctrl_finished_games by it for games per active player.",
		[]string{"window"}, nil)
)

type growthCollector struct {
	db  GrowthSource
	log *slog.Logger
	now func() time.Time

	mu     sync.Mutex
	readAt time.Time
	ok     bool
	last   GrowthReading
}

// NewGrowthCollector returns the growth collector over the database.
// log may be nil.
func NewGrowthCollector(db GrowthSource, log *slog.Logger) prometheus.Collector {
	if log == nil {
		log = slog.Default()
	}
	return &growthCollector{db: db, log: log, now: time.Now}
}

func (c *growthCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range []*prometheus.Desc{
		playersNewDesc, playersReturningDesc, retentionPreviousDesc, retentionRetainedDesc,
		cohortAccountsDesc, cohortActiveDesc, finishedGamesDesc, finishedSeatsDesc,
		finishedTurnsDesc, finishedRoundsDesc, finishedPlayersDesc,
	} {
		ch <- d
	}
}

// read returns the cached reading, refreshing it when it is older than
// GrowthCacheTTL. A failed refresh keeps the last good one and is tried
// again on the next scrape; ok is false only when there has never been
// one.
func (c *growthCollector) read(now time.Time) (GrowthReading, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.ok || now.Sub(c.readAt) >= GrowthCacheTTL {
		ctx, cancel := context.WithTimeout(context.Background(), growthQueryTimeout)
		defer cancel()
		r, err := c.db.Growth(ctx, now)
		if err != nil {
			c.log.Warn("metrics: reading growth numbers failed; serving the last read", "err", err)
		} else {
			c.last, c.readAt, c.ok = r, now, true
		}
	}
	return c.last, c.ok
}

func (c *growthCollector) Collect(ch chan<- prometheus.Metric) {
	r, ok := c.read(c.now())
	if !ok {
		// No read has ever succeeded: report nothing rather than a zero
		// that looks like a fact.
		return
	}
	g := func(d *prometheus.Desc, v int, lv ...string) {
		ch <- prometheus.MustNewConstMetric(d, prometheus.GaugeValue, float64(v), lv...)
	}
	for _, w := range usersPlayedWindows {
		x := r.Windows[w.label]
		g(playersNewDesc, x.NewPlayers, w.label)
		g(playersReturningDesc, x.ReturningPlayers, w.label)
		g(finishedGamesDesc, x.GamesFinished, w.label)
		g(finishedSeatsDesc, x.Seats, w.label)
		g(finishedTurnsDesc, x.Turns, w.label)
		g(finishedRoundsDesc, x.Rounds, w.label)
		g(finishedPlayersDesc, x.FinishedPlayers, w.label)
	}
	g(retentionPreviousDesc, r.RetentionPrevious)
	g(retentionRetainedDesc, r.RetentionRetained)
	for _, a := range signupCohortAges {
		x := r.Cohorts[a.label]
		g(cohortAccountsDesc, x.Accounts, a.label)
		g(cohortActiveDesc, x.Active, a.label)
	}
}
