package ws

// restore_rate.go writes restore points at most once per interval
// (#3823, ADR 0041's 2026-10-15 amendment).
//
// A restore point cost about 10 ms and 14 MB per commit at 5,000
// events, most of it re-encoding the event log, and a bot table
// commits several times a second. The owner's decision of 2026-10-15:
// rate-limit the writes, then encode each event once.
//
// WHAT STILL HOLDS
//
// Every commit still captures the game, as before, because only the
// capture says whether the state is a restore point (persist.go's
// rule). A capture that is one becomes the room's PENDING restore
// point, replacing any older pending one. It is written:
//
//   - at once, when the last write is at least the interval old (the
//     first restorable commit of a quiet spell pays no delay);
//   - otherwise when the interval since the last write runs out, by a
//     timer, so a table that goes quiet is written at its last
//     restorable state within the interval;
//   - at once when the game has ended (#2919: the final state is the
//     one a boot must bring back);
//   - on FlushRestorePoint, which shutdown runs for every room before
//     the shutdown census and again after the last commit can land.
//
// Because the pending point is a capture, not a "dirty" flag, a table
// that walks into a prompt after a restorable commit still gets that
// commit written: the file is the last clean state, exactly as before,
// only up to one interval later. A graceful restart (every deploy,
// rollback and Scryfall refresh restarts with SIGTERM) flushes, so it
// rewinds no table further than it did. A crash, or SIGKILL, can
// rewind a table by up to one interval more than it used to.
//
// A deleted room is closed first (closeRestorePoints), so a pending
// write or its timer cannot bring the file back after the reaper has
// removed it.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// DefaultRestorePointInterval is the server's restore-point interval
// when CMDCTRL_RESTORE_POINT_INTERVAL is unset (#3823): the longest a
// crash can rewind a busy table past its last clean state.
const DefaultRestorePointInterval = 5 * time.Second

// pendingRestorePoint is a restorable capture waiting to be written.
type pendingRestorePoint struct {
	seq         uint64
	snap        *game.GameSnapshot
	eventLogGen uint64
}

// restoreGate serialises publishing a restore point with closing the
// room, so a write cannot land after the reaper (RoomManager.Delete)
// removed the file. A leaf lock: nothing is locked inside it, so it
// can be taken under r.mu or under the lobby's lock.
type restoreGate struct {
	mu     sync.Mutex
	closed bool
}

// publish writes parts to path atomically, as one file, unless the gate
// is closed, and reports whether it did.
func (g *restoreGate) publish(path string, parts [][]byte) (bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false, nil
	}
	return true, writeFileAtomic(path, parts...)
}

// close shuts the gate for good.
func (g *restoreGate) close() {
	g.mu.Lock()
	g.closed = true
	g.mu.Unlock()
}

func (g *restoreGate) isClosed() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.closed
}

// SetRestorePointInterval sets how often this room may write its
// restore point. Zero or less writes on every restorable commit, which
// is NewRoom's default and what tests that read the file straight after
// an Apply rely on. RoomManager passes its own interval to every room
// it creates or restores.
func (r *Room) SetRestorePointInterval(d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.restoreEvery = max(d, 0)
}

// noteRestorePointLocked is the restore point's half of a commit at
// seq: capture, and write now or later. Caller MUST hold r.mu.
func (r *Room) noteRestorePointLocked(seq uint64) {
	if r.dumpDir == "" || r.restoreGate.isClosed() {
		return
	}
	p, ok := r.captureRestorePointLocked(seq)
	if !ok {
		// An ended table keeps its final state as its restore point,
		// so a deploy does not take it from the players still looking
		// at it (#2919; RestoreRooms brings it back for
		// EndedTableRetention). If that state cannot be written,
		// neither the older file on disk nor a pending capture may
		// bring the table back live: drop both.
		if r.Game.CurrentState() == game.StateEnded {
			r.restorePending = nil
			r.stopRestoreTimerLocked()
			r.RemoveRestorePoint()
		}
		return
	}
	r.restorePending = &p
	if r.restoreEvery <= 0 || p.snap.State == game.StateEnded ||
		time.Since(r.restoreLastWrite) >= r.restoreEvery {
		r.flushRestorePointLocked()
		return
	}
	r.armRestoreTimerLocked()
}

// FlushRestorePoint writes this room's pending restore point now, if
// it has one. Shutdown calls it for every room (RoomManager.FlushRestorePoints),
// so a graceful restart loses nothing to the rate limit. It reports
// what it did, for the shutdown record (#3037).
func (r *Room) FlushRestorePoint() FlushOutcome {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.flushRestorePointLocked()
}

// FlushOutcome is what one FlushRestorePoint did, spelled as the flush
// label of cmdctrl_boot_previous_shutdown_games.
type FlushOutcome string

const (
	// FlushWritten: a pending restore point was written.
	FlushWritten FlushOutcome = metrics.FlushWritten
	// FlushNothingPending: there was nothing to write.
	FlushNothingPending FlushOutcome = metrics.FlushNothingPending
	// FlushFailed: the write was attempted and failed.
	FlushFailed FlushOutcome = metrics.FlushFailed
)

// flushRestorePointLocked writes the pending restore point, if any.
// Failure is warn-only, as it always was: losing a restore point costs
// a rewind on the next deploy; refusing the player's action would cost
// them the move they just made. Caller MUST hold r.mu.
func (r *Room) flushRestorePointLocked() FlushOutcome {
	r.stopRestoreTimerLocked()
	p := r.restorePending
	if p == nil {
		return FlushNothingPending
	}
	r.restorePending = nil
	writes := r.restoreWrites
	err := r.publishRestorePointLocked(*p)
	if err != nil {
		r.log.Warn("restore point write failed", "err", err, "seq", p.seq)
	}
	outcome := FlushWritten
	switch {
	case err != nil:
		outcome = FlushFailed
	case r.restoreWrites == writes:
		// The room was deleted under us: the gate was closed and
		// nothing was published.
		outcome = FlushNothingPending
	}
	if p.snap.State == game.StateEnded {
		// The ended table's final state could not be written: the
		// older file is a live game (#2919).
		if err != nil {
			r.RemoveRestorePoint()
		}
		// An ended table writes nothing more unless the admin rewinds
		// it, so its encoded log (a few MB in a long game) is released;
		// a rewind rebuilds it.
		r.eventLog = game.EncodedEventLog{}
	}
	return outcome
}

// armRestoreTimerLocked makes sure a write is due when the interval
// since the last one runs out. Caller MUST hold r.mu.
func (r *Room) armRestoreTimerLocked() {
	if r.restoreTimer != nil {
		return
	}
	r.restoreTimerTok++
	tok := r.restoreTimerTok
	r.restoreTimer = time.AfterFunc(r.restoreEvery-time.Since(r.restoreLastWrite), func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		// A timer stopped too late to keep it from firing finds a
		// newer token (or none) and leaves the newer timer's work to it.
		if r.restoreTimer == nil || tok != r.restoreTimerTok {
			return
		}
		r.restoreTimer = nil
		if r.restoreGate.isClosed() {
			r.restorePending = nil
			return
		}
		r.flushRestorePointLocked()
	})
}

// stopRestoreTimerLocked disarms the trailing write. Caller MUST hold
// r.mu.
func (r *Room) stopRestoreTimerLocked() {
	if r.restoreTimer == nil {
		return
	}
	r.restoreTimer.Stop()
	r.restoreTimer = nil
}

// closeRestorePoints stops this room from ever writing its restore
// point again. RoomManager.Delete calls it before removing the files.
// It takes only the gate's leaf lock, never r.mu, because Delete runs
// under the lobby's lock; a pending capture and its timer find the
// gate closed and drop themselves.
func (r *Room) closeRestorePoints() {
	r.restoreGate.close()
}

// FlushRestorePoints writes every room's pending restore point. The
// server runs it at SIGTERM (before the shutdown census, so the census
// reads what is really on disk, and again once no commit can land),
// which is what keeps a deploy's restart from rewinding any table
// further than its last clean state (#3823).
func (m *RoomManager) FlushRestorePoints() FlushOutcomes {
	out := FlushOutcomes{}
	for _, r := range m.List() {
		out[r.Game.ID] = r.FlushRestorePoint()
	}
	return out
}

// FlushOutcomes is what a FlushRestorePoints did, by table.
type FlushOutcomes map[uuid.UUID]FlushOutcome

// Then folds a later flush into o: a table the later flush wrote (or
// failed to write) takes its result, because that is the newest word
// on its restore point; one it found nothing pending for keeps the
// earlier. The shutdown flushes twice, before the census and after the
// last commit can land, and the record wants both.
func (o FlushOutcomes) Then(later FlushOutcomes) FlushOutcomes {
	merged := make(FlushOutcomes, len(o)+len(later))
	for id, v := range o {
		merged[id] = v
	}
	for id, v := range later {
		if _, seen := merged[id]; !seen || v != FlushNothingPending {
			merged[id] = v
		}
	}
	return merged
}

// writeRestorePointLocked captures the game and publishes it as this
// room's restore point, atomically, IF the capture is restorable, now,
// whatever the rate limit. A pending restore point it supersedes is
// dropped.
//
// Returns (false, nil) when the game is holding continuations — that
// is the ordinary, expected case mid-prompt and is not an error. The
// previous restore point stays on disk.
//
// Caller MUST hold r.mu.
func (r *Room) writeRestorePointLocked(seq uint64) (bool, error) {
	if r.dumpDir == "" {
		return false, nil
	}
	p, ok := r.captureRestorePointLocked(seq)
	if !ok {
		return false, nil
	}
	r.restorePending = nil
	r.stopRestoreTimerLocked()
	if err := r.publishRestorePointLocked(p); err != nil {
		return false, err
	}
	return true, nil
}

// captureRestorePointLocked captures the game for a restore point and
// reports whether it is one. A capture that is not is tallied by what
// blocked it; one that is ends the current run of skips.
//
// Caller MUST hold r.mu.
func (r *Room) captureRestorePointLocked(seq uint64) (pendingRestorePoint, bool) {
	snap, gen := r.Game.CaptureRestorePoint()
	if !snap.Restorable() {
		// ADR 0041 P7 (#1497): count what blocked it, by kind, so the
		// shutdown census can say what keeps restore points stale
		// between deploys rather than only at the instant of one.
		r.skips.noteSkip(snap.Continuations)
		return pendingRestorePoint{}, false
	}
	r.skips.noteRestorable()
	return pendingRestorePoint{seq: seq, snap: snap, eventLogGen: gen}, true
}

// publishRestorePointLocked encodes p and writes it as this room's
// restore point, atomically.
//
// r.generation is read under r.mu (the caller already holds it) —
// every later write in this room's life carries forward the
// generation the room is CURRENTLY serving, so the file always
// answers "if the process died and came back right now, what
// generation would clients already holding this room's frames be
// told about". Only restoreOne ever advances it.
//
// Caller MUST hold r.mu.
func (r *Room) publishRestorePointLocked(p pendingRestorePoint) error {
	// Stamped on the attempt, not the success: a failing disk is tried
	// once an interval, not on every commit.
	r.restoreLastWrite = time.Now()
	parts, err := r.encodeRestorePointLocked(p)
	if err != nil {
		return fmt.Errorf("marshal restore point: %w", err)
	}
	published, err := r.restoreGate.publish(restorePointPath(r.dumpDir, r.Game.ID), parts)
	if err != nil || !published {
		return err
	}
	// Bookkeeping only, for the shutdown census (#524) — the write
	// above already succeeded, so this cannot be the thing that fails
	// a caller's Apply.
	r.lastRestorePoint = restorePointRecord{Seq: p.seq, At: time.Now().UTC()}
	r.restoreWrites++
	return nil
}

// eventSeqKey is the restore point's key that follows "events":
// GameSnapshot declares EventSeq right after Events, and EventSeq is
// not omitempty. It can appear in the encoding only as a key, since
// inside a JSON string its quotes are escaped.
var eventSeqKey = []byte(`,"eventSeq":`)

var (
	eventsHead = []byte(`,"events":[`)
	eventsTail = []byte(`]`)
)

// encodeRestorePointLocked marshals the restore point file for p with
// its event log spliced in from the room's encoded-event cache instead
// of re-encoded (#3823). It returns the file as parts, written one
// after another, so the cached log is never copied into a fresh
// buffer; the parts alias the cache and are good until the next call.
// Joined, they are byte-identical to json.Marshal of the whole file
// (TestRestorePointFromTheCacheIsByteIdentical). Should the key the log
// goes before be missing or appear twice, it marshals the file whole
// instead.
//
// Caller MUST hold r.mu.
func (r *Room) encodeRestorePointLocked(p pendingRestorePoint) ([][]byte, error) {
	file := restorePointFile{Seq: p.seq, Generation: r.generation, Snapshot: p.snap}
	whole := func() ([][]byte, error) {
		b, err := json.Marshal(file)
		return [][]byte{b}, err
	}
	if len(p.snap.Events) == 0 {
		return whole()
	}
	events, err := r.eventLog.Encode(p.snap.Events, p.eventLogGen)
	if err != nil {
		return nil, err
	}
	shell := *p.snap
	shell.Events = nil
	file.Snapshot = &shell
	payload, err := json.Marshal(file)
	if err != nil {
		return nil, err
	}
	at := bytes.Index(payload, eventSeqKey)
	if at < 0 || bytes.Contains(payload[at+len(eventSeqKey):], eventSeqKey) {
		file.Snapshot = p.snap
		return whole()
	}
	return [][]byte{payload[:at], eventsHead, events, eventsTail, payload[at:]}, nil
}
