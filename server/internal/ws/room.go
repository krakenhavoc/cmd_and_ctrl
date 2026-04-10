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
// Room is intentionally membership-free so that a future multi-room
// refactor (S04) doesn't have to reshuffle client tracking.
//
// Room is safe for concurrent use. All mutation + snapshot pairs go
// through Apply, which holds a single mutex across the dispatch,
// the sequence allocation, and the state capture — guaranteeing that
// seq numbers are allocated in the same order state progresses, and
// that no two dispatches can interleave their state captures.
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
	// written via CreateTemp + rename for atomicity.
	dumpDir string
}

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
// view of the resulting game state, marshals a snapshot frame, writes
// the recovery dump if enabled, and returns the frame bytes for
// broadcast. If fn returns an error, Apply returns it verbatim and
// does NOT increment seq or emit a snapshot.
//
// The room mutex is held for the whole mutate → seq → capture → dump
// sequence, so two concurrent clients cannot interleave state and
// cannot observe non-monotonic (seq, state) pairs.
func (r *Room) Apply(fn func() error) ([]byte, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := fn(); err != nil {
		return nil, 0, err
	}
	return r.captureLocked(true)
}

// Snapshot returns a snapshot frame of the current state without
// mutating anything and without advancing the sequence counter. Used
// to send the initial state to a just-joined client. The returned
// frame carries the CURRENT seq — the same value the last Apply
// broadcast to other clients. Under a join-during-action race, a
// new client can briefly see two consecutive frames with the same
// seq (one from Snapshot, one from a subsequent broadcast of the
// same state); both carry identical state and clients should treat
// snapshots idempotently.
//
// This "no-bump" semantics is deliberate: if Snapshot advanced the
// shared counter, existing clients would see a seq jump of +2 on
// their next broadcast (Apply→seq+1 AND Snapshot→seq+1) and flag it
// as a dropped frame.
func (r *Room) Snapshot() ([]byte, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.captureLocked(false)
}

// captureLocked is the common state-capture + marshal path used by
// both Apply and Snapshot. Caller MUST hold r.mu.
//
// If advanceSeq is true, the sequence counter is incremented — but
// only AFTER json.Marshal succeeds, so a marshal failure does not
// leave the counter advanced with no corresponding broadcast.
//
// If advanceSeq is false, the current seq is reused and the counter
// is not touched (see Snapshot).
func (r *Room) captureLocked(advanceSeq bool) ([]byte, uint64, error) {
	nextSeq := r.seq
	if advanceSeq {
		nextSeq = r.seq + 1
	}
	view := protocol.ViewOfGame(r.Game)

	payload, err := json.Marshal(protocol.SnapshotPayload{
		Seq:  nextSeq,
		Game: view,
	})
	if err != nil {
		return nil, 0, err
	}
	frame, err := json.Marshal(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindSnapshot,
		ID:      "",
		Payload: payload,
	})
	if err != nil {
		return nil, 0, err
	}

	// Marshal succeeded. Commit the seq advance (if any) BEFORE the
	// disk dump so the dump's seq matches what we're about to return.
	if advanceSeq {
		r.seq = nextSeq
	}

	if r.dumpDir != "" {
		if err := r.dumpSnapshotLocked(payload); err != nil {
			r.log.Warn("snapshot dump failed", "err", err, "seq", nextSeq)
		}
	}
	return frame, nextSeq, nil
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
