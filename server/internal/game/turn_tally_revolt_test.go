package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_tally_revolt_test.go — #2148. PlayerTurnTally.PermanentsLeft is
// revolt (CR 702.136): any permanent the player controlled leaving the
// battlefield, by any route, to any zone.

func revoltPush(g *Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: owner, Controller: owner})
	})
	return id
}

func revoltLeft(g *Game, p uuid.UUID) bool {
	var got bool
	g.WithWriteLock(func() { got = g.PermanentLeftThisTurn(p) })
	return got
}

func TestRevoltCountsEveryRoute(t *testing.T) {
	routes := map[string]func(g *Game, id uuid.UUID) error{
		"sacrifice": func(g *Game, id uuid.UUID) error { return g.SacrificePermanentForEffect(id) },
		"bounce":    func(g *Game, id uuid.UUID) error { return g.BounceToHandForEffect(id) },
		"exile":     func(g *Game, id uuid.UUID) error { return g.ExileCardForEffect(id) },
		"destroy":   func(g *Game, id uuid.UUID) error { return g.DestroyPermanentForEffect(id) },
	}
	for name, route := range routes {
		t.Run(name, func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			me := g.Seats[0].ID
			land := revoltPush(g, me, "Forest", "Basic Land — Forest")
			if revoltLeft(g, me) {
				t.Fatal("revolt before anything left")
			}
			var err error
			g.WithWriteLock(func() { err = route(g, land) })
			if err != nil {
				t.Fatal(err)
			}
			if !revoltLeft(g, me) {
				t.Errorf("%s of a land did not count", name)
			}
		})
	}
}

func TestRevoltCountsATokenDying(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	tok := revoltPush(g, me, "Servo", "Token Artifact Creature — Servo")
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(tok); err != nil {
			t.Fatal(err)
		}
	})
	if !revoltLeft(g, me) {
		t.Error("a token dying did not count")
	}
}

func TestRevoltIsPerController(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	theirs := revoltPush(g, opp, "Bear", "Creature — Bear")
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(theirs); err != nil {
			t.Fatal(err)
		}
	})
	if revoltLeft(g, me) {
		t.Error("an opponent's permanent leaving counted for me")
	}
	if !revoltLeft(g, opp) {
		t.Error("the opponent's own revolt was not recorded")
	}
}

func TestRevoltCountsAPermanentThatCameBack(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	id := revoltPush(g, me, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.MoveCardByID(zoneRefOf(g.Seats[0].Hand), zoneRefOf(g.Battlefield), id); err != nil {
		t.Fatal(err)
	}
	if !revoltLeft(g, me) {
		t.Error("a permanent that left and returned did not count")
	}
}

func TestRevoltResetsNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	id := revoltPush(g, me, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
	if !revoltLeft(g, me) {
		t.Fatal("setup: revolt not recorded")
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
	if revoltLeft(g, me) {
		t.Error("revolt survived into the next turn")
	}
}

func TestRevoltSurvivesTheSnapshotRoundTrip(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0].ID
	id := revoltPush(g, me, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(id) })
	back, err := throughJSON(t, g.CaptureSnapshot()).Restore()
	if err != nil {
		t.Fatal(err)
	}
	if !revoltLeft(back, me) {
		t.Error("revolt lost across the snapshot round trip")
	}
	cl := g.Clone()
	if !revoltLeft(cl, me) {
		t.Error("revolt lost across Clone")
	}
}
