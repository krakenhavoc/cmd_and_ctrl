package aiseat

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// stepped.go is the lockstep schedule's door into the runner (#1503).
//
// A production table runs one runner goroutine per bot seat (Start),
// and the scheduler decides which seat acts first after every commit.
// That is what a live table wants — a bot is never waiting on another
// bot's turn to think — and it is also why a seed does not fix a game:
// a defender's block lands before or after the active player's pass
// depending on which goroutine woke first, and the game forks there.
//
// A stepped runner takes the goroutine away and nothing else. The
// caller calls Step, seat by seat, in whatever fixed order it likes,
// and each call runs the very same act-loop a woken runner runs —
// enumeration, the policy, the observer, the forced answers, the loop
// breaker's holds and the dispatch. Only WHEN a seat acts is the
// caller's. So a caller that steps its seats in a fixed order on one
// goroutine plays the same moves in the same order on every run of a
// seed: the heuristic gate (#1409) and `boteval arena --lockstep` are
// the two that do.
//
// It is not a production table's schedule and nothing in the server
// uses it. Start is still the only runner the lobby builds.

// NewStepped builds a Runner that has no goroutine and no room
// subscription. It acts only when Step is called, and its Done channel
// is already closed: there is no loop to wait for.
//
// Three Config fields are switched off, because each is a wall-clock
// hold on behalf of somebody else and a stepped table has nobody else
// running while a seat holds:
//
//   - MinThink, and FollowTablePace, which re-derives it from the
//     table: there is no human watching the table to pace for.
//   - BlockGrace: the hold waits for a defender to finish declaring
//     blockers, and the defender cannot act until this Step returns.
//     It would cost its full duration on every combat and change
//     nothing — the pass is the same pass four seconds later.
//
// MaxThink stays: it is the policy's deadline, not pacing, and a model
// seat still needs one. It is also the one place wall clock can still
// reach a stepped game — a policy that overruns it falls back — which
// is why a model seat is only as reproducible as its endpoint.
func NewStepped(room *ws.Room, seat uuid.UUID, policy Policy, cfg Config, log *slog.Logger) *Runner {
	if log == nil {
		log = slog.Default()
	}
	cfg.MinThink = 0
	cfg.BlockGrace = 0
	cfg.FollowTablePace = false
	r := &Runner{
		room:    room,
		seat:    seat,
		policy:  policy,
		cfg:     cfg.withDefaults(),
		log:     log.With("bot_seat", seat.String(), "policy", policy.Name()),
		done:    make(chan struct{}),
		stepped: true,
	}
	close(r.done)
	return r
}

// Step runs one wake's act-loop on the caller's goroutine: the seat
// acts for as long as it has a decision it wants to take, up to
// MaxActionsPerWake, exactly as a woken runner would. It returns false
// when a woken runner would have exited — the game left StateActive,
// the seat conceded, or ctx is done.
//
// Step panics on a runner built by Start. That runner's own goroutine
// is already stepping it, and a second caller would race it for the
// same seat — two decisions in one window, which is not a schedule at
// all.
func (r *Runner) Step(ctx context.Context) bool {
	if !r.stepped {
		panic("aiseat: Step called on a runner built by Start; its own goroutine already steps it (use NewStepped)")
	}
	return r.step(ctx)
}

// Stepped reports whether this runner was built by NewStepped, so a
// harness can say which schedule a table actually ran.
func (r *Runner) Stepped() bool { return r.stepped }
