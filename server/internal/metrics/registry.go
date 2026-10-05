package metrics

import (
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Registry is the server's one registry: the one the metrics listener
// serves and the one main.go registers its state collectors on.
var Registry = NewRegistry()

// eventMetrics is every package-level event metric. NewRegistry
// registers exactly this list, so a new counter or histogram is
// added here as well as declared (see the package doc).
func eventMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		httpRequests,
		httpSeconds,
		// games.go and ws.go.
		gamesCreated,
		gamesStarted,
		gamesEnded,
		gameDuration,
		usersCreated,
		wsConnects,
		wsDisconnects,
		wsRejections,
		wsFrames,
		wsBroadcast,
		// engine.go
		actionsTotal,
		actionApplySeconds,
		effectErrors,
		bootRestoreGames,
		bootRestoreDegradedCards,
		// bots.go
		botDecisions,
		botDecisionSeconds,
		botFallbacks,
		botModelCalls,
		botModelCallSeconds,
		botModelTokens,
	}
}

// NewRegistry returns a fresh registry carrying the build info, the Go
// and process collectors, and every event metric. The event metrics are
// package-level, so two registries report the same values; a test that
// counts must compare before and after rather than expect zero.
func NewRegistry() *prometheus.Registry {
	r := prometheus.NewRegistry()
	r.MustRegister(Collectors()...)
	return r
}

// Collectors is everything NewRegistry registers: the Go and process
// collectors, the build info and every event metric. The state
// collectors main.go adds are not in it. A test that needs every name
// the server can report (the dashboards' and rules' name check)
// describes these and main's state collectors together.
func Collectors() []prometheus.Collector {
	return append([]prometheus.Collector{
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		newBuildInfo(),
	}, eventMetrics()...)
}

// buildCommit is the binary's vcs.revision, read once.
var buildCommit = readCommit()

// BuildCommit is the commit cmdctrl_build_info reports: the
// vcs.revision Go stamped into the binary, or "unknown" when there is
// none (a build outside a git checkout, or go test).
func BuildCommit() string { return buildCommit }

func readCommit() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" && s.Value != "" {
			return s.Value
		}
	}
	return "unknown"
}

func newBuildInfo() prometheus.Collector {
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "cmdctrl_build_info",
		Help:        "Always 1. The commit label is the binary's vcs.revision, or unknown.",
		ConstLabels: prometheus.Labels{"commit": buildCommit},
	})
	g.Set(1)
	return g
}
