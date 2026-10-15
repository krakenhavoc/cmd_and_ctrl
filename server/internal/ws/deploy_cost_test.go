package ws

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// deployReading is the deploy-cost metrics as the registry reports
// them, so a test can compare before and after (the metrics are
// package-level and other tests in the process move them too).
type deployReading struct {
	rewindCount, firstFrameCount uint64
	rewindActions, rewindSeconds float64
	rewound                      float64
	flush                        map[string]float64
}

func readDeploy(t *testing.T) deployReading {
	t.Helper()
	fams, err := metrics.Registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	r := deployReading{flush: map[string]float64{}}
	for _, f := range fams {
		for _, m := range f.GetMetric() {
			switch f.GetName() {
			case "cmdctrl_boot_rewind_actions":
				r.rewindCount = m.GetHistogram().GetSampleCount()
				r.rewindActions = m.GetHistogram().GetSampleSum()
			case "cmdctrl_boot_rewind_seconds":
				r.rewindSeconds = m.GetHistogram().GetSampleSum()
			case "cmdctrl_boot_first_frame_seconds":
				r.firstFrameCount = m.GetHistogram().GetSampleCount()
			case "cmdctrl_boot_rewound_games":
				r.rewound = m.GetGauge().GetValue()
			case "cmdctrl_boot_previous_shutdown_games":
				r.flush[m.GetLabel()[0].GetValue()] = m.GetGauge().GetValue()
			}
		}
	}
	return r
}

// shutdownAndRestart plays one deploy: the first manager flushes and
// writes its record, a second manager over the same directory arms the
// boot measurement and restores.
func shutdownAndRestart(t *testing.T, dir string, old *RoomManager, at time.Time, archived func(uuid.UUID) bool) (*RoomManager, []RestoreOutcome, ShutdownRecord) {
	t.Helper()
	flushes := old.FlushRestorePoints()
	rec := old.NewShutdownRecord(at, flushes, archived)
	if err := old.WriteShutdownRecord(rec); err != nil {
		t.Fatal(err)
	}
	next := restoreTestManager(t, dir)
	next.ArmBootMeasurement()
	return next, next.RestoreRooms(), rec
}

// The whole measurement, end to end: a table behind a continuation at
// shutdown comes back at its last clean state, and the boot reads that
// as a rewind of the actions it threw away; a table at a clean point
// comes back as a zero. The record is consumed.
func TestBootMeasuresTheRewindAgainstTheShutdownRecord(t *testing.T) {
	dir := t.TempDir()
	old := rateLimitedManager(t, dir, time.Hour)

	clean := newPersistGame(t)
	cleanRoom := old.Create(clean)
	setLife(t, cleanRoom, clean.Seats[0].ID, 33)

	stuck := newPersistGame(t)
	stuckRoom := old.Create(stuck)
	setLife(t, stuckRoom, stuck.Seats[0].ID, 38) // written at once
	setLife(t, stuckRoom, stuck.Seats[0].ID, 30) // held back by the interval
	cleanSeq := stuckRoom.Seq()
	openContinuation(t, stuckRoom, stuck.Seats[0].ID)
	setLife(t, stuckRoom, stuck.Seats[0].ID, 29) // mid-continuation: not a restore point
	liveSeq := stuckRoom.Seq()

	before := readDeploy(t)
	shutdownAt := time.Now()
	next, out, rec := shutdownAndRestart(t, dir, old, shutdownAt.Add(90*time.Second), nil)

	if len(out) != 2 {
		t.Fatalf("restored %d tables, want 2: %+v", len(out), out)
	}
	for _, g := range rec.Games {
		switch g.GameID {
		case stuck.ID:
			if g.LiveSeq != liveSeq || g.RestoreSeq != cleanSeq || g.SeqBehind != liveSeq-cleanSeq || g.Clean {
				t.Errorf("stuck table record = %+v, want live %d restore %d unclean", g, liveSeq, cleanSeq)
			}
			if g.Flush != metrics.FlushWritten {
				t.Errorf("stuck table flush = %q, want written (its clean state was pending)", g.Flush)
			}
		case clean.ID:
			if g.SeqBehind != 0 || !g.Clean || g.Flush != metrics.FlushNothingPending {
				t.Errorf("clean table record = %+v, want clean, nothing behind, nothing pending (its one commit was written at once)", g)
			}
		}
	}

	after := readDeploy(t)
	if got := after.rewindCount - before.rewindCount; got != 2 {
		t.Errorf("rewind observations rose by %d, want 2 (one per restored table)", got)
	}
	if got := after.rewindActions - before.rewindActions; got != float64(liveSeq-cleanSeq) {
		t.Errorf("actions rewound rose by %v, want %d", got, liveSeq-cleanSeq)
	}
	if got := after.rewound; got != 1 {
		t.Errorf("cmdctrl_boot_rewound_games = %v, want 1", got)
	}
	// The span runs from the restore point to the shutdown record: at
	// least the 90 s the record was dated ahead.
	if got := after.rewindSeconds - before.rewindSeconds; got < 90 {
		t.Errorf("seconds rewound rose by %v, want at least 90", got)
	}
	if got, want := after.flush[metrics.FlushWritten], 1.0; got != want {
		t.Errorf("previous shutdown flush=written = %v, want %v", got, want)
	}
	if got, want := after.flush[metrics.FlushNothingPending], 1.0; got != want {
		t.Errorf("previous shutdown flush=nothing_pending = %v, want %v", got, want)
	}
	if _, err := os.Stat(shutdownRecordPath(dir)); !os.IsNotExist(err) {
		t.Errorf("the record was not consumed: stat err = %v", err)
	}

	// First frame: each restored room is timed once, on its first frame.
	room := next.Get(stuck.ID)
	if !room.bootFirstFrame.Load() {
		t.Fatal("a restored room is not marked for first-frame timing")
	}
	room.noteFirstFrame()
	room.noteFirstFrame()
	if got := readDeploy(t).firstFrameCount - before.firstFrameCount; got != 1 {
		t.Errorf("two frames to one room observed %d first frames, want 1", got)
	}
}

// A boot with no record (the first after the build that introduces it,
// or after a crash) still times first frames and observes no rewind.
func TestBootWithoutARecordMeasuresOnlyFirstFrames(t *testing.T) {
	dir := t.TempDir()
	old := rateLimitedManager(t, dir, 0)
	g := newPersistGame(t)
	old.Create(g)
	setLife(t, old.Get(g.ID), g.Seats[0].ID, 31)

	before := readDeploy(t)
	next := restoreTestManager(t, dir)
	next.ArmBootMeasurement()
	if out := next.RestoreRooms(); len(out) != 1 || !out[0].Restored() {
		t.Fatalf("restore: %+v", out)
	}
	if got := readDeploy(t).rewindCount - before.rewindCount; got != 0 {
		t.Errorf("rewinds observed with no record: %d", got)
	}
	if !next.Get(g.ID).bootFirstFrame.Load() {
		t.Error("first-frame timing is not armed without a record")
	}
}

// Only the pass that follows ArmBootMeasurement is the boot's: the hot
// restore of deferred tables later in the process's life must not
// count as a boot, or its first frame would read minutes late.
func TestOnlyTheBootPassIsMeasured(t *testing.T) {
	dir := t.TempDir()
	old := rateLimitedManager(t, dir, 0)
	g := newPersistGame(t)
	old.Create(g)
	setLife(t, old.Get(g.ID), g.Seats[0].ID, 31)

	next := restoreTestManager(t, dir)
	next.ArmBootMeasurement()
	if out := next.RestoreRooms(); len(out) != 1 || !next.Get(g.ID).bootFirstFrame.Load() {
		t.Fatalf("boot restore: %+v", out)
	}

	// A table written afterwards and restored by a second pass, as the
	// hot restore does, is not the boot's.
	late := newPersistGame(t)
	writer := rateLimitedManager(t, dir, 0)
	setLife(t, writer.Create(late), late.Seats[0].ID, 12)
	if out := next.RestoreRooms(); len(out) != 1 || !out[0].Restored() {
		t.Fatalf("second restore: %+v", out)
	}
	if next.Get(late.ID).bootFirstFrame.Load() {
		t.Error("a restore pass nobody armed marked its room for first-frame timing")
	}
}

// A record that cannot be used is dropped, not trusted and not left to
// be mistaken for the next deploy's.
func TestUnusableRecordsAreDiscarded(t *testing.T) {
	for name, content := range map[string]string{
		"garbage":       `{"version": 1, "games": [`,
		"wrong version": `{"version": 99, "at": "2026-10-15T12:00:00Z", "games": []}`,
		"no time":       `{"version": 1, "games": []}`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.MkdirAll(dir+"/deploy", 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(shutdownRecordPath(dir), []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			m := restoreTestManager(t, dir)
			if rec := m.consumeShutdownRecord(); rec != nil {
				t.Errorf("read a record from %s: %+v", name, rec)
			}
			if _, err := os.Stat(shutdownRecordPath(dir)); !os.IsNotExist(err) {
				t.Errorf("the unusable record was left in place: %v", err)
			}
		})
	}
}

// A table the previous process had retired is counted in the flush
// totals but not in the rewind histograms; a table made after the
// record is not compared with anything.
func TestArchivedAndUnrecordedTablesAreNotRewinds(t *testing.T) {
	dir := t.TempDir()
	old := rateLimitedManager(t, dir, time.Hour)
	retired := newPersistGame(t)
	room := old.Create(retired)
	setLife(t, room, retired.Seats[0].ID, 38)
	setLife(t, room, retired.Seats[0].ID, 30)
	openContinuation(t, room, retired.Seats[0].ID)

	before := readDeploy(t)
	_, out, rec := shutdownAndRestart(t, dir, old, time.Now(),
		func(id uuid.UUID) bool { return id == retired.ID })
	if len(out) != 1 || !out[0].Restored() {
		t.Fatalf("restore: %+v", out)
	}
	if !rec.Games[0].Archived || rec.Games[0].SeqBehind == 0 {
		t.Fatalf("setup: record = %+v, want an archived table that is behind", rec.Games[0])
	}
	after := readDeploy(t)
	if got := after.rewindCount - before.rewindCount; got != 0 {
		t.Errorf("an archived table was observed %d times, want 0", got)
	}
	if after.rewound != 0 {
		t.Errorf("rewound = %v, want 0: archived tables do not count", after.rewound)
	}
}

// What the flush did is what the record says, per table: written when a
// point was pending, nothing_pending when it was already on disk, and
// no_restore_point for a table that never reached a clean state.
func TestShutdownRecordNamesWhatTheFlushDid(t *testing.T) {
	dir := t.TempDir()
	m := rateLimitedManager(t, dir, time.Hour)

	pending := newPersistGame(t)
	pr := m.Create(pending)
	setLife(t, pr, pending.Seats[0].ID, 38) // written at once
	setLife(t, pr, pending.Seats[0].ID, 30) // held back

	settled := newPersistGame(t)
	sr := m.Create(settled)
	setLife(t, sr, settled.Seats[0].ID, 38)
	sr.FlushRestorePoint() // already flushed before the shutdown's

	fresh := newPersistGame(t)
	fr := m.Create(fresh)
	openContinuation(t, fr, fresh.Seats[0].ID) // never clean: no file

	// A practice table writes nothing and is not in the record.
	m.CreateEphemeral(newPersistGame(t))

	flushes := m.FlushRestorePoints()
	if flushes[pending.ID] != FlushWritten || flushes[settled.ID] != FlushNothingPending {
		t.Fatalf("flush outcomes = %v", flushes)
	}
	rec := m.NewShutdownRecord(time.Now(), flushes, nil)
	got := map[uuid.UUID]string{}
	for _, g := range rec.Games {
		got[g.GameID] = g.Flush
	}
	want := map[uuid.UUID]string{
		pending.ID: metrics.FlushWritten,
		settled.ID: metrics.FlushNothingPending,
		fresh.ID:   metrics.FlushNoRestorePoint,
	}
	if len(got) != len(want) {
		t.Fatalf("record holds %d tables, want %d (the practice table is left out): %v", len(got), len(want), got)
	}
	for id, w := range want {
		if got[id] != w {
			t.Errorf("table flush = %q, want %q", got[id], w)
		}
	}
}

// The second flush, after the last commit can land, wins only where it
// did something: a table it found nothing pending for keeps the first
// flush's word.
func TestFlushOutcomesThen(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	first := FlushOutcomes{a: FlushWritten, b: FlushNothingPending}
	later := FlushOutcomes{a: FlushNothingPending, b: FlushFailed, c: FlushNothingPending}
	got := first.Then(later)
	want := FlushOutcomes{a: FlushWritten, b: FlushFailed, c: FlushNothingPending}
	for id, w := range want {
		if got[id] != w {
			t.Errorf("merged[%s] = %q, want %q", id, got[id], w)
		}
	}
}

// A room deleted under the flush (its gate is closed) wrote nothing and
// says so, rather than claiming a write.
func TestFlushOfADeletedRoomWroteNothing(t *testing.T) {
	dir := t.TempDir()
	m := rateLimitedManager(t, dir, time.Hour)
	g := newPersistGame(t)
	room := m.Create(g)
	setLife(t, room, g.Seats[0].ID, 38)
	setLife(t, room, g.Seats[0].ID, 30)
	m.Delete(g.ID)
	if got := room.FlushRestorePoint(); got != FlushNothingPending {
		t.Errorf("flush of a deleted room = %q, want nothing_pending", got)
	}
}

// No data directory, no record, and no error.
func TestNoDataDirWritesNoRecord(t *testing.T) {
	m := NewRoomManager(nil, "")
	if err := m.WriteShutdownRecord(m.NewShutdownRecord(time.Now(), nil, nil)); err != nil {
		t.Fatal(err)
	}
	m.ArmBootMeasurement()
	if out := m.RestoreRooms(); out != nil {
		t.Errorf("restored %v from nowhere", out)
	}
}
