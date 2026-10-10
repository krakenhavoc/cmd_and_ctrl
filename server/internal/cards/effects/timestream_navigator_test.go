package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// timestream_navigator_test.go — #2726: Timestream Navigator's cost puts
// itself on the bottom of its owner's library, and the ability needs the
// city's blessing. The engine half is game/bottom_self_cost_test.go.

const timestreamNavigatorOracle = "d0896996-ee31-402e-b642-4e7cca4929fe"

func TestTimestreamNavigatorWithoutTheBlessingCostsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	nav := pushCatalogPermanent(g, me.ID, "Timestream Navigator", "Creature — Human Pirate Wizard", timestreamNavigatorOracle, false)
	floatMana(t, g, me, "{U}{U}{U}{U}")
	turns := len(g.ExtraTurns)

	err := g.ActivateCatalogAbility(me.ID, nav, 0, game.ActivateAbilityParams{})
	if err == nil {
		t.Fatal("activated without the city's blessing")
	}
	if !g.Battlefield.Contains(nav) || len(g.ExtraTurns) != turns {
		t.Fatalf("a refused activation (%v) moved the creature or took a turn", err)
	}
}

func TestTimestreamNavigatorBottomsItselfAndTakesAnExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	nav := pushCatalogPermanent(g, me.ID, "Timestream Navigator", "Creature — Human Pirate Wizard", timestreamNavigatorOracle, false)
	grantBlessing(g, me)
	floatMana(t, g, me, "{U}{U}{U}{U}")
	turns := len(g.ExtraTurns)

	if err := g.ActivateCatalogAbility(me.ID, nav, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(nav) {
		t.Fatal("the Navigator is still on the battlefield once the ability is activated")
	}
	if len(me.Library.Cards) == 0 || me.Library.Cards[0].InstanceID != nav {
		t.Fatal("the Navigator is not on the bottom of its owner's library")
	}
	if len(g.ExtraTurns) != turns {
		t.Fatal("the extra turn came before the ability resolved")
	}
	passPriorityAroundTable(t, g)
	if len(g.ExtraTurns) != turns+1 {
		t.Fatalf("extra turns = %d, want %d", len(g.ExtraTurns), turns+1)
	}
	// The blessing is kept, so a second Navigator would work again, and
	// this one is gone for good.
	if !YouHaveTheCitysBlessing(g, me.ID) {
		t.Error("the city's blessing was lost")
	}
}

func TestTimestreamNavigatorIsFullAndAscends(t *testing.T) {
	spec, ok := Lookup(timestreamNavigatorOracle)
	if !ok {
		t.Fatal("Timestream Navigator is not in the catalog")
	}
	if spec.Completeness != CompletenessFull || !hasKeyword(spec.PrintedKeywords, game.KeywordAscend) {
		t.Errorf("completeness %v, keywords %v; want full with ascend", spec.Completeness, spec.PrintedKeywords)
	}
	if !spec.Activated[0].Cost.BottomSelf || !spec.Activated[0].Cost.Tap || spec.Activated[0].Cost.Mana != "{2}{U}{U}" {
		t.Errorf("cost = %+v, want {2}{U}{U}, {T}, bottom-self", spec.Activated[0].Cost)
	}
}
