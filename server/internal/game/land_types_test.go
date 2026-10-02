package game

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// land_types_test.go pins ADR 0109 §1 (#1881): CR 205.3i's list, CR
// 305.7's subtype replacement (land types only, CR 205.1a), and the
// setBasicLandTypes mod kind that does it from a resolved effect.

func TestLandTypesIsCR2053isList(t *testing.T) {
	want := []string{"Cave", "Desert", "Forest", "Gate", "Island", "Lair", "Locus", "Mine",
		"Mountain", "Plains", "Planet", "Power-Plant", "Sphere", "Swamp", "Tower", "Town", "Urza's"}
	if !reflect.DeepEqual(LandTypes, want) {
		t.Errorf("LandTypes = %v, want CR 205.3i's %v", LandTypes, want)
	}
	if !reflect.DeepEqual(BasicLandTypes, []string{"Plains", "Island", "Swamp", "Mountain", "Forest"}) {
		t.Errorf("BasicLandTypes = %v", BasicLandTypes)
	}
	for _, b := range BasicLandTypes {
		if !IsLandType(b) || !IsBasicLandType(b) {
			t.Errorf("%s: IsLandType %v, IsBasicLandType %v", b, IsLandType(b), IsBasicLandType(b))
		}
	}
	for _, s := range []string{"Gate", "urza's", "Urza’s", "Power-Plant", "cave"} {
		if !IsLandType(s) {
			t.Errorf("IsLandType(%q) = false", s)
		}
		if IsBasicLandType(s) {
			t.Errorf("IsBasicLandType(%q) = true", s)
		}
	}
	for _, s := range []string{"Dryad", "Elf", "Equipment", "Wastes", ""} {
		if IsLandType(s) {
			t.Errorf("IsLandType(%q) = true", s)
		}
	}
}

// TestSetLandSubtypesReplacesOnlyLandTypes is CR 205.1a's "from the
// appropriate set": a Dryad Arbor set to Island is an Island Dryad.
func TestSetLandSubtypesReplacesOnlyLandTypes(t *testing.T) {
	c := Characteristic{Subtypes: []string{"Forest", "Dryad"}, AllCreatureTypes: true}
	c.SetLandSubtypes([]string{"Island", "Island"})
	if !reflect.DeepEqual(c.Subtypes, []string{"Dryad", "Island"}) {
		t.Errorf("subtypes = %v, want [Dryad Island]", c.Subtypes)
	}
	if !c.AllCreatureTypes {
		t.Error("a land-type set took away \"every creature type\", a creature-type fact")
	}
	g := Characteristic{Subtypes: []string{"Urza's", "Tower"}}
	g.SetLandSubtypes([]string{"Mountain"})
	if !reflect.DeepEqual(g.Subtypes, []string{"Mountain"}) {
		t.Errorf("Urza's Tower set to Mountain: subtypes = %v, want [Mountain]", g.Subtypes)
	}
}

// pushScopedTestLand is a land permanent for these tests.
func pushScopedTestLand(g *Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Keywords:   []string{"hexproof"},
		Owner:      owner,
		Controller: owner,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: id, OldZone: ZoneHand, NewZone: ZoneBattlefield})
	})
	return id
}

// TestSetBasicLandTypesIsCR3057 is the kind's three clauses: the old
// land types go and the other subtypes stay, the rules-text abilities go
// and a granted one survives whenever it was granted, and the land taps
// for the new type's colour and not the old one's.
func TestSetBasicLandTypesIsCR3057(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	arbor := pushScopedTestLand(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad")
	// A grant from an OLDER effect: CR 305.7 "doesn't remove any
	// abilities that were granted to the land by other effects", and a
	// layer-6 grant lands after the layer-4 removal whatever its
	// timestamp.
	registerScopedEffectForTest(t, g, arbor, []Mod{AddKeywordsMod("flying")}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, arbor, []Mod{SetBasicLandTypesMod("island")}, IndefiniteDuration())

	c := scopedEffectChar(t, g, arbor)
	if !reflect.DeepEqual(c.Subtypes, []string{"Dryad", "Island"}) {
		t.Errorf("subtypes = %v, want [Dryad Island] (CR 205.1a)", c.Subtypes)
	}
	if !typeListHas(c.Types, "Land") || !typeListHas(c.Types, "Creature") {
		t.Errorf("types = %v, want Land Creature untouched (CR 305.7)", c.Types)
	}
	if !c.AbilitiesRemoved {
		t.Error("AbilitiesRemoved is not stamped: the rules-text abilities are still there")
	}
	if !reflect.DeepEqual(c.Abilities, []string{"flying"}) {
		t.Errorf("abilities = %v, want only the granted flying (printed hexproof gone)", c.Abilities)
	}
	var land Card
	g.WithWriteLock(func() {
		if c, ok := g.battlefieldCardLocked(arbor); ok {
			land = *c
		}
	})
	var produced []string
	for _, ab := range ManaAbilitiesForCard(land) {
		produced = append(produced, ab.Produced)
	}
	if !reflect.DeepEqual(produced, []string{"{U}"}) {
		t.Errorf("mana abilities produce %v, want only Island's {U} (CR 305.6)", produced)
	}
}

// TestSetBasicLandTypesNamesOnlyBasicLandTypes is the registration
// check: CR 305.7 is about the basic land types.
func TestSetBasicLandTypesNamesOnlyBasicLandTypes(t *testing.T) {
	for _, mods := range [][]Mod{{SetBasicLandTypesMod()}, {SetBasicLandTypesMod("Gate")}} {
		g := newActiveGame(t)
		id := pushScopedTestLand(g, g.Seats[0].ID, "Plains", "Basic Land — Plains")
		func() {
			defer func() {
				if r := recover(); r == nil || !strings.Contains(r.(string), "setBasicLandTypes") {
					t.Errorf("%v: registered without the setBasicLandTypes panic (got %v)", mods[0].Subtypes, r)
				}
			}()
			g.WithWriteLock(func() {
				g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id), mods, IndefiniteDuration(), "test")
			})
		}()
	}
}

// TestSetBasicLandTypesRestoreRefusesANonBasicType is the restore half
// of the same check.
func TestSetBasicLandTypesRestoreRefusesANonBasicType(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestLand(g, g.Seats[0].ID, "Plains", "Basic Land — Plains")
	registerScopedEffectForTest(t, g, id, []Mod{SetBasicLandTypesMod("Swamp")}, IndefiniteDuration())
	snap := g.CaptureSnapshot()
	if _, err := snap.RestoreStrict(); err != nil {
		t.Fatalf("a sound record does not restore: %v", err)
	}
	snap.ScopedEffects[0].Mods[0].Subtypes = []string{"Gate"}
	if _, err := snap.RestoreStrict(); err == nil {
		t.Error("a record setting a Gate restored")
	}
}

func TestLandTypeEffectsNameTheRecordsOnAPermanent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := pushScopedTestLand(g, me.ID, "Forest", "Basic Land — Forest")
	other := pushScopedTestLand(g, me.ID, "Plains", "Basic Land — Plains")
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(land),
			[]Mod{SetBasicLandTypesMod("Island")}, g.UntilEndOfTurnDuration(), "test")
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(land),
			[]Mod{AddSubtypesMod("Swamp"), ModifyPTMod(1, 1)}, g.UntilYourNextTurnDuration(me.ID), "Second")
		// Not a land type: not listed.
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(land),
			[]Mod{AddSubtypesMod("Orc")}, IndefiniteDuration(), "Third")
	})
	var got, none []LandTypeEffect
	g.WithWriteLock(func() {
		got = g.LandTypeEffectsForEffect(land)
		none = g.LandTypeEffectsForEffect(other)
	})
	want := []LandTypeEffect{
		{Types: []string{"Island"}, Until: "until end of turn", Source: "test"},
		{Types: []string{"Swamp"}, InAddition: true, Until: "until " + me.Name + "'s next turn", Source: "Second"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("land type effects = %+v, want %+v", got, want)
	}
	if len(none) != 0 {
		t.Errorf("an unaffected land lists %+v", none)
	}
}
