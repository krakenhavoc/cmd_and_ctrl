package aiseat

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// lockstep_export_test.go keeps #1409's test names for the lockstep
// harness in heuristic_game_test.go. Since #1503 the stepping door is
// production API (stepped.go: NewStepped and Runner.Step), so both
// names here are one-line delegations and add nothing of their own: a
// test-only copy of the door is how the harness and `boteval arena
// --lockstep` would have drifted apart.

// NewSteppedRunner is NewStepped.
func NewSteppedRunner(room *ws.Room, seat uuid.UUID, policy Policy, cfg Config, log *slog.Logger) *Runner {
	return NewStepped(room, seat, policy, cfg, log)
}

// StepForTest is Step.
func (r *Runner) StepForTest(ctx context.Context) bool {
	return r.Step(ctx)
}
