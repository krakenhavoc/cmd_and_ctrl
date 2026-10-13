package game

import (
	"errors"
	"testing"
)

// remove_seat_test.go pins #3093: an admin removing a seat from a
// started game is a concession made for the seat, so CR 800.4a holds
// exactly as it does for Concede, plus the one thing the removal adds,
// Player.Removed.

// TestRemoveSeatTakesEveryOwnedObjectOutOfTheGame is the headline.
func TestRemoveSeatTakesEveryOwnedObjectOutOfTheGame(t *testing.T) {
	p := newDepartureProbe(t)
	if err := p.g.RemoveSeatFromGame(p.b.ID); err != nil {
		t.Fatalf("RemoveSeatFromGame: %v", err)
	}
	if p.g.State != StateActive {
		t.Fatalf("three seats remain; the game should go on: %s", p.g.State)
	}
	p.assertBHasLeftTheGame(t)
	if !p.b.Eliminated || !p.b.Removed {
		t.Errorf("B: eliminated=%v removed=%v, want both", p.b.Eliminated, p.b.Removed)
	}
	if p.c.Removed || p.c.Eliminated {
		t.Error("a bystander was marked")
	}
}

// TestRemoveSeatIsLoggedAsAConcession: one departure, one log line,
// labelled the way a concession is.
func TestRemoveSeatIsLoggedAsAConcession(t *testing.T) {
	p := newDepartureProbe(t)
	if err := p.g.RemoveSeatFromGame(p.b.ID); err != nil {
		t.Fatal(err)
	}
	var concedes, eliminations int
	for _, ev := range p.g.Events {
		switch {
		case ev.Kind == EventConcede && ev.Actor == p.b.ID:
			concedes++
		case ev.Kind == EventPlayerEliminated && ev.Actor == p.b.ID && ev.Label == string(LossConcede):
			eliminations++
		}
	}
	if concedes != 1 || eliminations != 1 {
		t.Errorf("concede events %d, concede eliminations %d; want 1 and 1", concedes, eliminations)
	}
}

// TestRemoveSeatEndsTheControlEffectsTheSeatHeld: CR 800.4a's second
// step, through the removal rather than the concession.
func TestRemoveSeatEndsTheControlEffectsTheSeatHeld(t *testing.T) {
	p := newDepartureProbe(t)
	g := p.g
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != "thief-3093" {
			return nil
		}
		return []StaticAbility{controlGrabStatic(func(target *Card) bool {
			return target.InstanceID == p.cCreature
		})}
	})
	thief := NewCard("B's Control Magic", p.b.ID)
	thief.TypeLine = "Enchantment — Aura"
	thief.OracleID = "thief-3093"
	g.Battlefield.PushTop(thief)
	g.BumpLayerVersionForTest()
	if got := controllerOfCard(t, g, p.cCreature); got != p.b.ID {
		t.Fatalf("setup: C's creature should be under B's control, got %s", got)
	}
	if err := g.RemoveSeatFromGame(p.b.ID); err != nil {
		t.Fatal(err)
	}
	if got := controllerOfCard(t, g, p.cCreature); got != p.c.ID {
		t.Errorf("C's creature is controlled by %s after B was removed, want its owner %s", got, p.c.ID)
	}
}

// TestRemoveSeatDropsTheSeatsSpellsFromTheStack: CR 800.4a's stack
// clause, and a survivor's spell is left alone.
func TestRemoveSeatDropsTheSeatsSpellsFromTheStack(t *testing.T) {
	p := newDepartureProbe(t)
	g := p.g
	mine := guardedSpellOnStack(t, g, p.b)
	theirs := guardedSpellOnStack(t, g, p.c)
	if err := g.RemoveSeatFromGame(p.b.ID); err != nil {
		t.Fatal(err)
	}
	if zoneHoldingCard(g, mine) != "" {
		t.Errorf("the removed seat's spell is still in %q", zoneHoldingCard(g, mine))
	}
	if g.StackMeta[mine] != nil {
		t.Error("the removed seat's stack item survived")
	}
	if zoneHoldingCard(g, theirs) != "stack" || g.StackMeta[theirs] == nil {
		t.Error("a survivor's spell left the stack with the removed seat's")
	}
}

// TestRemoveSeatMovesPriorityOffTheSeat: the removed seat held
// priority; the table is not left waiting on a player who is gone.
func TestRemoveSeatMovesPriorityOffTheSeat(t *testing.T) {
	p := newDepartureProbe(t)
	g := p.g
	g.WithWriteLock(func() { g.Turn.PriorityHolder = p.b.Seat })
	if err := g.RemoveSeatFromGame(p.b.ID); err != nil {
		t.Fatal(err)
	}
	holder := g.Turn.PriorityHolder
	if holder == p.b.Seat || g.Seats[holder].Eliminated {
		t.Errorf("priority is with seat %d (eliminated=%v) after the removal", holder, g.Seats[holder].Eliminated)
	}
}

// TestRemoveActiveSeatEndsItsTurn: the removed seat's turn goes to the
// next seat in order, as for a concession (CR 800.4j, ADR 0059).
func TestRemoveActiveSeatEndsItsTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	if err := g.RemoveSeatFromGame(active.ID); err != nil {
		t.Fatal(err)
	}
	if g.Turn.ActiveSeat == active.Seat || g.Seats[g.Turn.ActiveSeat].Eliminated {
		t.Fatalf("the turn is still with seat %d", g.Turn.ActiveSeat)
	}
	if g.Turn.Step != StepUpkeep {
		t.Errorf("the next turn is at %s, want upkeep", g.Turn.Step)
	}
}

// TestRemoveSeatFromATwoPlayerGameEndsIt: one player left is the
// usual win check, and the survivor wins.
func TestRemoveSeatFromATwoPlayerGameEndsIt(t *testing.T) {
	g := newActiveGame(t)
	loser, winner := g.Seats[1], g.Seats[0]
	if err := g.RemoveSeatFromGame(loser.ID); err != nil {
		t.Fatal(err)
	}
	if g.State != StateEnded {
		t.Fatalf("state = %s, want ended", g.State)
	}
	if seat, ok := g.WinnerSeat(); !ok || seat != winner.Seat {
		t.Errorf("winner = seat %d (ok=%v), want seat %d", seat, ok, winner.Seat)
	}
	if !loser.Removed {
		t.Error("the removal that ended the game was not recorded on the seat")
	}
}

// TestRemoveSeatMarksAPlayerWhoAlreadyLeft: an admin can take the seat
// away from a player who conceded earlier, with no second concession.
func TestRemoveSeatMarksAPlayerWhoAlreadyLeft(t *testing.T) {
	p := newDepartureProbe(t)
	if err := p.g.Concede(p.b.ID); err != nil {
		t.Fatal(err)
	}
	if p.b.Removed {
		t.Fatal("a voluntary concession marked the seat removed")
	}
	events := len(p.g.Events)
	if err := p.g.RemoveSeatFromGame(p.b.ID); err != nil {
		t.Fatalf("RemoveSeatFromGame after a concession: %v", err)
	}
	if !p.b.Removed {
		t.Error("the seat was not marked")
	}
	if len(p.g.Events) != events {
		t.Errorf("a second departure logged %d events", len(p.g.Events)-events)
	}
}

func TestRemoveSeatRefusals(t *testing.T) {
	g := NewGame()
	pl, err := g.AddPlayer("P1", buildTestDeck("C1"))
	if err != nil {
		t.Fatal(err)
	}
	if err := g.RemoveSeatFromGame(pl.ID); !errors.Is(err, ErrGameNotActive) {
		t.Errorf("a lobby game: %v, want ErrGameNotActive", err)
	}
	a := newActiveGame(t)
	if err := a.RemoveSeatFromGame(pl.ID); !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("a stranger: %v, want ErrPlayerNotFound", err)
	}
	if err := a.RemoveSeatFromGame(a.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if err := a.RemoveSeatFromGame(a.Seats[0].ID); !errors.Is(err, ErrGameNotActive) {
		t.Errorf("an ended game: %v, want ErrGameNotActive", err)
	}
}

// TestSeatRemovedReadsTheFlag: false for a seated player, for one who
// conceded, and true for a removed seat or an id that is not a seat.
func TestSeatRemovedReadsTheFlag(t *testing.T) {
	p := newDepartureProbe(t)
	g := p.g
	if g.SeatRemoved(p.a.ID) {
		t.Error("a seated player reads as removed")
	}
	if err := g.Concede(p.c.ID); err != nil {
		t.Fatal(err)
	}
	if g.SeatRemoved(p.c.ID) {
		t.Error("a player who conceded reads as removed; they may keep spectating")
	}
	if err := g.RemoveSeatFromGame(p.b.ID); err != nil {
		t.Fatal(err)
	}
	if !g.SeatRemoved(p.b.ID) {
		t.Error("a removed seat does not read as removed")
	}
	if !g.SeatRemoved(NewCard("x", p.a.ID).InstanceID) {
		t.Error("an id that is no seat does not read as removed")
	}
}

// TestRemovedSurvivesCloneAndSnapshot: an undo clone and a restart
// both carry the flag.
func TestRemovedSurvivesCloneAndSnapshot(t *testing.T) {
	p := newDepartureProbe(t)
	if err := p.g.RemoveSeatFromGame(p.b.ID); err != nil {
		t.Fatal(err)
	}
	if !p.g.Clone().SeatRemoved(p.b.ID) {
		t.Error("Clone dropped Removed")
	}
	_, restored := roundTrip(t, p.g)
	if !restored.SeatRemoved(p.b.ID) {
		t.Error("a snapshot round trip dropped Removed")
	}
	if restored.SeatRemoved(p.a.ID) {
		t.Error("a snapshot round trip marked a bystander")
	}
}
