package game

import "testing"

// mill_no_loss_test.go — #767. CR 701.17b: a player told to mill more
// cards than their library holds "mill[s] as many as possible", and
// only an attempt to DRAW from an empty library loses the game
// (CR 704.5b). MillToZoneForEffect used to set LosesAtNextSBA whenever
// a run went past the bottom of the library, so every mill, every
// "exile the top N" and every until-run that never found its card
// eliminated a player the rules leave in the game.

// runOutAndCheck empties `p`'s library through `mill`, runs the state
// checks, and asserts the player is still in the game.
func runOutAndCheck(t *testing.T, g *Game, p *Player, mill func() error) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := mill(); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	if n := p.Library.Size(); n != 0 {
		t.Fatalf("library holds %d, want it run out", n)
	}
	if p.LosesAtNextSBA {
		t.Error("running a library out is not a draw: no loss is flagged (CR 701.17b)")
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if p.Eliminated {
		t.Error("the player is still in the game after the state checks")
	}
}

// Mill 93 against a 92-card library: the 92 go, nobody loses.
func TestMillMoreThanTheLibraryHoldsDoesNotLose(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	size := p.Library.Size()
	yard := p.Graveyard.Size()
	runOutAndCheck(t, g, p, func() error { return g.MillNForEffect(p.ID, size+1) })
	if got := p.Graveyard.Size() - yard; got != size {
		t.Errorf("milled %d, want the whole library of %d", got, size)
	}
}

// An unbounded until-run that never finds its card reads the library
// dry and stops there.
func TestUnboundedMillUntilThatNeverMatchesDoesNotLose(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	never := func(Card) bool { return false }
	runOutAndCheck(t, g, p, func() error {
		_, err := g.MillToZoneForEffect(p.ID, 0, ZoneGraveyard, never)
		return err
	})
}

// Helm of Obedience's shape: a bounded until-run whose bound is larger
// than the library and whose card is never found.
func TestBoundedMillUntilPastTheLibraryDoesNotLose(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	never := func(Card) bool { return false }
	runOutAndCheck(t, g, p, func() error {
		_, err := g.MillToZoneForEffect(p.ID, p.Library.Size()+5, ZoneGraveyard, never)
		return err
	})
}

// "Exile the top N cards" past the end of the library is not a mill
// and not a draw either.
func TestExileTopPastTheLibraryDoesNotLose(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	runOutAndCheck(t, g, p, func() error {
		moved, err := g.MillToZoneForEffect(p.ID, p.Library.Size()+4, ZoneExile, nil)
		if err == nil && len(moved) == 0 {
			t.Error("the library's cards are exiled")
		}
		return err
	})
}

// The CR 704.5b regression: the milled-out player's NEXT draw, from the
// library the mill left empty, still loses.
func TestDrawAfterMillingOutStillLoses(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	p := g.Seats[1]
	runOutAndCheck(t, g, p, func() error { return g.MillNForEffect(p.ID, p.Library.Size()+1) })

	if err := g.DrawCard(p.ID); err != ErrZoneEmpty {
		t.Fatalf("draw from the empty library: %v, want ErrZoneEmpty", err)
	}
	if !p.LosesAtNextSBA {
		t.Error("drawing from an empty library flags the loss")
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if !p.Eliminated {
		t.Error("and the state checks eliminate the player (CR 704.5b)")
	}
}
