// Package aiseat runs a bot in a seat: a goroutine that watches a
// ws.Room, asks a Policy which legal move to make whenever the seat
// has a decision, and dispatches it through the same action path a
// WebSocket client uses. ADR 0033 §2–§3.
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
//
// The JSON tags are load-bearing rather than decorative: a decision
// log records Inputs verbatim so that a window can be replayed
// offline through rules.Resolve and heuristic.Decide, both of which
// are pure functions of this struct. protocol.GameView and legal.Move
// are already fully tagged; these three make the whole thing
// round-trip.
type Input struct {
	View  protocol.GameView `json:"view"`
	Seat  uuid.UUID         `json:"seat"`
	Moves []legal.Move      `json:"moves"`
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

// TargetOrderer is an optional Policy extension (#687, ADR 0033 §1).
// A policy that implements it decides which of a spell's candidate
// targets survive the enumerator's expansion cap; one that does not
// gets the engine's candidate order, exactly as before.
//
// It exists because the cap is spent in candidate order, so the
// table's biggest threat could simply be absent from the move list —
// and no policy can pick a move it was never offered. Ranking a board
// is a POLICY question and `legal` may not import this package, so
// the ordering is injected rather than implemented down there.
//
// `in` carries the View and the Seat and NOT the moves: it is called
// to build the move list, so there is nothing yet to rank. The
// returned function is called once per candidate during that one
// enumeration and must not retain anything.
type TargetOrderer interface {
	TargetOrder(in Input) legal.TargetOrder
}

// CostFuelPricer is an optional Policy extension (#1013, ADR 0033 §1).
// A policy that implements it says what a card a CARD-SHAPED COST would
// eat is worth to KEEP — the blue card Force of Will pitches, the five
// cards an Uro exiles to escape — and the enumerator offers the
// cheapest payment first; one that does not gets the engine's zone
// order, exactly as before.
//
// TargetOrderer's twin, and separate from it for the reason
// legal.CostFuelOrder is a separate type: the two ask OPPOSITE
// questions. A target order ranks the board by importance and the
// enumerator keeps the top; a fuel price ranks the seat's own cards by
// what it would lose and the enumerator spends the BOTTOM. One
// interface with one method would let a policy answer one with the
// other and pitch its best card every time.
//
// `in` carries the View and the Seat and NOT the moves: it is called to
// build the move list. The returned function is called once per
// candidate during that one enumeration and must not retain anything.
type CostFuelPricer interface {
	CostFuelPrice(in Input) legal.CostFuelOrder
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

// decidesWithoutView marks RandomPolicy view-blind (#1261): Decide
// reads Input.Moves and nothing else, so the runner need not build a
// projection for it. See viewBlind.
func (p *RandomPolicy) decidesWithoutView() {}

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
