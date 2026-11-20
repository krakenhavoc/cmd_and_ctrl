package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// som_5577_b2_cards_test.go — Scars of Mirrodin, batch 2 (#5577). Each
// test plays the real catalog entry and asserts what its text says.

const (
	somB2TyrranaxOracle      = "8bed1349-4d0a-4544-ba66-c38e72cd2863"
	somB2WayfarerOracle      = "cd990c63-ba29-46f8-afa5-4cd5fe0b0166"
	somB2ReaverOracle        = "9bd62fa0-b99c-4b64-a6f6-608cf01f8ef3"
	somB2SeastriderOracle    = "249b722e-2f4c-4884-bebb-e967f4fd3238"
	somB2ThresherOracle      = "a8b9e29d-e757-4758-8c5e-7013853e2e92"
	somB2ScoriaOracle        = "1a6e8369-241f-4ddc-b4cd-a974d6b689a7"
	somB2NimOracle           = "02d4fe7d-ade4-40b1-b323-73a52c440525"
	somB2CystbearerOracle    = "81704164-bad9-4e52-9b69-3d49e92fe284"
	somB2FallenOracle        = "e02740e6-3eba-4f82-831b-35a9f122f07f"
	somB2ForgerOracle        = "9e623e0a-3c2d-4536-b12f-9437ef820499"
	somB2WardenOracle        = "3b17dc39-a564-4c61-b483-98ed3da4361b"
	somB2EdgewrightOracle    = "b718639d-cfba-4430-b5e3-1e644edd36d7"
	somB2GliderOracle        = "43e2e1ed-5c6a-49b4-8ab8-57861c02fc3e"
	somB2SunchaserOracle     = "5e07d314-119f-4918-9105-77543ccae7e6"
	somB2BrigadeOracle       = "ae32c0a5-0874-4cab-9a24-29fa572f3193"
	somB2BattlegearOracle    = "b81fed36-07c4-4903-b5b4-24e62d2ce239"
	somB2HarnessOracle       = "bd4ad383-8c40-4dda-bc9c-ba8a001d6882"
	somB2LifestaffOracle     = "e69b8388-1311-427f-9300-42badd5338ff"
	somB2BarrageOgreOracle   = "d556c71e-ce16-4229-880b-744790f93797"
	somB2HarvesterOracle     = "f105ccdd-65bc-4880-bbfe-f57fc0e060ea"
	somB2DarkslickOracle     = "cef11518-ecb2-4250-996c-a03fa9dfb09e"
	somB2ArchersOracle       = "bcaf3f32-c4e5-4d39-be98-bfc00327021b"
	somB2FerrovoreOracle     = "1f0f068a-fd4e-4a6f-9439-0e694e34a945"
	somB2HellionOracle       = "47d4da1e-b1a2-4a23-9b96-c85d5d20e3ef"
	somB2ScudderOracle       = "93f37df7-e1c4-457b-bc22-00c49b0c4f79"
	somB2NeurokReplicaOracle = "de8cdbd4-d8f7-4ce8-a69b-6d2020beff63"
	somB2DaredevilOracle     = "4be254e7-d082-4c10-8495-1ef8b80721e1"
	somB2ScrapmelterOracle   = "cef64ec6-cdfb-42a4-9a4a-3ef5448e93fc"
	somB2SaberclawOracle     = "4989d38e-332f-48f5-863b-1566921156c8"
	somB2SalvageScoutOracle  = "502e991c-f4de-4095-adac-cb0c25e44a7c"
	somB2SylvokReplicaOracle = "97818ed4-3590-4a3b-92ce-716f9cfdc78e"
	somB2HeartstokerOracle   = "0d996788-7193-45cd-bc6d-0f470521d1ca"
	somB2VulshokReplicaOracl = "da8e2236-88ae-4f30-884b-b961231fba8f"
	somB2MoriokReplicaOracle = "6273cd72-5330-4712-ae73-b058023f0435"
	somB2SolitonOracle       = "c04b4323-db6a-4ab1-99b7-eef3acad3c4f"
	somB2SeizeOracle         = "f6653d90-cd30-4b2b-9592-97d5769dc3d8"
	somB2UntamedMightOracle  = "4b6dd710-f7f2-4fc8-9b66-b48c7737b135"
	somB2VaultSkywardOracle  = "e2a32a0f-bf06-4cb3-a704-ff0f783f19e4"
	somB2PassageOracle       = "df702285-3577-47cc-aa1e-868921cdf6f2"
	somB2WingPunctureOracle  = "5de32ec7-3085-4b6b-a52e-5a1b3fdf07d3"
	somB2HaltOrderOracle     = "a5fdf1b2-da22-4fa6-810a-85c3f7587535"
	somB2InstillOracle       = "0040db73-ac11-4c15-a1ac-5a2c4c085f09"
	somB2DefianceOracle      = "73c7d938-5cec-4f96-b649-3f67bef29407"
	somB2TowerOracle         = "27618701-e8c9-4117-94f3-3836cf34990d"
	somB2TumbleMagnetOracle  = "0481adcd-b66d-432e-80ba-3a4718bb3a7d"
	somB2FlightBombOracle    = "9f679093-86a0-4211-8491-de1f6d81c483"
	somB2PanicBombOracle     = "4429a30e-17f1-4bcc-92f0-55ab1315adff"
	somB2OriginBombOracle    = "3de0e14f-c917-4c00-8341-1170ac9b3cbc"
	somB2HorizonBombOracle   = "12080ea9-0800-4601-b553-6711dd9f9806"
)

func TestSomB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		somB2TyrranaxOracle:      "Alpha Tyrranax",
		somB2WayfarerOracle:      "Loxodon Wayfarer",
		somB2ReaverOracle:        "Moriok Reaver",
		somB2SeastriderOracle:    "Plated Seastrider",
		somB2ThresherOracle:      "Razorfield Thresher",
		somB2ScoriaOracle:        "Scoria Elemental",
		somB2NimOracle:           "Contagious Nim",
		somB2CystbearerOracle:    "Cystbearer",
		somB2FallenOracle:        "Tel-Jilad Fallen",
		somB2ForgerOracle:        "Carapace Forger",
		somB2WardenOracle:        "Ghalma's Warden",
		somB2EdgewrightOracle:    "Auriok Edgewright",
		somB2GliderOracle:        "Snapsail Glider",
		somB2SunchaserOracle:     "Auriok Sunchaser",
		somB2BrigadeOracle:       "Ezuri's Brigade",
		somB2BattlegearOracle:    "Barbed Battlegear",
		somB2HarnessOracle:       "Strider Harness",
		somB2LifestaffOracle:     "Sylvok Lifestaff",
		somB2BarrageOgreOracle:   "Barrage Ogre",
		somB2HarvesterOracle:     "Corrupted Harvester",
		somB2DarkslickOracle:     "Darkslick Drake",
		somB2ArchersOracle:       "Ezuri's Archers",
		somB2FerrovoreOracle:     "Ferrovore",
		somB2HellionOracle:       "Flameborn Hellion",
		somB2ScudderOracle:       "Necrogen Scudder",
		somB2NeurokReplicaOracle: "Neurok Replica",
		somB2DaredevilOracle:     "Oxidda Daredevil",
		somB2ScrapmelterOracle:   "Oxidda Scrapmelter",
		somB2SaberclawOracle:     "Saberclaw Golem",
		somB2SalvageScoutOracle:  "Salvage Scout",
		somB2SylvokReplicaOracle: "Sylvok Replica",
		somB2HeartstokerOracle:   "Vulshok Heartstoker",
		somB2VulshokReplicaOracl: "Vulshok Replica",
		somB2MoriokReplicaOracle: "Moriok Replica",
		somB2SolitonOracle:       "Soliton",
		somB2SeizeOracle:         "Seize the Initiative",
		somB2UntamedMightOracle:  "Untamed Might",
		somB2VaultSkywardOracle:  "Vault Skyward",
		somB2PassageOracle:       "Whitesun's Passage",
		somB2WingPunctureOracle:  "Wing Puncture",
		somB2HaltOrderOracle:     "Halt Order",
		somB2InstillOracle:       "Instill Infection",
		somB2DefianceOracle:      "Tel-Jilad Defiance",
		somB2TowerOracle:         "Tower of Calamities",
		somB2TumbleMagnetOracle:  "Tumble Magnet",
		somB2FlightBombOracle:    "Flight Spellbomb",
		somB2PanicBombOracle:     "Panic Spellbomb",
		somB2OriginBombOracle:    "Origin Spellbomb",
		somB2HorizonBombOracle:   "Horizon Spellbomb",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- shared test helpers --------------------------------------------------------------

// somB2Settle passes priority until the stack is empty, answering every
// "you may pay" the Spellbombs ask `who` with `pay`. It stops early on
// any other open prompt (a search), which the caller answers.
func somB2Settle(t *testing.T, g *game.Game, who uuid.UUID, pay bool) {
	t.Helper()
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		if !hasPayUnlessFor(g, who) {
			return
		}
		answerPayUnless(t, g, who, pay)
	}
}

// somB2Tap taps a permanent for a test's setup.
func somB2Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Tapped = true })
}

// --- printed keywords ------------------------------------------------------------------

func TestSomB2CardsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Contagious Nim", "Creature — Phyrexian Zombie", somB2NimOracle, []string{"infect"}},
		{"Cystbearer", "Creature — Phyrexian Beast", somB2CystbearerOracle, []string{"infect"}},
		{"Tel-Jilad Fallen", "Creature — Phyrexian Elf Warrior", somB2FallenOracle, []string{"protection from artifacts", "infect"}},
		{"Darkslick Drake", "Creature — Phyrexian Drake", somB2DarkslickOracle, []string{"flying"}},
		{"Ezuri's Archers", "Creature — Elf Archer", somB2ArchersOracle, []string{"reach"}},
		{"Flameborn Hellion", "Creature — Hellion", somB2HellionOracle, []string{"haste"}},
		{"Necrogen Scudder", "Creature — Phyrexian Horror", somB2ScudderOracle, []string{"flying"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !hasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

// --- metalcraft ------------------------------------------------------------------------

func TestSomB2MetalcraftTurnsOnAtThreeArtifacts(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
		boost                  int
		keyword                string
	}{
		{"Carapace Forger", "Creature — Elf Artificer", somB2ForgerOracle, 2, 2, 2, ""},
		{"Ghalma's Warden", "Creature — Elephant Soldier", somB2WardenOracle, 2, 4, 2, ""},
		{"Auriok Edgewright", "Creature — Human Soldier", somB2EdgewrightOracle, 2, 2, 0, "double strike"},
		{"Snapsail Glider", "Artifact Creature — Construct", somB2GliderOracle, 2, 2, 0, "flying"},
		{"Auriok Sunchaser", "Creature — Human Soldier", somB2SunchaserOracle, 1, 1, 2, "flying"},
		{"Ezuri's Brigade", "Creature — Elf Warrior", somB2BrigadeOracle, 4, 4, 4, "trample"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
		b12Permanent(g, g.Seats[1].ID, "Theirs", "Artifact")
		// Two artifacts you control, the card itself counted when it is one.
		have := 0
		if tc.name == "Snapsail Glider" {
			have = 1
		}
		for ; have < 2; have++ {
			b12Permanent(g, me.ID, "Trinket", "Artifact")
		}
		if p, tough := effectivePTOf(t, g, id); p != tc.power || tough != tc.tough {
			t.Errorf("%s with two artifacts is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.tough)
		}
		if tc.keyword != "" && hasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s has %s with two artifacts", tc.name, tc.keyword)
		}
		third := b12Permanent(g, me.ID, "Third", "Artifact")
		if p, tough := effectivePTOf(t, g, id); p != tc.power+tc.boost || tough != tc.tough+tc.boost {
			t.Errorf("%s with three artifacts is %d/%d, want %d/%d", tc.name, p, tough, tc.power+tc.boost, tc.tough+tc.boost)
		}
		if tc.keyword != "" && !hasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s lacks %s with three artifacts", tc.name, tc.keyword)
		}
		killCreature(t, g, me.ID, third)
		if p, tough := effectivePTOf(t, g, id); p != tc.power || tough != tc.tough {
			t.Errorf("%s back at two artifacts is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.tough)
		}
		if tc.keyword != "" && hasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s kept %s back at two artifacts", tc.name, tc.keyword)
		}
	}
}

// --- Equipment -------------------------------------------------------------------------

func TestSomB2EquipmentBoostsAndGrants(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		keyword      string
	}{
		{"Barbed Battlegear", somB2BattlegearOracle, 6, 2, ""},
		{"Strider Harness", somB2HarnessOracle, 3, 4, "haste"},
		{"Sylvok Lifestaff", somB2LifestaffOracle, 3, 3, ""},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
		eq := b12Push(g, me.ID, tc.name, "Artifact — Equipment", tc.oracle, 0, 0)
		fillPool(me, 2)
		equipTo(t, g, me.ID, eq, bear)
		if p, tough := effectivePTOf(t, g, bear); p != tc.power || tough != tc.tough {
			t.Errorf("%s: the creature is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.tough)
		}
		if tc.keyword != "" && !hasKeywordOnBattlefield(t, g, bear, tc.keyword) {
			t.Errorf("%s: the creature lacks %s", tc.name, tc.keyword)
		}
	}
}

func TestSomB2SylvokLifestaffGainsWhenTheCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	somB1Equip(t, g, me.ID, "Sylvok Lifestaff", somB2LifestaffOracle, bear)
	life := me.Life
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

// --- sacrifice-an-artifact creatures ----------------------------------------------------

func TestSomB2BarrageOgreSacrificesAnArtifactForTwoDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	ogre := b12Push(g, me.ID, "Barrage Ogre", "Creature — Ogre Warrior", somB2BarrageOgreOracle, 3, 3)
	art := b12Permanent(g, me.ID, "Trinket", "Artifact")
	life := opp.Life
	ravB2MustActivate(t, g, me.ID, ogre, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{playerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{art},
	})
	if g.Battlefield.Contains(art) {
		t.Error("the artifact was not sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2 damage", life, opp.Life)
	}
}

func TestSomB2FerrovoreSacrificesAnArtifactForPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fv := b12Push(g, me.ID, "Ferrovore", "Creature — Beast", somB2FerrovoreOracle, 2, 2)
	art := b12Permanent(g, me.ID, "Trinket", "Artifact")
	ravB2MustActivate(t, g, me.ID, fv, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{art}}, "R")
	if g.Battlefield.Contains(art) {
		t.Error("the artifact was not sacrificed")
	}
	if p, tough := effectivePTOf(t, g, fv); p != 5 || tough != 2 {
		t.Errorf("Ferrovore is %d/%d, want 5/2", p, tough)
	}
	dcCleanup(g)
	if p, _ := effectivePTOf(t, g, fv); p != 2 {
		t.Errorf("the pump outlived the cleanup step (power %d)", p)
	}
}

func TestSomB2OxiddaDaredevilSacrificesAnArtifactForHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dd := b12Push(g, me.ID, "Oxidda Daredevil", "Creature — Goblin Artificer", somB2DaredevilOracle, 2, 1)
	art := b12Permanent(g, me.ID, "Trinket", "Artifact")
	if hasKeywordOnBattlefield(t, g, dd, "haste") {
		t.Fatal("the Daredevil has haste before activating")
	}
	ravB2MustActivate(t, g, me.ID, dd, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{art}})
	if g.Battlefield.Contains(art) {
		t.Error("the artifact was not sacrificed")
	}
	if !hasKeywordOnBattlefield(t, g, dd, "haste") {
		t.Error("the Daredevil did not gain haste")
	}
}

// --- self grants and untaps -------------------------------------------------------------

func TestSomB2SaberclawGolemGainsFirstStrikeUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	golem := b12Push(g, me.ID, "Saberclaw Golem", "Artifact Creature — Golem", somB2SaberclawOracle, 4, 2)
	ravB2MustActivate(t, g, me.ID, golem, 0, game.ActivateAbilityParams{}, "R")
	if !hasKeywordOnBattlefield(t, g, golem, "first strike") {
		t.Error("the Golem did not gain first strike")
	}
	dcCleanup(g)
	if hasKeywordOnBattlefield(t, g, golem, "first strike") {
		t.Error("first strike outlived the cleanup step")
	}
}

func TestSomB2SolitonUntapsItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sol := b12Push(g, me.ID, "Soliton", "Artifact Creature — Construct", somB2SolitonOracle, 3, 4)
	somB2Tap(g, sol)
	ravB2MustActivate(t, g, me.ID, sol, 0, game.ActivateAbilityParams{}, "U")
	if isTapped(g, sol) {
		t.Error("Soliton is still tapped")
	}
}

func TestSomB2CorruptedHarvesterRegeneratesBySacrificingACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	harv := b12Push(g, me.ID, "Corrupted Harvester", "Creature — Phyrexian Horror", somB2HarvesterOracle, 6, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, harv, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}, "B")
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(harv); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(harv) {
		t.Error("the regenerated Harvester was destroyed")
	}
}

// --- dies and enters triggers -------------------------------------------------------------

func TestSomB2DarkslickDrakeDrawsAsItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drake := b12Push(g, me.ID, "Darkslick Drake", "Creature — Phyrexian Drake", somB2DarkslickOracle, 2, 4)
	hand := me.Hand.Size()
	killCreature(t, g, me.ID, drake)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestSomB2NecrogenScudderCostsThreeLifeAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	rfdCast(t, g, "Necrogen Scudder", "Creature — Phyrexian Horror", somB2ScudderOracle, 3, 3)
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, me.Life)
	}
}

func TestSomB2OxiddaScrapmelterDestroysAnArtifactAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	b12Permanent(g, me.ID, "Mine", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Theirs", "Artifact")
	rfdCast(t, g, "Oxidda Scrapmelter", "Creature — Beast", somB2ScrapmelterOracle, 3, 3)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, theirs) != game.ZoneGraveyard {
		t.Errorf("the artifact is in %q, want the graveyard", b12ZoneOf(g, theirs))
	}
}

func TestSomB2VulshokHeartstokerPumpsACreatureAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rfdCast(t, g, "Vulshok Heartstoker", "Creature — Human Shaman", somB2HeartstokerOracle, 2, 2)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
}

// --- combat ---------------------------------------------------------------------------------

func TestSomB2EzurisArchersPumpOnlyBlockingAFlier(t *testing.T) {
	for _, flier := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		archers := b12Push(g, opp.ID, "Ezuri's Archers", "Creature — Elf Archer", somB2ArchersOracle, 1, 2)
		var kws []string
		if flier {
			kws = []string{"flying"}
		}
		atk := pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Attacker", TypeLine: "Creature — Bird",
			Power: 1, Toughness: 1, Keywords: kws, Owner: me.ID, Controller: me.ID,
		})
		attackIntoBlocks(t, g, opp.ID, atk)
		if err := g.DeclareBlocker(archers, atk); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
		lockInBlocks(t, g)
		passPriorityAroundTable(t, g)
		want := 1
		if flier {
			want = 4
		}
		if p, tough := effectivePTOf(t, g, archers); p != want || tough != 2 {
			t.Errorf("flier=%v: the Archers are %d/%d, want %d/2", flier, p, tough, want)
		}
	}
}

func TestSomB2FlamebornHellionMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	hellion := b12Push(g, me.ID, "Flameborn Hellion", "Creature — Hellion", somB2HellionOracle, 5, 4)
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != hellion {
		t.Fatalf("pass with the Hellion home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(hellion, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Hellion attacking: %v", err)
	}
}

// --- sacrifice-this activations --------------------------------------------------------------

func TestSomB2NeurokReplicaBouncesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	rep := b12Push(g, me.ID, "Neurok Replica", "Artifact Creature — Wizard", somB2NeurokReplicaOracle, 1, 4)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, rep, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}, "C", "U")
	if g.Battlefield.Contains(rep) {
		t.Error("the Replica was not sacrificed")
	}
	if !opp.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestSomB2SalvageScoutReturnsAnArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scout := b12Push(g, me.ID, "Salvage Scout", "Creature — Human Scout", somB2SalvageScoutOracle, 1, 1)
	relic := pushGraveyardCardTyped(me, "Old Relic", "Artifact")
	bear := pushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	if err := ravB2Activate(t, g, me.ID, scout, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}, "W"); err == nil {
		t.Fatal("the Scout targeted a creature card")
	}
	ravB2MustActivate(t, g, me.ID, scout, 0, game.ActivateAbilityParams{Targets: cardRefs(relic)}, "W")
	if g.Battlefield.Contains(scout) {
		t.Error("the Scout was not sacrificed")
	}
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card did not return to hand")
	}
}

func TestSomB2SylvokReplicaDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	rep := b12Push(g, me.ID, "Sylvok Replica", "Artifact Creature — Shaman", somB2SylvokReplicaOracle, 1, 3)
	ench := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	ravB2MustActivate(t, g, me.ID, rep, 0, game.ActivateAbilityParams{Targets: cardRefs(ench)}, "G")
	if g.Battlefield.Contains(rep) {
		t.Error("the Replica was not sacrificed")
	}
	if b12ZoneOf(g, ench) != game.ZoneGraveyard {
		t.Errorf("the enchantment is in %q, want the graveyard", b12ZoneOf(g, ench))
	}
}

func TestSomB2VulshokReplicaDealsThreeToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	rep := b12Push(g, me.ID, "Vulshok Replica", "Artifact Creature — Berserker", somB2VulshokReplicaOracl, 3, 1)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := ravB2Activate(t, g, me.ID, rep, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}, "C", "R"); err == nil {
		t.Fatal("the Replica targeted a creature")
	}
	life := opp.Life
	ravB2MustActivate(t, g, me.ID, rep, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}}, "C", "R")
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want 3 damage", life, opp.Life)
	}
}

func TestSomB2MoriokReplicaDrawsTwoAndLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rep := b12Push(g, me.ID, "Moriok Replica", "Artifact Creature — Warrior", somB2MoriokReplicaOracle, 2, 2)
	hand, life := me.Hand.Size(), me.Life
	ravB2MustActivate(t, g, me.ID, rep, 0, game.ActivateAbilityParams{}, "C", "B")
	if g.Battlefield.Contains(rep) {
		t.Error("the Replica was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, me.Life)
	}
}

// --- instants ------------------------------------------------------------------------------

func TestSomB2SeizeTheInitiativePumpsAndGrantsFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Seize the Initiative", "Instant", somB2SeizeOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("the target lacks first strike")
	}
	dcCleanup(g)
	if p, _ := effectivePTOf(t, g, bear); p != 2 || hasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("the effect outlived the cleanup step")
	}
}

func TestSomB2UntamedMightPumpsByX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castFromHand(t, g, game.Card{Name: "Untamed Might", TypeLine: "Instant", OracleID: somB2UntamedMightOracle, ManaCost: "{X}{G}"},
		game.CastSpellParams{XValue: 3, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the target is %d/%d, want 5/5", p, tough)
	}
}

func TestSomB2VaultSkywardGrantsFlyingAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	somB2Tap(g, bear)
	castCatalogSpell(t, g, "Vault Skyward", "Instant", somB2VaultSkywardOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the target lacks flying")
	}
	if isTapped(g, bear) {
		t.Error("the target is still tapped")
	}
}

func TestSomB2WhitesunsPassageGainsFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castCatalogSpell(t, g, "Whitesun's Passage", "Instant", somB2PassageOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5", life, me.Life)
	}
}

func TestSomB2WingPunctureHitsOnlyAFlier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	ground := b12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 2)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	if err := castCatalogSpellErr(t, g, "Wing Puncture", "Instant", somB2WingPunctureOracle, cardRefs(bear, ground)); err == nil {
		t.Fatal("Wing Puncture targeted a creature without flying")
	}
	castCatalogSpell(t, g, "Wing Puncture", "Instant", somB2WingPunctureOracle, cardRefs(bear, flier))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("the flier survived 3 damage")
	}
	if b12Card(t, g, bear).DamageMarked != 0 {
		t.Error("the flier dealt damage back")
	}
}

func TestSomB2HaltOrderCountersAnArtifactSpellAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	toMainForCost(t, g)
	caster, responder := seatsForCombat(g)
	trinket := castCatalogSpell(t, g, "Trinket", "Artifact", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	hand := responder.Hand.Size()
	spell := handCardFull(responder, "Halt Order", "Instant", "", somB2HaltOrderOracle, []string{"U"})
	if err := g.CastSpell(responder.ID, spell, game.CastSpellParams{Targets: cardRefs(trinket)}); err != nil {
		t.Fatalf("casting Halt Order: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(trinket) || g.Battlefield.Contains(trinket) {
		t.Error("the artifact spell was not countered")
	}
	if !caster.Graveyard.Contains(trinket) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
	if got := responder.Hand.Size() - hand; got != 1 {
		t.Errorf("hand moved by %d, want 1 (the draw)", got)
	}
}

func TestSomB2InstillInfectionShrinksAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Instill Infection", "Instant", somB2InstillOracle, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if got := counterCount(g, ogre, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters = %d, want 1", got)
	}
	// castCatalogSpell puts the spell in hand and casts it from there, so
	// the hand ends one card up: the draw.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw", hand, me.Hand.Size())
	}
}

func TestSomB2TelJiladDefianceGrantsProtectionAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Tel-Jilad Defiance", "Instant", somB2DefianceOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "protection from artifacts") {
		t.Error("the target lacks protection from artifacts")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw", hand, me.Hand.Size())
	}
	dcCleanup(g)
	if hasKeywordOnBattlefield(t, g, bear, "protection from artifacts") {
		t.Error("protection outlived the cleanup step")
	}
}

// --- artifacts -------------------------------------------------------------------------------

func TestSomB2TowerOfCalamitiesDealsTwelve(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	tower := b12Push(g, me.ID, "Tower of Calamities", "Artifact", somB2TowerOracle, 0, 0)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 10, 13)
	ravB2MustActivate(t, g, me.ID, tower, 0, game.ActivateAbilityParams{Targets: cardRefs(big)},
		"C", "C", "C", "C", "C", "C", "C", "C")
	if got := b12Card(t, g, big).DamageMarked; got != 12 {
		t.Errorf("damage marked %d, want 12", got)
	}
	if !isTapped(g, tower) {
		t.Error("the Tower did not tap as the cost")
	}
}

func TestSomB2TumbleMagnetEntersWithThreeAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	magnet := rfdCast(t, g, "Tumble Magnet", "Artifact", somB2TumbleMagnetOracle, 0, 0)
	if got := countersOn(g, magnet, game.CounterCharge); got != 3 {
		t.Fatalf("entered with %d charge counters, want 3", got)
	}
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, magnet, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if !isTapped(g, bear) {
		t.Error("the target was not tapped")
	}
	if got := countersOn(g, magnet, game.CounterCharge); got != 2 {
		t.Errorf("charge counters %d, want 2", got)
	}
}

// --- Spellbombs ------------------------------------------------------------------------------

func TestSomB2SpellbombsDrawOnlyIfYouPay(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, colour string
		pay                  bool
	}{
		{"Flight Spellbomb", somB2FlightBombOracle, "U", true},
		{"Panic Spellbomb", somB2PanicBombOracle, "R", true},
		{"Origin Spellbomb", somB2OriginBombOracle, "W", true},
		{"Horizon Spellbomb", somB2HorizonBombOracle, "G", true},
		{"Flight Spellbomb", somB2FlightBombOracle, "U", false},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bomb := b12Push(g, me.ID, tc.name, "Artifact", tc.oracle, 0, 0)
		hand := me.Hand.Size()
		killCreature(t, g, me.ID, bomb)
		b06AddMana(me, tc.colour)
		passPriorityAroundTable(t, g)
		answerPayUnless(t, g, me.ID, tc.pay)
		passPriorityAroundTable(t, g)
		want := 0
		if tc.pay {
			want = 1
		}
		if got := me.Hand.Size() - hand; got != want {
			t.Errorf("%s, pay=%v: drew %d, want %d", tc.name, tc.pay, got, want)
		}
	}
}

func TestSomB2FlightSpellbombGrantsFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bomb := b12Push(g, me.ID, "Flight Spellbomb", "Artifact", somB2FlightBombOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := ravB2Activate(t, g, me.ID, bomb, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(bomb) {
		t.Error("the Spellbomb was not sacrificed")
	}
	somB2Settle(t, g, me.ID, false)
	if !hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the target lacks flying")
	}
}

func TestSomB2PanicSpellbombStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	bomb := b12Push(g, me.ID, "Panic Spellbomb", "Artifact", somB2PanicBombOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	if err := ravB2Activate(t, g, me.ID, bomb, 0, game.ActivateAbilityParams{Targets: cardRefs(wall)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	somB2Settle(t, g, me.ID, false)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err == nil {
		t.Error("the target creature blocked")
	}
}

func TestSomB2OriginSpellbombMakesAMyr(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bomb := b12Push(g, me.ID, "Origin Spellbomb", "Artifact", somB2OriginBombOracle, 0, 0)
	if err := ravB2Activate(t, g, me.ID, bomb, 0, game.ActivateAbilityParams{}, "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	somB2Settle(t, g, me.ID, false)
	if got := tokensNamedControlledBy(g, "Myr", me.ID); got != 1 {
		t.Errorf("Myr tokens = %d, want 1", got)
	}
}

func TestSomB2HorizonSpellbombFindsABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bomb := b12Push(g, me.ID, "Horizon Spellbomb", "Artifact", somB2HorizonBombOracle, 0, 0)
	stapleLibraryCard(me, "Urborg, Tomb of Yawgmoth", "Legendary Land")
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	if err := ravB2Activate(t, g, me.ID, bomb, 0, game.ActivateAbilityParams{}, "C", "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	somB2Settle(t, g, me.ID, false)
	answerSearchByID(t, g, me.ID, forest)
	somB2Settle(t, g, me.ID, false)
	if !me.Hand.Contains(forest) {
		t.Error("the basic land did not reach the hand")
	}
}
