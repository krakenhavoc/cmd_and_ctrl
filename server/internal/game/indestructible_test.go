package game

import (
	"testing"

	"github.com/google/uuid"
)

// indestructible_test.go — S25 (#77). The point of these cases is
// not that indestructible works; it is that indestructible works
// ONLY where CR 702.12b says it does. Half the file is negative
// space: the zero-toughness SBA, the zero-loyalty SBA and sacrifice
// all still kill an indestructible permanent, and each of those is
// a rule players routinely get wrong, so each gets its own test
// rather than a shared table.

// pushIndestructibleCreature puts a 2/2 with printed indestructible
// onto the battlefield under `owner`. Uses the printed `Keywords`
// road rather than the catalog so the fixture needs no oracle ID;
// printedCharacteristic merges Card.Keywords into the base
// characteristic, so this survives a layer recompute.
func pushIndestructibleCreature(g *Game, owner *Player, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Golem"
	c.ManaCost = "{4}"
	c.Power, c.Toughness = 2, 2
	c.Keywords = []string{"indestructible"}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// pushVanillaGolem is the control: same body, no keyword.
func pushVanillaGolem(g *Game, owner *Player, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Golem"
	c.ManaCost = "{4}"
	c.Power, c.Toughness = 2, 2
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// TestIndestructibleSurvivesDestroyEffect is CR 701.7 + CR 702.12b at
// the catalog's destruction verb: every `DestroyTarget` in the
// effects package and every wrath that loops over the battlefield
// reaches `DestroyPermanentForEffect`, so gating there covers all of
// them at once.
func TestIndestructibleSurvivesDestroyEffect(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	safe := pushIndestructibleCreature(g, owner, "Darksteel Golem")
	doomed := pushVanillaGolem(g, owner, "Cardboard Golem")

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(safe); err != nil {
			t.Errorf("destroying an indestructible permanent: got %v, want nil", err)
		}
		if err := g.DestroyPermanentForEffect(doomed); err != nil {
			t.Errorf("destroying a plain permanent: got %v, want nil", err)
		}
	})

	if !g.Battlefield.Contains(safe) {
		t.Error("indestructible creature was destroyed by a destroy effect")
	}
	if g.Battlefield.Contains(doomed) {
		t.Error("plain creature survived a destroy effect — control case broken")
	}
}

// TestIndestructibleSurvivesLethalDamageAndKeepsIt covers CR 704.5g
// and the half of CR 702.12b that is easy to miss: the damage is
// still DEALT and still MARKED, the SBA just declines to act on it.
//
// The second half of the test is the reason that matters. Strip the
// keyword mid-turn — a control-change, a text-change, or in practice
// a turn-scoped grant that ended — and the creature dies immediately
// to damage it accumulated while protected. An implementation that
// "absorbed" the damage instead would leave it alive at full health.
func TestIndestructibleSurvivesLethalDamageAndKeepsIt(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushIndestructibleCreature(g, owner, "Darksteel Golem")

	if err := g.MarkDamage(id, 5); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })

	if !g.Battlefield.Contains(id) {
		t.Fatal("indestructible creature died to lethal marked damage")
	}
	card := findBattlefieldCard(g, id)
	if card.DamageMarked != 5 {
		t.Errorf("DamageMarked = %d, want 5 — damage must still be marked (CR 702.12b)", card.DamageMarked)
	}

	// Lose the keyword; the marked damage is still there and is
	// still lethal.
	g.WithWriteLock(func() {
		findBattlefieldCard(g, id).Keywords = nil
		g.layerVersion.Add(1)
		g.runStateChecksLocked()
	})
	if g.Battlefield.Contains(id) {
		t.Error("creature survived after losing indestructible with 5 damage marked on a 2-toughness body")
	}
}

// TestIndestructibleSurvivesDeathtouch is CR 704.5h. Deathtouch's
// whole trick is bypassing the toughness comparison, so it needs its
// own gate — filtering only on `DamageMarked >= toughness` would let
// a 1-damage deathtouch hit kill an indestructible creature.
//
// The flag is deliberately still SET on the card: it records a fact
// about the damage event, not about the creature, for the same
// reason the damage itself stays marked above.
func TestIndestructibleSurvivesDeathtouch(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushIndestructibleCreature(g, owner, "Darksteel Golem")

	g.WithWriteLock(func() {
		findBattlefieldCard(g, id).MarkedLethalByDeathtouch = true
		g.runStateChecksLocked()
	})

	if !g.Battlefield.Contains(id) {
		t.Fatal("indestructible creature died to a deathtouch mark")
	}
	if !findBattlefieldCard(g, id).MarkedLethalByDeathtouch {
		t.Error("MarkedLethalByDeathtouch was cleared — the flag records the damage event, not the creature's state")
	}
}

// TestIndestructibleDiesToZeroToughness is CR 704.5f, the rule this
// implementation most had to be careful NOT to break. That SBA puts
// the creature into its owner's graveyard; it does not destroy it,
// so indestructible is irrelevant and a -1/-1 counter pile still
// works. Putting the check one level down in
// routeBattlefieldCardToOwnerGraveyardLocked would have broken this.
func TestIndestructibleDiesToZeroToughness(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushIndestructibleCreature(g, owner, "Darksteel Golem")

	g.WithWriteLock(func() {
		findBattlefieldCard(g, id).Counters = map[string]int{"-1/-1": 2}
		g.layerVersion.Add(1)
		g.runStateChecksLocked()
	})

	if g.Battlefield.Contains(id) {
		t.Error("indestructible creature at 0 toughness survived — CR 704.5f is not destruction")
	}
}

// TestIndestructibleDoesNotStopSacrifice is CR 701.17b. Sacrifice
// and destruction share an exit ramp
// (routeBattlefieldCardToOwnerGraveyardLocked) and differ in every
// other respect; this test is what pins the gate above that ramp
// rather than inside it.
func TestIndestructibleDoesNotStopSacrifice(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushIndestructibleCreature(g, owner, "Darksteel Golem")

	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(id); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})

	if g.Battlefield.Contains(id) {
		t.Error("indestructible permanent survived being sacrificed — sacrifice is not destruction")
	}
}

// TestIndestructiblePlaneswalkerStillDiesAtZeroLoyalty is CR 704.5i,
// the other "put into its owner's graveyard" SBA. The paper ruling
// people are always surprised by: an indestructible planeswalker
// that spends itself down to 0 loyalty still dies.
func TestIndestructiblePlaneswalkerStillDiesAtZeroLoyalty(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	c := NewCard("Indestructible Walker", owner.ID)
	c.TypeLine = "Legendary Planeswalker — Test"
	c.Keywords = []string{"indestructible"}
	c.Counters = map[string]int{CounterLoyalty: 0}
	g.Battlefield.PushTop(c)

	g.WithWriteLock(func() { g.runStateChecksLocked() })

	if g.Battlefield.Contains(c.InstanceID) {
		t.Error("indestructible planeswalker at 0 loyalty survived — CR 704.5i is not destruction")
	}
}

// TestTurnScopedIndestructibleGrantProtectsUntilItExpires is the
// Heroic Intervention / Boros Charm shape end to end: a layer-6
// grant registered in the turn-scoped registry really does reach
// HasKeyword through the recompute, and really does stop mattering
// once the cleanup sweep drops it.
//
// It also pins the recompute call inside
// destroyBattlefieldPermanentLocked. A grant registered moments
// earlier in the same resolution frame has bumped layerVersion but
// not refreshed `effective`; without that call, a Wrath resolving
// after a Heroic Intervention would kill everything.
func TestTurnScopedIndestructibleGrantProtectsUntilItExpires(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushVanillaGolem(g, owner, "Cardboard Golem")

	g.WithWriteLock(func() {
		g.RegisterTurnScopedStaticForEffect(StaticAbility{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
				return target.InstanceID == id
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Abilities = append(c.Abilities, "indestructible")
			},
		}, uuid.New(), "test — indestructible until end of turn")

		// No explicit recompute here on purpose: the destroy path
		// owes us one.
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if !g.Battlefield.Contains(id) {
		t.Fatal("creature with a turn-scoped indestructible grant was destroyed")
	}

	// Cleanup sweeps the grant (CR 514.2); the creature is plain
	// cardboard again.
	g.WithWriteLock(func() {
		g.Turn.Number++
		g.ClearExpiredTurnScopedStaticsLocked()
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect after expiry: %v", err)
		}
	})
	if g.Battlefield.Contains(id) {
		t.Error("creature survived destruction after its indestructible grant expired")
	}
}

// TestIndestructibleIsACanonicalKeyword guards the invariant stated
// on `canonicalKeywords`: a token is in that table exactly when the
// engine honours it. The deck importer filters Scryfall's keyword
// array through CanonicalKeyword, so failing this would silently
// strip indestructible off every imported Avacyn and Darksteel
// Citadel and none of the tests above would notice.
func TestIndestructibleIsACanonicalKeyword(t *testing.T) {
	if kw, ok := CanonicalKeyword("Indestructible"); !ok || kw != "indestructible" {
		t.Errorf("CanonicalKeyword(\"Indestructible\") = (%q, %v), want (\"indestructible\", true)", kw, ok)
	}
}
