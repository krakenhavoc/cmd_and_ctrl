package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// The dashboards and alert rules in deploy/monitoring/ (ADR 0123 §7,
// §9, §11) query the server's metrics by name. Nothing else ties them
// to the code: a renamed metric would leave a panel empty and an alert
// silent, and promtool cannot tell, because a query for a name nobody
// reports is valid PromQL. These tests are that tie. They run with the
// server's tests, which CI runs on any change outside client/, and so
// on every change to deploy/monitoring/ and internal/metrics alike.

// monitoringDir is deploy/monitoring/, from this package's directory.
const monitoringDir = "../../../deploy/monitoring"

// textfileMetrics are cmdctrl_ names the server does not register:
// files other jobs write for Alloy's textfile collector.
var textfileMetrics = map[string]string{
	"cmdctrl_offsite_backup_last_success_timestamp_seconds":  "scripts/backup-offsite.sh",
	"cmdctrl_offsite_backup_last_exit_code":                  "scripts/backup-offsite.sh",
	"cmdctrl_monitoring_sync_last_success_timestamp_seconds": "the monitoring VM's config sync (HomeLab)",
	"cmdctrl_monitoring_sync_last_run_timestamp_seconds":     "the monitoring VM's config sync (HomeLab)",
	"cmdctrl_monitoring_sync_last_run_success":               "the monitoring VM's config sync (HomeLab)",
}

// datasourceUIDs are the monitoring VM's Grafana datasources.
var datasourceUIDs = map[string]bool{"prometheus": true, "loki": true, "alertmanager": true}

var cmdctrlName = regexp.MustCompile(`cmdctrl_[a-z0-9_]+`)

// monitoringFiles lists the files the monitoring VM syncs: the rule
// files and the dashboards. The rule tests are CI-only and not read.
func monitoringFiles(t *testing.T) (rules, dashboards []string) {
	t.Helper()
	var err error
	if rules, err = filepath.Glob(filepath.Join(monitoringDir, "rules", "*.yml")); err != nil {
		t.Fatal(err)
	}
	if dashboards, err = filepath.Glob(filepath.Join(monitoringDir, "dashboards", "*.json")); err != nil {
		t.Fatal(err)
	}
	if len(rules) == 0 || len(dashboards) == 0 {
		t.Fatalf("found %d rule files and %d dashboards under %s; want both", len(rules), len(dashboards), monitoringDir)
	}
	return rules, dashboards
}

// serverMetricNames is every metric family the server can report: the
// registry's own collectors and every state collector main registers,
// read with Describe so a vector nothing has incremented yet is named.
func serverMetricNames(t *testing.T) map[string]bool {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	dir := t.TempDir()
	database, err := db.Open(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	mgr := ws.NewRoomManager(log, dir)
	rec := &recordingRegisterer{}
	// With a database, so the users and database collectors register too.
	registerGameMetrics(rec, log, mgr, lobby.NewLobbyWithStore(mgr, lobby.NewSQLStore(database)), ws.NewHub(log),
		database, users.NewSQLStore(database, nil))
	if len(rec.collectors) == 0 {
		t.Fatal("registerGameMetrics registered nothing")
	}
	return metricstest.Names(t, append(metrics.Collectors(), rec.collectors...)...)
}

// recordingRegisterer keeps what is registered on it, so the test can
// describe exactly the collectors main registers.
type recordingRegisterer struct{ collectors []prometheus.Collector }

func (r *recordingRegisterer) Register(c prometheus.Collector) error {
	r.collectors = append(r.collectors, c)
	return nil
}

func (r *recordingRegisterer) MustRegister(cs ...prometheus.Collector) {
	r.collectors = append(r.collectors, cs...)
}

func (r *recordingRegisterer) Unregister(prometheus.Collector) bool { return false }

// known reports whether a name used in a query is a metric the server
// registers or a textfile metric: as written, or as a histogram's or
// summary's _bucket, _sum or _count series.
func known(registered map[string]bool, name string) bool {
	if registered[name] || textfileMetrics[name] != "" {
		return true
	}
	for _, suffix := range []string{"_bucket", "_sum", "_count"} {
		if base, ok := strings.CutSuffix(name, suffix); ok && registered[base] {
			return true
		}
	}
	return false
}

// ADR 0123 §11: every cmdctrl_ name a dashboard or a rule uses is one
// the server registers, or one of the textfile metrics.
func TestMonitoringConfigNamesRegisteredMetrics(t *testing.T) {
	registered := serverMetricNames(t)
	for name := range textfileMetrics {
		if registered[name] {
			t.Errorf("%s is registered by the server; drop it from textfileMetrics", name)
		}
	}
	rules, dashboards := monitoringFiles(t)
	used := 0
	for _, path := range append(rules, dashboards...) {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		for _, name := range cmdctrlName.FindAllString(string(raw), -1) {
			if seen[name] {
				continue
			}
			seen[name] = true
			used++
			if !known(registered, name) {
				t.Errorf("%s uses %s, which the server does not register (and is not a textfile metric)", filepath.Base(path), name)
			}
		}
	}
	if used == 0 {
		t.Error("no cmdctrl_ names found in deploy/monitoring/: is the pattern still right?")
	}
}

// ADR 0123 §9 and the monitoring VM's provisioning: each dashboard is
// JSON with a unique cmdctrl- uid and title, an env variable, and
// datasources named only by the VM's three uids. A dashboard exported
// for sharing names its datasource with a ${DS_…} input instead, which
// file provisioning cannot resolve.
func TestMonitoringDashboardsAreProvisionable(t *testing.T) {
	_, dashboards := monitoringFiles(t)
	uids := map[string]string{}
	titles := map[string]string{}
	for _, path := range dashboards {
		base := filepath.Base(path)
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var d map[string]any
		if err := json.Unmarshal(raw, &d); err != nil {
			t.Errorf("%s: not JSON: %v", base, err)
			continue
		}
		if strings.Contains(string(raw), "${DS_") || d["__inputs"] != nil {
			t.Errorf("%s: names a datasource through an import input (${DS_…} or __inputs); use the uid", base)
		}
		uid, _ := d["uid"].(string)
		if !strings.HasPrefix(uid, "cmdctrl-") {
			t.Errorf("%s: uid %q does not start with cmdctrl-", base, uid)
		}
		if other, dup := uids[uid]; dup {
			t.Errorf("%s and %s share uid %q", other, base, uid)
		}
		uids[uid] = base
		title, _ := d["title"].(string)
		if title == "" {
			t.Errorf("%s: no title", base)
		}
		if other, dup := titles[title]; dup {
			t.Errorf("%s and %s share title %q", other, base, title)
		}
		titles[title] = base
		if !hasEnvVariable(d) {
			t.Errorf("%s: no env template variable", base)
		}
		for _, bad := range datasourceRefs(d) {
			t.Errorf("%s: datasource %s is not one of the monitoring VM's (prometheus, loki, alertmanager)", base, bad)
		}
	}
}

func hasEnvVariable(d map[string]any) bool {
	templating, _ := d["templating"].(map[string]any)
	list, _ := templating["list"].([]any)
	for _, v := range list {
		if m, ok := v.(map[string]any); ok && m["name"] == "env" {
			return true
		}
	}
	return false
}

// datasourceRefs walks a dashboard and returns every datasource
// reference that does not name one of datasourceUIDs by uid.
func datasourceRefs(v any) []string {
	var bad []string
	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, child := range x {
				if k == "datasource" {
					ref, ok := child.(map[string]any)
					uid, _ := ref["uid"].(string)
					if !ok || !datasourceUIDs[uid] {
						s, _ := json.Marshal(child)
						bad = append(bad, string(s))
					}
					continue
				}
				walk(child)
			}
		case []any:
			for _, child := range x {
				walk(child)
			}
		}
	}
	walk(v)
	sort.Strings(bad)
	return bad
}
