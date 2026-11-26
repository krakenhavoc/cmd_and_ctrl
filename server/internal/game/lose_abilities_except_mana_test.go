package game

import "testing"

// lose_abilities_except_mana_test.go — #6246 (CR 613.1f, CR 605.1a): a
// removal that keeps mana abilities composes with one that keeps nothing,
// in either order.

const (
	exceptManaLandOracle = "test-6246-mana-land"
	exceptManaSunOracle  = "test-6246-sun"
	exceptManaAllOracle  = "test-6246-remove-all"
)

func withExceptManaCatalog(t *testing.T) {
	t.Helper()
	prevMana := CatalogManaAbilities
	CatalogManaAbilities = func(k string) []ManaAbilityShape {
		if k == exceptManaLandOracle {
			return []ManaAbilityShape{{TapCost: true, Produced: "{C}", Label: "Add {C}"}}
		}
		if prevMana != nil {
			return prevMana(k)
		}
		return nil
	}
	t.Cleanup(func() { CatalogManaAbilities = prevMana })
	lands := func(target *Card, _ *Game, _ *Card) bool { return target.IsLand() }
	withStaticAbilities(t, func(key string) []StaticAbility {
		switch key {
		case exceptManaSunOracle:
			return []StaticAbility{{Layer: Layer6Ability, RemovesAbilities: true, KeepsManaAbilities: true, AppliesTo: lands,
				Apply: func(*Characteristic, *Card, *Game, *Card) {}}}
		case exceptManaAllOracle:
			return []StaticAbility{{Layer: Layer6Ability, RemovesAbilities: true, AppliesTo: lands,
				Apply: func(*Characteristic, *Card, *Game, *Card) {}}}
		}
		return nil
	})
}

func exceptManaRows(g *Game, c Card) int {
	n := 0
	g.WithWriteLock(func() {
		g.layerVersion.Add(1)
		g.RecomputeLayersIfStaleLocked()
		if live := findBattlefieldCard(g, c.InstanceID); live != nil {
			n = len(manaAbilitiesOf(live))
		}
	})
	return n
}

func TestKeepManaRemovalComposesWithAFullRemoval(t *testing.T) {
	for _, order := range [][]string{
		{exceptManaSunOracle, exceptManaAllOracle},
		{exceptManaAllOracle, exceptManaSunOracle},
	} {
		withExceptManaCatalog(t)
		g := newActiveGame(t)
		me := g.Seats[0]
		land := Card{Name: "Mana Land", TypeLine: "Land", OracleID: exceptManaLandOracle, Owner: me.ID, Controller: me.ID}
		land.InstanceID = pushTypedTestCard(g, land)
		if n := exceptManaRows(g, land); n != 1 {
			t.Fatalf("setup: %d mana rows, want 1", n)
		}
		pushTypedTestCard(g, Card{Name: order[0], TypeLine: "Enchantment", OracleID: order[0], Owner: me.ID, Controller: me.ID})
		want := 1
		if order[0] == exceptManaAllOracle {
			want = 0
		}
		if n := exceptManaRows(g, land); n != want {
			t.Errorf("%v: after the first removal %d mana rows, want %d", order, n, want)
		}
		pushTypedTestCard(g, Card{Name: order[1], TypeLine: "Enchantment", OracleID: order[1], Owner: me.ID, Controller: me.ID})
		if n := exceptManaRows(g, land); n != 0 {
			t.Errorf("%v: with both removals %d mana rows, want 0: a full removal takes the mana ability whatever its order", order, n)
		}
	}
}
