package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m15_5556_b1_cards_test.go — Magic 2015 batch 1 (#5556). Every test
// plays the real catalog entry through the engine.

const (
	m15B1AeronautTinkererOracle     = "a15bb807-ea31-4bff-b573-c249c4c78dd6"
	m15B1AmphinPathmageOracle       = "53e4273c-3292-47a0-90d1-8a0eb0eaf381"
	m15B1BelligerentSliverOracle    = "7d54a535-e924-47c7-81db-e04df1c3f44e"
	m15B1BloodHostOracle            = "ac193909-7e81-4082-ad02-2a60d142a0ab"
	m15B1BorderlandMarauderOracle   = "11fd2650-2150-4e62-8b64-e820d734a231"
	m15B1BrawlersPlateOracle        = "9249adf2-8607-424f-a73b-fcddaf124af9"
	m15B1MossBeastOracle            = "32f5f67b-bba3-427c-ab24-13a50acdc611"
	m15B1CarrionCrowOracle          = "2f9de785-a80a-4356-a4b4-0e0243e8ea2d"
	m15B1CovenantOfBloodOracle      = "20ca6a15-ba88-4efd-b5bd-da5938e9a43e"
	m15B1CrowdsFavorOracle          = "7a23915f-8bce-4aad-8395-0c736450a62d"
	m15B1RiverMarshalOracle         = "544c4cb3-cd43-4670-9883-e9545dad9cd6"
	m15B1FeralIncarnationOracle     = "31fefd48-6fff-41bf-a559-35a1699357ca"
	m15B1FestergloomOracle          = "a653ac26-e9c9-4319-88f6-c83cc64d3bdb"
	m15B1FleshToDustOracle          = "2f05f3ed-e3b7-4225-b3f2-c12eda335237"
	m15B1FrostLynxOracle            = "9400e348-b0c6-4aa3-b92c-1fbc8bc3baf9"
	m15B1GeistOfTheMoorsOracle      = "5c9d814b-4257-47bc-a1df-c3480bc038ba"
	m15B1InfernoFistOracle          = "0ccc75a0-1bc2-461f-ab6f-0f182328a863"
	m15B1JorubaiMurkLurkerOracle    = "12e091d0-3fd0-4058-84a9-529023b74c74"
	m15B1KirdChieftainOracle        = "3a7cca31-ccca-459b-9ac6-c8f384136ba5"
	m15B1KrenkosEnforcerOracle      = "0d9787e3-f482-4a98-95af-0438ed09c96c"
	m15B1MarkedByHonorOracle        = "aa5a7940-976d-4d96-9ddf-1c3891885c7e"
	m15B1MeditationPuzzleOracle     = "724c29a2-1783-401e-887c-2189afd5a944"
	m15B1MilitaryIntelOracle        = "464bfcdb-a82b-4573-92a4-c5e463724fbe"
	m15B1MinersBaneOracle           = "3ca75f61-60bf-487f-8cb3-b955cdd81e2a"
	m15B1NecromancersAssistOracle   = "2f72f111-450a-4da2-81fd-cbc4f9c1ef9d"
	m15B1NetcasterSpiderOracle      = "0ddb3fa0-447c-40c5-928a-541f3d543014"
	m15B1NightfireGiantOracle       = "2f351af6-0c94-4881-b91a-0211645c43c8"
	m15B1NimbusOfTheIslesOracle     = "3d529824-8e82-4c44-8e28-ed3419740c11"
	m15B1NissasExpeditionOracle     = "653010bc-45af-427e-a294-e0cc28d31821"
	m15B1ParagonEternalWildsOracle  = "942a461a-49aa-4eeb-b399-a2afba41ade9"
	m15B1ParagonFierceDefOracle     = "0f94d1a7-351f-4d37-8dab-1ebf2e8ee213"
	m15B1ParagonGatheringOracle     = "38212d24-40fd-404f-a8bb-24ed0d023fe8"
	m15B1ParagonNewDawnsOracle      = "f064e5e3-4772-4f6d-b8fa-4f36d4fcdb42"
	m15B1ParagonOpenGravesOracle    = "39375b5d-f509-4482-8d34-797dae17d38b"
	m15B1PillarOfLightOracle        = "643c1ee3-a653-4a6c-b705-a61f3b67e38c"
	m15B1ResearchAssistantOracle    = "30399525-ab58-4ed0-8d87-3decb0dd4862"
	m15B1SacredArmoryOracle         = "37d74df9-cf87-4b82-842a-3941064078fb"
	m15B1ScrapyardMongrelOracle     = "b6c8f2fe-b097-49af-ab59-16590dedaf32"
	m15B1ShadowcloakVampireOracle   = "00fc773a-72aa-4b58-a17c-8ac0f17c4c85"
	m15B1ShamanOfSpringOracle       = "b1ee0250-83ea-45a9-b5e5-3fa74bc7f26b"
	m15B1SliverHivelordOracle       = "e4fe33e7-b952-4152-8c83-81d948756d2f"
	m15B1SunbladeElfOracle          = "1b67e36a-6d74-41f7-b534-f7e3b6bf7017"
	m15B1TyrantsMachineOracle       = "d9cb3d86-cb3a-4e2e-a24e-ed3cfe5233d0"
	m15B1UnmakeTheGravesOracle      = "d9b99c58-8e2e-496c-8288-eda934f49162"
	m15B1VineweftOracle             = "f4a2d9eb-7f2b-4ee6-b38e-cf6e3a837b49"
	m15B1VoidSnareOracle            = "06e2ce8e-7341-4448-8625-cb1574eb179d"
	m15B1WillForgedGolemOracle      = "25fb0396-7355-4ff1-8ad3-d5a82adf6cc1"
	m15B1WitchsFamiliarOracle       = "abe06b4d-e7f0-4125-8cfc-e36296e8bacd"
	m15B1BearType                   = "Creature — Bear"
	m15B1SliverType                 = "Creature — Sliver"
	m15B1FirstStrike                = "first strike"
	m15B1Trample                    = "trample"
	m15B1BeastToken                 = "Beast"
	m15B1AnArtifactRock             = "Artifact"
	m15B1NonCreatureEnchantmentType = "Enchantment"
)

// m15B1Seats is the active player and the next one, from a main phase.
func m15B1Seats(t *testing.T, g *game.Game) (me, opp *game.Player) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	return effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
}

// --- every card is in the catalog ----------------------------------------

func TestM15B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		m15B1AeronautTinkererOracle, m15B1AmphinPathmageOracle, m15B1BelligerentSliverOracle, m15B1BloodHostOracle,
		m15B1BorderlandMarauderOracle, m15B1BrawlersPlateOracle, m15B1MossBeastOracle, m15B1CarrionCrowOracle,
		m15B1CovenantOfBloodOracle, m15B1CrowdsFavorOracle, m15B1RiverMarshalOracle, m15B1FeralIncarnationOracle,
		m15B1FestergloomOracle, m15B1FleshToDustOracle, m15B1FrostLynxOracle, m15B1GeistOfTheMoorsOracle,
		m15B1InfernoFistOracle, m15B1JorubaiMurkLurkerOracle, m15B1KirdChieftainOracle, m15B1KrenkosEnforcerOracle,
		m15B1MarkedByHonorOracle, m15B1MeditationPuzzleOracle, m15B1MilitaryIntelOracle, m15B1MinersBaneOracle,
		m15B1NecromancersAssistOracle, m15B1NetcasterSpiderOracle, m15B1NightfireGiantOracle, m15B1NimbusOfTheIslesOracle,
		m15B1NissasExpeditionOracle, m15B1ParagonEternalWildsOracle, m15B1ParagonFierceDefOracle, m15B1ParagonGatheringOracle,
		m15B1ParagonNewDawnsOracle, m15B1ParagonOpenGravesOracle, m15B1PillarOfLightOracle, m15B1ResearchAssistantOracle,
		m15B1SacredArmoryOracle, m15B1ScrapyardMongrelOracle, m15B1ShadowcloakVampireOracle, m15B1ShamanOfSpringOracle,
		m15B1SliverHivelordOracle, m15B1SunbladeElfOracle, m15B1TyrantsMachineOracle, m15B1UnmakeTheGravesOracle,
		m15B1VineweftOracle, m15B1VoidSnareOracle, m15B1WillForgedGolemOracle, m15B1WitchsFamiliarOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- printed keywords and convoke ------------------------------------------

func TestM15B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keyword                string
	}{
		{"Geist of the Moors", "Creature — Spirit", m15B1GeistOfTheMoorsOracle, "flying"},
		{"Nimbus of the Isles", "Creature — Elemental", m15B1NimbusOfTheIslesOracle, "flying"},
		{"Krenko's Enforcer", "Creature — Goblin Warrior", m15B1KrenkosEnforcerOracle, "intimidate"},
		{"Carrion Crow", "Creature — Zombie Bird", m15B1CarrionCrowOracle, "flying"},
		{"Netcaster Spider", "Creature — Spider", m15B1NetcasterSpiderOracle, "reach"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestM15B1ConvokeCardsDeclareConvoke(t *testing.T) {
	for _, oracle := range []string{
		m15B1WillForgedGolemOracle, m15B1CovenantOfBloodOracle, m15B1CrowdsFavorOracle, m15B1MeditationPuzzleOracle,
		m15B1FeralIncarnationOracle, m15B1NissasExpeditionOracle, m15B1UnmakeTheGravesOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not in the catalog", oracle)
		}
		if spec.TapCost == nil {
			t.Errorf("%s declares no convoke", spec.Name)
		}
	}
}

func TestM15B1WitchsFamiliarIsVanilla(t *testing.T) {
	spec, ok := effects.Lookup(m15B1WitchsFamiliarOracle)
	if !ok {
		t.Fatal("Witch's Familiar is not in the catalog")
	}
	if len(spec.Activated) != 0 || len(spec.Triggered) != 0 || len(spec.Static) != 0 {
		t.Error("Witch's Familiar has rules")
	}
}

func TestM15B1CarrionCrowEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	crow := effectstest.CastHoldCreature(t, g, "Carrion Crow", "Creature — Zombie Bird", m15B1CarrionCrowOracle, 2, 2)
	if !g.Battlefield.Contains(crow) {
		t.Fatal("Carrion Crow did not resolve")
	}
	if !effectstest.HmlTapped(t, g, crow) {
		t.Error("Carrion Crow entered untapped")
	}
}

// --- triggered abilities ----------------------------------------------------

func TestM15B1ShamanOfSpringDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Shaman of Spring", "Creature — Elf Shaman", m15B1ShamanOfSpringOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestM15B1NecromancersAssistantMillsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	library, graveyard := me.Library.Size(), me.Graveyard.Size()
	effectstest.CastHoldCreature(t, g, "Necromancer's Assistant", "Creature — Zombie", m15B1NecromancersAssistOracle, 3, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
	if got := me.Graveyard.Size() - graveyard; got != 3 {
		t.Errorf("%d cards reached the graveyard, want 3", got)
	}
}

func TestM15B1BorderlandMarauderPumpsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	id := effectstest.B12Push(g, me.ID, "Borderland Marauder", "Creature — Human Warrior", m15B1BorderlandMarauderOracle, 1, 2)
	effectstest.DeclareAttack(t, g, them.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 2 {
		t.Errorf("attacking, the Marauder is %d/%d, want 3/2", p, tough)
	}
}

func TestM15B1NetcasterSpiderPumpsBlockingAFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := effectstest.B12Push(g, opp.ID, "Netcaster Spider", "Creature — Spider", m15B1NetcasterSpiderOracle, 2, 3)
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Bird", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AttackIntoBlocks(t, g, opp.ID, bird)
	if err := g.DeclareBlocker(spider, bird); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, spider); p != 4 || tough != 3 {
		t.Errorf("blocking a flier, the Spider is %d/%d, want 4/3", p, tough)
	}
}

func TestM15B1FrostLynxTapsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", m15B1BearType, 2, 2)
	effectstest.CastHoldCreature(t, g, "Frost Lynx", "Creature — Elemental Cat", m15B1FrostLynxOracle, 2, 2)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Error("the opponent's creature is untapped")
	}
}

func TestM15B1MilitaryIntelligenceDrawsForTwoAttackers(t *testing.T) {
	for _, tc := range []struct {
		name      string
		attackers int
		draws     int
	}{
		{"two attackers", 2, 1},
		{"one attacker", 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := effectstest.SeatsForCombat(g)
			effectstest.PushCatalogPermanent(g, me.ID, "Military Intelligence", m15B1NonCreatureEnchantmentType, m15B1MilitaryIntelOracle, false)
			var attackers []uuid.UUID
			for i := 0; i < tc.attackers; i++ {
				attackers = append(attackers, effectstest.B12Creature(g, me.ID, "Raider", m15B1BearType, 2, 2))
			}
			effectstest.AdvanceToMain(t, g)
			hand := me.Hand.Size()
			effectstest.DeclareAttack(t, g, them.ID, attackers...)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("drew %d, want %d", got, tc.draws)
			}
		})
	}
}

// --- static abilities -------------------------------------------------------

func TestM15B1SliverLordsGrantOnlyYourSlivers(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Belligerent Sliver", m15B1SliverType, m15B1BelligerentSliverOracle, "menace"},
		{"Sliver Hivelord", "Legendary Creature — Sliver", m15B1SliverHivelordOracle, "indestructible"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			lord := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			mine := effectstest.B12Creature(g, me.ID, "My Sliver", m15B1SliverType, 1, 1)
			bear := effectstest.B12Creature(g, me.ID, "My Bear", m15B1BearType, 2, 2)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Sliver", m15B1SliverType, 1, 1)
			for _, id := range []uuid.UUID{lord, mine} {
				if !effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
					t.Errorf("a Sliver you control has no %s", tc.keyword)
				}
			}
			if effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("a non-Sliver has %s", tc.keyword)
			}
			if effectstest.HasEffectiveKeyword(t, g, theirs, tc.keyword) {
				t.Errorf("an opponent's Sliver has %s", tc.keyword)
			}
		})
	}
}

func TestM15B1AeronautTinkererFliesWithAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tinkerer := effectstest.B12Push(g, me.ID, "Aeronaut Tinkerer", "Creature — Human Artificer", m15B1AeronautTinkererOracle, 2, 3)
	if effectstest.HasEffectiveKeyword(t, g, tinkerer, "flying") {
		t.Fatal("the Tinkerer flies without an artifact")
	}
	effectstest.B12Permanent(g, me.ID, "Rock", m15B1AnArtifactRock)
	if !effectstest.HasEffectiveKeyword(t, g, tinkerer, "flying") {
		t.Error("the Tinkerer has no flying with an artifact")
	}
}

func TestM15B1ScrapyardMongrelWithAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dog := effectstest.B12Push(g, me.ID, "Scrapyard Mongrel", "Creature — Dog", m15B1ScrapyardMongrelOracle, 3, 3)
	if p, tough := effectstest.PtOf(t, g, dog); p != 3 || tough != 3 {
		t.Fatalf("without an artifact the Mongrel is %d/%d, want 3/3", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, dog, m15B1Trample) {
		t.Fatal("the Mongrel has trample without an artifact")
	}
	effectstest.B12Permanent(g, me.ID, "Rock", m15B1AnArtifactRock)
	if p, tough := effectstest.PtOf(t, g, dog); p != 5 || tough != 3 {
		t.Errorf("with an artifact the Mongrel is %d/%d, want 5/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dog, m15B1Trample) {
		t.Error("the Mongrel has no trample with an artifact")
	}
}

func TestM15B1LandGatedPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, land string
		p, tough                     int
	}{
		{"Dauntless River Marshal", "Creature — Human Soldier", m15B1RiverMarshalOracle, "Island", 2, 1},
		{"Jorubai Murk Lurker", "Creature — Leech", m15B1JorubaiMurkLurkerOracle, "Swamp", 1, 3},
		{"Kird Chieftain", "Creature — Ape", m15B1KirdChieftainOracle, "Forest", 3, 3},
		{"Nightfire Giant", "Creature — Zombie Giant", m15B1NightfireGiantOracle, "Mountain", 4, 3},
		{"Sunblade Elf", "Creature — Elf Warrior", m15B1SunbladeElfOracle, "Plains", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			effectstest.B12Permanent(g, opp.ID, tc.land, "Basic Land — "+tc.land)
			if p, tough := effectstest.PtOf(t, g, id); p != tc.p || tough != tc.tough {
				t.Fatalf("with only an opponent's %s it is %d/%d, want %d/%d", tc.land, p, tough, tc.p, tc.tough)
			}
			effectstest.B12Permanent(g, me.ID, tc.land, "Basic Land — "+tc.land)
			if p, tough := effectstest.PtOf(t, g, id); p != tc.p+1 || tough != tc.tough+1 {
				t.Errorf("with your %s it is %d/%d, want %d/%d", tc.land, p, tough, tc.p+1, tc.tough+1)
			}
		})
	}
}

// --- the Paragons -------------------------------------------------------------

func TestM15B1Paragons(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, color, other, keyword string
		mana                                []string
	}{
		{"Paragon of Eternal Wilds", m15B1ParagonEternalWildsOracle, "G", "R", m15B1Trample, []string{"G"}},
		{"Paragon of Fierce Defiance", m15B1ParagonFierceDefOracle, "R", "G", "haste", []string{"R"}},
		{"Paragon of Gathering Mists", m15B1ParagonGatheringOracle, "U", "W", "flying", []string{"U"}},
		{"Paragon of New Dawns", m15B1ParagonNewDawnsOracle, "W", "U", "vigilance", []string{"W"}},
		{"Paragon of Open Graves", m15B1ParagonOpenGravesOracle, "B", "W", "deathtouch", []string{"C", "C", "B"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := m15B1Seats(t, g)
			paragon := effectstest.B12Push(g, me.ID, tc.name, "Creature — Human", tc.oracle, 2, 2)
			effectstest.HmlSetColors(g, paragon, tc.color)
			same := effectstest.B12Creature(g, me.ID, "Same", m15B1BearType, 2, 2)
			effectstest.HmlSetColors(g, same, tc.color)
			other := effectstest.B12Creature(g, me.ID, "Other", m15B1BearType, 2, 2)
			effectstest.HmlSetColors(g, other, tc.other)
			theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m15B1BearType, 2, 2)
			effectstest.HmlSetColors(g, theirs, tc.color)
			if p, tough := effectstest.PtOf(t, g, same); p != 3 || tough != 3 {
				t.Errorf("your creature of the colour is %d/%d, want 3/3", p, tough)
			}
			for _, id := range []uuid.UUID{paragon, other, theirs} {
				if p, tough := effectstest.PtOf(t, g, id); p != 2 || tough != 2 {
					t.Errorf("a creature outside the anthem is %d/%d, want 2/2", p, tough)
				}
			}
			effectstest.ApaMana(me, tc.mana...)
			for _, bad := range []uuid.UUID{paragon, other, theirs} {
				if err := g.ActivateCatalogAbility(me.ID, paragon, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bad)}); err == nil {
					t.Fatal("the Paragon targeted a creature that isn't another one of its colour you control")
				}
			}
			effectstest.ApaActivate(t, g, me, paragon, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(same)})
			if !effectstest.HasEffectiveKeyword(t, g, same, tc.keyword) {
				t.Errorf("the target has no %s", tc.keyword)
			}
			if !effectstest.HmlTapped(t, g, paragon) {
				t.Error("the Paragon did not tap")
			}
		})
	}
}

// --- activated abilities -------------------------------------------------------

func TestM15B1AmphinPathmageMakesACreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m15B1Seats(t, g)
	mage := effectstest.B12Push(g, me.ID, "Amphin Pathmage", "Creature — Salamander Wizard", m15B1AmphinPathmageOracle, 3, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m15B1BearType, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", m15B1BearType, 2, 2)
	effectstest.ApaMana(me, "C", "C", "U")
	effectstest.ApaActivate(t, g, me, mage, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.AttackIntoBlocks(t, g, opp.ID, bear)
	if err := g.DeclareBlocker(blocker, bear); err == nil {
		t.Error("a creature blocked the unblockable target")
	}
}

func TestM15B1BloodHostSacrificesAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	host := effectstest.B12Push(g, me.ID, "Blood Host", "Creature — Vampire", m15B1BloodHostOracle, 3, 3)
	victim := effectstest.B12Creature(g, me.ID, "Victim", m15B1BearType, 2, 2)
	effectstest.ApaMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, host, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{host}}); err == nil {
		t.Fatal("the Host sacrificed itself")
	}
	life := me.Life
	effectstest.ApaActivate(t, g, me, host, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{victim}})
	if g.Battlefield.Contains(victim) {
		t.Error("the other creature was not sacrificed")
	}
	if got := effectstest.PlusOneCounters(g, host); got != 1 {
		t.Errorf("the Host has %d +1/+1 counters, want 1", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestM15B1CarnivorousMossBeastGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	beast := effectstest.B12Push(g, me.ID, "Carnivorous Moss-Beast", "Creature — Plant Elemental Beast", m15B1MossBeastOracle, 4, 5)
	effectstest.ApaMana(me, "C", "C", "C", "C", "C", "G", "G")
	effectstest.ApaActivate(t, g, me, beast, 0, game.ActivateAbilityParams{})
	if got := effectstest.PlusOneCounters(g, beast); got != 1 {
		t.Errorf("the Moss-Beast has %d +1/+1 counters, want 1", got)
	}
}

func TestM15B1MinersBanePumpsAndTramples(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	bane := effectstest.B12Push(g, me.ID, "Miner's Bane", "Creature — Elemental", m15B1MinersBaneOracle, 6, 3)
	effectstest.ApaMana(me, "C", "C", "R")
	effectstest.ApaActivate(t, g, me, bane, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, bane); p != 7 || tough != 3 {
		t.Errorf("the Bane is %d/%d, want 7/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bane, m15B1Trample) {
		t.Error("the Bane has no trample")
	}
}

func TestM15B1ResearchAssistantLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	assistant := effectstest.B12Push(g, me.ID, "Research Assistant", "Creature — Human Wizard", m15B1ResearchAssistantOracle, 1, 3)
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "C", "C", "C", "U")
	effectstest.ApaActivate(t, g, me, assistant, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d before the discard, want %d", got, hand+1)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
	if !effectstest.HmlTapped(t, g, assistant) {
		t.Error("the Assistant did not tap")
	}
}

func TestM15B1ShadowcloakVampireFliesForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	vampire := effectstest.B12Push(g, me.ID, "Shadowcloak Vampire", "Creature — Vampire", m15B1ShadowcloakVampireOracle, 4, 3)
	life := me.Life
	effectstest.ApaActivate(t, g, me, vampire, 0, game.ActivateAbilityParams{})
	if me.Life != life-2 {
		t.Errorf("life %d after paying, want %d", me.Life, life-2)
	}
	if !effectstest.HasEffectiveKeyword(t, g, vampire, "flying") {
		t.Error("the Vampire has no flying")
	}
}

func TestM15B1TappersTapTheTarget(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		mana                   []string
	}{
		{"Dauntless River Marshal", "Creature — Human Soldier", m15B1RiverMarshalOracle, []string{"C", "C", "C", "U"}},
		{"Tyrant's Machine", m15B1AnArtifactRock, m15B1TyrantsMachineOracle, []string{"C", "C", "C", "C"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := m15B1Seats(t, g)
			source := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 1)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", m15B1BearType, 2, 2)
			effectstest.ApaMana(me, tc.mana...)
			effectstest.ApaActivate(t, g, me, source, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(theirs)})
			if !effectstest.HmlTapped(t, g, theirs) {
				t.Error("the target is untapped")
			}
		})
	}
}

func TestM15B1JorubaiMurkLurkerGrantsLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	leech := effectstest.B12Push(g, me.ID, "Jorubai Murk Lurker", "Creature — Leech", m15B1JorubaiMurkLurkerOracle, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m15B1BearType, 2, 2)
	effectstest.ApaMana(me, "C", "B")
	effectstest.ApaActivate(t, g, me, leech, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the target has no lifelink")
	}
}

func TestM15B1TargetPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		mana                   []string
		p, tough               int
		trample                bool
	}{
		{"Kird Chieftain", "Creature — Ape", m15B1KirdChieftainOracle, []string{"C", "C", "C", "C", "G"}, 4, 4, true},
		{"Sacred Armory", m15B1AnArtifactRock, m15B1SacredArmoryOracle, []string{"C", "C"}, 3, 2, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := m15B1Seats(t, g)
			source := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
			bear := effectstest.B12Creature(g, me.ID, "Bear", m15B1BearType, 2, 2)
			effectstest.ApaMana(me, tc.mana...)
			effectstest.ApaActivate(t, g, me, source, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
			if p, tough := effectstest.PtOf(t, g, bear); p != tc.p || tough != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
			if got := effectstest.HasEffectiveKeyword(t, g, bear, m15B1Trample); got != tc.trample {
				t.Errorf("trample = %v, want %v", got, tc.trample)
			}
		})
	}
}

func TestM15B1NightfireGiantShootsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m15B1Seats(t, g)
	giant := effectstest.B12Push(g, me.ID, "Nightfire Giant", "Creature — Zombie Giant", m15B1NightfireGiantOracle, 4, 3)
	life := opp.Life
	effectstest.ApaMana(me, "C", "C", "C", "C", "R")
	effectstest.ApaActivate(t, g, me, giant, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
}

func TestM15B1SunbladeElfPumpsYourTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m15B1Seats(t, g)
	elf := effectstest.B12Push(g, me.ID, "Sunblade Elf", "Creature — Elf Warrior", m15B1SunbladeElfOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m15B1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m15B1BearType, 2, 2)
	effectstest.ApaMana(me, "C", "C", "C", "C", "W")
	effectstest.ApaActivate(t, g, me, elf, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("your Bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the opponent's Bear is %d/%d, want 2/2", p, tough)
	}
}

// --- Equipment and Auras ----------------------------------------------------

func TestM15B1BrawlersPlateEquips(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m15B1BearType, 2, 2)
	plate := effectstest.SeedEquipment(g, me.ID, "Brawler's Plate", m15B1BrawlersPlateOracle)
	if err := effectstest.StrictEquip(t, g, me.ID, plate, bear, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("equip {4}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the equipped Bear is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, m15B1Trample) {
		t.Error("the equipped Bear has no trample")
	}
}

func TestM15B1MarkedByHonorPumpsAndGrantsVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Marked by Honor", m15B1MarkedByHonorOracle, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the host is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the host has no vigilance")
	}
}

func TestM15B1VineweftPumpsAndReturnsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Vineweft", m15B1VineweftOracle, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the host is %d/%d, want 3/3", p, tough)
	}
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Vineweft", TypeLine: effectstest.AuraTypeLine, OracleID: m15B1VineweftOracle,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.FloatForTest(g, me, "CCCCG")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{4}{G}: return this card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("Vineweft did not return to hand")
	}
}

func TestM15B1InfernoFistEnchantsYourCreatureAndShoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m15B1Seats(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m15B1BearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Inferno Fist", effectstest.AuraTypeLine, m15B1InfernoFistOracle, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("Inferno Fist enchanted a creature you don't control")
	}
	bear := effectstest.AuraBear(g, me.ID)
	fist := effectstest.AuraCast(t, g, "Inferno Fist", m15B1InfernoFistOracle, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the host is %d/%d, want 4/2", p, tough)
	}
	life := opp.Life
	effectstest.ApaMana(me, "R")
	effectstest.ApaActivate(t, g, me, fist, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
	if g.Battlefield.Contains(fist) {
		t.Error("the Aura was not sacrificed")
	}
}

// --- the spells -------------------------------------------------------------------

func TestM15B1VoidSnareBouncesANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := m15B1Seats(t, g)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErr(t, g, "Void Snare", "Sorcery", m15B1VoidSnareOracle, effectstest.CardTarget(land)); err == nil {
		t.Fatal("Void Snare targeted a land")
	}
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m15B1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Void Snare", "Sorcery", m15B1VoidSnareOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestM15B1FleshToDustDestroysPastRegeneration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := m15B1Seats(t, g)
	boa := effectstest.VisShield(t, g, opp, 2, 3)
	effectstest.CastCatalogSpell(t, g, "Flesh to Dust", "Instant", m15B1FleshToDustOracle, effectstest.CardTarget(boa))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(boa) {
		t.Error("a regeneration shield saved the target")
	}
}

func TestM15B1FestergloomSparesBlackCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m15B1Seats(t, g)
	black := effectstest.B12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	effectstest.HmlSetColors(g, black, "B")
	green := effectstest.B12Creature(g, opp.ID, "Bear", m15B1BearType, 2, 2)
	effectstest.HmlSetColors(g, green, "G")
	effectstest.CastCatalogSpell(t, g, "Festergloom", "Sorcery", m15B1FestergloomOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, black); p != 2 || tough != 2 {
		t.Errorf("the black creature is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, green); p != 1 || tough != 1 {
		t.Errorf("the green creature is %d/%d, want 1/1", p, tough)
	}
}

func TestM15B1PillarOfLightExilesOnlyToughnessFourOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := m15B1Seats(t, g)
	small := effectstest.B12Creature(g, opp.ID, "Small", m15B1BearType, 5, 3)
	if err := effectstest.CastCatalogSpellErr(t, g, "Pillar of Light", "Instant", m15B1PillarOfLightOracle, effectstest.CardTarget(small)); err == nil {
		t.Fatal("Pillar of Light targeted a creature with toughness 3")
	}
	big := effectstest.B12Creature(g, opp.ID, "Big", m15B1BearType, 1, 4)
	effectstest.CastCatalogSpell(t, g, "Pillar of Light", "Instant", m15B1PillarOfLightOracle, effectstest.CardTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(big) {
		t.Error("the creature was not exiled")
	}
}

func TestM15B1CovenantOfBloodDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m15B1Seats(t, g)
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Covenant of Blood", "Sorcery", m15B1CovenantOfBloodOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-4 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, theirs-4)
	}
	if me.Life != mine+4 {
		t.Errorf("life %d, want %d", me.Life, mine+4)
	}
}

func TestM15B1CrowdsFavorPumpsAndGrantsFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m15B1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Crowd's Favor", "Instant", m15B1CrowdsFavorOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, m15B1FirstStrike) {
		t.Error("the Bear has no first strike")
	}
}

func TestM15B1MeditationPuzzleGainsEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Meditation Puzzle", "Instant", m15B1MeditationPuzzleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+8 {
		t.Errorf("life %d, want %d", me.Life, life+8)
	}
}

func TestM15B1FeralIncarnationMakesThreeBeasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	before := effectstest.CountBattlefieldNamed(g, me.ID, m15B1BeastToken)
	effectstest.CastCatalogSpell(t, g, "Feral Incarnation", "Sorcery", m15B1FeralIncarnationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, m15B1BeastToken) - before; got != 3 {
		t.Errorf("%d Beasts were created, want 3", got)
	}
}

func TestM15B1NissasExpeditionSearchesForUpToTwoBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	for i := 0; i < 3; i++ {
		effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	}
	effectstest.CastCatalogSpell(t, g, "Nissa's Expedition", "Sorcery", m15B1NissasExpeditionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || c.SearchMax != 2 {
		t.Fatalf("the search is for up to two, got %+v", c)
	}
}

func TestM15B1UnmakeTheGravesReturnsTwoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m15B1Seats(t, g)
	rock := effectstest.WthGraveyardCard(me, "Rock", m15B1AnArtifactRock)
	if err := effectstest.CastCatalogSpellErr(t, g, "Unmake the Graves", "Instant", m15B1UnmakeTheGravesOracle, effectstest.CardTarget(rock)); err == nil {
		t.Fatal("Unmake the Graves targeted a noncreature card")
	}
	one := effectstest.WthGraveyardCard(me, "One", m15B1BearType)
	two := effectstest.WthGraveyardCard(me, "Two", m15B1BearType)
	effectstest.CastCatalogSpell(t, g, "Unmake the Graves", "Instant", m15B1UnmakeTheGravesOracle, effectstest.CardTargets(one, two))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{one, two} {
		if !me.Hand.Contains(id) || me.Graveyard.Contains(id) {
			t.Error("a creature card did not return to hand")
		}
	}
}
