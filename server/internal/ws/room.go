package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
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

	// generation is this room's restore generation (#523, ADR 0044
	// decision 5): the zero value for a room that has never been
	// rebuilt from disk, bumped once by restoreOne every time it IS
	// rebuilt from a restore point, and otherwise constant for the
	// rest of the room's in-memory life — every action just advances
	// seq within the generation the room already has. Guarded by mu
	// like seq, though in practice nothing but restoreOne ever writes
	// it (before the room is registered, so no lock is needed there).
	// Carried on every snapshot frame (protocol.SnapshotPayload) and
	// in every later restore-point write, so a client can tell a
	// server restart's rewind apart from an ordinary dropped frame:
	// within one generation seq is non-decreasing; a generation change
	// means the server rebuilt this room from an earlier point on
	// purpose and the client must discard and re-render.
	generation uint64

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
	// Snapshot does NOT push (it's a no-op for state). Only the top
	// undoSnapshotDepth entries hold a clone, and a bot's commit holds
	// none: those entries are barriers nobody can pop (ADR 0194).
	//
	// Per-entry caller tracking gates Undo: a seated player may only
	// pop an entry whose caller matches their own seat. Admin
	// (Caller == uuid.Nil) bypasses the gate. This prevents one
	// player from rewinding an opponent's action mid-game.
	undoStack []undoEntry

	// subs are the commit observers registered via Subscribe — today
	// the S31 bot runners. Guarded by subMu, never by mu: notify runs
	// AFTER mu is released so an observer may call Apply from its
	// wake without deadlocking. Added in S31 sub-PR 3.
	subMu sync.Mutex
	subs  map[chan struct{}]struct{}

	// lastActed is when each seat last committed something of its own:
	// an action, a bundle or an undo (ADR 0162 §2, "acting resets the
	// clock"). The notifier reads it through LastActed. Guarded by
	// actMu, never by mu, so a read never waits on a commit.
	actMu     sync.Mutex
	lastActed map[uuid.UUID]time.Time

	// pendingAnnotation is consumed by the next captureLocked and
	// written onto the replay line / crash dump it produces. Set
	// under mu by ApplyBundle and cleared by the capture, so it can
	// never bleed onto an unrelated commit. Nil for every ordinary
	// action, which is all of them but bot improvisation.
	pendingAnnotation *protocol.ReplayAnnotation

	// host is the designated table host (ADR 0075 §2.1), set by the
	// lobby through SetHost and moved on by resolveHost when that seat
	// leaves the game. Guarded by hostMu, never by mu, so the lobby
	// can set it from inside an ApplyExternal fn. See host.go.
	hostMu sync.Mutex
	host   uuid.UUID

	// playmats maps a seat to the URL of its owner's playmat (ADR
	// 0128), set by the lobby through SetPlaymat and stamped on every
	// capture. Guarded by playmatMu, never by mu, for the same reason
	// as host. See playmat.go.
	playmatMu sync.Mutex
	playmats  map[uuid.UUID]string
	// playmatWashes is each seat's owner-set wash, stamped beside its
	// URL (ADR 0128 amendment). Same guard.
	playmatWashes map[uuid.UUID]int

	// voice is the table's Discord voice room (ADR 0159 §6), set by the
	// lobby through SetVoice and stamped on every capture. Guarded by
	// voiceMu, never by mu, for the same reason as host. See voice.go.
	voiceMu sync.Mutex
	voice   *protocol.VoiceView

	// rematch is the host's rematch offer on an ended table (#3285, ADR
	// 0164), set by the lobby through SetRematch and stamped on every
	// capture. Guarded by rematchMu, never by mu, for the same reason as
	// host. See rematch.go.
	rematchMu sync.Mutex
	rematch   *protocol.RematchView

	// awayMu guards the away state of ADR 0177 (away.go): which seats
	// a bot is standing in for, and which the table may move on
	// without. The lobby sets both; never guarded by mu, so the lobby
	// may set them from inside an ApplyExternal fn.
	awayMu   sync.Mutex
	standIns map[uuid.UUID]standIn
	movable  map[uuid.UUID]bool
	// stackPos is where each seat dragged the stack panel (#3429, ADR
	// 0176), set by the lobby through SetStackPosition and stamped on
	// every capture. Guarded by stackPosMu, never by mu, for the same
	// reason as host. See stack_position.go.
	stackPosMu sync.Mutex
	stackPos   map[uuid.UUID]protocol.StackPositionView

	// lastRestorePoint is the seq and wall time of this room's most
	// recently WRITTEN restore point (persist.go), so that a shutdown
	// census (#524) can report how far a live table has drifted from
	// what disk holds without re-reading or re-marshalling the file.
	// Guarded by mu, like seq: set only inside publishRestorePointLocked
	// on a successful write, and by restoreOne at boot (before the
	// room is registered, so no lock is needed there). The zero value
	// (a zero At) means "no restore point has ever been written" —
	// legitimate for a brand-new room and read that way by
	// ShutdownReport.
	lastRestorePoint restorePointRecord

	// skips is the per-kind tally of captures that were NOT restore
	// points (ADR 0041 phase 3, P7, #1497): the evidence the shutdown
	// census cannot give, since it reads one instant per deploy and an
	// idle table at that instant hides everything that happened
	// mid-game. Guarded by mu; written only by captureRestorePointLocked.
	skips skipTally

	// The restore-point rate limit and the encoded event log (#3823,
	// restore_rate.go). All guarded by mu except restoreGate, a leaf
	// lock of its own. restoreEvery is how often a restore point may be
	// written (zero: every restorable commit); restoreLastWrite when the
	// last write was attempted; restorePending the newest restorable
	// capture not yet written; restoreTimer the trailing write, with
	// restoreTimerTok telling a stale firing from the current one;
	// restoreWrites counts the files written, for tests; eventLog is
	// each event's JSON, encoded once.
	restoreEvery     time.Duration
	restoreLastWrite time.Time
	restorePending   *pendingRestorePoint
	restoreTimer     *time.Timer
	restoreTimerTok  uint64
	restoreWrites    int
	restoreGate      restoreGate
	eventLog         game.EncodedEventLog

	// tableRolls is the hub's rate limit on roll_table_die (ADR 0121
	// §5, table_roll.go): when each seat last rolled at the table.
	// Guarded by its own mutex, never by mu. In memory only; a restart
	// forgets it, which costs at most one extra roll.
	tableRolls tableRollLimiter

	// chat is the last few chat lines, for a player report's excerpt
	// (chat_recent.go). Its own mutex; in memory only.
	chat chatRing

	// loopBudget overrides game.MaxLoopReplayDispatches for this room's
	// loop replays (loop_replay.go); zero is the default. Set by tests.
	loopBudget int
}

// restorePointRecord is Room's bookkeeping about its own last written
// restore point. See the lastRestorePoint field comment above.
type restorePointRecord struct {
	Seq uint64
	At  time.Time
}

// skipTally counts the captures captureRestorePointLocked skipped, by
// census kind, and the runs of consecutive skips between two restorable
// captures. In-memory only; it dies with the process, which is
// fine, because the shutdown census logs it on the way out.
type skipTally struct {
	// ByKind is skipped captures per census counter name
	// (game.ContinuationCensus.Kinds). A capture blocked by two kinds
	// counts once under each.
	ByKind map[string]int
	// Run is the current run of consecutive skipped captures, in
	// actions; zero right after a restorable capture.
	Run int
	// LongestRun is the longest run this process has seen.
	LongestRun int
}

// noteSkip records one skipped capture.
func (s *skipTally) noteSkip(c game.ContinuationCensus) {
	if s.ByKind == nil {
		s.ByKind = map[string]int{}
	}
	for kind := range c.Kinds() {
		s.ByKind[kind]++
	}
	s.Run++
	if s.Run > s.LongestRun {
		s.LongestRun = s.Run
	}
}

// noteRestorable records a capture that is a restore point, written or
// pending (restore_rate.go): the current run ends.
func (s *skipTally) noteRestorable() { s.Run = 0 }

// copyByKind returns a copy of the per-kind map for a report.
func (s *skipTally) copyByKind() map[string]int {
	if len(s.ByKind) == 0 {
		return nil
	}
	out := make(map[string]int, len(s.ByKind))
	for k, v := range s.ByKind {
		out[k] = v
	}
	return out
}

// undoEntry is one slot on Room.undoStack — the pre-action game
// state plus the seat ID of the player who triggered the action.
// Caller == uuid.Nil means the action came from an admin / spectator
// session and the entry is undoable by any seated player or admin.
type undoEntry struct {
	// pre is the game before the commit, or nil for a barrier: an
	// entry nobody can pop, admin included (ADR 0194). barrier says
	// why it holds no clone.
	pre     *game.Game
	barrier undoBarrier
	caller  uuid.UUID
	// freeUndo exempts this entry from the undoing player's
	// UndosRemaining budget. Set only for entries a player is
	// cleaning up after somebody else — today, a bot improvisation
	// (ADR 0033 §8).
	//
	// The budget exists to police the social cost of taking back
	// YOUR OWN move, and it refreshes once per turn. An improvisation
	// is the bot asserting a rules interpretation the engine could
	// not execute; a human correcting it is doing maintenance on a
	// catalog gap, not rewinding their own play. Charging for that
	// would make the careful response (check it, fix it) cost more
	// than the lazy one (let it stand) — backwards for a feature
	// whose entire safety argument is that it is reversible.
	freeUndo bool

	// autoAnswered is the prompt this entry's commit answered for its
	// chooser with their standing answer (ADR 0127 §4), or uuid.Nil.
	// Undoing such an entry reopens the prompt, and the room marks it
	// asked by hand so the next commit does not answer it again (§6).
	// Room state, never persisted.
	autoAnswered uuid.UUID
	// autoAnswerSeq is that answer's EventAutoAnswer seq, which the
	// room stamps on the chooser's view while this entry is the top of
	// the stack (stampUndoAutoAnswerLocked).
	autoAnswerSeq uint64
}

// undoBarrier says why an undo entry holds no clone (ADR 0194). The
// zero value is an entry that holds one.
type undoBarrier uint8

const (
	// barrierNone: the entry holds its pre-commit clone.
	barrierNone undoBarrier = iota
	// barrierBot: a bot seat's commit. Nobody may take back a bot's
	// move, admin included, so the room never clones for one; nothing
	// under it can be reached either (owner decision 1).
	barrierBot
	// barrierDepth: the entry fell below the top undoSnapshotDepth
	// entries and its clone was dropped (owner decision 2).
	barrierDepth
)

// undoStackCap bounds the per-room undo ring. The most common use is
// "I clicked the wrong card, let me back up one or two", not deep
// history rewind. Entries below the top undoSnapshotDepth keep only
// their metadata (who made them), so the cap costs next to nothing.
const undoStackCap = 32

// undoSnapshotDepth is how many entries from the top of the ring keep
// their full pre-commit clone (ADR 0194, K = 8). A clone of a four-seat
// midgame is about 650 KB (BenchmarkCloneMidgame: ~400 cards at
// 1.4 KB), so the ring holds at most about 5 MB a table, and none at a
// table of bots, whose commits leave barriers.
const undoSnapshotDepth = 8

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

// ErrGameOverUndo is returned by Undo when the game has ended and the
// caller is a seated player: the end is final (ADR 0057 Decision 5,
// the owner's answer to question 1). A game that has announced a
// winner, played the win sound and been written to the database does
// not come back. The admin (caller == uuid.Nil) can still undo, for
// mistakes.
var ErrGameOverUndo = errSentinel("ws: the game is over; only the admin can undo past its end")

// ErrUndoPastSecretVote is returned by Undo when the seated caller's
// entry is the commit that revealed a secret vote (ADR 0180 decision
// 4): taking it back would let the last voter vote again knowing every
// other ballot. The admin (caller == uuid.Nil) can still undo.
var ErrUndoPastSecretVote = errSentinel("ws: that action revealed a secret vote and can't be taken back")

// ErrUndoPastBot is returned by Undo when the top entry is a bot's
// commit (ADR 0194 owner decision 1): a bot's move is never taken back,
// by anyone, admin included, and nothing before it can be reached.
var ErrUndoPastBot = errSentinel("ws: a bot has acted since; a bot's move can't be taken back")

// ErrUndoTooDeep is returned by Undo when the top entry is older than
// the last undoSnapshotDepth actions (ADR 0194 owner decision 2), for
// every caller, admin included.
var ErrUndoTooDeep = errSentinel("ws: undo only goes back 8 actions")

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
//
// Apply counts the commit in cmdctrl_actions_total with type "other";
// a caller that knows the action's type uses ApplyAction.
func (r *Room) Apply(caller uuid.UUID, fn func() error) (protocol.GameView, uint64, error) {
	return r.ApplyAction(ActionTag{Actor: caller}, fn)
}

// ApplyAction is Apply for a caller that says what the commit is, for
// cmdctrl_actions_total (ADR 0123 §3): tag.Type is the action's type
// and tag.Actor the seat that acted, which also stamps the undo entry
// exactly as Apply's caller does (uuid.Nil is the admin).
func (r *Room) ApplyAction(tag ActionTag, fn func() error) (protocol.GameView, uint64, error) {
	st := r.startCommit(tag)
	view, seq, err := r.apply(tag.Actor, fn, &st)
	st.record()
	if err == nil {
		r.noteActed(tag.Actor)
		r.notify()
	}
	return view, seq, err
}

func (r *Room) apply(caller uuid.UUID, fn func() error, st *commitStat) (protocol.GameView, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer st.hold()()

	// Stash a pre-mutation clone for the undo stack BEFORE running
	// fn. Done eagerly so a failing fn doesn't grow history with a
	// no-op snapshot. If fn fails we drop the clone on the floor.
	// A bot's commit takes none: it leaves a barrier (ADR 0194).
	barrier := r.undoBarrierFor(caller)
	var pre *game.Game
	if barrier == barrierNone {
		pre = r.Game.Clone()
	}
	if err := fn(); err != nil {
		return protocol.GameView{}, 0, err
	}
	st.applied = true
	r.pushUndoLocked(undoEntry{pre: pre, barrier: barrier, caller: caller})
	return r.autoAnswerThenCaptureLocked()
}

// undoBarrierFor is barrierBot when stamp is a bot seat, whose commits
// leave a barrier rather than a clone (ADR 0194), and barrierNone for
// everyone else: a person, an agent, the admin (uuid.Nil). Caller MUST
// hold r.mu.
func (r *Room) undoBarrierFor(stamp uuid.UUID) undoBarrier {
	if stamp == uuid.Nil {
		return barrierNone
	}
	if bot, _ := r.Game.SeatDriver(stamp); bot {
		return barrierBot
	}
	return barrierNone
}

// pushUndoLocked pushes one entry onto the undo ring. A barrier entry
// drops whatever clone it was handed. The entry that falls below the
// top undoSnapshotDepth drops its clone and becomes a barrier, so the
// ring never holds more than undoSnapshotDepth clones (ADR 0194).
// Caller MUST hold r.mu.
func (r *Room) pushUndoLocked(e undoEntry) {
	if e.barrier != barrierNone {
		e.pre = nil
	}
	r.undoStack = append(r.undoStack, e)
	if n := len(r.undoStack) - undoSnapshotDepth - 1; n >= 0 {
		if old := &r.undoStack[n]; old.barrier == barrierNone {
			old.pre = nil
			old.barrier = barrierDepth
		}
	}
	if len(r.undoStack) > undoStackCap {
		// Drop the oldest entry, shifting the rest down in place.
		r.undoStack = append(r.undoStack[:0], r.undoStack[1:]...)
	}
}

// maxAutoAnswersPerCommit bounds the automatic answers one commit can
// set off. Each answer is one prompt leaving the queue, and a prompt is
// asked only by a resolution or an answer, so a real table never gets
// near it; it is a backstop against a card whose answer queues the same
// question again.
const maxAutoAnswersPerCommit = 64

// autoAnswerThenCaptureLocked answers, one commit each, every prompt a
// seat has a standing answer for (ADR 0127 §4), then captures the
// frame the caller returns. Each automatic answer is its own commit:
// the frame of the commit before it is captured first (a replay line
// and a dump each), then the game is cloned for the undo entry,
// answered, and the entry pushed — stamped with the CHOOSER, so they
// can take it back, and free (owner decision 8), since they did not
// click it. The last frame is returned as the caller's.
//
// It runs under the room lock with no goroutine, timer or subscriber,
// so a prompt raised by any commit — a human's, a bot's, the admin's,
// a lobby step — is answered the same way, also when no socket is
// connected, and a lockstep arena replays the same answers in the same
// order. Caller MUST hold r.mu.
//
// #2881: an active player who passed the turn passes again once a
// commit hands them priority. actions.Dispatch already does that for
// every action; the settle here covers the commits that are not
// actions (a lobby step) and every automatic answer, whose pass is
// part of the answer's commit.
func (r *Room) autoAnswerThenCaptureLocked() (protocol.GameView, uint64, error) {
	r.Game.SettlePassTurn()
	for i := 0; i < maxAutoAnswersPerCommit; i++ {
		choiceID, chooser, ok := r.Game.NextAutoAnswer()
		if !ok {
			break
		}
		if _, _, err := r.captureLocked(true); err != nil {
			return protocol.GameView{}, 0, err
		}
		pre := r.Game.Clone()
		seq, err := r.Game.AutoAnswer(choiceID)
		if err != nil {
			// Should not happen: NextAutoAnswer just named it under the
			// same lock. Put the game back and ask the player instead,
			// so the loop cannot spin on it.
			r.log.Warn("automatic answer failed", "choice", choiceID, "err", err)
			r.rollbackLocked(pre)
			r.Game.MarkAskedByHand(choiceID, game.AskedByHandUndone)
			continue
		}
		r.Game.SettlePassTurn()
		// A bot never has a standing answer, so this is never a
		// barrier today; the clone is taken anyway, for the rollback.
		r.pushUndoLocked(undoEntry{pre: pre, barrier: r.undoBarrierFor(chooser), caller: chooser, freeUndo: true, autoAnswered: choiceID, autoAnswerSeq: seq})
	}
	// ADR 0193: a commit that answered the last seat of a loop proposal
	// (or moved the game past a recording) is followed, under the same
	// lock, by the replay — a commit and an undo entry of its own,
	// captured in the frame returned here.
	r.Game.TendLoopRecording()
	if r.Game.LoopReplayDue() {
		r.replayLoopLocked()
	}
	return r.captureLocked(true)
}

// Bundle is a group of mutations that commit or fail together, and
// that the undo stack treats as ONE entry. S31 sub-PR 8 (ADR 0033 §8)
// is the first caller: a bot improvising an effect the catalog cannot
// execute emits a handful of sandbox verbs, and four verbs that
// half-apply leave a board nobody can reason about.
type Bundle struct {
	// Caller stamps the resulting undo entry. uuid.Nil means any
	// seated player (or admin) may pop it — which is what an
	// improvisation wants: the bot is not a person who can be asked
	// to take its move back, so everyone at the table can.
	Caller uuid.UUID

	// Steps run in order under the room lock. The first error rolls
	// the whole bundle back and leaves no undo entry behind.
	Steps []func() error

	// FreeUndo exempts the resulting entry from the undoing player's
	// UndosRemaining budget — see undoEntry.freeUndo.
	FreeUndo bool

	// Annotation tags the replay line this commit produces. Nil
	// leaves the line untagged.
	Annotation *protocol.ReplayAnnotation

	// Actor is the seat whose commit this is, for the seat_kind of
	// cmdctrl_actions_total (a bot's improvisation names the bot's
	// seat); uuid.Nil is the admin. It is not the undo stamp, which is
	// Caller: an improvisation is nil-stamped so anyone may undo it.
	Actor uuid.UUID
}

// ErrEmptyBundle is returned by ApplyBundle with no steps. A bundle
// that does nothing should not mint an undo entry.
var ErrEmptyBundle = errSentinel("ws: bundle has no steps")

// ApplyBundle runs every step of b under a single hold of the room
// lock, as one atomic commit: one sequence number, one snapshot, one
// replay line, one undo entry.
//
// Atomicity is the point. Apply's contract is that a failing fn leaves
// no undo entry — true for a single dispatch, which either mutates or
// does not, but false the moment fn makes several mutations and the
// third one fails. ApplyBundle closes that by restoring the
// pre-bundle clone (the same mechanism Undo uses) on any step's
// error, so the caller sees the bundle as all-or-nothing and the
// table never sees a half-applied effect.
//
// On success the whole bundle reverts as a single Undo.
//
// The bundle counts once in cmdctrl_actions_total, as type "bundle",
// with b.Actor's seat kind.
func (r *Room) ApplyBundle(b Bundle) (protocol.GameView, uint64, error) {
	st := r.startCommit(ActionTag{Actor: b.Actor, fixedType: metrics.ActionTypeBundle})
	view, seq, err := r.applyBundle(b, &st)
	st.record()
	if err == nil {
		r.noteActed(b.Actor)
		r.notify()
	}
	return view, seq, err
}

func (r *Room) applyBundle(b Bundle, st *commitStat) (protocol.GameView, uint64, error) {
	if len(b.Steps) == 0 {
		return protocol.GameView{}, 0, ErrEmptyBundle
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	defer st.hold()()

	// A bundle always clones, because a failing step rolls back to it;
	// under a bot's stamp the entry drops it once the bundle commits
	// (ADR 0194). An improvisation is nil-stamped and keeps it.
	barrier := r.undoBarrierFor(b.Caller)
	pre := r.Game.Clone()
	for i, step := range b.Steps {
		if step == nil {
			r.rollbackLocked(pre)
			return protocol.GameView{}, 0, fmt.Errorf("ws: bundle step %d is nil", i)
		}
		if err := step(); err != nil {
			// Roll the earlier steps back before returning. Without
			// this the caller gets an error AND a partly-mutated
			// game with no undo entry pointing at it.
			r.rollbackLocked(pre)
			return protocol.GameView{}, 0, fmt.Errorf("ws: bundle step %d: %w", i, err)
		}
	}

	st.applied = true
	r.pushUndoLocked(undoEntry{pre: pre, barrier: barrier, caller: b.Caller, freeUndo: b.FreeUndo})
	r.pendingAnnotation = b.Annotation
	return r.autoAnswerThenCaptureLocked()
}

// rollbackLocked restores the game to a pre-mutation clone. Caller
// MUST hold r.mu. Same restore path as Undo, minus the stack and
// budget bookkeeping — a rolled-back bundle never happened, so
// nothing about it is recorded.
func (r *Room) rollbackLocked(pre *game.Game) {
	r.Game.WithWriteLock(func() {
		r.Game.RestoreFrom(pre)
	})
}

// ApplyExternal is Apply for lobby-side (HTTP) mutations — join, deck
// upload, start. It runs fn under the room lock and captures seq +
// view exactly like Apply, but records no undo entry: these are game
// setup steps, not player actions a seat may take back. The returned
// view + seq must still reach connected clients — see
// Hub.BroadcastState; a lobby mutation that skips this path is
// invisible to anyone already sitting on the game page.
//
// ApplyExternal counts the commit in cmdctrl_actions_total as a lobby
// setup step: type "lobby", seat_kind "human". A caller dispatching an
// action that mints no undo entry uses ApplyExternalAction.
func (r *Room) ApplyExternal(fn func() error) (protocol.GameView, uint64, error) {
	return r.ApplyExternalAction(ActionTag{fixedType: metrics.ActionTypeLobby, seatKind: metrics.SeatKindHuman}, fn)
}

// ApplyExternalClearingUndo is ApplyExternal for a mutation that must
// not be taken back: on success it empties the undo ring. An admin
// removing a seat from a started game (#3093) is that case. Every
// entry already on the ring holds a clone from before the seat left,
// so undoing one would restore the removed player to the table, and
// their session would be refused while their seat played on.
func (r *Room) ApplyExternalClearingUndo(fn func() error) (protocol.GameView, uint64, error) {
	return r.ApplyExternal(func() error {
		if err := fn(); err != nil {
			return err
		}
		r.undoStack = nil
		return nil
	})
}

// ApplyExternalAction is ApplyExternal for an action that mints no undo
// entry (actions.MintsNoUndo, the table settings), tagged as
// ApplyAction tags one.
func (r *Room) ApplyExternalAction(tag ActionTag, fn func() error) (protocol.GameView, uint64, error) {
	st := r.startCommit(tag)
	view, seq, err := r.applyExternal(fn, &st)
	st.record()
	if err == nil {
		r.notify()
	}
	return view, seq, err
}

func (r *Room) applyExternal(fn func() error, st *commitStat) (protocol.GameView, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer st.hold()()

	if err := fn(); err != nil {
		return protocol.GameView{}, 0, err
	}
	st.applied = true
	return r.autoAnswerThenCaptureLocked()
}

// Subscribe registers a commit observer and returns its wake channel
// plus an unsubscribe func. The channel is edge-triggered with
// capacity 1 and a non-blocking send: a subscriber that is busy sees
// ONE wake for any number of commits that happened meanwhile and is
// expected to re-read the room (Snapshot, or the game directly) on
// wake rather than trust any payload — there is none. Fired after
// every successful Apply / ApplyExternal / Undo, after r.mu has been
// released, so a subscriber may call back into the room from its wake.
// Added in S31 sub-PR 3 for the bot runners.
func (r *Room) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	r.subMu.Lock()
	if r.subs == nil {
		r.subs = make(map[chan struct{}]struct{})
	}
	r.subs[ch] = struct{}{}
	r.subMu.Unlock()
	return ch, func() {
		r.subMu.Lock()
		delete(r.subs, ch)
		r.subMu.Unlock()
	}
}

// noteActed stamps seat's latest commit for LastActed. uuid.Nil (the
// admin, the lobby) is nobody's seat and is not stamped.
func (r *Room) noteActed(seat uuid.UUID) {
	if seat == uuid.Nil {
		return
	}
	r.actMu.Lock()
	defer r.actMu.Unlock()
	if r.lastActed == nil {
		r.lastActed = make(map[uuid.UUID]time.Time)
	}
	r.lastActed[seat] = time.Now()
}

// LastActed is when seat last committed an action, a bundle or an undo
// in this process, and zero if it has not (ADR 0162 §2).
func (r *Room) LastActed(seat uuid.UUID) time.Time {
	r.actMu.Lock()
	defer r.actMu.Unlock()
	return r.lastActed[seat]
}

// notify wakes every subscriber. Non-blocking: a full channel means a
// wake is already pending for that subscriber, which is all an
// edge-trigger promises. Caller must NOT hold r.mu.
func (r *Room) notify() {
	r.subMu.Lock()
	defer r.subMu.Unlock()
	for ch := range r.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
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
// cursor enters that player's untap step. Admin bypasses the budget,
// and so does any entry flagged freeUndo — a bot improvisation, which
// a human undoes as maintenance rather than as a take-back.
//
// Returns ErrNothingToUndo (empty stack), ErrGameOverUndo (the game
// has ended and the caller is not the admin — ADR 0057), ErrUndoPastBot
// or ErrUndoTooDeep (the top entry is a barrier, for every caller —
// ADR 0194), ErrNotYourUndo (top entry belongs to another seat), or
// game.ErrNoUndosRemaining (budget at 0).
// Bumps seq on success so connected clients see a regular snapshot
// frame and don't have to special-case the rewind. The undo itself
// is NOT pushed onto the stack — no redo at v1.
func (r *Room) Undo(caller uuid.UUID) (protocol.GameView, uint64, error) {
	view, seq, err := r.undo(caller)
	if err == nil {
		r.noteActed(caller)
		r.notify()
	}
	return view, seq, err
}

func (r *Room) undo(caller uuid.UUID) (protocol.GameView, uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.undoStack) == 0 {
		return protocol.GameView{}, 0, ErrNothingToUndo
	}
	top := r.undoStack[len(r.undoStack)-1]

	// ADR 0057: the end is final for everyone but the admin. Checked
	// before the caller gate, because whoever took the last action —
	// often the loser, passing priority into an effect win — owns the
	// entry that ended the game.
	if caller != uuid.Nil && r.Game.CurrentState() == game.StateEnded {
		return protocol.GameView{}, 0, ErrGameOverUndo
	}

	// ADR 0194: a barrier holds no clone, so nobody can pop it, the
	// admin included. Checked before the caller gate, because "a bot
	// has acted since" is the reason that will not go away.
	if err := top.barrier.err(); err != nil {
		return protocol.GameView{}, 0, err
	}

	// Caller gate. Admin (uuid.Nil) bypasses; otherwise the top
	// entry's caller must match.
	if caller != uuid.Nil && top.caller != uuid.Nil && top.caller != caller {
		return protocol.GameView{}, 0, ErrNotYourUndo
	}
	if caller != uuid.Nil && r.revealedSecretVoteSinceLocked(top) {
		return protocol.GameView{}, 0, ErrUndoPastSecretVote
	}

	// Budget pre-check (peek, don't decrement). We need to know the
	// pre-restore game has budget so the undo isn't half-applied:
	// if SpendUndo would fail, we bail BEFORE touching state.
	// Decrement happens after RestoreFrom so it isn't clobbered by
	// the snapshot's pre-spend budget value.
	//
	// A free entry (today: a bot improvisation) skips the budget
	// entirely — see undoEntry.freeUndo for why cleaning up after a
	// bot should not cost a player their own take-back.
	spendBudget := caller != uuid.Nil && !top.freeUndo
	if spendBudget {
		if !r.Game.HasUndoBudget(caller) {
			return protocol.GameView{}, 0, game.ErrNoUndosRemaining
		}
	}

	r.undoStack = r.undoStack[:len(r.undoStack)-1]
	r.Game.WithWriteLock(func() {
		r.Game.RestoreFrom(top.pre)
	})
	// ADR 0127 §6: an undone automatic answer reopens its prompt, and
	// the player answers it by hand. Without the mark the next commit
	// would answer it again.
	if top.autoAnswered != uuid.Nil {
		r.Game.MarkAskedByHand(top.autoAnswered, game.AskedByHandUndone)
	}
	// ADR 0193 §2 and owner decision 8: an undo ends any loop
	// recording, since the history it describes is gone. Undoing a
	// replay returns to the demonstrated iteration's end with it spent.
	r.Game.DiscardLoopRecording("an undo was taken")

	// Spend the budget AFTER restore — the restore reset the budget
	// to its pre-action value (which had not yet been spent), so the
	// debit needs to land on top of that.
	if spendBudget {
		if err := r.Game.SpendUndo(caller); err != nil {
			// Should not happen: peek above confirmed budget > 0
			// and we hold r.mu so no concurrent spend could race.
			// Surface as internal so it stands out if it ever fires.
			return protocol.GameView{}, 0, err
		}
	}
	return r.captureLocked(true)
}

// revealedSecretVoteSinceLocked reports whether the live game revealed
// a secret vote after the entry's clone was taken (ADR 0180 decision 4).
// The clone is the room's own and nobody else holds it. Caller MUST
// hold r.mu.
func (r *Room) revealedSecretVoteSinceLocked(e undoEntry) bool {
	if e.pre == nil {
		return false
	}
	return r.Game.RevealedSecretVoteAfter(e.pre.LastEventSeq())
}

// err is the refusal an undo of a barrier entry gets, or nil for an
// entry that holds its clone.
func (b undoBarrier) err() error {
	switch b {
	case barrierBot:
		return ErrUndoPastBot
	case barrierDepth:
		return ErrUndoTooDeep
	}
	return nil
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

// ViewFor returns the view a client seated at viewerID receives: the
// full projection with the room's own stamps (host, undo, playmats),
// then the per-seat filter. A bot runner reads its input here, so it
// sees the same bytes a human at that seat does (aiseat's
// TestBotInputViewIsByteIdenticalToTheHumanView). Unlike Snapshot it
// writes nothing to disk and leaves the pending annotation alone.
func (r *Room) ViewFor(viewerID string) protocol.GameView {
	r.mu.Lock()
	view := protocol.ViewOfGame(r.Game)
	r.stampHostLocked(&view)
	r.stampUndoAutoAnswerLocked(&view)
	r.stampUndoBlockedLocked(&view)
	r.stampPlaymatsLocked(&view)
	r.stampVoiceLocked(&view)
	r.stampRematchLocked(&view)
	r.stampStackPositionsLocked(&view)
	r.stampStandInsLocked(&view)
	r.mu.Unlock()
	return protocol.FilterViewFor(view, viewerID)
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

// Generation returns the room's current restore generation (#523).
// Cheap, like Seq — a plain field read under the same mutex — and
// safe to call on every broadcast: the value only ever changes once,
// at restore, before the room is registered and visible to callers.
func (r *Room) Generation() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.generation
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
	r.stampHostLocked(&view)
	r.stampUndoAutoAnswerLocked(&view)
	r.stampUndoBlockedLocked(&view)
	r.stampPlaymatsLocked(&view)
	r.stampVoiceLocked(&view)
	r.stampRematchLocked(&view)
	r.stampStackPositionsLocked(&view)
	r.stampStandInsLocked(&view)

	// Consume any annotation the committing caller left for this
	// capture. Cleared unconditionally — including on the Snapshot
	// path, which does not advance seq — so a stale tag can never
	// attach itself to a later, unrelated line.
	annotation := r.pendingAnnotation
	r.pendingAnnotation = nil

	// If crash recovery is enabled, marshal the full (unfiltered)
	// payload now, BEFORE committing the seq advance. A marshal
	// failure must not leave r.seq advanced.
	var dumpPayload []byte
	if r.dumpDir != "" {
		payload, err := json.Marshal(protocol.SnapshotPayload{
			Seq:        nextSeq,
			Generation: r.generation,
			Game:       view,
			Annotation: annotation,
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
	//
	// The write itself is rate-limited (restore_rate.go, #3823): the
	// capture is taken here, and written now or within the interval.
	if r.dumpDir != "" && advanceSeq {
		r.noteRestorePointLocked(nextSeq)
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
	defer func() { _ = f.Close() }()
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

// stampUndoAutoAnswerLocked marks the chooser's seat with the seq of
// their automatic answer while it is the top undo entry (ADR 0127 §6,
// the owner's 2026-10-07 amendment), so the client knows exactly when
// its notice's Undo would take that answer back. Room state, like the
// host: the engine does not know the undo stack. Caller MUST hold r.mu.
func (r *Room) stampUndoAutoAnswerLocked(view *protocol.GameView) {
	if len(r.undoStack) == 0 {
		return
	}
	top := r.undoStack[len(r.undoStack)-1]
	if top.pre == nil || top.autoAnswered == uuid.Nil || top.autoAnswerSeq == 0 {
		return
	}
	id := top.caller.String()
	for i := range view.Seats {
		if view.Seats[i].ID == id {
			view.Seats[i].UndoAutoAnswer = top.autoAnswerSeq
		}
	}
}

// stampUndoBlockedLocked tells each seat why an `undo` from it would be
// refused by the stack, using the same gates as undo(): an empty stack,
// the game's end (checked before the caller gate, ADR 0057), a barrier
// (a bot's move, or past the snapshot depth, ADR 0194), and a top
// entry some other seat owns. Room state, like the host: the engine
// does not know the undo stack. Caller MUST hold r.mu.
func (r *Room) stampUndoBlockedLocked(view *protocol.GameView) {
	ended := r.Game.CurrentState() == game.StateEnded
	var top undoEntry
	revealed := false
	if len(r.undoStack) > 0 {
		top = r.undoStack[len(r.undoStack)-1]
		revealed = r.revealedSecretVoteSinceLocked(top)
	}
	for i := range view.Seats {
		id, err := uuid.Parse(view.Seats[i].ID)
		switch {
		case len(r.undoStack) == 0:
			view.Seats[i].UndoBlocked = protocol.UndoBlockedNothing
		case ended:
			view.Seats[i].UndoBlocked = protocol.UndoBlockedGameOver
		case top.barrier == barrierBot:
			view.Seats[i].UndoBlocked = protocol.UndoBlockedPastBot
		case top.barrier == barrierDepth:
			view.Seats[i].UndoBlocked = protocol.UndoBlockedTooDeep
		case err == nil && top.caller != uuid.Nil && top.caller != id:
			view.Seats[i].UndoBlocked = protocol.UndoBlockedNotYours
		case revealed:
			view.Seats[i].UndoBlocked = protocol.UndoBlockedSecretVote
		}
	}
}
