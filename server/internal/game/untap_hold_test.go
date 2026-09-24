package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// untap_hold_test.go — #1313, ADR 0058's 2026-09-23 amendment: an
// UntapSkip that carries a CR 611.2 Duration ("it doesn't untap during
// its controller's untap step for as long as you control Ty Lee").

// holdWhileYouControl records a "for as long as <controller> controls
// <source>" hold on target, failing the test if the duration never
// starts.
func holdWhileYouControl(t *testing.T, g *Game, source, controller, target uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		d, ok := g.ForAsLongAsYouControlDuration(source, controller)
		if !ok {
			t.Fatal("duration did not start with the source on the battlefield under its controller")
		}
		if err := g.HoldUntappedForEffect(target, d); err != nil {
			t.Fatal(err)
		}
	})
}

func TestUntapHoldLastsEveryControllerStepUntilTheSourceLeaves(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	source := pushTappedPermanent(g, a.ID, "Ty Lee", "", "Creature", false)
	target := pushTappedPermanent(g, b.ID, "Held", "", "Creature", true)
	holdWhileYouControl(t, g, source, a.ID, target)
	// A second identical hold is not added again.
	holdWhileYouControl(t, g, source, a.ID, target)
	if got := len(cardByIDForUntapTest(g, target).NextUntapSkips); got != 1 {
		t.Fatalf("hold count after a duplicate = %d, want 1", got)
	}

	for step := 0; step < 2; step++ {
		g.WithWriteLock(func() { g.performUntapStepLocked(b.Seat) })
		c := cardByIDForUntapTest(g, target)
		if !c.Tapped {
			t.Fatalf("untap step %d: held creature untapped", step+1)
		}
		if len(c.NextUntapSkips) != 1 {
			t.Fatalf("untap step %d used the hold up: %#v", step+1, c.NextUntapSkips)
		}
	}

	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: a.ID}, source); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.performUntapStepLocked(b.Seat) })
	c := cardByIDForUntapTest(g, target)
	if c.Tapped {
		t.Fatal("creature stayed tapped after the hold's source left the battlefield")
	}
	if len(c.NextUntapSkips) != 0 {
		t.Fatalf("expired hold was not dropped: %#v", c.NextUntapSkips)
	}
}

// The Dungeon Geists ruling (2019-07-12): "If another player gains
// control of Dungeon Geists, its effect expires. It won't keep the
// creature from untapping anymore, even if you later regain control."
// No untap step runs between the loss and the regain, and exactly ONE
// layer pass sees b in control, so the sweep at the end of that pass
// is the only thing that can have ended the hold.
func TestUntapHoldEndsForGoodWhenControlOfTheSourceIsLost(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	source := pushTappedPermanent(g, a.ID, "Dungeon Geists", "", "Creature", false)
	target := pushTappedPermanent(g, b.ID, "Held", "", "Creature", true)
	holdWhileYouControl(t, g, source, a.ID, target)

	g.WithWriteLock(func() {
		if !g.GainControlForEffect(source, source, b.ID, g.UntilEndOfTurnDuration(), "threaten") {
			t.Fatal("control change refused")
		}
		g.recomputeLayersLocked()
		if cardByIDForUntapTest(g, source).Controller != b.ID {
			t.Fatal("control did not move")
		}
		g.ClearEndOfTurnScopedStaticsLocked()
		g.recomputeLayersLocked()
		if cardByIDForUntapTest(g, source).Controller != a.ID {
			t.Fatal("control did not come back")
		}
		g.performUntapStepLocked(b.Seat)
	})
	if cardByIDForUntapTest(g, target).Tapped {
		t.Fatal("hold came back when control of its source was regained")
	}
}

// "Its controller's untap step" is read at each step, and only the
// controller's own step is narrowed: a Seedborn-style untap during
// another player's step still untaps a held permanent, as it does for
// an ADR 0058 Decision 1 restriction.
func TestUntapHoldFollowsTheHeldPermanentsControllerAndOnlyTheirStep(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	const muse = "hold-seedborn"
	source := pushTappedPermanent(g, a.ID, "Source", "", "Creature", false)
	target := pushTappedPermanent(g, b.ID, "Held", "", "Creature", true)
	pushTappedPermanent(g, a.ID, "Muse", muse, "Creature", false)
	withCatalogUntapStepPermissions(t, func(key string) []UntapStepPermission {
		if key != muse {
			return nil
		}
		return []UntapStepPermission{{
			AppliesTo: func(_ *Game, source *Card, player uuid.UUID) bool { return source.Controller != player },
			Untaps:    func(_ *Game, source, target *Card) bool { return target.Controller == source.Controller },
		}}
	})
	holdWhileYouControl(t, g, source, a.ID, target)

	// Control of the held creature moves to a: a's untap step holds it.
	g.WithWriteLock(func() {
		cardByIDForUntapTest(g, target).Controller = a.ID
		g.performUntapStepLocked(a.Seat)
	})
	if !cardByIDForUntapTest(g, target).Tapped {
		t.Fatal("hold did not follow the held creature to its new controller's untap step")
	}
	// a's Seedborn Muse untaps it during b's untap step.
	g.WithWriteLock(func() { g.performUntapStepLocked(b.Seat) })
	if cardByIDForUntapTest(g, target).Tapped {
		t.Fatal("hold stopped a Seedborn-style untap during another player's untap step")
	}
	if got := len(cardByIDForUntapTest(g, target).NextUntapSkips); got != 1 {
		t.Fatalf("hold was used up by another player's step: %d entries", got)
	}
}

func TestUntapHoldWhileSourceRemainsTapped(t *testing.T) {
	g := newActiveGame(t)
	a := g.Seats[0]
	source := pushTappedPermanent(g, a.ID, "Rust Tick", "", "Artifact Creature", true)
	target := pushTappedPermanent(g, a.ID, "Held artifact", "", "Artifact", true)
	upright := pushTappedPermanent(g, a.ID, "Upright", "", "Creature", false)
	g.WithWriteLock(func() {
		if _, ok := g.ForAsLongAsSourceTappedDuration(upright); ok {
			t.Fatal("an untapped source started a 'remains tapped' duration")
		}
		d, ok := g.ForAsLongAsSourceTappedDuration(source)
		if !ok {
			t.Fatal("tapped source did not start the duration")
		}
		_ = g.HoldUntappedForEffect(target, d)
		// CR 502.3: the set is determined before anything untaps, so
		// the held artifact stays tapped in the same step the source
		// untaps in.
		g.performUntapStepLocked(a.Seat)
	})
	if cardByIDForUntapTest(g, source).Tapped {
		t.Fatal("source did not untap")
	}
	if !cardByIDForUntapTest(g, target).Tapped {
		t.Fatal("held artifact untapped in the same step as its source")
	}
	// The source has untapped: the duration is over, and tapping the
	// source again does not revive it.
	g.WithWriteLock(func() {
		g.recomputeLayersLocked()
		cardByIDForUntapTest(g, source).Tapped = true
		g.performUntapStepLocked(a.Seat)
	})
	if cardByIDForUntapTest(g, target).Tapped {
		t.Fatal("held artifact stayed tapped after its source untapped")
	}
}

func TestUntapHoldCloneSnapshotAndZoneReset(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	source := pushTappedPermanent(g, a.ID, "Source", "", "Creature", false)
	target := pushTappedPermanent(g, b.ID, "Held", "", "Creature", true)
	holdWhileYouControl(t, g, source, a.ID, target)
	want := *cardByIDForUntapTest(g, target).NextUntapSkips[0].While

	clone := g.Clone()
	g.WithWriteLock(func() { g.RestoreFrom(clone) })
	if got := cardByIDForUntapTest(g, target).NextUntapSkips; len(got) != 1 || got[0].While == nil || *got[0].While != want {
		t.Fatalf("RestoreFrom hold = %#v", got)
	}

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	restored.WithWriteLock(func() {
		held := cardByIDForUntapTest(restored, target)
		if len(held.NextUntapSkips) != 1 || held.NextUntapSkips[0].While == nil || *held.NextUntapSkips[0].While != want {
			t.Fatalf("snapshot-restored hold = %#v", held.NextUntapSkips)
		}
		restored.performUntapStepLocked(b.Seat)
		if !cardByIDForUntapTest(restored, target).Tapped {
			t.Fatal("restored hold did not hold")
		}
	})

	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: b.ID}, target); err != nil {
		t.Fatal(err)
	}
	if got := g.Seats[1].Graveyard.Cards[0].NextUntapSkips; len(got) != 0 {
		t.Fatalf("zone change kept the hold: %#v", got)
	}
}

func TestUntapHeldLockedReportsOnlyLiveHolds(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	source := pushTappedPermanent(g, a.ID, "Source", "", "Creature", false)
	held := pushTappedPermanent(g, b.ID, "Held", "", "Creature", true)
	marked := pushTappedPermanent(g, b.ID, "Marked", "", "Creature", true)
	holdWhileYouControl(t, g, source, a.ID, held)
	g.WithWriteLock(func() { _ = g.SkipNextUntapForEffect(marked, uuid.Nil) })
	g.WithWriteLock(func() {
		if !g.UntapHeldLocked(cardByIDForUntapTest(g, held)) {
			t.Error("live hold not reported")
		}
		if g.UntapHeldLocked(cardByIDForUntapTest(g, marked)) {
			t.Error("a one-shot marker was reported as a hold")
		}
		cardByIDForUntapTest(g, source).Controller = b.ID
		if g.UntapHeldLocked(cardByIDForUntapTest(g, held)) {
			t.Error("expired hold still reported before the sweep")
		}
	})
}
