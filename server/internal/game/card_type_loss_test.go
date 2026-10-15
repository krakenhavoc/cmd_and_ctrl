package game

import (
	"reflect"
	"testing"
)

// card_type_loss_test.go — #2041: losing a card type loses the subtypes
// that went with it (CR 205.1a), applied by the layer pass to every
// layer-4 effect.

// A creature that stops being a creature loses its creature types, and
// HasSubtype and CreatureTypesOf say so. When the removal ends they are
// back.
func TestLosingCreatureLosesCreatureTypes(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Artifact Creature — Alien Angel")
	registerScopedEffectForTest(t, g, id, []Mod{RemoveTypesMod("Creature")}, IndefiniteDuration())
	c := scopedEffectChar(t, g, id)
	if !reflect.DeepEqual(c.Types, []string{"Artifact"}) || len(c.Subtypes) != 0 {
		t.Fatalf("types %v subtypes %v, want [Artifact] and none (CR 205.1a)", c.Types, c.Subtypes)
	}
	var angel bool
	var creatureTypes []string
	g.WithWriteLock(func() {
		if card, ok := g.battlefieldCardLocked(id); ok {
			angel, creatureTypes = card.HasSubtype("Angel"), CreatureTypesOf(card)
		}
	})
	if angel || len(creatureTypes) != 0 {
		t.Errorf("HasSubtype(Angel) = %v, CreatureTypesOf = %v; want false and none", angel, creatureTypes)
	}

	g.WithWriteLock(func() {
		g.ScopedEffects = nil
		g.layerVersion.Add(1)
	})
	if c := scopedEffectChar(t, g, id); !reflect.DeepEqual(c.Subtypes, []string{"Alien", "Angel"}) {
		t.Errorf("after the removal ends: subtypes %v, want [Alien Angel]", c.Subtypes)
	}
}

// An artifact creature that loses Creature keeps its artifact types; a
// land creature keeps its land types.
func TestLosingCreatureKeepsOtherTypesSubtypes(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     []string
	}{
		{"Artifact Creature — Equipment Lizard", []string{"Equipment"}},
		{"Land Creature — Forest Dryad", []string{"Forest"}},
		{"Enchantment Creature — Aura Spirit", []string{"Aura"}},
	} {
		g := newActiveGame(t)
		id := pushTypedTestPermanent(g, g.Seats[0].ID, tc.typeLine)
		registerScopedEffectForTest(t, g, id, []Mod{RemoveTypesMod("Creature")}, IndefiniteDuration())
		if c := scopedEffectChar(t, g, id); !reflect.DeepEqual(c.Subtypes, tc.want) {
			t.Errorf("%s: subtypes %v, want %v", tc.typeLine, c.Subtypes, tc.want)
		}
	}
}

// Kindred shares the creature types (CR 205.3m), so a Kindred object
// that stops being a creature keeps them, and keeps a changeling's
// every type.
func TestKindredKeepsCreatureTypesWhenCreatureGoes(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Kindred Artifact Creature — Goblin Equipment", KeywordChangeling)
	registerScopedEffectForTest(t, g, id, []Mod{RemoveTypesMod("Creature")}, IndefiniteDuration())
	c := scopedEffectChar(t, g, id)
	if !reflect.DeepEqual(c.Subtypes, []string{"Goblin", "Equipment"}) || !c.AllCreatureTypes {
		t.Errorf("subtypes %v (every type %v), want [Goblin Equipment] and every creature type", c.Subtypes, c.AllCreatureTypes)
	}
}

// A changeling that stops being a creature is no longer every creature
// type.
func TestLosingCreatureEndsAChangelingsEveryType(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Shapeshifter", KeywordChangeling)
	registerScopedEffectForTest(t, g, id, []Mod{RemoveTypesMod("Creature")}, IndefiniteDuration())
	if c := scopedEffectChar(t, g, id); c.AllCreatureTypes || len(c.Subtypes) != 0 {
		t.Errorf("subtypes %v (every type %v), want none", c.Subtypes, c.AllCreatureTypes)
	}
}

// Timestamps decide (CR 613.7). "Isn't a creature" first and a later
// "is an artifact creature": a creature again, without the creature
// types the removal took, which it removes "for the entire time the
// object's card type is removed". The other order: not a creature, and
// the animation's own creature type goes too.
func TestCardTypeLossFollowsTimestampOrder(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Artifact Creature — Alien Angel")
	registerScopedEffectForTest(t, g, id, []Mod{RemoveTypesMod("Creature")}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, id, []Mod{AddTypesMod("Creature")}, IndefiniteDuration())
	c := scopedEffectChar(t, g, id)
	if !typeListHas(c.Types, "Creature") || len(c.Subtypes) != 0 {
		t.Errorf("removal then animation: types %v subtypes %v, want a creature with no creature types", c.Types, c.Subtypes)
	}

	g = newActiveGame(t)
	id = pushTypedTestPermanent(g, g.Seats[0].ID, "Artifact — Equipment")
	registerScopedEffectForTest(t, g, id, []Mod{AddTypesMod("Creature"), AddSubtypesMod("Angel")}, IndefiniteDuration())
	if c := scopedEffectChar(t, g, id); !reflect.DeepEqual(c.Subtypes, []string{"Equipment", "Angel"}) {
		t.Fatalf("control: animated subtypes %v, want [Equipment Angel]", c.Subtypes)
	}
	registerScopedEffectForTest(t, g, id, []Mod{RemoveTypesMod("Creature")}, IndefiniteDuration())
	c = scopedEffectChar(t, g, id)
	if typeListHas(c.Types, "Creature") || !reflect.DeepEqual(c.Subtypes, []string{"Equipment"}) {
		t.Errorf("animation then removal: types %v subtypes %v, want a noncreature Equipment", c.Types, c.Subtypes)
	}
}

// The rule is every list's, not the creature list's: a planeswalker
// that stops being one loses its planeswalker type, a land its land
// types, an artifact its artifact types.
func TestLosingAnyCardTypeLosesItsSubtypes(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		remove   string
		want     []string
	}{
		{"Legendary Planeswalker — Gideon", "Planeswalker", nil},
		{"Artifact Land — Forest", "Land", nil},
		{"Artifact Creature — Equipment Lizard", "Artifact", []string{"Lizard"}},
		{"Battle — Siege", "Battle", nil},
	} {
		g := newActiveGame(t)
		id := pushTypedTestPermanent(g, g.Seats[0].ID, tc.typeLine)
		registerScopedEffectForTest(t, g, id, []Mod{RemoveTypesMod(tc.remove)}, IndefiniteDuration())
		c := scopedEffectChar(t, g, id)
		if len(c.Subtypes) != len(tc.want) || (len(tc.want) > 0 && !reflect.DeepEqual(c.Subtypes, tc.want)) {
			t.Errorf("%s without %s: subtypes %v, want %v", tc.typeLine, tc.remove, c.Subtypes, tc.want)
		}
	}
}

// dropSubtypesOfLostTypes never writes through the slice it was handed:
// the layer pass's baseline is cached.
func TestDropSubtypesOfLostTypesWritesANewSlice(t *testing.T) {
	shared := []string{"Alien", "Equipment", "Angel"}
	c := Characteristic{Types: []string{"Artifact"}, Subtypes: shared}
	c.dropSubtypesOfLostTypes(correlatedSubtypeSets([]string{"Artifact", "Creature"}))
	if !reflect.DeepEqual(c.Subtypes, []string{"Equipment"}) {
		t.Errorf("subtypes %v, want [Equipment]", c.Subtypes)
	}
	if !reflect.DeepEqual(shared, []string{"Alien", "Equipment", "Angel"}) {
		t.Errorf("the original slice was written: %v", shared)
	}
}
