package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// untap_cant_test.go — #4574: "can't become untapped" (CR 614.17). The
// primitive every untap goes through refuses it, whoever asks.

// lockWhileYouControl records a "can't become untapped for as long as
// <controller> controls <source>" lock on target.
func lockWhileYouControl(t *testing.T, g *Game, source, controller, target uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		d, ok := g.ForAsLongAsYouControlDuration(source, controller)
		if !ok {
			t.Fatal("duration did not start with the source on the battlefield under its controller")
		}
		if err := g.CantUntapForEffect(target, d); err != nil {
			t.Fatal(err)
		}
	})
}

func TestCantBecomeUntappedHoldBlocksEveryUntapPathUntilTheSourceLeaves(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	source := pushTappedPermanent(g, a.ID, "Locker", "", "Creature", false)
	target := pushTappedPermanent(g, b.ID, "Locked", "", "Creature", true)
	lockWhileYouControl(t, g, source, a.ID, target)
	// A second identical lock is not added again, and a plain hold with the
	// same duration is a different entry.
	lockWhileYouControl(t, g, source, a.ID, target)
	if got := len(cardByIDForUntapTest(g, target).NextUntapSkips); got != 1 {
		t.Fatalf("lock count after a duplicate = %d, want 1", got)
	}
	stayed := func(what string) {
		t.Helper()
		if !cardByIDForUntapTest(g, target).Tapped {
			t.Fatalf("%s untapped a creature that can't become untapped", what)
		}
		if n := untapEventsFor(g, target); n != 0 {
			t.Fatalf("%s announced %d untap events for it, want 0", what, n)
		}
	}

	g.WithWriteLock(func() { g.performUntapStepLocked(b.Seat) })
	stayed("its controller's untap step")

	// An effect ("untap target creature").
	if err := untapEffect(g, target); err != nil {
		t.Fatal(err)
	}
	stayed("an untap effect")

	// The sandbox untap and untap-everything.
	if err := g.TapCard(target, false); err != nil {
		t.Fatal(err)
	}
	stayed("the sandbox untap")
	if err := g.UntapAll(b.ID); err != nil {
		t.Fatal(err)
	}
	stayed("untap all")

	// Another player's step, through an untap-step permission (Seedborn
	// Muse), is not the controller's step and must not slip through either.
	const muse = "cant-untap-seedborn"
	pushTappedPermanent(g, b.ID, "Muse", muse, "Creature", false)
	withCatalogUntapStepPermissions(t, func(key string) []UntapStepPermission {
		if key != muse {
			return nil
		}
		return []UntapStepPermission{{
			AppliesTo: func(_ *Game, source *Card, player uuid.UUID) bool { return source.Controller != player },
			Untaps:    func(_ *Game, source, target *Card) bool { return target.Controller == source.Controller },
		}}
	})
	g.WithWriteLock(func() { g.performUntapStepLocked(a.Seat) })
	stayed("another player's untap step with a Seedborn Muse out")

	// The wire's reading.
	g.WithWriteLock(func() {
		c := cardByIDForUntapTest(g, target)
		if !g.CantBecomeUntappedLocked(c) || !g.UntapHeldLocked(c) || !g.CantBecomeUntappedCheckerLocked()(c) {
			t.Error("a live lock is not reported by the readers")
		}
	})

	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: a.ID}, source); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if g.CantBecomeUntappedLocked(cardByIDForUntapTest(g, target)) {
			t.Error("the lock is still reported after its source left")
		}
	})
	if err := untapEffect(g, target); err != nil {
		t.Fatal(err)
	}
	if cardByIDForUntapTest(g, target).Tapped {
		t.Fatal("an untap effect did nothing after the lock's source left")
	}
}

// A creature that can't become untapped does not lose its stun counter to
// an untap that does not happen.
func TestCantBecomeUntappedKeepsAStunCounter(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	source := pushTappedPermanent(g, a.ID, "Locker", "", "Creature", false)
	target := pushTappedPermanent(g, b.ID, "Locked", "", "Creature", true)
	cardByIDForUntapTest(g, target).Counters = map[string]int{CounterStun: 1}
	lockWhileYouControl(t, g, source, a.ID, target)
	g.WithWriteLock(func() { g.untapPermanentLocked(cardByIDForUntapTest(g, target)) })
	if got := cardByIDForUntapTest(g, target).Counters[CounterStun]; got != 1 {
		t.Fatalf("stun counters = %d, want the one it started with", got)
	}
}

// A plain hold stays what it was: it keeps only its controller's step.
func TestPlainHoldStillAllowsAnUntapEffect(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	source := pushTappedPermanent(g, a.ID, "Ty Lee", "", "Creature", false)
	target := pushTappedPermanent(g, b.ID, "Held", "", "Creature", true)
	holdWhileYouControl(t, g, source, a.ID, target)
	g.WithWriteLock(func() {
		if g.CantBecomeUntappedLocked(cardByIDForUntapTest(g, target)) {
			t.Error("a plain hold was read as can't-become-untapped")
		}
	})
	if err := untapEffect(g, target); err != nil {
		t.Fatal(err)
	}
	if cardByIDForUntapTest(g, target).Tapped {
		t.Fatal("an untap effect should still untap a creature that only misses its untap step")
	}
}

// A static source (Frozen in Ice) marked CantBecomeUntapped blocks every
// untap, where an ordinary restriction blocks only the controller's step.
func TestStaticCantBecomeUntappedBlocksEffectsButOrdinaryRestrictionDoesNot(t *testing.T) {
	g := newActiveGame(t)
	a := g.Seats[0]
	const frozen, plain = "cant-untap-aura", "doesnt-untap-aura"
	locked := pushTappedPermanent(g, a.ID, "Locked", "", "Creature", true)
	slowed := pushTappedPermanent(g, a.ID, "Slowed", "", "Creature", true)
	pushTappedPermanent(g, a.ID, "Frozen", frozen, "Enchantment — Aura", false)
	pushTappedPermanent(g, a.ID, "Plain", plain, "Enchantment — Aura", false)
	withCatalogUntapStepRestrictions(t, func(key string) []UntapStepRestriction {
		switch key {
		case frozen:
			return []UntapStepRestriction{{
				CantBecomeUntapped: true,
				Restricts:          func(target *Card, _ *Game, _ *Card) bool { return target.InstanceID == locked },
			}}
		case plain:
			return []UntapStepRestriction{{
				Restricts: func(target *Card, _ *Game, _ *Card) bool { return target.InstanceID == slowed },
			}}
		}
		return nil
	})
	g.WithWriteLock(func() { g.performUntapStepLocked(a.Seat) })
	if !cardByIDForUntapTest(g, locked).Tapped || !cardByIDForUntapTest(g, slowed).Tapped {
		t.Fatal("the untap step untapped a restricted permanent")
	}
	if err := untapEffect(g, locked); err != nil {
		t.Fatal(err)
	}
	if err := untapEffect(g, slowed); err != nil {
		t.Fatal(err)
	}
	if !cardByIDForUntapTest(g, locked).Tapped {
		t.Error("an untap effect untapped a permanent that can't become untapped")
	}
	if cardByIDForUntapTest(g, slowed).Tapped {
		t.Error("an untap effect should untap a permanent that only misses its untap step")
	}
}

func TestCantBecomeUntappedHoldSurvivesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	source := pushTappedPermanent(g, a.ID, "Locker", "", "Creature", false)
	target := pushTappedPermanent(g, b.ID, "Locked", "", "Creature", true)
	lockWhileYouControl(t, g, source, a.ID, target)

	clone := g.Clone()
	g.WithWriteLock(func() { g.RestoreFrom(clone) })
	if got := cardByIDForUntapTest(g, target).NextUntapSkips; len(got) != 1 || !got[0].CantBecomeUntapped {
		t.Fatalf("RestoreFrom lock = %#v", got)
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
	if err := untapEffect(restored, target); err != nil {
		t.Fatal(err)
	}
	restored.WithWriteLock(func() {
		if !cardByIDForUntapTest(restored, target).Tapped {
			t.Fatal("the restored lock did not hold against an untap effect")
		}
	})
}

// untapEffect is "untap target permanent" as a resolving effect: under the
// write lock, through the effect API.
func untapEffect(g *Game, id uuid.UUID) error {
	var err error
	g.WithWriteLock(func() { err = g.UntapTargetForEffect(id) })
	return err
}
