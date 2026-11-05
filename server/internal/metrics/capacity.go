package metrics

import "github.com/prometheus/client_golang/prometheus"

// The capacity metrics (ADR 0148, #2977): the caps the server holds
// itself to, what it carries against them, how often a request was
// turned away, and how many tables the boot left on disk.
//
//	cmdctrl_capacity_limit{resource}       the cap; 0 means the cap is off
//	cmdctrl_capacity_used{resource}        what is carried now
//	cmdctrl_capacity_refusals_total{reason}  requests turned away
//	cmdctrl_boot_restore_deferred_games    tables not brought back because the load
//	                                       cap was full: set at boot, and by the hot
//	                                       restore to those still waiting
//
// And the idle cleanup's (ADR 0211 §6), which keeps abandoned tables
// from holding the cap:
//
//	cmdctrl_capacity_idle_after_seconds{kind}  the configured threshold; 0 is off
//	cmdctrl_capacity_idle_closed_total{kind}   tables the sweep closed
//	cmdctrl_capacity_idle_warned_total         started games warned before a close
//	cmdctrl_capacity_idle_unloaded_total       closed tables unloaded after the keep window
//
// The two gauges are a collector read at scrape time, like the tables
// collector; the counter and the boot gauge are event metrics.

// Resources of cmdctrl_capacity_limit and cmdctrl_capacity_used.
const (
	// CapacityTableLoad is the weighted load of the unended tables
	// (capacity.Limits.TableCost summed).
	CapacityTableLoad = "table_load"
	// CapacityWSConnections is the live WebSocket connections.
	CapacityWSConnections = "ws_connections"
)

var capacityResources = []string{CapacityTableLoad, CapacityWSConnections}

// Reasons of cmdctrl_capacity_refusals_total.
const (
	// RefuseTableLoad: a table (or a practice table) would have pushed
	// the load past the cap.
	RefuseTableLoad = "table_load"
	// RefuseBotSeat: a bot seat would have pushed the load past it.
	RefuseBotSeat = "bot_seat"
	// RefuseWSConnections: a WebSocket from someone with no seat at the
	// table was turned away at the connection cap.
	RefuseWSConnections = "ws_connections"
	// RefuseJoinConnections: a join or spectate was turned away at the
	// connection cap, before the socket it would have opened.
	RefuseJoinConnections = "join_connections"
	// RefuseCreateAccount / RefuseCreateIP: the per-account and per-IP
	// table-creation rates.
	RefuseCreateAccount = "create_account"
	RefuseCreateIP      = "create_ip"
)

var refusalReasons = []string{
	RefuseTableLoad, RefuseBotSeat, RefuseWSConnections, RefuseJoinConnections,
	RefuseCreateAccount, RefuseCreateIP,
}

// Kinds of the idle cleanup's metrics (ADR 0211 §6): an unstarted
// table, a started game nobody moved in, and a game left to its bots.
const (
	IdleKindLobby = "lobby"
	IdleKindGame  = "game"
	IdleKindBots  = "bots"
)

var idleKindLabels = []string{IdleKindLobby, IdleKindGame, IdleKindBots}

var (
	idleAfterSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cmdctrl_capacity_idle_after_seconds",
		Help: "The idle cleanup's threshold, by kind: how long an unstarted table (lobby), a started game with no human move (game) or a game left to its bots (bots) may stay idle before the server closes it. 0 means that cleanup is off.",
	}, []string{"kind"})
	idleClosed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_capacity_idle_closed_total",
		Help: "Tables the idle sweep closed, by kind.",
	}, []string{"kind"})
	idleWarned = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cmdctrl_capacity_idle_warned_total",
		Help: "Started games the idle sweep warned before closing them.",
	})
	idleUnloaded = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cmdctrl_capacity_idle_unloaded_total",
		Help: "Idle-closed tables unloaded from memory and disk after the keep window (the replay is kept).",
	})
)

// SetIdleAfter records the idle cleanup's configured thresholds; a
// zero duration is that cleanup off. Called once at boot.
func SetIdleAfter(lobby, game, bots float64) {
	idleAfterSeconds.WithLabelValues(IdleKindLobby).Set(lobby)
	idleAfterSeconds.WithLabelValues(IdleKindGame).Set(game)
	idleAfterSeconds.WithLabelValues(IdleKindBots).Set(bots)
}

// IdleClosed counts one table the idle sweep closed, by kind.
func IdleClosed(kind string) { idleClosed.WithLabelValues(kind).Inc() }

// IdleWarned counts one started game warned before an idle close.
func IdleWarned() { idleWarned.Inc() }

// IdleUnloaded counts one idle-closed table unloaded after its keep
// window.
func IdleUnloaded() { idleUnloaded.Inc() }

var (
	capacityRefusals = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_capacity_refusals_total",
		Help: "Requests turned away by a capacity cap or a table-creation rate, by reason.",
	}, []string{"reason"})

	bootRestoreDeferred = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cmdctrl_boot_restore_deferred_games",
		Help: "Unended tables left on disk, not restored, because the table-load cap was full: set at boot, and again by POST /admin/capacity/restore-deferred to those still waiting. Zero when everything fit.",
	})

	capacityLimitDesc = prometheus.NewDesc("cmdctrl_capacity_limit",
		"The capacity cap for a resource. 0 means the cap is off.",
		[]string{"resource"}, nil)
	capacityUsedDesc = prometheus.NewDesc("cmdctrl_capacity_used",
		"What the server carries now against the cap for a resource: table_load is the weighted load of the unended tables, ws_connections the live sockets.",
		[]string{"resource"}, nil)
)

func init() {
	for _, r := range refusalReasons {
		capacityRefusals.WithLabelValues(r)
	}
	for _, k := range idleKindLabels {
		idleAfterSeconds.WithLabelValues(k)
		idleClosed.WithLabelValues(k)
	}
}

// CapacityRefused counts one request turned away, by reason.
func CapacityRefused(reason string) { capacityRefusals.WithLabelValues(reason).Inc() }

// SetBootRestoreDeferred sets the boot gauge of tables left on disk by
// the load cap. Called once at boot, with zero when nothing was.
func SetBootRestoreDeferred(n int) { bootRestoreDeferred.Set(float64(n)) }

// CapacityReading is what the capacity collector reads at scrape time.
// A zero limit means that cap is off.
type CapacityReading struct {
	Load      int
	LoadLimit int
	Conns     int
	ConnLimit int
}

// CapacitySource reads the current usage and limits. main.go
// implements it over the lobby (load) and the hub (connections).
type CapacitySource interface {
	CapacityReading() CapacityReading
}

type capacityCollector struct{ src CapacitySource }

// NewCapacityCollector returns the collector of cmdctrl_capacity_limit
// and cmdctrl_capacity_used. main.go registers it on Registry.
func NewCapacityCollector(src CapacitySource) prometheus.Collector {
	return &capacityCollector{src: src}
}

func (c *capacityCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- capacityLimitDesc
	ch <- capacityUsedDesc
}

func (c *capacityCollector) Collect(ch chan<- prometheus.Metric) {
	r := c.src.CapacityReading()
	limit := map[string]int{CapacityTableLoad: r.LoadLimit, CapacityWSConnections: r.ConnLimit}
	used := map[string]int{CapacityTableLoad: r.Load, CapacityWSConnections: r.Conns}
	for _, res := range capacityResources {
		ch <- prometheus.MustNewConstMetric(capacityLimitDesc, prometheus.GaugeValue, float64(limit[res]), res)
		ch <- prometheus.MustNewConstMetric(capacityUsedDesc, prometheus.GaugeValue, float64(used[res]), res)
	}
}
