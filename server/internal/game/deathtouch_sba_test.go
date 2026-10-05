package game

import "testing"

// deathtouch_sba_test.go — #2319. CR 704.5h destroys a creature dealt
// deathtouch damage "since the last time state-based actions were
// checked", so the mark is consumed by the pass that reads it.

func pushDeathtouchSource(t *testing.T, g *Game) *Card {
	t.Helper()
	id := pushKeywordCreature(t, g, g.Seats[0], 1, 1, "deathtouch")
	return findCard(g, id)
}

func TestDeathtouchDamageDestroysNonIndestructible(t *testing.T) {
	g := newActiveGame(t)
	src := pushDeathtouchSource(t, g)
	victim := pushVanillaGolem(g, g.Seats[1], "Bear")
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(src.InstanceID, victim, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
		g.runStateChecksLocked()
	})
	if g.Battlefield.Contains(victim) {
		t.Error("creature dealt deathtouch damage survived the SBA check")
	}
}

// The #2319 case: indestructible for the check that follows the
// damage, then loses it. The damage was already counted, so it lives.
func TestDeathtouchMarkNotRecountedAfterIndestructibleEnds(t *testing.T) {
	g := newActiveGame(t)
	src := pushDeathtouchSource(t, g)
	victim := pushVanillaGolem(g, g.Seats[1], "Cardboard Golem")
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(src.InstanceID, g.PinnedObjectsLocked(victim),
			[]Mod{AddKeywordsMod("indestructible")}, g.UntilEndOfTurnDuration(),
			"test — indestructible until end of turn")
		if err := g.DealDamageToCreatureForEffect(src.InstanceID, victim, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
		g.runStateChecksLocked()
	})
	if !g.Battlefield.Contains(victim) {
		t.Fatal("indestructible creature died to deathtouch damage")
	}
	c := findBattlefieldCard(g, victim)
	if c.MarkedLethalByDeathtouch {
		t.Error("deathtouch mark outlived the SBA check")
	}
	if c.DamageMarked != 1 {
		t.Errorf("DamageMarked = %d, want 1 (CR 704.5g keeps counting it)", c.DamageMarked)
	}
	g.WithWriteLock(func() {
		g.Turn.Seq++
		g.ClearEndOfTurnScopedStaticsLocked()
		g.runStateChecksLocked()
	})
	if !g.Battlefield.Contains(victim) {
		t.Error("creature destroyed by stale deathtouch damage after losing indestructible")
	}
}

// Indestructible granted in response, before any SBA check, still
// saves the creature at the one check that counts the damage.
func TestDeathtouchDamageThenIndestructibleBeforeCheckSurvives(t *testing.T) {
	g := newActiveGame(t)
	src := pushDeathtouchSource(t, g)
	victim := pushVanillaGolem(g, g.Seats[1], "Cardboard Golem")
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(src.InstanceID, victim, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
		g.RegisterScopedEffectForEffect(src.InstanceID, g.PinnedObjectsLocked(victim),
			[]Mod{AddKeywordsMod("indestructible")}, g.UntilEndOfTurnDuration(),
			"test — indestructible in response")
		g.runStateChecksLocked()
	})
	if !g.Battlefield.Contains(victim) {
		t.Error("creature that gained indestructible before the check was destroyed")
	}
}

// A first-strike deathtouch attacker kills its blocker in the
// first-strike step, before regular damage.
func TestFirstStrikeDeathtouchKillsBlockerAndSurvives(t *testing.T) {
	g := newActiveGame(t)
	atk := pushKeywordCreature(t, g, g.Seats[0], 1, 1, "deathtouch", "first strike")
	ogre := pushKeywordCreature(t, g, g.Seats[1], 5, 5)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(ogre, atk); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	if findCard(g, ogre) != nil {
		t.Error("blocker should die to first-strike deathtouch damage")
	}
	if findCard(g, atk) == nil {
		t.Error("first-strike attacker should survive: the blocker died before dealing regular damage")
	}
}
