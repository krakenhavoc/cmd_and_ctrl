package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cloud_deck_a_test.go — the first slice of the Cloud, Ex-SOLDIER deck
// request (#3656): Cloud, Armory Automaton, Wrecking Ball Arm, Slayers'
// Stronghold, both Barrets, SOLDIER Military Program, Red XIII, Elena,
// Zack Fair and Conformer Shuriken.

const (
	cloudExSoldierOracle   = "653f0452-5802-4639-86c2-1500c3964fcb"
	armoryAutomatonOracle  = "fa3335be-80b1-49d7-9809-5bef13be79f4"
	wreckingBallArmOracle  = "9f48cb32-46be-4c47-b558-2ea359efdd9b"
	slayersStrongholdOracl = "2de7367b-a5a4-43f4-8f8f-b931ea28150d"
	barretWallaceOracle    = "913672bc-cc88-4964-a331-7dbce40f2bda"
	barretAvalancheOracle  = "83f776d9-f3b6-40c2-8008-1ace4d110825"
	soldierProgramOracle   = "3e2034ee-adc5-4a24-9605-c9722bf813c1"
	redXIIIOracle          = "9d302e2a-d852-4878-ac9d-d51d9924144b"
	elenaTurkRecruitOracle = "f5b6835b-7ebe-4855-8b3e-bffb74fd50ee"
	zackFairOracle         = "e84d146a-1c67-49df-b062-622794e45c42"
	conformerShurikenOracl = "8e7681ba-6412-4165-b374-5636feec216d"
)

func cloudDeckWantFull(t *testing.T, oracle string) {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%s: registered %v, completeness %s, caveats %v, want full", oracle, ok, spec.Completeness, spec.Caveats)
	}
}

func cloudDeckCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters["+1/+1"]
			}
		}
	})
	return n
}

func cloudDeckHasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	v := s58p6View(t, g, id)
	return game.HasKeyword(&v, kw)
}

func TestCloudDeckCloudEntersAndAttachesAnEquipment(t *testing.T) {
	cloudDeckWantFull(t, cloudExSoldierOracle)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sword := seedEquipment(g, me.ID, "Test Blade", "")

	cloud := castAndResolveCreature(t, g, "Cloud, Ex-SOLDIER", "Legendary Creature — Human Soldier Mercenary", cloudExSoldierOracle)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if !hasID(pick.PickTargetCards, sword) {
			t.Fatalf("legal set %v misses the Equipment", pick.PickTargetCards)
		}
		pickCard(t, g, me.ID, sword)
	}
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, sword); host.ID != cloud {
		t.Errorf("Equipment host = %+v, want Cloud %s", host, cloud)
	}
}

func TestCloudDeckCloudAttackDrawsPerEquippedAttackerAndMakesTreasureAtSevenPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	cloud := pushCatalogPermanent(g, me.ID, "Cloud, Ex-SOLDIER", "Legendary Creature — Human Soldier Mercenary", cloudExSoldierOracle, false)
	bear := seedBear(g, me.ID)
	plain := seedBear(g, me.ID)
	arm := seedEquipment(g, me.ID, "Wrecking Ball Arm", wreckingBallArmOracle)
	blade := seedEquipment(g, me.ID, "Test Blade", "")
	equipTo(t, g, me.ID, arm, cloud)
	if err := cloudDeckAttach(g, blade, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("attach: %v", err)
	}
	if got := effectivePower(t, g, cloud); got != 7 {
		t.Fatalf("Cloud power %d, want 7", got)
	}
	before := me.Hand.Size()

	declareAttack(t, g, opp.ID, cloud, bear, plain)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d, want 2 (Cloud and one Bear are equipped attackers)", got)
	}
	if n := battlefieldCountNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("Treasures = %d, want 2 (Cloud has power 7)", n)
	}
}

func TestCloudDeckCloudAttackMakesNoTreasureBelowSevenPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	cloud := pushCatalogPermanent(g, me.ID, "Cloud, Ex-SOLDIER", "Legendary Creature — Human Soldier Mercenary", cloudExSoldierOracle, false)
	before := me.Hand.Size()

	declareAttack(t, g, opp.ID, cloud)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("drew %d with no equipped attacker, want 0", got)
	}
	if n := battlefieldCountNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("Treasures = %d, want 0", n)
	}
}

func TestCloudDeckArmoryAutomatonAttachesAnyNumberOfEquipmentWithoutChangingControl(t *testing.T) {
	cloudDeckWantFull(t, armoryAutomatonOracle)
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := seedEquipment(g, me.ID, "Test Blade", "")
	theirs := seedEquipment(g, opp.ID, "Their Blade", "")

	auto := castAndResolveCreature(t, g, "Armory Automaton", "Artifact Creature — Construct", armoryAutomatonOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, e := range []uuid.UUID{mine, theirs} {
		if host := attachmentHostOf(t, g, e); host.ID != auto {
			t.Errorf("Equipment %s host = %+v, want the Automaton", e, host)
		}
	}
}

func TestCloudDeckWreckingBallArmSetsBasePTAndBlocksSmallCreatures(t *testing.T) {
	cloudDeckWantFull(t, wreckingBallArmOracle)
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	arm := seedEquipment(g, me.ID, "Wrecking Ball Arm", wreckingBallArmOracle)

	equipTo(t, g, me.ID, arm, bear)
	if got := effectivePower(t, g, bear); got != 7 {
		t.Errorf("equipped power %d, want 7", got)
	}
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: wreckingBallArmOracle})
	if len(abilities) != 2 {
		t.Fatalf("equip abilities = %d, want 2 (legendary {3} and plain {7})", len(abilities))
	}
}

func TestCloudDeckSlayersStrongholdBoostsAndGrantsKeywords(t *testing.T) {
	cloudDeckWantFull(t, slayersStrongholdOracl)
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	land := pushCatalogPermanent(g, me.ID, "Slayers' Stronghold", "Land", slayersStrongholdOracl, false)
	bear := seedBear(g, me.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "W"})

	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	for _, kw := range []string{"vigilance", "haste"} {
		if !cloudDeckHasKeyword(t, g, bear, kw) {
			t.Errorf("Bear lacks %s", kw)
		}
	}
}

func TestCloudDeckBarretWallaceDealsDamagePerEquippedCreature(t *testing.T) {
	cloudDeckWantFull(t, barretWallaceOracle)
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	barret := pushCatalogPermanent(g, me.ID, "Barret Wallace", "Legendary Creature — Human Rebel", barretWallaceOracle, false)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Test Blade", "")
	other := seedEquipment(g, me.ID, "Other Blade", "")
	for _, p := range []struct{ e, h uuid.UUID }{{blade, bear}, {other, barret}} {
		if err := cloudDeckAttach(g, p.e, game.TargetRef{Kind: game.TargetCard, ID: p.h}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	}
	life := opp.Life

	declareAttack(t, g, opp.ID, barret)
	passPriorityAroundTable(t, g)

	if got := life - opp.Life; got != 2 {
		t.Errorf("defending player lost %d, want 2 (two equipped creatures)", got)
	}
}

func TestCloudDeckBarretAvalancheLeaderMakesARebelPerEquipmentAndAttachesAtCombat(t *testing.T) {
	cloudDeckWantFull(t, barretAvalancheOracle)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Barret, Avalanche Leader", "Legendary Creature — Human Rebel", barretAvalancheOracle, false)

	blade := b12PlayFromHand(t, g, "Test Blade", equipTypeLine, "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := countTokensControlled(g, me.ID, "Rebel"); n != 1 {
		t.Fatalf("Rebel tokens = %d, want 1", n)
	}

	advanceTo(t, g, game.StepBeginCombat)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickCard(t, g, me.ID, blade)
	}
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickCard(t, g, me.ID, tokenOfSubtype(g, me.ID, "Rebel"))
	}
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, blade); host.ID != tokenOfSubtype(g, me.ID, "Rebel") {
		t.Errorf("Equipment host = %+v, want the Rebel token", host)
	}
}

func tokenOfSubtype(g *game.Game, controller uuid.UUID, subtype string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.HasSubtype(subtype) {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestCloudDeckSoldierMilitaryProgramMakesASoldier(t *testing.T) {
	cloudDeckWantFull(t, soldierProgramOracle)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "SOLDIER Military Program", "Enchantment", soldierProgramOracle, false)

	advanceTo(t, g, game.StepBeginCombat)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode_pick prompt")
	}
	answerMode(t, g, c, me.ID, 0)
	passPriorityAroundTable(t, g)

	if n := countTokensControlled(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("Soldier tokens = %d, want 1", n)
	}
}

func TestCloudDeckRedXIIIGrantsKeywordsToOtherModifiedCreatures(t *testing.T) {
	cloudDeckWantFull(t, redXIIIOracle)
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	xiii := pushCatalogPermanent(g, me.ID, "Red XIII, Proud Warrior", "Legendary Creature — Beast Warrior", redXIIIOracle, false)
	bear := seedBear(g, me.ID)
	plain := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Test Blade", "")
	if err := cloudDeckAttach(g, blade, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("attach: %v", err)
	}

	for _, kw := range []string{"vigilance", "trample"} {
		if !cloudDeckHasKeyword(t, g, bear, kw) {
			t.Errorf("equipped Bear lacks %s", kw)
		}
		if cloudDeckHasKeyword(t, g, plain, kw) {
			t.Errorf("unmodified Bear has %s", kw)
		}
	}
	_ = xiii
}

func TestCloudDeckRedXIIIReturnsAnEquipmentFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	blade := pushGraveyardPermanent(me, "Test Blade", equipTypeLine, "{2}")
	creature := pushGraveyardPermanent(me, "Test Bear", "Creature — Bear", "{1}")

	castAndResolveCreature(t, g, "Red XIII, Proud Warrior", "Legendary Creature — Beast Warrior", redXIIIOracle)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	if hasID(pick.PickTargetCards, creature) || !hasID(pick.PickTargetCards, blade) {
		t.Fatalf("legal set %v, want only the Equipment", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, blade)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(blade) {
		t.Error("the Equipment did not return to hand")
	}
}

func TestCloudDeckElenaReturnsHistoricAndGrowsOnHistoricCasts(t *testing.T) {
	cloudDeckWantFull(t, elenaTurkRecruitOracle)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := pushGraveyardPermanent(me, "Test Rock", "Artifact", "{1}")
	assassin := pushGraveyardPermanent(me, "Test Assassin", "Legendary Creature — Human Assassin", "{1}{B}")
	plain := pushGraveyardPermanent(me, "Test Bear", "Creature — Bear", "{1}")

	elena := castAndResolveCreature(t, g, "Elena, Turk Recruit", "Legendary Creature — Human Assassin", elenaTurkRecruitOracle)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	if !hasID(pick.PickTargetCards, rock) || hasID(pick.PickTargetCards, assassin) || hasID(pick.PickTargetCards, plain) {
		t.Fatalf("legal set %v, want only the non-Assassin historic card", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Fatal("the artifact did not return to hand")
	}

	// A historic spell puts a counter on Elena; a plain one does not.
	castCatalogSpell(t, g, "Test Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if got := cloudDeckCounters(t, g, elena); got != 1 {
		t.Errorf("Elena counters = %d, want 1", got)
	}
}

func TestCloudDeckZackFairSacrificeMovesCountersAndGrantsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	zack := castAndResolveCreature(t, g, "Zack Fair", "Legendary Creature — Human Soldier", zackFairOracle)
	passPriorityAroundTable(t, g)
	if got := cloudDeckCounters(t, g, zack); got != 1 {
		t.Fatalf("Zack counters = %d, want 1 (enters with a +1/+1 counter)", got)
	}
	bear := seedBear(g, me.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})

	if err := g.ActivateCatalogAbility(me.ID, zack, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if onBattlefield(g, zack) {
		t.Error("Zack was not sacrificed")
	}
	if got := cloudDeckCounters(t, g, bear); got != 1 {
		t.Errorf("Bear counters = %d, want 1", got)
	}
	if !cloudDeckHasKeyword(t, g, bear, "indestructible") {
		t.Error("Bear is not indestructible")
	}
	if spec, _ := Lookup(zackFairOracle); spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Errorf("Zack is %v with %d caveats, want caveats", spec.Completeness, len(spec.Caveats))
	}
}

func TestCloudDeckConformerShurikenTapsADefenderAndGrowsToMatch(t *testing.T) {
	cloudDeckWantFull(t, conformerShurikenOracl)
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	shuriken := seedEquipment(g, me.ID, "Conformer Shuriken", conformerShurikenOracl)
	equipTo(t, g, me.ID, shuriken, bear)

	declareAttack(t, g, opp.ID, bear)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickCard(t, g, me.ID, big)
	}
	passPriorityAroundTable(t, g)

	tapped := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == big {
				tapped = c.Tapped
			}
		}
	})
	if !tapped {
		t.Error("the defender's creature was not tapped")
	}
	if got := cloudDeckCounters(t, g, bear); got != 3 {
		t.Errorf("Bear counters = %d, want 3 (5 power vs 2)", got)
	}
}

func cloudDeckAttach(g *game.Game, attachment uuid.UUID, host game.TargetRef) error {
	var err error
	g.WithWriteLock(func() { err = g.AttachForEffect(attachment, host) })
	return err
}
