package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_tally_void_test.go — #2128. TurnTally.NonlandPermanentsLeft and
// SpellsWarped are void (CR 207.2c): a nonland permanent left the
// battlefield this turn, or a spell was cast for its warp cost this turn
// (CR 702.185c). Both are table-wide.

func voidActive(g *Game) bool {
	var got bool
	g.WithWriteLock(func() { got = g.VoidThisTurn() })
	return got
}

func TestVoidCountsEveryRouteOfANonlandPermanent(t *testing.T) {
	routes := map[string]func(g *Game, id uuid.UUID) error{
		"sacrifice": func(g *Game, id uuid.UUID) error { return g.SacrificePermanentForEffect(id) },
		"bounce":    func(g *Game, id uuid.UUID) error { return g.BounceToHandForEffect(id) },
		"exile":     func(g *Game, id uuid.UUID) error { return g.ExileCardForEffect(id) },
		"destroy":   func(g *Game, id uuid.UUID) error { return g.DestroyPermanentForEffect(id) },
	}
	for name, route := range routes {
		t.Run(name, func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			rock := revoltPush(g, g.Seats[0].ID, "Sol Ring", "Artifact")
			if voidActive(g) {
				t.Fatal("void before anything left")
			}
			var err error
			g.WithWriteLock(func() { err = route(g, rock) })
			if err != nil {
				t.Fatal(err)
			}
			if !voidActive(g) {
				t.Errorf("%s of a nonland permanent did not turn void on", name)
			}
		})
	}
}

// A land leaving is not void, however it leaves, and neither is a land
// that is also a creature (CR 205.2b: it satisfies every effect that applies to a land).
func TestVoidIgnoresALandLeaving(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	forest := revoltPush(g, me, "Forest", "Basic Land — Forest")
	arbor := revoltPush(g, me, "Dryad Arbor", "Land Creature — Forest Dryad")
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(forest); err != nil {
			t.Fatal(err)
		}
		if err := g.DestroyPermanentForEffect(arbor); err != nil {
			t.Fatal(err)
		}
	})
	if voidActive(g) {
		t.Error("a land leaving turned void on")
	}
	// The per-player revolt cell still saw both: the two are different facts.
	if !revoltLeft(g, me) {
		t.Error("revolt did not see the lands leave")
	}
}

func TestVoidCountsATokenCeasingToExist(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	tok := revoltPush(g, g.Seats[0].ID, "Servo", "Token Artifact Creature — Servo")
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(tok); err != nil {
			t.Fatal(err)
		}
	})
	if !voidActive(g) {
		t.Error("a token dying did not turn void on")
	}
}

// Void is a fact about the table, not about a seat: an opponent's
// permanent leaving counts for everyone.
func TestVoidIsTableWide(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	theirs := revoltPush(g, g.Seats[2].ID, "Bear", "Creature — Bear")
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(theirs); err != nil {
			t.Fatal(err)
		}
	})
	if !voidActive(g) {
		t.Error("an opponent's permanent leaving did not turn void on")
	}
}

// Only a cast for the warp cost is warped: another alternative cost, the
// printed cost and no cost at all are not.
func TestVoidCountsASpellCastForItsWarpCostOnly(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.WithWriteLock(func() {
		g.noteSpellWarpedLocked(nil)
		g.noteSpellWarpedLocked(&AlternativeCost{Key: "evoke", ManaCost: "{1}"})
	})
	if voidActive(g) {
		t.Fatal("a cast that was not for the warp cost turned void on")
	}
	g.WithWriteLock(func() { g.noteSpellWarpedLocked(&AlternativeCost{Key: "warp", ManaCost: "{R}", WarpExile: true}) })
	if !voidActive(g) {
		t.Error("a cast for the warp cost did not turn void on")
	}
	if got := g.TurnTally.SpellsWarped; got != 1 {
		t.Errorf("SpellsWarped = %d, want 1", got)
	}
}

func TestVoidResetsNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	rock := revoltPush(g, g.Seats[0].ID, "Sol Ring", "Artifact")
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(rock)
		g.noteSpellWarpedLocked(&AlternativeCost{WarpExile: true})
	})
	if !voidActive(g) {
		t.Fatal("setup: void not recorded")
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
	if voidActive(g) {
		t.Error("void survived into the next turn")
	}
	if g.TurnTally.NonlandPermanentsLeft != 0 || g.TurnTally.SpellsWarped != 0 {
		t.Errorf("cells not cleared: %+v", g.TurnTally)
	}
}

// A restore taken while void is live keeps it, through a JSON snapshot, a
// Clone (undo) and each cell on its own.
func TestVoidSurvivesTheSnapshotRoundTrip(t *testing.T) {
	for name, setup := range map[string]func(g *Game){
		"permanent left": func(g *Game) {
			rock := revoltPush(g, g.Seats[0].ID, "Sol Ring", "Artifact")
			g.WithWriteLock(func() { _ = g.ExileCardForEffect(rock) })
		},
		"spell warped": func(g *Game) {
			g.WithWriteLock(func() { g.noteSpellWarpedLocked(&AlternativeCost{WarpExile: true}) })
		},
	} {
		t.Run(name, func(t *testing.T) {
			g := newRestorableGame(t)
			setup(g)
			back, err := throughJSON(t, g.CaptureSnapshot()).Restore()
			if err != nil {
				t.Fatal(err)
			}
			if !voidActive(back) {
				t.Error("void lost across the snapshot round trip")
			}
			if !voidActive(g.Clone()) {
				t.Error("void lost across Clone")
			}
		})
	}
}
