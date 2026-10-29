package effects

import (
	"testing"

	"github.com/google/uuid"
)

// msc_1_cards_test.go — Marvel Super Heroes Commander, first batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	msc1AimBotOracle         = "bd269d87-5f72-4559-bf2b-270e2ebe6154"
	msc1CaptainsDefenseOracl = "c3f83115-398f-4738-9bf2-b37bd363e9f6"
	msc1CityPigeonOracle     = "ac8cae63-270c-4f73-b29b-50f8f2395fd9"
	msc1CoastalPeakOracle    = "0e25faa2-efb0-4ca5-b280-18c38faa860c"
	msc1CommonCrookOracle    = "ebcbfdba-34c7-40eb-91a9-d306ff5ae8b9"
	msc1ConcFireOracle       = "2fcb2342-202d-4baa-82ad-66d782142264"
	msc1DenialOracle         = "99eff6e9-5911-4f89-bad1-0888067a7783"
	msc1ServoGuardsOracle    = "296e79f9-57cb-4af7-9e18-48d2d655b537"
	msc1ArmorOracle          = "f7ed3cd6-d138-4c12-93bb-91629492a5fa"
	msc1FuriousOracle        = "efefb339-bb7c-4b26-abed-8434f955122b"
)

func TestMsc1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc1AimBotOracle:         "A.I.M. Bot",
		msc1CaptainsDefenseOracl: "Captain's Defense",
		msc1CityPigeonOracle:     "City Pigeon",
		msc1CoastalPeakOracle:    "Coastal Peak",
		msc1CommonCrookOracle:    "Common Crook",
		msc1ConcFireOracle:       "Concentrated Fire",
		msc1DenialOracle:         "Dismissive Denial",
		msc1ServoGuardsOracle:    "Doom's Servo-Guards",
		msc1ArmorOracle:          "Experimental Armor",
		msc1FuriousOracle:        "Furious Strength",
	}
	for oracle, name := range want {
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

func TestMsc1AimBotFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bot := b12Push(g, me.ID, "A.I.M. Bot", "Artifact Creature — Robot Villain", msc1AimBotOracle, 2, 3)
	assertKeywords(t, g, bot, "flying")
}

func TestMsc1CaptainsDefensePumpsABlockerAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	crtBlock(g, blocker, attacker, me.ID)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Captain's Defense", "Instant", msc1CaptainsDefenseOracl, b16TargetCard(blocker))
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, blocker), effectiveToughness(t, g, blocker); p != 4 || tg != 4 {
		t.Errorf("blocker is %d/%d, want 4/4", p, tg)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want the spell's own card gone and one card drawn (+1 from the push, 0 net)", hand, got)
	}
}

func TestMsc1CaptainsDefenseNeedsABlockingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	err := castCatalogSpellErr(t, g, "Captain's Defense", "Instant", msc1CaptainsDefenseOracl, b16TargetCard(idle))
	if err == nil {
		t.Error("a creature that is not blocking was accepted as the target")
	}
}

func TestMsc1CityPigeonFliesAndLeavesAFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pigeon := b12Push(g, me.ID, "City Pigeon", "Creature — Bird", msc1CityPigeonOracle, 1, 1)
	assertKeywords(t, g, pigeon, "flying")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pigeon) })
	passPriorityAroundTable(t, g)
	if n := mshB1Tokens(g, me.ID, "Food"); n != 1 {
		t.Errorf("%d Food tokens after the Pigeon left, want 1", n)
	}
}

func TestMsc1CoastalPeakEntersTappedTapsForEitherColourAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playLandFromHand(t, g, "Coastal Peak", msc1CoastalPeakOracle)
	top100AssertEnteredTapped(t, g, id, "Coastal Peak")
	for _, colour := range []string{"U", "R"} {
		me.ManaPool = nil
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
		b28TapForMana(t, g, me.ID, id, colour)
		if got := poolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("tapped for {%s}: pool %v", colour, got)
		}
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	cycled, _ := cycleFromHand(t, g2, "Coastal Peak", "Land — Island Mountain", msc1CoastalPeakOracle, "{C}{C}")
	if !me2.Graveyard.Contains(cycled) {
		t.Error("the cycled land is not in the graveyard")
	}
}

func TestMsc1CommonCrookLeavesATreasureWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	crook := b12Push(g, me.ID, "Common Crook", "Creature — Human Rogue Villain", msc1CommonCrookOracle, 2, 2)
	if treasureCount(g) != 0 {
		t.Fatal("a Treasure exists before the Crook dies")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(crook) })
	passPriorityAroundTable(t, g)
	if n := treasureCount(g); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
}

func TestMsc1ConcentratedFireDealsFive(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b12Creature(g, opp.ID, "Five Toughness", "Creature — Wall", 0, 5)
	big := b12Creature(g, opp.ID, "Six Toughness", "Creature — Wall", 0, 6)
	castCatalogSpell(t, g, "Concentrated Fire", "Instant", msc1ConcFireOracle, b16TargetCard(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a creature with toughness 5 survived 5 damage")
	}
	castCatalogSpell(t, g, "Concentrated Fire", "Instant", msc1ConcFireOracle, b16TargetCard(big))
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, big)
	if !ok || card.DamageMarked != 5 {
		t.Errorf("toughness 6 creature: on battlefield %v, damage %d, want 5", ok, card.DamageMarked)
	}
}

func TestMsc1DismissiveDenialCountersASpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetPlayer(me.ID))
	castCatalogSpell(t, g, "Dismissive Denial", "Instant", msc1DenialOracle, b16TargetCard(bolt))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != 40 {
		t.Error("the Bolt is countered")
	}
}

func TestMsc1DismissiveDenialBasicLandcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Command Tower", "Land"),
	)
	id, _ := cycleFromHand(t, g, "Dismissive Denial", "Instant", msc1DenialOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the landcycling search did not ask")
	}
	if searchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("a nonbasic land was offered")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the basic land did not reach the hand")
	}
}

func TestMsc1ServoGuardsGainTwoAndMillTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three")
	life, yard, lib := me.Life, me.Graveyard.Size(), me.Library.Size()
	castHoldCreature(t, g, "Doom's Servo-Guards", "Artifact Creature — Robot Villain", msc1ServoGuardsOracle, 2, 2)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	if got := me.Graveyard.Size() - yard; got != 2 {
		t.Errorf("%d cards milled, want 2", got)
	}
	if got := lib - me.Library.Size(); got != 2 {
		t.Errorf("library fell by %d, want 2", got)
	}
}

func TestMsc1ExperimentalArmorGivesPlusOneFlyingAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	armor := b12Push(g, me.ID, "Experimental Armor", "Artifact — Equipment", msc1ArmorOracle, 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "C", "C")
	equipTo(t, g, me.ID, armor, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("equipped bear is %d/%d, want 3/3", p, tg)
	}
	assertKeywords(t, g, bear, "flying", "haste")
}

func TestMsc1FuriousStrengthGivesPlusTwoAndMenace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Furious Strength", msc1FuriousOracle, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("enchanted bear is %d/%d, want 4/4", p, tg)
	}
	assertKeywords(t, g, bear, "menace")
}
