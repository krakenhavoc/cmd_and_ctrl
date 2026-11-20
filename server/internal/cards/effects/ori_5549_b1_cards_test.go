package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ori_5549_b1_cards_test.go — the first batch of Magic Origins cards
// (#5549). Every test plays the real catalog entry through the engine.

const (
	oriAbbotOfKeralKeepOracle        = "09be0f69-9c7f-491f-964d-bfbbc2d4251f"
	oriAkroanJailerOracle            = "e3ad2c77-2a41-4a4a-a6bf-e542d0162b88"
	oriAkroanSergeantOracle          = "8a206c6f-4f6f-4c31-9dab-d73a57db989c"
	oriAmprynTacticianOracle         = "4ae19803-23a7-4724-a4cd-dc6f3c16fb24"
	oriAnointerOfChampionsOracle     = "f74ef5b0-9222-43d0-a919-aaf28dc21b59"
	oriAspiringAeronautOracle        = "b2e31c0f-8914-4a18-bc05-44efb7e0f18c"
	oriAvenBattlePriestOracle        = "696fb1fa-5535-4f73-8235-ad4416dcfdd0"
	oriBlazingHellhoundOracle        = "598a7410-35e8-4d62-9bf1-d44ee7717040"
	oriBlessedSpiritsOracle          = "5fb899c1-3a2f-4b5e-a82d-fece51bdf786"
	oriBondedConstructOracle         = "c7556be3-649a-4994-8aaa-a0f258856682"
	oriBoundingKrasisOracle          = "c6736bac-314b-48af-8722-c16a0d66affa"
	oriCitadelCastellanOracle        = "b9e93832-93b7-409d-91aa-13757225824d"
	oriClashOfWillsOracle            = "56aec5bd-a8e8-403f-b339-0fc817426428"
	oriConclaveNaturalistsOracle     = "cb5de977-d420-41cc-85f0-6ffee8bbeff1"
	oriDeadbridgeShamanOracle        = "a57389ad-7f2a-47a6-8ae6-2821503cbb28"
	oriEnlightenedAsceticOracle      = "8fe257e5-b54a-4e56-8040-d2eaa0bb0a6a"
	oriEyeblightAssassinOracle       = "09cacd93-5c20-4993-9ee2-a756204796f7"
	oriFetidImpOracle                = "cdb68656-633d-436e-ad56-0fdb4814bd19"
	oriFieryImpulseOracle            = "27c84ad1-1491-459b-bf8a-0146062084dc"
	oriFirefiendElementalOracle      = "59c898d8-14f0-4645-84e7-aeedb2ee6c5e"
	oriGhirapurGearcrafterOracle     = "710dc9b1-7d32-4330-95a8-3e91a668202a"
	oriGuardianAutomatonOracle       = "cd29d0fd-685f-4c26-b69c-e96f1c0d59c4"
	oriHealingHandsOracle            = "3cc48835-3ac0-4774-b380-f9b21d2dc974"
	oriHeavyInfantryOracle           = "d07707ef-fa4e-423f-8a1f-095d19100321"
	oriHitchclawRecluseOracle        = "241ec1fc-cc50-468e-8ee3-f95a82a3e07b"
	oriIroassChampionOracle          = "8f2252f6-69b9-4c56-b078-803db301a874"
	oriJhessianThiefOracle           = "b917c3aa-2d54-49e0-924a-dfd02eaf4b2d"
	oriKnightOfThePilgrimsRoadOracle = "40999aa6-6424-42e4-9d47-33bee6bbb886"
	oriLightningJavelinOracle        = "a6f6b062-cd7e-4631-9921-40e054ab835d"
	oriMageRingBullyOracle           = "a0b4fa95-4c61-40ff-b645-a1dbe8825750"
	oriMalakirCullbladeOracle        = "4fbaf052-a8c3-4395-8f40-0cb3a10d5ccd"
	oriPharikasDiscipleOracle        = "d7798312-657c-4f4a-9e2d-906284caf50f"
	oriPrismRingOracle               = "129ee7d6-e3df-40ef-9502-7aad10183031"
	oriRabidBloodsuckerOracle        = "e6b0c47b-9758-4ee3-b831-0af0650660fe"
	oriReaveSoulOracle               = "8eb328e6-332e-435c-9733-4c5336f17824"
	oriRhoxMaulersOracle             = "019461d9-63ae-43d3-a9d4-bd745d0cf46c"
	oriRingwardenOwlOracle           = "4a8b3316-a01a-4206-bbed-d7aac6c834bd"
	oriSeparatistVoidmageOracle      = "d9fabfa8-b930-428d-af4e-640fb310a5b2"
	oriShamblingGhoulOracle          = "2acc1657-d716-41d7-aedf-9b080a58ec83"
	oriSkysnareSpiderOracle          = "b820c29e-2428-4074-a024-9502848484d6"
	oriStalwartAvenOracle            = "09cc2cda-14d1-424e-a688-1b49b1b4c392"
	oriSubterraneanScoutOracle       = "19f66bfa-bdfc-430a-ae39-4535f0b28324"
	oriSuppressionBondsOracle        = "e844ded0-821a-4378-8f58-bf1e1c0269d5"
	oriThunderclapWyvernOracle       = "15276d3b-a117-44bf-87c3-c17e032e4a26"
	oriTopanFreebladeOracle          = "9e2cd0d5-1537-412b-9ecd-989ab4ac2700"
	oriUnholyHungerOracle            = "ed8d8eb5-fbc8-4b7f-9b75-53ab118e72de"
	oriVeteransSidearmOracle         = "f038c11e-b070-40fd-9d0f-e031f7bc2bd2"
	oriVrynWingmareOracle            = "c7c241f2-d3cc-4032-a9da-149de73c2539"
	oriWarHornOracle                 = "99bc739e-c745-47d8-8401-2f34b6becd85"
	oriWarOracleOracle               = "aa26fb50-1b53-488e-90d1-b0c7af273622"

	oriBearType    = "Creature — Bear"
	oriSoldierType = "Creature — Human Soldier"
)

// oriAllOracles is every card this batch adds.
var oriAllOracles = []string{
	oriAbbotOfKeralKeepOracle, oriAkroanJailerOracle, oriAkroanSergeantOracle, oriAmprynTacticianOracle,
	oriAnointerOfChampionsOracle, oriAspiringAeronautOracle, oriAvenBattlePriestOracle, oriBlazingHellhoundOracle,
	oriBlessedSpiritsOracle, oriBondedConstructOracle, oriBoundingKrasisOracle, oriCitadelCastellanOracle,
	oriClashOfWillsOracle, oriConclaveNaturalistsOracle, oriDeadbridgeShamanOracle, oriEnlightenedAsceticOracle,
	oriEyeblightAssassinOracle, oriFetidImpOracle, oriFieryImpulseOracle, oriFirefiendElementalOracle,
	oriGhirapurGearcrafterOracle, oriGuardianAutomatonOracle, oriHealingHandsOracle, oriHeavyInfantryOracle,
	oriHitchclawRecluseOracle, oriIroassChampionOracle, oriJhessianThiefOracle, oriKnightOfThePilgrimsRoadOracle,
	oriLightningJavelinOracle, oriMageRingBullyOracle, oriMalakirCullbladeOracle, oriPharikasDiscipleOracle,
	oriPrismRingOracle, oriRabidBloodsuckerOracle, oriReaveSoulOracle, oriRhoxMaulersOracle,
	oriRingwardenOwlOracle, oriSeparatistVoidmageOracle, oriShamblingGhoulOracle, oriSkysnareSpiderOracle,
	oriStalwartAvenOracle, oriSubterraneanScoutOracle, oriSuppressionBondsOracle, oriThunderclapWyvernOracle,
	oriTopanFreebladeOracle, oriUnholyHungerOracle, oriVeteransSidearmOracle, oriVrynWingmareOracle,
	oriWarHornOracle, oriWarOracleOracle,
}

// oriLibraryTop is the top card of p's library.
func oriLibraryTop(t *testing.T, p *game.Player) uuid.UUID {
	t.Helper()
	if p.Library.Size() == 0 {
		t.Fatal("the library is empty")
	}
	return p.Library.Cards[p.Library.Size()-1].InstanceID
}

// --- every card is in the catalog ----------------------------------------

func TestOri5549EveryCardIsRegisteredFull(t *testing.T) {
	if len(oriAllOracles) != 50 {
		t.Fatalf("%d cards listed, want 50", len(oriAllOracles))
	}
	for _, oracle := range oriAllOracles {
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

// --- printed keywords -------------------------------------------------------

func TestOri5549KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, oracle string
		keywords     []string
	}{
		{"Akroan Sergeant", oriAkroanSergeantOracle, []string{"first strike"}},
		{"Citadel Castellan", oriCitadelCastellanOracle, []string{"vigilance"}},
		{"Firefiend Elemental", oriFirefiendElementalOracle, []string{"haste"}},
		{"Hitchclaw Recluse", oriHitchclawRecluseOracle, []string{"reach"}},
		{"Iroas's Champion", oriIroassChampionOracle, []string{"double strike"}},
		{"Pharika's Disciple", oriPharikasDiscipleOracle, []string{"deathtouch"}},
		{"Rhox Maulers", oriRhoxMaulersOracle, []string{"trample"}},
		{"Ringwarden Owl", oriRingwardenOwlOracle, []string{"flying", "prowess"}},
		{"Skysnare Spider", oriSkysnareSpiderOracle, []string{"vigilance", "reach"}},
		{"Stalwart Aven", oriStalwartAvenOracle, []string{"flying"}},
		{"Topan Freeblade", oriTopanFreebladeOracle, []string{"vigilance"}},
		{"War Oracle", oriWarOracleOracle, []string{"lifelink"}},
		{"Abbot of Keral Keep", oriAbbotOfKeralKeepOracle, []string{"prowess"}},
		{"Aspiring Aeronaut", oriAspiringAeronautOracle, []string{"flying"}},
		{"Aven Battle Priest", oriAvenBattlePriestOracle, []string{"flying"}},
		{"Blessed Spirits", oriBlessedSpiritsOracle, []string{"flying"}},
		{"Fetid Imp", oriFetidImpOracle, []string{"flying"}},
		{"Jhessian Thief", oriJhessianThiefOracle, []string{"prowess"}},
		{"Mage-Ring Bully", oriMageRingBullyOracle, []string{"prowess"}},
		{"Thunderclap Wyvern", oriThunderclapWyvernOracle, []string{"flash", "flying"}},
		{"Vryn Wingmare", oriVrynWingmareOracle, []string{"flying"}},
		{"Bounding Krasis", oriBoundingKrasisOracle, []string{"flash"}},
		{"Rabid Bloodsucker", oriRabidBloodsuckerOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, "Creature — Test", tc.oracle, 1, 1)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s (abilities %v)", tc.name, kw, abilities)
				}
			}
		})
	}
}

// Renown N (CR 702.112): combat damage to a player puts N +1/+1
// counters on it and it becomes renowned.
func TestOri5549RenownCreaturesBecomeRenowned(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		n            int
	}{
		{"Akroan Sergeant", oriAkroanSergeantOracle, 1},
		{"Citadel Castellan", oriCitadelCastellanOracle, 2},
		{"Firefiend Elemental", oriFirefiendElementalOracle, 1},
		{"Knight of the Pilgrim's Road", oriKnightOfThePilgrimsRoadOracle, 1},
		{"Pharika's Disciple", oriPharikasDiscipleOracle, 1},
		{"Rhox Maulers", oriRhoxMaulersOracle, 2},
		{"Stalwart Aven", oriStalwartAvenOracle, 1},
		{"Topan Freeblade", oriTopanFreebladeOracle, 1},
		{"War Oracle", oriWarOracleOracle, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := pushRenownCard(g, me.ID, tc.name, tc.oracle, 2, 2)
			if renownedNow(g, id) {
				t.Fatal("renowned before it dealt combat damage")
			}
			attackWith(t, g, opp.ID, id)
			passPriorityAroundTable(t, g)
			if !renownedNow(g, id) || plusOneCounters(g, id) != tc.n {
				t.Errorf("renowned %v, %d counters; want renowned with %d", renownedNow(g, id), plusOneCounters(g, id), tc.n)
			}
		})
	}
}

// Prowess: a noncreature spell pumps the Owl until end of turn.
func TestOri5549RingwardenOwlProwess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	owl := b12Push(g, me.ID, "Ringwarden Owl", "Creature — Bird", oriRingwardenOwlOracle, 3, 3)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, owl); p != 4 || tough != 4 {
		t.Errorf("the Owl is %d/%d after a noncreature spell, want 4/4", p, tough)
	}
}

// --- entry and death triggers ------------------------------------------------

func TestOri5549AbbotOfKeralKeepExilesTheTopCardToPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	top := oriLibraryTop(t, me)
	castHoldCreature(t, g, "Abbot of Keral Keep", "Creature — Human Monk", oriAbbotOfKeralKeepOracle, 2, 1)
	passPriorityAroundTable(t, g)
	if me.Library.Contains(top) || !g.Exile.Contains(top) {
		t.Error("the top card of the library was not exiled")
	}
}

func TestOri5549AmprynTacticianPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Mine", oriBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", oriBearType, 2, 2)
	tactician := castHoldCreature(t, g, "Ampryn Tactician", oriSoldierType, oriAmprynTacticianOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, tactician); p != 4 || tough != 4 {
		t.Errorf("the Tactician is %d/%d, want 4/4", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestOri5549ThopterMakers(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Aspiring Aeronaut", "Creature — Human Artificer", oriAspiringAeronautOracle},
		{"Ghirapur Gearcrafter", "Creature — Human Artificer", oriGhirapurGearcrafterOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 2)
			passPriorityAroundTable(t, g)
			if n := countBattlefieldNamed(g, me.ID, "Thopter"); n != 1 {
				t.Fatalf("%d Thopters, want 1", n)
			}
			thopter := findBattlefieldByName(g, "Thopter")
			if !hasEffectiveKeyword(t, g, thopter, "flying") {
				t.Error("the Thopter has no flying")
			}
		})
	}
}

func TestOri5549AvenBattlePriestGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castHoldCreature(t, g, "Aven Battle Priest", "Creature — Bird Cleric", oriAvenBattlePriestOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestOri5549RabidBloodsuckerDrainsEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	castHoldCreature(t, g, "Rabid Bloodsucker", "Creature — Vampire", oriRabidBloodsuckerOracle, 3, 2)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if s.Life != lives[s.ID]-2 {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, lives[s.ID]-2)
		}
	}
}

func TestOri5549GuardianAutomatonGainsThreeWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Guardian Automaton", "Artifact Creature — Construct", oriGuardianAutomatonOracle, 3, 3)
	life := me.Life
	b27Kill(g, id)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestOri5549DeadbridgeShamanMakesAnOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d2dSetHand(opp, 3)
	id := b12Push(g, me.ID, "Deadbridge Shaman", "Creature — Elf Shaman", oriDeadbridgeShamanOracle, 3, 1)
	b27Kill(g, id)
	g.RunStateChecksForTest()
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("the trigger targeted its own controller")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 1 {
		t.Errorf("the opponent owes %d discards, want 1", n)
	}
}

func TestOri5549HeavyInfantryTapsAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", oriBearType, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", oriBearType, 2, 2)
	castHoldCreature(t, g, "Heavy Infantry", oriSoldierType, oriHeavyInfantryOracle, 3, 4)
	if pick := latestPickTarget(g, me.ID); pick != nil && hasID(pick.PickTargetCards, mine) {
		t.Error("a creature I control was offered")
	}
	answerPickTargetIfOpen(t, g, theirs)
	passPriorityAroundTable(t, g)
	if !hmlTapped(t, g, theirs) {
		t.Error("the opposing creature is untapped")
	}
	if hmlTapped(t, g, mine) {
		t.Error("my creature was tapped")
	}
}

func TestOri5549EyeblightAssassinShrinksAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", oriBearType, 3, 3)
	castHoldCreature(t, g, "Eyeblight Assassin", "Creature — Elf Assassin", oriEyeblightAssassinOracle, 2, 2)
	answerPickTargetIfOpen(t, g, theirs)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the target is %d/%d, want 2/2", p, tough)
	}
}

func TestOri5549OptionalDestroyOnEntry(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, victimType string
	}{
		{"Conclave Naturalists", "Creature — Dryad", oriConclaveNaturalistsOracle, "Artifact"},
		{"Enlightened Ascetic", "Creature — Cat Monk", oriEnlightenedAsceticOracle, "Enchantment"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			victim := b12Permanent(g, opp.ID, "Victim", tc.victimType)
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 1)
			answerPickTargetIfOpen(t, g, victim)
			answerLatestTriggerPrompt(t, g, me.ID, true)
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) || !opp.Graveyard.Contains(victim) {
				t.Error("the target was not destroyed")
			}
		})
	}
}

// Enlightened Ascetic can't target an artifact: with none but an
// artifact on the battlefield there is no prompt at all (CR 603.3d).
func TestOri5549EnlightenedAsceticIgnoresArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	castHoldCreature(t, g, "Enlightened Ascetic", "Creature — Cat Monk", oriEnlightenedAsceticOracle, 1, 1)
	if latestPickTarget(g, me.ID) != nil || owedMays(g, me.ID) > 0 {
		t.Error("the trigger asked something with no enchantment to target")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was destroyed")
	}
}

func TestOri5549SeparatistVoidmageBouncesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", oriBearType, 2, 2)
	castHoldCreature(t, g, "Separatist Voidmage", "Creature — Human Wizard", oriSeparatistVoidmageOracle, 2, 2)
	answerPickTargetIfOpen(t, g, theirs)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestOri5549BoundingKrasisTapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", oriBearType, 2, 2)
	castHoldCreature(t, g, "Bounding Krasis", "Creature — Fish Lizard", oriBoundingKrasisOracle, 3, 3)
	answerPickTargetIfOpen(t, g, theirs)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	if !hmlTapped(t, g, theirs) {
		t.Error("the creature was not tapped")
	}
}

func TestOri5549ShamblingGhoulEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := castHoldCreature(t, g, "Shambling Ghoul", "Creature — Zombie", oriShamblingGhoulOracle, 2, 3)
	if !hmlTapped(t, g, id) {
		t.Error("Shambling Ghoul entered untapped")
	}
}

func TestOri5549SubterraneanScoutMakesASmallCreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", oriBearType, 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", oriBearType, 2, 2)
	castHoldCreature(t, g, "Subterranean Scout", "Creature — Goblin Scout", oriSubterraneanScoutOracle, 2, 1)
	answerPickTargetIfOpen(t, g, bear)
	passPriorityAroundTable(t, g)
	brAttack(t, g, bear)
	if err := g.DeclareBlocker(blocker, bear); err == nil {
		t.Error("the targeted creature was blocked")
	}
}

// --- attack and combat triggers -------------------------------------------------

func TestOri5549JhessianThiefDrawsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thief := b12Push(g, me.ID, "Jhessian Thief", "Creature — Human Rogue", oriJhessianThiefOracle, 1, 3)
	hand := me.Hand.Size()
	dealCombatDamage(g, me.ID, thief, opp.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestOri5549MageRingBullyMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bully := b12Push(g, me.ID, "Mage-Ring Bully", "Creature — Human Warrior", oriMageRingBullyOracle, 2, 2)
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != bully {
		t.Fatalf("pass with the Bully home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(bully, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Bully attacking: %v", err)
	}
}

func TestOri5549BondedConstructCantAttackAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	construct := apaPush(g, me.ID, me.ID, apaCreature("Bonded Construct", oriBondedConstructOracle, 2, 1))
	bear := apaPush(g, me.ID, me.ID, apaCreature("Bear", "", 2, 2))
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(construct, opp.ID); err != nil {
		t.Fatalf("declaring the Construct first was refused: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with the Construct alone: %v, want ErrAttackLimit", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("the companion was refused: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("passing with two attackers: %v", err)
	}
}

func TestOri5549MalakirCullbladeGrowsOnlyOnOpponentsDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blade := b12Push(g, me.ID, "Malakir Cullblade", "Creature — Vampire Warrior", oriMalakirCullbladeOracle, 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", oriBearType, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", oriBearType, 2, 2)
	b27Kill(g, theirs)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := countersOn(g, blade, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after an opponent's creature died, want 1", n)
	}
	b27Kill(g, mine)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := countersOn(g, blade, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after my own creature died, want still 1", n)
	}
}

func TestOri5549BlessedSpiritsGrowsOnEnchantmentSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spirits := b12Push(g, me.ID, "Blessed Spirits", "Creature — Spirit", oriBlessedSpiritsOracle, 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Anthem", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, spirits, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1 (only the enchantment spell)", n)
	}
}

func TestOri5549PrismRingGainsLifeForTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushChosenColorPermanent(t, g, me.ID, "Prism Ring", "Artifact", oriPrismRingOracle, "R")
	advanceToMain(t, g)
	before := me.Life
	red := game.Card{Name: "Red Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}}
	if err := socB19CastWithPool(g, me, red, "R", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d after a red spell, want %d", me.Life, before+1)
	}
	blue := game.Card{Name: "Blue Thing", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"}}
	if err := socB19CastWithPool(g, me, blue, "U", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d after a blue spell, want unchanged %d", me.Life, before+1)
	}
}

// --- activated abilities ------------------------------------------------------------

func TestOri5549AkroanJailerTapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	jailer := b12Push(g, me.ID, "Akroan Jailer", oriSoldierType, oriAkroanJailerOracle, 1, 1)
	bear := b12Creature(g, opp.ID, "Bear", oriBearType, 2, 2)
	apaMana(me, "W", "C", "C")
	apaActivate(t, g, me, jailer, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !hmlTapped(t, g, bear) {
		t.Error("the target is untapped")
	}
	if !hmlTapped(t, g, jailer) {
		t.Error("the Jailer did not tap")
	}
}

func TestOri5549AnointerOfChampionsPumpsOnlyAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	anointer := b12Push(g, me.ID, "Anointer of Champions", "Creature — Human Cleric", oriAnointerOfChampionsOracle, 1, 1)
	raider := b12Creature(g, me.ID, "Raider", oriSoldierType, 2, 2)
	idle := b12Creature(g, me.ID, "Idle", oriSoldierType, 2, 2)
	brAttack(t, g, raider)
	if err := g.ActivateCatalogAbility(me.ID, anointer, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(idle)}); err == nil {
		t.Fatal("the Anointer targeted a creature that isn't attacking")
	}
	apaActivate(t, g, me, anointer, 0, game.ActivateAbilityParams{Targets: cardTarget(raider)})
	if p, tough := effectivePTOf(t, g, raider); p != 3 || tough != 3 {
		t.Errorf("the attacker is %d/%d, want 3/3", p, tough)
	}
}

func TestOri5549BlazingHellhoundSacrificesAnotherToPing(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	hound := b12Push(g, me.ID, "Blazing Hellhound", "Creature — Elemental Dog", oriBlazingHellhoundOracle, 4, 3)
	bear := b12Creature(g, me.ID, "Bear", oriBearType, 2, 2)
	apaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, hound, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{hound}, Targets: []game.TargetRef{playerTarget(opp.ID)},
	}); err == nil {
		t.Fatal("the Hellhound sacrificed itself to pay for its own ability")
	}
	life := opp.Life
	apaActivate(t, g, me, hound, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear}, Targets: []game.TargetRef{playerTarget(opp.ID)},
	})
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
}

func TestOri5549FetidImpGainsDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	imp := b12Push(g, me.ID, "Fetid Imp", "Creature — Imp", oriFetidImpOracle, 1, 2)
	if hasEffectiveKeyword(t, g, imp, "deathtouch") {
		t.Fatal("the Imp has deathtouch before activating")
	}
	apaMana(me, "B")
	apaActivate(t, g, me, imp, 0, game.ActivateAbilityParams{})
	if !hasEffectiveKeyword(t, g, imp, "deathtouch") {
		t.Error("the Imp has no deathtouch")
	}
}

// --- statics ---------------------------------------------------------------------------

func TestOri5549ThunderclapWyvernPumpsYourOtherFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wyvern := b12Push(g, me.ID, "Thunderclap Wyvern", "Creature — Drake", oriThunderclapWyvernOracle, 2, 3)
	flyer := apaPush(g, me.ID, me.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	walker := b12Creature(g, me.ID, "Walker", oriBearType, 2, 2)
	theirs := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	if p, tough := effectivePTOf(t, g, flyer); p != 2 || tough != 2 {
		t.Errorf("your flyer is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, wyvern); p != 2 || tough != 3 {
		t.Errorf("the Wyvern is %d/%d, want 2/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, walker); p != 2 || tough != 2 {
		t.Errorf("a creature without flying is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("an opponent's flyer is %d/%d, want 1/1", p, tough)
	}
}

func TestOri5549VeteransSidearmPumpsTheEquippedCreature(t *testing.T) {
	bnsB23Pump(t, oriVeteransSidearmOracle, "Veteran's Sidearm", 1, 1, "")
}

func TestOri5549WarHornPumpsYourAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "War Horn", "Artifact", oriWarHornOracle, false)
	bear := b12Creature(g, me.ID, "Bear", oriBearType, 2, 2)
	home := b12Creature(g, me.ID, "Home Bear", oriBearType, 2, 2)
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Fatalf("a creature not attacking has power %d, want 2", p)
	}
	declareAttack(t, g, opp.ID, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("an attacking Bear is %d/%d, want 3/2", p, tough)
	}
	if p, _ := effectivePTOf(t, g, home); p != 2 {
		t.Errorf("a creature that stayed home has power %d, want 2", p)
	}
}

func TestOri5549VrynWingmareTaxesNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Vryn Wingmare", "Creature — Pegasus", oriVrynWingmareOracle, 2, 1)
	advanceToMain(t, g)
	bolt := game.Card{Name: "Red Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}}
	if err := socB19CastWithPool(g, me, bolt, "R", 1, game.CastSpellParams{}); err == nil {
		t.Fatal("a noncreature spell was cast without the {1} tax")
	}
	bolt2 := game.Card{Name: "Red Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}}
	if err := socB19CastWithPool(g, me, bolt2, "R", 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("a noncreature spell with the tax paid: %v", err)
	}
	passPriorityAroundTable(t, g)
	goblin := game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", ManaCost: "{R}", Colors: []string{"R"}, Power: 1, Toughness: 1}
	if err := socB19CastWithPool(g, me, goblin, "R", 1, game.CastSpellParams{}); err != nil {
		t.Errorf("a creature spell was taxed: %v", err)
	}
}

func TestOri5549SuppressionBondsStopsBlocking(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, g, "Suppression Bonds", "Enchantment — Aura", oriSuppressionBondsOracle, cardTarget(land)); err == nil {
		t.Fatal("Suppression Bonds enchanted a land")
	}
	blocker := b12Creature(g, opp.ID, "Blocker", oriBearType, 2, 2)
	castCatalogSpell(t, g, "Suppression Bonds", "Enchantment — Aura", oriSuppressionBondsOracle, cardTarget(blocker))
	passPriorityAroundTable(t, g)
	attacker := b12Creature(g, me.ID, "Attacker", oriBearType, 2, 2)
	brAttack(t, g, attacker)
	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("the enchanted creature blocked")
	}
}

// --- the spells ------------------------------------------------------------------------

func TestOri5549ClashOfWillsCountersUnlessXIsPaid(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := b13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := b13HandCard(me, "Clash of Wills", "Instant", oriClashOfWillsOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Targets: b16TargetCard(bolt)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := answerPayUnless(t, g, opp.ID, false)
	if c.PayCost != "{2}" {
		t.Errorf("the counter asks for %q, want {2}", c.PayCost)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the spell was not countered")
	}
}

func TestOri5549ReaveSoulDestroysOnlyASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	big := b12Creature(g, opp.ID, "Big", oriBearType, 4, 4)
	small := b12Creature(g, opp.ID, "Small", oriBearType, 3, 3)
	if err := castCatalogSpellErr(t, g, "Reave Soul", "Sorcery", oriReaveSoulOracle, cardTarget(big)); err == nil {
		t.Fatal("Reave Soul targeted a creature with power 4")
	}
	castCatalogSpell(t, g, "Reave Soul", "Sorcery", oriReaveSoulOracle, cardTarget(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the creature survived")
	}
}

func TestOri5549HealingHandsGainsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	life, mine, hand := opp.Life, me.Life, me.Hand.Size()
	castCatalogSpell(t, g, "Healing Hands", "Sorcery", oriHealingHandsOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life+4 {
		t.Errorf("the target is at %d, want %d", opp.Life, life+4)
	}
	if me.Life != mine {
		t.Errorf("the caster's life changed to %d", me.Life)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestOri5549LightningJavelinDealsThreeThenScries(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	life := opp.Life
	castCatalogSpell(t, g, "Lightning Javelin", "Sorcery", oriLightningJavelinOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-3)
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scrying %d cards, want 1", len(c.ScryCards))
	}
}

// Spell mastery counts the instants and sorceries in the caster's
// graveyard as the spell resolves.
func TestOri5549FieryImpulseDealsThreeWithSpellMastery(t *testing.T) {
	for _, tc := range []struct {
		name    string
		spells  int
		damage  int
		comment string
	}{
		{"without", 1, 2, "one instant in the graveyard"},
		{"with", 2, 3, "two instants and sorceries in the graveyard"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, opp := top100ActiveSeat(g), top100Opponent(g)
			for i := 0; i < tc.spells; i++ {
				wthGraveyardCard(me, "Spent", "Instant")
			}
			wthGraveyardCard(me, "Old Bear", oriBearType)
			ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
			castCatalogSpell(t, g, "Fiery Impulse", "Instant", oriFieryImpulseOracle, cardTarget(ogre))
			passPriorityAroundTable(t, g)
			if got := findCardAnywhere(t, g, ogre).DamageMarked; got != tc.damage {
				t.Errorf("%s: the target has %d damage, want %d", tc.comment, got, tc.damage)
			}
		})
	}
}

func TestOri5549UnholyHungerGainsTwoWithSpellMastery(t *testing.T) {
	for _, tc := range []struct {
		name   string
		spells int
		gain   int
	}{
		{"without", 1, 0},
		{"with", 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, opp := top100ActiveSeat(g), top100Opponent(g)
			for i := 0; i < tc.spells; i++ {
				wthGraveyardCard(me, "Spent", "Sorcery")
			}
			bear := b12Creature(g, opp.ID, "Bear", oriBearType, 2, 2)
			life := me.Life
			castCatalogSpell(t, g, "Unholy Hunger", "Instant", oriUnholyHungerOracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(bear) {
				t.Error("the creature survived")
			}
			if me.Life != life+tc.gain {
				t.Errorf("life %d, want %d", me.Life, life+tc.gain)
			}
		})
	}
}
