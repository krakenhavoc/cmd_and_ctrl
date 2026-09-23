package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch02_test.go — card-level coverage for the card-coverage
// roadmap's batch 02 (#295, `edhrec_rank` 238–360): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation, land play or attack. Every batch-specific
// name is b02-prefixed — three batches are being written in parallel.

const (
	b02UndergroundSeaOracle      = "4b22be3a-8ce1-47d1-b82e-6c3ccfb0548b"
	b02VolcanicIslandOracle      = "c718911c-c955-4eb9-9e16-be4bd49a4e4e"
	b02TropicalIslandOracle      = "74b7fe23-5d3a-4092-8d78-7c0eba8f6f73"
	b02TundraOracle              = "02418479-9455-417f-a6a1-004356faff37"
	b02SundownPassOracle         = "5ad0b405-cca4-475e-985c-4d7e3599d87e"
	b02HauntedRidgeOracle        = "e2a37967-4212-4553-9f77-bcb613405807"
	b02OvergrownFarmlandOracle   = "709d2f10-1585-48c3-9058-ddd5f62f0452"
	b02DeathcapGladeOracle       = "f6d24565-5b32-4eff-b2e0-6e2c25516ff0"
	b02ArcaneSanctumOracle       = "7d7cf15c-06b9-4062-a1eb-32614c458a3b"
	b02JungleShrineOracle        = "2e69537c-c898-4e13-a72d-ce3957a90304"
	b02SimicGrowthChamberOracle  = "046f5783-cc7b-416a-8cf6-2bcef9c2cc1a"
	b02GolgariRotFarmOracle      = "1b301478-b14f-4ef8-94e6-9647d582eabe"
	b02SeatOfTheSynodOracle      = "39451b4d-cd7a-40da-b457-cb51b609173f"
	b02ScavengerGroundsOracle    = "5ece7d03-9ee7-4953-a06e-9d8e41874903"
	b02InfernalGraspOracle       = "94f0a572-e91c-4b56-a5d1-6cbbeabd210d"
	b02WitheringTormentOracle    = "ffce81c5-1b58-4882-a4e7-6f8d7cb170de"
	b02DamnOracle                = "b01d61cc-9844-4191-86a0-f2db6d42d6e5"
	b02SnapOracle                = "ac914d98-221e-426c-8a50-342896b15f9e"
	b02ArchmageEmeritusOracle    = "8305d576-21d8-4ce7-8eda-a7cd9793aca5"
	b02BalefulStrixOracle        = "37688720-03de-4eca-a82d-a0afe8d58adc"
	b02GuardianProjectOracle     = "4f9e07ae-6341-4b46-9f77-f17ab659d266"
	b02LotusCobraOracle          = "8ad91f64-ccab-4edc-bd54-b2ee9267d614"
	b02AvengerOfZendikarOracle   = "4ba5b3f6-503b-43e6-b66e-4f8c55cffed7"
	b02LoranOfTheThirdPathOracle = "b3d81980-76f2-44e2-b1c9-01e30c726312"
	b02SyrKonradOracle           = "14c3ff84-1e82-4606-a433-869fc52cc382"
)

// b02PushLand puts a land on the battlefield under owner, tapped or
// not, with no events (a fixture, not an entry).
func b02PushLand(g *game.Game, owner uuid.UUID, name string, tapped bool) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Land",
		Owner: owner, Controller: owner, Tapped: tapped,
	})
	return id
}

// b02TopOfLibrary puts a card with the given type line on top of a
// player's library (PushTop is the top: the last element).
func b02TopOfLibrary(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func b02ToMainPhase(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// --- registration --------------------------------------------------

func TestBatch02CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b02UndergroundSeaOracle:      "Underground Sea",
		b02VolcanicIslandOracle:      "Volcanic Island",
		b02TropicalIslandOracle:      "Tropical Island",
		b02TundraOracle:              "Tundra",
		b02SundownPassOracle:         "Sundown Pass",
		b02HauntedRidgeOracle:        "Haunted Ridge",
		b02OvergrownFarmlandOracle:   "Overgrown Farmland",
		b02DeathcapGladeOracle:       "Deathcap Glade",
		b02ArcaneSanctumOracle:       "Arcane Sanctum",
		b02JungleShrineOracle:        "Jungle Shrine",
		b02SimicGrowthChamberOracle:  "Simic Growth Chamber",
		b02GolgariRotFarmOracle:      "Golgari Rot Farm",
		b02SeatOfTheSynodOracle:      "Seat of the Synod",
		b02ScavengerGroundsOracle:    "Scavenger Grounds",
		b02InfernalGraspOracle:       "Infernal Grasp",
		b02WitheringTormentOracle:    "Withering Torment",
		b02DamnOracle:                "Damn",
		b02SnapOracle:                "Snap",
		b02ArchmageEmeritusOracle:    "Archmage Emeritus",
		b02BalefulStrixOracle:        "Baleful Strix",
		b02GuardianProjectOracle:     "Guardian Project",
		b02LotusCobraOracle:          "Lotus Cobra",
		b02AvengerOfZendikarOracle:   "Avenger of Zendikar",
		b02LoranOfTheThirdPathOracle: "Loran of the Third Path",
		b02SyrKonradOracle:           "Syr Konrad, the Grim",
	}
	if len(want) != 25 {
		t.Fatalf("the batch is 25 cards, the table lists %d", len(want))
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

// The twelve land rows across four tables, each producing its printed
// mana — the loop-leak canary every cycle carries.
func TestBatch02LandRowsProduceTheirPrintedMana(t *testing.T) {
	type row struct {
		produced     string
		replacements int
		triggers     int
	}
	rows := map[string]row{
		"Sundown Pass":         {"{R|W}", 1, 0},
		"Haunted Ridge":        {"{B|R}", 1, 0},
		"Overgrown Farmland":   {"{G|W}", 1, 0},
		"Deathcap Glade":       {"{B|G}", 1, 0},
		"Underground Sea":      {"{U|B}", 0, 0},
		"Volcanic Island":      {"{U|R}", 0, 0},
		"Tropical Island":      {"{G|U}", 0, 0},
		"Tundra":               {"{W|U}", 0, 0},
		"Arcane Sanctum":       {"{W|U|B}", 1, 0},
		"Jungle Shrine":        {"{R|G|W}", 1, 0},
		"Simic Growth Chamber": {"{G}{U}", 1, 1},
		"Golgari Rot Farm":     {"{B}{G}", 1, 1},
	}
	found := 0
	for _, spec := range All() {
		want, ok := rows[spec.Name]
		if !ok {
			continue
		}
		found++
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", spec.Name, spec.ManaAbilities, want.produced)
		}
		if len(spec.Replacements) != want.replacements {
			t.Errorf("%s: %d replacements, want %d", spec.Name, len(spec.Replacements), want.replacements)
		}
		if len(spec.Triggered) != want.triggers {
			t.Errorf("%s: %d triggers, want %d", spec.Name, len(spec.Triggered), want.triggers)
		}
	}
	if found != len(rows) {
		t.Errorf("found %d of %d land rows", found, len(rows))
	}
}

// --- lands ---------------------------------------------------------

func TestHauntedRidgeEntersTappedWithOneOtherLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	id := playLandFromHand(t, g, "Haunted Ridge", b02HauntedRidgeOracle)
	top100AssertEnteredTapped(t, g, id, "Haunted Ridge")
}

func TestDeathcapGladeEntersUntappedWithTwoOtherLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	id := playLandFromHand(t, g, "Deathcap Glade", b02DeathcapGladeOracle)
	top100AssertEnteredUntapped(t, g, id, "Deathcap Glade")
}

// A dual has no drawback: untapped, and either colour on demand.
func TestUndergroundSeaEntersUntappedAndOffersBothColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playLandFromHand(t, g, "Underground Sea", b02UndergroundSeaOracle)
	top100AssertEnteredUntapped(t, g, id, "Underground Sea")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("a dual must ask U or B, got %+v", pick)
	}
}

func TestArcaneSanctumEntersTappedAndOffersThreeColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playLandFromHand(t, g, "Arcane Sanctum", b02ArcaneSanctumOracle)
	top100AssertEnteredTapped(t, g, id, "Arcane Sanctum")

	ready := seedPermanentWithOracle(g, me.ID, "Arcane Sanctum", "Land", b02ArcaneSanctumOracle)
	if err := g.ActivateManaAbility(me.ID, ready, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 3 {
		t.Fatalf("a tri-land must ask W, U or B, got %+v", pick)
	}
}

// #1337: the bounce is a resolution-time choice (ReturnOneYouControl
// / ChoosePermanents), not a target picked when the trigger goes on
// the stack.
func TestSimicGrowthChamberEntersTappedAndBouncesAChosenLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	id := playLandFromHand(t, g, "Simic Growth Chamber", b02SimicGrowthChamberOracle)
	top100AssertEnteredTapped(t, g, id, "Simic Growth Chamber")

	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Simic Growth Chamber queued no return pick")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}

	if !me.Hand.Contains(forest) {
		t.Error("the chosen land should be back in hand")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Chamber itself should stay")
	}
}

func TestSeatOfTheSynodIsAnArtifactLandThatTapsForBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seat := seedPermanentWithOracle(g, me.ID, "Seat of the Synod", "Artifact Land", b02SeatOfTheSynodOracle)
	card, _ := battlefieldCard(g, seat)
	if !card.IsArtifact() || !card.IsLand() {
		t.Fatalf("Seat of the Synod must be both an artifact and a land: %q", card.TypeLine)
	}
	if err := g.ActivateManaAbility(me.ID, seat, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U]", got)
	}
}

func TestScavengerGroundsSacrificesItselfToExileAllGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grounds := pushCatalogPermanent(g, me.ID, "Scavenger Grounds", "Land — Desert", b02ScavengerGroundsOracle, false)
	myDead := pushGraveyardCardForTest(me, "My Dead")
	theirDead := pushGraveyardCardForTest(opp, "Their Dead")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, grounds, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{grounds},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(grounds) {
		t.Error("the Grounds is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{myDead, theirDead} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s should be in exile", id)
		}
	}
}

func TestScavengerGroundsRefusesANonDesert(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grounds := pushCatalogPermanent(g, me.ID, "Scavenger Grounds", "Land — Desert", b02ScavengerGroundsOracle, false)
	forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, grounds, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{forest},
	}); err == nil {
		t.Error("a Forest was accepted for 'Sacrifice a Desert'")
	}
	if !g.Battlefield.Contains(grounds) || !g.Battlefield.Contains(forest) {
		t.Error("a refused activation must pay nothing")
	}
}

// --- removal -------------------------------------------------------

func TestInfernalGraspDestroysAndCostsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedCreature(g, "Bear", opp.ID)
	before := me.Life
	castCatalogSpell(t, g, "Infernal Grasp", "Instant", b02InfernalGraspOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bear) {
		t.Error("the creature should be in its owner's graveyard")
	}
	if me.Life != before-2 {
		t.Errorf("caster life %d -> %d, want -2", before, me.Life)
	}
}

func TestWitheringTormentDestroysAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glory := seedPermanentFor(g, opp.ID, "Glory", "Enchantment")
	before := me.Life
	castCatalogSpell(t, g, "Withering Torment", "Instant", b02WitheringTormentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: glory}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(glory) {
		t.Error("the enchantment survived")
	}
	if me.Life != before-2 {
		t.Errorf("caster life %d -> %d, want -2", before, me.Life)
	}
}

func TestWitheringTormentRefusesAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	b02ToMainPhase(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Withering Torment", TypeLine: "Instant",
		OracleID: b02WitheringTormentOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err == nil {
		t.Error("an artifact was accepted for 'target creature or enchantment'")
	}
}

func TestDamnDestroysOneCreatureWhenCastForBlack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedCreature(g, "Mine", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	castCatalogSpell(t, g, "Damn", "Sorcery", b02DamnOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target survived")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the single-target cast must not sweep")
	}
}

func TestDamnOverloadedDestroysEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedCreature(g, "Mine", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	b02ToMainPhase(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Damn", TypeLine: "Sorcery",
		OracleID: b02DamnOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overloaded cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{mine, theirs} {
		if g.Battlefield.Contains(c) {
			t.Errorf("%s survived the overloaded Damn", c)
		}
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("Damn overloaded is creatures only")
	}
}

func TestSnapBouncesAndUntapsTwoOfYourTappedLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b02PushLand(g, me.ID, "Island A", true)
	b := b02PushLand(g, me.ID, "Island B", true)
	c := b02PushLand(g, me.ID, "Island C", true)
	theirs := b02PushLand(g, opp.ID, "Their Island", true)
	bear := seedCreature(g, "Bear", opp.ID)

	castCatalogSpell(t, g, "Snap", "Instant", b02SnapOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if !opp.Hand.Contains(bear) {
		t.Error("the creature should be back in its owner's hand")
	}
	answerChooseCards(t, g, me.ID, a, b) // 3 tapped lands offered, pick 2
	untapped := 0
	for _, id := range []uuid.UUID{a, b, c} {
		if card, _ := battlefieldCard(g, id); !card.Tapped {
			untapped++
		}
	}
	if untapped != 2 {
		t.Errorf("%d of my lands untapped, want exactly 2", untapped)
	}
	if card, _ := battlefieldCard(g, theirs); !card.Tapped {
		t.Error("an opponent's land was untapped")
	}
}

// --- triggers ------------------------------------------------------

func TestArchmageEmeritusDrawsOnAnInstantNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Archmage Emeritus", "Creature — Human Wizard", b02ArchmageEmeritusOracle, false)
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want +1 for the instant", before, got)
	}

	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after a creature spell, want unchanged at %d", got, before+1)
	}
}

func TestBalefulStrixFliesHasDeathtouchAndDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	strix := castCatalogSpell(t, g, "Baleful Strix", "Artifact Creature — Bird", b02BalefulStrixOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(strix) {
		t.Fatal("the Strix did not resolve to the battlefield")
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want +1 from the ETB", before, got)
	}
	abilities := effectiveAbilities(t, g, strix)
	if !eotHasAbility(abilities, "flying") || !eotHasAbility(abilities, "deathtouch") {
		t.Errorf("abilities %v, want flying and deathtouch", abilities)
	}
}

func TestGuardianProjectDrawsOnlyForAFreshName(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Guardian Project", "Enchantment", b02GuardianProjectOracle, false)
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("hand %d -> %d, want +1 for a fresh name", before, got)
	}

	// A second creature with the same name as one you control: no draw.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after a duplicate name, want unchanged", got)
	}

	// A name matching a creature card in your graveyard: no draw.
	pushGraveyardCardForTest(me, "Ghost")
	castCatalogSpell(t, g, "Ghost", "Creature — Spirit", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after a name in the graveyard, want unchanged", got)
	}

	// A token: no draw.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after a token, want unchanged", got)
	}
}

func TestLotusCobraAddsManaOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Lotus Cobra", "Creature — Snake", b02LotusCobraOracle, false)

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)

	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("landfall should queue an any-colour pick, got %+v", pick)
	}
}

func TestAvengerOfZendikarMakesPlantsThenGrowsThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLandOnBattlefield(g, me.ID, "Forest A", "Basic Land — Forest")
	seedLandOnBattlefield(g, me.ID, "Forest B", "Basic Land — Forest")
	seedLandOnBattlefield(g, me.ID, "Forest C", "Basic Land — Forest")
	seedLandOnBattlefield(g, g.Seats[1].ID, "Their Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Avenger of Zendikar", "Creature — Elemental", b02AvengerOfZendikarOracle, nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Plant"); n != 3 {
		t.Fatalf("%d Plants, want 3 (one per land YOU control)", n)
	}

	playLandFromHand(t, g, "Forest D", "")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Plant" {
			continue
		}
		if c.Counters["+1/+1"] != 1 {
			t.Errorf("a Plant has %d +1/+1 counters after landfall, want 1", c.Counters["+1/+1"])
		}
	}
}

func TestLoranDestroysAnArtifactOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")

	castCatalogSpell(t, g, "Loran of the Third Path", "Legendary Creature — Human Artificer", b02LoranOfTheThirdPathOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived Loran's ETB")
	}
}

func TestLoranTapsToDrawForYouAndAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loran := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loran of the Third Path",
		TypeLine: "Legendary Creature — Human Artificer", OracleID: b02LoranOfTheThirdPathOracle,
		Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if !eotHasAbility(effectiveAbilities(t, g, loran), "vigilance") {
		t.Error("printed vigilance did not reach the effective abilities")
	}
	meBefore, oppBefore := me.Hand.Size(), opp.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, loran, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != meBefore+1 || opp.Hand.Size() != oppBefore+1 {
		t.Errorf("hands me %d->%d opp %d->%d, want +1 each", meBefore, me.Hand.Size(), oppBefore, opp.Hand.Size())
	}
}

func TestLoranTapAbilityRefusesYourself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	loran := pushCatalogPermanent(g, me.ID, "Loran of the Third Path", "Legendary Creature — Human Artificer", b02LoranOfTheThirdPathOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, loran, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Error("the controller was accepted for 'target opponent'")
	}
}

func TestSyrKonradPingsOnDeathMillAndLeavingYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Legendary Creature — Human Knight", b02SyrKonradOracle, false)
	before := lifeOfOpponents(g)
	assertDrained := func(step string, n int) {
		t.Helper()
		for i, b := range before {
			if got := g.Seats[i+1].Life; got != b-n {
				t.Errorf("%s: opponent %d %d -> %d, want -%d", step, i+1, b, got, n)
			}
		}
	}

	// Another creature dies (an opponent's).
	bear := seedCreature(g, "Bear", opp.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	assertDrained("death", 1)

	// A creature card milled from a library — not from the battlefield.
	b02TopOfLibrary(opp, "Milled Bear", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	assertDrained("mill", 2)

	// A milled non-creature is silent.
	b02TopOfLibrary(opp, "Milled Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	assertDrained("mill noncreature", 2)

	// A creature card leaving YOUR graveyard.
	mine := pushGraveyardCardForTest(me, "Old Bear")
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(mine) })
	passPriorityAroundTable(t, g)
	assertDrained("left my graveyard", 3)

	// A creature card leaving an OPPONENT's graveyard is not yours.
	theirs := pushGraveyardCardForTest(opp, "Their Old Bear")
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs) })
	passPriorityAroundTable(t, g)
	assertDrained("left their graveyard", 3)
}

func TestSyrKonradMillsEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	konrad := pushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Legendary Creature — Human Knight", b02SyrKonradOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"})
	sizes := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		sizes[i] = p.Library.Size()
	}
	if err := g.ActivateCatalogAbility(me.ID, konrad, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := p.Library.Size(); got != sizes[i]-1 {
			t.Errorf("seat %d library %d -> %d, want -1", i, sizes[i], got)
		}
	}
}
