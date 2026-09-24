package aiseat_test

import (
	"context"
	"math/rand/v2"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// view_cost_test.go pins #1261's runner half: a seat builds its view
// only when it has a decision to make with it.
//
// The nightly's bot gates went red on a budget, not a stall — every
// game finished, just slower than the job allowed — and the biggest
// single cost was a heuristic seat building a full projection on
// every commit whether or not it had anything to do. That shape is
// invisible to every other test (a whole game still plays, only
// slower), so this one counts the hook the view feeds rather than
// timing anything.

// orderCounter is a policy with #687's ordering hook that counts how
// often the runner built a view to ask it. enumerationOrder calls
// TargetOrder exactly when it has built the seat's view, so the count
// is the number of views the runner paid for.
type orderCounter struct {
	orders  atomic.Int64
	decided atomic.Int64
}

func (p *orderCounter) Name() string { return "order-counter" }

func (p *orderCounter) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	p.decided.Add(1)
	if i := aiseat.PassIndex(in.Moves); i >= 0 {
		return aiseat.Decision{Index: i, Reason: "pass"}, nil
	}
	return aiseat.Decision{Index: aiseat.Decline}, nil
}

func (p *orderCounter) TargetOrder(aiseat.Input) legal.TargetOrder {
	p.orders.Add(1)
	return nil
}

// TestASeatWithNoDecisionBuildsNoView: a seat with nothing to do —
// not the priority holder, no prompt owed, no combat declaration —
// wakes, finds nothing, and parks without building its view. Before
// #1261 it built the view first and enumerated second, so every
// commit at a four-seat table cost four projections.
func TestASeatWithNoDecisionBuildsNoView(t *testing.T) {
	room := newRoom(t, 4, 1261)
	g := room.Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	holder := g.Snapshot().Turn.PriorityHolder
	idle := g.Seats[(holder+1)%len(g.Seats)].ID
	if moves := legal.EnumerateFor(g, idle); len(moves) != 0 {
		t.Fatalf("precondition: the non-holder seat has %d moves, want none", len(moves))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &orderCounter{}
	r := aiseat.Start(ctx, room, idle, p, aiseat.Config{}, nil, testLogger())
	waitFor(t, "the idle seat's runner to park", r.Idle)

	if n := p.orders.Load(); n != 0 {
		t.Errorf("a seat with no decision built its view %d times; want 0", n)
	}
	if n := p.decided.Load(); n != 0 {
		t.Errorf("a seat with no decision was asked to decide %d times; want 0", n)
	}
}

// TestASeatWithADecisionStillOrdersItsMoves is the other half: the
// pre-check must not cost the policy its hook. The priority holder
// has a decision, so its view is built, TargetOrder is asked, and the
// policy decides over the ordered list.
func TestASeatWithADecisionStillOrdersItsMoves(t *testing.T) {
	room := newRoom(t, 4, 1261)
	g := room.Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	holder := g.Seats[g.Snapshot().Turn.PriorityHolder].ID

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &orderCounter{}
	r := aiseat.Start(ctx, room, holder, p, aiseat.Config{}, nil, testLogger())
	waitFor(t, "the holder to decide", func() bool { return p.decided.Load() > 0 })
	waitFor(t, "the holder's runner to park", r.Idle)

	if p.orders.Load() == 0 {
		t.Error("the priority holder decided without its ordering hook being asked")
	}
	if o, d := p.orders.Load(), p.decided.Load(); o != d {
		t.Errorf("ordering hook asked %d times for %d decisions; want one view per decision", o, d)
	}
}

// TestAViewBlindActiveBotStillHoldsForBlockers: a random seat decides
// without a view (#1261), and the block-grace hold used to read the
// decision's view to find out which step it was in. Handed the empty
// frame it would read no step at all and pass straight out of
// declare_blockers — the bug the grace exists to prevent, back on the
// seat a new player meets first (ADR 0076's tutorial bot is `random`).
// TestActiveBotHoldsPassForBlockers is the same board with a policy
// that always gets its view.
func TestAViewBlindActiveBotStillHoldsForBlockers(t *testing.T) {
	room := newRoom(t, 2, 9)
	g := room.Game
	bot, def := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	atk, blk := uuid.New(), uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: atk, Name: "Attacker", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: bot.ID, Controller: bot.ID})
	g.Battlefield.PushTop(game.Card{InstanceID: blk, Name: "Blocker", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: def.ID, Controller: def.ID})
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.DeclareAttacker(atk, def.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	// The random seat's only move is the pass, so whatever it draws it
	// is deciding to pass — which is the move the grace holds.
	if moves := legal.EnumerateFor(g, bot.ID); len(moves) != 1 || moves[0].Kind != legal.KindPass {
		t.Fatalf("precondition: the active seat's moves are %v, want only the pass", moves)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := aiseat.Start(ctx, room, bot.ID, aiseat.NewRandomPolicy(rand.NewPCG(9, 9)),
		aiseat.Config{BlockGrace: 2 * time.Second}, nil, testLogger())
	time.Sleep(300 * time.Millisecond)
	// The pass is HELD, not merely unanswered: with no runner on the
	// defending seat a pass would leave the step where it is, so the
	// step alone cannot tell a held pass from a made one. The priority
	// holder and the runner's own count can.
	snap := g.Snapshot()
	if snap.Turn.Step != game.StepDeclareBlockers || g.Seats[snap.Turn.PriorityHolder].ID != bot.ID {
		t.Fatalf("a view-blind active bot passed inside the grace window (step %s, priority with seat %d)",
			snap.Turn.Step, snap.Turn.PriorityHolder)
	}
	if n := r.Stats().Passes; n != 0 {
		t.Fatalf("a view-blind active bot passed %d times inside the grace window; want 0", n)
	}
}
