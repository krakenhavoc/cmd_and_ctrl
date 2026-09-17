package game

import (
	"testing"

	"github.com/google/uuid"
)

// battlefield_exit_damage_test.go is the #816 regression suite: which
// battlefield exits clear the damage marked on a permanent.
//
// #813 (ADR 0013 §5d) moved the clear off the destroy entry point and
// onto the landed outcome of a destruction, which fixed the case a
// replacement rewrote — and left every OTHER exit clearing nothing.
// An exiled or bounced creature kept the number: it showed in the new
// zone, and it came back onto the battlefield with the card when it
// was replayed, to die to the first ping. CR 400.7: the card in its
// new zone is a NEW OBJECT, with no memory of the damage the old one
// took.
//
// The clear now lives in MoveCard's one battlefield-exit cleanup, so
// it is true of every route off the battlefield by construction. What
// these tests pin is that it is still true LATE — after the CR 614
// window, after the LKI snapshot — because clearing early is the bug
// #708 fixed and it must not come back by the front door.

// damagedCreature (destroy_damage_test.go) seats a damaged creature.

// exiledCard returns the card with this ID from exile, or fails.
func exiledCard(t *testing.T, g *Game, id uuid.UUID) *Card {
	t.Helper()
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == id {
			return &g.Exile.Cards[i]
		}
	}
	t.Fatalf("card %s is not in exile", id)
	return nil
}

// zoneCard returns the card with this ID from the given zone, or fails.
func zoneCard(t *testing.T, z *Zone, id uuid.UUID) *Card {
	t.Helper()
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			return &z.Cards[i]
		}
	}
	t.Fatalf("card %s is not in %s", id, z.Kind)
	return nil
}

// TestExiledCreatureLeavesItsDamageOnTheBattlefield is the issue, in
// its shortest form: the number does not travel.
func TestExiledCreatureLeavesItsDamageOnTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 4, 3)

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})

	if c := exiledCard(t, g, id); c.DamageMarked != 0 {
		t.Errorf("DamageMarked in exile = %d, want 0 (CR 400.7: a new object)", c.DamageMarked)
	}
}

// TestFlickeredCreatureReturnsUndamaged is the shape a player
// actually meets: Flicker, Cloudshift, Ephemerate. Out to exile and
// straight back, and the 4/4 that comes back has taken no damage.
//
// The mid-flight assertion is the load-bearing one. The exile →
// battlefield return mints a new instance and scrubs the card on the
// way in (effect_api.go), so the far end of the flicker was already
// clean; what was NOT clean was the card sitting in exile in between,
// which is what every other reader of that zone sees.
func TestFlickeredCreatureReturnsUndamaged(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 4, 3)

	var newID uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
		if c := exiledCard(t, g, id); c.DamageMarked != 0 {
			t.Errorf("DamageMarked mid-flicker = %d, want 0", c.DamageMarked)
		}
		var err error
		newID, err = g.ReturnFromExileToBattlefieldForEffect(id, owner.ID, false)
		if err != nil {
			t.Fatalf("ReturnFromExileToBattlefieldForEffect: %v", err)
		}
	})

	back := findBattlefieldCard(g, newID)
	if back == nil {
		t.Fatal("the flickered creature did not come back")
	}
	if back.DamageMarked != 0 {
		t.Errorf("DamageMarked after the flicker = %d, want 0", back.DamageMarked)
	}
}

// TestReanimatedCreatureArrivesUndamaged — the graveyard half. The
// destroy path has cleared the damage since #813; this pins that the
// card the reanimation picks up is clean whatever route put it there.
func TestReanimatedCreatureArrivesUndamaged(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 4, 3)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
		if c := zoneCard(t, owner.Graveyard, id); c.DamageMarked != 0 {
			t.Errorf("DamageMarked in the graveyard = %d, want 0", c.DamageMarked)
		}
		if err := g.ReturnFromGraveyardForEffect(id, ZoneBattlefield); err != nil {
			t.Fatalf("ReturnFromGraveyardForEffect: %v", err)
		}
	})

	back := findBattlefieldCard(g, id)
	if back == nil {
		t.Fatal("the reanimated creature is not on the battlefield")
	}
	if back.DamageMarked != 0 {
		t.Errorf("DamageMarked after the reanimation = %d, want 0", back.DamageMarked)
	}
}

// TestBouncedCreatureIsCleanInHandAndOnItsWayBack — Unsummon, then
// replayed. The hand copy carries nothing, and neither does the
// permanent it becomes.
func TestBouncedCreatureIsCleanInHandAndOnItsWayBack(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 4, 3)

	var newID uuid.UUID
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		if c := zoneCard(t, owner.Hand, id); c.DamageMarked != 0 {
			t.Errorf("DamageMarked in hand = %d, want 0", c.DamageMarked)
		}
		var err error
		newID, err = g.PutFromHandOntoBattlefieldForEffect(id, HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})

	back := findBattlefieldCard(g, newID)
	if back == nil {
		t.Fatal("the bounced creature did not come back down")
	}
	if back.DamageMarked != 0 {
		t.Errorf("DamageMarked after the replay = %d, want 0", back.DamageMarked)
	}
}

// TestDeathtouchMarkDoesNotSurviveTheExit — CR 702.2c's flag is the
// damage's companion and goes with it. A creature marked lethal that
// is exiled and returned is a new object, and the next state-based
// check must not kill it.
func TestDeathtouchMarkDoesNotSurviveTheExit(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 4, 0)
	findBattlefieldCard(g, id).MarkedLethalByDeathtouch = true

	var newID uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
		if c := exiledCard(t, g, id); c.MarkedLethalByDeathtouch {
			t.Error("the deathtouch mark followed the card into exile")
		}
		var err error
		newID, err = g.ReturnFromExileToBattlefieldForEffect(id, owner.ID, false)
		if err != nil {
			t.Fatalf("ReturnFromExileToBattlefieldForEffect: %v", err)
		}
		g.runStateChecksLocked()
	})

	if findBattlefieldCard(g, newID) == nil {
		t.Fatal("the returned creature was destroyed by a deathtouch mark it no longer carries")
	}
}

// TestTheExitWindowStillSeesTheDamage is #708's guarantee, restated
// for the exits that were not part of #708. The CR 614 window opens
// BEFORE anything moves, so a replacement that reads how much damage
// is on the permanent gets the real number — the clear is part of the
// move, not part of asking for it.
func TestTheExitWindowStillSeesTheDamage(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 4, 3)

	seen := -1
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == id
			},
			Replace: func(_ *ReplacementEvent, g *Game, _ *Card) error {
				if c := findBattlefieldCard(g, id); c != nil {
					seen = c.DamageMarked
				}
				return nil
			},
			Label: "reads the damage on its way out",
		})
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})

	if seen != 3 {
		t.Errorf("the replacement read %d damage, want 3 — the clear moved back before the window", seen)
	}
}

// TestDiesTriggerLKIStillDescribesTheCreatureThatDied — the LKI
// question the exit cleanup raises (CR 603.10).
//
// snapshotLKILocked runs while the permanent is still on the
// battlefield, one line before the MoveCard that now does the
// clearing, so an LTB trigger reads the creature that was there. The
// snapshot is a Characteristic and has never carried DamageMarked;
// what the trigger's `source` card carries is the object in the NEW
// zone, and CR 400.7 says that one has no damage on it. Both halves
// are asserted here so a future change to either is a visible one.
func TestDiesTriggerLKIStillDescribesTheCreatureThatDied(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-exit-lki-oracle"
	id := damagedCreature(g, owner, 4, 3)
	findBattlefieldCard(g, id).OracleID = oracle

	var lkiPower, lkiToughness, sourceDamage int
	var lkiName string
	withCatalogTriggers(t, func(oracleID string) []TriggeredAbility {
		if oracleID != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventLTB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ Event, source *Card, lki Characteristic, _ *Game) *StackItem {
				lkiPower, lkiToughness, lkiName = lki.Power, lki.Toughness, lki.Name
				sourceDamage = source.DamageMarked
				return nil
			},
		}}
	})

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})

	if lkiName == "" {
		t.Fatal("the LTB trigger never fired, so nothing was read")
	}
	if lkiPower != 1 || lkiToughness != 4 {
		t.Errorf("LKI P/T = %d/%d, want 1/4 — the snapshot must precede the exit cleanup",
			lkiPower, lkiToughness)
	}
	if sourceDamage != 0 {
		t.Errorf("the exiled object carries %d damage, want 0 (CR 400.7)", sourceDamage)
	}
}
