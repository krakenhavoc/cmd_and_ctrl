package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// withClock pins the boot clock to start+d for the test.
func withClock(t *testing.T, d time.Duration) {
	t.Helper()
	prev := now
	now = func() time.Time { return processStart.Add(d) }
	t.Cleanup(func() { now = prev })
}

func histCount(t *testing.T, name string) uint64 {
	t.Helper()
	fams, err := Registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fams {
		if f.GetName() == name {
			return f.GetMetric()[0].GetHistogram().GetSampleCount()
		}
	}
	t.Fatalf("%s is not registered", name)
	return 0
}

func histSum(t *testing.T, name string) float64 {
	t.Helper()
	fams, err := Registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fams {
		if f.GetName() == name {
			return f.GetMetric()[0].GetHistogram().GetSampleSum()
		}
	}
	t.Fatalf("%s is not registered", name)
	return 0
}

// Ready is the time from the boot clock's zero, and the phases are
// what the boot reports for them; a phase nobody named is dropped
// rather than minted as a label.
func TestBootReadyAndPhases(t *testing.T) {
	withClock(t, 12500*time.Millisecond)
	if got := BootReady(); got != 12500*time.Millisecond {
		t.Errorf("BootReady returned %v, want 12.5s", got)
	}
	if got := testutil.ToFloat64(bootReadySeconds); got != 12.5 {
		t.Errorf("cmdctrl_boot_ready_seconds = %v, want 12.5", got)
	}

	BootPhase(BootPhaseCardIndex, 4*time.Second)
	BootPhase(BootPhaseRestore, 1500*time.Millisecond)
	BootPhase("not-a-phase", time.Hour)
	if got := testutil.ToFloat64(bootPhaseSeconds.WithLabelValues(BootPhaseCardIndex)); got != 4 {
		t.Errorf("card_index = %v, want 4", got)
	}
	if got := testutil.ToFloat64(bootPhaseSeconds.WithLabelValues(BootPhaseRestore)); got != 1.5 {
		t.Errorf("restore = %v, want 1.5", got)
	}
	if n := testutil.CollectAndCount(bootPhaseSeconds); n != len(bootPhaseLabels) {
		t.Errorf("%d phase series, want the closed set of %d", n, len(bootPhaseLabels))
	}
}

func TestBootFirstFrameObservesSinceStart(t *testing.T) {
	withClock(t, 30*time.Second)
	n, sum := histCount(t, "cmdctrl_boot_first_frame_seconds"), histSum(t, "cmdctrl_boot_first_frame_seconds")
	if got := BootFirstFrame(); got != 30*time.Second {
		t.Errorf("BootFirstFrame returned %v, want 30s", got)
	}
	if got := histCount(t, "cmdctrl_boot_first_frame_seconds") - n; got != 1 {
		t.Errorf("observations rose by %d, want 1", got)
	}
	if got := histSum(t, "cmdctrl_boot_first_frame_seconds") - sum; got != 30 {
		t.Errorf("sum rose by %v, want 30", got)
	}
}

// A table that came back exactly where it stood is an observation of
// zero, so the le=0 bucket counts the tables that did not rewind; a
// negative span (clock skew between the two processes) is zero too.
func TestBootRewindObservesZerosAndClampsSpan(t *testing.T) {
	n, sum := histCount(t, "cmdctrl_boot_rewind_actions"), histSum(t, "cmdctrl_boot_rewind_actions")
	secs := histSum(t, "cmdctrl_boot_rewind_seconds")

	BootRewind(0, 0)
	BootRewind(7, 90*time.Second)
	BootRewind(1, -5*time.Second)

	if got := histCount(t, "cmdctrl_boot_rewind_actions") - n; got != 3 {
		t.Errorf("observations rose by %d, want 3", got)
	}
	if got := histSum(t, "cmdctrl_boot_rewind_actions") - sum; got != 8 {
		t.Errorf("actions rewound rose by %v, want 8", got)
	}
	if got := histSum(t, "cmdctrl_boot_rewind_seconds") - secs; got != 90 {
		t.Errorf("seconds rewound rose by %v, want 90", got)
	}
}

// The previous shutdown's gauges take what the record says, and a
// negative downtime (clock skew) reads as zero.
func TestPreviousShutdown(t *testing.T) {
	SetPreviousShutdown(PreviousShutdown{
		Downtime: 8 * time.Second,
		Flush:    map[string]int{FlushWritten: 3, FlushNothingPending: 5, FlushFailed: 1, "not-a-flush": 9},
		Rewound:  2,
	})
	if got := testutil.ToFloat64(bootDowntimeSeconds); got != 8 {
		t.Errorf("downtime = %v, want 8", got)
	}
	for f, want := range map[string]float64{FlushWritten: 3, FlushNothingPending: 5, FlushFailed: 1, FlushNoRestorePoint: 0} {
		if got := testutil.ToFloat64(bootPreviousShutdownGames.WithLabelValues(f)); got != want {
			t.Errorf("flush=%s = %v, want %v", f, got, want)
		}
	}
	if n := testutil.CollectAndCount(bootPreviousShutdownGames); n != len(flushLabels) {
		t.Errorf("%d flush series, want the closed set of %d", n, len(flushLabels))
	}
	if got := testutil.ToFloat64(bootRewoundGames); got != 2 {
		t.Errorf("rewound = %v, want 2", got)
	}
	SetPreviousShutdown(PreviousShutdown{Downtime: -time.Second})
	if got := testutil.ToFloat64(bootDowntimeSeconds); got != 0 {
		t.Errorf("a negative downtime reads %v, want 0", got)
	}
}

// Every deploy metric, fed values outside its sets, still passes the
// label guard.
func TestDeployMetricsPassTheLabelGuard(t *testing.T) {
	BootReady()
	for _, p := range append(append([]string{}, bootPhaseLabels...), "bogus") {
		BootPhase(p, time.Second)
	}
	BootFirstFrame()
	BootRewind(3, time.Minute)
	SetPreviousShutdown(PreviousShutdown{Flush: map[string]int{FlushWritten: 1, "bogus": 1}})
	if err := CheckClosedLabels(NewRegistry()); err != nil {
		t.Fatal(err)
	}
}
