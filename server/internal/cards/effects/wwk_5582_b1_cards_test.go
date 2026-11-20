package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wwk_5582_b1_cards_test.go — Worldwake batch 1 (#5582). Every test plays
// the real catalog entry through the engine.

const (
	wwkBattleHurdaOracle        = "46762944-8364-45c4-8008-e693e97ddc46"
	wwkGoblinRoughriderOracle   = "a920d37e-e172-41f7-bfbf-c18821dcbed7"
	wwkGoliathSphinxOracle      = "73115500-4468-43b3-9151-13e3da9050c5"
	wwkGrapplerSpiderOracle     = "0ae0c880-a508-471f-a90b-029c417f6b3b"
	wwkJagwaspSwarmOracle       = "376d90f6-2ceb-4d9e-a04d-3dd5bafd1e8b"
	wwkLeatherbackBalothOracle  = "0ee14128-3bec-4b65-8ee4-619337d4ed45"
	wwkMarshThreaderOracle      = "7b2eb6f2-bd13-4eb6-b91a-c0fa8f3d96af"
	wwkHorizonDrakeOracle       = "9c724912-5afd-4813-affb-4dbda64fd4ee"
	wwkApexHawksOracle          = "c025d558-aae8-4146-923b-a98035025a42"
	wwkEnclaveEliteOracle       = "27c84b2e-fc66-45a6-a5ad-d3ef204d5077"
	wwkGnarlidPackOracle        = "be6e835c-822e-45ff-a992-249d4d9781dc"
	wwkQuagVampiresOracle       = "1d28ed9a-cc4e-48d9-a1a6-7504fc6cc56d"
	wwkSkitterOfLizardsOracle   = "14646c85-12ad-4b99-a2af-e4a790e034ef"
	wwkLightkeeperOracle        = "7f31aa43-cf75-4488-b47f-e16e923a2d01"
	wwkDeathforgeShamanOracle   = "1d3cfdd3-e16f-4955-9b81-0d6fc981c168"
	wwkBloodhuskRitualistOracle = "93e1d3a6-e12f-4a0a-be0f-fbf568d89f1a"
	wwkHedronRoverOracle        = "1b2826ea-a3d5-4b8a-b362-0b0880d1cc48"
	wwkFledglingGriffinOracle   = "314748a6-4e7f-4afb-b77e-5ef366e02b90"
	wwkSnappingCreeperOracle    = "bb4e1340-da63-46b8-a5e3-79dfedec3e78"
	wwkCausticCrawlerOracle     = "91c83b1e-32d9-4bb4-9a71-05fbb0cb8134"
	wwkCosisRavagerOracle       = "1253de62-6218-401c-9d98-a5ec8617e0d8"
	wwkHadaFreebladeOracle      = "e9ad9df4-2fd6-4a22-b861-0a8d28f2fe10"
	wwkGraypeltHunterOracle     = "b0553bf6-ecc1-4d08-9777-19827658355b"
	wwkBojukaBrigandOracle      = "a1d82070-3198-4ead-9431-d90661e29741"
	wwkHalimarExcavatorOracle   = "fd3e37c9-93bf-4f3e-a279-22afbffd8d43"
	wwkAkoumBattlesingerOracle  = "eb448589-28e1-49d4-ba3e-7ddb4341fdc5"
	wwkBullRushOracle           = "0140cf53-71ef-4d37-9197-48aab35ca4a2"
	wwkIonasJudgmentOracle      = "abbfbe50-4caf-48d0-b8f1-f2c9212bcd2f"
	wwkVeteransReflexesOracle   = "efee0ca0-0897-4972-99c4-432561b6e9f8"
	wwkJoinTheRanksOracle       = "aa16307a-9c1f-4538-872c-95206e2b7a6e"
	wwkAetherTradewindsOracle   = "3a0dd318-9523-4740-b0aa-a911032e3851"
	wwkTombHexOracle            = "4b3bc59c-5439-4ec8-b25e-7493fa1cd3fd"
	wwkRestForTheWearyOracle    = "c7f08962-6c54-4fd5-a55e-7513e758af32"
	wwkMysteriesOfTheDeepOracle = "3708eaae-9306-46b2-ae41-7ce6c50be3c7"
	wwkLoamLionOracle           = "18961e25-3feb-4dba-a77f-240a4d664ecb"
	wwkRuthlessCullbladeOracle  = "b7748ad6-a8cd-45e7-a4d6-4ee747722657"
	wwkSummitApesOracle         = "01799149-4bca-417b-9e89-14493fd46c79"
	wwkSejiriMerfolkOracle      = "74231a46-7e2d-497f-ae66-b40f8989551c"
	wwkKitesailApprenticeOracle = "c2157f2c-a3a4-47aa-b33e-b2c65e11c2e5"
	wwkClawsOfValakutOracle     = "94b86278-bfa6-4be4-8114-09b6c6c85724"
	wwkCunningSparkmageOracle   = "ab4a25d8-e1d0-454e-9d06-f7960a0352bc"
	wwkSurrakarBanisherOracle   = "d48fd4b7-b60c-4932-abed-a8df1184db25"
	wwkPilgrimsEyeOracle        = "66155010-cc2c-449a-85d5-92c95aade514"
	wwkKorFirewalkerOracle      = "fb37ff1c-7461-4a16-b27f-6eef0db937fe"
	wwkNovablastWurmOracle      = "c3aa2554-1cea-45e7-8fc7-1c09bd82fadb"
	wwkGrotagThrasherOracle     = "fcfe4750-5311-45d2-8d8e-7b71e69368a2"
	wwkSmolderingSpiresOracle   = "cfa3288d-e521-4a13-bcb3-7950a94e1746"
	wwkBearType                 = "Creature — Bear"
	wwkAllyFriendType           = "Creature — Kor Ally"
	wwkFirstStrike              = "first strike"
)

func TestWwk5582EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		wwkBattleHurdaOracle: "Battle Hurda", wwkGoblinRoughriderOracle: "Goblin Roughrider",
		wwkGoliathSphinxOracle: "Goliath Sphinx", wwkGrapplerSpiderOracle: "Grappler Spider",
		wwkJagwaspSwarmOracle: "Jagwasp Swarm", wwkLeatherbackBalothOracle: "Leatherback Baloth",
		wwkMarshThreaderOracle: "Marsh Threader", wwkHorizonDrakeOracle: "Horizon Drake",
		wwkApexHawksOracle: "Apex Hawks", wwkEnclaveEliteOracle: "Enclave Elite",
		wwkGnarlidPackOracle: "Gnarlid Pack", wwkQuagVampiresOracle: "Quag Vampires",
		wwkSkitterOfLizardsOracle: "Skitter of Lizards", wwkLightkeeperOracle: "Lightkeeper of Emeria",
		wwkDeathforgeShamanOracle: "Deathforge Shaman", wwkBloodhuskRitualistOracle: "Bloodhusk Ritualist",
		wwkHedronRoverOracle: "Hedron Rover", wwkFledglingGriffinOracle: "Fledgling Griffin",
		wwkSnappingCreeperOracle: "Snapping Creeper", wwkCausticCrawlerOracle: "Caustic Crawler",
		wwkCosisRavagerOracle: "Cosi's Ravager", wwkHadaFreebladeOracle: "Hada Freeblade",
		wwkGraypeltHunterOracle: "Graypelt Hunter", wwkBojukaBrigandOracle: "Bojuka Brigand",
		wwkHalimarExcavatorOracle: "Halimar Excavator", wwkAkoumBattlesingerOracle: "Akoum Battlesinger",
		wwkBullRushOracle: "Bull Rush", wwkIonasJudgmentOracle: "Iona's Judgment",
		wwkVeteransReflexesOracle: "Veteran's Reflexes", wwkJoinTheRanksOracle: "Join the Ranks",
		wwkAetherTradewindsOracle: "Aether Tradewinds", wwkTombHexOracle: "Tomb Hex",
		wwkRestForTheWearyOracle: "Rest for the Weary", wwkMysteriesOfTheDeepOracle: "Mysteries of the Deep",
		wwkLoamLionOracle: "Loam Lion", wwkRuthlessCullbladeOracle: "Ruthless Cullblade",
		wwkSummitApesOracle: "Summit Apes", wwkSejiriMerfolkOracle: "Sejiri Merfolk",
		wwkKitesailApprenticeOracle: "Kitesail Apprentice", wwkClawsOfValakutOracle: "Claws of Valakut",
		wwkCunningSparkmageOracle: "Cunning Sparkmage", wwkSurrakarBanisherOracle: "Surrakar Banisher",
		wwkPilgrimsEyeOracle: "Pilgrim's Eye", wwkKorFirewalkerOracle: "Kor Firewalker",
		wwkNovablastWurmOracle: "Novablast Wurm", wwkGrotagThrasherOracle: "Grotag Thrasher",
		wwkSmolderingSpiresOracle: "Smoldering Spires",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered as %q, %v with %d caveats; want its name, Full and none", name, spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- shared test helpers ---------------------------------------------------------

// wwkCast casts a creature card with a printed body from the active
// seat's hand in a main phase and returns its ID. It does not pass
// priority.
func wwkCast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := top100ActiveSeat(g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// wwkPickTarget passes priority until chooser has a pick_target prompt
// open, then answers it with ref.
func wwkPickTarget(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef) {
	t.Helper()
	for i := 0; i < 12 && latestPickTarget(g, chooser) == nil; i++ {
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

func wwkCardRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// wwkLand seeds a land with the given basic subtype under owner.
func wwkLand(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return b12Permanent(g, owner, subtype, "Basic Land — "+subtype)
}

// wwkSetTapped taps a battlefield permanent directly.
func wwkSetTapped(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Tapped = true })
}

// --- printed keywords --------------------------------------------------------------

func TestWwk5582KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Battle Hurda", "Creature — Giant", wwkBattleHurdaOracle, []string{wwkFirstStrike}},
		{"Goliath Sphinx", "Creature — Sphinx", wwkGoliathSphinxOracle, []string{"flying"}},
		{"Grappler Spider", "Creature — Spider", wwkGrapplerSpiderOracle, []string{"reach"}},
		{"Jagwasp Swarm", "Creature — Insect", wwkJagwaspSwarmOracle, []string{"flying"}},
		{"Marsh Threader", "Creature — Kor Scout", wwkMarshThreaderOracle, []string{"swampwalk"}},
		{"Horizon Drake", "Creature — Drake", wwkHorizonDrakeOracle, []string{"flying", "protection from lands"}},
		{"Apex Hawks", "Creature — Bird", wwkApexHawksOracle, []string{"flying"}},
		{"Enclave Elite", "Creature — Merfolk Soldier", wwkEnclaveEliteOracle, []string{"islandwalk"}},
		{"Quag Vampires", "Creature — Vampire Rogue", wwkQuagVampiresOracle, []string{"swampwalk"}},
		{"Skitter of Lizards", "Creature — Lizard", wwkSkitterOfLizardsOracle, []string{"haste"}},
		{"Lightkeeper of Emeria", "Creature — Angel", wwkLightkeeperOracle, []string{"flying"}},
		{"Graypelt Hunter", "Creature — Human Warrior Ally", wwkGraypeltHunterOracle, []string{"trample"}},
		{"Akoum Battlesinger", "Creature — Human Berserker Ally", wwkAkoumBattlesingerOracle, []string{"haste"}},
		{"Cunning Sparkmage", "Creature — Human Shaman", wwkCunningSparkmageOracle, []string{"haste"}},
		{"Pilgrim's Eye", "Artifact Creature — Thopter", wwkPilgrimsEyeOracle, []string{"flying"}},
		{"Kor Firewalker", "Creature — Kor Soldier", wwkKorFirewalkerOracle, []string{"protection from red"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s (abilities %v)", tc.name, kw, abilities)
				}
			}
		})
	}
}

// --- multikicker ---------------------------------------------------------------------

func TestWwk5582MultikickerCreaturesEnterWithACounterPerKick(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Gnarlid Pack", "Creature — Beast", wwkGnarlidPackOracle, 2, 2},
		{"Apex Hawks", "Creature — Bird", wwkApexHawksOracle, 2, 2},
		{"Enclave Elite", "Creature — Merfolk Soldier", wwkEnclaveEliteOracle, 2, 2},
		{"Quag Vampires", "Creature — Vampire Rogue", wwkQuagVampiresOracle, 1, 1},
		{"Skitter of Lizards", "Creature — Lizard", wwkSkitterOfLizardsOracle, 1, 1},
	} {
		for _, kicks := range []int{0, 2} {
			t.Run(tc.name, func(t *testing.T) {
				g := newCatalogGame(t)
				params := game.CastSpellParams{}
				for i := 0; i < kicks; i++ {
					params.OptionalCosts = append(params.OptionalCosts, 0)
				}
				id := wwkCast(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, params)
				passPriorityAroundTable(t, g)
				if !g.Battlefield.Contains(id) {
					t.Fatalf("%s did not resolve", tc.name)
				}
				if got := b12Counter(t, g, id, game.CounterPlusOne); got != kicks {
					t.Errorf("kicked %d times: %d +1/+1 counters", kicks, got)
				}
			})
		}
	}
}

func TestWwk5582LightkeeperGainsTwoLifePerKick(t *testing.T) {
	for _, kicks := range []int{0, 2} {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		life := me.Life
		params := game.CastSpellParams{}
		for i := 0; i < kicks; i++ {
			params.OptionalCosts = append(params.OptionalCosts, 0)
		}
		wwkCast(t, g, "Lightkeeper of Emeria", "Creature — Angel", wwkLightkeeperOracle, 2, 4, params)
		passPriorityAroundTable(t, g)
		passPriorityAroundTable(t, g)
		if me.Life != life+2*kicks {
			t.Errorf("kicked %d times: life %d, want %d", kicks, me.Life, life+2*kicks)
		}
	}
}

func TestWwk5582DeathforgeShamanDealsTwicePerKick(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	life := opp.Life
	wwkCast(t, g, "Deathforge Shaman", "Creature — Ogre Shaman", wwkDeathforgeShamanOracle, 4, 3,
		game.CastSpellParams{OptionalCosts: []int{0, 0}})
	passPriorityAroundTable(t, g)
	wwkPickTarget(t, g, me.ID, playerTarget(opp.ID))
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 4 {
		t.Errorf("the player took %d damage, want 4", got)
	}
}

func TestWwk5582BloodhuskRitualistDiscardsPerKick(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	d2dSetHand(opp, 3)
	wwkCast(t, g, "Bloodhusk Ritualist", "Creature — Vampire Shaman", wwkBloodhuskRitualistOracle, 2, 2,
		game.CastSpellParams{OptionalCosts: []int{0, 0}})
	passPriorityAroundTable(t, g)
	wwkPickTarget(t, g, me.ID, playerTarget(opp.ID))
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 2 {
		t.Errorf("the opponent owes %d discards, want 2", n)
	}
}

// --- landfall ---------------------------------------------------------------------------

func TestWwk5582HedronRoverGrowsOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	rover := b12Push(g, me.ID, "Hedron Rover", "Artifact Creature — Construct", wwkHedronRoverOracle, 2, 2)
	playLandFromHand(t, g, "Plains", "")
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, rover); p != 4 || tough != 4 {
		t.Errorf("the Rover is %d/%d, want 4/4", p, tough)
	}
}

func TestWwk5582LandfallKeywordGrants(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Fledgling Griffin", "Creature — Griffin", wwkFledglingGriffinOracle, "flying"},
		{"Snapping Creeper", "Creature — Plant", wwkSnappingCreeperOracle, "vigilance"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("%s has %s before any land entered", tc.name, tc.keyword)
			}
			playLandFromHand(t, g, "Forest", "")
			passPriorityAroundTable(t, g)
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s has no %s after landfall", tc.name, tc.keyword)
			}
		})
	}
}

func TestWwk5582CausticCrawlerShrinksATargetOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	b12Push(g, me.ID, "Caustic Crawler", "Creature — Insect", wwkCausticCrawlerOracle, 4, 3)
	bear := b12Creature(g, opp.ID, "Bear", wwkBearType, 2, 2)
	playLandFromHand(t, g, "Swamp", "")
	wwkPickTarget(t, g, me.ID, wwkCardRef(bear))
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 1 || tough != 1 {
		t.Errorf("the Bear is %d/%d, want 1/1", p, tough)
	}
}

func TestWwk5582CosisRavagerPingsAPlayerOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	b12Push(g, me.ID, "Cosi's Ravager", "Creature — Elemental", wwkCosisRavagerOracle, 2, 2)
	life := opp.Life
	playLandFromHand(t, g, "Mountain", "")
	wwkPickTarget(t, g, me.ID, playerTarget(opp.ID))
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-1)
	}
}

// --- Allies -------------------------------------------------------------------------------

func TestWwk5582AllyCounterOnItsOwnAndAnotherAllysEntry(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Hada Freeblade", "Creature — Human Soldier Ally", wwkHadaFreebladeOracle, 0, 1},
		{"Graypelt Hunter", "Creature — Human Warrior Ally", wwkGraypeltHunterOracle, 2, 2},
		{"Bojuka Brigand", "Creature — Human Warrior Ally", wwkBojukaBrigandOracle, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := top100ActiveSeat(g)
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			answerLatestTriggerPrompt(t, g, me.ID, true)
			passPriorityAroundTable(t, g)
			if got := b12Counter(t, g, id, game.CounterPlusOne); got != 1 {
				t.Fatalf("after its own entry: %d counters, want 1", got)
			}
			castHoldCreature(t, g, "Bear", wwkBearType, "", 2, 2)
			if n := owedMays(g, me.ID); n != 0 {
				t.Fatalf("a non-Ally entering asked %d questions", n)
			}
			castHoldCreature(t, g, "Kor Friend", wwkAllyFriendType, "", 1, 1)
			answerLatestTriggerPrompt(t, g, me.ID, true)
			passPriorityAroundTable(t, g)
			if got := b12Counter(t, g, id, game.CounterPlusOne); got != 2 {
				t.Errorf("after another Ally's entry: %d counters, want 2", got)
			}
		})
	}
}

func TestWwk5582BojukaBrigandCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Raider", wwkBearType, 2, 2)
	brigand := b12Push(g, opp.ID, "Bojuka Brigand", "Creature — Human Warrior Ally", wwkBojukaBrigandOracle, 1, 1)
	bear := b12Creature(g, opp.ID, "Bear", wwkBearType, 2, 2)
	brAttack(t, g, attacker)
	if err := g.DeclareBlocker(brigand, attacker); err == nil {
		t.Error("Bojuka Brigand blocked, but it can't block")
	}
	if err := g.DeclareBlocker(bear, attacker); err != nil {
		t.Errorf("an ordinary creature can't block: %v", err)
	}
}

func TestWwk5582HalimarExcavatorMillsPerAlly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	b12Push(g, me.ID, "Halimar Excavator", "Creature — Human Wizard Ally", wwkHalimarExcavatorOracle, 1, 3)
	lib, gy := opp.Library.Size(), opp.Graveyard.Size()
	castHoldCreature(t, g, "Kor Friend", wwkAllyFriendType, "", 1, 1)
	wwkPickTarget(t, g, me.ID, playerTarget(opp.ID))
	passPriorityAroundTable(t, g)
	if got := lib - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2 (two Allies)", got)
	}
	if got := opp.Graveyard.Size() - gy; got != 2 {
		t.Errorf("%d cards reached the graveyard, want 2", got)
	}
}

func TestWwk5582AkoumBattlesingerPumpsAllies(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	singer := b12Push(g, me.ID, "Akoum Battlesinger", "Creature — Human Berserker Ally", wwkAkoumBattlesingerOracle, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", wwkBearType, 2, 2)
	friend := castHoldCreature(t, g, "Kor Friend", wwkAllyFriendType, "", 1, 1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, singer); p != 2 {
		t.Errorf("the Battlesinger has power %d, want 2", p)
	}
	if p, _ := ptOf(t, g, friend); p != 2 {
		t.Errorf("the other Ally has power %d, want 2", p)
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("a non-Ally has power %d, want 2", p)
	}
}

// --- spells ------------------------------------------------------------------------------

func TestWwk5582BullRushPumps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	bear := b12Creature(g, top100ActiveSeat(g).ID, "Bear", wwkBearType, 2, 2)
	castCatalogSpell(t, g, "Bull Rush", "Instant", wwkBullRushOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 4/2", p, tough)
	}
}

func TestWwk5582IonasJudgmentExilesAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Iona's Judgment", "Sorcery", wwkIonasJudgmentOracle, cardTarget(rock)); err == nil {
		t.Fatal("Iona's Judgment targeted an artifact")
	}
	aura := b12Permanent(g, opp.ID, "Pact", "Enchantment")
	castCatalogSpell(t, g, "Iona's Judgment", "Sorcery", wwkIonasJudgmentOracle, cardTarget(aura))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !inExile(g, aura) {
		t.Error("the enchantment was not exiled")
	}
}

func TestWwk5582VeteransReflexesPumpsAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	bear := b12Creature(g, top100ActiveSeat(g).ID, "Bear", wwkBearType, 2, 2)
	wwkSetTapped(g, bear)
	castCatalogSpell(t, g, "Veteran's Reflexes", "Instant", wwkVeteransReflexesOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
	if hmlTapped(t, g, bear) {
		t.Error("the Bear is still tapped")
	}
}

func TestWwk5582JoinTheRanksMakesTwoSoldierAllies(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	castCatalogSpell(t, g, "Join the Ranks", "Instant", wwkJoinTheRanksOracle, nil)
	passPriorityAroundTable(t, g)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Soldier Ally" && c.HasSubtype("Ally") && c.HasSubtype("Soldier") {
				n++
			}
		}
	})
	if n != 2 {
		t.Errorf("%d Soldier Ally tokens, want 2", n)
	}
}

func TestWwk5582AetherTradewindsBouncesOneOfEach(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "Mine", wwkBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", wwkBearType, 2, 2)
	if err := castCatalogSpellErr(t, g, "Aether Tradewinds", "Instant", wwkAetherTradewindsOracle,
		[]game.TargetRef{wwkCardRef(theirs), wwkCardRef(mine)}); err == nil {
		t.Fatal("the targets were accepted in the wrong slots")
	}
	castCatalogSpell(t, g, "Aether Tradewinds", "Instant", wwkAetherTradewindsOracle,
		[]game.TargetRef{wwkCardRef(mine), wwkCardRef(theirs)})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("your permanent did not return to your hand")
	}
	if !opp.Hand.Contains(theirs) {
		t.Error("their permanent did not return to its owner's hand")
	}
}

func TestWwk5582TombHexShrinksMoreAfterLandfall(t *testing.T) {
	for _, tc := range []struct {
		landfall bool
		want     int
	}{{false, 3}, {true, 1}} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		if tc.landfall {
			playLandFromHand(t, g, "Swamp", "")
		}
		bear := b12Creature(g, top100Opponent(g).ID, "Bear", wwkBearType, 5, 5)
		castCatalogSpell(t, g, "Tomb Hex", "Instant", wwkTombHexOracle, cardTarget(bear))
		passPriorityAroundTable(t, g)
		if p, tough := ptOf(t, g, bear); p != tc.want || tough != tc.want {
			t.Errorf("landfall %v: the Bear is %d/%d, want %d/%d", tc.landfall, p, tough, tc.want, tc.want)
		}
	}
}

func TestWwk5582RestForTheWearyGainsMoreAfterLandfall(t *testing.T) {
	for _, tc := range []struct {
		landfall bool
		want     int
	}{{false, 4}, {true, 8}} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := top100ActiveSeat(g)
		if tc.landfall {
			playLandFromHand(t, g, "Plains", "")
		}
		life := me.Life
		castCatalogSpell(t, g, "Rest for the Weary", "Instant", wwkRestForTheWearyOracle, []game.TargetRef{playerTarget(me.ID)})
		passPriorityAroundTable(t, g)
		if got := me.Life - life; got != tc.want {
			t.Errorf("landfall %v: gained %d, want %d", tc.landfall, got, tc.want)
		}
	}
}

func TestWwk5582MysteriesOfTheDeepDrawsMoreAfterLandfall(t *testing.T) {
	for _, tc := range []struct {
		landfall bool
		want     int
	}{{false, 2}, {true, 3}} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := top100ActiveSeat(g)
		if tc.landfall {
			playLandFromHand(t, g, "Island", "")
		}
		hand := me.Hand.Size()
		castCatalogSpell(t, g, "Mysteries of the Deep", "Instant", wwkMysteriesOfTheDeepOracle, nil)
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != tc.want {
			t.Errorf("landfall %v: drew %d, want %d", tc.landfall, got, tc.want)
		}
	}
}

// --- statics ----------------------------------------------------------------------------

func TestWwk5582LoamLionNeedsAForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lion := b12Push(g, me.ID, "Loam Lion", "Creature — Cat", wwkLoamLionOracle, 1, 1)
	if p, tough := ptOf(t, g, lion); p != 1 || tough != 1 {
		t.Fatalf("without a Forest the Lion is %d/%d, want 1/1", p, tough)
	}
	wwkLand(g, me.ID, "Forest")
	if p, tough := ptOf(t, g, lion); p != 2 || tough != 3 {
		t.Errorf("with a Forest the Lion is %d/%d, want 2/3", p, tough)
	}
}

func TestWwk5582RuthlessCullbladeWhenAnOpponentIsLow(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blade := b12Push(g, me.ID, "Ruthless Cullblade", "Creature — Vampire Warrior", wwkRuthlessCullbladeOracle, 2, 1)
	if p, tough := ptOf(t, g, blade); p != 2 || tough != 1 {
		t.Fatalf("with every opponent high the Cullblade is %d/%d, want 2/1", p, tough)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 10-opp.Life) })
	if opp.Life != 10 {
		t.Fatalf("the opponent is at %d, want 10", opp.Life)
	}
	if p, tough := ptOf(t, g, blade); p != 4 || tough != 2 {
		t.Errorf("with an opponent at 10 the Cullblade is %d/%d, want 4/2", p, tough)
	}
}

func TestWwk5582LandConditionalKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, land string
		keywords                     []string
	}{
		{"Summit Apes", "Creature — Ape", wwkSummitApesOracle, "Mountain", []string{"menace"}},
		{"Sejiri Merfolk", "Creature — Merfolk Soldier", wwkSejiriMerfolkOracle, "Plains", []string{wwkFirstStrike, "lifelink"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 1)
			for _, kw := range tc.keywords {
				if hasEffectiveKeyword(t, g, id, kw) {
					t.Fatalf("%s has %s without a %s", tc.name, kw, tc.land)
				}
			}
			wwkLand(g, me.ID, tc.land)
			for _, kw := range tc.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s with a %s", tc.name, kw, tc.land)
				}
			}
		})
	}
}

func TestWwk5582KitesailApprenticeWhileEquipped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kor := b12Push(g, me.ID, "Kitesail Apprentice", "Creature — Kor Soldier", wwkKitesailApprenticeOracle, 1, 1)
	if hasEffectiveKeyword(t, g, kor, "flying") {
		t.Fatal("the unequipped Apprentice has flying")
	}
	sword := b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	attachForTest(t, g, sword, kor)
	if p, tough := ptOf(t, g, kor); p != 2 || tough != 2 {
		t.Errorf("the equipped Apprentice is %d/%d, want 2/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, kor, "flying") {
		t.Error("the equipped Apprentice has no flying")
	}
}

func TestWwk5582ClawsOfValakutCountsYourMountains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", wwkBearType, 2, 2)
	wwkLand(g, me.ID, "Mountain")
	wwkLand(g, me.ID, "Mountain")
	wwkLand(g, opp.ID, "Mountain")
	claws := pushCatalogPermanent(g, me.ID, "Claws of Valakut", "Enchantment — Aura", wwkClawsOfValakutOracle, false)
	attachForTest(t, g, claws, bear)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the enchanted Bear is %d/%d, want 4/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, wwkFirstStrike) {
		t.Error("the enchanted Bear has no first strike")
	}
}

// --- the rest ----------------------------------------------------------------------------

func TestWwk5582CunningSparkmagePings(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mage := b12Push(g, me.ID, "Cunning Sparkmage", "Creature — Human Shaman", wwkCunningSparkmageOracle, 0, 1)
	life := opp.Life
	apaActivate(t, g, me, mage, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-1)
	}
	if !hmlTapped(t, g, mage) {
		t.Error("the Sparkmage did not tap")
	}
}

func TestWwk5582SurrakarBanisherBouncesATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	idle := b12Creature(g, opp.ID, "Idle", wwkBearType, 2, 2)
	tired := b12Creature(g, opp.ID, "Tired", wwkBearType, 2, 2)
	wwkSetTapped(g, tired)
	castHoldCreature(t, g, "Surrakar Banisher", "Creature — Surrakar", wwkSurrakarBanisherOracle, 3, 3)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, wwkCardRef(idle)); err == nil {
			t.Fatal("the Banisher targeted an untapped creature")
		}
	}
	wwkPickTarget(t, g, me.ID, wwkCardRef(tired))
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(tired) || g.Battlefield.Contains(tired) {
		t.Error("the tapped creature did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("the untapped creature left the battlefield")
	}
}

func TestWwk5582PilgrimsEyeFetchesABasicLand(t *testing.T) {
	jmpB6FetchesABasicToHand(t, "Pilgrim's Eye", "Artifact Creature — Thopter", wwkPilgrimsEyeOracle, 1, 1)
}

func TestWwk5582KorFirewalkerGainsOnARedSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, opp.ID, "Kor Firewalker", "Creature — Kor Soldier", wwkKorFirewalkerOracle, 2, 2)
	life := opp.Life
	white := dstHand(me, "White Thing", "Sorcery", "{W}", "W")
	if err := g.CastSpell(me.ID, white, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast white: %v", err)
	}
	if owedMays(g, opp.ID) != 0 || triggerOnStackFrom(g, "Kor Firewalker") {
		t.Error("a white spell triggered the Firewalker")
	}
	passPriorityAroundTable(t, g)
	red := dstHand(me, "Red Thing", "Sorcery", "{R}", "R")
	if err := g.CastSpell(me.ID, red, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast red: %v", err)
	}
	b06AnswerAllTriggerPrompts(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)
	if opp.Life != life+1 {
		t.Errorf("a red spell: the Firewalker's controller has %d life, want %d", opp.Life, life+1)
	}
}

func TestWwk5582NovablastWurmDestroysAllOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := b12Push(g, me.ID, "Novablast Wurm", "Creature — Wurm", wwkNovablastWurmOracle, 7, 7)
	mine := b12Creature(g, me.ID, "Mine", wwkBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", wwkBearType, 2, 2)
	declareAttack(t, g, opp.ID, wurm)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Error("the Wurm destroyed itself")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("another creature survived")
	}
}

func TestWwk5582GrotagThrasherStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thrasher := b12Push(g, me.ID, "Grotag Thrasher", "Creature — Lizard", wwkGrotagThrasherOracle, 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", wwkBearType, 2, 2)
	declareAttack(t, g, opp.ID, thrasher)
	wwkPickTarget(t, g, me.ID, wwkCardRef(theirs))
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, theirs); c == nil || !game.Restricted(c, game.CantBlock) {
		t.Error("the targeted creature can still block")
	}
}

func TestWwk5582SmolderingSpiresEntersTappedAndStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	theirs := b12Creature(g, opp.ID, "Their Bear", wwkBearType, 2, 2)
	spires := playLandFromHand(t, g, "Smoldering Spires", wwkSmolderingSpiresOracle)
	if !hmlTapped(t, g, spires) {
		t.Error("Smoldering Spires entered untapped")
	}
	wwkPickTarget(t, g, me.ID, wwkCardRef(theirs))
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, theirs); c == nil || !game.Restricted(c, game.CantBlock) {
		t.Error("the targeted creature can still block")
	}
}
