package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b21_test.go — Commander deck cards, batch 21 (#5368).

const (
	cmdrB21MigrationPathOracle = "db4907f3-4071-47c4-9329-96b4813435be"
	cmdrB21BattlementOracle    = "585f62dc-4461-42f1-a3a4-b19a1e550d2d"
	cmdrB21CaryatidOracle      = "13d4c46b-c2d6-44cc-a252-4a991d471854"
	cmdrB21WitnessOracle       = "a52a06aa-56fc-4845-bbcd-a54f16369b34"
	cmdrB21TowerDefenseOracle  = "58853200-efbe-41cc-87f1-720841308f4d"
	cmdrB21TitanOracle         = "a95309ac-f920-474d-81b5-0f68c8a54f34"
)

func cmdrB21AssertFull(t *testing.T, oracle, name string) {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", name)
	}
	if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
	}
}

func TestCmdrB21MigrationPathFetchesTwoBasicsTappedAndCycles(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21MigrationPathOracle, "Migration Path")
	g := newCatalogGame(t)
	me := g.Seats[0]
	stapleLibraryCard(me, "Basic Forest", "Basic Land — Forest")
	stapleLibraryCard(me, "Basic Island", "Basic Land — Island")
	stapleLibraryCard(me, "Basic Swamp", "Basic Land — Swamp")
	castCatalogSpell(t, g, "Migration Path", "Sorcery", cmdrB21MigrationPathOracle, nil)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c == nil {
		t.Fatal("no search prompt")
	}
	answerSearchNamed(t, g, me.ID, "Basic Forest", "Basic Island")
	passPriorityAroundTable(t, g)
	n := 0
	for _, bc := range g.Battlefield.Cards {
		if bc.Controller == me.ID && bc.IsLand() {
			n++
			if !bc.Tapped {
				t.Errorf("%s entered untapped", bc.Name)
			}
		}
	}
	if n != 2 {
		t.Errorf("%d lands fetched, want 2", n)
	}

	// Cycling {2}.
	advanceTo(t, g, game.StepPrecombatMain)
	id := pushCatalogHandCard(me, "Migration Path", "Sorcery", cmdrB21MigrationPathOracle)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("cycled with an empty pool")
	}
	apaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card should be in the graveyard")
	}
}

func TestCmdrB21OvergrownBattlementCountsCreaturesWithDefender(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21BattlementOracle, "Overgrown Battlement")
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushCatalogPermanent(g, me.ID, "Overgrown Battlement", "Creature — Wall", cmdrB21BattlementOracle, false)
	tapForMana(t, g, me.ID, a)
	if got := sortedPool(me); len(got) != 1 || got[0] != "G" {
		t.Fatalf("alone: pool = %v, want {G}", got)
	}
	me.ManaPool.EmptyPool()
	b := pushCatalogPermanent(g, me.ID, "Overgrown Battlement", "Creature — Wall", cmdrB21BattlementOracle, false)
	_ = b
	pushCatalogPermanent(g, me.ID, "Sylvan Caryatid", "Creature — Plant", cmdrB21CaryatidOracle, false)
	// An opponent's defender and a creature without defender don't count.
	pushCatalogPermanent(g, g.Seats[1].ID, "Overgrown Battlement", "Creature — Wall", cmdrB21BattlementOracle, false)
	pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	// Tapping the first one: three defenders (two Battlements, one Caryatid).
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == a {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	tapForMana(t, g, me.ID, a)
	if got := sortedPool(me); len(got) != 3 {
		t.Errorf("pool = %v, want three {G}", got)
	}
}

func TestCmdrB21SylvanCaryatidTapsForAnyColourAndHasHexproof(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21CaryatidOracle, "Sylvan Caryatid")
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Sylvan Caryatid", "Creature — Plant", cmdrB21CaryatidOracle, false)
	if !hasEffectiveKeyword(t, g, id, "hexproof") || !hasEffectiveKeyword(t, g, id, "defender") {
		t.Error("Sylvan Caryatid should have defender and hexproof")
	}
	tapForMana(t, g, me.ID, id)
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("pick = %+v, want one mana of any of five colours", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool = %v, want {U}", got)
	}
}

func TestCmdrB21TimelessWitnessReturnsACardMandatorilyAndEternalizes(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21WitnessOracle, "Timeless Witness")
	g := newCatalogGame(t)
	me := g.Seats[0]
	body := pushGraveyardCardForTest(me, "Old Body")
	castAndResolveCreature(t, g, "Timeless Witness", "Creature — Human Shaman", cmdrB21WitnessOracle)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("the trigger asked \"you may\"; Timeless Witness's return is mandatory")
	}
	pickCard(t, g, me.ID, body)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(body) {
		t.Error("the targeted card did not return to hand")
	}

	// Eternalize {5}{G}{G}: a 4/4 black Zombie copy, whose own trigger runs.
	other := pushGraveyardCardForTest(me, "Second Body")
	wit := seedGraveyardCard(t, g, "Timeless Witness", "Creature — Human Shaman", cmdrB21WitnessOracle)
	if err := g.ActivateCatalogAbility(me.ID, wit, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("eternalized with an empty pool")
	}
	apaMana(me, "C", "C", "C", "C", "C", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, wit, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("eternalize: %v", err)
	}
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(wit) {
		t.Error("the eternalized card should be exiled")
	}
	if !me.Hand.Contains(other) {
		t.Error("the token's enters trigger did not return the second card")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Timeless Witness" && c.Controller == me.ID && c.InstanceID != wit {
			if !c.IsToken() {
				continue
			}
			found = true
			if c.Power != 4 || c.Toughness != 4 {
				t.Errorf("token %d/%d, want a 4/4 token", c.Power, c.Toughness)
			}
		}
	}
	if !found {
		t.Error("no eternalized token")
	}
}

func TestCmdrB21TowerDefenseBoostsAndGrantsReachToYourCreaturesOnly(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21TowerDefenseOracle, "Tower Defense")
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Tower Defense", "Instant", cmdrB21TowerDefenseOracle, nil)
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, mine) != 2 || effectiveToughness(t, g, mine) != 7 {
		t.Errorf("my bear %d/%d, want 2/7", effectivePower(t, g, mine), effectiveToughness(t, g, mine))
	}
	if !hasEffectiveKeyword(t, g, mine, "reach") {
		t.Error("my bear should have reach")
	}
	if effectiveToughness(t, g, theirs) != 2 || hasEffectiveKeyword(t, g, theirs, "reach") {
		t.Error("their bear was affected")
	}
}

func TestCmdrB21ToweringTitanCountsToughnessAndSacrificesForTrample(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21TitanOracle, "Towering Titan")
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := pushCatalogPermanent(g, me.ID, "Sylvan Caryatid", "Creature — Plant", cmdrB21CaryatidOracle, false) // 1/1 seeded, defender
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 9, 9)
	toughness := effectiveToughness(t, g, wall) + 3
	titan := castAndResolveCreature(t, g, "Towering Titan", "Creature — Giant", cmdrB21TitanOracle)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, titan); got != toughness {
		t.Errorf("Titan entered with %d counters, want %d (other creatures you control)", got, toughness)
	}

	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, titan, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("sacrificed a creature without defender")
	}
	if err := g.ActivateCatalogAbility(me.ID, titan, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{wall}}); err != nil {
		t.Fatalf("sacrifice a defender: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, still := battlefieldCard(g, wall); still {
		t.Error("the defender was not sacrificed")
	}
	if !hasEffectiveKeyword(t, g, titan, "trample") || !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("all creatures should have trample")
	}
	their := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opp.ID && hasEffectiveKeyword(t, g, c.InstanceID, "trample") {
			their++
		}
	}
	if their != 1 {
		t.Errorf("%d of the opponent's creatures have trample, want 1 (all creatures)", their)
	}
}
