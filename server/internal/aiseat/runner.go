package aiseat

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// ErrNoMoves is returned by a Policy asked to decide with nothing to
// decide between. Runners never call Decide with an empty list.
var ErrNoMoves = errors.New("aiseat: no legal moves")

// Broadcaster pushes a committed view to every WS client on a game.
// Satisfied by *ws.Hub. Nil is allowed (tests); the game still
// advances, nobody watching is told.
type Broadcaster interface {
	BroadcastState(gameID uuid.UUID, seq uint64, view protocol.GameView)
}

// Config paces and bounds a runner. DefaultConfig is what production
// uses; the zero value is the same minus pacing (MinThink 0), which is
// what tests want.
type Config struct {
	// MinThink holds a fast decision so the table never feels
	// precognitive. Zero disables pacing; DefaultConfig sets 700ms.
	MinThink time.Duration
	// MaxThink is the hard deadline on Policy.Decide. On expiry the
	// runner takes the pass move (or the first move when passing is
	// not on offer) and logs the miss. Default 2s.
	MaxThink time.Duration
	// MaxActionsPerWake bounds one wake's act-loop so a runaway policy
	// cannot monopolise the room. Default 64.
	MaxActionsPerWake int
	// MaxConsecutiveRejects is how many dispatcher rejections in a row
	// the runner tolerates before forcing a pass. Rejections should be
	// rare — every offered move is sound at enumeration time — and a
	// streak means the board is changing under the bot. Default 3.
	MaxConsecutiveRejects int
	// BlockGrace is how long an ACTIVE bot holds its pass during the
	// declare-blockers step while any defending seat still has a legal
	// block to declare. The engine takes blocks during the step rather
	// than as a turn-based action before priority, so without this a
	// bot would pass, the table would wrap, and the step would end
	// before a human (or a slower bot) had blocked. Zero disables;
	// DefaultConfig sets 4s.
	BlockGrace time.Duration
	// Narrate posts the policy's Decision.Reason for every non-pass
	// move as a bot_reasoning chat line. Every client receives it and
	// only clients with "show bot reasoning" on render it, because
	// S11.5 settings live in the browser and the server has no
	// per-player store to gate on. Off in the zero value (tests and
	// the soak harness want silence); DefaultConfig turns it on.
	//
	// Improvisation announcements are NOT gated by this. They are
	// mandatory disclosure, not narration — see improvise.go.
	Narrate bool
}

const (
	defaultMinThink   = 700 * time.Millisecond
	defaultMaxThink   = 2 * time.Second
	defaultMaxActions = 64
	defaultMaxRejects = 3
	defaultBlockGrace = 4 * time.Second
	rejectionsKept    = 32
)

// DefaultConfig is the production pacing: 700ms minimum think, 2s
// hard deadline, 64 actions per wake, 3 rejections before a forced
// pass, 4s block grace, reasoning narrated to clients that asked for
// it.
func DefaultConfig() Config {
	return Config{MinThink: defaultMinThink, BlockGrace: defaultBlockGrace, Narrate: true}.withDefaults()
}

func (c Config) withDefaults() Config {
	if c.MaxThink <= 0 {
		c.MaxThink = defaultMaxThink
	}
	if c.MaxActionsPerWake <= 0 {
		c.MaxActionsPerWake = defaultMaxActions
	}
	if c.MaxConsecutiveRejects <= 0 {
		c.MaxConsecutiveRejects = defaultMaxRejects
	}
	return c
}

// Stats are the runner's counters, readable while it runs.
type Stats struct {
	Decisions int64 // Policy.Decide calls that returned a move
	Applied   int64 // moves the dispatcher accepted
	Rejected  int64 // moves the dispatcher refused
	Fallbacks int64 // decisions replaced by the fallback move
	Passes    int64 // pass_priority moves applied
	// Improvisations are announced sandbox bundles the room accepted;
	// ImprovRefused are the ones this runner would not apply (failed
	// validation, or the bundle rolled back). Both are worth watching:
	// a bot improvising constantly means a deck outrunning the
	// catalog, and refusals mean a policy asking for things it may
	// not have. S31 sub-PR 8.
	Improvisations int64
	ImprovRefused  int64
	// Rejections are the most recent rejected moves (up to 32), so a
	// test or an operator can tell a step race from an enumerator bug.
	Rejections []Rejection
}

// Rejection is one move the dispatcher refused.
type Rejection struct {
	Type  string
	Label string
	Err   error
}

// Runner drives one bot seat. Start it with Start; it exits when ctx
// is cancelled or the game leaves StateActive.
type Runner struct {
	room   *ws.Room
	seat   uuid.UUID
	policy Policy
	cfg    Config
	bc     Broadcaster
	log    *slog.Logger

	decisions, applied, rejected, fallbacks, passes atomic.Int64
	improvisations, improvRefused                   atomic.Int64
	rejMu                                           sync.Mutex
	rejections                                      []Rejection
	done                                            chan struct{}
}

// Start launches a runner goroutine for seat in room. bc may be nil.
// The runner acts immediately on the state as it stands (the seat
// may already be owed a decision) and then on every room commit.
func Start(ctx context.Context, room *ws.Room, seat uuid.UUID, policy Policy, cfg Config, bc Broadcaster, log *slog.Logger) *Runner {
	if log == nil {
		log = slog.Default()
	}
	r := &Runner{
		room:   room,
		seat:   seat,
		policy: policy,
		cfg:    cfg.withDefaults(),
		bc:     bc,
		log:    log.With("bot_seat", seat.String(), "policy", policy.Name()),
		done:   make(chan struct{}),
	}
	go r.loop(ctx)
	return r
}

// Done is closed when the runner has exited.
func (r *Runner) Done() <-chan struct{} { return r.done }

// Seat is the seat this runner plays.
func (r *Runner) Seat() uuid.UUID { return r.seat }

// PolicyName is the name of the policy actually driving this seat —
// "random", "heuristic", "assisted", "strong".
//
// It exists because the one bug this layer cannot otherwise show you
// is a seat whose tier and whose policy disagree: a seat labelled
// `heuristic` that is quietly running the random policy plays every
// window, commits every move, and looks from the table exactly like a
// bad bot. Nothing in the protocol carries the difference, so the
// only way to assert it is to ask the runner.
func (r *Runner) PolicyName() string { return r.policy.Name() }

// Stats snapshots the counters.
func (r *Runner) Stats() Stats {
	r.rejMu.Lock()
	rej := append([]Rejection(nil), r.rejections...)
	r.rejMu.Unlock()
	return Stats{
		Decisions:      r.decisions.Load(),
		Applied:        r.applied.Load(),
		Rejected:       r.rejected.Load(),
		Fallbacks:      r.fallbacks.Load(),
		Passes:         r.passes.Load(),
		Improvisations: r.improvisations.Load(),
		ImprovRefused:  r.improvRefused.Load(),
		Rejections:     rej,
	}
}

func (r *Runner) recordRejection(mv legal.Move, err error) {
	r.rejected.Add(1)
	r.rejMu.Lock()
	defer r.rejMu.Unlock()
	r.rejections = append(r.rejections, Rejection{Type: mv.Type, Label: mv.Label, Err: err})
	if len(r.rejections) > rejectionsKept {
		r.rejections = r.rejections[len(r.rejections)-rejectionsKept:]
	}
}

func (r *Runner) loop(ctx context.Context) {
	defer close(r.done)
	wake, unsubscribe := r.room.Subscribe()
	defer unsubscribe()
	if !r.step(ctx) {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-wake:
			if !r.step(ctx) {
				return
			}
		}
	}
}

// step acts for as long as the seat has a decision, up to
// MaxActionsPerWake. Returns false when the runner should exit (the
// game is over or ctx is done).
func (r *Runner) step(ctx context.Context) bool {
	rejects := 0
	for n := 0; n < r.cfg.MaxActionsPerWake; n++ {
		if ctx.Err() != nil {
			return false
		}
		if r.room.Game.CurrentState() != game.StateActive {
			return false
		}
		moves := legal.EnumerateFor(r.room.Game, r.seat)
		if len(moves) == 0 {
			return true
		}
		// The decision, with the policy's view built from the seat's
		// filtered projection only.
		started := time.Now()
		in := Input{
			View:  protocol.ViewOfGameFor(r.room.Game, r.seat.String()),
			Seat:  r.seat,
			Moves: moves,
		}
		if r.concede(ctx, in) {
			return false
		}
		// Improvisation comes before the decision, because it is what
		// a policy does INSTEAD of taking an offered move: the line it
		// wants needs an effect the catalog cannot execute. The bundle
		// moved the board, so re-enumerate rather than deciding
		// against a stale move list. Bounded by MaxActionsPerWake like
		// everything else in this loop.
		if r.improvise(ctx, in, started) {
			continue
		}
		idx, reason := r.decide(ctx, in)
		if idx == Decline {
			// The policy wants nothing from this window and does not
			// hold priority (decide turns a decline into a pass when
			// one is on offer), so nothing is waiting on us. Sleep
			// until the next commit.
			r.log.Debug("bot declines", "reason", reason)
			return true
		}
		if rejects >= r.cfg.MaxConsecutiveRejects {
			// The board keeps changing under us, or the engine will
			// not take the answer the policy keeps choosing. Stop
			// guessing: yield priority if we can, and otherwise take
			// an answer the engine cannot refuse.
			//
			// The second branch is the one that matters. A seat that
			// owes a pending choice is offered that choice's answers
			// and nothing else, so there is no pass here — and
			// returning meant sleeping with the prompt still open,
			// which holds the whole table until someone takes the
			// seat by hand (#544). A suboptimal legal answer, a
			// search's "fail to find", is not a good move. It is a
			// move, and the game continues.
			switch pi, si := PassIndex(moves), SafeIndex(moves); {
			case pi >= 0:
				idx, reason = pi, "forced pass after repeated rejections"
			case si >= 0:
				idx, reason = si, "forced always-legal answer after repeated rejections"
				r.log.Warn("bot forced onto the always-legal answer",
					"move", moves[si].Label, "rejects", rejects)
			default:
				return true
			}
		}
		mv := moves[idx]
		r.pace(ctx, started)
		if mv.Kind == legal.KindPass && r.shouldHoldForBlockers(in.View) {
			r.holdForBlockers(ctx)
		}
		if ctx.Err() != nil {
			return false
		}

		view, seq, err := r.room.Apply(r.seat, func() error {
			return actions.Dispatch(r.room.Game, actions.Action{
				Type:   actions.Type(mv.Type),
				Player: mv.Player,
				Caller: r.seat,
				Params: mv.Params,
			})
		})
		if err != nil {
			rejects++
			r.recordRejection(mv, err)
			r.log.Warn("bot move rejected", "move", mv.Label, "type", mv.Type, "err", err)
			continue
		}
		rejects = 0
		r.applied.Add(1)
		if mv.Kind == legal.KindPass {
			r.passes.Add(1)
		} else {
			r.narrate(mv.Label, reason)
		}
		r.log.Debug("bot move", "move", mv.Label, "reason", reason, "seq", seq)
		if r.bc != nil {
			r.bc.BroadcastState(r.room.Game.ID, seq, view)
		}
	}
	return true
}

// decide runs the policy under the MaxThink deadline and maps any
// failure — error, timeout, out-of-range index — to the fallback
// move. Returns the chosen index and a reason for the log.
func (r *Runner) decide(ctx context.Context, in Input) (int, string) {
	dctx, cancel := context.WithTimeout(ctx, r.cfg.MaxThink)
	defer cancel()
	d, err := r.policy.Decide(dctx, in)
	switch {
	case err != nil:
		r.log.Warn("policy failed; falling back", "err", err)
	case d.Index == Decline:
		// A decline from a seat that holds priority would stall the
		// table, so it becomes the pass it was standing in for.
		r.decisions.Add(1)
		if pi := PassIndex(in.Moves); pi >= 0 {
			return pi, "decline → pass"
		}
		return Decline, d.Reason
	case d.Index < 0 || d.Index >= len(in.Moves):
		r.log.Warn("policy returned an out-of-range move; falling back", "index", d.Index, "moves", len(in.Moves))
	default:
		r.decisions.Add(1)
		return d.Index, d.Reason
	}
	r.fallbacks.Add(1)
	if pi := PassIndex(in.Moves); pi >= 0 {
		return pi, "fallback: pass"
	}
	return 0, "fallback: first legal move"
}

// pace holds the runner so the decision takes at least MinThink of
// wall-clock, measured from when enumeration began.
func (r *Runner) pace(ctx context.Context, started time.Time) {
	if r.cfg.MinThink <= 0 {
		return
	}
	r.hold(ctx, r.cfg.MinThink-time.Since(started))
}

func (r *Runner) hold(ctx context.Context, d time.Duration) {
	if d <= 0 {
		return
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// holdForBlockers waits up to BlockGrace, re-checking every 100ms so
// the pass goes through as soon as every defender has blocked or run
// out of blocks.
func (r *Runner) holdForBlockers(ctx context.Context) {
	deadline := time.Now().Add(r.cfg.BlockGrace)
	for time.Now().Before(deadline) && ctx.Err() == nil {
		r.hold(ctx, 100*time.Millisecond)
		if !r.shouldHoldForBlockers(protocol.ViewOfGameFor(r.room.Game, r.seat.String())) {
			return
		}
	}
}

// shouldHoldForBlockers reports whether passing now would end a
// declare-blockers step that a defending seat may still want to act
// in: this seat is the active player, attackers are declared, and
// some other live seat has a legal block it has not made. Uses the
// same enumerator the defenders do, so "may still want to" is exact
// rather than a guess about untapped creatures.
func (r *Runner) shouldHoldForBlockers(view protocol.GameView) bool {
	if r.cfg.BlockGrace <= 0 || view.Turn.Step != string(game.StepDeclareBlockers) {
		return false
	}
	as := view.Turn.ActiveSeat
	if as < 0 || as >= len(view.Seats) || view.Seats[as].ID != r.seat.String() {
		return false
	}
	for _, seat := range view.Seats {
		if seat.ID == r.seat.String() || seat.Eliminated {
			continue
		}
		id, err := uuid.Parse(seat.ID)
		if err != nil {
			continue
		}
		for _, m := range legal.EnumerateFor(r.room.Game, id) {
			if m.Kind == legal.KindBlock {
				return true
			}
		}
	}
	return false
}
