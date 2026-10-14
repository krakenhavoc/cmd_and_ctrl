package aiseat_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// controlled_turn_test.go — ADR 0167 §7: bots and a turn another player
// plays. Every table is played lockstep (NewSteppedRunner), so each
// test is a fixed sequence of decisions rather than a race.
//
// Alice controls Bob during Bob's turn; Carol is neither.

// watching is a policy that passes when it can and otherwise takes the
// first move, and records every move list it is handed.
type watching struct {
	mu    sync.Mutex
	asked [][]legal.Move
}

func (w *watching) Name() string { return "watching" }

func (w *watching) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	w.mu.Lock()
	w.asked = append(w.asked, append([]legal.Move(nil), in.Moves...))
	w.mu.Unlock()
	if i := aiseat.PassIndex(in.Moves); i >= 0 {
		return aiseat.Decision{Index: i}, nil
	}
	return aiseat.Decision{Index: 0}, nil
}

// sawMoveFor reports whether the policy was ever handed a move for seat.
func (w *watching) sawMoveFor(seat uuid.UUID) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, ms := range w.asked {
		for _, m := range ms {
			if m.Player == seat {
				return true
			}
		}
	}
	return false
}

func (w *watching) calls() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.asked)
}

// controlledTurnTable is a three-seat table on Bob's controlled turn.
func controlledTurnTable(t *testing.T) (room *ws.Room, alice, bob, carol uuid.UUID, controlledSeq int) {
	t.Helper()
	room = newRoom(t, 3, 167)
	g := room.Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	a := g.Turn.ActiveSeat
	alice, bob, carol = g.Seats[a].ID, g.Seats[(a+1)%3].ID, g.Seats[(a+2)%3].ID
	g.WithWriteLock(func() {
		if !g.BeginTurnControlForEffect(game.TurnControl{Controller: alice, Controlled: bob, Source: uuid.New(), SourceName: "Emrakul, the Promised End"}) {
			t.Fatal("BeginTurnControlForEffect refused")
		}
	})
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.ControllerOfPlayer(bob) != alice {
		t.Fatal("setup: Bob's turn is not controlled")
	}
	return room, alice, bob, carol, g.Turn.Seq
}

func turnSeq(g *game.Game) int {
	var s int
	g.ReadSnapshot(func() { s = g.Turn.Seq })
	return s
}

func landsOf(g *game.Game, seat uuid.UUID) int {
	var n int
	g.ReadSnapshot(func() { n = botLandsLocked(g, seat) })
	return n
}

// stepUntilTurnEnds steps the runners in seat order until the turn with
// Seq seq is over, and fails if the table stops moving first.
func stepUntilTurnEnds(t *testing.T, g *game.Game, seq int, runners ...*aiseat.Runner) {
	t.Helper()
	ctx := context.Background()
	for i := 0; i < 400; i++ {
		for _, r := range runners {
			if turnSeq(g) > seq {
				return
			}
			r.Step(ctx)
		}
	}
	t.Fatalf("turn %d did not end: step %s, priority %d", seq, g.Turn.Step, g.Turn.PriorityHolder)
}

// A bot controlling a human plays the human's turn passively: it passes
// and makes the forced answers, plays none of their cards, and its own
// policy is never handed a move for them.
func TestBotControllingAHumanPlaysTheirTurnPassively(t *testing.T) {
	room, alice, bob, carol, seq := controlledTurnTable(t)
	g := room.Game
	alicePol, carolPol := &watching{}, &watching{}
	ar := aiseat.NewSteppedRunner(room, alice, alicePol, aiseat.Config{}, testLogger())
	cr := aiseat.NewSteppedRunner(room, carol, carolPol, aiseat.Config{}, testLogger())
	lands := landsOf(g, bob)

	stepUntilTurnEnds(t, g, seq, ar, cr)

	if alicePol.sawMoveFor(bob) {
		t.Error("Alice's own policy was handed a move for Bob")
	}
	if got := landsOf(g, bob); got != lands {
		t.Errorf("Bob's lands went from %d to %d: the passive policy played his cards", lands, got)
	}
	if st := ar.Stats(); st.Applied == 0 || st.Rejected != 0 {
		t.Errorf("Alice's runner: %+v, want moves applied for Bob and none rejected", st)
	}
	if g.ControllerOfPlayer(bob) != bob {
		t.Error("Bob is still controlled after his turn")
	}
	var active uuid.UUID
	var extra bool
	g.ReadSnapshot(func() { active, extra = g.Seats[g.Turn.ActiveSeat].ID, g.Turn.Extra })
	if active != bob || !extra {
		t.Errorf("after the controlled turn: active %v extra %v, want Bob's extra turn", active, extra)
	}
}

// A human controlling a bot: the bot is offered nothing all turn, and
// the human plays it through the actions layer.
func TestHumanControllingABotLeavesTheBotIdle(t *testing.T) {
	room, alice, bob, carol, seq := controlledTurnTable(t)
	g := room.Game
	bobPol, carolPol := &watching{}, &watching{}
	br := aiseat.NewSteppedRunner(room, bob, bobPol, aiseat.Config{}, testLogger())
	cr := aiseat.NewSteppedRunner(room, carol, carolPol, aiseat.Config{}, testLogger())
	ctx := context.Background()
	for i := 0; i < 400 && turnSeq(g) == seq; i++ {
		br.Step(ctx)
		if len(legal.EnumerateFor(g, bob)) != 0 {
			t.Fatal("the controlled bot is offered moves")
		}
		// Alice (a person) answers whatever Bob owes, as Bob.
		for _, m := range legal.EnumerateFor(g, alice) {
			if m.Player != bob {
				continue
			}
			if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: alice, Params: m.Params}); err != nil {
				t.Fatalf("Alice's move for Bob %q: %v", m.Label, err)
			}
			break
		}
		// …and her own decisions, by passing.
		for _, m := range legal.EnumerateFor(g, alice) {
			if m.Player == alice && m.Kind == legal.KindPass {
				if err := actions.Dispatch(g, actions.Action{Type: actions.TypePassPriority, Player: alice, Caller: alice}); err != nil {
					t.Fatalf("Alice's pass: %v", err)
				}
				break
			}
		}
		cr.Step(ctx)
	}
	if turnSeq(g) == seq {
		t.Fatal("Bob's controlled turn did not end")
	}
	if st := br.Stats(); st.Applied != 0 || st.Decisions != 0 {
		t.Errorf("the controlled bot acted during its controlled turn: %+v", st)
	}
	if bobPol.calls() != 0 {
		t.Errorf("the controlled bot's policy was asked %d times", bobPol.calls())
	}
}

// A bot controlling a bot: the controlled bot does nothing, and the
// controller plays its turn passively.
func TestBotControllingABot(t *testing.T) {
	room, alice, bob, carol, seq := controlledTurnTable(t)
	g := room.Game
	alicePol, bobPol, carolPol := &watching{}, &watching{}, &watching{}
	ar := aiseat.NewSteppedRunner(room, alice, alicePol, aiseat.Config{}, testLogger())
	br := aiseat.NewSteppedRunner(room, bob, bobPol, aiseat.Config{}, testLogger())
	cr := aiseat.NewSteppedRunner(room, carol, carolPol, aiseat.Config{}, testLogger())

	stepUntilTurnEnds(t, g, seq, ar, br, cr)

	if st := br.Stats(); st.Applied != 0 {
		t.Errorf("the controlled bot applied %d moves during its controlled turn", st.Applied)
	}
	if bobPol.calls() != 0 {
		t.Errorf("the controlled bot's policy was asked %d times", bobPol.calls())
	}
	if alicePol.sawMoveFor(bob) {
		t.Error("the controller's own policy was handed a move for Bob")
	}
	if st := ar.Stats(); st.Applied == 0 {
		t.Error("the controlling bot made no move for Bob")
	}
}
