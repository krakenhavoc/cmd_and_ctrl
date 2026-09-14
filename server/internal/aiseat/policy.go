// Package aiseat runs a bot in a seat: a goroutine that watches a
// ws.Room, asks a Policy which legal move to make whenever the seat
// has a decision, and dispatches it through the same action path a
// WebSocket client uses. ADR 0024 §2–§3.
//
// The hidden-information guarantee is structural: a Policy receives
// an Input built only from the seat's filtered protocol.GameView and
// the legal.Move list, never a *game.Game. Real policies (heuristic,
// model-backed) land in subpackages under aiseat/ that are forbidden
// from importing internal/game. That ban is enforced by
// TestPolicyPackagesDoNotImportGame in aiseat/heuristic — it walks
// every package under aiseat/ rather than just its own, so a policy
// written later is covered without its author having to know.
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

// Decline is the Decision.Index a policy returns to mean "I do not
// want to do any of these right now".
//
// It exists because not every window the enumerator opens is a window
// the seat must act in. The one that matters is declare-blockers: a
// defender is offered blocks BEFORE priority reaches it (blocking is
// a turn-based action, not a response), so the move list is blocks
// and nothing else, and a policy that has finished assigning its
// blockers has no pass to reach for. Without a decline it would have
// to keep declaring blocks until it ran out of creatures — which is
// exactly the "chump-block with everything" behaviour a heuristic
// policy exists to avoid.
//
// Declining is only honoured when the seat does NOT hold priority.
// When a pass is on offer the runner converts a decline into that
// pass, because a seat that holds priority and does nothing stalls
// the table: no action, no commit, no wake, no game.
const Decline = -1

// Conceder is an optional Policy extension. A policy that implements
// it is asked, before each decision, whether the position is lost;
// answering true makes the runner concede the seat and exit.
//
// It is separate from Decide because conceding is not a legal MOVE:
// `legal` deliberately does not enumerate it (a random policy that
// could concede would scoop out of the fuzzer's first game), so the
// only way a policy can express "I am done" is out of band.
type Conceder interface {
	ShouldConcede(in Input) bool
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

// SafeIndex returns the index of the first move the enumerator
// marked legal.Move.AlwaysLegal, or -1.
//
// It is PassIndex generalised, and it exists because passing is not
// always on offer: a seat that owes a pending choice is enumerated
// that choice's answers and NOTHING else, so PassIndex is -1 for the
// whole time the prompt is open. A runner that has run out of
// rejections there has no way to put the game down, and the table
// stops (#544). The always-legal answer is the way out: it is a
// worse move than the one the policy wanted, and enormously better
// than a seat that stops playing.
//
// Pass is itself marked AlwaysLegal, so where both exist this finds
// whichever the enumerator listed first; callers that specifically
// want to yield priority should still ask PassIndex.
func SafeIndex(moves []legal.Move) int {
	for i, m := range moves {
		if m.AlwaysLegal {
			return i
		}
	}
	return -1
}
