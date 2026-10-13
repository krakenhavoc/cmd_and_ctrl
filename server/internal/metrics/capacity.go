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
