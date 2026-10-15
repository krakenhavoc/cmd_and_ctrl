package metrics

import (
	"math"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// The deploy-cost metrics (ADR 0195 phase 0, #3037): what a restart
// costs a live table. They are set once per boot and describe the
// deploy that produced THIS process, so a series is one deploy's
// reading and each new process is the next deploy. They answer the
// three questions the ADR's deadline and memory budget wait on:
//
//	cmdctrl_boot_ready_seconds             process start until the listener is bound
//	cmdctrl_boot_phase_seconds{phase}      the parts of that: database, card index, restore pass
//	cmdctrl_boot_downtime_seconds          the previous process's last word until this one started
//	cmdctrl_boot_first_frame_seconds       process start until a restored table sent its first frame
//	cmdctrl_boot_rewind_actions            per restored table, actions the restart threw away
//	cmdctrl_boot_rewind_seconds            per restored table, the time those actions spanned
//	cmdctrl_boot_rewound_games             restored tables that came back behind where they stood
//	cmdctrl_boot_previous_shutdown_games{flush}
//	                                       the previous process's tables by what its
//	                                       shutdown flush did with their restore point
//
// A process cannot be scraped after it has stopped, so the shutdown
// side is a record on disk (ws.ShutdownRecord) that the next boot reads
// and turns into the last four. The first boot after a build that
// writes no record (the one that introduces this) has none: the
// downtime gauge reads NaN and the rewind histograms stay empty. The
// shutdown census log lines still carry that deploy's rewind.
//
// RSS and the Go heap are the standard process_resident_memory_bytes
// and go_memstats_heap_inuse_bytes, which the Server dashboard already
// plots; the deploy row of the Engine dashboard sets them beside the
// table count.

// Values of the phase label on cmdctrl_boot_phase_seconds.
const (
	// BootPhaseDatabase is opening and migrating the SQLite database.
	BootPhaseDatabase = "database"
	// BootPhaseCardIndex is loading the Scryfall card index.
	BootPhaseCardIndex = "card_index"
	// BootPhaseRestore is the restore pass: reading every restore point
	// and rebuilding its table.
	BootPhaseRestore = "restore"
)

var bootPhaseLabels = []string{BootPhaseDatabase, BootPhaseCardIndex, BootPhaseRestore}

// Values of the flush label on cmdctrl_boot_previous_shutdown_games:
// what the previous process's shutdown flush (ws.Room.FlushRestorePoint)
// did for one table.
const (
	// FlushWritten: a restore point was pending and was written, so the
	// table's newest clean state is on disk.
	FlushWritten = "written"
	// FlushNothingPending: no restore point was pending, because the
	// last one had already been written (the common case) or the table
	// was holding a continuation and had no clean state to write.
	FlushNothingPending = "nothing_pending"
	// FlushFailed: the write was attempted and failed. The older file,
	// if any, is what the next boot restored.
	FlushFailed = "failed"
	// FlushNoRestorePoint: the table has no restore point on disk at all,
	// so the next boot could not bring it back.
	FlushNoRestorePoint = "no_restore_point"
)

var flushLabels = []string{FlushWritten, FlushNothingPending, FlushFailed, FlushNoRestorePoint}

var (
	bootReadySeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cmdctrl_boot_ready_seconds",
		Help: "Seconds from the process starting until its listener was bound and the server could serve (ADR 0195 phase 0, #3037). Zero until then.",
	})

	bootPhaseSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cmdctrl_boot_phase_seconds",
		Help: "Seconds the boot spent in one phase: opening the database, loading the card index, and the restore pass. The rest of cmdctrl_boot_ready_seconds is everything else.",
	}, []string{"phase"})

	bootDowntimeSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cmdctrl_boot_downtime_seconds",
		Help: "Seconds between the previous process's shutdown record and this process starting: the time nothing answered. NaN when the previous process left no record (a crash, or a build that wrote none). A large value is a stopped server, not a deploy.",
	})

	bootFirstFrameSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "cmdctrl_boot_first_frame_seconds",
		Help: "Per table restored at boot, seconds from the process starting until the first snapshot frame went to a client. The client's reconnect backoff is in it; a table nobody reconnects to is not observed.",
		// A fast boot is a few seconds and the ADR 0148 load test saw
		// seven minutes; the players' backoff adds on top of both.
		Buckets: []float64{1, 2, 5, 10, 15, 30, 60, 120, 300, 600, 1800},
	})

	bootRewindActions = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "cmdctrl_boot_rewind_actions",
		Help: "Per table restored at boot, how many actions behind the previous process's last commit it came back (0: no rewind). Observed only when the previous process left a shutdown record.",
		// le=0 is the tables that did not rewind.
		Buckets: []float64{0, 1, 2, 3, 5, 10, 25, 50, 100, 250, 500},
	})

	bootRewindSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "cmdctrl_boot_rewind_seconds",
		Help: "Per table restored at boot, the seconds between its restore point and the previous process's shutdown: how much play the rewind covered (0: no rewind). Observed only when the previous process left a shutdown record.",
		// An upper bound on the play lost: it is measured from when the
		// restore point was captured, not from the last lost action.
		Buckets: []float64{0, 1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600},
	})

	bootRewoundGames = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cmdctrl_boot_rewound_games",
		Help: "Tables the last boot restored behind where the previous process left them. Zero when none did, or when there was no shutdown record to compare against.",
	})

	bootPreviousShutdownGames = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cmdctrl_boot_previous_shutdown_games",
		Help: "Tables the previous process held at shutdown, by what its shutdown flush did with their restore point: written, nothing_pending, failed, or no_restore_point (nothing on disk to restore).",
	}, []string{"flush"})
)

// processStart is when this package initialised, within milliseconds of
// the Go runtime handing over: the zero of every boot time here. Its
// imports initialise first, so card catalog init() functions, which
// run after it, are inside the measured time. The exec and the dynamic
// loader before it are not; process_start_time_seconds has those.
var processStart = time.Now()

// now is time.Now, replaced by tests.
var now = time.Now

func init() {
	bootDowntimeSeconds.Set(math.NaN())
	for _, p := range bootPhaseLabels {
		bootPhaseSeconds.WithLabelValues(p)
	}
	for _, f := range flushLabels {
		bootPreviousShutdownGames.WithLabelValues(f)
	}
}

// ProcessStart is the instant the boot clock started.
func ProcessStart() time.Time { return processStart }

// BootPhase records how long one boot phase took. A phase outside the
// closed set is dropped.
func BootPhase(phase string, d time.Duration) {
	if !slices.Contains(bootPhaseLabels, phase) {
		return
	}
	bootPhaseSeconds.WithLabelValues(phase).Set(d.Seconds())
}

// BootReady records that the server is ready to serve: the listener is
// bound. It returns the time that took, for the log line beside it.
func BootReady() time.Duration {
	d := now().Sub(processStart)
	bootReadySeconds.Set(d.Seconds())
	return d
}

// BootFirstFrame observes one restored table sending its first snapshot
// frame, and returns the time since the process started.
func BootFirstFrame() time.Duration {
	d := now().Sub(processStart)
	bootFirstFrameSeconds.Observe(d.Seconds())
	return d
}

// BootRewind observes one restored table's rewind: the actions it came
// back behind the previous process's last commit and the play that
// spanned. Zeros mean it came back exactly where it stood.
func BootRewind(actions uint64, span time.Duration) {
	bootRewindActions.Observe(float64(actions))
	bootRewindSeconds.Observe(max(span, 0).Seconds())
}

// PreviousShutdown is what the previous process's shutdown record says,
// as the boot hands it to SetPreviousShutdown.
type PreviousShutdown struct {
	// Downtime is the previous record's time to this process's start.
	Downtime time.Duration
	// Flush counts the previous process's tables by flush result. A key
	// outside the Flush values is dropped.
	Flush map[string]int
	// Rewound is how many restored tables came back behind their last
	// commit.
	Rewound int
}

// SetPreviousShutdown sets the gauges the previous shutdown's record
// explains. The boot calls it once, after the restore pass, and only
// when there was a record; without one the gauges keep their start
// values (NaN downtime, zero counts).
func SetPreviousShutdown(p PreviousShutdown) {
	bootDowntimeSeconds.Set(max(p.Downtime, 0).Seconds())
	for _, f := range flushLabels {
		bootPreviousShutdownGames.WithLabelValues(f).Set(float64(p.Flush[f]))
	}
	bootRewoundGames.Set(float64(p.Rewound))
}
