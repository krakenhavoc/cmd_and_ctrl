package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// technodrome_test.go — #4678's card. The engine and enumerator halves
// of the power floor are in game/power_floor_test.go and
// legal/power_floor_test.go.

const technodromeOracle = "51c0b46c-c379-420a-a4a4-d18724c759aa"

func TestTechnodromeIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(technodromeOracle)
	if !ok || spec.Name != "Technodrome" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Technodrome: registered %v, %+v", ok, spec)
	}
}

func TestTechnodromeDrawsGrowsAndAttacksAtSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	drome := effectstest.PushDiesCreatureForTest(g, me.ID, "Technodrome", technodromeOracle, "Artifact Creature — Construct", 3, 3)
	rock := effectstest.PushPermanentForTest(g, me.ID, "Rock", "", "Artifact")
	effectstest.AssertKeywords(t, g, drome, "reach", "trample")
	if a, b := effectstest.TechnodromeRestricted(t, g, drome); !a || !b {
		t.Fatalf("a 3/3 Technodrome can attack %v / block %v, want neither", !a, !b)
	}

	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, drome, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, rock) {
		t.Error("the sacrificed artifact is still there")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d (drew a card)", got, hand+1)
	}
	if got := effectstest.B12Counter(t, g, drome, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if a, b := effectstest.TechnodromeRestricted(t, g, drome); !a || !b {
		t.Fatal("a 4/4 Technodrome may attack or block")
	}

	if err := g.AddCounter(drome, game.CounterPlusOne, 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if a, b := effectstest.TechnodromeRestricted(t, g, drome); a || b {
		t.Fatalf("a 6/6 Technodrome still can't attack %v / block %v", a, b)
	}
	if err := g.AddCounter(drome, game.CounterMinusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if a, b := effectstest.TechnodromeRestricted(t, g, drome); !a || !b {
		t.Fatal("a 5/5 Technodrome may attack or block")
	}
}

func TestTechnodromeCannotSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	drome := effectstest.PushDiesCreatureForTest(g, me.ID, "Technodrome", technodromeOracle, "Artifact Creature — Construct", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, drome, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{drome}}); err == nil {
		t.Fatal("sacrificing itself must be refused (\"another artifact\")")
	}
}
