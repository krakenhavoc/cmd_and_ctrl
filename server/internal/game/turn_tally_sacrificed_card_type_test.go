package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_tally_sacrificed_card_type_test.go — #5224. TurnTally.
// SacrificedCardTypes is "if you've sacrificed an artifact this turn":
// the permanents a player sacrificed this turn, per card type they had
// as they were sacrificed, a token included.

func sacrificedArtifacts(g *Game, p uuid.UUID) int {
	var n int
	g.WithWriteLock(func() { n = g.SacrificedWithCardTypeThisTurn(p, "artifact") })
	return n
}

func sacrificeForTest(t *testing.T, g *Game, p, id uuid.UUID) {
	t.Helper()
	if err := g.SacrificePermanent(p, id); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
}

// A token counts: the Clue is gone (CR 704.5d) by the time anyone asks,
// and the tally recorded it at the sacrifice. The Food tally beside it
// is unchanged.
func TestSacrificedArtifactCountsAToken(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	clue := pushToken(g, me, "Clue", "Token Artifact — Clue")
	food := pushToken(g, me, "Food", "Token Artifact — Food")
	sacrificeForTest(t, g, me.ID, clue)
	nowhere(t, g, clue)
	if got := sacrificedArtifacts(g, me.ID); got != 1 {
		t.Errorf("a sacrificed Clue token counted %d artifacts, want 1", got)
	}
	sacrificeForTest(t, g, me.ID, food)
	if got := sacrificedArtifacts(g, me.ID); got != 2 {
		t.Errorf("after a Food too: %d artifacts, want 2", got)
	}
	g.WithWriteLock(func() {
		if n := g.SacrificedWithSubtypeThisTurn(me.ID, "Food"); n != 1 {
			t.Errorf("the Food readers' tally = %d, want 1", n)
		}
	})
}

// An artifact creature is one artifact: each card type it has is bumped
// once, so it also counts once as a creature.
func TestSacrificedArtifactCountsAnArtifactCreatureOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	thopter := celebrationEnter(t, g, me, "Ornithopter", "Artifact Creature — Thopter")
	sacrificeForTest(t, g, me, thopter)
	if got := sacrificedArtifacts(g, me); got != 1 {
		t.Errorf("an artifact creature counted %d artifacts, want 1", got)
	}
	g.WithWriteLock(func() {
		if n := g.SacrificedWithCardTypeThisTurn(me, "Creature"); n != 1 {
			t.Errorf("as a creature: %d, want 1", n)
		}
	})
}

// A non-artifact sacrifice does not count, and the count is the
// sacrificing player's.
func TestSacrificedArtifactIgnoresANonArtifactAndOtherPlayers(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	bear := celebrationEnter(t, g, me, "Grizzly Bears", "Creature — Bear")
	rock := celebrationEnter(t, g, opp, "Mind Stone", "Artifact")
	sacrificeForTest(t, g, me, bear)
	if got := sacrificedArtifacts(g, me); got != 0 {
		t.Errorf("a sacrificed creature counted %d artifacts, want 0", got)
	}
	sacrificeForTest(t, g, opp, rock)
	if got := sacrificedArtifacts(g, me); got != 0 {
		t.Errorf("an opponent's sacrifice counted %d for me, want 0", got)
	}
	if got := sacrificedArtifacts(g, opp); got != 1 {
		t.Errorf("the opponent's own count is %d, want 1", got)
	}
}

func TestSacrificedArtifactResetsNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	rock := celebrationEnter(t, g, me, "Mind Stone", "Artifact")
	sacrificeForTest(t, g, me, rock)
	if sacrificedArtifacts(g, me) != 1 {
		t.Fatal("setup: the sacrifice was not counted")
	}
	seat := g.Turn.ActiveSeat
	for i := 0; i < 40 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == seat {
		t.Fatal("setup: the turn never passed")
	}
	if got := sacrificedArtifacts(g, me); got != 0 {
		t.Errorf("the count survived into the next turn: %d", got)
	}
}

// A restore taken mid-turn keeps the count, through a JSON snapshot and
// a Clone (undo).
func TestSacrificedArtifactSurvivesTheSnapshotRoundTrip(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0].ID
	rock := celebrationEnter(t, g, me, "Mind Stone", "Artifact")
	sacrificeForTest(t, g, me, rock)
	back, err := throughJSON(t, g.CaptureSnapshot()).Restore()
	if err != nil {
		t.Fatal(err)
	}
	if got := sacrificedArtifacts(back, me); got != 1 {
		t.Errorf("after the snapshot round trip: %d, want 1", got)
	}
	if got := sacrificedArtifacts(g.Clone(), me); got != 1 {
		t.Errorf("after Clone: %d, want 1", got)
	}
}
