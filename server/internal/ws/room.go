package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"

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

	// undoStack holds pre-mutation Game clones paired with the seat
	// ID of the player whose action produced the post-state. Oldest
	// first; capped at undoStackCap. Each Apply pushes one entry;
	// Snapshot does NOT push (it's a no-op for state).
	//
	// Per-entry caller tracking gates Undo: a seated player may only
	// pop an entry whose caller matches their own seat. Admin
	// (Caller == uuid.Nil) bypasses the gate. This prevents one
	// player from rewinding an opponent's action mid-game.
	undoStack []undoEntry
}

// undoEntry is one slot on Room.undoStack — the pre-action game
// state plus the seat ID of the player who triggered the action.
// Caller == uuid.Nil means the action came from an admin / spectator
// session and the entry is undoable by any seated player or admin.
type undoEntry struct {
	pre    *game.Game
	caller uuid.UUID
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

// ErrNotYourUndo is returned by Undo when the seated caller asked to
// pop a stack entry whose stored caller is a different seat. Casual
// games' "I'd like to take that back" social protocol — you ask the
// opponent to undo their move first if they've acted since you. Admin
// (caller == uuid.Nil) bypasses the check. Added in S11.
var ErrNotYourUndo = errSentinel("ws: top of undo stack is another seat's action")

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
func (r *Room) Apply(caller uuid.UUID, fn func() error) (protocol.GameView, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Stash a pre-mutation clone for the undo stack BEFORE running
	// fn. Done eagerly so a failing fn doesn't grow history with a
	// no-op snapshot. If fn fails we drop the clone on the floor.
	pre := r.Game.Clone()
	if err := fn(); err != nil {
		return protocol.GameView{}, 0, err
	}
	r.undoStack = append(r.undoStack, undoEntry{pre: pre, caller: caller})
	if len(r.undoStack) > undoStackCap {
		// Drop the oldest entry. Trim by reslicing forward — the
		// underlying array's first slot is now unreachable, GC'd on
		// next allocation. Keeping the cap small bounds the leak.
		r.undoStack = append(r.undoStack[:0], r.undoStack[1:]...)
	}
	return r.captureLocked(true)
}

// ApplyExternal is Apply for lobby-side (HTTP) mutations — join, deck
// upload, start. It runs fn under the room lock and captures seq +
// view exactly like Apply, but records no undo entry: these are game
// setup steps, not player actions a seat may take back. The returned
// view + seq must still reach connected clients — see
// Hub.BroadcastState; a lobby mutation that skips this path is
// invisible to anyone already sitting on the game page.
func (r *Room) ApplyExternal(fn func() error) (protocol.GameView, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := fn(); err != nil {
		return protocol.GameView{}, 0, err
	}
	return r.captureLocked(true)
}

// Undo pops the most recent pre-mutation entry off the undo stack,
// restores the live Game to that state, captures a fresh snapshot,
// and returns it for broadcast.
//
// Authorization: a seated caller may only pop an entry whose stored
// caller matches their own seat — preventing one player from rewinding
// an opponent's move mid-game. Admin / spectator (caller == uuid.Nil)
// bypasses both checks; any seated player may undo their own admin-
// stamped entries (entries whose stored caller is uuid.Nil).
//
// Budget: a successful seated-caller undo also debits the caller's
// per-turn UndosRemaining via game.SpendUndo. Refreshed when the
// cursor enters that player's untap step. Admin bypasses the budget.
//
// Returns ErrNothingToUndo (empty stack), ErrNotYourUndo (top entry
// belongs to another seat), or game.ErrNoUndosRemaining (budget at 0).
// Bumps seq on success so connected clients see a regular snapshot
// frame and don't have to special-case the rewind. The undo itself
// is NOT pushed onto the stack — no redo at v1.
func (r *Room) Undo(caller uuid.UUID) (protocol.GameView, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.undoStack) == 0 {
		return protocol.GameView{}, 0, ErrNothingToUndo
	}
	top := r.undoStack[len(r.undoStack)-1]

	// Caller gate. Admin (uuid.Nil) bypasses; otherwise the top
	// entry's caller must match.
	if caller != uuid.Nil && top.caller != uuid.Nil && top.caller != caller {
		return protocol.GameView{}, 0, ErrNotYourUndo
	}

	// Budget pre-check (peek, don't decrement). We need to know the
	// pre-restore game has budget so the undo isn't half-applied:
	// if SpendUndo would fail, we bail BEFORE touching state.
	// Decrement happens after RestoreFrom so it isn't clobbered by
	// the snapshot's pre-spend budget value.
	if caller != uuid.Nil {
		if peekUndoBudget(r.Game, caller) <= 0 {
			return protocol.GameView{}, 0, game.ErrNoUndosRemaining
		}
	}

	r.undoStack = r.undoStack[:len(r.undoStack)-1]
	r.Game.WithWriteLock(func() {
		r.Game.RestoreFrom(top.pre)
	})

	// Spend the budget AFTER restore — the restore reset the budget
	// to its pre-action value (which had not yet been spent), so the
	// debit needs to land on top of that.
	if caller != uuid.Nil {
		if err := r.Game.SpendUndo(caller); err != nil {
			// Should not happen: peek above confirmed budget > 0
			// and we hold r.mu so no concurrent spend could race.
			// Surface as internal so it stands out if it ever fires.
			return protocol.GameView{}, 0, err
		}
	}
	return r.captureLocked(true)
}

// peekUndoBudget returns the named player's current UndosRemaining
// without mutating state. Used by Room.Undo to validate budget
// before doing the more expensive state restore. Calls PlayerByID
// directly (which takes its own read lock) — no nested locking.
func peekUndoBudget(g *game.Game, playerID uuid.UUID) int {
	p := g.PlayerByID(playerID)
	if p == nil {
		return 0
	}
	return p.UndosRemaining
}

// IsErrNothingToUndo reports whether err is one of the expected
// "undo can't proceed" sentinels — used by the hub to map server
// errors to wire codes without importing the game package directly.
func IsErrNothingToUndo(err error) bool { return errors.Is(err, ErrNothingToUndo) }
func IsErrNotYourUndo(err error) bool   { return errors.Is(err, ErrNotYourUndo) }

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

// Seq returns the current sequence counter without capturing a view.
// Cheap — the hub's post-admit re-check uses it to detect whether a
// broadcast landed in the pre-stage→admit gap before paying for a
// full Snapshot (which builds the whole view and rewrites the
// crash-recovery dump).
func (r *Room) Seq() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.seq
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
		// Replay log: every Apply (advanceSeq=true) appends one
		// JSONL line to the per-game replay file. Snapshot reads
		// (advanceSeq=false) are deliberately excluded — they're a
		// no-op on game state and would inflate the log without
		// adding information.
		if advanceSeq {
			if err := r.appendReplayLocked(dumpPayload); err != nil {
				r.log.Warn("replay log append failed", "err", err, "seq", nextSeq)
			}
		}
	}

	// Restore point (see persist.go). Distinct from the dump above:
	// that one is a protocol.GameView for forensics and cannot
	// rebuild a game; this one is the domain snapshot that can.
	//
	// Written only when the resulting state holds no live Go
	// continuations, so what sits on disk is always the most recent
	// state a restart can rebuild EXACTLY. A game mid-prompt skips
	// the write and keeps its previous restore point — the deploy
	// rewinds it rather than resurrecting it wrong.
	//
	// Failure is warn-only for the same reason the dump's is: losing
	// a restore point costs a rewind on the next deploy; refusing the
	// player's action costs them the move they just made.
	if r.dumpDir != "" && advanceSeq {
		if r.Game.CurrentState() == game.StateEnded {
			// Nothing left to resume. Drop the file so every future
			// boot does not rebuild a finished table.
			r.RemoveRestorePoint()
		} else if _, err := r.writeRestorePointLocked(nextSeq); err != nil {
			r.log.Warn("restore point write failed", "err", err, "seq", nextSeq)
		}
	}
	return view, nextSeq, nil
}

// appendReplayLocked appends one JSON line (the marshaled snapshot
// payload + '\n') to the per-game replay log. The crash-recovery dump
// already captures the latest snapshot in a single file; the replay
// log is the additive history equivalent — every successful Apply
// produces exactly one line, in order, so a downstream consumer can
// re-derive the full game timeline by streaming it back.
//
// Format: JSONL (one protocol.SnapshotPayload per line). Open with
// O_APPEND so concurrent writes from different goroutines (which
// shouldn't happen — caller holds r.mu — but defense in depth) get
// atomic POSIX-compliant single-line appends.
//
// Caller MUST hold r.mu. ReplayPath returns the on-disk location.
func (r *Room) appendReplayLocked(payload []byte) error {
	path := replayPath(r.dumpDir, r.Game.ID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("append replay line: %w", err)
	}
	return nil
}

// ReplayPath returns the on-disk path to the replay JSONL for this
// room's game, or an empty string when crash recovery / replay is
// disabled (dumpDir empty). Used by the lobby's GET
// /games/{id}/replay handler to stream the file back to the client.
func (r *Room) ReplayPath() string {
	if r.dumpDir == "" {
		return ""
	}
	return replayPath(r.dumpDir, r.Game.ID)
}

// snapshotPath / replayPath are the single definition of the on-disk
// artifact layout under dumpDir. Room's writers and RoomManager's
// Delete cleanup both go through these, so the layout can't drift
// between the writer and the reaper (a drifted reaper would silently
// stop removing deleted games' files — os.IsNotExist hides the miss).
func snapshotPath(dumpDir string, id uuid.UUID) string {
	return filepath.Join(dumpDir, "games", id.String()+".json")
}

func replayPath(dumpDir string, id uuid.UUID) string {
	return filepath.Join(dumpDir, "replays", id.String()+".jsonl")
}

// dumpSnapshotLocked writes the snapshot payload to the crash-recovery
// directory. Uses CreateTemp + rename for an atomic publish that
// survives crashes without leaving a partial file. Caller MUST hold
// r.mu (or serialize dumps some other way) — CreateTemp gives us a
// unique tmp filename even if we didn't, but the rest of Apply needs
// the mutex anyway.
func (r *Room) dumpSnapshotLocked(payload []byte) error {
	finalPath := snapshotPath(r.dumpDir, r.Game.ID)
	gamesDir := filepath.Dir(finalPath)
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

	if err := os.Rename(tmpPath, finalPath); err != nil {
		// Rename failed — the tmp file is still on disk. Remove it
		// explicitly since the defer above was disarmed.
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename tmp: %w", err)
	}
	return nil
}
