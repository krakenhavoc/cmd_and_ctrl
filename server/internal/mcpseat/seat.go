package mcpseat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Config is how the owner set the seat up. None of it is reachable from a
// tool: the model cannot widen the origin list or the absorbed rules.
type Config struct {
	// Origins is the --allow-origin list. Nil means DefaultOrigins.
	Origins *OriginAllowlist
	// Absorb is the Layer A rules answered without the model (§4). Nil
	// means DefaultAbsorb.
	Absorb map[string]bool
	// StateDir holds the saved sessions. Empty means DefaultStateDir.
	StateDir string
	// LogOutput receives the log. Never stdout, which carries the MCP
	// protocol. Nil means stderr.
	LogOutput io.Writer
	// LogLevel is the log's level.
	LogLevel slog.Level
	// HTTPClient and Dialer are for tests; nil means the defaults.
	HTTPClient *http.Client
	Dialer     *websocket.Dialer
	// AckTimeout bounds act's wait for the server's ack or error (§6.4).
	// Past it, act answers `unknown` and the model reads the state.
	AckTimeout time.Duration
	// Now is the clock; nil means time.Now.
	Now func() time.Time
}

// Limits from §3 and §8.
const (
	maxRejections     = 3
	chatMaxRunes      = 500
	maxCardNameValue  = 200
	defaultWait       = 25
	maxWait           = 50
	firstSnapshotWait = 15 * time.Second
	movesReplyWait    = 15 * time.Second
	sinceLogLines     = 30
	sinceChatLines    = 20
	defaultAckTimeout = 10 * time.Second
	defaultName       = "Agent"
)

// winState is where a decision window stands.
type winState int

const (
	// winPending: being resolved (the full list fetched, Layer A run).
	winPending winState = iota
	// winAuto: answered without the model; waiting for the server.
	winAuto
	// winDecision: a real choice, for the model.
	winDecision
	// winActed: the model's move was accepted.
	winActed
)

// windowKey names the state a window was opened in.
type windowKey struct{ gen, seq uint64 }

func (k windowKey) token() string { return fmt.Sprintf("w%d.%d", k.gen, k.seq) }

// window is one decision window: the seat's full move list in one state.
// Every index a tool shows refers to one window's list.
type window struct {
	key        windowKey
	token      string
	moves      []wireMove
	cuts       []cutReport
	partial    bool
	state      winState
	rejections int
	openedAt   time.Time
	kind       string
}

// reply is a frame answering one of the seat's own frames, by id, or the
// news that the socket dropped before one came.
type reply struct {
	kind    string // kindAck, protocol.KindError, kindLegalMovesReply, or "dropped"
	ack     ackPayload
	err     protocol.ErrorPayload
	payload json.RawMessage
}

const replyDropped = "dropped"

// Seat is the one seat this process holds. One binary, one seat: two
// seats are two MCP server entries (Out of scope).
type Seat struct {
	cfg   Config
	log   *slog.Logger
	sec   *secrets
	api   *api
	store stateStore
	now   func() time.Time

	actions *bucket // 4/s, burst 8 — Layer A's answers count
	chat    *bucket // one line per 5 s
	cards   *bucket // 5 lookups/s

	opMu sync.Mutex // serialises join, set_deck and leave

	mu       sync.Mutex
	client   string
	sess     *savedSession
	conn     *wsConn
	stopAuto chan struct{}
	autoDone chan struct{}
	poke     chan struct{}
	changed  chan struct{} // closed and replaced on every change

	isConnected bool
	endReason   string // why the connection ended for good
	haveView    bool
	view        *protocol.GameView
	key         windowKey
	moves       []wireMove
	truncated   bool

	win          *window
	lastCounted  windowKey
	passUntil    bool
	lastActive   int
	reportedOver bool
	noMoveReq    bool // the server answered legal_moves_request as unknown
	probed       bool // the wire probe has been sent for this seat

	pending map[string]chan reply
	moveSem chan struct{}

	autoSince   map[string]int
	logSeqShown uint64
	chatSince   []string

	cardCache map[string]*cardMeta
	stats     stats
}

// NewSeat builds a seat from the owner's configuration.
func NewSeat(cfg Config) (*Seat, error) {
	if cfg.Origins == nil {
		al, err := NewOriginAllowlist(nil)
		if err != nil {
			return nil, err
		}
		cfg.Origins = al
	}
	if cfg.Absorb == nil {
		cfg.Absorb = map[string]bool{}
		for _, r := range DefaultAbsorb {
			cfg.Absorb[r] = true
		}
	}
	if cfg.StateDir == "" {
		d, err := DefaultStateDir()
		if err != nil {
			return nil, err
		}
		cfg.StateDir = d
	}
	if cfg.LogOutput == nil {
		cfg.LogOutput = os.Stderr
	}
	if cfg.AckTimeout <= 0 {
		cfg.AckTimeout = defaultAckTimeout
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	sec := newSecrets()
	s := &Seat{
		cfg:        cfg,
		sec:        sec,
		log:        newLogger(cfg.LogOutput, sec, cfg.LogLevel),
		api:        newAPI(cfg.HTTPClient),
		store:      stateStore{root: cfg.StateDir},
		now:        now,
		actions:    newBucket(4, 8, now),
		chat:       newBucket(0.2, 1, now),
		cards:      newBucket(5, 5, now),
		changed:    make(chan struct{}),
		pending:    map[string]chan reply{},
		moveSem:    make(chan struct{}, 1),
		autoSince:  map[string]int{},
		cardCache:  map[string]*cardMeta{},
		stats:      newStats(),
		lastActive: -1,
	}
	return s, nil
}

// Logger is the seat's scrubbing logger, for main's own lines.
func (s *Seat) Logger() *slog.Logger { return s.log }

// SetClient records the MCP client's clientInfo.name, for the badge (§7).
func (s *Seat) SetClient(name string) {
	if name == "" {
		return
	}
	s.mu.Lock()
	s.client = name
	s.mu.Unlock()
}

// Scrub removes every credential the seat holds from text. transport.go
// runs every tool result through it.
func (s *Seat) Scrub(text string) string { return s.sec.scrub(text) }

// NoteToolBytes counts what a tool returned, for §9's report.
func (s *Seat) NoteToolBytes(n int) {
	s.mu.Lock()
	s.stats.toolBytes += n
	s.mu.Unlock()
}

// Close leaves the table connection without touching the saved session,
// so a restarted binary can reattach. Called when the MCP client goes.
func (s *Seat) Close() {
	s.mu.Lock()
	conn := s.conn
	stop, done := s.stopAuto, s.autoDone
	s.conn, s.stopAuto, s.autoDone = nil, nil, nil
	s.mu.Unlock()
	if stop != nil {
		close(stop)
		<-done
	}
	if conn != nil {
		conn.close()
	}
}

// broadcastLocked wakes every waiter. Callers hold s.mu.
func (s *Seat) broadcastLocked() {
	close(s.changed)
	s.changed = make(chan struct{})
}

func (s *Seat) pokeAutopilot() {
	if s.poke == nil {
		return
	}
	select {
	case s.poke <- struct{}{}:
	default:
	}
}

// --- frameSink: called on the connection's read goroutine ---------------

func (s *Seat) connected() {
	s.mu.Lock()
	s.isConnected = true
	probe := !s.probed
	s.probed = true
	s.broadcastLocked()
	s.mu.Unlock()
	if probe {
		go s.probeWire()
	}
}

// legacyAckWait is how long act waits on a server that predates ADR 0122
// PR 5, which never acknowledges anything: long enough for an error frame
// to arrive, short enough that a test game is not mostly waiting.
const legacyAckWait = 2 * time.Second

// probeWire asks once, on the first connection, whether the server knows
// legal_moves_request. A server from before PR 5 answers "unknown or
// unsupported kind", and it sends no ack either, so act stops waiting
// the full AckTimeout for one and says why its answer is `unknown`.
// Nothing is inferred from the probe about any move: act still reports
// only an ack or an error as an outcome.
func (s *Seat) probeWire() {
	ctx, cancel := context.WithTimeout(context.Background(), movesReplyWait)
	defer cancel()
	select {
	case s.moveSem <- struct{}{}:
	case <-ctx.Done():
		return
	}
	defer func() { <-s.moveSem }()
	ch, err := s.sendFrame(kindLegalMovesRequest, legalMovesRequest{})
	if err != nil {
		return
	}
	select {
	case r := <-ch:
		if r.kind == string(protocol.KindError) && r.err.Code == protocol.CodeBadRequest &&
			strings.Contains(r.err.Message, "unknown or unsupported kind") {
			s.mu.Lock()
			s.noMoveReq = true
			s.mu.Unlock()
			s.log.Info("this server predates ADR 0122 PR 5: no acks and no full move lists")
		}
	case <-ctx.Done():
	}
}

// ackWait is how long act and concede wait for the server's answer.
func (s *Seat) ackWait() (time.Duration, bool) {
	s.mu.Lock()
	legacy := s.noMoveReq
	s.mu.Unlock()
	if legacy && legacyAckWait < s.cfg.AckTimeout {
		return legacyAckWait, true
	}
	return s.cfg.AckTimeout, legacy
}

func (s *Seat) dropped() {
	s.mu.Lock()
	s.isConnected = false
	for id, ch := range s.pending {
		ch <- reply{kind: replyDropped}
		delete(s.pending, id)
	}
	// An automatic answer lost with the socket is decided again on the
	// first snapshot after the redial; a decision stays the model's.
	if s.win != nil && (s.win.state == winPending || s.win.state == winAuto) {
		s.win = nil
	}
	s.broadcastLocked()
	s.mu.Unlock()
}

func (s *Seat) ended(reason string) {
	s.mu.Lock()
	s.isConnected = false
	s.endReason = reason
	for id, ch := range s.pending {
		ch <- reply{kind: replyDropped}
		delete(s.pending, id)
	}
	s.broadcastLocked()
	s.mu.Unlock()
	s.log.Warn("table connection ended", "reason", reason)
}

func (s *Seat) frame(f protocol.Frame) {
	switch string(f.Kind) {
	case string(protocol.KindSnapshot):
		s.onSnapshot(f.Payload)
	case kindAck:
		var p ackPayload
		_ = json.Unmarshal(f.Payload, &p)
		s.deliver(f.ID, reply{kind: kindAck, ack: p})
	case string(protocol.KindError):
		var p protocol.ErrorPayload
		_ = json.Unmarshal(f.Payload, &p)
		if !s.deliver(f.ID, reply{kind: string(protocol.KindError), err: p}) {
			s.log.Info("server error", "code", p.Code, "message", p.Message)
		}
	case kindLegalMovesReply:
		s.deliver(f.ID, reply{kind: kindLegalMovesReply, payload: f.Payload})
	case string(protocol.KindChat):
		var p protocol.ChatPayload
		if json.Unmarshal(f.Payload, &p) != nil {
			return
		}
		s.mu.Lock()
		if s.sess == nil || p.AuthorID != s.sess.PlayerID.String() {
			s.chatSince = append(s.chatSince, chatLine(p))
			if len(s.chatSince) > sinceChatLines {
				s.chatSince = s.chatSince[len(s.chatSince)-sinceChatLines:]
			}
		}
		s.mu.Unlock()
	}
}

// deliver hands a reply to whoever is waiting on that id.
func (s *Seat) deliver(id string, r reply) bool {
	if id == "" {
		return false
	}
	s.mu.Lock()
	ch, ok := s.pending[id]
	if ok {
		delete(s.pending, id)
	}
	s.mu.Unlock()
	if ok {
		ch <- r
	}
	return ok
}

func (s *Seat) onSnapshot(raw json.RawMessage) {
	var p protocol.SnapshotPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		s.log.Warn("unreadable snapshot", "err", err)
		return
	}
	sm, err := decodeSnapshotMoves(raw)
	if err != nil {
		s.log.Warn("unreadable move list", "err", err)
		return
	}
	s.mu.Lock()
	// docs/protocol.md: seq never falls within a generation; a new
	// generation is a deliberate rewind and is always taken.
	if s.haveView && p.Generation == s.key.gen && p.Seq < s.key.seq {
		s.mu.Unlock()
		return
	}
	view := p.Game
	s.view = &view
	s.haveView = true
	s.key = windowKey{gen: p.Generation, seq: p.Seq}
	s.moves = sm.Game.LegalMoves
	s.truncated = sm.Game.Truncated
	if self := s.selfLocked(); self != nil {
		// The start of the seat's own turn clears pass_until (§4), as
		// the browser's safety belt does.
		if view.Turn.ActiveSeat == self.Seat && s.lastActive != self.Seat {
			s.passUntil = false
		}
	}
	s.lastActive = view.Turn.ActiveSeat
	over := gameOver(&view) && !s.reportedOver
	if over {
		s.reportedOver = true
	}
	sess := s.sess
	s.broadcastLocked()
	s.mu.Unlock()
	if over {
		s.mu.Lock()
		line := s.stats.oneLine()
		s.mu.Unlock()
		s.log.Info("game over", "report", line)
		if sess != nil {
			if err := s.store.remove(sess.Origin, sess.GameID); err != nil {
				s.log.Warn("could not delete the saved session", "err", err)
			}
		}
	}
	s.pokeAutopilot()
}

func gameOver(v *protocol.GameView) bool {
	return v.State == "ended" || v.Outcome != nil
}

// selfLocked is this seat's PlayerView in the current view.
func (s *Seat) selfLocked() *protocol.PlayerView {
	if s.view == nil || s.sess == nil {
		return nil
	}
	return mySeat(s.view, s.sess.PlayerID.String())
}

// --- the autopilot ------------------------------------------------------

// autopilot answers trivial windows as they open (§4), with no MinThink
// delay, the way a browser's autopass does.
func (s *Seat) autopilot(stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-stop
		cancel()
	}()
	for {
		select {
		case <-stop:
			return
		case <-s.poke:
		}
		s.step(ctx)
	}
}

// legalMoves strips the wire extras off a list, for Layer A.
func legalMoves(ws []wireMove) []legal.Move {
	out := make([]legal.Move, len(ws))
	for i := range ws {
		out[i] = ws[i].Move
	}
	return out
}

// step opens a window for the current state, if the seat owes a decision
// and none is open for it yet, and answers it or hands it to the model.
func (s *Seat) step(ctx context.Context) {
	s.mu.Lock()
	if s.sess == nil || !s.haveView || s.view.State != "active" || len(s.moves) == 0 {
		s.mu.Unlock()
		return
	}
	if self := s.selfLocked(); self == nil || self.Eliminated {
		s.mu.Unlock()
		return
	}
	key := s.key
	if s.win != nil && s.win.key == key {
		s.mu.Unlock()
		return
	}
	w := &window{key: key, token: key.token(), moves: append([]wireMove(nil), s.moves...), state: winPending}
	s.win = w
	if key != s.lastCounted {
		s.lastCounted = key
		s.stats.windows++
		if s.truncated {
			s.stats.truncatedWindows++
		}
	}
	truncated, noReq := s.truncated, s.noMoveReq
	s.mu.Unlock()

	// §4: when the frame says the list was cut, the full list is fetched
	// before Layer A or the model sees anything.
	if truncated {
		if noReq {
			s.mu.Lock()
			w.partial = true
			s.mu.Unlock()
		} else {
			rep, err := s.requestMoves(ctx, "")
			s.mu.Lock()
			if s.win != w {
				s.mu.Unlock()
				return
			}
			switch {
			case err == nil && (rep.Generation != key.gen || rep.Seq != key.seq):
				// The board moved while the request was out: the
				// next snapshot opens the next window.
				s.win = nil
				s.mu.Unlock()
				s.pokeAutopilot()
				return
			case err == nil:
				w.moves, w.cuts = rep.Moves, rep.Truncated
			default:
				w.partial = true
				s.log.Info("full move list unavailable; using the cut list", "err", err)
			}
			s.mu.Unlock()
		}
	}

	s.mu.Lock()
	if s.win != w {
		s.mu.Unlock()
		return
	}
	view, me := s.view, s.sess.PlayerID
	ans, auto := decideAutomatically(view, me, legalMoves(w.moves), s.cfg.Absorb, s.passUntil)
	if !auto {
		s.openDecisionLocked(w)
		s.mu.Unlock()
		return
	}
	w.state = winAuto
	s.autoSince[ans.rule]++
	s.stats.absorbed[ans.rule]++
	s.mu.Unlock()

	if err := s.actions.wait(ctx); err != nil {
		return
	}
	ch, err := s.sendAction(w.moves[ans.index], nil)
	if err != nil {
		s.mu.Lock()
		if s.win == w {
			s.win = nil
		}
		s.mu.Unlock()
		s.log.Info("automatic answer not sent", "rule", ans.rule, "err", err)
		return
	}
	go s.awaitAuto(w, ch)
}

// openDecisionLocked hands a window to the model.
func (s *Seat) openDecisionLocked(w *window) {
	w.state = winDecision
	w.openedAt = s.now()
	w.kind = windowKind(s.view, s.sess.PlayerID.String(), w.moves)
	s.stats.shown++
	s.broadcastLocked()
}

// awaitAuto waits for the server's answer to an automatic move. A refusal
// hands the window to the model, which the binary never answers for it.
func (s *Seat) awaitAuto(w *window, ch chan reply) {
	t := time.NewTimer(s.cfg.AckTimeout)
	defer t.Stop()
	var r reply
	select {
	case r = <-ch:
	case <-t.C:
		return
	}
	if r.kind != string(protocol.KindError) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.autoErrors++
	if s.win == w && w.state == winAuto {
		w.rejections++
		s.openDecisionLocked(w)
	}
	s.log.Info("automatic answer refused", "code", r.err.Code, "message", r.err.Message)
}

// sendAction sends a move as an action frame with a fresh id, and returns
// the channel its ack or error will arrive on.
func (s *Seat) sendAction(m wireMove, params json.RawMessage) (chan reply, error) {
	if params == nil {
		params = m.Params
	}
	p := protocol.ActionPayload{Type: m.Type, Params: params}
	if m.Player != uuid.Nil {
		p.Player = m.Player.String()
	}
	return s.sendFrame(string(protocol.KindAction), p)
}

// sendFrame sends a frame that the server answers by id.
func (s *Seat) sendFrame(kind string, payload any) (chan reply, error) {
	id := uuid.NewString()
	ch := make(chan reply, 1)
	s.mu.Lock()
	conn := s.conn
	if conn == nil {
		s.mu.Unlock()
		return nil, errNotConnected
	}
	s.pending[id] = ch
	s.mu.Unlock()
	if err := conn.send(kind, id, payload); err != nil {
		s.mu.Lock()
		delete(s.pending, id)
		s.mu.Unlock()
		return nil, err
	}
	return ch, nil
}

// errNoMoveRequest is an older server's answer to legal_moves_request.
var errNoMoveRequest = errors.New("this server cannot send a full move list (it predates ADR 0122 PR 5)")

// requestMoves asks for the seat's full move list, or one card's moves
// with the caps lifted (§6.1, §6.2). At most one request is in flight.
func (s *Seat) requestMoves(ctx context.Context, source string) (*legalMovesReply, error) {
	select {
	case s.moveSem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-s.moveSem }()
	s.mu.Lock()
	s.stats.fullListRequests++
	s.mu.Unlock()
	ch, err := s.sendFrame(kindLegalMovesRequest, legalMovesRequest{Source: source})
	if err != nil {
		return nil, err
	}
	t := time.NewTimer(movesReplyWait)
	defer t.Stop()
	var r reply
	select {
	case r = <-ch:
	case <-t.C:
		return nil, errors.New("no answer to the move-list request")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	switch r.kind {
	case kindLegalMovesReply:
		var rep legalMovesReply
		if err := json.Unmarshal(r.payload, &rep); err != nil {
			return nil, err
		}
		return &rep, nil
	case string(protocol.KindError):
		if r.err.Code == protocol.CodeBadRequest && strings.Contains(r.err.Message, "unknown or unsupported kind") {
			s.mu.Lock()
			s.noMoveReq = true
			s.mu.Unlock()
			return nil, errNoMoveRequest
		}
		return nil, fmt.Errorf("the server refused the move-list request: %s", r.err.Message)
	default:
		return nil, errNotConnected
	}
}
