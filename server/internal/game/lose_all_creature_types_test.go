package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// lose_all_creature_types_test.go — #4691, ModLoseAllCreatureTypes.

// goblinLordTestOracle keys the "other Goblins get +1/+1" static the
// lord tests install.
const goblinLordTestOracle = "00000000-0000-0000-0000-000000004691"

// cardHasSubtype is HasSubtype on the battlefield card after a recompute.
func cardHasSubtype(t *testing.T, g *Game, id uuid.UUID, subtype string) bool {
	t.Helper()
	c := layeredBattlefieldCard(t, g, id)
	return c.HasSubtype(subtype)
}

// Every creature type goes and nothing comes; the card types, the
// supertype and the subtype of another kind stay (CR 205.1a).
func TestLoseAllCreatureTypesKeepsCardTypesAndOtherSubtypes(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Legendary Artifact Creature — Equipment Elf Warrior")
	registerScopedEffectForTest(t, g, id, []Mod{LoseAllCreatureTypesMod()}, IndefiniteDuration())
	c := scopedEffectChar(t, g, id)
	if !reflect.DeepEqual(c.Types, []string{"Artifact", "Creature"}) {
		t.Errorf("types = %v, want [Artifact Creature]", c.Types)
	}
	if !reflect.DeepEqual(c.Subtypes, []string{"Equipment"}) {
		t.Errorf("subtypes = %v, want [Equipment]", c.Subtypes)
	}
	if !typeListHas(c.Supertypes, "Legendary") {
		t.Errorf("supertypes = %v, want Legendary kept", c.Supertypes)
	}
	if cardHasSubtype(t, g, id, "Elf") {
		t.Error("HasSubtype still answers for a lost creature type")
	}
}

// A changeling under it is no creature type, and keeps the changeling
// keyword: the effect removes types, not abilities (CR 702.73a).
func TestLoseAllCreatureTypesEndsAChangelingsEveryType(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Shapeshifter", KeywordChangeling)
	if !cardHasSubtype(t, g, id, "Elf") {
		t.Fatal("control: the changeling is not an Elf")
	}
	registerScopedEffectForTest(t, g, id, []Mod{LoseAllCreatureTypesMod()}, IndefiniteDuration())
	c := scopedEffectChar(t, g, id)
	if c.AllCreatureTypes || len(c.Subtypes) != 0 {
		t.Errorf("subtypes = %v (all %v), want none", c.Subtypes, c.AllCreatureTypes)
	}
	card := layeredBattlefieldCard(t, g, id)
	if card.HasSubtype("Elf") || card.HasSubtype("Shapeshifter") || HasAllCreatureTypes(&card) {
		t.Error("the changeling is still a creature type")
	}
	if !HasKeyword(&card, KeywordChangeling) {
		t.Error("the changeling keyword went; losing creature types removes no ability")
	}
}

// An earlier "gains all creature types" is applied first and stripped
// (CR 613.7).
func TestLoseAllCreatureTypesStripsAnEarlierAllCreatureTypes(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Goblin")
	registerScopedEffectForTest(t, g, id, []Mod{AllCreatureTypesMod()}, IndefiniteDuration())
	if !cardHasSubtype(t, g, id, "Elf") {
		t.Fatal("control: the Goblin with all creature types is not an Elf")
	}
	registerScopedEffectForTest(t, g, id, []Mod{LoseAllCreatureTypesMod()}, IndefiniteDuration())
	c := scopedEffectChar(t, g, id)
	if c.AllCreatureTypes || len(c.Subtypes) != 0 {
		t.Errorf("subtypes = %v (all %v), want none: the later loss wins", c.Subtypes, c.AllCreatureTypes)
	}
	if cardHasSubtype(t, g, id, "Elf") || cardHasSubtype(t, g, id, "Goblin") {
		t.Error("the creature still has a creature type")
	}
}

// A later "gains all creature types" (Oko, Lorwyn Liege's +2) applies
// on top and wins (CR 613.7).
func TestALaterAllCreatureTypesWinsOverLoseAllCreatureTypes(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Goblin")
	registerScopedEffectForTest(t, g, id, []Mod{LoseAllCreatureTypesMod()}, IndefiniteDuration())
	if cardHasSubtype(t, g, id, "Goblin") {
		t.Fatal("control: the Goblin kept its type")
	}
	registerScopedEffectForTest(t, g, id, []Mod{AllCreatureTypesMod()}, IndefiniteDuration())
	if c := scopedEffectChar(t, g, id); !c.AllCreatureTypes {
		t.Error("the later all-creature-types effect did not apply")
	}
	if !cardHasSubtype(t, g, id, "Elf") || !cardHasSubtype(t, g, id, "Goblin") {
		t.Error("the creature is not every creature type after the later effect")
	}
}

// A lord stops pumping the creature while the effect lasts, and pumps
// it again once the effect ends with the turn (CR 611.2a).
func TestLoseAllCreatureTypesTurnsOffALordUntilEndOfTurn(t *testing.T) {
	g := newActiveGame(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != goblinLordTestOracle {
			return nil
		}
		return []StaticAbility{{
			Layer:    Layer7PT,
			SubLayer: SubLayer7C_Modify,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.IsCreature() && target.InstanceID != source.InstanceID && target.HasSubtype("Goblin")
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power++
				c.Toughness++
			},
		}}
	})
	me := g.Seats[0].ID
	pushTypedTestCard(g, Card{
		Name: "Goblin Lord", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2,
		OracleID: goblinLordTestOracle, Owner: me, Controller: me,
	})
	goblin := pushTypedTestPermanent(g, me, "Creature — Goblin")
	shifter := pushTypedTestPermanent(g, me, "Creature — Shapeshifter", KeywordChangeling)
	for _, id := range []uuid.UUID{goblin, shifter} {
		if c := scopedEffectChar(t, g, id); c.Power != 3 || c.Toughness != 3 {
			t.Fatalf("control: %s is %d/%d, want 3/3 under the lord", id, c.Power, c.Toughness)
		}
	}

	for _, id := range []uuid.UUID{goblin, shifter} {
		var ok bool
		g.WithWriteLock(func() {
			ok = g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id),
				[]Mod{LoseAllCreatureTypesMod()}, g.UntilEndOfTurnDuration(), "test — loses all creature types")
		})
		if !ok {
			t.Fatal("RegisterScopedEffectForEffect registered nothing")
		}
	}
	for _, id := range []uuid.UUID{goblin, shifter} {
		if c := scopedEffectChar(t, g, id); c.Power != 2 || c.Toughness != 2 {
			t.Errorf("%s is %d/%d under the effect, want 2/2: the lord still sees a Goblin", id, c.Power, c.Toughness)
		}
		if cardHasSubtype(t, g, id, "Goblin") {
			t.Errorf("%s is still a Goblin", id)
		}
	}

	advancePastScopedCleanup(t, g)
	for _, id := range []uuid.UUID{goblin, shifter} {
		if c := scopedEffectChar(t, g, id); c.Power != 3 || c.Toughness != 3 {
			t.Errorf("%s is %d/%d after the turn, want 3/3: the effect did not end", id, c.Power, c.Toughness)
		}
		if !cardHasSubtype(t, g, id, "Goblin") {
			t.Errorf("%s is not a Goblin after the turn", id)
		}
	}
}

// The record survives a snapshot: the kind is written by name and a
// restored game still applies it.
func TestLoseAllCreatureTypesSurvivesASnapshot(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Goblin")
	registerScopedEffectForTest(t, g, id, []Mod{LoseAllCreatureTypesMod()}, IndefiniteDuration())
	_, restored := roundTrip(t, g)
	if c := scopedEffectChar(t, restored, id); len(c.Subtypes) != 0 || c.AllCreatureTypes {
		t.Errorf("restored subtypes = %v (all %v), want none", c.Subtypes, c.AllCreatureTypes)
	}
}
