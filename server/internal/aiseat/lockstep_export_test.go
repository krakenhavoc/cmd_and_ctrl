package aiseat

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// lockstep_export_test.go is the test-only door the lockstep harness
// in heuristic_game_test.go walks through (#1409). It is a _test.go
// file in package aiseat, so it compiles into this package's test
// binary and nowhere else: production code cannot reach either
// function, and production runners are exactly what they were.
//
// Why a door rather than a scheduler inside the runner: a seeded table
// was not a replayable game because four runner goroutines interleave
// differently on every run — a defender's block lands before or after
// the active player's pass depending on which goroutine the scheduler
// woke first. The fix that keeps production untouched is to take the
// goroutines away in the test and call the very same act-loop (step)
// from one goroutine, seat by seat, in a fixed order. Everything the
// runner does inside a window — enumeration, the policy, the observer,
// forced answers, the loop-breaker holds, dispatch — is the shipped
// code; only WHEN a seat gets to act is the harness's.

// NewSteppedRunner builds a Runner that has no goroutine and no room
// subscription. It acts only when StepForTest is called, and its Done
// channel is already closed — there is no loop to wait for.
func NewSteppedRunner(room *ws.Room, seat uuid.UUID, policy Policy, cfg Config, log *slog.Logger) *Runner {
	if log == nil {
		log = slog.Default()
	}
	r := &Runner{
		room:   room,
		seat:   seat,
		policy: policy,
		cfg:    cfg.withDefaults(),
		log:    log.With("bot_seat", seat.String(), "policy", policy.Name()),
		done:   make(chan struct{}),
	}
	close(r.done)
	return r
}

// StepForTest runs one wake's act-loop on the caller's goroutine: the
// seat acts for as long as it has a decision it wants to take, up to
// MaxActionsPerWake, exactly as a woken runner would. It returns false
// when the runner would have exited (the game left StateActive, or
// ctx is done).
func (r *Runner) StepForTest(ctx context.Context) bool {
	return r.step(ctx)
}
