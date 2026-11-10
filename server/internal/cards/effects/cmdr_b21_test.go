package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b21_test.go — Commander deck cards, batch 21 (#5368).

const (
	cmdrB21MigrationPathOracle = "db4907f3-4071-47c4-9329-96b4813435be"
	cmdrB21BattlementOracle    = "585f62dc-4461-42f1-a3a4-b19a1e550d2d"
	cmdrB21CaryatidOracle      = "13d4c46b-c2d6-44cc-a252-4a991d471854"
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
