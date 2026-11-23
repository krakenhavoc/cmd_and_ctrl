package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b36_test.go — S82 bonus sheets, batch 36 (Foundations Jumpstart):
// eleven cards through the real catalog.

const (
	bnsOrchardStrider    = "c6906032-e08b-478f-bc71-cb1184f60bad"
	bnsPaladinBloodstain = "f340018d-8b68-45d5-a72e-a922be68365d"
	bnsPathmakerInitiate = "c0247a19-9c8f-422c-88cf-9f298c3a8c96"
	bnsPestermite        = "83cf4235-eb4e-4c55-839a-f2740d41ed5e"
	bnsPhantomNinja      = "e7a13f43-1171-4808-9964-c5ad424027cb"
	bnsPharikasChosen    = "b5a26a1b-d2be-406e-a6a9-1993045313b5"
	bnsPrimevalHerald    = "495348d7-bc11-4209-8b0f-72db2e6c6993"
	bnsPurpleCrab        = "df4569d7-ee75-4ba0-bbc3-7766ac686523"
	bnsRealmSeekers      = "ccb66cdc-909a-42c1-a10f-3886e9ed1936"
	bnsAmplimancer       = "1c264f6e-caa7-44cb-8719-86212292679c"
)

func bnsBasicForest(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: p.ID, Controller: p.ID})
	return id
}

func TestBnsB36Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsOrchardStrider: "Orchard Strider", bnsPaladinBloodstain: "Paladin of the Bloodstained",
		bnsPathmakerInitiate: "Pathmaker Initiate",
		bnsPestermite:        "Pestermite", bnsPhantomNinja: "Phantom Ninja",
		bnsPharikasChosen: "Pharika's Chosen", bnsPrimevalHerald: "Primeval Herald",
		bnsPurpleCrab: "Purple-Crystal Crab", bnsRealmSeekers: "Realm Seekers",
		bnsAmplimancer: "Reckless Amplimancer",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestBnsB36KeywordsAndStatics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	for _, c := range []struct {
		oracle, name string
		kws          []string
	}{
		{bnsPestermite, "Pestermite", []string{"flash", "flying"}},
		{bnsPharikasChosen, "Pharika's Chosen", []string{"deathtouch"}},
		{bnsPrimevalHerald, "Primeval Herald", []string{"trample"}},
	} {
		id := effectstest.BlbB1Push(g, me.ID, c.name, "Creature", c.oracle, 1, 1)
		for _, kw := range c.kws {
			if !effectstest.P1g2HasKeyword(g, id, kw) {
				t.Errorf("%s lacks %s", c.name, kw)
			}
		}
	}
	ninja := effectstest.BlbB1Push(g, me.ID, "Phantom Ninja", "Creature — Illusion Ninja", bnsPhantomNinja, 2, 2)
	if effectstest.RestrictionsOf(t, g, ninja)&game.CantBeBlocked == 0 {
		t.Error("Phantom Ninja can be blocked")
	}
	_ = opp
}

func TestBnsB36OrchardStriderMakesTwoFoods(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Orchard Strider", "Creature — Treefolk", bnsOrchardStrider, "{4}{G}{G}", 6, 4,
		"G", "G", "G", "G", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Food"); n != 2 {
		t.Errorf("want 2 Food tokens, got %d", n)
	}
	_ = me
}

func TestBnsB36OrchardStriderBasicLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	forest := bnsBasicForest(me)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Orchard Strider", TypeLine: "Creature — Treefolk", OracleID: bnsOrchardStrider,
		ManaCost: "{4}{G}{G}", Power: 6, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("landcycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	found := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == forest {
			found = true
		}
	}
	if !found {
		t.Error("the basic land did not reach the hand")
	}
}

func TestBnsB36PaladinMakesALifelinkVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Paladin of the Bloodstained", "Creature — Vampire Knight", bnsPaladinBloodstain, "{3}{W}", 3, 2,
		"W", "W", "W", "W"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	var tok game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Vampire" {
			tok = c
		}
	}
	if tok.InstanceID == uuid.Nil {
		t.Fatal("no Vampire token")
	}
	if !effectstest.P1g2HasKeyword(g, tok.InstanceID, "lifelink") || tok.Power != 1 || tok.Toughness != 1 {
		t.Errorf("token is %d/%d, want a 1/1 with lifelink", tok.Power, tok.Toughness)
	}
}

func TestBnsB36PathmakerMakesASmallCreatureUnblockableAndRefusesABigOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	init := effectstest.BlbB1Push(g, me.ID, "Pathmaker Initiate", "Creature — Human Wizard", bnsPathmakerInitiate, 2, 1)
	small := effectstest.BlbB1Push(g, me.ID, "Small", "Creature — Bear", "", 2, 2)
	big := effectstest.BlbB1Push(g, me.ID, "Big", "Creature — Bear", "", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.Unsick(g, init)
	err := g.ActivateCatalogAbility(me.ID, init, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(big)})
	if err == nil {
		t.Fatal("a power-3 creature was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, init, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, small)&game.CantBeBlocked == 0 {
		t.Error("the target can still be blocked")
	}
}

func TestBnsB36PestermiteTapsOrUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Pestermite", "Creature — Faerie Rogue", bnsPestermite, "{2}{U}", 2, 1,
		"U", "U", "U"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("\"tap it\" left the target untapped")
	}
}

func TestBnsB36PrimevalHeraldSearchesWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	forest := bnsBasicForest(me)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Primeval Herald", "Creature — Elf Scout", bnsPrimevalHerald, "{3}{G}", 3, 1,
		"G", "G", "G", "G"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	c, ok := effectstest.BattlefieldCard(g, forest)
	if !ok {
		t.Fatal("the land did not enter the battlefield")
	}
	if !c.Tapped {
		t.Error("the land entered untapped")
	}
}

func TestBnsB36CrabDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	crab := effectstest.BlbB1Push(g, me.ID, "Purple-Crystal Crab", "Creature — Crab", bnsPurpleCrab, 1, 1)
	before := effectstest.HandSizeOf(g, me)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(crab) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSizeOf(g, me); got != before+1 {
		t.Errorf("hand %d -> %d, want one card drawn", before, got)
	}
}

func TestBnsB36RealmSeekersCountsAllHandsAndSearchesForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	forest := bnsBasicForest(me)
	var total int
	id, err := effectstest.BlbHCastWithMana(t, g, "Realm Seekers", "Creature — Elf Scout", bnsRealmSeekers, "{4}{G}{G}", 0, 0,
		"G", "G", "G", "G", "G", "G")
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			total += p.Hand.Size()
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCountersOf(g, id, game.CounterPlusOne); got != total {
		t.Fatalf("entered with %d counters, want %d (the cards in all hands)", got, total)
	}
	if total == 0 {
		t.Skip("no cards in hand to count")
	}
	effectstest.B06AddMana(me, "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if got := effectstest.BattlefieldCountersOf(g, id, game.CounterPlusOne); got != total-1 {
		t.Errorf("%d counters left, want %d", got, total-1)
	}
	_ = opp
}

func TestBnsB36AmplimancerDoublesPowerAndToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	id := effectstest.BlbB1Push(g, me.ID, "Reckless Amplimancer", "Creature — Elf Druid", bnsAmplimancer, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G", "G", "G", "G", "G", "G", "G", "G", "G", "G")
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	var p, tough int
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(id)
		p, tough = c.PowerForComparison(), c.CurrentToughness()
	})
	if p != 8 || tough != 8 {
		t.Errorf("got %d/%d, want 8/8 after two activations", p, tough)
	}
}
