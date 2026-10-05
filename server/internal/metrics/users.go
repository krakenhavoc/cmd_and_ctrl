package metrics

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// The users collector (ADR 0123 §3): registered accounts and how many
// of them played recently, read from SQLite and cached for
// UsersCacheTTL. main.go registers it only when there is a database;
// with none, cmdctrl_users and cmdctrl_users_played are absent.
//
// "Played in the window" is the distinct accounts seated at a table
// that was active at some point in the window. Two sources make it:
//
//   - the rows (UserSource.LastPlayed): each account's latest end of a
//     table it sat at, the table's ended_at or, for one closed by
//     archiving, its archived_at. An account counts in a window when
//     that end is inside it.
//   - the lobby (LiveUserSource): the accounts seated at a table that
//     is active right now, which played in every window.
//
// users.last_seen_at is not used: it is written only at sign-in.
// Neither is a games row in state active with no end: it is either a
// live table, which the lobby reports, or one that did not come back
// after a restart, which nobody is playing.

// UsersCacheTTL is how long the users collector reuses its last read
// of the database.
const UsersCacheTTL = 60 * time.Second

// usersQueryTimeout bounds one refresh of the cache.
const usersQueryTimeout = 5 * time.Second

// Windows of cmdctrl_users_played.
var usersPlayedWindows = []struct {
	label string
	span  time.Duration
}{
	{"1d", 24 * time.Hour},
	{"7d", 7 * 24 * time.Hour},
	{"30d", 30 * 24 * time.Hour},
}

var windowLabels = []string{"1d", "7d", "30d"}

// UsersPlayedWindow is the span of cmdctrl_users_played{window=label},
// and false for a label the gauge does not have. The admin views'
// played filter (ADR 0124 §3.1) reads its windows here, so the page and
// the tile count over the same span.
func UsersPlayedWindow(label string) (time.Duration, bool) {
	for _, w := range usersPlayedWindows {
		if w.label == label {
			return w.span, true
		}
	}
	return 0, false
}

// UserSource is the database half of the users collector.
type UserSource interface {
	// CountUsers is the number of users rows.
	CountUsers(ctx context.Context) (int, error)
	// LastPlayed maps each account seated at a started table whose end
	// is at or after since to the latest such end.
	LastPlayed(ctx context.Context, since time.Time) (map[string]time.Time, error)
}

// LiveUserSource is the lobby half: the accounts (users ids) seated at
// a table that is active now, practice and archived tables excluded.
type LiveUserSource interface {
	MetricsLiveUsers() []string
}

var (
	usersDesc = prometheus.NewDesc("cmdctrl_users",
		"Registered accounts (users rows). Read from the database, cached 60 s.", nil, nil)
	usersPlayedDesc = prometheus.NewDesc("cmdctrl_users_played",
		"Distinct accounts seated at a table that was active in the window. Read from the database and the lobby, cached 60 s.",
		[]string{"window"}, nil)
)

type usersCollector struct {
	db   UserSource
	live LiveUserSource
	log  *slog.Logger
	now  func() time.Time

	mu      sync.Mutex
	readAt  time.Time
	ok      bool
	total   int
	lastEnd map[string]time.Time
}

// NewUsersCollector returns the users collector over the database and
// the lobby. log may be nil.
func NewUsersCollector(db UserSource, live LiveUserSource, log *slog.Logger) prometheus.Collector {
	if log == nil {
		log = slog.Default()
	}
	return &usersCollector{db: db, live: live, log: log, now: time.Now}
}

func (c *usersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- usersDesc
	ch <- usersPlayedDesc
}

// read returns the cached database read, refreshing it when it is
// older than UsersCacheTTL. A failed refresh keeps the last good read
// (and is tried again on the next scrape); ok is false only when there
// has never been one.
func (c *usersCollector) read(now time.Time) (total int, lastEnd map[string]time.Time, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.ok || now.Sub(c.readAt) >= UsersCacheTTL {
		ctx, cancel := context.WithTimeout(context.Background(), usersQueryTimeout)
		defer cancel()
		oldest := usersPlayedWindows[len(usersPlayedWindows)-1].span
		n, err := c.db.CountUsers(ctx)
		var ends map[string]time.Time
		if err == nil {
			ends, err = c.db.LastPlayed(ctx, now.Add(-oldest))
		}
		if err != nil {
			c.log.Warn("metrics: reading user counts failed; serving the last read", "err", err)
		} else {
			c.total, c.lastEnd, c.readAt, c.ok = n, ends, now, true
		}
	}
	return c.total, c.lastEnd, c.ok
}

func (c *usersCollector) Collect(ch chan<- prometheus.Metric) {
	now := c.now()
	total, lastEnd, ok := c.read(now)
	if !ok {
		// No read has ever succeeded: report nothing rather than a zero
		// that looks like a fact.
		return
	}
	// The lobby is read after the collector's own lock is released.
	live := c.live.MetricsLiveUsers()

	ch <- prometheus.MustNewConstMetric(usersDesc, prometheus.GaugeValue, float64(total))
	for _, w := range usersPlayedWindows {
		since := now.Add(-w.span)
		played := make(map[string]bool, len(live))
		for _, u := range live {
			played[u] = true
		}
		for u, end := range lastEnd {
			if !end.Before(since) {
				played[u] = true
			}
		}
		ch <- prometheus.MustNewConstMetric(usersPlayedDesc, prometheus.GaugeValue, float64(len(played)), w.label)
	}
}
