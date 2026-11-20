package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zen_5583_b2_cards_test.go — Zendikar batch 2 (#5583). Every test plays
// the real catalog entry through the engine.

const (
	zenB2KrakenHatchlingOracle     = "08e6c786-d553-4c23-9137-cfad84146739"
	zenB2PillarfieldOxOracle       = "0f7c70c4-9795-4c76-8b27-043942a963c6"
	zenB2ShatterskullGiantOracle   = "aa241f75-ce34-49a8-8ffd-785172249bcf"
	zenB2StoneworkPumaOracle       = "841b1010-a7d3-415e-b76d-7b50fea909cb"
	zenB2VastwoodGorgerOracle      = "a17a86a3-9f9c-4e09-93e6-e543a70733bc"
	zenB2BogTattersOracle          = "e1e67a2f-61b6-49c5-8bcf-5d2b3f0df35d"
	zenB2CaravanHurdaOracle        = "2c0b8cc3-b1c1-498d-b34c-f317c2c9468f"
	zenB2CliffThreaderOracle       = "27882623-b5ad-49f2-b5ab-f0bdb93268a1"
	zenB2GiantScorpionOracle       = "557485d9-7537-4297-a589-c29d0f1ecd98"
	zenB2ShepherdOfTheLostOracle   = "d8c70f2f-985c-4982-97eb-99da3810e048"
	zenB2SkyRuinDrakeOracle        = "d9b7881b-915a-42f1-bd2f-0aaa900a8981"
	zenB2ZendikarFarguideOracle    = "01ab37a4-21d2-49f4-9b85-cb522a4d04ad"
	zenB2WelkinTernOracle          = "010c660a-a67b-48f0-a63a-15470eb89f69"
	zenB2TerraStomperOracle        = "d688baac-0574-4d88-b7eb-a7384c298ef5"
	zenB2SteppeLynxOracle          = "b16cef02-dcf8-4866-ac42-6ee929ddbea2"
	zenB2TerritorialBalothOracle   = "f772e767-a931-4891-94eb-d92a8ca7f8c1"
	zenB2WindriderEelOracle        = "9a3b16ae-210b-4d9b-83db-2801edfba394"
	zenB2PlatedGeopedeOracle       = "d1efc4fd-2978-4434-8fdf-ddc05484ebb1"
	zenB2HedronScrabblerOracle     = "180d9c48-6d6c-4d20-965a-0ee2ad57f99b"
	zenB2HagraCrocodileOracle      = "8bca3213-d8e1-4915-bd9a-9a291cdec1cd"
	zenB2GeyserGliderOracle        = "840d1002-abf4-4d05-945a-05c3fe9a2ebd"
	zenB2SurrakarMarauderOracle    = "23c52252-1e65-4e04-98d3-0ac06a565016"
	zenB2ShoalSerpentOracle        = "8a1d14c2-ca31-4711-b385-f4af0c1c0b04"
	zenB2MakindiShieldmateOracle   = "b8a65e49-2379-42bf-823e-c208c2550aeb"
	zenB2NimanaSellSwordOracle     = "5093df46-92cb-4f79-a5cf-4351c1fcf36d"
	zenB2OranRiefSurvivalistOracle = "36b4c44a-7682-4876-ae2a-35112cb55ec0"
	zenB2TuktukGruntsOracle        = "200ce414-8583-4bbd-88a6-ca30babb613d"
	zenB2UmaraRaptorOracle         = "a58ee84f-1d9c-4924-b7b1-14a9b2ba3b98"
	zenB2KazanduBlademasterOracle  = "133f5d30-d883-493e-93a1-cf9583db460b"
	zenB2HighlandBerserkerOracle   = "633849f6-fad7-49d7-8bf5-2a3a4510965f"
	zenB2JoragaBardOracle          = "a7eef84a-0b56-41d5-a298-dc3cf199e125"
	zenB2SeascapeAerialistOracle   = "877c6bb4-487c-4eea-b4af-f28fe1515fb6"
	zenB2OnduClericOracle          = "f4232466-dd6a-49bf-be6c-95905c3ded17"
	zenB2KazuulWarlordOracle       = "78504edb-6536-4a11-ac1c-7d53684f2226"
	zenB2CryptRipperOracle         = "18bcf4c5-7abc-4ca5-b805-4bbebf4e433c"
	zenB2MoltenRavagerOracle       = "697c851d-8475-479e-8783-cebdbe9a892d"
	zenB2CallerOfGalesOracle       = "86fd8f95-0f1c-42c0-a63e-567c59ce89c4"
	zenB2MerfolkSeastalkersOracle  = "86d592e1-4b2d-4411-9370-48326cf55cf1"
	zenB2GoblinWarPaintOracle      = "67ca79d7-9064-4605-8625-b1cfe5cb1b45"
	zenB2NimbusWingsOracle         = "0f35e73b-6a38-4185-a7d3-d237d67ba1cd"
	zenB2TrustyMacheteOracle       = "394fae8f-3757-4e08-b97c-5d7c451f4e72"
	zenB2SpidersilkNetOracle       = "04979f18-1ba8-47a3-a078-ab6b84bfac72"
	zenB2SlaughterCryOracle        = "924e20d5-b8a0-4883-9273-97ca01bf702a"
	zenB2ShieldmatesBlessingOracle = "4ed55e9c-84ea-42a3-a32f-469da56b3eb5"
	zenB2NarrowEscapeOracle        = "c96803dc-b69b-4fc3-a39b-1aad6d800e86"
	zenB2WindborneChargeOracle     = "b3ef64ad-a57e-44d9-9797-3f479741c497"

	zenB2BearType = "Creature — Bear"
	zenB2AllyType = "Creature — Kor Scout Ally"
)

var zenB2AllOracles = []string{
	zenB2KrakenHatchlingOracle, zenB2PillarfieldOxOracle, zenB2ShatterskullGiantOracle, zenB2StoneworkPumaOracle,
	zenB2VastwoodGorgerOracle, zenB2BogTattersOracle, zenB2CaravanHurdaOracle, zenB2CliffThreaderOracle,
	zenB2GiantScorpionOracle, zenB2ShepherdOfTheLostOracle, zenB2SkyRuinDrakeOracle, zenB2ZendikarFarguideOracle,
	zenB2WelkinTernOracle, zenB2TerraStomperOracle, zenB2SteppeLynxOracle, zenB2TerritorialBalothOracle,
	zenB2WindriderEelOracle, zenB2PlatedGeopedeOracle, zenB2HedronScrabblerOracle, zenB2HagraCrocodileOracle,
	zenB2GeyserGliderOracle, zenB2SurrakarMarauderOracle, zenB2ShoalSerpentOracle, zenB2MakindiShieldmateOracle,
	zenB2NimanaSellSwordOracle, zenB2OranRiefSurvivalistOracle, zenB2TuktukGruntsOracle, zenB2UmaraRaptorOracle,
	zenB2KazanduBlademasterOracle, zenB2HighlandBerserkerOracle, zenB2JoragaBardOracle, zenB2SeascapeAerialistOracle,
	zenB2OnduClericOracle, zenB2KazuulWarlordOracle, zenB2CryptRipperOracle, zenB2MoltenRavagerOracle,
	zenB2CallerOfGalesOracle, zenB2MerfolkSeastalkersOracle, zenB2GoblinWarPaintOracle, zenB2NimbusWingsOracle,
	zenB2TrustyMacheteOracle, zenB2SpidersilkNetOracle, zenB2SlaughterCryOracle, zenB2ShieldmatesBlessingOracle,
	zenB2NarrowEscapeOracle, zenB2WindborneChargeOracle,
}

// --- every card is in the catalog ----------------------------------------

func TestZenB2EveryCardIsRegisteredFull(t *testing.T) {
	if len(zenB2AllOracles) != 46 {
		t.Fatalf("%d cards listed, want 46", len(zenB2AllOracles))
	}
	for _, oracle := range zenB2AllOracles {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// The vanilla bodies register no ability at all.
func TestZenB2VanillaCreaturesHaveNoAbilities(t *testing.T) {
	for _, oracle := range []string{
		zenB2KrakenHatchlingOracle, zenB2PillarfieldOxOracle, zenB2ShatterskullGiantOracle,
		zenB2StoneworkPumaOracle, zenB2VastwoodGorgerOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not in the catalog", oracle)
		}
		if len(spec.PrintedKeywords)+len(spec.Triggered)+len(spec.Activated)+len(spec.Static) != 0 {
			t.Errorf("%s registers abilities; it has no rules text", spec.Name)
		}
	}
}

func TestZenB2TerraStomperCantBeCountered(t *testing.T) {
	spec, ok := Lookup(zenB2TerraStomperOracle)
	if !ok {
		t.Fatal("Terra Stomper is not in the catalog")
	}
	if !spec.CantBeCountered {
		t.Error("Terra Stomper can be countered")
	}
}

// --- printed keywords --------------------------------------------------------

func TestZenB2KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Bog Tatters", "Creature — Wraith", zenB2BogTattersOracle, []string{"swampwalk"}},
		{"Caravan Hurda", "Creature — Giant", zenB2CaravanHurdaOracle, []string{"lifelink"}},
		{"Cliff Threader", "Creature — Kor Scout", zenB2CliffThreaderOracle, []string{"mountainwalk"}},
		{"Giant Scorpion", "Creature — Scorpion", zenB2GiantScorpionOracle, []string{"deathtouch"}},
		{"Shepherd of the Lost", "Creature — Angel", zenB2ShepherdOfTheLostOracle, []string{"flying", "first strike", "vigilance"}},
		{"Sky Ruin Drake", "Creature — Drake", zenB2SkyRuinDrakeOracle, []string{"flying"}},
		{"Zendikar Farguide", "Creature — Elemental", zenB2ZendikarFarguideOracle, []string{"forestwalk"}},
		{"Welkin Tern", "Creature — Bird", zenB2WelkinTernOracle, []string{"flying"}},
		{"Terra Stomper", "Creature — Beast", zenB2TerraStomperOracle, []string{"trample"}},
		{"Windrider Eel", "Creature — Fish", zenB2WindriderEelOracle, []string{"flying"}},
		{"Plated Geopede", "Creature — Insect", zenB2PlatedGeopedeOracle, []string{"first strike"}},
		{"Shoal Serpent", "Creature — Serpent", zenB2ShoalSerpentOracle, []string{"defender"}},
		{"Makindi Shieldmate", "Creature — Kor Soldier Ally", zenB2MakindiShieldmateOracle, []string{"defender"}},
		{"Tuktuk Grunts", "Creature — Goblin Warrior Ally", zenB2TuktukGruntsOracle, []string{"haste"}},
		{"Umara Raptor", "Creature — Bird Ally", zenB2UmaraRaptorOracle, []string{"flying"}},
		{"Kazandu Blademaster", "Creature — Human Soldier Ally", zenB2KazanduBlademasterOracle, []string{"first strike", "vigilance"}},
		{"Crypt Ripper", "Creature — Shade", zenB2CryptRipperOracle, []string{"haste"}},
		{"Merfolk Seastalkers", "Creature — Merfolk Scout", zenB2MerfolkSeastalkersOracle, []string{"islandwalk"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- block restrictions ---------------------------------------------------------

func TestZenB2WelkinTernBlocksOnlyFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := b12Creature(g, me.ID, "Walker", zenB2BearType, 2, 2)
	flyer := apaPush(g, me.ID, me.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	tern := b12Push(g, opp.ID, "Welkin Tern", "Creature — Bird", zenB2WelkinTernOracle, 2, 1)
	brAttack(t, g, walker, flyer)
	if err := g.DeclareBlocker(tern, walker); err == nil {
		t.Error("Welkin Tern blocked a creature without flying")
	}
	if err := g.DeclareBlocker(tern, flyer); err != nil {
		t.Errorf("Welkin Tern can't block a flyer: %v", err)
	}
}

func TestZenB2HagraCrocodileCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := b12Creature(g, me.ID, "Walker", zenB2BearType, 2, 2)
	croc := b12Push(g, opp.ID, "Hagra Crocodile", "Creature — Crocodile", zenB2HagraCrocodileOracle, 3, 1)
	brAttack(t, g, walker)
	if err := g.DeclareBlocker(croc, walker); err == nil {
		t.Error("Hagra Crocodile blocked")
	}
}

// --- landfall -------------------------------------------------------------------

// zenB2PlayPlains plays a Plains for the active seat in its main phase.
func zenB2PlayPlains(t *testing.T, g *game.Game) {
	t.Helper()
	b31PlayLand(t, g, "Plains", "Basic Land — Plains", "")
	passPriorityAroundTable(t, g)
}

func TestZenB2LandfallPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		gainP, gainT           int
	}{
		{"Steppe Lynx", "Creature — Cat", zenB2SteppeLynxOracle, 0, 1, 2, 2},
		{"Territorial Baloth", "Creature — Beast", zenB2TerritorialBalothOracle, 4, 4, 2, 2},
		{"Windrider Eel", "Creature — Fish", zenB2WindriderEelOracle, 2, 2, 2, 2},
		{"Plated Geopede", "Creature — Insect", zenB2PlatedGeopedeOracle, 1, 1, 2, 2},
		{"Hedron Scrabbler", "Artifact Creature — Construct", zenB2HedronScrabblerOracle, 1, 1, 1, 1},
		{"Hagra Crocodile", "Creature — Crocodile", zenB2HagraCrocodileOracle, 3, 1, 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			zenB2PlayPlains(t, g)
			if p, tough := effectivePTOf(t, g, id); p != tc.power+tc.gainP || tough != tc.toughness+tc.gainT {
				t.Errorf("%s is %d/%d after a land entered, want %d/%d",
					tc.name, p, tough, tc.power+tc.gainP, tc.toughness+tc.gainT)
			}
		})
	}
}

func TestZenB2LandfallGrantsAKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Geyser Glider", "Creature — Elemental Beast", zenB2GeyserGliderOracle, "flying"},
		{"Surrakar Marauder", "Creature — Surrakar", zenB2SurrakarMarauderOracle, "intimidate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("%s has %s before any land entered", tc.name, tc.keyword)
			}
			zenB2PlayPlains(t, g)
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("landfall did not give %s %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestZenB2ShoalSerpentLosesDefenderOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	serpent := b12Push(g, me.ID, "Shoal Serpent", "Creature — Serpent", zenB2ShoalSerpentOracle, 5, 5)
	if !hasEffectiveKeyword(t, g, serpent, "defender") {
		t.Fatal("Shoal Serpent has no defender")
	}
	zenB2PlayPlains(t, g)
	if hasEffectiveKeyword(t, g, serpent, "defender") {
		t.Error("landfall did not remove defender")
	}
}

// --- Allies ------------------------------------------------------------------------

// Its own entry and another Ally's each offer a +1/+1 counter; a
// creature that is not an Ally offers nothing.
func TestZenB2AlliesPutACounterOnThemselves(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Makindi Shieldmate", "Creature — Kor Soldier Ally", zenB2MakindiShieldmateOracle, 0, 3},
		{"Nimana Sell-Sword", "Creature — Human Warrior Ally", zenB2NimanaSellSwordOracle, 2, 2},
		{"Oran-Rief Survivalist", "Creature — Human Warrior Ally", zenB2OranRiefSurvivalistOracle, 1, 1},
		{"Tuktuk Grunts", "Creature — Goblin Warrior Ally", zenB2TuktukGruntsOracle, 2, 2},
		{"Umara Raptor", "Creature — Bird Ally", zenB2UmaraRaptorOracle, 1, 1},
		{"Kazandu Blademaster", "Creature — Human Soldier Ally", zenB2KazanduBlademasterOracle, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			answerLatestTriggerPrompt(t, g, me.ID, true)
			passPriorityAroundTable(t, g)
			if got := plusOnes(g, id); got != 1 {
				t.Fatalf("%d +1/+1 counters after it entered, want 1", got)
			}
			castHoldCreature(t, g, "Recruit", "Creature — Human Ally", "", 1, 1)
			answerLatestTriggerPrompt(t, g, me.ID, true)
			passPriorityAroundTable(t, g)
			if got := plusOnes(g, id); got != 2 {
				t.Fatalf("%d +1/+1 counters after another Ally entered, want 2", got)
			}
			castHoldCreature(t, g, "Bear", zenB2BearType, "", 2, 2)
			if n := owedMays(g, me.ID); n != 0 {
				t.Errorf("a creature that is not an Ally offered %d triggers", n)
			}
		})
	}
}

func TestZenB2AllyCounterCanBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castHoldCreature(t, g, "Nimana Sell-Sword", "Creature — Human Warrior Ally", zenB2NimanaSellSwordOracle, 2, 2)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, id); got != 0 {
		t.Errorf("%d +1/+1 counters after declining, want 0", got)
	}
}

func TestZenB2AlliesGrantAKeywordToAllies(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Highland Berserker", "Creature — Human Berserker Ally", zenB2HighlandBerserkerOracle, "first strike"},
		{"Joraga Bard", "Creature — Elf Rogue Bard Ally", zenB2JoragaBardOracle, "vigilance"},
		{"Seascape Aerialist", "Creature — Merfolk Wizard Ally", zenB2SeascapeAerialistOracle, "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			ally := b12Creature(g, me.ID, "Scout", zenB2AllyType, 1, 1)
			bear := b12Creature(g, me.ID, "Bear", zenB2BearType, 2, 2)
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			answerLatestTriggerPrompt(t, g, me.ID, true)
			passPriorityAroundTable(t, g)
			for _, a := range []uuid.UUID{id, ally} {
				if !hasEffectiveKeyword(t, g, a, tc.keyword) {
					t.Errorf("an Ally has no %s", tc.keyword)
				}
			}
			if hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("a creature that is not an Ally gained %s", tc.keyword)
			}
		})
	}
}

func TestZenB2OnduClericGainsLifePerAlly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Scout", zenB2AllyType, 1, 1)
	b12Creature(g, me.ID, "Bear", zenB2BearType, 2, 2)
	life := me.Life
	castHoldCreature(t, g, "Ondu Cleric", "Creature — Kor Cleric Ally", zenB2OnduClericOracle, 1, 1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d (two Allies)", me.Life, life+2)
	}
}

func TestZenB2KazuulWarlordCountersEachAlly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ally := b12Creature(g, me.ID, "Scout", zenB2AllyType, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", zenB2BearType, 2, 2)
	warlord := castHoldCreature(t, g, "Kazuul Warlord", "Creature — Minotaur Warrior Ally", zenB2KazuulWarlordOracle, 3, 3)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if plusOnes(g, warlord) != 1 || plusOnes(g, ally) != 1 {
		t.Errorf("counters: Warlord %d, Ally %d, want 1 each", plusOnes(g, warlord), plusOnes(g, ally))
	}
	if n := plusOnes(g, bear); n != 0 {
		t.Errorf("a creature that is not an Ally got %d counters", n)
	}
}

// --- activated abilities -------------------------------------------------------------

func TestZenB2Firebreathers(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, color string
		power, toughness              int
		wantP, wantT                  int
	}{
		{"Crypt Ripper", "Creature — Shade", zenB2CryptRipperOracle, "B", 2, 2, 3, 3},
		{"Molten Ravager", "Creature — Elemental", zenB2MoltenRavagerOracle, "R", 0, 4, 1, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			me.ManaPool.AddMana(game.ManaToken{Color: tc.color})
			apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if p, tough := effectivePTOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestZenB2CallerOfGalesGrantsFlying(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	caller := b12Push(g, me.ID, "Caller of Gales", "Creature — Merfolk Wizard", zenB2CallerOfGalesOracle, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", zenB2BearType, 2, 2)
	fillPool(me, 1)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	apaActivate(t, g, me, caller, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
	if !hmlTapped(t, g, caller) {
		t.Error("Caller of Gales did not tap")
	}
}

func TestZenB2MerfolkSeastalkersTapsACreatureWithoutFlying(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	stalkers := b12Push(g, me.ID, "Merfolk Seastalkers", "Creature — Merfolk Scout", zenB2MerfolkSeastalkersOracle, 2, 3)
	bear := b12Creature(g, opp.ID, "Bear", zenB2BearType, 2, 2)
	fillPool(me, 2)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	apaActivate(t, g, me, stalkers, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !hmlTapped(t, g, bear) {
		t.Error("the target did not tap")
	}
}

func TestZenB2MerfolkSeastalkersCantTargetAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	stalkers := b12Push(g, me.ID, "Merfolk Seastalkers", "Creature — Merfolk Scout", zenB2MerfolkSeastalkersOracle, 2, 3)
	flyer := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	fillPool(me, 2)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	if err := g.ActivateCatalogAbility(me.ID, stalkers, 0, game.ActivateAbilityParams{Targets: cardTarget(flyer), Strict: true}); err == nil {
		t.Error("Merfolk Seastalkers targeted a creature with flying")
	}
}

// --- Auras and Equipment ---------------------------------------------------------------

func TestZenB2AurasPumpAndGrant(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword string
		wantP, wantT          int
	}{
		{"Goblin War Paint", zenB2GoblinWarPaintOracle, "haste", 4, 4},
		{"Nimbus Wings", zenB2NimbusWingsOracle, "flying", 3, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := top100ActiveSeat(g)
			bear := b12Creature(g, me.ID, "Bear", zenB2BearType, 2, 2)
			castCatalogSpell(t, g, tc.name, "Enchantment — Aura", tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tough := effectivePTOf(t, g, bear); p != tc.wantP || tough != tc.wantT {
				t.Errorf("the enchanted creature is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
			if !hasKeywordOnBattlefield(t, g, bear, tc.keyword) {
				t.Errorf("the enchanted creature has no %s", tc.keyword)
			}
		})
	}
}

func TestZenB2EquipmentBoostsAndGrants(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		keywords     []string
	}{
		{"Trusty Machete", zenB2TrustyMacheteOracle, 4, 3, nil},
		{"Spidersilk Net", zenB2SpidersilkNetOracle, 2, 4, []string{"reach"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			bear := b12Creature(g, me.ID, "Bear", zenB2BearType, 2, 2)
			eq := b12Push(g, me.ID, tc.name, "Artifact — Equipment", tc.oracle, 0, 0)
			fillPool(me, 2)
			equipTo(t, g, me.ID, eq, bear)
			if p, tough := effectivePTOf(t, g, bear); p != tc.power || tough != tc.tough {
				t.Errorf("the creature is %d/%d, want %d/%d", p, tough, tc.power, tc.tough)
			}
			for _, kw := range tc.keywords {
				if !hasKeywordOnBattlefield(t, g, bear, kw) {
					t.Errorf("the creature lacks %s", kw)
				}
			}
		})
	}
}

// --- instants and sorceries -------------------------------------------------------------

func TestZenB2SlaughterCryPumpsAndGrantsFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", zenB2BearType, 2, 2)
	castCatalogSpell(t, g, "Slaughter Cry", "Instant", zenB2SlaughterCryOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the target is %d/%d, want 5/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the target has no first strike")
	}
}

// A 3-point shield against 5 damage lets 2 through.
func TestZenB2ShieldmatesBlessingPreventsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	bear := pushBear(g, me.ID, "Shielded Bear", 10)
	castCatalogSpell(t, g, "Shieldmate's Blessing", "Instant", zenB2ShieldmatesBlessingOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 5)
	})
	if got := damageMarkedOn(g, bear); got != 2 {
		t.Errorf("marked = %d, want 2", got)
	}
}

func TestZenB2NarrowEscapeBouncesAndGainsFour(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", zenB2BearType, 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Narrow Escape", "Instant", zenB2NarrowEscapeOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the target did not return to its owner's hand")
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestZenB2NarrowEscapeCantTargetAnOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	bear := b12Creature(g, opp.ID, "Bear", zenB2BearType, 2, 2)
	id := b13HandCard(me, "Narrow Escape", "Instant", zenB2NarrowEscapeOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardTarget(bear)}); err == nil {
		t.Error("Narrow Escape targeted a permanent an opponent controls")
	}
}

func TestZenB2WindborneChargePumpsTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	a := b12Creature(g, me.ID, "Bear A", zenB2BearType, 2, 2)
	b := b12Creature(g, me.ID, "Bear B", zenB2BearType, 2, 2)
	castCatalogSpell(t, g, "Windborne Charge", "Sorcery", zenB2WindborneChargeOracle, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectivePTOf(t, g, id); p != 4 || tough != 4 {
			t.Errorf("a target is %d/%d, want 4/4", p, tough)
		}
		if !hasEffectiveKeyword(t, g, id, "flying") {
			t.Error("a target has no flying")
		}
	}
}
