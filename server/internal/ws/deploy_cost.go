package ws

// deploy_cost.go is the measurement half of ADR 0195 phase 0 (#3037):
// what a restart costs a live table, readable in Grafana after the
// fact.
//
// The shutdown census (shutdown_report.go) already logs, per table,
// whether a restart would rewind it and by how many actions, and the
// restore pass logs each table that came back. What neither can do is
// pair the two: a process that has stopped cannot be scraped, and the
// process that boots next knows only the restore point it read. So the
// shutdown writes a small record of what it left (this file's
// ShutdownRecord), and the next boot reads it, compares each table's
// last commit with the restore point it came back from, and sets the
// cmdctrl_boot_* metrics (metrics/deploy.go). The record is consumed:
// it is removed as it is read, so a later crash cannot be mistaken for
// a deploy by reusing an old one.
//
// Nothing here is on a hot path and nothing here can fail a boot or a
// shutdown: a missing, unreadable or stale record only means the
// rewind histograms have nothing to say that time.

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// shutdownRecordVersion is the record's format. A record of a version
// this build does not know is ignored, not guessed at.
const shutdownRecordVersion = 1

// shutdownRecordPath is where the shutdown leaves its record, beside
// the restore points it describes.
func shutdownRecordPath(dumpDir string) string {
	return filepath.Join(dumpDir, "deploy", "shutdown-record.json")
}

// ShutdownRecord is what a process leaves on graceful shutdown for the
// next boot: one entry per table that writes restore points.
type ShutdownRecord struct {
	Version int `json:"version"`
	// At is when the record was taken, which is just before the process
	// exits: the last moment any table could have moved.
	At    time.Time            `json:"at"`
	Games []ShutdownRecordGame `json:"games"`
}

// ShutdownRecordGame is one table's entry.
type ShutdownRecordGame struct {
	GameID uuid.UUID `json:"game_id"`
	// Archived is a retired table. Its entry counts in the flush totals
	// and is left out of the rewind histograms: nobody is at it.
	Archived bool `json:"archived,omitempty"`
	// LiveSeq is the last commit the table made.
	LiveSeq uint64 `json:"live_seq"`
	// RestoreSeq is the seq of the restore point on disk after the
	// flush; zero when there is none.
	RestoreSeq uint64 `json:"restore_seq"`
	// SeqBehind is LiveSeq - RestoreSeq: the actions the restart throws
	// away, which the next boot confirms against the file it reads.
	SeqBehind uint64 `json:"seq_behind"`
	// Clean reports that the live state was itself a usable restore
	// point, so a flushed table has SeqBehind 0.
	Clean bool `json:"clean"`
	// Flush is what the shutdown flush did: a metrics.Flush value.
	Flush string `json:"flush"`
}

// NewShutdownRecord builds the record for the rooms now registered. at
// is the record's time; flushes is what FlushRestorePoints reported
// (a table it did not mention is read as nothing pending); archived
// says whether the lobby has retired a table, as for LogShutdownCensus.
// A room that writes no restore points (a practice table) is left out.
func (m *RoomManager) NewShutdownRecord(at time.Time, flushes FlushOutcomes, archived func(uuid.UUID) bool) ShutdownRecord {
	rec := ShutdownRecord{Version: shutdownRecordVersion, At: at.UTC()}
	for _, r := range m.List() {
		if r.dumpDir == "" {
			continue
		}
		rep := r.ShutdownReport()
		g := ShutdownRecordGame{
			GameID:     rep.GameID,
			LiveSeq:    rep.LiveSeq,
			RestoreSeq: rep.RestoreSeq,
			SeqBehind:  rep.SeqBehind,
			Clean:      rep.Clean,
			Flush:      string(FlushNothingPending),
		}
		if archived != nil {
			g.Archived = archived(rep.GameID)
		}
		if f, ok := flushes[rep.GameID]; ok {
			g.Flush = string(f)
		}
		if !rep.HasRestorePoint {
			g.Flush = metrics.FlushNoRestorePoint
		}
		rec.Games = append(rec.Games, g)
	}
	return rec
}

// WriteShutdownRecord publishes rec atomically for the next boot. A
// manager with no data directory writes nothing.
func (m *RoomManager) WriteShutdownRecord(rec ShutdownRecord) error {
	if m.dumpDir == "" {
		return nil
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal shutdown record: %w", err)
	}
	return writeFileAtomic(shutdownRecordPath(m.dumpDir), b)
}

// LogShutdownRecord says, in one line, what the record holds: how the
// flush went and what a restart will rewind. The census lines before
// it name the tables; this one is the total an operator reads first.
func LogShutdownRecord(log *slog.Logger, rec ShutdownRecord) {
	if log == nil {
		log = slog.Default()
	}
	flush := map[string]int{}
	var rewound int
	var actions uint64
	for _, g := range rec.Games {
		flush[g.Flush]++
		if g.SeqBehind > 0 && !g.Archived {
			rewound++
			actions += g.SeqBehind
		}
	}
	level := slog.LevelInfo
	if flush[metrics.FlushFailed] > 0 {
		level = slog.LevelWarn
	}
	log.Log(context.Background(), level, "shutdown record written",
		"games", len(rec.Games), "flush", flush, "would_rewind", rewound, "actions_rewound", actions)
}

// bootMeasurement is the boot restore pass's measuring state: the
// previous process's record, when there was a readable one.
type bootMeasurement struct {
	prev *ShutdownRecord
}

// ArmBootMeasurement makes the NEXT restore pass the boot's: it reads
// and removes the previous process's shutdown record, marks each table
// that pass restores for first-frame timing, and, after the pass, sets
// the cmdctrl_boot_* metrics from the two. The server calls it once at
// boot, before RestoreFromDisk. A later pass (the hot restore of the
// tables the load cap deferred) is not the boot's and measures nothing.
func (m *RoomManager) ArmBootMeasurement() {
	bm := &bootMeasurement{prev: m.consumeShutdownRecord()}
	m.mu.Lock()
	m.boot = bm
	m.mu.Unlock()
}

// takeBootMeasurement hands the armed measurement to the pass that
// starts now, once.
func (m *RoomManager) takeBootMeasurement() *bootMeasurement {
	m.mu.Lock()
	defer m.mu.Unlock()
	bm := m.boot
	m.boot = nil
	return bm
}

// consumeShutdownRecord reads the previous shutdown record and removes
// it. Nil when there is none, or it cannot be used.
func (m *RoomManager) consumeShutdownRecord() *ShutdownRecord {
	if m.dumpDir == "" {
		return nil
	}
	path := shutdownRecordPath(m.dumpDir)
	raw, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			m.log.Warn("shutdown record unreadable; this boot cannot say how far tables rewound", "path", path, "err", err)
		}
		return nil
	}
	if err := os.Remove(path); err != nil {
		m.log.Warn("shutdown record could not be removed", "path", path, "err", err)
	}
	var rec ShutdownRecord
	if err := json.Unmarshal(raw, &rec); err != nil || rec.Version != shutdownRecordVersion || rec.At.IsZero() {
		m.log.Warn("shutdown record not understood; this boot cannot say how far tables rewound", "path", path, "err", err)
		return nil
	}
	return &rec
}

// observeBoot is the boot pass's measurement, run on its outcomes: it
// marks the restored rooms for first-frame timing, and, with a record,
// observes each one's rewind and sets the previous-shutdown gauges.
func (m *RoomManager) observeBoot(bm *bootMeasurement, outcomes []RestoreOutcome) {
	for _, o := range outcomes {
		if o.Restored() {
			o.Room.bootFirstFrame.Store(true)
		}
	}
	if bm.prev == nil {
		m.log.Info("deploy cost: no shutdown record from the previous process; rewinds are in its shutdown census lines")
		return
	}
	rec := bm.prev
	byGame := make(map[uuid.UUID]ShutdownRecordGame, len(rec.Games))
	flush := map[string]int{}
	for _, g := range rec.Games {
		byGame[g.GameID] = g
		flush[g.Flush]++
	}

	var rewound, measured, unrecorded int
	var actions uint64
	for _, o := range outcomes {
		if !o.Restored() {
			continue
		}
		prev, ok := byGame[o.GameID]
		if !ok {
			// Made after the record, or the previous process was not the
			// one that wrote this table's file.
			unrecorded++
			continue
		}
		if prev.Archived {
			continue
		}
		var behind uint64
		var span time.Duration
		if prev.LiveSeq > o.RestoredSeq {
			behind = prev.LiveSeq - o.RestoredSeq
			span = max(rec.At.Sub(o.RestorePointAt), 0)
			rewound++
			actions += behind
			m.log.Warn("deploy cost: table rewound",
				"game_id", o.GameID, "actions_behind", behind, "span", span.Round(time.Second).String(),
				"live_seq", prev.LiveSeq, "restored_seq", o.RestoredSeq, "flush", prev.Flush)
		}
		measured++
		metrics.BootRewind(behind, span)
	}

	downtime := max(metrics.ProcessStart().Sub(rec.At), 0)
	metrics.SetPreviousShutdown(metrics.PreviousShutdown{Downtime: downtime, Flush: flush, Rewound: rewound})
	m.log.Info("deploy cost: previous shutdown compared",
		"downtime", downtime.Round(time.Millisecond).String(),
		"previous_games", len(rec.Games), "flush", flush,
		"restored_measured", measured, "restored_unrecorded", unrecorded,
		"rewound", rewound, "actions_rewound", actions)
}

// noteFirstFrame records that a snapshot frame has just gone to a
// client of this room. The first one after a boot restore is the number
// cmdctrl_boot_first_frame_seconds holds; every later call is one
// atomic load.
func (r *Room) noteFirstFrame() {
	if !r.bootFirstFrame.Load() || !r.bootFirstFrame.CompareAndSwap(true, false) {
		return
	}
	d := metrics.BootFirstFrame()
	r.log.Info("deploy cost: first frame after boot", "since_process_start", d.Round(time.Millisecond).String())
}
