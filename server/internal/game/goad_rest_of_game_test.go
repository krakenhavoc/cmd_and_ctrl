package game

import (
	"testing"

	"github.com/google/uuid"
)

// goad_rest_of_game_test.go pins #3522: "goaded for the rest of the game"
// (Dack Fayden, Helping Hand). The marker has no end of its own; it ends
// when the permanent leaves the battlefield and nowhere else.

func goadForRestOfGame(t *testing.T, g *Game, id, by uuid.UUID) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() { ok = g.GoadForRestOfGameForEffect(id, by) })
	if !ok {
		t.Fatalf("GoadForRestOfGameForEffect(%s) refused a battlefield creature", id)
	}
}

// attackAtC declares the goaded X at C, the one opponent no goad names, so
// the table can move on to the next turn.
func attackAtC(t *testing.T, tb goadTable) {
	t.Helper()
	advanceIntoStep(t, tb.g, StepDeclareAttackers)
	if err := tb.g.DeclareAttacker(tb.x, tb.c.ID); err != nil {
		t.Fatalf("X at C: %v", err)
	}
}

// A goad for the rest of the game outlives its goader's next turn, and a
// whole lap of the table, while an ordinary goad beside it ends on time.
func TestGoadForRestOfGameSurvivesTurnChanges(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goadForRestOfGame(t, g, tb.x, tb.a.ID)
	goad(t, g, tb.x, tb.b.ID)
	attackAtC(t, tb)

	advanceToStepOfSeat(t, g, tb.aSeat, StepUpkeep)
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.a.ID, tb.b.ID) {
		t.Fatalf("A's own turn began and the goaders are %v, want [A B]: A's goad is for the rest of the game", got)
	}
	advanceToStepOfSeat(t, g, tb.bSt, StepUpkeep)
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.a.ID) {
		t.Fatalf("after B's turn began, goaders = %v, want [A]: B's ordinary goad ends, A's does not", got)
	}
	// A full lap and then some: back to the creature's controller and
	// round to A again.
	advanceToStepOfSeat(t, g, tb.meSeat, StepDeclareAttackers)
	if err := g.DeclareAttacker(tb.x, tb.c.ID); err != nil {
		t.Fatalf("X at C: %v", err)
	}
	advanceToStepOfSeat(t, g, tb.aSeat, StepUpkeep)
	advanceToStepOfSeat(t, g, tb.meSeat, StepDeclareAttackers)
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.a.ID) {
		t.Fatalf("after a lap the goaders are %v, want [A]", got)
	}
	// And it is still enforced: X may not attack its goader.
	requirementErr(t, g.DeclareAttacker(tb.x, tb.a.ID))
	if err := g.DeclareAttacker(tb.x, tb.c.ID); err != nil {
		t.Fatalf("X at C: %v", err)
	}
}

// The goad ends with the permanent: a creature that leaves and comes back
// is a new object (CR 400.7) and is not goaded.
func TestGoadForRestOfGameEndsWhenThePermanentLeaves(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goadForRestOfGame(t, g, tb.x, tb.a.ID)
	if err := g.DestroyPermanentForEffect(tb.x); err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
	c, ok := g.LookupCardForEffect(tb.x)
	if !ok {
		t.Fatal("the destroyed creature is nowhere")
	}
	if len(c.Goads) != 0 {
		t.Fatalf("a creature in the graveyard still carries goads: %+v", c.Goads)
	}
	var again bool
	g.WithWriteLock(func() { again = g.GoadForRestOfGameForEffect(tb.x, tb.a.ID) })
	if again {
		t.Error("goaded a card that is not on the battlefield")
	}
}

// A later "until your next turn" goad by the same player never shortens a
// goad for the rest of the game: the effects overlap and the longer one
// stands.
func TestOrdinaryGoadDoesNotShortenARestOfGameGoad(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goadForRestOfGame(t, g, tb.x, tb.a.ID)
	goad(t, g, tb.x, tb.a.ID)
	attackAtC(t, tb)
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.a.ID) {
		t.Fatalf("a second goad by A made the goaders %v, want one entry for A", got)
	}
	advanceToStepOfSeat(t, g, tb.aSeat, StepUpkeep)
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.a.ID) {
		t.Fatalf("A's refresh ended A's goad for the rest of the game: %v", got)
	}
}

// Changing hands leaves the goad standing: it names the goader, not the
// controller.
func TestRestOfGameGoadSurvivesAControlChange(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goadForRestOfGame(t, g, tb.x, tb.me.ID)
	g.WithWriteLock(func() {
		g.GainControlForEffect(tb.x, tb.x, tb.b.ID, IndefiniteDuration(), "test theft")
	})
	c, _ := battlefieldCardByID(g, tb.x)
	if c.Controller != tb.b.ID {
		t.Fatalf("controller = %v, want B", c.Controller)
	}
	if got := goadersOf(t, g, tb.x); !sameGoaders(got, tb.me.ID) {
		t.Fatalf("goaders after the theft = %v, want the original goader", got)
	}
}

// The flag is written to a restore point and read back, and the restored
// game still holds the goad across the goader's turn.
func TestRestOfGameGoadRoundTripsTheSnapshot(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goadForRestOfGame(t, g, tb.x, tb.a.ID)
	attackAtC(t, tb)
	_, restored := roundTrip(t, g)
	after, ok := battlefieldCardByID(restored, tb.x)
	if !ok || len(after.Goads) != 1 || !after.Goads[0].ForRestOfGame {
		t.Fatalf("goads after the round trip = %+v, want one rest-of-game goad", after.Goads)
	}
	advanceToStepOfSeat(t, restored, tb.aSeat, StepUpkeep)
	if got := goadersOf(t, restored, tb.x); !sameGoaders(got, tb.a.ID) {
		t.Fatalf("restored goaders after A's turn began = %v, want [A]", got)
	}
}
