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
// # The import rule
//
// This package sits under aiseat/ and is bound by ADR 0033 §3 like
// every other subpackage: no internal/game, ever. It reads
// aiseat.DecisionEvent and writes JSON, and that is all it does.
package decisionlog

import (
	"bufio"
	"compress/gzip"
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
	Applied        bool   `json:"applied"`
	RejectErr      string `json:"reject_err,omitempty"`
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

	Trace     aiseat.Trace `json:"trace"`
	Decision  Decision     `json:"decision"`
	Final     Final        `json:"final"`
	LatencyMS float64      `json:"latency_ms"`
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
			Applied:        ev.Applied,
			RejectErr:      errText(ev.RejectErr),
		},
		LatencyMS: float64(ev.Latency) / float64(time.Millisecond),
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
	return &GameLog{
		path:  path,
		mode:  l.opt.Mode,
		max:   l.opt.MaxBytesPerGame,
		log:   l.opt.Log.With("decision_log", path),
		f:     f,
		w:     bufio.NewWriterSize(f, 64<<10),
		bytes: info.Size(),
	}, nil
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
	Dropped int64  `json:"dropped"`
	Bytes   int64  `json:"bytes"`
}

// GameLog is one game's file. It implements aiseat.DecisionObserver
// and io.Closer.
//
// Every bot seat at the table shares one of these and each calls
// Observe from its own goroutine, so the mutex is doing real work:
// it serialises the four writers AND guarantees that a record is one
// whole line, which is the only thing a JSONL reader can rely on.
type GameLog struct {
	path string
	mode Mode
	max  int64
	log  *slog.Logger

	mu      sync.Mutex
	f       *os.File
	w       *bufio.Writer
	records int64
	dropped int64
	bytes   int64
	warned  bool
	closed  bool
}

// Path is the file being written.
func (g *GameLog) Path() string { return g.path }

// Mode is the fullness this log was opened with.
func (g *GameLog) Mode() Mode { return g.mode }

// Stats snapshots the counters.
func (g *GameLog) Stats() Stats {
	g.mu.Lock()
	defer g.mu.Unlock()
	return Stats{Path: g.path, Records: g.records, Dropped: g.dropped, Bytes: g.bytes}
}

// Observe writes one record.
//
// It never returns an error and never panics a bot seat: a decision
// log that cannot write is a diagnostic that stopped working, and
// stopping the game over it would be a strictly worse outcome than
// losing the diagnostic. Failures are counted and logged once.
func (g *GameLog) Observe(ev aiseat.DecisionEvent) {
	if !Includes(g.mode, ev) {
		return
	}
	line, err := json.Marshal(FromEvent(ev, g.mode))
	if err != nil {
		g.mu.Lock()
		g.dropped++
		g.warnOnceLocked("bot decision log: a record would not marshal; it was dropped", "err", err)
		g.mu.Unlock()
		return
	}
	line = append(line, '\n')

	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		g.dropped++
		return
	}
	if g.max > 0 && g.bytes+int64(len(line)) > g.max {
		g.dropped++
		g.warnOnceLocked("bot decision log has hit its size cap; further records for this game are DROPPED (raise the cap, or use a narrower CMDCTRL_BOT_DECISION_LOG_MODE)",
			"cap_bytes", g.max, "records", g.records)
		return
	}
	n, werr := g.w.Write(line)
	g.bytes += int64(n)
	if werr != nil {
		g.dropped++
		g.warnOnceLocked("bot decision log write failed; records are being dropped", "err", werr)
		return
	}
	g.records++
}

// warnOnceLocked logs at most one WARN per game. Caller holds g.mu.
// One line, because the failure that matters here is a game-long
// condition (the cap, a full disk) and a per-window warning would
// bury the log it is warning about.
func (g *GameLog) warnOnceLocked(msg string, args ...any) {
	if g.warned {
		return
	}
	g.warned = true
	g.log.Warn(msg, args...)
}

// Close flushes and closes the file. Idempotent.
func (g *GameLog) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil
	}
	g.closed = true
	ferr := g.w.Flush()
	cerr := g.f.Close()
	if ferr != nil {
		return ferr
	}
	return cerr
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
