package testsl

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// leaA1Vanilla are the batch's bodies with no rules text.
var leaA1Vanilla = map[string]string{
	"6a462a69-3e42-41de-a3aa-a488d9f38d69": "Craw Wurm",
	"3c97c311-7ad5-47ec-b421-f6c3bfbda9fb": "Earth Elemental",
	"83c8a3a6-2e1a-4e26-8847-6d066f42d906": "Gray Ogre",
	"14c8f55d-d177-4c25-a931-ebeb9e6062a0": "Grizzly Bears",
	"342199e0-15b6-4824-83da-25caef2592b3": "Hill Giant",
	"8f1dae40-b307-446e-bbd2-86aa35813871": "Hurloon Minotaur",
	"b7c0bb85-fb87-4c73-bc1b-7b4dc763c7e8": "Ironroot Treefolk",
	"218d9277-c179-4de3-9c7f-79b5a6d4fa38": "Merfolk of the Pearl Trident",
	"a37159df-f6d7-4db6-85de-0ea77f425993": "Mons's Goblin Raiders",
	"ac41171e-c454-49e9-9004-c082ae099630": "Obsianus Golem",
	"c071be90-0531-40cc-af46-0cbe80c4ddd4": "Pearled Unicorn",
	"e0fefaf0-da20-4d58-8db7-019dba16c780": "Scathe Zombies",
	"c470fbe1-0538-46ec-8741-3143f5e78af8": "Water Elemental",
}

// leaA1Keyworded are the batch's creatures whose whole text is one
// printed keyword.
var leaA1Keyworded = map[string]struct {
	name    string
	keyword string
}{
	"7744bae4-a8b7-44a5-9b4c-0048ad4cc448": {"Air Elemental", "flying"},
	"c39ea5f9-6ec0-4697-897b-779e326754a7": {"Mahamoti Djinn", "flying"},
	"bbd30183-524c-4b93-b953-90853ec3f39f": {"Phantom Monster", "flying"},
	"01f07ab9-3c3e-4fea-9cd6-543a5b7056b8": {"Roc of Kher Ridges", "flying"},
	"a1f20695-6f08-4d5c-9fba-b0018bee298e": {"Scryb Sprites", "flying"},
	"b2a8fcf9-2ec2-40a9-9955-90f8605b11a3": {"War Mammoth", "trample"},
	"534ce8e2-53a1-407c-9881-d3b5347d43c3": {"Elvish Archers", "first strike"},
}

func TestLeaA1EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		effectstest.LeaA1ShatterOracle:   "Shatter",
		effectstest.LeaA1StoneRainOracle: "Stone Rain",
		effectstest.LeaA1SinkholeOracle:  "Sinkhole",
		effectstest.LeaA1IceStormOracle:  "Ice Storm",
		effectstest.LeaA1TerrorOracle:    "Terror",
	}
	for oracle, name := range leaA1Vanilla {
		all[oracle] = name
	}
	for oracle, k := range leaA1Keyworded {
		all[oracle] = k.name
	}
	if len(all) != 25 {
		t.Fatalf("the batch lists %d cards, want 25", len(all))
	}
	for oracle, name := range all {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestLeaA1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for oracle, name := range leaA1Vanilla {
		spec, _ := effects.Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", name)
		}
	}
}

func TestLeaA1KeywordedCreaturesPrintOneKeyword(t *testing.T) {
	for oracle, k := range leaA1Keyworded {
		spec, _ := effects.Lookup(oracle)
		if !slices.Equal(spec.PrintedKeywords, []string{k.keyword}) {
			t.Errorf("%s: printed keywords %v, want [%s]", k.name, spec.PrintedKeywords, k.keyword)
		}
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
			t.Errorf("%s declares more than its keyword", k.name)
		}
	}
}

// The printed keyword is live on the battlefield.
func TestLeaA1AirElementalHasFlyingOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elemental := effectstest.B12Push(g, me.ID, "Air Elemental", "Creature — Elemental", "7744bae4-a8b7-44a5-9b4c-0048ad4cc448", 4, 4)
	if !effectstest.EffectiveAbilitiesContain(t, g, elemental, "flying") {
		t.Error("Air Elemental should have flying on the battlefield")
	}
}

func TestLeaA1ShatterDestroysAnArtifactOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Shatter", "Instant", effectstest.LeaA1ShatterOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Fatal("a nonartifact creature should not be a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Shatter", "Instant", effectstest.LeaA1ShatterOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact should be in its owner's graveyard")
	}
}

func TestLeaA1LandDestructionDestroysALandOnly(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Stone Rain", effectstest.LeaA1StoneRainOracle},
		{"Sinkhole", effectstest.LeaA1SinkholeOracle},
		{"Ice Storm", effectstest.LeaA1IceStormOracle},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1]
			land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
			bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
			if err := effectstest.CastCatalogSpellErr(t, g, c.name, "Sorcery", c.oracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
				t.Fatal("a creature should not be a legal target")
			}
			effectstest.CastCatalogSpell(t, g, c.name, "Sorcery", c.oracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(land) || !opp.Graveyard.Contains(land) {
				t.Error("the land should be in its owner's graveyard")
			}
		})
	}
}

func TestLeaA1TerrorRefusesBlackAndArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Bear", TypeLine: "Creature — Bear", Colors: []string{"B"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	for _, id := range []uuid.UUID{black, golem} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Terror", "Instant", effectstest.LeaA1TerrorOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: id}}); err == nil {
			t.Errorf("%v should not be a legal target", id)
		}
	}
}

// "It can't be regenerated": a shield on the creature does not save it
// (CR 701.19c).
func TestLeaA1TerrorIgnoresARegenerationShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.P1Regenerate(t, g, bear)
	effectstest.CastCatalogSpell(t, g, "Terror", "Instant", effectstest.LeaA1TerrorOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the regenerated creature should still be destroyed")
	}
}
