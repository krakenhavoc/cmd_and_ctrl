package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// next_seat_view_test.go — ADR 0188 Decision 9, #3566: TurnView.next_seat
// is the engine's own answer to "whose turn is next", public to every
// viewer, and present for seat 0 (a pointer, because omitempty would drop
// a bare 0).

func TestNextSeatIsOnTheTurnViewForEveryViewer(t *testing.T) {
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	active := g.Turn.ActiveSeat
	want := (active + 1) % len(g.Seats)

	v := ViewOfGame(g)
	if v.Turn.NextSeat == nil || *v.Turn.NextSeat != want {
		t.Fatalf("next_seat = %v, want %d", v.Turn.NextSeat, want)
	}
	for _, viewer := range g.Seats {
		fv := FilterViewFor(v, viewer.ID.String())
		if fv.Turn.NextSeat == nil || *fv.Turn.NextSeat != want {
			t.Errorf("viewer %s sees next_seat %v, want %d", viewer.Name, fv.Turn.NextSeat, want)
		}
	}

	// An extra turn queued for the active seat is the next turn.
	g.WithWriteLock(func() { g.TakeExtraTurnsForEffect(g.Seats[active].ID, uuid.Nil, 1) })
	if v := ViewOfGame(g); v.Turn.NextSeat == nil || *v.Turn.NextSeat != active {
		t.Errorf("next_seat = %v with an extra turn queued, want %d", v.Turn.NextSeat, active)
	}
	// And in the extra turn, the next one is the normal rotation's.
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if v := ViewOfGame(g); !v.Turn.Extra || v.Turn.NextSeat == nil || *v.Turn.NextSeat != want {
		t.Errorf("in the extra turn next_seat = %v, want %d", v.Turn.NextSeat, want)
	}
}

// TestNextSeatZeroIsOnTheWire: seat 0 is a real answer.
func TestNextSeatZeroIsOnTheWire(t *testing.T) {
	zero := 0
	b, err := json.Marshal(TurnView{NextSeat: &zero})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"next_seat":0`) {
		t.Errorf("next_seat 0 is missing from %s", b)
	}
	b, err = json.Marshal(TurnView{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "next_seat") {
		t.Errorf("an unset next_seat is on the wire: %s", b)
	}
}
