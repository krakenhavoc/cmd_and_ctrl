package metrics

import (
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// labelSet is one row of the label table: a label name's closed set of
// values, as a check.
type labelSet struct {
	// values describes the set for an error message.
	values string
	allows func(string) bool
}

func oneOf(vs ...string) labelSet {
	return labelSet{
		values: strings.Join(vs, "|"),
		allows: func(v string) bool { return slices.Contains(vs, v) },
	}
}

// labelSets is THE table of label names a cmdctrl_ metric may carry,
// each with the closed set of values it may take (ADR 0123 §3). A new
// label is a new row here, in the same change as the metric that uses
// it. A set must be closed: an enum, a fixed list, or patterns the code
// registered. Never an ID, a name, an address, a URL or free text.
var labelSets = map[string]labelSet{
	// The matched ServeMux pattern (http.go), or "unmatched".
	"route": {
		values: "a pattern registered on a metrics.ServeMux, or " + RouteUnmatched,
		allows: func(v string) bool { return v == RouteUnmatched || routes.has(v) },
	},
	"method": oneOf(methodLabels...),
	"code":   oneOf(codeLabels...),
	// cmdctrl_build_info: one value per binary.
	"commit": {
		values: "this binary's commit",
		allows: func(v string) bool { return v == buildCommit },
	},
	// Games and players (games.go, tables.go, users.go).
	"state":    oneOf(tableStateLabels...),
	"archived": oneOf(archivedLabels...),
	"kind":     oneOf(seatKindLabels...),
	"account":  oneOf(accountLabels...),
	"seats":    oneOf(seatsLabels...),
	"window":   oneOf(windowLabels...),
	// WebSocket (ws.go, tables.go).
	"role":   oneOf(roleLabels...),
	"reason": oneOf(rejectLabels...),
}

// familyLabelSets are rows for one metric family only, consulted
// before labelSets. They are for a label name that two families use
// with different closed sets (direction is in|out on the frames and
// prompt|completion on the model tokens; outcome is a game's result
// here and a restore's elsewhere), so neither family's set has to be
// widened into the other's.
var familyLabelSets = map[string]map[string]labelSet{
	"cmdctrl_games_ended_total": {
		"outcome": oneOf(outcomeLabels...),
	},
	"cmdctrl_ws_frames_total": {
		"direction": oneOf(frameDirectionLabels...),
		"type":      oneOf(frameTypeLabels...),
	},
}

// standardLabelSets is the same table for the upstream Go and process
// collectors' families (go_*, process_*), which this package does not
// name.
var standardLabelSets = map[string]labelSet{
	// go_info{version}: the Go release this binary was built with.
	"version": {
		values: runtime.Version(),
		allows: func(v string) bool { return v == runtime.Version() },
	},
}

// bannedSubstrings may not appear in a metric or label name: each would
// be a metric about one person, one table or one address (ADR 0123 §3,
// ADR 0017 §9).
var bannedSubstrings = []string{"id", "user", "name", "ip", "remote", "url"}

// standardPrefixes are the upstream collectors' families.
var standardPrefixes = []string{"go_", "process_"}

// nameExemptions are metric names allowed despite containing a banned
// substring, each with its reason. The upstream families (go_*,
// process_*) are exempt by prefix instead, below: their names are fixed
// by client_golang, cannot carry data, and one of them trips the rule
// on an ordinary word (go_memstats_heap_idle_bytes, "id" in "idle").
// Their LABELS are still checked. A label name is never exempt.
var nameExemptions = map[string]string{
	// ADR 0123 §3 names these. Each counts accounts; none carries one.
	"cmdctrl_users":               "a count of users rows (ADR 0123 §3)",
	"cmdctrl_users_played":        "a count of accounts that played in a window (ADR 0123 §3)",
	"cmdctrl_users_created_total": "a count of new users rows (ADR 0123 §3)",
}

func isStandard(family string) bool {
	for _, p := range standardPrefixes {
		if strings.HasPrefix(family, p) {
			return true
		}
	}
	return false
}

func bannedIn(name string) string {
	for _, s := range bannedSubstrings {
		if strings.Contains(name, s) {
			return s
		}
	}
	return ""
}

// CheckClosedLabels gathers g and reports every family or series that
// breaks the label rule: a family that is neither cmdctrl_ nor an
// upstream collector's, a metric or label name containing a banned
// substring, a label name with no row in the table, or a value outside
// its row's set. It is what TestMetricLabelsAreClosedSets runs, and
// exported so the server's own tests can run it after driving real
// routes.
func CheckClosedLabels(g prometheus.Gatherer) error {
	families, err := g.Gather()
	if err != nil {
		return fmt.Errorf("gather: %w", err)
	}
	var errs []error
	for _, f := range families {
		name := f.GetName()
		standard := isStandard(name)
		sets := labelSets
		switch {
		case standard:
			sets = standardLabelSets
		case !strings.HasPrefix(name, "cmdctrl_"):
			errs = append(errs, fmt.Errorf("%s: not a cmdctrl_ metric nor an upstream go_/process_ one", name))
			continue
		}
		if _, exempt := nameExemptions[name]; !exempt && !standard {
			if s := bannedIn(name); s != "" {
				errs = append(errs, fmt.Errorf("%s: metric name contains %q", name, s))
			}
		}
		for _, m := range f.GetMetric() {
			for _, lp := range m.GetLabel() {
				ln, lv := lp.GetName(), lp.GetValue()
				if s := bannedIn(ln); s != "" {
					errs = append(errs, fmt.Errorf("%s: label name %q contains %q", name, ln, s))
					continue
				}
				set, ok := familyLabelSets[name][ln]
				if !ok {
					set, ok = sets[ln]
				}
				if !ok {
					errs = append(errs, fmt.Errorf("%s: label %q is not in the label table (labels.go)", name, ln))
					continue
				}
				if !set.allows(lv) {
					errs = append(errs, fmt.Errorf("%s: label %s=%q is outside its set (%s)", name, ln, lv, set.values))
				}
			}
		}
	}
	return errors.Join(errs...)
}
