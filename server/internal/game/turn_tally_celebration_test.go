package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_tally_celebration_test.go — #5164. PlayerTurnTally.
// NonlandPermanentsEntered is celebration's count (CR 207.2c): the
// nonland permanents that entered the battlefield under a player's
// control this turn, each entry once whatever its types, tokens
// included, one that has since left still counted.

func celebrationCount(g *Game, p uuid.UUID) int {
	var n int
	g.WithWriteLock(func() { n = g.NonlandPermanentsEnteredThisTurn(p) })
	return n
}

// celebrationEnter puts a card from its owner's hand onto the battlefield
// under that player's control, through the ordinary entry pipeline, and
// returns the permanent's ID.
func celebrationEnter(t *testing.T, g *Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		g.PlayerByIDForEffect(owner).Hand.PushTop(Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 1, Toughness: 1, Owner: owner, Controller: owner})
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, HandEntryOptions{Controller: owner})
	})
	if err != nil || entered == uuid.Nil {
		t.Fatalf("put %s onto the battlefield: %v (entered %v)", name, err, entered)
	}
	return entered
}

// An artifact creature is one permanent, not one per card type.
func TestCelebrationCountsAnArtifactCreatureOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	celebrationEnter(t, g, me, "Ornithopter", "Artifact Creature — Thopter")
	if got := celebrationCount(g, me); got != 1 {
		t.Errorf("an artifact creature entering counted %d, want 1", got)
	}
	celebrationEnter(t, g, me, "Sol Ring", "Artifact")
	if got := celebrationCount(g, me); got != 2 {
		t.Errorf("after an artifact too: %d, want 2", got)
	}
}

func TestCelebrationCountsATokenEntering(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me, Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1}, 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := celebrationCount(g, me); got != 2 {
		t.Errorf("two tokens created counted %d, want 2", got)
	}
}

// A land does not count, and neither does a land that is also a creature
// (CR 205.2b: it is a land).
func TestCelebrationIgnoresALandEntering(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	celebrationEnter(t, g, me, "Forest", "Basic Land — Forest")
	celebrationEnter(t, g, me, "Dryad Arbor", "Land Creature — Forest Dryad")
	if got := celebrationCount(g, me); got != 0 {
		t.Errorf("two lands entering counted %d, want 0", got)
	}
	if got := g.TurnTallyFor(me).LandsEntered; got != 2 {
		t.Errorf("landfall's tally saw %d lands, want 2: the two are different facts", got)
	}
}

// One that entered and then left still entered this turn.
func TestCelebrationCountsAPermanentThatHasSinceLeft(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	bear := celebrationEnter(t, g, me, "Grizzly Bears", "Creature — Bear")
	rock := celebrationEnter(t, g, me, "Sol Ring", "Artifact")
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
		if err := g.ExileCardForEffect(rock); err != nil {
			t.Fatal(err)
		}
	})
	if got := celebrationCount(g, me); got != 2 {
		t.Errorf("after both left: %d, want 2", got)
	}
}

// The count is the controller's: an opponent's permanent entering is
// theirs, not mine.
func TestCelebrationIsPerPlayer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	celebrationEnter(t, g, opp, "Grizzly Bears", "Creature — Bear")
	if got := celebrationCount(g, me); got != 0 {
		t.Errorf("an opponent's permanent counted %d for me, want 0", got)
	}
	if got := celebrationCount(g, opp); got != 1 {
		t.Errorf("the opponent's own count is %d, want 1", got)
	}
}

func TestCelebrationResetsNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	celebrationEnter(t, g, me, "Grizzly Bears", "Creature — Bear")
	celebrationEnter(t, g, me, "Sol Ring", "Artifact")
	if celebrationCount(g, me) != 2 {
		t.Fatal("setup: the entries were not counted")
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
	if got := celebrationCount(g, me); got != 0 {
		t.Errorf("the count survived into the next turn: %d", got)
	}
}

// A restore taken mid-turn keeps the count, through a JSON snapshot and
// a Clone (undo).
func TestCelebrationSurvivesTheSnapshotRoundTrip(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0].ID
	bear := celebrationEnter(t, g, me, "Grizzly Bears", "Creature — Bear")
	celebrationEnter(t, g, me, "Sol Ring", "Artifact")
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	back, err := throughJSON(t, g.CaptureSnapshot()).Restore()
	if err != nil {
		t.Fatal(err)
	}
	if got := celebrationCount(back, me); got != 2 {
		t.Errorf("after the snapshot round trip: %d, want 2", got)
	}
	if got := celebrationCount(g.Clone(), me); got != 2 {
		t.Errorf("after Clone: %d, want 2", got)
	}
}
