package aiseat

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// controlled_turn.go — a bot playing another player's turn (CR 723.1,
// ADR 0167 §7: Emrakul, the Promised End).
//
// While a bot controls another seat's whole turn, its move list holds
// its own moves and that seat's, each with Player set to the seat it is
// for. The runner splits them (splitControlledTurnMoves): its own moves
// go to its policy as always, and the controlled seat's go to
// controlledTurnPolicy, which never calls a model. It passes priority,
// declares no attackers, and answers each prompt with the enumerator's
// first answer, or with the declining answer when one is marked
// always_legal. That is legal, never stalls the table, and never plays
// the controlled player's cards against their interests or in the
// bot's favour. A policy that uses the turn well is a follow-up: the
// evaluation code assumes "my seat", and pricing another seat's board
// is design work of its own.
//
// A controlled bot is offered nothing (the enumerator gives a
// controlled seat no moves), so its runner never wakes to act.

// controlledTurnPolicy decides for a seat whose turn the bot is
// playing. A Policy, so a decision log can name it; it wraps the seat's
// own policy only to say whose it is (capability.go's Unwrap), and
// never asks it anything.
type controlledTurnPolicy struct {
	inner Policy
}

func (p controlledTurnPolicy) Name() string {
	if p.inner == nil {
		return "controlled-turn"
	}
	return p.inner.Name() + "/controlled-turn"
}

// Unwrap names the seat's own policy (capability.go).
func (p controlledTurnPolicy) Unwrap() Policy { return p.inner }

func (p controlledTurnPolicy) Decide(_ context.Context, in Input) (Decision, error) {
	if len(in.Moves) == 0 {
		return Decision{}, ErrNoMoves
	}
	if i := PassIndex(in.Moves); i >= 0 {
		return Decision{Index: i, Reason: "controlled turn: pass"}, nil
	}
	if i := SafeIndex(in.Moves); i >= 0 {
		return Decision{Index: i, Reason: "controlled turn: the declining answer"}, nil
	}
	return Decision{Index: 0, Reason: "controlled turn: the first answer"}, nil
}

// playControlledTurn takes one decision for the seat whose turn this
// bot is playing: controlledTurnPolicy picks from moves, and the move is
// sent as this bot (Caller) for that seat (Player), which the actions
// layer runs as that seat. ok is false when the runner should exit;
// applied says whether the engine took the move.
//
// Paced like the seat's own decisions, except that a lone pass is not
// held for (#2898). A pass is held, not sent, while the CR 732 loop
// breaker suspends automatic passing, as the bot's own pass is — and,
// like it, sent through a mandatory loop's notice (#2506, CR 732.4).
func (r *Runner) playControlledTurn(ctx context.Context, moves []legal.Move, started time.Time, minThink time.Duration) (ok, applied bool) {
	var pol Policy = controlledTurnPolicy{inner: r.policy}
	if len(moves) > 0 && r.room.Game.ControlUntilOfPlayer(moves[0].Player) == game.ControlUntilSearch {
		// #3998: a search the bot controls is answered in its interest.
		pol = controlledSearchPolicy{inner: r.policy, g: r.room.Game}
	}
	d, err := pol.Decide(ctx, Input{Seat: r.seat, Moves: moves})
	if err != nil || d.Index < 0 || d.Index >= len(moves) {
		return true, false
	}
	mv := moves[d.Index]
	if mv.Kind == legal.KindPass && r.room.Game.AutoPassSuspended() && !r.room.Game.MandatoryLoopStanding() {
		r.log.Warn("bot holding the controlled turn: automatic passing is suspended by the loop breaker (CR 732)")
		return true, false
	}
	if !onlyPass(moves) {
		r.pace(ctx, started, minThink)
	}
	if ctx.Err() != nil {
		return false, false
	}
	view, seq, err := r.apply(actions.Action{
		Type:   actions.Type(mv.Type),
		Player: mv.Player,
		Caller: r.seat,
		Params: mv.Params,
	})
	if err != nil {
		r.recordRejection(mv, err)
		r.log.Warn("bot move for a controlled seat rejected", "move", mv.Label, "type", mv.Type, "player", mv.Player, "err", err)
		return true, false
	}
	r.applied.Add(1)
	if mv.Kind == legal.KindPass {
		r.passes.Add(1)
	}
	r.log.Debug("bot move for a controlled seat", "move", mv.Label, "player", mv.Player, "reason", d.Reason, "seq", seq)
	if r.bc != nil {
		r.bc.BroadcastState(r.room.Game.ID, seq, view)
	}
	return true, true
}

// splitControlledTurnMoves separates the seat's own moves from those it
// makes for a seat whose turn (or combat phase, #3999) it is playing
// (ADR 0167 §7) or whose search it controls (#3998, controlled_search.go). A move for a seat it
// controls only during a resolution (Word of Command, ADR 0166) stays
// with its own: its policy has always answered those.
func splitControlledTurnMoves(g *game.Game, seat uuid.UUID, moves []legal.Move) (own, controlled []legal.Move) {
	for _, m := range moves {
		if m.Player == seat || m.Player == uuid.Nil {
			own = append(own, m)
			continue
		}
		if until := g.ControlUntilOfPlayer(m.Player); until.PlaysForThem() || until == game.ControlUntilSearch {
			controlled = append(controlled, m)
			continue
		}
		own = append(own, m)
	}
	if len(controlled) == 0 {
		return moves, nil
	}
	return own, controlled
}
