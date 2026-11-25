package game

import (
	"testing"

	"github.com/google/uuid"
)

// single_source_lethal_test.go — #6209: "This creature can't be
// destroyed by lethal damage unless lethal damage dealt by a single
// source is marked on it" (Ogre Enforcer).
//
// CR 704.5g: "If a creature has toughness greater than 0, it has damage
// marked on it, and the total damage marked on it is greater than or
// equal to its toughness, that creature has been dealt lethal damage and
// is destroyed." The clause asks who marked it, so the engine keeps the
// marked damage per source.

const ogreEnforcerOracle = "test-ogre-enforcer"

// withSingleSourceLethalOnly installs the static for one test, the way
// withCatalogAttackTaxes installs a tax.
func withSingleSourceLethalOnly(t *testing.T, oracle string) {
	t.Helper()
	prev := CatalogSingleSourceLethalOnly
	t.Cleanup(func() { CatalogSingleSourceLethalOnly = prev })
	CatalogSingleSourceLethalOnly = func(key string) []SingleSourceLethalOnly {
		if key == oracle {
			return []SingleSourceLethalOnly{{Label: "This creature can't be destroyed by lethal damage unless lethal damage dealt by a single source is marked on it."}}
		}
		return nil
	}
}

// pushEnforcer puts a 4-toughness creature carrying the static on the
// battlefield.
func pushEnforcer(g *Game, owner uuid.UUID) uuid.UUID {
	id := pushScopedTestCreature(g, owner, 3, 4)
	g.WithWriteLock(func() {
		if c := findBattlefieldCard(g, id); c != nil {
			c.Name = "Ogre Enforcer"
			c.OracleID = ogreEnforcerOracle
		}
	})
	return id
}

func alive(g *Game, id uuid.UUID) bool {
	var ok bool
	g.ReadSnapshot(func() { ok = findBattlefieldCard(g, id) != nil })
	return ok
}

// TestEnforcerSurvivesTwoSourcesAndDiesToOne — the headline: 2+2 from
// two sources leaves it standing, 4 from one destroys it.
func TestEnforcerSurvivesTwoSourcesAndDiesToOne(t *testing.T) {
	withSingleSourceLethalOnly(t, ogreEnforcerOracle)
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	enforcer := pushEnforcer(g, opp)
	one := pushScopedTestCreature(g, me, 2, 2)
	two := pushScopedTestCreature(g, me, 2, 2)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(one, enforcer, 2); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(two, enforcer, 2); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if !alive(g, enforcer) {
		t.Fatal("2 + 2 from two sources destroyed it")
	}
	var marked int
	g.ReadSnapshot(func() { marked = findBattlefieldCard(g, enforcer).DamageMarked })
	if marked != 4 {
		t.Errorf("damage marked = %d, want 4 — the damage is marked either way (CR 704.5g)", marked)
	}

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(one, enforcer, 2); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if alive(g, enforcer) {
		t.Error("4 from one source left it standing")
	}
}

// TestEnforcerDiesToOneBigSource — a single source over the threshold in
// one hit destroys it.
func TestEnforcerDiesToOneBigSource(t *testing.T) {
	withSingleSourceLethalOnly(t, ogreEnforcerOracle)
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	enforcer := pushEnforcer(g, opp)
	big := pushScopedTestCreature(g, me, 5, 5)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(big, enforcer, 5); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if alive(g, enforcer) {
		t.Error("5 from one source left it standing")
	}
}

// TestEnforcerClauseIsTheCreaturesOwn — a creature without the static is
// destroyed by the plain CR 704.5g sum, from any number of sources.
func TestEnforcerClauseIsTheCreaturesOwn(t *testing.T) {
	withSingleSourceLethalOnly(t, ogreEnforcerOracle)
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	plain := pushScopedTestCreature(g, opp, 3, 4)
	one := pushScopedTestCreature(g, me, 2, 2)
	two := pushScopedTestCreature(g, me, 2, 2)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(one, plain, 2); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(two, plain, 2); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if alive(g, plain) {
		t.Error("an ordinary creature survived 2 + 2 from two sources")
	}
}

// TestEnforcerSplitShrinksWithTheMark — removing damage removes it from
// the per-source split too, so a creature healed below lethal is not
// destroyed by the damage that was there before.
func TestEnforcerSplitShrinksWithTheMark(t *testing.T) {
	withSingleSourceLethalOnly(t, ogreEnforcerOracle)
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	enforcer := pushEnforcer(g, opp)
	one := pushScopedTestCreature(g, me, 2, 2)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(one, enforcer, 3); err != nil {
			t.Fatal(err)
		}
	})
	var split []SourceMark
	g.ReadSnapshot(func() { split = findBattlefieldCard(g, enforcer).MarkedBySource })
	if len(split) != 1 || split[0].Source != one || split[0].Amount != 3 {
		t.Fatalf("split = %+v, want 3 from the one source", split)
	}
	g.WithWriteLock(func() { g.HealAllDamageForEffect(enforcer) })
	g.ReadSnapshot(func() {
		c := findBattlefieldCard(g, enforcer)
		if c.DamageMarked != 0 || len(c.MarkedBySource) != 0 {
			t.Errorf("after removing all damage: marked = %d, split = %+v", c.DamageMarked, c.MarkedBySource)
		}
	})
	// And it takes four from one source again to destroy it: the healed
	// damage counts for nothing.
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(one, enforcer, 3); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if !alive(g, enforcer) {
		t.Error("3 after a heal destroyed it")
	}
}

// TestLethalFromASingleSourceReadsTheThreshold — the predicate itself:
// false with nothing marked, false when two sources share the total,
// true once one of them is at the threshold.
func TestLethalFromASingleSourceReadsTheThreshold(t *testing.T) {
	c := &Card{TypeLine: "Creature — Ogre", Power: 3, Toughness: 4, PrintedPTKnown: true}
	if LethalFromASingleSource(c) {
		t.Error("nothing marked reads as lethal from one source")
	}
	c.DamageMarked = 4
	c.noteMarkedDamageSource(uuid.New(), 2)
	c.noteMarkedDamageSource(uuid.New(), 2)
	if LethalFromASingleSource(c) {
		t.Error("2 + 2 from two sources reads as lethal from one")
	}
	c.DamageMarked = 8
	c.noteMarkedDamageSource(c.MarkedBySource[0].Source, 2)
	if !LethalFromASingleSource(c) {
		t.Error("4 from one source does not read as lethal from one")
	}
	c.releaseMarkedDamage(8)
	if len(c.MarkedBySource) != 0 {
		t.Errorf("releasing everything left %+v", c.MarkedBySource)
	}
}
