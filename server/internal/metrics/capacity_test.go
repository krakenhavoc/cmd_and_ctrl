package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeCapacity CapacityReading

func (f fakeCapacity) CapacityReading() CapacityReading { return CapacityReading(f) }

func TestCapacityCollectorReportsLimitAndUsePerResource(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewCapacityCollector(fakeCapacity{Load: 42, LoadLimit: 180, Conns: 9, ConnLimit: 0}))
	err := testutil.GatherAndCompare(reg, strings.NewReader(`
# HELP cmdctrl_capacity_limit The capacity cap for a resource. 0 means the cap is off.
# TYPE cmdctrl_capacity_limit gauge
cmdctrl_capacity_limit{resource="table_load"} 180
cmdctrl_capacity_limit{resource="ws_connections"} 0
# HELP cmdctrl_capacity_used What the server carries now against the cap for a resource: table_load is the weighted load of the unended tables, ws_connections the live sockets.
# TYPE cmdctrl_capacity_used gauge
cmdctrl_capacity_used{resource="table_load"} 42
cmdctrl_capacity_used{resource="ws_connections"} 9
`), "cmdctrl_capacity_limit", "cmdctrl_capacity_used")
	if err != nil {
		t.Error(err)
	}
}

func TestCapacityRefusalsAreCountedByReason(t *testing.T) {
	for _, reason := range refusalReasons {
		before := testutil.ToFloat64(capacityRefusals.WithLabelValues(reason))
		CapacityRefused(reason)
		if got := testutil.ToFloat64(capacityRefusals.WithLabelValues(reason)); got != before+1 {
			t.Errorf("%s: %v, want %v", reason, got, before+1)
		}
	}
}

func TestPlayerReportedCountsOneStoredReport(t *testing.T) {
	before := testutil.ToFloat64(playerReports)
	PlayerReported()
	if got := testutil.ToFloat64(playerReports); got != before+1 {
		t.Errorf("cmdctrl_player_reports_total = %v, want %v", got, before+1)
	}
}

func TestBootRestoreDeferredGauge(t *testing.T) {
	SetBootRestoreDeferred(7)
	if got := testutil.ToFloat64(bootRestoreDeferred); got != 7 {
		t.Errorf("gauge = %v, want 7", got)
	}
	SetBootRestoreDeferred(0)
	if got := testutil.ToFloat64(bootRestoreDeferred); got != 0 {
		t.Errorf("gauge = %v, want 0", got)
	}
}
