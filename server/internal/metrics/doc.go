// Package metrics is the server's Prometheus instrumentation (ADR 0123
// §2, §3). It imports only github.com/prometheus/client_golang, so any
// package may import it: ws, lobby, aiseat and db all will. The aiseat
// import gate bans only internal/game, and this package never imports
// anything from this module.
//
// # The registry
//
// Everything registers on Registry, never on prometheus.DefaultRegisterer.
// NewRegistry builds a fresh one with the same contents, so a test can
// gather without sharing a registry with the process. Registry carries:
//
//   - cmdctrl_build_info{commit}, the binary's vcs.revision ("unknown"
//     when the build has none);
//   - the Go runtime and process collectors (go_*, process_*);
//   - every event metric in eventMetrics.
//
// The state collectors are not in it: they read live objects, so
// main.go registers them on Registry once it has built them, all in
// registerGameMetrics (cmd/server/metrics_state.go). Today that is
// engine.go's restore-point collector (the rooms), tables.go's (the
// lobby and the hub), and, only when there is a database, users.go's
// and db.go's. Collectors lists everything else.
//
// The server serves Registry on its own loopback-only listener
// (CMDCTRL_METRICS_ADDR, listen.go), never on the public mux.
//
// # Adding a metric
//
// An EVENT metric (a counter or a histogram something increments when
// it happens) is a package-level variable in this package, in a file
// named for its area (games.go, ws.go, engine.go, bots.go), with a
// small exported function the caller uses to record the event, so the
// label values are spelled here and nowhere else. Add the variable to
// eventMetrics in registry.go; NewRegistry registers that list.
//
// A STATE gauge (rooms, seats, sockets, users) is a prometheus.Collector
// that reads the live state at scrape time, so it can never drift from
// the truth. Its type lives here and takes a small interface for the
// state it reads, because this package may not import lobby or ws
// (they import it). main.go constructs it with the live RoomManager,
// Hub, Lobby or database and calls Registry.MustRegister.
//
// The dashboards and alert rules in deploy/monitoring/ query these
// names. TestMonitoringConfigNamesRegisteredMetrics (cmd/server) fails
// on a name they use that the server does not register, so a rename
// here updates them in the same change (docs/monitoring.md).
//
// # Labels are closed sets
//
// No metric carries a user, player, seat or game ID, a name, a remote
// address, a URL or free text (ADR 0123 §3, after ADR 0017 §9). Every
// label name a cmdctrl_ metric uses must be a row in labelSets
// (labels.go), with the closed set of values it may take. A new label
// is a new row there, in the same change as the metric.
// TestMetricLabelsAreClosedSets gathers the registry and fails on a
// label name outside that table, a value outside its row's set, or a
// metric or label name containing id, user, name, ip, remote or url.
// A label name that two families use with different closed sets (an
// outcome, a direction, a type) gets a row per family in
// familyLabelSets instead, which the guard consults first.
// A metric name the ADR fixes that trips the substring rule (ADR 0123
// names cmdctrl_users, cmdctrl_users_played and
// cmdctrl_users_created_total) goes in nameExemptions with its reason;
// a label name never does.
//
// # Routes
//
// cmdctrl_http_requests_total and cmdctrl_http_request_seconds carry the
// matched ServeMux pattern as `route`, never the path. Every mux the
// server builds is a *ServeMux from this package, which records its
// patterns as they are registered (the closed set the label guard
// checks against) and, after it dispatches, records the pattern it
// matched in the request's route holder. InstrumentHTTP puts the holder
// in the context; the innermost mux's answer wins. See http.go.
package metrics
