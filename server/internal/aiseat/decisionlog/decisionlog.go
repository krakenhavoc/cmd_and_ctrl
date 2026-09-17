// Package decisionlog writes one JSONL file per game recording what
// every bot seat was asked, what it decided, and what happened to the
// answer.
//
// # What it is for
//
// ADR 0033 claims a funnel that absorbs most windows, a heuristic
// good enough to fall back to, and a model that plays better than the
// heuristic when it is reached. None of those three is measured
// today. The obstacle is not the maths — it is that the evidence
// (the prompt, the reply, the ranking Layer B threw away, the reason
// the runner overruled the policy) exists only as locals inside one
// decision and is gone a microsecond later. This package is where it
// goes instead.
//
// A record is REPLAYABLE, which is the property that matters:
// rules.Resolve and heuristic.Decide are pure functions of an
// aiseat.Input, so a record with the Input in it can be re-decided
// offline, years later, by a binary that never touched a game. That
// is what makes a log of these a position suite, a regression gate
// and a blunder review rather than a pile of prose.
//
// # Hidden information, and why this is operator-only
//
// Each record's Input.View is the seat's OWN filtered view — the same
// bytes a human in that chair receives — so no single record leaks
// anything its seat could not see. The FILE is the problem: it
// aggregates every bot seat at the table, so reading it end to end
// shows four hands at once.
//
// Therefore: the log is **off by default**, it is written under a
// directory the operator names (CMDCTRL_BOT_DECISION_LOG), it is
// NEVER served over HTTP, and it is never attached to a bug report.
// Turning it on for a table with humans in it records the bots' views
// of that table, which is public information plus the bots' own
// hands; turning it on is still a decision an operator makes
// deliberately.
//
// # Reading a file back
//
// Scan walks the lines in order and hands each Record to a callback.
// One thing a reader has to know: the static system block — the rules
// primer and the deck list, several KiB, byte-identical on every
// window of a game — is written ONCE per distinct block. Every record
// carries SystemHash; only the first record with a given hash carries
// Trace.Prompt.System. A reader that wants the system text for a later
// record looks up its SystemHash among the ones it has already seen,
// which is one map and the natural shape of a sequential scan anyway.
// Trace.Prompt.User, the per-decision half, is always present.
//
// # The import rule
//
// This package sits under aiseat/ and is bound by ADR 0033 §3 like
// every other subpackage: no internal/game, ever. It reads
// aiseat.DecisionEvent and writes JSON, and that is all it does.
package decisionlog

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// RecordVersion is the schema version stamped on every line. Bump it
// when a field changes meaning; a reader that does not recognise a
// version should say so rather than guess.
const RecordVersion = 1

// DefaultMaxBytesPerGame bounds one game's file. A four-seat game is
// a few thousand windows and a full record carries the whole view, so
// `all` mode runs to tens of MiB; 256 MiB is headroom for a long game
// and a hard stop before a stuck table fills a disk.
const DefaultMaxBytesPerGame int64 = 256 << 20

// scanTokenMax is the largest line Scan will read. A record in `all`
// mode holds a four-player board view plus a prompt, which is
// comfortably past bufio's 64 KiB default; 64 MiB is the same cap
// random_game_test.go uses on replay frames.
const scanTokenMax = 64 << 20

// Mode decides how much of each window is written.
type Mode string

const (
	// ModeEscalated is the default and the one to leave on. Every
	// window is recorded, but the full Input — the board view, which
	// is almost all of the bytes — only for windows that LEFT Layer
	// A. The rest keep their move list and their trace, which is
	// enough to count with and to find the window again, and drop the
	// view, which for a forced pass says nothing anybody will read.
	ModeEscalated Mode = "escalated"
	// ModeAll writes the full Input for every window. Use it to
	// harvest positions or to chase a bug through a whole game; it is
	// roughly five times the size.
	ModeAll Mode = "all"
	// ModeModel writes ONLY the windows that reached a model, with
	// their full Input. It is the mode for reviewing a model tier's
	// play: the windows Layer A and Layer B settled are not what a
	// model tier is being judged on.
	ModeModel Mode = "model"
)

// ParseMode validates a configured mode. Empty is ModeEscalated.
func ParseMode(s string) (Mode, error) {
	switch m := Mode(strings.ToLower(strings.TrimSpace(s))); m {
	case "":
		return ModeEscalated, nil
	case ModeEscalated, ModeAll, ModeModel:
		return m, nil
	default:
		return "", fmt.Errorf("decisionlog: unknown mode %q (want %s, %s or %s)",
			s, ModeEscalated, ModeAll, ModeModel)
	}
}

// --- the record ------------------------------------------------------

// Decision is what the policy returned, before the runner applied any
// of its own rules to it.
type Decision struct {
	Index  int    `json:"index"`
	Reason string `json:"reason,omitempty"`
	// Err is the policy's error as text. An `error` does not survive
	// a round trip through JSON and the message is the whole value.
	Err string `json:"err,omitempty"`
}

// Final is what the runner actually dispatched, and what the engine
// said about it.
type Final struct {
	Index  int    `json:"index"`
	Label  string `json:"label,omitempty"`
	Reason string `json:"reason,omitempty"`
	// RunnerFallback is the runner's own fallback cause — timeout,
	// policy-error, out-of-range, decline-pass, forced-pass,
	// forced-always-legal — empty when the policy's answer was taken
	// as given. It is NOT Trace.Fallback, which is the funnel's
	// internal classification of a model failure.
	RunnerFallback string `json:"runner_fallback,omitempty"`
	// Forced is what the runner did INSTEAD of dispatching the answer
	// it had: forced-pass, forced-always-legal, no-legal-answer,
	// cancelled. Empty when it dispatched. It is a second axis, not a
	// refinement of RunnerFallback — a window can time out and then be
	// forced onto the always-legal answer, and both facts matter.
	Forced    string `json:"forced,omitempty"`
	Applied   bool   `json:"applied"`
	RejectErr string `json:"reject_err,omitempty"`
}

// Record is one decision window on disk.
//
// Input and Moves are alternatives: a full record carries Input (with
// the move list inside it) and a compact one carries Moves alone. A
// reader that wants to replay a window needs Input; a reader that
// only wants to count needs neither.
type Record struct {
	V          int    `json:"v"`
	Game       string `json:"game"`
	Seat       string `json:"seat"`
	SeatIndex  int    `json:"seat_index"`
	Seq        uint64 `json:"seq,omitempty"`
	Turn       int    `json:"turn"`
	Step       string `json:"step,omitempty"`
	ActiveSeat int    `json:"active_seat"`
	Priority   int    `json:"priority"`
	Policy     string `json:"policy,omitempty"`

	Input *aiseat.Input `json:"input,omitempty"`
	Moves []legal.Move  `json:"moves,omitempty"`

	Trace    aiseat.Trace `json:"trace"`
	Decision Decision     `json:"decision"`
	Final    Final        `json:"final"`
	// SystemHash identifies the static system block this window was
	// sent, empty when no model call was assembled. It is present on
	// every model record; the block's TEXT is present only on the
	// first record that carries this hash — see the package doc.
	SystemHash string  `json:"system_hash,omitempty"`
	LatencyMS  float64 `json:"latency_ms"`
}

// Escalated reports whether this window left Layer A — the windows a
// model tier is actually judged on, and the ones worth a full record.
func (r Record) Escalated() bool { return r.Trace.Layer != "A" }

// Includes reports whether mode writes this window at all.
func Includes(mode Mode, ev aiseat.DecisionEvent) bool {
	if mode != ModeModel {
		return true
	}
	return ev.Trace.Model != "" || ev.Trace.Prompt != nil
}

// FromEvent projects one runner event into a record, at the fullness
// mode asks for. Callers gate on Includes first.
func FromEvent(ev aiseat.DecisionEvent, mode Mode) Record {
	turn := ev.Input.View.Turn
	rec := Record{
		V:          RecordVersion,
		Game:       ev.Game.String(),
		Seat:       ev.Seat.String(),
		SeatIndex:  seatIndex(ev),
		Seq:        ev.Seq,
		Turn:       turn.Number,
		Step:       turn.Step,
		ActiveSeat: turn.ActiveSeat,
		Priority:   turn.PriorityHolder,
		Policy:     ev.Policy,
		Trace:      ev.Trace,
		Decision:   Decision{Index: ev.Decision.Index, Reason: ev.Decision.Reason, Err: errText(ev.DecisionErr)},
		Final: Final{
			Index:          ev.Index,
			Label:          ev.Label,
			Reason:         ev.Reason,
			RunnerFallback: ev.Fallback,
			Forced:         ev.Forced,
			Applied:        ev.Applied,
			RejectErr:      errText(ev.RejectErr),
		},
		SystemHash: SystemHash(ev.Trace.Prompt),
		LatencyMS:  float64(ev.Latency) / float64(time.Millisecond),
	}
	full := mode == ModeAll || mode == ModeModel || rec.Escalated()
	if full {
		in := ev.Input
		rec.Input = &in
	} else {
		rec.Moves = ev.Input.Moves
	}
	return rec
}

// SystemHash identifies a static system block by content. Empty for a
// window that assembled no prompt.
//
// It is short (16 hex characters, 64 bits) because it is an identity
// within ONE game file, not a cryptographic commitment: a collision
// would need two different static blocks in the same game, and the
// static block is fixed per seat at construction.
func SystemHash(p *aiseat.Prompt) string {
	if p == nil || len(p.System) == 0 {
		return ""
	}
	h := sha256.New()
	for _, b := range p.System {
		_, _ = h.Write([]byte(b))
		// A separator, so ["ab","c"] and ["a","bc"] are not the same
		// block. They never are in practice; a hash that says they
		// are is still wrong.
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func seatIndex(ev aiseat.DecisionEvent) int {
	me := ev.Seat.String()
	for i, s := range ev.Input.View.Seats {
		if s.ID == me {
			return i
		}
	}
	return -1
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// --- the writer ------------------------------------------------------

// Options configure a Logger.
type Options struct {
	// Dir is where the per-game files go. Created 0700 if missing.
	Dir string
	// Mode is the fullness. Empty is ModeEscalated.
	Mode Mode
	// MaxBytesPerGame caps one game's file. Zero takes
	// DefaultMaxBytesPerGame; negative disables the cap.
	MaxBytesPerGame int64
	// Log receives the one WARN a game gets when it hits the cap.
	Log *slog.Logger
}

// Logger opens one GameLog per game.
type Logger struct {
	opt Options
}

// New validates the options and creates the directory.
func New(opt Options) (*Logger, error) {
	if strings.TrimSpace(opt.Dir) == "" {
		return nil, errors.New("decisionlog: no directory configured")
	}
	mode, err := ParseMode(string(opt.Mode))
	if err != nil {
		return nil, err
	}
	opt.Mode = mode
	if opt.MaxBytesPerGame == 0 {
		opt.MaxBytesPerGame = DefaultMaxBytesPerGame
	}
	if opt.Log == nil {
		opt.Log = slog.Default()
	}
	// 0700: the file holds four seats' views of one table. Nothing
	// else on the box has business reading it.
	if err := os.MkdirAll(opt.Dir, 0o700); err != nil {
		return nil, fmt.Errorf("decisionlog: create %s: %w", opt.Dir, err)
	}
	return &Logger{opt: opt}, nil
}

// Dir is where logs are written.
func (l *Logger) Dir() string { return l.opt.Dir }

// Mode is the configured fullness.
func (l *Logger) Mode() Mode { return l.opt.Mode }

// Path is where this game's log will be written.
func Path(dir string, gameID uuid.UUID) string {
	return filepath.Join(dir, gameID.String()+".decisions.jsonl")
}

// OpenGame creates the log file for one game.
func (l *Logger) OpenGame(gameID uuid.UUID) (*GameLog, error) {
	path := Path(l.opt.Dir, gameID)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("decisionlog: open %s: %w", path, err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("decisionlog: stat %s: %w", path, err)
	}
	return newGameLog(path, l.opt.Mode, l.opt.MaxBytesPerGame,
		l.opt.Log.With("decision_log", path), f, info.Size()), nil
}

// DecisionLogger adapts l to aiseat.DecisionLogger, which is the
// interface Manager holds. It exists so that aiseat never imports
// this package — that would be a cycle, since this package reads
// aiseat's types.
func (l *Logger) DecisionLogger() aiseat.DecisionLogger {
	return aiseat.DecisionLoggerFunc(func(gameID uuid.UUID) (aiseat.GameDecisionLog, error) {
		g, err := l.OpenGame(gameID)
		if err != nil {
			// A typed nil in a non-nil interface is the classic way
			// to make `if log != nil` lie. Return the untyped one.
			return nil, err
		}
		return g, nil
	})
}

// Stats is a GameLog's counters.
type Stats struct {
	Path    string `json:"path"`
	Records int64  `json:"records"`
	// Dropped is every record that did not reach the file, by cause:
	// DroppedQueue when the seat goroutines outran the writer,
	// DroppedCap when the per-game byte cap was reached, DroppedError
	// when marshalling or writing failed.
	Dropped      int64 `json:"dropped"`
	DroppedQueue int64 `json:"dropped_queue"`
	DroppedCap   int64 `json:"dropped_cap"`
	DroppedError int64 `json:"dropped_error"`
	Bytes        int64 `json:"bytes"`
}

// systemHashKey is the cheap gate dedupeSystem screens lines with.
var systemHashKey = []byte(`"system_hash":`)

// queueDepth is how many marshalled records may be in flight between
// the seat goroutines and the writer.
//
// 256 is sized against the burst the runner can actually produce:
// MaxActionsPerWake is 64 and four seats can be woken by the same
// commit, so 256 windows back-to-back is the worst case a table
// reaches. Anything shallower drops lines during ordinary play — a
// four-seat burst against a 64-deep queue lost a quarter of them —
// and the corpus these files exist to build is worth the memory.
// The bound is the depth times one marshalled record, so ~15 MiB in
// `all` mode on a four-player board and far less in `escalated`.
const queueDepth = 256

// GameLog is one game's file. It implements aiseat.DecisionObserver
// and io.Closer.
//
// # Why the write is asynchronous
//
// Observe is called inline on a bot seat's own goroutine, between the
// decision and the next one. A synchronous write puts a shared mutex
// and — every time the 64 KiB buffer fills, which one full record
// does on its own — a disk write on that path, so four seats would
// serialise on the disk while a table waited. A diagnostic that slows
// the thing it is diagnosing measures its own overhead.
//
// So Observe marshals (on the caller's goroutine, where the data is
// hot and the cost is microseconds) and hands the bytes to one writer
// goroutine through a bounded channel. A full channel DROPS and
// counts, exactly as the byte cap does: the log is a diagnostic, and
// the alternative — blocking a seat until the disk catches up — is
// the failure this design exists to avoid.
type GameLog struct {
	path string
	mode Mode
	max  int64
	log  *slog.Logger

	queue chan []byte
	done  chan struct{} // closed by the writer when it has drained

	// sendMu guards the send on queue against Close closing it. A
	// sender holds it for reading; Close takes it for writing, so no
	// send can be in flight when the channel closes. This is the one
	// race that would take a bot seat down with a panic rather than
	// losing a diagnostic line.
	sendMu    sync.RWMutex
	sendsShut bool

	closeOnce sync.Once
	closeErr  error

	mu      sync.Mutex
	records int64
	dropped struct{ queue, cap, err int64 }
	bytes   int64
	warned  bool
}

func newGameLog(path string, mode Mode, max int64, log *slog.Logger, f *os.File, size int64) *GameLog {
	g := &GameLog{
		path:  path,
		mode:  mode,
		max:   max,
		log:   log,
		queue: make(chan []byte, queueDepth),
		done:  make(chan struct{}),
		bytes: size,
	}
	go g.write(f)
	return g
}

// Path is the file being written.
func (g *GameLog) Path() string { return g.path }

// Mode is the fullness this log was opened with.
func (g *GameLog) Mode() Mode { return g.mode }

// Stats snapshots the counters. Records and Bytes lag by whatever is
// still in the queue; after Close they are exact.
func (g *GameLog) Stats() Stats {
	g.mu.Lock()
	defer g.mu.Unlock()
	return Stats{
		Path:         g.path,
		Records:      g.records,
		Dropped:      g.dropped.queue + g.dropped.cap + g.dropped.err,
		DroppedQueue: g.dropped.queue,
		DroppedCap:   g.dropped.cap,
		DroppedError: g.dropped.err,
		Bytes:        g.bytes,
	}
}

// Observe hands one record to the writer.
//
// It never blocks a bot seat, never returns an error and never panics
// one: a decision log that cannot keep up is a diagnostic that lost
// some lines, and stopping the game over it would be strictly worse.
func (g *GameLog) Observe(ev aiseat.DecisionEvent) {
	if !Includes(g.mode, ev) {
		return
	}
	// Marshalling happens here, on the caller's goroutine, rather than
	// in the writer: the data is hot, the cost is microseconds, and
	// the event — which holds the seat's whole view — does not then
	// have to stay reachable while the queue drains. It is done
	// BEFORE the send lock so Close never waits on it.
	line, err := json.Marshal(FromEvent(ev, g.mode))
	if err != nil {
		g.drop(&g.dropped.err, "bot decision log: a record would not marshal; it was dropped")
		return
	}
	line = append(line, '\n')

	g.sendMu.RLock()
	defer g.sendMu.RUnlock()
	if g.sendsShut {
		// Closed: nothing is draining the queue any more.
		g.drop(&g.dropped.queue, "")
		return
	}
	select {
	case g.queue <- line:
	default:
		g.drop(&g.dropped.queue, "bot decision log cannot keep up with the table; records are being DROPPED (the disk is slow, or the mode is too verbose — try CMDCTRL_BOT_DECISION_LOG_MODE=escalated)")
	}
}

// drop counts one lost record and logs at most one WARN per game.
func (g *GameLog) drop(counter *int64, warn string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	*counter++
	if warn == "" || g.warned {
		return
	}
	g.warned = true
	g.log.Warn(warn, "records", g.records, "cap_bytes", g.max)
}

// write is the single writer goroutine: the only thing that touches
// the file, which is what makes one record one whole line.
//
// It also applies the system-block dedupe. The static half of a
// model prompt is several KiB and byte-identical on every window of a
// game (that is the whole point of a prompt cache), so writing it
// thousands of times is most of the file. It is written once per
// distinct block and thereafter carried by SystemHash alone — see the
// package doc. The dedupe lives here because "first in THIS FILE" is
// a property of the file, and only the writer knows it.
func (g *GameLog) write(f *os.File) {
	defer close(g.done)
	w := bufio.NewWriterSize(f, 64<<10)
	seen := make(map[string]bool, 4)
	for line := range g.queue {
		line = dedupeSystem(line, seen)
		g.mu.Lock()
		over := g.max > 0 && g.bytes+int64(len(line)) > g.max
		g.mu.Unlock()
		if over {
			g.drop(&g.dropped.cap, "bot decision log has hit its size cap; further records for this game are DROPPED (raise the cap, or use a narrower CMDCTRL_BOT_DECISION_LOG_MODE)")
			continue
		}
		n, err := w.Write(line)
		g.mu.Lock()
		g.bytes += int64(n)
		g.mu.Unlock()
		if err != nil {
			g.drop(&g.dropped.err, "bot decision log write failed; records are being dropped")
			continue
		}
		g.mu.Lock()
		g.records++
		g.mu.Unlock()
	}
	// Written before close(g.done), which Close waits on, so the
	// handoff is ordered without another lock.
	if err := w.Flush(); err != nil {
		g.log.Error("bot decision log could not be flushed", "err", err)
		g.closeErr = err
	}
	if err := f.Close(); err != nil {
		g.log.Error("bot decision log could not be closed", "err", err)
		if g.closeErr == nil {
			g.closeErr = err
		}
	}
}

// dedupeSystem strips Trace.Prompt.System from a marshalled record
// whose SystemHash has already been written to this file.
//
// It works on the marshalled bytes rather than the Record because the
// Record was built on a seat's goroutine and the writer is the only
// place that knows what this file has already seen. A record with no
// system block, or the first one with a given hash, passes through
// untouched.
func dedupeSystem(line []byte, seen map[string]bool) []byte {
	// Cheap gate first. Most records are not model windows, carry no
	// system_hash at all (omitempty), and must not cost a JSON parse
	// of a 60 KiB line on the writer goroutine to find that out.
	if !bytes.Contains(line, systemHashKey) {
		return line
	}
	var probe struct {
		SystemHash string `json:"system_hash"`
	}
	if err := json.Unmarshal(line, &probe); err != nil || probe.SystemHash == "" {
		return line
	}
	if !seen[probe.SystemHash] {
		seen[probe.SystemHash] = true
		return line
	}
	var rec Record
	if err := json.Unmarshal(line, &rec); err != nil {
		return line
	}
	if rec.Trace.Prompt == nil || len(rec.Trace.Prompt.System) == 0 {
		return line
	}
	// Copy the prompt before editing it: FromEvent took the Prompt
	// POINTER off the event, so the struct behind it may be shared
	// with the policy's own records.
	p := *rec.Trace.Prompt
	p.System = nil
	rec.Trace.Prompt = &p
	out, err := json.Marshal(rec)
	if err != nil {
		return line
	}
	return append(out, '\n')
}

// Close stops accepting records, drains what is queued, flushes and
// closes the file. Idempotent, and safe to call while seats are still
// observing — a record that races the close is dropped and counted.
func (g *GameLog) Close() error {
	g.closeOnce.Do(func() {
		g.sendMu.Lock()
		g.sendsShut = true
		close(g.queue)
		g.sendMu.Unlock()
		<-g.done
	})
	return g.closeErr
}

// --- reading ---------------------------------------------------------

// Scan reads a decision log and calls fn for each record, in order.
// A `.gz` suffix is decompressed on the way through, so an archived
// log reads the same as a live one. fn returning an error stops the
// scan and Scan returns it.
func Scan(path string, fn func(Record) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	var r io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		zr, zerr := gzip.NewReader(f)
		if zerr != nil {
			return fmt.Errorf("decisionlog: %s: %w", path, zerr)
		}
		defer func() { _ = zr.Close() }()
		r = zr
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 1<<20), scanTokenMax)
	line := 0
	for sc.Scan() {
		line++
		raw := sc.Bytes()
		if len(raw) == 0 {
			continue
		}
		var rec Record
		if err := json.Unmarshal(raw, &rec); err != nil {
			return fmt.Errorf("decisionlog: %s line %d: %w", path, line, err)
		}
		if err := fn(rec); err != nil {
			return err
		}
	}
	return sc.Err()
}
