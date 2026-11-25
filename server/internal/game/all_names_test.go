package game

import (
	"testing"

	"github.com/google/uuid"
)

// all_names_test.go — #6257: Spy Kit's "Equipped creature ... has all
// names of nonlegendary creature cards in addition to its name". CR
// 201.2a: two objects have the same name if they have at least one name
// in common, even if one of them has additional names.

const testAllNamesOracle = "test-all-nonlegendary-creature-names"

// withAllNames wires a card database that knows two nonlegendary
// creature cards, and a static giving the test card all their names in
// layer 3 (CR 612), as Spy Kit gives them to the equipped creature.
func withAllNames(t *testing.T) {
	t.Helper()
	prevNames := NonlegendaryCreatureCardName
	t.Cleanup(func() { NonlegendaryCreatureCardName = prevNames })
	NonlegendaryCreatureCardName = func(name string) bool {
		return name == "Grizzly Bears" || name == "Llanowar Elves"
	}
	prev := CatalogStaticAbilities
	t.Cleanup(func() { CatalogStaticAbilities = prev })
	CatalogStaticAbilities = func(key string) []StaticAbility {
		if key == testAllNamesOracle {
			return []StaticAbility{{
				Layer:     Layer3Text,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID },
				Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.AllNonlegendaryCreatureNames = true },
			}}
		}
		if prev == nil {
			return nil
		}
		return prev(key)
	}
}

// pushNamed puts a creature named `name` onto the battlefield under
// `controller`, with all those names when `allNames`. Caller holds g.mu.
func pushNamed(g *Game, controller uuid.UUID, name, typeLine string, allNames bool) uuid.UUID {
	c := NewCard(name, controller)
	c.TypeLine = typeLine
	c.Power, c.Toughness = 2, 2
	if allNames {
		c.OracleID = testAllNamesOracle
	}
	g.Battlefield.PushTop(c)
	g.layerVersion.Add(1)
	g.RecomputeLayersIfStaleLocked()
	return c.InstanceID
}

// TestAllNamesMatchEveryNonlegendaryCreatureCardName: the creature has
// its own name and every name the card database calls a nonlegendary
// creature card's, and no other; one without the effect has only its own.
func TestAllNamesMatchEveryNonlegendaryCreatureCardName(t *testing.T) {
	withAllNames(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	var spy, plain uuid.UUID
	g.WithWriteLock(func() {
		spy = pushNamed(g, me.ID, "Shadow Spy", "Creature — Human Rogue", true)
		plain = pushNamed(g, me.ID, "Plain Rogue", "Creature — Human Rogue", false)
	})
	var s, p Card
	g.ReadSnapshot(func() {
		s, p = *findBattlefieldCard(g, spy), *findBattlefieldCard(g, plain)
	})
	for _, name := range []string{"Shadow Spy", "Grizzly Bears", "Llanowar Elves"} {
		if !CardNameMatches(s, name) || !PermanentHasName(s, name) {
			t.Errorf("the creature with all names does not have the name %q", name)
		}
	}
	if CardNameMatches(s, "Krenko, Mob Boss") || PermanentHasName(s, "Krenko, Mob Boss") {
		t.Error("it has a name no nonlegendary creature card has")
	}
	if CardNameMatches(p, "Grizzly Bears") || PermanentHasName(p, "Grizzly Bears") {
		t.Error("a creature without the effect has Grizzly Bears' name")
	}
	if !HasAllNonlegendaryCreatureNames(s) || HasAllNonlegendaryCreatureNames(p) {
		t.Error("HasAllNonlegendaryCreatureNames disagrees with the effect")
	}
}

// TestAllNamesWithoutACardDatabaseIsOnlyItsOwnName: with no card
// database wired, the engine knows no names to add.
func TestAllNamesWithoutACardDatabaseIsOnlyItsOwnName(t *testing.T) {
	withAllNames(t)
	NonlegendaryCreatureCardName = nil
	g := newActiveGame(t)
	var spy uuid.UUID
	g.WithWriteLock(func() { spy = pushNamed(g, g.Seats[0].ID, "Shadow Spy", "Creature — Human Rogue", true) })
	var s Card
	g.ReadSnapshot(func() { s = *findBattlefieldCard(g, spy) })
	if CardNameMatches(s, "Grizzly Bears") {
		t.Error("matched a name with no card database")
	}
	if !CardNameMatches(s, "Shadow Spy") {
		t.Error("lost its own name")
	}
}

// TestAllNamesLegendRule: two legendary creatures one player controls
// that both have all those names share a name (CR 201.2a), so the legend
// rule applies to them; one alone, or one beside a legend with an
// unrelated name, is left alone.
func TestAllNamesLegendRule(t *testing.T) {
	withAllNames(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	var a, b uuid.UUID
	g.WithWriteLock(func() {
		a = pushNamed(g, me.ID, "Legend One", "Legendary Creature — Human", true)
		pushNamed(g, me.ID, "Legend Two", "Legendary Creature — Human", false)
		if got := g.legendRuleChoicesLocked(); len(got) != 0 {
			t.Errorf("one legend with all names beside an unrelated legend: violations %v, want none", got)
		}
		b = pushNamed(g, me.ID, "Legend Three", "Legendary Creature — Human", true)
		got := g.legendRuleChoicesLocked()[me.ID]
		if len(got) != 2 || !containsID(got, a) || !containsID(got, b) {
			t.Errorf("two legends with all names: violation %v, want both", got)
		}
	})
}
