package aiseat

import (
	"context"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
)

// concede asks a Conceder policy whether the seat is done, and scoops
// it if so. Returns true when the seat conceded and the runner should
// exit.
//
// Conceding goes through actions.Dispatch like every other bot move
// (ADR 0033 §2 — zero protocol forking), so it lands in the replay,
// the undo history and the broadcast exactly as a human's concede
// does.
func (r *Runner) concede(ctx context.Context, in Input) bool {
	c, ok := r.policy.(Conceder)
	if !ok || !c.ShouldConcede(in) {
		return false
	}
	if ctx.Err() != nil {
		return false
	}
	view, seq, err := r.room.Apply(r.seat, func() error {
		return actions.Dispatch(r.room.Game, actions.Action{
			Type:   actions.TypeConcede,
			Player: r.seat,
			Caller: r.seat,
		})
	})
	if err != nil {
		// A concede the engine refuses (the game already ended, say)
		// is not worth retrying; play on and let the next wake sort
		// it out.
		r.log.Warn("bot concede rejected", "err", err)
		return false
	}
	r.applied.Add(1)
	r.log.Info("bot conceded")
	if r.bc != nil {
		r.bc.BroadcastState(r.room.Game.ID, seq, view)
	}
	return true
}
