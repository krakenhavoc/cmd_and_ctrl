package game

import "testing"

// decked_cause_test.go — #4879: a player who draws from an empty library
// leaves with the cause empty_draw (CR 704.5b), on the elimination event
// and on the player, and a loss to life does not say so.

func TestDrawFromEmptyLibraryLeavesWithEmptyDraw(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	g.WithWriteLock(func() {
		p.Library.Cards = nil
		if err := g.DrawNForEffect(p.ID, 1); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	runChecks(g)
	if !p.Eliminated {
		t.Fatalf("%s drew from an empty library and is still in", p.Name)
	}
	ev, ok := lastEvent(g, EventPlayerEliminated)
	if !ok || ev.Actor != p.ID || ev.Label != string(LossEmptyDraw) {
		t.Fatalf("elimination %+v, want %s", ev, LossEmptyDraw)
	}
	if p.LossCause != LossEmptyDraw {
		t.Errorf("Player.LossCause %q, want %q", p.LossCause, LossEmptyDraw)
	}
	// It survives a snapshot round trip.
	r, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got := r.Seats[1].LossCause; got != LossEmptyDraw {
		t.Errorf("restored LossCause %q, want %q", got, LossEmptyDraw)
	}
}

func TestLossToLifeIsNotEmptyDraw(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	g.WithWriteLock(func() { p.Life = 0 })
	runChecks(g)
	ev, ok := lastEvent(g, EventPlayerEliminated)
	if !ok || ev.Actor != p.ID {
		t.Fatalf("no elimination for %s at 0 life", p.Name)
	}
	if ev.Label == string(LossEmptyDraw) || p.LossCause == LossEmptyDraw {
		t.Fatalf("a loss to life reads as decked: event %q, player %q", ev.Label, p.LossCause)
	}
	if p.LossCause != LossLife {
		t.Errorf("Player.LossCause %q, want %q", p.LossCause, LossLife)
	}
	if q := g.Seats[2]; q.LossCause != "" {
		t.Errorf("a player still in has LossCause %q", q.LossCause)
	}
}
