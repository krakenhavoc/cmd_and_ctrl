package metrics

import (
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// engine.go is ADR 0123 §3's "Game engine health" table, less the two
// database gauges: every commit through a room, the effect errors the
// engine survives, how far a deploy would rewind the live tables, and
// what the last boot's restore pass brought back.

// Fixed values of the type label for a commit that is not one action
// from the actions enum.
const (
	// ActionTypeOther is an action type outside the enum (a client can
	// send any string), or a commit whose caller named no type.
	ActionTypeOther = "other"
	// ActionTypeBundle is a ws.Room.ApplyBundle commit: several
	// mutations that commit or roll back together (a bot's
	// improvisation, a managed spawn). It counts once.
	ActionTypeBundle = "bundle"
	// ActionTypeLobby is a lobby-side setup commit through
	// ws.Room.ApplyExternal: a join, a deck, the start, a bot seat.
	ActionTypeLobby = "lobby"
)

// Values of the seat_kind label.
const (
	SeatKindHuman = "human"
	SeatKindBot   = "bot"
	SeatKindAgent = "agent"
	SeatKindAdmin = "admin"
)

// Values of the result label on the action metrics.
const (
	ResultApplied  = "applied"
	ResultRejected = "rejected"
)

// Values of the outcome label on cmdctrl_boot_restore_games.
const (
	RestoreOutcomeRestored  = "restored"
	RestoreOutcomeEnded     = "ended"
	RestoreOutcomeAbandoned = "abandoned"
)

var (
	actionSeatKindLabels = []string{SeatKindHuman, SeatKindBot, SeatKindAgent, SeatKindAdmin}
	actionResultLabels   = []string{ResultApplied, ResultRejected}
	restoreOutcomeLabels = []string{RestoreOutcomeRestored, RestoreOutcomeEnded, RestoreOutcomeAbandoned}
	fixedActionTypes     = []string{ActionTypeOther, ActionTypeBundle, ActionTypeLobby}
)

var (
	actionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_actions_total",
		Help: "Commits through a room (Apply, ApplyBundle, ApplyExternal), by action type, the kind of seat that made them, and whether they applied.",
	}, []string{"type", "seat_kind", "result"})

	actionApplySeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "cmdctrl_action_apply_seconds",
		Help: "Time a commit held the room lock: dispatch, state capture, dump and restore-point write.",
		// An ordinary action is a few milliseconds; a restore-point
		// write on a big board, or a slow disk, is what the upper
		// buckets are for.
		Buckets: []float64{.0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
	}, []string{"result"})

	effectErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cmdctrl_effect_errors_total",
		Help: "EventEffectErrors emitted: a card's effect failed mid-resolution and the engine carried on.",
	})

	bootRestoreGames = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cmdctrl_boot_restore_games",
		Help: "The last boot's restore pass: restore points that came back, that were for ended games past their retention and dropped, and that were abandoned.",
	}, []string{"outcome"})

	bootRestoreDegradedCards = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cmdctrl_boot_restore_degraded_cards",
		Help: "Cards the last boot's restore pass flagged AbilitiesLostOnRestore (fewer catalog abilities than captured, or a lost stack ability).",
	})
)

// The two children of the apply histogram, fetched once so a commit
// does no label lookup for them.
var (
	applySecondsApplied  = actionApplySeconds.WithLabelValues(ResultApplied)
	applySecondsRejected = actionApplySeconds.WithLabelValues(ResultRejected)
)

// actionTypes is the actions enum, registered by the package that owns
// it (ws registers actions.Types at init): the closed set of the type
// label, less fixedActionTypes. This package imports nothing from the
// module, so it cannot read the enum itself.
var actionTypes = &closedSet{set: map[string]bool{}}

// RegisterActionTypes adds action types to the type label's closed set.
// A type never registered is recorded as "other".
func RegisterActionTypes(types ...string) {
	for _, t := range types {
		actionTypes.add(t)
	}
}

// ActionTypes returns every registered action type, sorted.
func ActionTypes() []string { return actionTypes.list() }

func actionTypeLabel(t string) string {
	switch t {
	case ActionTypeBundle, ActionTypeLobby:
		return t
	}
	if actionTypes.has(t) {
		return t
	}
	return ActionTypeOther
}

func seatKindLabel(k string) string {
	switch k {
	case SeatKindHuman, SeatKindBot, SeatKindAgent, SeatKindAdmin:
		return k
	}
	return SeatKindHuman
}

// actionKey is one cmdctrl_actions_total series.
type actionKey struct {
	typ, kind string
	applied   bool
}

// actionCounters caches each series' child, so a commit after the
// first of its kind does one read-locked map lookup and no allocation.
var actionCounters = struct {
	mu sync.RWMutex
	m  map[actionKey]prometheus.Counter
}{m: map[actionKey]prometheus.Counter{}}

func actionCounter(k actionKey) prometheus.Counter {
	actionCounters.mu.RLock()
	c, ok := actionCounters.m[k]
	actionCounters.mu.RUnlock()
	if ok {
		return c
	}
	result := ResultRejected
	if k.applied {
		result = ResultApplied
	}
	c = actionsTotal.WithLabelValues(k.typ, k.kind, result)
	actionCounters.mu.Lock()
	actionCounters.m[k] = c
	actionCounters.mu.Unlock()
	return c
}

// RecordAction counts one commit through a room and observes how long
// it held the room lock. typ is an action type (anything unregistered
// counts as "other") or one of the fixed ActionType values; seatKind is
// one of the SeatKind values. applied is false when the commit was
// refused and rolled back.
func RecordAction(typ, seatKind string, applied bool, held time.Duration) {
	actionCounter(actionKey{typ: actionTypeLabel(typ), kind: seatKindLabel(seatKind), applied: applied}).Inc()
	if applied {
		applySecondsApplied.Observe(held.Seconds())
	} else {
		applySecondsRejected.Observe(held.Seconds())
	}
}

// EffectError counts one EventEffectError.
func EffectError() { effectErrors.Inc() }

// SetBootRestore sets the boot restore gauges from the restore pass.
// Called once at boot, after the pass, with zeros when there was
// nothing to restore.
func SetBootRestore(restored, ended, abandoned, degradedCards int) {
	bootRestoreGames.WithLabelValues(RestoreOutcomeRestored).Set(float64(restored))
	bootRestoreGames.WithLabelValues(RestoreOutcomeEnded).Set(float64(ended))
	bootRestoreGames.WithLabelValues(RestoreOutcomeAbandoned).Set(float64(abandoned))
	bootRestoreDegradedCards.Set(float64(degradedCards))
}

// RestorePointLag is what the restore-point collector reads at scrape
// time (see RestorePointReader).
type RestorePointLag struct {
	// Oldest is the age of the oldest restore point among the active
	// rooms that are behind theirs: how much a deploy now would
	// rewind. Zero when no room is behind.
	Oldest time.Duration
	// Behind is how many active rooms have applied a commit their
	// restore point does not hold.
	Behind int
}

// RestorePointReader is the live state the restore-point collector
// reads: ws.RoomManager.
type RestorePointReader interface {
	RestorePointLag(now time.Time) RestorePointLag
}

// restorePointCollector serves cmdctrl_restore_point_age_seconds and
// cmdctrl_rooms_behind_restore_point from the live rooms at scrape time.
type restorePointCollector struct {
	src  RestorePointReader
	age  *prometheus.Desc
	rows *prometheus.Desc
	now  func() time.Time
}

// NewRestorePointCollector returns the collector for the two
// restore-point gauges over src. main.go registers it on Registry with
// the live RoomManager.
func NewRestorePointCollector(src RestorePointReader) prometheus.Collector {
	return &restorePointCollector{
		src: src,
		age: prometheus.NewDesc("cmdctrl_restore_point_age_seconds",
			"Age of the oldest restore point among active rooms that have moved past theirs: how much a deploy now would rewind (ADR 0044). Zero when every active room is at its restore point.",
			nil, nil),
		rows: prometheus.NewDesc("cmdctrl_rooms_behind_restore_point",
			"Active rooms whose last applied commit is not yet captured in a restore point.",
			nil, nil),
		now: time.Now,
	}
}

func (c *restorePointCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.age
	ch <- c.rows
}

func (c *restorePointCollector) Collect(ch chan<- prometheus.Metric) {
	lag := c.src.RestorePointLag(c.now())
	ch <- prometheus.MustNewConstMetric(c.age, prometheus.GaugeValue, lag.Oldest.Seconds())
	ch <- prometheus.MustNewConstMetric(c.rows, prometheus.GaugeValue, float64(lag.Behind))
}

// closedSet is a label's closed set that code adds to at init or
// registration time, like the route set.
type closedSet struct {
	mu  sync.RWMutex
	set map[string]bool
}

func (s *closedSet) add(v string) {
	s.mu.Lock()
	s.set[v] = true
	s.mu.Unlock()
}

func (s *closedSet) has(v string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set[v]
}

func (s *closedSet) list() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.set))
	for v := range s.set {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

// actionTypeSet is the type label's row: every registered action type
// and the fixed values.
func actionTypeSet() labelSet {
	return labelSet{
		values: "an action type registered with RegisterActionTypes, or " + strings.Join(fixedActionTypes, "|"),
		allows: func(v string) bool { return slices.Contains(fixedActionTypes, v) || actionTypes.has(v) },
	}
}
