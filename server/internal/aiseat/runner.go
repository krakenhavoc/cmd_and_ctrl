package aiseat

import (
	"context"
	"encoding/json"
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
	// Observer receives one DecisionEvent per decision window, after
	// the dispatcher's answer is known. Nil (the default) costs
	// nothing: no event is built.
	//
	// It is called inline on the runner's goroutine, so an observer
	// that blocks delays a bot seat. One observer is normally shared
	// by every seat at a table — aiseat/decisionlog's per-game log is
	// exactly that — so it must be safe from several goroutines.
	//
	// The improvise path is deliberately NOT observed: an
	// improvisation is a bundle the policy asked for instead of a
	// move, there is no index and no move list entry, and
	// improvise.go already announces it to the table.
	Observer DecisionObserver
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
//
// The JSON tags are here because these numbers leave the process: the
// eval harness reports them per policy, and a report nobody can
// machine-read is a report that gets retyped.
type Stats struct {
	Decisions int64 `json:"decisions"` // Policy.Decide calls that returned a move
	Applied   int64 `json:"applied"`   // moves the dispatcher accepted
	Rejected  int64 `json:"rejected"`  // moves the dispatcher refused
	Fallbacks int64 `json:"fallbacks"` // decisions replaced by the fallback move
	Passes    int64 `json:"passes"`    // pass_priority moves applied
	// Improvisations are announced sandbox bundles the room accepted;
	// ImprovRefused are the ones this runner would not apply (failed
	// validation, or the bundle rolled back). Both are worth watching:
	// a bot improvising constantly means a deck outrunning the
	// catalog, and refusals mean a policy asking for things it may
	// not have. S31 sub-PR 8.
	Improvisations int64 `json:"improvisations"`
	ImprovRefused  int64 `json:"improv_refused"`
	// Rejections are the most recent rejected moves (up to 32), so a
	// test or an operator can tell a step race from an enumerator bug.
	Rejections []Rejection `json:"rejections,omitempty"`
	// Latency is the distribution of whole-decision durations over
	// the most recent 1024 windows — #505's "latency percentiles",
	// which until now were unreachable from outside the runner.
	Latency Percentiles `json:"latency"`
	// Spend is what this seat has cost in model calls and tokens,
	// split into deciding and improvising (#735, ADR 0033 §5). Zero
	// for every policy that is not a Spender — which is every tier
	// with no model in it, and the true answer for them.
	Spend Spend `json:"spend"`
}

// Rejection is one move the dispatcher refused.
type Rejection struct {
	Type  string
	Label string
	Err   error
}

// MarshalJSON renders Err as a string. An `error` marshals to `{}` by
// default, which turns the single most useful field of a rejection —
// why the engine said no — into two braces.
func (r Rejection) MarshalJSON() ([]byte, error) {
	out := struct {
		Type  string `json:"type"`
		Label string `json:"label,omitempty"`
		Err   string `json:"err,omitempty"`
	}{Type: r.Type, Label: r.Label}
	if r.Err != nil {
		out.Err = r.Err.Error()
	}
	return json.Marshal(out)
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
	latMu                                           sync.Mutex
	latencies                                       []time.Duration
	latNext                                         int
	done                                            chan struct{}

	// idle is true while the loop is parked on its wake channel, and
	// wake is that channel, installed by Start before it returns.
	// Both guarded by idleMu because Idle is read from another
	// goroutine. Parked is not the same as having nothing to do, which
	// is why Idle reads the channel as well as the flag — see Idle.
	idleMu sync.Mutex
	idle   bool
	wake   <-chan struct{}
}

// Start launches a runner goroutine for seat in room. bc may be nil.
// The runner acts immediately on the state as it stands (the seat
// may already be owed a decision) and then on every room commit.
//
// The room subscription is taken HERE, on the caller's goroutine,
// before Start returns — not on the runner's goroutine once it is
// scheduled. That makes "Start returned" mean "this seat is listening
// from now on", which is the only ordering a caller can establish at
// all: the runner goroutine is not scheduled by whoever seated it, so
// with the subscription inside the loop there was a window — unbounded
// on a loaded machine — in which a commit reached every other seat and
// not this one.
//
// The window cost a wake, not a decision: the first step reads the
// live game, so it sees the effect of a commit it was not told about.
// It is still a real defect and not only a test problem, because a
// seat that declines a window PARKS until the next commit, and a commit
// that landed in the window is one this seat will now never be woken
// for. Seat the last bot at a table while another seat's move is in
// flight and that bot can sit out the rest of the game waiting for a
// wake that has already been and gone. #938 is that shape, seen from a
// test: the observed seat declined, parked, and the only other writer
// had already finished.
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
	wake, unsubscribe := room.Subscribe()
	// Under idleMu because Idle reads r.wake from another goroutine,
	// and a caller may hold the *Runner before the loop is scheduled.
	r.idleMu.Lock()
	r.wake = wake
	r.idleMu.Unlock()
	go r.loop(ctx, wake, unsubscribe)
	return r
}

// Done is closed when the runner has exited.
func (r *Runner) Done() <-chan struct{} { return r.done }

// Idle reports whether the runner has finished acting and is parked
// waiting for the next room commit, with no wake already pending. A
// runner that has exited is idle too: it has nothing left to do
// either.
//
// It exists for the tests, and it is the answer to #848. "Has the bot
// finished?" was previously a sleep long enough to probably be true,
// and a probably that fails on a loaded machine is a flake. Nothing
// else observable says it: a room sequence standing still means only
// that nothing has committed YET, and a runner that is mid-decision
// looks exactly like one that has stopped. The runner knows, so it
// says so.
//
// "No wake already pending" is asked of the wake channel HERE, at the
// moment of the call, and that is #924. The flag alone was set from
// `len(wake) == 0` as the loop parked, so a commit landing one
// instruction later left a runner that reported itself idle with a
// wake sitting in the channel — parked between waking and dispatching,
// which reads exactly like parked with nothing to do. The window is a
// scheduler quantum wide on a single-CPU machine, which is why it was
// a CI flake and not a local one.
//
// Idle is still a fact about this instant. A commit from another seat
// can wake the runner immediately afterwards, which is a bot with new
// work rather than a bot that lied — so a test that needs a table to
// have STOPPED asks every runner at it, and asks the game what it is
// holding, in the same wait.
func (r *Runner) Idle() bool {
	select {
	case <-r.done:
		return true
	default:
	}
	r.idleMu.Lock()
	defer r.idleMu.Unlock()
	// wake is installed by Start, so it is never nil here; idle is
	// false until the loop parks for the first time.
	return r.idle && len(r.wake) == 0
}

func (r *Runner) setIdle(v bool) {
	r.idleMu.Lock()
	r.idle = v
	r.idleMu.Unlock()
}

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
		Latency:        PercentilesOf(r.latencySnapshot()),
		Spend:          r.spend(),
	}
}

// spend asks the policy what this seat has cost so far, through the
// wrapper chain — the shipped model tiers implement Spender on the
// funnel itself, but a seat wrapped by a harness or a test must not
// stop reporting. That is #1060's hole, one interface later, and the
// reason every lookup in this package goes through Capability.
//
// A policy that cannot reach a model is not a Spender and contributes
// a zero Spend, which is what a `random` or `heuristic` seat costs.
func (r *Runner) spend() Spend {
	sp, ok := Capability[Spender](r.policy)
	if !ok {
		return Spend{}
	}
	return sp.Spend()
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

// loop takes the wake channel and its unsubscribe from Start, which
// subscribed before it returned — see Start for why the subscription
// may not happen here.
func (r *Runner) loop(ctx context.Context, wake <-chan struct{}, unsubscribe func()) {
	defer close(r.done)
	defer unsubscribe()
	if !r.step(ctx) {
		return
	}
	for {
		// Parked. A wake already sitting in the channel is work in
		// hand — the select below takes it without blocking — so it is
		// not idleness, and Idle must not report it as such. Idle asks
		// the channel itself, so this only has to say "parked": the
		// answer then stays right for a wake that arrives after this
		// line rather than before it (#924).
		r.setIdle(true)
		select {
		case <-ctx.Done():
			return
		case <-wake:
			r.setIdle(false)
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
		// #687 / #1013: the policy may order the enumerator's target
		// expansion and price the cards a cost would eat, and both are
		// reads of the seat's own view — so for a policy with either
		// opinion the view is built first and reused for the decision
		// below. A policy with neither pays for nothing:
		// enumerationOrder answers a zero Options without touching the
		// projection, and the enumeration is byte-identical to what it
		// was.
		started := time.Now()
		opts, ordering := r.enumerationOrder()
		moves := legal.EnumerateForWithOptions(r.room.Game, r.seat, opts)
		if len(moves) == 0 {
			return true
		}
		// The decision, with the policy's view built from the seat's
		// filtered projection only.
		in := Input{
			View:  ordering.View,
			Seat:  r.seat,
			Moves: moves,
		}
		if in.View.ID == "" {
			in.View = protocol.ViewOfGameFor(r.room.Game, r.seat.String())
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
		out := r.decide(ctx, in)
		idx, reason := out.index, out.reason
		if idx == Decline {
			// The policy wants nothing from this window and does not
			// hold priority (decide turns a decline into a pass when
			// one is on offer), so nothing is waiting on us. Sleep
			// until the next commit.
			r.log.Debug("bot declines", "reason", reason)
			r.observe(in, out, nil, 0, false, nil)
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
			fi, freason, forced := forcedAnswer(moves)
			out.forced = forced
			switch forced {
			case ForcedAlwaysLegal:
				r.log.Warn("bot forced onto the always-legal answer",
					"move", moves[fi].Label, "rejects", rejects)
			case ForcedNoLegalAnswer:
				// Nothing left to play: no pass, no unconditional
				// answer, and the policy's picks keep bouncing. The
				// seat puts the window down, which is #544's shape,
				// so it is reported rather than only returned.
				r.log.Warn("bot has run out of answers and is leaving the window open",
					"rejects", rejects, "moves", len(moves))
				r.observe(in, out, nil, 0, false, nil)
				return true
			}
			idx, reason = fi, freason
			out.index, out.reason = idx, reason
		}
		mv := moves[idx]
		// #628, CR 726: the engine has flagged a trigger loop, so
		// automatic passing is suspended for the whole table. A bot's
		// pass is as automatic as a browser's autopass toggle, so the
		// seat holds instead of feeding the loop another iteration.
		// The runner is edge-triggered on room commits, so holding
		// here costs nothing: nothing commits until somebody makes a
		// real decision, and the game state carries the notice saying
		// which ability and how many times. A non-pass move is still
		// played — and playing one clears the notice like any other
		// decision, which is how a bot-only table gets moving again
		// when it has something else to do.
		if mv.Kind == legal.KindPass && r.room.Game.AutoPassSuspended() {
			r.log.Warn("bot holding: automatic passing is suspended by the loop breaker (CR 726)")
			r.observe(in, out, &mv, 0, false, nil)
			return true
		}
		// #810, the other shape of the same runaway. A trigger loop
		// runs itself and the breaker stops it by suspending automatic
		// PASSES. An activation loop is fed one activation at a time,
		// so the move that keeps it turning is not a pass and the line
		// above never sees it — a bot handed a free, repeatable
		// ability took it until the wall clock ran out. A bot's
		// activation of the permanent the notice names is as automatic
		// as its pass, so it holds on that too, and holds on nothing
		// else: any other move is progress and clears the notice like
		// any other decision.
		//
		// The notice names a permanent and one of its abilities; this
		// matches on the permanent. Holding a second ability of the
		// same source for as long as the notice stands is the
		// conservative direction — the notice means "stop feeding this
		// card" — and it costs a bot nothing it could not do on the
		// next turn.
		if mv.Kind == legal.KindActivate {
			if n := r.room.Game.CurrentLoopNotice(); n != nil && n.Source == mv.Source {
				r.log.Warn("bot holding: the loop breaker named this permanent's ability (CR 726)",
					"move", mv.Label, "ability", n.Label, "count", n.Count)
				r.observe(in, out, &mv, 0, false, nil)
				return true
			}
		}
		r.pace(ctx, started)
		if mv.Kind == legal.KindPass && r.shouldHoldForBlockers(in.View) {
			r.holdForBlockers(ctx)
		}
		if ctx.Err() != nil {
			// The decision was made and paid for; the pacing hold
			// outlived the runner. Report the window rather than
			// dropping it, so "one event per window the policy
			// decided in" holds even on the way out.
			out.forced = ForcedCancelled
			r.observe(in, out, &mv, 0, false, nil)
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
			r.observe(in, out, &mv, 0, false, err)
			continue
		}
		rejects = 0
		r.applied.Add(1)
		r.observe(in, out, &mv, seq, true, nil)
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

// outcome is what one call to decide produced: the index the runner
// will dispatch, and everything an observer needs to say why.
//
// It is a struct rather than the old (int, string) pair because the
// pair could not distinguish "the policy chose the pass" from "the
// policy timed out and the runner took the pass", and those are
// opposite facts about a bot seat. The reason string said so in
// English, which is not a thing a report can count.
type outcome struct {
	// index is the move to dispatch, or Decline.
	index int
	// reason is the free text that reaches the log and the table.
	reason string
	// fallback is the runner's own fallback cause, empty when the
	// policy's answer was taken as given.
	fallback string
	// forced is what the runner did instead of dispatching that
	// answer, empty when it dispatched it. Independent of fallback:
	// a window that timed out AND then had its fallback pass rejected
	// into the always-legal answer sets both.
	forced string
	// decision and err are exactly what the policy returned.
	decision Decision
	err      error
	// traced says whether trace came from the policy.
	traced bool
	trace  Trace
	// latency is the whole Decide call, deadline included.
	latency time.Duration
}

// decide runs the policy under the MaxThink deadline and maps any
// failure — error, timeout, out-of-range index — to the fallback
// move.
func (r *Runner) decide(ctx context.Context, in Input) outcome {
	dctx, cancel := context.WithTimeout(ctx, r.cfg.MaxThink)
	defer cancel()

	started := time.Now()
	var (
		d   Decision
		tr  Trace
		err error
	)
	traced := false
	// A Tracer answers the window AND shows its working. Nothing else
	// changes: DecideTraced returns what Decide would have, so this
	// is the same decision with the prompt, the reply and the
	// heuristic's ranking attached.
	if t, ok := Capability[Tracer](r.policy); ok && r.cfg.Observer != nil {
		d, tr, err = t.DecideTraced(dctx, in)
		traced = true
	} else {
		d, err = r.policy.Decide(dctx, in)
		tr = Trace{HeuristicIndex: Decline}
		if _, isRandom := r.policy.(*RandomPolicy); isRandom {
			tr.Layer = TraceLayerRandom
		}
	}
	out := outcome{decision: d, err: err, traced: traced, trace: tr, latency: time.Since(started)}
	r.observeLatency(out.latency)

	switch {
	case err != nil:
		r.log.Warn("policy failed; falling back", "err", err)
		out.fallback = FallbackPolicyError
		// A deadline miss is a different operational problem from a
		// policy that threw, and on a self-hosted model it is the
		// likely one. Both fall back the same way; only the label
		// differs, and the label is what a report counts.
		if errors.Is(err, context.DeadlineExceeded) {
			out.fallback = FallbackTimeout
		}
	case d.Index == Decline:
		// A decline from a seat that holds priority would stall the
		// table, so it becomes the pass it was standing in for.
		r.decisions.Add(1)
		if pi := PassIndex(in.Moves); pi >= 0 {
			out.index, out.reason, out.fallback = pi, "decline → pass", FallbackDeclinePass
			return out
		}
		out.index, out.reason = Decline, d.Reason
		return out
	case d.Index < 0 || d.Index >= len(in.Moves):
		r.log.Warn("policy returned an out-of-range move; falling back", "index", d.Index, "moves", len(in.Moves))
		out.fallback = FallbackOutOfRange
	default:
		r.decisions.Add(1)
		out.index, out.reason = d.Index, d.Reason
		return out
	}
	r.fallbacks.Add(1)
	if pi := PassIndex(in.Moves); pi >= 0 {
		out.index, out.reason = pi, "fallback: pass"
		return out
	}
	out.index, out.reason = 0, "fallback: first legal move"
	return out
}

// forcedAnswer is what the runner takes when the policy's choices
// have been refused MaxConsecutiveRejects times in a row: yield
// priority if that is on offer, otherwise take the answer the engine
// cannot refuse, otherwise admit there is nothing.
//
// It is its own function because the third case is the one that
// matters and the hardest to reach: it needs an enumerator and an
// engine that disagree, in a window with no pass and no unconditional
// answer, which is #544's shape and is not constructible from a real
// game in a test. Splitting it out makes the rule checkable on a move
// list rather than only on a table.
//
// Returns the index to dispatch (Decline when there is none), the
// reason for the log, and the Forced cause.
func forcedAnswer(moves []legal.Move) (int, string, string) {
	if pi := PassIndex(moves); pi >= 0 {
		return pi, "forced pass after repeated rejections", ForcedPass
	}
	if si := SafeIndex(moves); si >= 0 {
		return si, "forced always-legal answer after repeated rejections", ForcedAlwaysLegal
	}
	return Decline, "no legal answer left after repeated rejections", ForcedNoLegalAnswer
}

// observe emits one DecisionEvent. mv is the move that was
// dispatched, nil when none was.
func (r *Runner) observe(in Input, out outcome, mv *legal.Move, seq uint64, applied bool, rejectErr error) {
	obs := r.cfg.Observer
	if obs == nil {
		return
	}
	ev := DecisionEvent{
		Game:        r.room.Game.ID,
		Seat:        r.seat,
		Policy:      r.policy.Name(),
		Seq:         seq,
		Input:       in,
		Traced:      out.traced,
		Trace:       out.trace,
		Decision:    out.decision,
		DecisionErr: out.err,
		Fallback:    out.fallback,
		Forced:      out.forced,
		Index:       out.index,
		Reason:      out.reason,
		Latency:     out.latency,
		Applied:     applied,
		RejectErr:   rejectErr,
	}
	if mv != nil {
		ev.Label = mv.Label
	}
	obs.Observe(ev)
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
