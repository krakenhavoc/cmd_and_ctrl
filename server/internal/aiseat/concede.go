package aiseat

import (
	"context"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// concede asks a Conceder policy whether the seat is done, and scoops
// it if so. Returns true when the seat conceded and the runner should
// exit.
//
// Conceding goes through actions.Dispatch like every other bot move
// (ADR 0033 §2 — zero protocol forking), so it lands in the replay,
// the undo history and the broadcast exactly as a human's concede
// does.
// targetOrder is the policy's answer to "which targets matter", or
// nil when it has no opinion — the ordering hook #687 threads into
// the enumerator. Nil is the whole of the old behaviour, so a policy
// that does not implement TargetOrderer (random, and any policy
// written before this) is untouched.
// It also hands back the Input it built, so the decision below reuses
// that one projection instead of building a second: the ordering and
// the decision describe the same instant, and a bot seat still costs
// exactly one view per decision.
// #1013 added the second hook, and it shares the one projection for
// the same reason the decision does: the ordering, the fuel price and
// the decision must describe the same instant, and a bot seat still
// costs exactly one view per decision.
func (r *Runner) enumerationOrder() (legal.Options, Input) {
	orderer, wantsOrder := r.policy.(TargetOrderer)
	pricer, wantsFuel := r.policy.(CostFuelPricer)
	if !wantsOrder && !wantsFuel {
		return legal.Options{}, Input{}
	}
	in := Input{View: protocol.ViewOfGameFor(r.room.Game, r.seat.String()), Seat: r.seat}
	var opts legal.Options
	if wantsOrder {
		opts.OrderTargets = orderer.TargetOrder(in)
	}
	if wantsFuel {
		opts.OrderCostFuel = pricer.CostFuelPrice(in)
	}
	return opts, in
}

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
