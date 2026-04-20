package ws

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Room bundles an authoritative game.Game with the wire-format
// infrastructure needed to broadcast snapshots of it: a monotonic
// sequence counter, a slog logger, and optional on-disk crash
// recovery. The actual set of connected clients lives on the Hub;
// Room is intentionally membership-free so the RoomManager can own
// lifecycle concerns independently.
//
// Room is safe for concurrent use. All mutation + snapshot pairs go
// through Apply, which holds a single mutex across the dispatch,
// the sequence allocation, and the state capture — guaranteeing that
// seq numbers are allocated in the same order state progresses, and
// that no two dispatches can interleave their state captures.
//
// S04: Apply and Snapshot return a protocol.GameView (the full,
// unfiltered view) instead of pre-marshalled frame bytes. The hub
// layers per-client visibility filtering on top, which means the
// same captured view can be marshalled into N different tailored
// frames per broadcast — one per connected viewer.
type Room struct {
	Game *game.Game
	log  *slog.Logger

	// mu serializes the mutate → capture → (optional dump) sequence
	// in Apply and Snapshot. Taken before any game lock.
	mu  sync.Mutex
	seq uint64

	// dumpDir is the root for crash-recovery snapshots. Empty string
	// disables disk writes entirely. Each snapshot lands at
	//   <dumpDir>/games/<game-id>.json
	// written via CreateTemp + rename for atomicity. Dumps always
	// contain the full unfiltered view — crash recovery is not a
	// player-facing surface, so visibility filtering does not apply.
	dumpDir string

	// undoStack holds pre-mutation Game clones, oldest first.
	// Capped at undoStackCap; older entries fall off the front when
	// the cap is exceeded. Each Apply pushes the *pre-action* clone
	// before running fn, so Undo restores the state immediately
	// before the most recent action. Snapshot does NOT push (it's a
	// no-op for state).
	undoStack []*game.Game
}

// undoStackCap bounds the per-room undo ring. 32 is a casual-game-
// reasonable depth — the most common use is "I clicked the wrong
// card, let me back up one or two", not deep history rewind.
// Keeping it small bounds the per-room memory footprint (each clone
// is ~10-100 KB depending on board state).
const undoStackCap = 32

// ErrNothingToUndo is returned by Undo when the undo stack is empty
// (no actions have been applied since the room was created or since
// the stack was cleared on game start). The hub turns this into a
// `bad_request` error frame addressed back to the originator.
var ErrNothingToUndo = errSentinel("ws: nothing to undo")

type errSentinel string

func (e errSentinel) Error() string { return string(e) }

// NewRoom constructs a Room wrapping the given game. If log is nil,
// slog.Default() is used. If dumpDir is empty, crash-recovery writes
// are disabled.
func NewRoom(g *game.Game, log *slog.Logger, dumpDir string) *Room {
	if log == nil {
		log = slog.Default()
	}
	return &Room{
		Game:    g,
		log:     log,
		dumpDir: dumpDir,
	}
}

// Apply serializes a mutation against the room: it runs fn (which
// should be a single game-state mutation, typically actions.Dispatch),
// then atomically allocates a sequence number, captures a wire-format
// view of the resulting game state, writes the recovery dump if
// enabled, and returns the view + seq for the hub to filter and
// broadcast. If fn returns an error, Apply returns it verbatim and
// does NOT increment seq or emit a view.
//
// The room mutex is held for the whole mutate → seq → capture → dump
// sequence, so two concurrent clients cannot interleave state and
// cannot observe non-monotonic (seq, state) pairs.
func (r *Room) Apply(fn func() error) (protocol.GameView, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Stash a pre-mutation clone for the undo stack BEFORE running
	// fn. Done eagerly so a failing fn doesn't grow history with a
	// no-op snapshot. If fn fails we drop the clone on the floor.
	pre := r.Game.Clone()
	if err := fn(); err != nil {
		return protocol.GameView{}, 0, err
	}
	r.undoStack = append(r.undoStack, pre)
	if len(r.undoStack) > undoStackCap {
		// Drop the oldest entry. Trim by reslicing forward — the
		// underlying array's first slot is now unreachable, GC'd on
		// next allocation. Keeping the cap small bounds the leak.
		r.undoStack = append(r.undoStack[:0], r.undoStack[1:]...)
	}
	return r.captureLocked(true)
}

// Undo pops the most recent pre-mutation clone off the undo stack,
// restores the live Game to that state, captures a fresh snapshot,
// and returns it for broadcast. Returns ErrNothingToUndo if the
// stack is empty (no actions applied since the last clear).
//
// Undo bumps seq the same way a regular Apply does so connected
// clients see a normal snapshot frame and don't have to special-case
// the rewind. The undo itself is NOT pushed onto the stack — undoing
// an undo would create a cycle; if you want redo, use a separate
// future redo stack (out of scope for v1).
func (r *Room) Undo() (protocol.GameView, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.undoStack) == 0 {
		return protocol.GameView{}, 0, ErrNothingToUndo
	}
	prev := r.undoStack[len(r.undoStack)-1]
	r.undoStack = r.undoStack[:len(r.undoStack)-1]

	// Restore under the game's own write lock so any concurrent
	// readers (PlayerByID, etc.) see a consistent transition.
	r.Game.WithWriteLock(func() {
		r.Game.RestoreFrom(prev)
	})
	return r.captureLocked(true)
}

// Snapshot returns the current view without mutating anything and
// without advancing the sequence counter. Used to send the initial
// state to a just-joined client. The returned view carries the CURRENT
// seq — the same value the last Apply broadcast to other clients.
// Under a join-during-action race, a new client can briefly see two
// consecutive frames with the same seq (one from Snapshot, one from a
// subsequent broadcast of the same state); both carry identical state
// and clients should treat snapshots idempotently.
//
// This "no-bump" semantics is deliberate: if Snapshot advanced the
// shared counter, existing clients would see a seq jump of +2 on
// their next broadcast (Apply→seq+1 AND Snapshot→seq+1) and flag it
// as a dropped frame.
func (r *Room) Snapshot() (protocol.GameView, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.captureLocked(false)
}

// captureLocked is the common state-capture path used by both Apply
// and Snapshot. Caller MUST hold r.mu.
//
// If advanceSeq is true, the sequence counter is incremented — but
// only AFTER the crash-recovery dump marshal succeeds, so a marshal
// failure doesn't leave the counter advanced with no corresponding
// on-disk record.
//
// The full-fidelity view is returned to the caller; the hub layer is
// responsible for per-client visibility filtering and marshalling.
func (r *Room) captureLocked(advanceSeq bool) (protocol.GameView, uint64, error) {
	nextSeq := r.seq
	if advanceSeq {
		nextSeq = r.seq + 1
	}
	view := protocol.ViewOfGame(r.Game)

	// If crash recovery is enabled, marshal the full (unfiltered)
	// payload now, BEFORE committing the seq advance. A marshal
	// failure must not leave r.seq advanced.
	var dumpPayload []byte
	if r.dumpDir != "" {
		payload, err := json.Marshal(protocol.SnapshotPayload{
			Seq:  nextSeq,
			Game: view,
		})
		if err != nil {
			return protocol.GameView{}, 0, fmt.Errorf("marshal snapshot: %w", err)
		}
		dumpPayload = payload
	}

	// Marshal succeeded (or was skipped). Commit the seq advance before
	// writing to disk so the dump's seq matches what we're about to
	// return.
	if advanceSeq {
		r.seq = nextSeq
	}

	if dumpPayload != nil {
		if err := r.dumpSnapshotLocked(dumpPayload); err != nil {
			r.log.Warn("snapshot dump failed", "err", err, "seq", nextSeq)
		}
	}
	return view, nextSeq, nil
}

// dumpSnapshotLocked writes the snapshot payload to the crash-recovery
// directory. Uses CreateTemp + rename for an atomic publish that
// survives crashes without leaving a partial file. Caller MUST hold
// r.mu (or serialize dumps some other way) — CreateTemp gives us a
// unique tmp filename even if we didn't, but the rest of Apply needs
// the mutex anyway.
func (r *Room) dumpSnapshotLocked(payload []byte) error {
	gamesDir := filepath.Join(r.dumpDir, "games")
	if err := os.MkdirAll(gamesDir, 0o755); err != nil {
		return err
	}

	// CreateTemp guarantees a unique filename — even if two goroutines
	// somehow reach this code concurrently (which Apply's mutex
	// prevents today), their writes cannot collide on the tmp path.
	pattern := r.Game.ID.String() + ".*.json.tmp"
	tmp, err := os.CreateTemp(gamesDir, pattern)
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	// If anything below fails, clean up the orphaned tmp file.
	defer func() {
		if tmp != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(payload); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close tmp: %w", err)
	}
	tmp = nil // disable the defer's cleanup now that we're about to rename

	finalPath := filepath.Join(gamesDir, r.Game.ID.String()+".json")
	if err := os.Rename(tmpPath, finalPath); err != nil {
		// Rename failed — the tmp file is still on disk. Remove it
		// explicitly since the defer above was disarmed.
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename tmp: %w", err)
	}
	return nil
}
