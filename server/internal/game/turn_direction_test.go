package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_direction_test.go — #6223, "reverse the game's turn order"
// (Aeon Engine). CR 103.1: the default turn order proceeds clockwise
// from the starting player; CR 101.4: the next player in turn order
// makes their choices next. Reversing the order sends both the other
// way round the table.

func passTurnInOrder(t *testing.T, g *Game, want int, when string) {
	t.Helper()
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.Turn.ActiveSeat != want {
		t.Fatalf("%s: seat %d's turn, want seat %d", when, g.Turn.ActiveSeat, want)
	}
}

// Seats A-B-C-D; A reverses the order on A's turn. The next turn is
// D's, then C's, then B's, then A's again, and a second reversal turns
// it back.
func TestReversedTurnOrderRotatesTheOtherWay(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: seat %d starts", g.Turn.ActiveSeat)
	}
	g.WithWriteLock(func() { g.ReverseTurnOrderForEffect() })
	if seat, ok := g.NextTurnSeat(); !ok || seat != 3 {
		t.Errorf("forecast next turn = %d (%v), want seat 3", seat, ok)
	}
	passTurnInOrder(t, g, 3, "first turn after the reversal")
	passTurnInOrder(t, g, 2, "second turn after the reversal")
	passTurnInOrder(t, g, 1, "third turn after the reversal")
	passTurnInOrder(t, g, 0, "the rotation comes back round")
	g.WithWriteLock(func() { g.ReverseTurnOrderForEffect() })
	passTurnInOrder(t, g, 1, "a second reversal restores the default order")
}

// CR 117.3d: passing priority hands it to the next player in turn order.
func TestReversedTurnOrderPassesPriorityTheOtherWay(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToMainForTurnOrder(t, g)
	g.WithWriteLock(func() { g.ReverseTurnOrderForEffect() })
	holder := g.Turn.PriorityHolder
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if want := (holder + 3) % 4; g.Turn.PriorityHolder != want {
		t.Errorf("priority went to seat %d, want seat %d", g.Turn.PriorityHolder, want)
	}
}

func advanceToMainForTurnOrder(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 10 && g.Turn.Step != StepPrecombatMain; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Step != StepPrecombatMain || g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("setup: step %s, priority %d", g.Turn.Step, g.Turn.PriorityHolder)
	}
}

// CR 603.3b with CR 101.4: simultaneous triggers go on the stack APNAP,
// the active player's first, then each other player's in turn order.
// Reversed, seat 3's come before seat 1's.
func TestReversedTurnOrderPutsTriggersOnTheStackInTheNewOrder(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToMainForTurnOrder(t, g)
	ids := make([]uuid.UUID, 4)
	g.WithWriteLock(func() {
		g.ReverseTurnOrderForEffect()
		for s, p := range g.Seats {
			ids[s] = uuid.New()
			g.PendingTriggers = append(g.PendingTriggers, &StackItem{
				ID: ids[s], Kind: StackItemTriggered, Controller: p.ID, Owner: p.ID, Label: p.Name + "'s trigger",
			})
		}
		if !g.drainPendingTriggersAPNAPLocked() {
			t.Fatal("the drain held")
		}
	})
	// Placement order (Seq ascending) is the APNAP order: 0, 3, 2, 1.
	order := []int{0, 3, 2, 1}
	for i := 1; i < len(order); i++ {
		prev, cur := g.StackMeta[ids[order[i-1]]], g.StackMeta[ids[order[i]]]
		if prev == nil || cur == nil {
			t.Fatal("a trigger did not reach the stack")
		}
		if prev.Seq >= cur.Seq {
			t.Errorf("seat %d's trigger was placed after seat %d's", order[i-1], order[i])
		}
	}
	var got []uuid.UUID
	g.WithWriteLock(func() { got = g.apnapSeatsExceptLocked(uuid.Nil) })
	for i, s := range order {
		if got[i] != g.Seats[s].ID {
			t.Errorf("APNAP walk position %d is %s, want seat %d", i, got[i], s)
		}
	}
}

// The reversal is game state: a snapshot keeps it, an undo puts it back,
// and a game that never reversed captures no key for it.
func TestReversedTurnOrderSurvivesARestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if g.CaptureSnapshot().TurnOrderReversed {
		t.Fatal("a fresh game captured a reversed order")
	}
	g.WithWriteLock(func() { g.ReverseTurnOrderForEffect() })
	_, restored := roundTrip(t, g)
	if !restored.TurnOrderReversedForEffect() {
		t.Fatal("the restore dropped the reversal")
	}
	passTurnInOrder(t, restored, 3, "after a restore")
	if c := g.Clone(); !c.TurnOrderReversed {
		t.Error("a clone dropped the reversal")
	}
}

func TestSeatInTurnOrderWrapsBothWays(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	cases := []struct{ from, k, want, wantRev int }{
		{0, 0, 0, 0}, {0, 1, 1, 3}, {3, 1, 0, 2}, {1, 5, 2, 0}, {2, -1, 1, 3},
	}
	for _, c := range cases {
		if got := g.SeatInTurnOrderForEffect(c.from, c.k); got != c.want {
			t.Errorf("default (%d, %d) = %d, want %d", c.from, c.k, got, c.want)
		}
	}
	g.WithWriteLock(func() { g.ReverseTurnOrderForEffect() })
	for _, c := range cases {
		if got := g.SeatInTurnOrderForEffect(c.from, c.k); got != c.wantRev {
			t.Errorf("reversed (%d, %d) = %d, want %d", c.from, c.k, got, c.wantRev)
		}
	}
}
