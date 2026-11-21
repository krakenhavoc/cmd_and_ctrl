package testsms

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestMsc1AimBotFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bot := effectstest.B12Push(g, me.ID, "A.I.M. Bot", "Artifact Creature — Robot Villain", msc1AimBotOracle, 2, 3)
	effectstest.AssertKeywords(t, g, bot, "flying")
}

func TestMsc1CaptainsDefensePumpsABlockerAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CrtBlock(g, blocker, attacker, me.ID)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Captain's Defense", "Instant", msc1CaptainsDefenseOracl, effectstest.B16TargetCard(blocker))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, blocker), effectstest.EffectiveToughness(t, g, blocker); p != 4 || tg != 4 {
		t.Errorf("blocker is %d/%d, want 4/4", p, tg)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want the spell's own card gone and one card drawn (+1 from the push, 0 net)", hand, got)
	}
}

func TestMsc1CaptainsDefenseNeedsABlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Captain's Defense", "Instant", msc1CaptainsDefenseOracl, effectstest.B16TargetCard(idle))
	if err == nil {
		t.Error("a creature that is not blocking was accepted as the target")
	}
}

func TestMsc1CityPigeonFliesAndLeavesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pigeon := effectstest.B12Push(g, me.ID, "City Pigeon", "Creature — Bird", msc1CityPigeonOracle, 1, 1)
	effectstest.AssertKeywords(t, g, pigeon, "flying")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pigeon) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.MshB1Tokens(g, me.ID, "Food"); n != 1 {
		t.Errorf("%d Food tokens after the Pigeon left, want 1", n)
	}
}

func TestMsc1CoastalPeakEntersTappedTapsForEitherColourAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PlayLandFromHand(t, g, "Coastal Peak", msc1CoastalPeakOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Coastal Peak")
	for _, colour := range []string{"U", "R"} {
		me.ManaPool = nil
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
		effectstest.B28TapForMana(t, g, me.ID, id, colour)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("tapped for {%s}: pool %v", colour, got)
		}
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	cycled, _ := effectstest.CycleFromHand(t, g2, "Coastal Peak", "Land — Island Mountain", msc1CoastalPeakOracle, "{C}{C}")
	if !me2.Graveyard.Contains(cycled) {
		t.Error("the cycled land is not in the graveyard")
	}
}

func TestMsc1CommonCrookLeavesATreasureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crook := effectstest.B12Push(g, me.ID, "Common Crook", "Creature — Human Rogue Villain", msc1CommonCrookOracle, 2, 2)
	if effectstest.TreasureCount(g) != 0 {
		t.Fatal("a Treasure exists before the Crook dies")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(crook) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TreasureCount(g); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
}

func TestMsc1ConcentratedFireDealsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Five Toughness", "Creature — Wall", 0, 5)
	big := effectstest.B12Creature(g, opp.ID, "Six Toughness", "Creature — Wall", 0, 6)
	effectstest.CastCatalogSpell(t, g, "Concentrated Fire", "Instant", msc1ConcFireOracle, effectstest.B16TargetCard(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a creature with toughness 5 survived 5 damage")
	}
	effectstest.CastCatalogSpell(t, g, "Concentrated Fire", "Instant", msc1ConcFireOracle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, big)
	if !ok || card.DamageMarked != 5 {
		t.Errorf("toughness 6 creature: on battlefield %v, damage %d, want 5", ok, card.DamageMarked)
	}
}

func TestMsc1DismissiveDenialCountersASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetPlayer(me.ID))
	effectstest.CastCatalogSpell(t, g, "Dismissive Denial", "Instant", msc1DenialOracle, effectstest.B16TargetCard(bolt))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != 40 {
		t.Error("the Bolt is countered")
	}
}

func TestMsc1DismissiveDenialBasicLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Command Tower", "Land"),
	)
	id, _ := effectstest.CycleFromHand(t, g, "Dismissive Denial", "Instant", msc1DenialOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the landcycling search did not ask")
	}
	if effectstest.SearchOptionNamed(g, c, "Command Tower") != uuid.Nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	life, yard, lib := me.Life, me.Graveyard.Size(), me.Library.Size()
	effectstest.CastHoldCreature(t, g, "Doom's Servo-Guards", "Artifact Creature — Robot Villain", msc1ServoGuardsOracle, 2, 2)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	armor := effectstest.B12Push(g, me.ID, "Experimental Armor", "Artifact — Equipment", msc1ArmorOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.EquipTo(t, g, me.ID, armor, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("equipped bear is %d/%d, want 3/3", p, tg)
	}
	effectstest.AssertKeywords(t, g, bear, "flying", "haste")
}

func TestMsc1FuriousStrengthGivesPlusTwoAndMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Furious Strength", msc1FuriousOracle, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("enchanted bear is %d/%d, want 4/4", p, tg)
	}
	effectstest.AssertKeywords(t, g, bear, "menace")
}
