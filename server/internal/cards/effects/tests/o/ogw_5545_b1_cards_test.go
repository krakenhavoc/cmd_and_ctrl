package testso

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ogw_5545_b1_cards_test.go — Oath of the Gatewatch, batch 1 (#5545):
// two vanilla bodies, keyword creatures, the five tapped duals and
// Wastes, the colorless-mana Eldrazi activations, the seven cohort
// creatures, enters triggers and a run of spells whose text a catalogued
// card already prints. Each test plays the real catalog entry through
// the engine and asserts what its text says.

const (
	ogwB1AncientCrabOracle      = "6bab6d0a-cd06-45e3-b780-d5354ee2a032"
	ogwB1CanopyGorgerOracle     = "1d614b37-d578-4f35-8fd8-3c7800d5f742"
	ogwB1AffaProtectorOracle    = "5b38a1e4-a6af-4094-9392-70a87759e20d"
	ogwB1MakindiAeronautOracle  = "9f980829-90c1-43e2-a103-915f38317ed0"
	ogwB1StormchaserMageOracle  = "32902854-2c64-431e-9d9c-9a23e52afb1b"
	ogwB1UmaraEntanglerOracle   = "d32cf19d-a974-4161-851d-dcbf37069e26"
	ogwB1BalothPupOracle        = "cb06cf58-3cc6-429a-9bb0-be4b1c53de30"
	ogwB1KorScythemasterOracle  = "21083222-12b6-4892-8078-1859140b7cc3"
	ogwB1ChitinousCloakOracle   = "d5487098-fd24-49af-9634-4827a5c2eac4"
	ogwB1CinderBarrensOracle    = "f155a05f-9f5e-4875-a407-103b85ce30ee"
	ogwB1MeanderingRiverOracle  = "2d9663be-c466-4191-85e2-a69ce0965432"
	ogwB1SubmergedBoneyardOracl = "f27d52e6-aab9-4f95-ae46-33d1173bf4fe"
	ogwB1TimberGorgeOracle      = "00b34fab-5a80-4a4d-b6cf-72479197677a"
	ogwB1TranquilExpanseOracle  = "a5478263-47c9-447f-9c7a-c77ce0752947"
	ogwB1WastesOracle           = "05d24b0c-904a-46b6-b42a-96a4d91a0dd4"
	ogwB1SeersLanternOracle     = "4b3dd7e5-e0b5-4773-bb16-37b757ae74d8"
	ogwB1WardenOracle           = "448da389-81db-4931-9cf7-3fde0ce9392f"
	ogwB1BlindingDroneOracle    = "c924e240-02a0-46b5-b004-d5c42db277ec"
	ogwB1HavocSowerOracle       = "098e6926-4a6a-4bbb-96bc-1f35cc7982ae"
	ogwB1ShriekerOracle         = "b2b5683b-f8ea-4af6-b847-953ab4a267c4"
	ogwB1MawOfKozilekOracle     = "11b529f0-e365-4500-b54a-44e2ac6e709f"
	ogwB1SlaughterDroneOracle   = "047e09c0-cbd9-45ba-9729-bf35ebe24e69"
	ogwB1StalkingDroneOracle    = "2659cc53-d20d-476f-bc64-4b35a4288d0e"
	ogwB1EssenceDepleterOracle  = "19ea37cc-e715-4a3c-87f3-5ed9d44f42d9"
	ogwB1ProphetOracle          = "6bc82093-0246-4c60-bf30-0b605bef01cd"
	ogwB1RelentlessHunterOracle = "6881e3ac-9e3d-49d2-b2ac-63feb6f0f2ff"
	ogwB1KorSkyClimberOracle    = "298a71d4-6dc1-444e-9057-0d25aef7fb64"
	ogwB1OnduWarClericOracle    = "d011153e-d12a-4eb1-b105-64729e37be12"
	ogwB1MalakirOracle          = "6ebfce76-82a2-4168-bd64-99cf16882dee"
	ogwB1ZulaportOracle         = "017f4471-db5a-4ad0-97b3-b1f12b52578a"
	ogwB1DranasChosenOracle     = "185b3b9f-f798-4ca4-af99-b17a24fcc1c6"
	ogwB1MundasVanguardOracle   = "1d7ee372-3a3d-4e23-a1ce-b8b8cdd4bdae"
	ogwB1SpawnbinderOracle      = "2d49c50b-3e2d-4e74-9b8a-c2f9a495e470"
	ogwB1ZadasCommandoOracle    = "d57d1988-5962-4df9-8342-cf04233fd6c7"
	ogwB1SaddlebackLagacOracle  = "3d720d55-3df8-40fd-bc93-ad99c1105249"
	ogwB1GladehartOracle        = "4d937ea3-bfd4-4cdf-b722-08f2b99ea2a5"
	ogwB1LoamLarvaOracle        = "528a19d7-ee79-45a5-b27f-e0bff4213193"
	ogwB1CinderHellionOracle    = "9e1ad977-2814-4336-93ec-57b7736a629f"
	ogwB1ScionSummonerOracle    = "fa60f3a6-5834-4059-adea-dd17366b7885"
	ogwB1VampireEnvoyOracle     = "cfc50aef-b17b-4fc1-af07-fe43b89689b7"
	ogwB1RealityHemorrhageOracl = "21465dcd-c93a-4a2c-ae10-e22ebd5cae2f"
	ogwB1OblivionStrikeOracle   = "ad3564f5-11f7-456b-8881-4db00a18a2ed"
	ogwB1ImmolatingGlareOracle  = "046b6989-5b56-4b52-a655-109b1eb26a93"
	ogwB1NaturalStateOracle     = "f500f4cc-d5ee-4995-b7b5-96a5f63dfb47"
	ogwB1TarSnareOracle         = "3621bd8c-c883-462f-ad56-c01deecb2720"
	ogwB1TearsOfValakutOracle   = "fce335fb-bc67-4fa4-8d38-097f2a661332"
	ogwB1VoidShatterOracle      = "3d9f854e-1ee5-4aa1-a33b-d3ae08f15dd0"
	ogwB1AlliedReinforcementsOr = "2baf2ee6-e290-4a2b-89a0-0e725edde06d"
	ogwB1LeadByExampleOracle    = "9a50418c-3cdc-4e3e-b02c-6da646093120"
	ogwB1SlipThroughSpaceOracle = "2516949d-75bb-4aea-a41d-0298bb129471"

	ogwB1BearType  = "Creature — Bear"
	ogwB1AllyType  = "Creature — Kor Ally"
	ogwB1DroneType = "Creature — Eldrazi Drone"
)

// ogwB1All is every card of the batch, by oracle ID.
var ogwB1All = map[string]string{
	ogwB1AncientCrabOracle:      "Ancient Crab",
	ogwB1CanopyGorgerOracle:     "Canopy Gorger",
	ogwB1AffaProtectorOracle:    "Affa Protector",
	ogwB1MakindiAeronautOracle:  "Makindi Aeronaut",
	ogwB1StormchaserMageOracle:  "Stormchaser Mage",
	ogwB1UmaraEntanglerOracle:   "Umara Entangler",
	ogwB1BalothPupOracle:        "Baloth Pup",
	ogwB1KorScythemasterOracle:  "Kor Scythemaster",
	ogwB1ChitinousCloakOracle:   "Chitinous Cloak",
	ogwB1CinderBarrensOracle:    "Cinder Barrens",
	ogwB1MeanderingRiverOracle:  "Meandering River",
	ogwB1SubmergedBoneyardOracl: "Submerged Boneyard",
	ogwB1TimberGorgeOracle:      "Timber Gorge",
	ogwB1TranquilExpanseOracle:  "Tranquil Expanse",
	ogwB1WastesOracle:           "Wastes",
	ogwB1SeersLanternOracle:     "Seer's Lantern",
	ogwB1WardenOracle:           "Warden of Geometries",
	ogwB1BlindingDroneOracle:    "Blinding Drone",
	ogwB1HavocSowerOracle:       "Havoc Sower",
	ogwB1ShriekerOracle:         "Kozilek's Shrieker",
	ogwB1MawOfKozilekOracle:     "Maw of Kozilek",
	ogwB1SlaughterDroneOracle:   "Slaughter Drone",
	ogwB1StalkingDroneOracle:    "Stalking Drone",
	ogwB1EssenceDepleterOracle:  "Essence Depleter",
	ogwB1ProphetOracle:          "Prophet of Distortion",
	ogwB1RelentlessHunterOracle: "Relentless Hunter",
	ogwB1KorSkyClimberOracle:    "Kor Sky Climber",
	ogwB1OnduWarClericOracle:    "Ondu War Cleric",
	ogwB1MalakirOracle:          "Malakir Soothsayer",
	ogwB1ZulaportOracle:         "Zulaport Chainmage",
	ogwB1DranasChosenOracle:     "Drana's Chosen",
	ogwB1MundasVanguardOracle:   "Munda's Vanguard",
	ogwB1SpawnbinderOracle:      "Spawnbinder Mage",
	ogwB1ZadasCommandoOracle:    "Zada's Commando",
	ogwB1SaddlebackLagacOracle:  "Saddleback Lagac",
	ogwB1GladehartOracle:        "Gladehart Cavalry",
	ogwB1LoamLarvaOracle:        "Loam Larva",
	ogwB1CinderHellionOracle:    "Cinder Hellion",
	ogwB1ScionSummonerOracle:    "Scion Summoner",
	ogwB1VampireEnvoyOracle:     "Vampire Envoy",
	ogwB1RealityHemorrhageOracl: "Reality Hemorrhage",
	ogwB1OblivionStrikeOracle:   "Oblivion Strike",
	ogwB1ImmolatingGlareOracle:  "Immolating Glare",
	ogwB1NaturalStateOracle:     "Natural State",
	ogwB1TarSnareOracle:         "Tar Snare",
	ogwB1TearsOfValakutOracle:   "Tears of Valakut",
	ogwB1VoidShatterOracle:      "Void Shatter",
	ogwB1AlliedReinforcementsOr: "Allied Reinforcements",
	ogwB1LeadByExampleOracle:    "Lead by Example",
	ogwB1SlipThroughSpaceOracle: "Slip Through Space",
}

// ogwB1Charge adds n colorless mana ({C}) to p's pool.
func ogwB1Charge(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
}

// ogwB1Active is the active player in a main phase.
func ogwB1Active(t *testing.T, g *game.Game) *game.Player {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	return effectstest.Top100ActiveSeat(g)
}

// --- every card is in the catalog ----------------------------------------

func TestOgw5545B1EveryCardIsRegisteredFull(t *testing.T) {
	if len(ogwB1All) != 50 {
		t.Fatalf("the batch lists %d cards, want 50", len(ogwB1All))
	}
	for oracle, name := range ogwB1All {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOgw5545B1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for _, oracle := range []string{ogwB1AncientCrabOracle, ogwB1CanopyGorgerOracle} {
		spec, _ := effects.Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s is a vanilla body and declares abilities", spec.Name)
		}
	}
}

// Devoid (CR 702.114a) is declared on every devoid card of the batch.
func TestOgw5545B1DevoidCardsDeclareDevoid(t *testing.T) {
	for _, oracle := range []string{
		ogwB1BlindingDroneOracle, ogwB1HavocSowerOracle, ogwB1ShriekerOracle, ogwB1MawOfKozilekOracle,
		ogwB1SlaughterDroneOracle, ogwB1StalkingDroneOracle, ogwB1EssenceDepleterOracle, ogwB1ProphetOracle,
		ogwB1ScionSummonerOracle, ogwB1RealityHemorrhageOracl, ogwB1OblivionStrikeOracle, ogwB1VoidShatterOracle,
		ogwB1SlipThroughSpaceOracle,
	} {
		spec, _ := effects.Lookup(oracle)
		if !slices.Contains(spec.PrintedKeywords, game.KeywordDevoid) {
			t.Errorf("%s does not declare devoid", spec.Name)
		}
	}
}

// --- printed keywords ------------------------------------------------------

func TestOgw5545B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Affa Protector", "Creature — Human Soldier Ally", ogwB1AffaProtectorOracle, []string{"vigilance"}},
		{"Makindi Aeronaut", "Creature — Kor Scout Ally", ogwB1MakindiAeronautOracle, []string{"flying"}},
		{"Stormchaser Mage", "Creature — Human Wizard", ogwB1StormchaserMageOracle, []string{"flying", "haste", "prowess"}},
		{"Umara Entangler", "Creature — Merfolk Rogue Ally", ogwB1UmaraEntanglerOracle, []string{"prowess"}},
		{"Warden of Geometries", ogwB1DroneType, ogwB1WardenOracle, []string{"vigilance"}},
		{"Zada's Commando", "Creature — Goblin Archer Ally", ogwB1ZadasCommandoOracle, []string{"first strike"}},
		{"Cinder Hellion", "Creature — Hellion", ogwB1CinderHellionOracle, []string{"trample"}},
		{"Vampire Envoy", "Creature — Vampire Cleric Ally", ogwB1VampireEnvoyOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- statics ---------------------------------------------------------------

func TestOgw5545B1BalothPupTramplesOnlyWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	pup := effectstest.B12Push(g, g.Seats[0].ID, "Baloth Pup", "Creature — Beast", ogwB1BalothPupOracle, 3, 1)
	if effectstest.HasEffectiveKeyword(t, g, pup, "trample") {
		t.Fatal("Baloth Pup has trample with no counter on it")
	}
	effectstest.SncCounter(g, pup, game.CounterPlusOne, 1)
	if !effectstest.HasEffectiveKeyword(t, g, pup, "trample") {
		t.Error("Baloth Pup has no trample with a +1/+1 counter on it")
	}
}

func TestOgw5545B1KorScythemasterHasFirstStrikeOnlyWhileAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	scythe := effectstest.B12Push(g, me.ID, "Kor Scythemaster", "Creature — Kor Soldier Ally", ogwB1KorScythemasterOracle, 3, 1)
	if effectstest.HasEffectiveKeyword(t, g, scythe, "first strike") {
		t.Fatal("Kor Scythemaster has first strike before it attacks")
	}
	effectstest.DeclareAttack(t, g, them.ID, scythe)
	if !effectstest.HasEffectiveKeyword(t, g, scythe, "first strike") {
		t.Error("Kor Scythemaster has no first strike while attacking")
	}
}

func TestOgw5545B1ChitinousCloakEquipsForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := ogwB1Active(t, g)
	cloak := effectstest.PushCatalogPermanent(g, me.ID, "Chitinous Cloak", "Artifact — Equipment", ogwB1ChitinousCloakOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", ogwB1BearType, 2, 2)
	effectstest.ApaMana(me, "C", "C", "C")
	effectstest.ApaActivate(t, g, me, cloak, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the equipped Bear is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("the equipped Bear has no menace")
	}
}

// --- lands -------------------------------------------------------------------

func TestOgw5545B1TappedDualsEnterTappedAndTapForEitherColor(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		a, b         string
	}{
		{"Cinder Barrens", ogwB1CinderBarrensOracle, "B", "R"},
		{"Meandering River", ogwB1MeanderingRiverOracle, "W", "U"},
		{"Submerged Boneyard", ogwB1SubmergedBoneyardOracl, "U", "B"},
		{"Timber Gorge", ogwB1TimberGorgeOracle, "R", "G"},
		{"Tranquil Expanse", ogwB1TranquilExpanseOracle, "G", "W"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := ogwB1Active(t, g)
			played := effectstest.B12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
			if !effectstest.HmlTapped(t, g, played) {
				t.Errorf("%s entered untapped", tc.name)
			}
			for _, color := range []string{tc.a, tc.b} {
				land := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
				if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{color}}); err != nil {
					t.Fatalf("tapping %s for {%s}: %v", tc.name, color, err)
				}
				if got := me.ManaPool.CountColor(color); got != 1 {
					t.Errorf("the pool holds %d {%s}, want 1", got, color)
				}
				me.ManaPool.EmptyPool()
			}
		})
	}
}

func TestOgw5545B1ColorlessSourcesTapForC(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Wastes", "Basic Land", ogwB1WastesOracle},
		{"Seer's Lantern", "Artifact", ogwB1SeersLanternOracle},
		{"Warden of Geometries", ogwB1DroneType, ogwB1WardenOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := ogwB1Active(t, g)
			id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("tapping %s: %v", tc.name, err)
			}
			if got := me.ManaPool.CountColor("C"); got != 1 {
				t.Errorf("the pool holds %d {C}, want 1", got)
			}
			if !effectstest.HmlTapped(t, g, id) {
				t.Errorf("%s did not tap", tc.name)
			}
		})
	}
}

func TestOgw5545B1SeersLanternScriesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := ogwB1Active(t, g)
	lantern := effectstest.PushCatalogPermanent(g, me.ID, "Seer's Lantern", "Artifact", ogwB1SeersLanternOracle, false)
	effectstest.ApaMana(me, "C", "C")
	effectstest.ApaActivate(t, g, me, lantern, 0, game.ActivateAbilityParams{})
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scrying %d cards, want 1", len(c.ScryCards))
	}
}

// --- colorless-mana activations ----------------------------------------------

// {C} is paid only with colorless mana (CR 107.4c): black mana can't pay
// Maw of Kozilek's cost.
func TestOgw5545B1ColorlessCostRefusesColoredMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := ogwB1Active(t, g)
	maw := effectstest.B12Push(g, me.ID, "Maw of Kozilek", ogwB1DroneType, ogwB1MawOfKozilekOracle, 2, 5)
	effectstest.ApaMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, maw, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("black mana paid {C}")
	}
}

func TestOgw5545B1SelfPumps(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		mana         []string
		p, tough     int
		keyword      string
	}{
		{"Havoc Sower", ogwB1HavocSowerOracle, []string{"C", "C"}, 5, 4, ""},
		{"Kozilek's Shrieker", ogwB1ShriekerOracle, []string{"C"}, 4, 2, "menace"},
		{"Maw of Kozilek", ogwB1MawOfKozilekOracle, []string{"C"}, 4, 3, ""},
		{"Slaughter Drone", ogwB1SlaughterDroneOracle, []string{"C"}, 2, 2, "deathtouch"},
		{"Stalking Drone", ogwB1StalkingDroneOracle, []string{"C"}, 3, 4, ""},
		{"Relentless Hunter", ogwB1RelentlessHunterOracle, []string{"C", "R", "G"}, 4, 4, "trample"},
		{"Kor Sky Climber", ogwB1KorSkyClimberOracle, []string{"C", "W"}, 3, 2, "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := ogwB1Active(t, g)
			base := map[string][2]int{
				"Havoc Sower": {3, 3}, "Kozilek's Shrieker": {3, 2}, "Maw of Kozilek": {2, 5},
				"Slaughter Drone": {2, 2}, "Stalking Drone": {2, 2}, "Relentless Hunter": {3, 3}, "Kor Sky Climber": {3, 2},
			}[tc.name]
			id := effectstest.B12Push(g, me.ID, tc.name, ogwB1DroneType, tc.oracle, base[0], base[1])
			if tc.keyword != "" && effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("%s has %s before activating", tc.name, tc.keyword)
			}
			effectstest.ApaMana(me, tc.mana...)
			effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if p, tough := effectstest.PtOf(t, g, id); p != tc.p || tough != tc.tough {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.p, tc.tough)
			}
			if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestOgw5545B1StalkingDroneActivatesOnlyOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := ogwB1Active(t, g)
	drone := effectstest.B12Push(g, me.ID, "Stalking Drone", ogwB1DroneType, ogwB1StalkingDroneOracle, 2, 2)
	ogwB1Charge(me, 1)
	effectstest.ApaActivate(t, g, me, drone, 0, game.ActivateAbilityParams{})
	ogwB1Charge(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, drone, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Stalking Drone activated twice in one turn")
	}
}

func TestOgw5545B1BlindingDroneTapsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := ogwB1Active(t, g), effectstest.Top100Opponent(g)
	drone := effectstest.B12Push(g, me.ID, "Blinding Drone", ogwB1DroneType, ogwB1BlindingDroneOracle, 1, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", ogwB1BearType, 2, 2)
	ogwB1Charge(me, 1)
	effectstest.ApaActivate(t, g, me, drone, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !effectstest.HmlTapped(t, g, bear) {
		t.Error("the target is untapped")
	}
	if !effectstest.HmlTapped(t, g, drone) {
		t.Error("Blinding Drone did not tap")
	}
}

func TestOgw5545B1EssenceDepleterDrainsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := ogwB1Active(t, g), effectstest.Top100Opponent(g)
	depleter := effectstest.B12Push(g, me.ID, "Essence Depleter", ogwB1DroneType, ogwB1EssenceDepleterOracle, 2, 3)
	mine, theirs := me.Life, opp.Life
	ogwB1Charge(me, 2)
	effectstest.ApaActivate(t, g, me, depleter, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if opp.Life != theirs-1 || me.Life != mine+1 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, mine+1, theirs-1)
	}
}

func TestOgw5545B1ProphetOfDistortionDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := ogwB1Active(t, g)
	prophet := effectstest.B12Push(g, me.ID, "Prophet of Distortion", ogwB1DroneType, ogwB1ProphetOracle, 1, 2)
	hand := me.Hand.Size()
	ogwB1Charge(me, 4)
	effectstest.ApaActivate(t, g, me, prophet, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

// --- cohort --------------------------------------------------------------------

// ogwB1Cohort puts a cohort creature and an untapped Ally on the active
// player's side of the table, in a main phase.
func ogwB1Cohort(t *testing.T, g *game.Game, name, typeLine, oracle string) (me *game.Player, cohort, ally uuid.UUID) {
	t.Helper()
	me = ogwB1Active(t, g)
	cohort = effectstest.B12Push(g, me.ID, name, typeLine, oracle, 2, 2)
	ally = effectstest.B12Creature(g, me.ID, "Ally", ogwB1AllyType, 1, 1)
	return me, cohort, ally
}

// The second half of the cost is an untapped Ally other than the cohort
// creature, which its own {T} has already tapped.
func TestOgw5545B1CohortNeedsAnotherUntappedAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, cleric, _ := ogwB1Cohort(t, g, "Ondu War Cleric", "Creature — Human Cleric Ally", ogwB1OnduWarClericOracle)
	bear := effectstest.B12Creature(g, me.ID, "Bear", ogwB1BearType, 2, 2)
	for name, ids := range map[string][]uuid.UUID{
		"no Ally":    nil,
		"a non-Ally": {bear},
		"the cleric": {cleric},
	} {
		if err := g.ActivateCatalogAbility(me.ID, cleric, 0, game.ActivateAbilityParams{Strict: true, TapIDs: ids}); err == nil {
			t.Errorf("cohort paid with %s", name)
		}
	}
}

func TestOgw5545B1CohortAbilities(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		target                 string // "", "opponent", "creature"
		check                  func(t *testing.T, g *game.Game, me, opp *game.Player, cohort, ally, victim uuid.UUID, life, oppLife, hand int)
	}{
		{"Ondu War Cleric", "Creature — Human Cleric Ally", ogwB1OnduWarClericOracle, "",
			func(t *testing.T, _ *game.Game, me, _ *game.Player, _, _, _ uuid.UUID, life, _, _ int) {
				if me.Life != life+2 {
					t.Errorf("life %d, want %d", me.Life, life+2)
				}
			}},
		{"Malakir Soothsayer", "Creature — Vampire Shaman Ally", ogwB1MalakirOracle, "",
			func(t *testing.T, _ *game.Game, me, _ *game.Player, _, _, _ uuid.UUID, life, _, hand int) {
				if me.Hand.Size() != hand+1 || me.Life != life-1 {
					t.Errorf("hand %d life %d, want %d and %d", me.Hand.Size(), me.Life, hand+1, life-1)
				}
			}},
		{"Zulaport Chainmage", "Creature — Human Shaman Ally", ogwB1ZulaportOracle, "opponent",
			func(t *testing.T, _ *game.Game, _, opp *game.Player, _, _, _ uuid.UUID, _, oppLife, _ int) {
				if opp.Life != oppLife-2 {
					t.Errorf("the opponent is at %d, want %d", opp.Life, oppLife-2)
				}
			}},
		{"Zada's Commando", "Creature — Goblin Archer Ally", ogwB1ZadasCommandoOracle, "opponent",
			func(t *testing.T, _ *game.Game, _, opp *game.Player, _, _, _ uuid.UUID, _, oppLife, _ int) {
				if opp.Life != oppLife-1 {
					t.Errorf("the opponent is at %d, want %d", opp.Life, oppLife-1)
				}
			}},
		{"Drana's Chosen", "Creature — Vampire Shaman Ally", ogwB1DranasChosenOracle, "",
			func(t *testing.T, g *game.Game, me, _ *game.Player, _, _, _ uuid.UUID, _, _, _ int) {
				zombie, ok := effectstest.VisBattlefieldNamed(g, me.ID, "Zombie")
				if !ok {
					t.Fatal("no Zombie token")
				}
				if !zombie.Tapped || zombie.Power != 2 || zombie.Toughness != 2 {
					t.Errorf("the Zombie is %d/%d tapped=%v, want a tapped 2/2", zombie.Power, zombie.Toughness, zombie.Tapped)
				}
			}},
		{"Munda's Vanguard", "Creature — Kor Knight Ally", ogwB1MundasVanguardOracle, "",
			func(t *testing.T, g *game.Game, _, _ *game.Player, cohort, ally, victim uuid.UUID, _, _, _ int) {
				if effectstest.PlusOnes(g, cohort) != 1 || effectstest.PlusOnes(g, ally) != 1 {
					t.Error("not every creature you control got a counter")
				}
				if effectstest.PlusOnes(g, victim) != 0 {
					t.Error("an opponent's creature got a counter")
				}
			}},
		{"Spawnbinder Mage", "Creature — Human Wizard Ally", ogwB1SpawnbinderOracle, "creature",
			func(t *testing.T, g *game.Game, _, _ *game.Player, _, _, victim uuid.UUID, _, _, _ int) {
				if !effectstest.HmlTapped(t, g, victim) {
					t.Error("the target is untapped")
				}
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, cohort, ally := ogwB1Cohort(t, g, tc.name, tc.typeLine, tc.oracle)
			opp := effectstest.Top100Opponent(g)
			victim := effectstest.B12Creature(g, opp.ID, "Victim", ogwB1BearType, 2, 2)
			params := game.ActivateAbilityParams{TapIDs: []uuid.UUID{ally}}
			switch tc.target {
			case "opponent":
				params.Targets = []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
			case "creature":
				params.Targets = effectstest.CardTarget(victim)
			}
			life, oppLife, hand := me.Life, opp.Life, me.Hand.Size()
			effectstest.ApaActivate(t, g, me, cohort, 0, params)
			if !effectstest.HmlTapped(t, g, cohort) || !effectstest.HmlTapped(t, g, ally) {
				t.Error("the cost did not tap the cohort creature and the Ally")
			}
			tc.check(t, g, me, opp, cohort, ally, victim, life, oppLife, hand)
		})
	}
}

// --- triggered abilities -----------------------------------------------------------

func TestOgw5545B1SaddlebackLagacSupportsTwoOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := ogwB1Active(t, g)
	a := effectstest.B12Creature(g, me.ID, "A", ogwB1BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", ogwB1BearType, 2, 2)
	lagac := effectstest.CastHoldCreature(t, g, "Saddleback Lagac", "Creature — Lizard", ogwB1SaddlebackLagacOracle, 3, 1)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no support target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, append(effectstest.TdCard(a), effectstest.TdCard(b)...)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, a) != 1 || effectstest.PlusOnes(g, b) != 1 {
		t.Error("support did not put a counter on each target")
	}
	if effectstest.PlusOnes(g, lagac) != 0 {
		t.Error("support put a counter on the Lagac itself")
	}
}

func TestOgw5545B1GladehartCavalrySupportsAndGainsForCounteredDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := ogwB1Active(t, g)
	marked := effectstest.B12Creature(g, me.ID, "Marked", ogwB1BearType, 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", ogwB1BearType, 2, 2)
	effectstest.CastHoldCreature(t, g, "Gladehart Cavalry", "Creature — Elf Knight", ogwB1GladehartOracle, 6, 6)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no support target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.TdCard(marked)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, marked) != 1 {
		t.Fatal("support did not put a counter on the target")
	}
	life := me.Life
	effectstest.B27Kill(g, plain)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("a creature with no counter died and life went to %d", me.Life)
	}
	effectstest.B27Kill(g, marked)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d after a countered creature died, want %d", me.Life, life+2)
	}
}

func TestOgw5545B1LoamLarvaPutsABasicLandOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := ogwB1Active(t, g)
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Grizzly Bears", "Creature — Bear"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
	)
	effectstest.CastHoldCreature(t, g, "Loam Larva", "Creature — Insect", ogwB1LoamLarvaOracle, 1, 3)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger did not offer a search")
	}
	if effectstest.SearchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Error("a nonland card is a legal find")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.Name != "Forest" {
		t.Errorf("the top card is %q, want Forest", top.Name)
	}
	if me.Library.Size() != 3 {
		t.Errorf("library = %d, want 3", me.Library.Size())
	}
}

func TestOgw5545B1CinderHellionHitsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := ogwB1Active(t, g), effectstest.Top100Opponent(g)
	life := opp.Life
	effectstest.CastHoldCreature(t, g, "Cinder Hellion", "Creature — Hellion", ogwB1CinderHellionOracle, 4, 4)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
}

func TestOgw5545B1ScionSummonerMakesAScion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := ogwB1Active(t, g)
	effectstest.CastHoldCreature(t, g, "Scion Summoner", ogwB1DroneType, ogwB1ScionSummonerOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Eldrazi Scion"); got != 1 {
		t.Errorf("%d Eldrazi Scions, want 1", got)
	}
}

func TestOgw5545B1VampireEnvoyGainsWhenItTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	envoy := effectstest.B12Push(g, me.ID, "Vampire Envoy", "Creature — Vampire Cleric Ally", ogwB1VampireEnvoyOracle, 1, 4)
	life := me.Life
	effectstest.DeclareAttack(t, g, them.ID, envoy)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d after it attacked, want %d", me.Life, life+1)
	}
}

// --- the spells ----------------------------------------------------------------

func TestOgw5545B1RealityHemorrhageDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Reality Hemorrhage", "Instant", ogwB1RealityHemorrhageOracl, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-2)
	}
}

func TestOgw5545B1OblivionStrikeExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", ogwB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Oblivion Strike", "Sorcery", ogwB1OblivionStrikeOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Error("the creature was not exiled")
	}
}

func TestOgw5545B1ImmolatingGlareDestroysOnlyAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	raider := effectstest.B12Creature(g, me.ID, "Raider", ogwB1BearType, 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", ogwB1BearType, 2, 2)
	effectstest.DeclareAttack(t, g, them.ID, raider)
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Immolating Glare", "Instant", ogwB1ImmolatingGlareOracle),
		game.CastSpellParams{Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("Immolating Glare targeted a creature that isn't attacking")
	}
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Immolating Glare", "Instant", ogwB1ImmolatingGlareOracle),
		game.CastSpellParams{Targets: effectstest.CardTarget(raider)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) {
		t.Error("the attacker survived")
	}
}

func TestOgw5545B1NaturalStateDestroysOnlyCheapArtifactsAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	cheap := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Cheap Rock", TypeLine: "Artifact", ManaCost: "{3}"})
	pricey := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Pricey Rock", TypeLine: "Artifact", ManaCost: "{4}"})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", ogwB1BearType, 2, 2)
	for _, bad := range []uuid.UUID{pricey, bear} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Natural State", "Instant", ogwB1NaturalStateOracle, effectstest.CardTarget(bad)); err == nil {
			t.Fatalf("Natural State targeted %s", effectstest.FindCardAnywhere(t, g, bad).Name)
		}
	}
	effectstest.CastCatalogSpell(t, g, "Natural State", "Instant", ogwB1NaturalStateOracle, effectstest.CardTarget(cheap))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the mana value 3 artifact survived")
	}
}

func TestOgw5545B1TarSnareShrinks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Tar Snare", "Instant", ogwB1TarSnareOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, ogre); p != 2 || tough != 3 {
		t.Errorf("the Ogre is %d/%d, want 2/3", p, tough)
	}
}

func TestOgw5545B1TearsOfValakutHitsOnlyAFlyer(t *testing.T) {
	if spec, _ := effects.Lookup(ogwB1TearsOfValakutOracle); !spec.CantBeCountered {
		t.Error("Tears of Valakut can be countered")
	}
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", ogwB1BearType, 5, 6)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 5, Toughness: 6, Keywords: []string{"flying"}})
	if err := effectstest.CastCatalogSpellErr(t, g, "Tears of Valakut", "Instant", ogwB1TearsOfValakutOracle, effectstest.CardTarget(walker)); err == nil {
		t.Fatal("Tears of Valakut targeted a creature without flying")
	}
	effectstest.CastCatalogSpell(t, g, "Tears of Valakut", "Instant", ogwB1TearsOfValakutOracle, effectstest.CardTarget(flyer))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, flyer).DamageMarked; got != 5 {
		t.Errorf("the flyer has %d damage, want 5", got)
	}
}

func TestOgw5545B1VoidShatterExilesTheSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := effectstest.B13HandCard(me, "Void Shatter", "Instant", ogwB1VoidShatterOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(bolt) || !g.Exile.Contains(bolt) {
		t.Error("the countered spell was not exiled")
	}
}

func TestOgw5545B1AlliedReinforcementsMakesTwoKnights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := ogwB1Active(t, g)
	effectstest.CastCatalogSpell(t, g, "Allied Reinforcements", "Sorcery", ogwB1AlliedReinforcementsOr, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Knight Ally"); got != 2 {
		t.Errorf("%d Knight Ally tokens, want 2", got)
	}
}

func TestOgw5545B1LeadByExampleSupportsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := ogwB1Active(t, g), effectstest.Top100Opponent(g)
	a := effectstest.B12Creature(g, me.ID, "A", ogwB1BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", ogwB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lead by Example", "Instant", ogwB1LeadByExampleOracle, append(effectstest.CardTarget(a), effectstest.CardTarget(b)...))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, a) != 1 || effectstest.PlusOnes(g, b) != 1 {
		t.Error("support did not put a counter on each target")
	}
}

func TestOgw5545B1SlipThroughSpaceDrawsAndMakesItUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", ogwB1BearType, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", ogwB1BearType, 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Slip Through Space", "Sorcery", ogwB1SlipThroughSpaceOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
	effectstest.BrAttack(t, g, bear)
	if err := g.DeclareBlocker(blocker, bear); err == nil {
		t.Error("the creature was blocked")
	}
}
