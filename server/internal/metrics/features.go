package metrics

import (
	"slices"

	"github.com/prometheus/client_golang/prometheus"
)

// The feature-use counters (ADR 0123 amendment, #3557): how often a
// few features are used, as plain totals. None carries a person, a
// table or a card; the only labels are closed sets spelled here.
// They answer "is this used at all", not "who used it".

// Table kinds of cmdctrl_tables_started_total.
const (
	// TableReal is a table that counts (games.go's started tables).
	TableReal = "real"
	// TablePractice is a practice table (one human against bots, never
	// stored, never in the game counters).
	TablePractice = "practice"
)

var tableKindLabels = []string{TableReal, TablePractice}

// Reasons of cmdctrl_own_key_fallbacks_total: why an own-key bot seat
// stopped using its owner's model key and played the built-in
// heuristic (ADR 0182). They are short codes for modelkeys' fixed
// sentences, which this package may not import. OwnKeyOther takes
// anything else.
const (
	OwnKeyRefused     = "refused"
	OwnKeyQuota       = "quota"
	OwnKeyModel       = "model"
	OwnKeyUnreadable  = "unreadable"
	OwnKeyPrivate     = "private"
	OwnKeyUnreachable = "unreachable"
	OwnKeyTimeout     = "timeout"
	OwnKeyRateLimited = "rate_limited"
	OwnKeyBadReply    = "bad_reply"
	OwnKeyRemoved     = "removed"
	OwnKeyDailyLimit  = "daily_limit"
	OwnKeyCallLimit   = "call_limit"
	OwnKeyOther       = "other"
)

var ownKeyReasonLabels = []string{
	OwnKeyRefused, OwnKeyQuota, OwnKeyModel, OwnKeyUnreadable, OwnKeyPrivate,
	OwnKeyUnreachable, OwnKeyTimeout, OwnKeyRateLimited, OwnKeyBadReply,
	OwnKeyRemoved, OwnKeyDailyLimit, OwnKeyCallLimit, OwnKeyOther,
}

var (
	tablesStarted = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_tables_started_total",
		Help: "Tables that left the lobby, by kind: real, or practice (which cmdctrl_games_started_total leaves out).",
	}, []string{"kind"})

	ownKeySeatsAdded = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cmdctrl_own_key_seats_added_total",
		Help: "Bot seats added that run on their owner's own model key (ADR 0182).",
	})

	ownKeyFallbacks = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_own_key_fallbacks_total",
		Help: "Own-key bot seats that stopped using their model key and played the built-in heuristic, by reason.",
	}, []string{"reason"})

	statsPageLoads = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cmdctrl_stats_page_loads_total",
		Help: "GET /me/stats requests answered: how often the statistics page is opened.",
	})

	rematchesOffered = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cmdctrl_rematches_offered_total",
		Help: "Rematches offered at a finished table.",
	})
)

func init() {
	for _, k := range tableKindLabels {
		tablesStarted.WithLabelValues(k)
	}
	for _, r := range ownKeyReasonLabels {
		ownKeyFallbacks.WithLabelValues(r)
	}
}

// TableStarted counts a table leaving the lobby, practice or not.
func TableStarted(practice bool) {
	if practice {
		tablesStarted.WithLabelValues(TablePractice).Inc()
		return
	}
	tablesStarted.WithLabelValues(TableReal).Inc()
}

// OwnKeySeatAdded counts one own-key bot seat added to a table.
func OwnKeySeatAdded() { ownKeySeatsAdded.Inc() }

// OwnKeyFallback counts an own-key seat dropping to the heuristic for
// reason (one of the OwnKey* codes; anything else counts as other).
func OwnKeyFallback(reason string) {
	if !slices.Contains(ownKeyReasonLabels, reason) {
		reason = OwnKeyOther
	}
	ownKeyFallbacks.WithLabelValues(reason).Inc()
}

// StatsPageLoaded counts one answered GET /me/stats.
func StatsPageLoaded() { statsPageLoads.Inc() }

// RematchOffered counts one rematch offered.
func RematchOffered() { rematchesOffered.Inc() }
