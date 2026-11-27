package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o616Sunweb    = "832f96a4-0f4d-47bb-bc1c-3884e73feabc"
	o616Tiger     = "8ea35158-1c1a-46da-b466-0eafb376e464"
	o616Beetle    = "96415c51-2d44-4a96-9caa-9c4fbde59fca"
	o616Troops    = "9e9d5242-424d-4f61-b625-806295dbb0c7"
	o616Drake     = "b13cc6be-ee35-491e-a437-f7fc191f24c7"
	o616WallOfAir = "b2d3da40-e2f7-4480-9da9-33019d6f4071"
	o616Dryad     = "bb1963f9-261c-4d3f-a401-90aaced77e7f"
	o616Wildcat   = "dbf7b3ef-1ca4-4b0d-bec2-f5e60b150013"
)

func TestS82OldB616EveryCardIsRegisteredFullWithItsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		oracle, name, typeLine string
		keywords               []string
	}{
		{o616Sunweb, "Sunweb", "Creature — Wall", []string{"defender", "flying"}},
		{o616Tiger, "Sabretooth Tiger", "Creature — Cat", []string{"first strike"}},
		{o616Beetle, "Plague Beetle", "Creature — Insect", []string{"swampwalk"}},
		{o616Troops, "Standing Troops", "Creature — Human Soldier", []string{"vigilance"}},
		{o616Drake, "Fighting Drake", "Creature — Drake", []string{"flying"}},
		{o616WallOfAir, "Wall of Air", "Creature — Wall", []string{"defender", "flying"}},
		{o616Dryad, "Rushwood Dryad", "Creature — Dryad", []string{"forestwalk"}},
		{o616Wildcat, "Canyon Wildcat", "Creature — Cat", []string{"mountainwalk"}},
	} {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok {
			t.Errorf("%s is not registered", tc.name)
			continue
		}
		if spec.Name != tc.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", tc.name, spec.Name, spec.Completeness, spec.Caveats)
		}
		id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
		for _, kw := range tc.keywords {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

func TestS82OldB616WallsWithDefenderCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushCatalogPermanent(g, me.ID, "Wall of Air", "Creature — Wall", o616WallOfAir, false)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, opp.ID); err == nil {
		t.Error("a creature with defender attacked")
	}
}

func TestS82OldB616WallOfAirBlocksFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	wall := effectstest.PushCatalogPermanent(g, opp.ID, "Wall of Air", "Creature — Wall", o616WallOfAir, false)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, flier, wall)
}

func TestS82OldB616LandwalkIsUnblockableOnlyAgainstTheLandType(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, land string
	}{
		{"Plague Beetle", o616Beetle, "Creature — Insect", "Swamp"},
		{"Rushwood Dryad", o616Dryad, "Creature — Dryad", "Forest"},
		{"Canyon Wildcat", o616Wildcat, "Creature — Cat", "Mountain"},
	} {
		for _, withLand := range []bool{true, false} {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			walker := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			blocker := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
			if withLand {
				effectstest.B12Permanent(g, opp.ID, tc.land, "Basic Land — "+tc.land)
			} else {
				effectstest.B12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
			}
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(walker, opp.ID); err != nil {
				t.Fatalf("%s attack: %v", tc.name, err)
			}
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			err := g.DeclareBlocker(blocker, walker)
			if withLand && err == nil {
				t.Errorf("%s was blocked by a player who controls a %s", tc.name, tc.land)
			}
			if !withLand && err != nil {
				t.Errorf("%s could not be blocked without a %s: %v", tc.name, tc.land, err)
			}
		}
	}
}

func TestS82OldB616SunwebCantBlockSmallCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, me.ID, "Two-Two", "Creature — Bear", 2, 2)
	large := effectstest.B12Creature(g, me.ID, "Three-Three", "Creature — Ogre", 3, 3)
	sunweb := effectstest.PushCatalogPermanent(g, opp.ID, "Sunweb", "Creature — Wall", o616Sunweb, false)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: small, Target: opp.ID}, {Attacker: large, Target: opp.ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(sunweb, small); err == nil {
		t.Error("Sunweb blocked a creature with power 2")
	}
	if err := g.DeclareBlocker(sunweb, large); err != nil {
		t.Errorf("Sunweb could not block a creature with power 3: %v", err)
	}
}

func TestS82OldB616StandingTroopsAttackWithoutTapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	troops := effectstest.PushCatalogPermanent(g, me.ID, "Standing Troops", "Creature — Human Soldier", o616Troops, false)
	effectstest.DeclareAttack(t, g, opp.ID, troops)
	if c := effectstest.FindBattlefieldCardForTest(g, troops); c == nil || c.Tapped {
		t.Error("a vigilant attacker tapped")
	}
}
