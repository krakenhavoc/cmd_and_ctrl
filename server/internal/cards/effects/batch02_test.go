package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch02_test.go — card-level coverage for the card-coverage
// roadmap's batch 02 (#295, `edhrec_rank` 238–360). One test per
// observable behaviour, driven through a real cast / land play /
// activation rather than by calling primitives directly.

const (
	sundownPassOracle          = "5ad0b405-cca4-475e-985c-4d7e3599d87e"
	hauntedRidgeOracle         = "e2a37967-4212-4553-9f77-bcb613405807"
	overgrownFarmlandOracle    = "709d2f10-1585-48c3-9058-ddd5f62f0452"
	deathcapGladeOracle        = "f6d24565-5b32-4eff-b2e0-6e2c25516ff0"
	undergroundSeaOracle       = "4b22be3a-8ce1-47d1-b82e-6c3ccfb0548b"
	volcanicIslandOracle       = "c718911c-c955-4eb9-9e16-be4bd49a4e4e"
	tropicalIslandOracle       = "74b7fe23-5d3a-4092-8d78-7c0eba8f6f73"
	tundraOracle               = "02418479-9455-417f-a6a1-004356faff37"
	arcaneSanctumOracle        = "7d7cf15c-06b9-4062-a1eb-32614c458a3b"
	jungleShrineOracle         = "2e69537c-c898-4e13-a72d-ce3957a90304"
	ketriaTriomeOracle         = "6bae00e8-06cf-4ac4-a1cc-757e454109fe"
	simicGrowthChamberOracle   = "046f5783-cc7b-416a-8cf6-2bcef9c2cc1a"
	golgariRotFarmOracle       = "1b301478-b14f-4ef8-94e6-9647d582eabe"
	seatOfTheSynodOracle       = "39451b4d-cd7a-40da-b457-cb51b609173f"
	darksteelCitadelOracle     = "8dc067bf-f78f-4ac4-b6e7-b305c42cf0bc"
	scavengerGroundsOracle     = "5ece7d03-9ee7-4953-a06e-9d8e41874903"
	infernalGraspOracle        = "94f0a572-e91c-4b56-a5d1-6cbbeabd210d"
	witheringTormentOracle     = "ffce81c5-1b58-4882-a4e7-6f8d7cb170de"
	balefulStrixOracle         = "37688720-03de-4eca-a82d-a0afe8d58adc"
	entombOracle               = "299fc083-0834-4064-8344-f895aff68867"
	buriedAliveOracle          = "8203c621-a1a0-4865-8c9a-0d4064c86107"
	damnOracle                 = "b01d61cc-9844-4191-86a0-f2db6d42d6e5"
	snapOracle                 = "ac914d98-221e-426c-8a50-342896b15f9e"
	seethingSongOracle         = "64bf8929-f5f2-4d50-8667-13b1d007bcfc"
	manaGeyserOracle           = "a8dba58b-2956-492e-ae30-49db2ae68e53"
	harrowOracle               = "705509e9-a034-4a5a-9c65-66f58748b8a2"
	diabolicIntentOracle       = "038519b9-bca8-4b27-b5ac-2409595469d0"
	grayMerchantOracle         = "38f3b157-0df4-409b-89cc-086e1531cd5b"
	archmageEmeritusOracle     = "8305d576-21d8-4ce7-8eda-a7cd9793aca5"
	syrKonradOracle            = "14c3ff84-1e82-4606-a433-869fc52cc382"
	lotusCobraOracle           = "8ad91f64-ccab-4edc-bd54-b2ee9267d614"
	guardianProjectOracle      = "4f9e07ae-6341-4b46-9f77-f17ab659d266"
	loranOfTheThirdPathOracle  = "b3d81980-76f2-44e2-b1c9-01e30c726312"
	avengerOfZendikarOracle    = "4ba5b3f6-503b-43e6-b66e-4f8c55cffed7"
	craterhoofBehemothOracle   = "8c52bd39-0586-48ca-b263-17210cf9feb6"
	decanterOfEndlessWaterOrcl = "8ae98ef8-8f52-4877-a08c-1fae5514184e"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The four land
// cycles are loops over tables, so a transposed row is invisible
// until someone plays that exact card — this is the canary.
func TestBatch02CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		sundownPassOracle:          "Sundown Pass",
		hauntedRidgeOracle:         "Haunted Ridge",
		overgrownFarmlandOracle:    "Overgrown Farmland",
		deathcapGladeOracle:        "Deathcap Glade",
		undergroundSeaOracle:       "Underground Sea",
		volcanicIslandOracle:       "Volcanic Island",
		tropicalIslandOracle:       "Tropical Island",
		tundraOracle:               "Tundra",
		arcaneSanctumOracle:        "Arcane Sanctum",
		jungleShrineOracle:         "Jungle Shrine",
		ketriaTriomeOracle:         "Ketria Triome",
		simicGrowthChamberOracle:   "Simic Growth Chamber",
		golgariRotFarmOracle:       "Golgari Rot Farm",
		seatOfTheSynodOracle:       "Seat of the Synod",
		darksteelCitadelOracle:     "Darksteel Citadel",
		scavengerGroundsOracle:     "Scavenger Grounds",
		infernalGraspOracle:        "Infernal Grasp",
		witheringTormentOracle:     "Withering Torment",
		balefulStrixOracle:         "Baleful Strix",
		entombOracle:               "Entomb",
		buriedAliveOracle:          "Buried Alive",
		damnOracle:                 "Damn",
		snapOracle:                 "Snap",
		seethingSongOracle:         "Seething Song",
		manaGeyserOracle:           "Mana Geyser",
		harrowOracle:               "Harrow",
		diabolicIntentOracle:       "Diabolic Intent",
		grayMerchantOracle:         "Gray Merchant of Asphodel",
		archmageEmeritusOracle:     "Archmage Emeritus",
		syrKonradOracle:            "Syr Konrad, the Grim",
		lotusCobraOracle:           "Lotus Cobra",
		guardianProjectOracle:      "Guardian Project",
		loranOfTheThirdPathOracle:  "Loran of the Third Path",
		avengerOfZendikarOracle:    "Avenger of Zendikar",
		craterhoofBehemothOracle:   "Craterhoof Behemoth",
		decanterOfEndlessWaterOrcl: "Decanter of Endless Water",
	}
	if len(want) != 36 {
		t.Fatalf("the batch ships 36 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// Each new land row's printed colours, read straight off the Spec.
// A copy-paste that gives two rows the same pair is the failure this
// catches, and it is otherwise silent.
func TestBatch02LandRowsProduceTheirPrintedColours(t *testing.T) {
	want := map[string]string{
		// slowlands
		"Sundown Pass":       "{R|W}",
		"Haunted Ridge":      "{B|R}",
		"Overgrown Farmland": "{G|W}",
		"Deathcap Glade":     "{B|G}",
		// original duals
		"Underground Sea": "{U|B}",
		"Volcanic Island": "{U|R}",
		"Tropical Island": "{G|U}",
		"Tundra":          "{W|U}",
		// tri-lands
		"Arcane Sanctum": "{W|U|B}",
		"Jungle Shrine":  "{R|G|W}",
		// karoos (fixed pair, not a pipe)
		"Simic Growth Chamber": "{G}{U}",
		"Golgari Rot Farm":     "{B}{G}",
	}
	found := 0
	for _, spec := range All() {
		w, ok := want[spec.Name]
		if !ok {
			continue
		}
		found++
		if len(spec.ManaAbilities) != 1 {
			t.Errorf("%s: %d mana abilities, want 1", spec.Name, len(spec.ManaAbilities))
			continue
		}
		if got := spec.ManaAbilities[0].Produced; got != w {
			t.Errorf("%s: produces %s, want %s", spec.Name, got, w)
		}
	}
	if found != len(want) {
		t.Errorf("found %d of %d land rows", found, len(want))
	}
}

// --- lands: entry behaviour ----------------------------------------

func TestSundownPassEntersTappedWithOneOtherLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	id := playLandFromHand(t, g, "Sundown Pass", sundownPassOracle)
	top100AssertEnteredTapped(t, g, id, "Sundown Pass")
}

func TestDeathcapGladeEntersUntappedWithTwoOtherLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	id := playLandFromHand(t, g, "Deathcap Glade", deathcapGladeOracle)
	top100AssertEnteredUntapped(t, g, id, "Deathcap Glade")
}

// An original dual has no entry clause at all — the single thing that
// could go wrong is someone copy-pasting a tapland's Replacements
// slice onto it.
func TestUndergroundSeaEntersUntappedAndAlwaysHas(t *testing.T) {
	g := newCatalogGame(t)
	spec, _ := Lookup(undergroundSeaOracle)
	if len(spec.Replacements) != 0 {
		t.Errorf("Underground Sea declares %d replacements; it has no entry clause", len(spec.Replacements))
	}
	id := playLandFromHand(t, g, "Underground Sea", undergroundSeaOracle)
	top100AssertEnteredUntapped(t, g, id, "Underground Sea")
}

func TestArcaneSanctumEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Arcane Sanctum", arcaneSanctumOracle)
	top100AssertEnteredTapped(t, g, id, "Arcane Sanctum")
}

func TestKetriaTriomeEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Ketria Triome", ketriaTriomeOracle)
	top100AssertEnteredTapped(t, g, id, "Ketria Triome")
}

// The artifact lands are the two in the batch that must NOT enter
// tapped — that is most of why they are played.
func TestSeatOfTheSynodEntersUntappedAndTapsForBlue(t *testing.T) {
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Seat of the Synod", seatOfTheSynodOracle)
	top100AssertEnteredUntapped(t, g, id, "Seat of the Synod")

	me := g.Seats[g.Turn.ActiveSeat]
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U]", got)
	}
}

func TestDarksteelCitadelTapsForColorlessAndDeclaresIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Darksteel Citadel", "Artifact Land", darksteelCitadelOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	// The keyword is declared even though the engine does not yet
	// enforce it (#176) — so the day enforcement lands, the card is
	// already correct.
	if !eotHasAbility(effectiveAbilities(t, g, land), "indestructible") {
		t.Error("printed indestructible did not reach the effective abilities")
	}
}

// --- karoo lands ---------------------------------------------------

func TestSimicGrowthChamberEntersTappedAndBouncesALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	handBefore := me.Hand.Size()

	id := playLandFromHand(t, g, "Simic Growth Chamber", simicGrowthChamberOracle)
	top100AssertEnteredTapped(t, g, id, "Simic Growth Chamber")

	// The bounce is modelled as a target clause, so the controller
	// answers a pick_target prompt before the trigger reaches the
	// stack. See the declared simplification on karoo_lands.go.
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) {
		t.Error("Simic Growth Chamber did not return a land to hand")
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want +1 from the bounce", handBefore, me.Hand.Size())
	}
}

// --- Scavenger Grounds ---------------------------------------------

func TestScavengerGroundsExilesEveryGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := seedPermanentWithOracle(g, me.ID, "Scavenger Grounds", "Land — Desert", scavengerGroundsOracle)
	mine := batch01GraveyardCard(me, "Old Bear", "Creature — Bear")
	theirs := batch01GraveyardCard(opp, "Old Wolf", "Creature — Wolf")

	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{land},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The land is sacrificed on announce, as a cost.
	if g.Battlefield.Contains(land) {
		t.Error("the Desert sacrifice is a cost and should be paid at announce")
	}
	passPriorityAroundTable(t, g)

	if me.Graveyard.Contains(mine) {
		t.Error("the controller's own graveyard must be exiled too — the card is symmetric")
	}
	if opp.Graveyard.Contains(theirs) {
		t.Error("the opponent's graveyard was not exiled")
	}
	if !g.Exile.Contains(mine) || !g.Exile.Contains(theirs) {
		t.Error("exiled cards did not land in exile")
	}
}

// --- removal -------------------------------------------------------

func TestInfernalGraspKillsAndCostsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := seedCreature(g, "Big Bear", opp.ID)
	before := me.Life

	castCatalogSpell(t, g, "Infernal Grasp", "Instant", infernalGraspOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("Infernal Grasp did not destroy its target")
	}
	if me.Life != before-2 {
		t.Errorf("life %d → %d, want %d (life LOSS, not damage)", before, me.Life, before-2)
	}
}

func TestWitheringTormentAnswersAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := pushCatalogPermanent(g, opp.ID, "Rhystic Study", "Enchantment", "", false)
	before := me.Life

	castCatalogSpell(t, g, "Withering Torment", "Instant", witheringTormentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("Withering Torment did not destroy the enchantment")
	}
	if me.Life != before-2 {
		t.Errorf("life %d → %d, want %d", before, me.Life, before-2)
	}
}

func TestDamnKillsOneCreatureWhenCastForItsManaCost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := seedCreature(g, "Their Bear", opp.ID)
	survivor := seedCreature(g, "My Bear", me.ID)

	castCatalogSpell(t, g, "Damn", "Sorcery", damnOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("Damn did not destroy its target")
	}
	if !g.Battlefield.Contains(survivor) {
		t.Error("the targeted mode must not sweep the board")
	}
}

func TestDamnOverloadedSweepsEveryCreatureIncludingYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := seedCreature(g, "My Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)

	castWithAltCost(t, g, "Damn", "Sorcery", damnOracle, "overload")
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("overloaded Damn must destroy EACH creature, yours included")
	}
}

// --- tutors --------------------------------------------------------

func TestEntombPutsTheChosenCardInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Reanimation Target", "Something Else")

	castCatalogSpell(t, g, "Entomb", "Instant", entombOracle, nil)
	passPriorityAroundTable(t, g)

	answerSearchNamed(t, g, me.ID, "Reanimation Target")
	if !graveyardHasNamed(me, "Reanimation Target") {
		t.Error("Entomb did not put the chosen card into the graveyard")
	}
}

func TestBuriedAliveTakesThreeCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, n := range []string{"Body A", "Body B", "Body C", "Body D"} {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: n, TypeLine: "Creature — Zombie",
			Owner: me.ID, Controller: me.ID,
		})
	}

	castCatalogSpell(t, g, "Buried Alive", "Sorcery", buriedAliveOracle, nil)
	passPriorityAroundTable(t, g)

	answerSearchNamed(t, g, me.ID, "Body A", "Body B", "Body C")
	for _, n := range []string{"Body A", "Body B", "Body C"} {
		if !graveyardHasNamed(me, n) {
			t.Errorf("%s should be in the graveyard", n)
		}
	}
	if graveyardHasNamed(me, "Body D") {
		t.Error("Buried Alive takes up to THREE, not everything")
	}
}

func TestDiabolicIntentEatsACreatureAndTutorsToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := seedCreature(g, "Fodder", me.ID)
	seedLibrary(me, "The Combo Piece")

	castWithSacrifice(t, g, "Diabolic Intent", "Sorcery", diabolicIntentOracle, fodder)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice is an additional COST and is paid at announce")
	}
	passPriorityAroundTable(t, g)

	answerSearchNamed(t, g, me.ID, "The Combo Piece")
	if !handHasNamed(me, "The Combo Piece") {
		t.Error("Diabolic Intent did not put the chosen card into hand")
	}
}

func TestHarrowSacrificesALandAndFetchesTwoBasics(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	for _, n := range []string{"Island", "Mountain", "Plains"} {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: n, TypeLine: "Basic Land — " + n,
			Owner: me.ID, Controller: me.ID,
		})
	}

	castWithSacrifice(t, g, "Harrow", "Instant", harrowOracle, land)
	if g.Battlefield.Contains(land) {
		t.Error("Harrow's land sacrifice is a cost, paid at announce")
	}
	passPriorityAroundTable(t, g)

	answerSearchNamed(t, g, me.ID, "Island", "Mountain")
	fetched := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller != me.ID {
			continue
		}
		if c.Name == "Island" || c.Name == "Mountain" {
			fetched++
			if c.Tapped {
				t.Errorf("%s entered tapped; Harrow has no tapped clause", c.Name)
			}
		}
	}
	if fetched != 2 {
		t.Errorf("Harrow put %d lands onto the battlefield, want 2", fetched)
	}
}

// --- mana ----------------------------------------------------------

func TestSeethingSongAddsFiveRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	castCatalogSpell(t, g, "Seething Song", "Instant", seethingSongOracle, nil)
	passPriorityAroundTable(t, g)

	got := batch01PoolColors(me)
	if len(got) != 5 {
		t.Fatalf("pool %v, want five tokens", got)
	}
	for _, c := range got {
		if c != "R" {
			t.Errorf("pool %v, want all R", got)
			break
		}
	}
}

func TestManaGeyserCountsOnlyTappedOpponentLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]

	// Two tapped opponent lands, one untapped opponent land, one
	// tapped land of my own — only the first two count.
	for i := 0; i < 2; i++ {
		id := seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
		setTappedForTest(g, id, true)
	}
	seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	mine := seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	setTappedForTest(g, mine, true)

	castCatalogSpell(t, g, "Mana Geyser", "Sorcery", manaGeyserOracle, nil)
	passPriorityAroundTable(t, g)

	got := batch01PoolColors(me)
	if len(got) != 2 {
		t.Errorf("pool %v, want two R (two tapped opponent lands)", got)
	}
}

func TestLotusCobraTriggersOnALandDrop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Lotus Cobra", "Creature — Snake", lotusCobraOracle, false)

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)

	// "One mana of any color" is a pipe, so it lands as a colour
	// pick rather than straight into the pool.
	if pick := riderLatestManaPick(g, me.ID); pick == nil {
		t.Fatal("landfall did not queue the colour pick")
	}
}

func TestDecanterOfEndlessWaterLiftsTheHandSizeCapAndFixes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spec, _ := Lookup(decanterOfEndlessWaterOrcl)
	if !spec.NoMaxHandSize {
		t.Error("Decanter of Endless Water must declare NoMaxHandSize")
	}
	rock := seedPermanentWithOracle(g, me.ID, "Decanter of Endless Water", "Artifact", decanterOfEndlessWaterOrcl)
	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("with no commander the pick must offer all five colours, got %+v", pick)
	}
}

// --- Snap ----------------------------------------------------------

func TestSnapBouncesAndUntapsTwoOfYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := seedCreature(g, "Their Bear", opp.ID)
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		id := seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
		setTappedForTest(g, id, true)
		lands = append(lands, id)
	}

	castCatalogSpell(t, g, "Snap", "Instant", snapOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("Snap did not bounce its target")
	}
	untapped := 0
	for _, id := range lands {
		if c, ok := lookupForTest(g, id); ok && !c.Tapped {
			untapped++
		}
	}
	if untapped != 2 {
		t.Errorf("%d lands untapped, want exactly 2 (up to two)", untapped)
	}
}

// --- creatures -----------------------------------------------------

func TestBalefulStrixDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()

	id := castCatalogSpell(t, g, "Baleful Strix", "Artifact Creature — Bird", balefulStrixOracle, nil)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d → %d, want +1 from the ETB draw", before, me.Hand.Size())
	}
	abilities := effectiveAbilities(t, g, id)
	if !eotHasAbility(abilities, "flying") || !eotHasAbility(abilities, "deathtouch") {
		t.Errorf("abilities %v, want flying and deathtouch", abilities)
	}
}

func TestGrayMerchantDrainsForDevotionAndGainsTheTotal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// Two black pips already on the board, plus Gary's own {B}{B}
	// once he lands, is devotion 4. The hybrid counts too.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Pip", TypeLine: "Enchantment",
		ManaCost: "{1}{B}", Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hybrid Pip", TypeLine: "Enchantment",
		ManaCost: "{B/R}", Owner: me.ID, Controller: me.ID,
	})
	lifeBefore := me.Life
	oppLife := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			oppLife[p.ID] = p.Life
		}
	}

	// The Gary that resolves carries its own mana cost, so devotion
	// sees its {B}{B} as well: 2 + 2 = 4.
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Gray Merchant of Asphodel",
		TypeLine: "Creature — Zombie", OracleID: grayMerchantOracle,
		ManaCost: "{3}{B}{B}", Power: 2, Toughness: 4,
		Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	const wantX = 4
	drained := 0
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		if got := oppLife[p.ID] - p.Life; got != wantX {
			t.Errorf("opponent lost %d, want %d (devotion)", got, wantX)
		}
		drained += wantX
	}
	if me.Life != lifeBefore+drained {
		t.Errorf("life %d → %d, want +%d (the TOTAL lost, not X)", lifeBefore, me.Life, drained)
	}
}

func TestArchmageEmeritusDrawsOnEachInstantYouCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Archmage Emeritus", "Creature — Human Wizard", archmageEmeritusOracle, false)
	victim := seedCreature(g, "Their Bear", opp.ID)
	before := me.Hand.Size()

	// castCatalogSpell seeds the spell INTO hand first, so the
	// arithmetic is: +1 seeded, −1 cast, +1 magecraft = +1.
	castCatalogSpell(t, g, "Infernal Grasp", "Instant", infernalGraspOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d → %d, want %d (seeded, cast, magecraft draw)", before, me.Hand.Size(), before+1)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("sanity: the Grasp should still have resolved")
	}
}

func TestArchmageEmeritusIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Archmage Emeritus", "Creature — Human Wizard", archmageEmeritusOracle, false)
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Some Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)

	// +1 seeded, −1 cast, and nothing came back.
	if me.Hand.Size() != before {
		t.Errorf("hand %d → %d: magecraft must not fire on a creature spell", before, me.Hand.Size())
	}
}

func TestSyrKonradPingsWhenACreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Legendary Creature — Human Knight", syrKonradOracle, false)
	victim := seedCreature(g, "Doomed Bear", opp.ID)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}

	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(victim)
	})
	passPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		want := before[p.ID]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", p.Seat, p.Life, want)
		}
	}
}

func TestSyrKonradDoesNotPingOnHisOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	konrad := pushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Legendary Creature — Human Knight", syrKonradOracle, false)
	before := opp.Life

	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(konrad)
	})
	passPriorityAroundTable(t, g)

	if opp.Life != before {
		t.Errorf("opponent life %d → %d: the clause is ANOTHER creature", before, opp.Life)
	}
	_ = me
}

func TestSyrKonradMillsEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	konrad := pushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Legendary Creature — Human Knight", syrKonradOracle, false)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Library.Size()
	}

	if err := g.ActivateCatalogAbility(me.ID, konrad, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		if p.Library.Size() != before[p.ID]-1 {
			t.Errorf("seat %d library %d → %d, want -1", p.Seat, before[p.ID], p.Library.Size())
		}
	}
}

func TestGuardianProjectDrawsForAUniqueCreatureAndNotForARepeat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Guardian Project", "Enchantment", guardianProjectOracle, false)

	// castCatalogSpell seeds into hand, so a drawing cast is +1 net
	// (+1 seeded, −1 cast, +1 drawn) and a non-drawing cast is flat.
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Unique Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d → %d, want %d (seeded, cast, drew)", before, me.Hand.Size(), before+1)
	}

	// A second copy of the same name: the intervening-if is false.
	before = me.Hand.Size()
	castCatalogSpell(t, g, "Unique Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Errorf("hand %d → %d: a same-named creature must not draw", before, me.Hand.Size())
	}
}

func TestGuardianProjectIgnoresTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Guardian Project", "Enchantment", guardianProjectOracle, false)
	before := me.Hand.Size()

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, SpiritToken(), 1)
	})
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != before {
		t.Errorf("hand %d → %d: the clause is NONTOKEN", before, me.Hand.Size())
	}
}

func TestLoranDestroysAnArtifactOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	rock := pushCatalogPermanent(g, opp.ID, "Sol Ring", "Artifact", "", false)

	castAndResolveCreature(t, g, "Loran of the Third Path",
		"Legendary Creature — Human Artificer", loranOfTheThirdPathOracle)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("Loran's ETB did not destroy the chosen artifact")
	}
}

func TestLoranDrawsForYouAndTheTargetedOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Seeded rather than cast, so she is past summoning sickness —
	// the tap ability really is gated on it, which is printed
	// behaviour and not what this test is about.
	loran := pushCatalogPermanent(g, me.ID, "Loran of the Third Path",
		"Legendary Creature — Human Artificer", loranOfTheThirdPathOracle, false)

	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, loran, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != myHand+1 || opp.Hand.Size() != theirHand+1 {
		t.Errorf("hands %d→%d and %d→%d, want +1 each",
			myHand, me.Hand.Size(), theirHand, opp.Hand.Size())
	}
}

func TestAvengerOfZendikarMakesAPlantPerLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	}

	castCatalogSpell(t, g, "Avenger of Zendikar", "Creature — Elemental", avengerOfZendikarOracle, nil)
	passPriorityAroundTable(t, g)

	plants := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Plant" {
			plants++
			if c.CurrentPower() != 0 || c.CurrentToughness() != 1 {
				t.Errorf("Plant is %d/%d, want 0/1", c.CurrentPower(), c.CurrentToughness())
			}
		}
	}
	if plants != 3 {
		t.Errorf("%d Plants, want 3 (one per land)", plants)
	}
}

func TestCraterhoofPumpsAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := seedCreature(g, "Bear", me.ID)

	castCatalogSpell(t, g, "Craterhoof Behemoth", "Creature — Beast", craterhoofBehemothOracle, nil)
	passPriorityAroundTable(t, g)

	// Bear + Craterhoof = X of 2.
	c, ok := lookupForTest(g, bear)
	if !ok {
		t.Fatal("the bear vanished")
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("bear is %d/%d, want 4/4 (2/2 +2/+2)", c.CurrentPower(), c.CurrentToughness())
	}
	if !eotHasAbility(effectiveAbilities(t, g, bear), "trample") {
		t.Error("Craterhoof did not grant trample")
	}
}

// --- small local helpers -------------------------------------------

func graveyardHasNamed(p *game.Player, name string) bool {
	for _, c := range p.Graveyard.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func handHasNamed(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func setTappedForTest(g *game.Game, id uuid.UUID, tapped bool) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = tapped
				return
			}
		}
	})
}

func lookupForTest(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() {
		c, ok = g.LookupCardForEffect(id)
	})
	return c, ok
}
