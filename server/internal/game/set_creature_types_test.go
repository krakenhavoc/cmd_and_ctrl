package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// set_creature_types_test.go — #2144, ModSetCreatureTypes.

// pushTypedTestPermanent puts a permanent with the given type line onto
// the battlefield, with an entry event so a record can pin it.
func pushTypedTestPermanent(g *Game, owner uuid.UUID, typeLine string, keywords ...string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: "Typed Thing", TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		Keywords: keywords,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: id, OldZone: ZoneHand, NewZone: ZoneBattlefield})
	})
	return id
}

// An artifact creature keeps its card types and its artifact subtype;
// only its creature types are replaced (CR 205.1a).
func TestSetCreatureTypesKeepsCardTypesAndOtherSubtypes(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Legendary Artifact Creature — Equipment Elf Warrior")
	registerScopedEffectForTest(t, g, id, []Mod{SetCreatureTypesMod("Frog")}, IndefiniteDuration())
	c := scopedEffectChar(t, g, id)
	if !reflect.DeepEqual(c.Types, []string{"Artifact", "Creature"}) {
		t.Errorf("types = %v, want [Artifact Creature]", c.Types)
	}
	if !reflect.DeepEqual(c.Subtypes, []string{"Equipment", "Frog"}) {
		t.Errorf("subtypes = %v, want [Equipment Frog]", c.Subtypes)
	}
	if !typeListHas(c.Supertypes, "Legendary") {
		t.Errorf("supertypes = %v, want Legendary kept", c.Supertypes)
	}
}

// A changeling is every creature type until the set, and only the new
// type after it.
func TestSetCreatureTypesEndsAChangelingsEveryType(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Shapeshifter", KeywordChangeling)
	if c := scopedEffectChar(t, g, id); !c.AllCreatureTypes {
		t.Fatal("control: the changeling is not every creature type")
	}
	registerScopedEffectForTest(t, g, id, []Mod{SetCreatureTypesMod("Halfling", "Scout")}, IndefiniteDuration())
	c := scopedEffectChar(t, g, id)
	if c.AllCreatureTypes || !reflect.DeepEqual(c.Subtypes, []string{"Halfling", "Scout"}) {
		t.Errorf("subtypes = %v (all %v), want only [Halfling Scout]", c.Subtypes, c.AllCreatureTypes)
	}
	var elf, scout bool
	g.WithWriteLock(func() {
		if card, ok := g.battlefieldCardLocked(id); ok {
			elf, scout = card.HasSubtype("Elf"), card.HasSubtype("Scout")
		}
	})
	if elf || !scout {
		t.Error("HasSubtype still answers for the old types, or not for the new one")
	}
}

// A noncreature permanent can't have a creature type (CR 205.3d).
func TestSetCreatureTypesGivesANoncreatureNone(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Artifact — Vehicle")
	registerScopedEffectForTest(t, g, id, []Mod{SetCreatureTypesMod("Frog")}, IndefiniteDuration())
	if c := scopedEffectChar(t, g, id); !reflect.DeepEqual(c.Subtypes, []string{"Vehicle"}) {
		t.Errorf("subtypes = %v, want [Vehicle]", c.Subtypes)
	}
}

func TestSetCreatureTypesModProblem(t *testing.T) {
	for _, tc := range []struct {
		mod  Mod
		want bool
	}{
		{SetCreatureTypesMod("Frog"), false},
		{SetCreatureTypesMod("Halfling", "Rogue"), false},
		{SetCreatureTypesMod(), true},
		{SetCreatureTypesMod("Forest"), true},
		{AddSubtypesMod("Forest"), false},
	} {
		if got := setCreatureTypesModProblem(tc.mod) != ""; got != tc.want {
			t.Errorf("%v %v: problem = %v, want %v", tc.mod.Kind, tc.mod.Subtypes, got, tc.want)
		}
	}
}
