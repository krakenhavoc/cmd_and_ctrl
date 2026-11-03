package protocol

import "testing"

// decked_cause_view_test.go — #4879: the wire says why a player left.
// A draw from an empty library (CR 704.5b) is cause "empty_draw" on the
// public log line and loss_cause "empty_draw" on the seat, for every
// viewer; a loss to life says "life".

func TestDeckedLossIsOnTheLogAndTheSeat(t *testing.T) {
	g := newTwoSeatGame(t)
	a, b := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		a.Library.Cards = nil
		if err := g.DrawNForEffect(a.ID, 1); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	g.RunStateChecksForTest()
	v := ViewOfGame(g)
	elim := logOfKind(v, LogEliminated)
	if len(elim) != 1 || elim[0].Cause != "empty_draw" {
		t.Fatalf("eliminated lines %+v, want one with cause empty_draw", elim)
	}
	if elim[0].Text != "A lost the game (drew from an empty library)" {
		t.Errorf("eliminated text %q", elim[0].Text)
	}
	for _, viewer := range []string{a.ID.String(), b.ID.String()} {
		fv := FilterViewFor(v, viewer)
		if got := fv.Seats[a.Seat].LossCause; got != "empty_draw" {
			t.Errorf("viewer %s: seat loss_cause %q, want empty_draw", viewer, got)
		}
		if got := fv.Seats[b.Seat].LossCause; got != "" {
			t.Errorf("viewer %s: the winner's loss_cause %q, want none", viewer, got)
		}
		if l := logOfKind(fv, LogEliminated); len(l) != 1 || l[0].Cause != "empty_draw" {
			t.Errorf("viewer %s: eliminated lines %+v", viewer, l)
		}
	}
}

func TestLifeLossIsNotDecked(t *testing.T) {
	g := newTwoSeatGame(t)
	a := g.Seats[0]
	g.WithWriteLock(func() { a.Life = 0 })
	g.RunStateChecksForTest()
	v := ViewOfGame(g)
	elim := logOfKind(v, LogEliminated)
	if len(elim) != 1 || elim[0].Cause != "life" {
		t.Fatalf("eliminated lines %+v, want one with cause life", elim)
	}
	if got := v.Seats[a.Seat].LossCause; got != "life" {
		t.Errorf("seat loss_cause %q, want life", got)
	}
}
