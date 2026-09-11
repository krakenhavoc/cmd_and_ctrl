// Package aiseat runs a bot in a seat: a goroutine that watches a
// ws.Room, asks a Policy which legal move to make whenever the seat
// has a decision, and dispatches it through the same action path a
// WebSocket client uses. ADR 0024 §2–§3.
//
// The hidden-information guarantee is structural: a Policy receives
// an Input built only from the seat's filtered protocol.GameView and
// the legal.Move list, never a *game.Game. Real policies (heuristic,
// model-backed) land in subpackages under aiseat/ that are forbidden
// from importing internal/game; the import test ships with the first
// of them (sub-PR 6).
package aiseat

import (
	"context"
	"math/rand/v2"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Input is everything a Policy may know when deciding. View is the
// seat's own filtered view — byte-identical to what a human client in
// that seat receives — and Moves is the closed list of legal moves.
type Input struct {
	View  protocol.GameView
	Seat  uuid.UUID
	Moves []legal.Move
}

// Decision names a move by index into Input.Moves. A policy cannot
// return a move that was not offered. Reason is free text for the
// log and, behind a setting, the table.
type Decision struct {
	Index  int
	Reason string
}

// Policy decides. Decide must respect ctx — the runner imposes a
// hard deadline and falls back when it expires — and must be safe to
// call from one goroutine at a time per seat.
type Policy interface {
	Name() string
	Decide(ctx context.Context, in Input) (Decision, error)
}

// RandomPolicy picks uniformly among the legal moves. It is the
// `random` tier: the engine fuzzer, and the baseline every other
// policy is measured against. Deterministic under a seeded source.
type RandomPolicy struct {
	rng *rand.Rand
}

// NewRandomPolicy returns a RandomPolicy over the given source; nil
// uses the process-global generator.
func NewRandomPolicy(src rand.Source) *RandomPolicy {
	if src == nil {
		return &RandomPolicy{}
	}
	return &RandomPolicy{rng: rand.New(src)}
}

func (p *RandomPolicy) Name() string { return "random" }

func (p *RandomPolicy) Decide(_ context.Context, in Input) (Decision, error) {
	if len(in.Moves) == 0 {
		return Decision{}, ErrNoMoves
	}
	var i int
	if p.rng != nil {
		i = p.rng.IntN(len(in.Moves))
	} else {
		i = rand.IntN(len(in.Moves))
	}
	return Decision{Index: i, Reason: "random"}, nil
}

// PassIndex returns the index of the pass_priority move in moves, or
// -1. Runners fall back to it when a policy fails.
func PassIndex(moves []legal.Move) int {
	for i, m := range moves {
		if m.Kind == legal.KindPass {
			return i
		}
	}
	return -1
}
