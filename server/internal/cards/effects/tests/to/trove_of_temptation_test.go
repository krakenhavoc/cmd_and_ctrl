package teststo

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trove_of_temptation_test.go — "Each opponent must attack you or a
// planeswalker you control with at least one creature each combat if
// able" (#6087), and the end-step Treasure.

const troveOfTemptationOracle = "f624c0f1-1c09-4b7a-b140-11329f27e3c7"

// TestTroveOfTemptationWantsAnAttackOnItsController: an attack elsewhere
// does not obey it — with one creature, that attack is refused outright;
// with two, the pass is refused until one attacks Trove's controller.
func TestTroveOfTemptationWantsAnAttackOnItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, troveSide, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, troveSide.ID, "Trove of Temptation", "Enchantment", troveOfTemptationOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	giant := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)

	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) {
		t.Fatalf("empty declaration under Trove = %v, want an attack-requirement refusal", err)
	}
	want := "At least one creature must attack " + troveSide.Name + " or a planeswalker they control this combat if able (Trove of Temptation)."
	if got := re.Sentence(me.ID); got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if err := g.DeclareAttacker(bear, other.ID); err != nil {
		t.Fatalf("the Bears at another opponent while the Giant can still obey: %v", err)
	}
	if err := g.PassPriority(); !errors.As(err, &re) {
		t.Fatalf("pass with no attack on Trove's controller = %v, want a refusal", err)
	}
	if err := g.DeclareAttacker(giant, troveSide.ID); err != nil {
		t.Fatalf("the Giant at Trove's controller: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with Trove obeyed: %v", err)
	}

	// One creature: an attack at anybody else is refused.
	g2 := effectstest.NewCatalogGame(t)
	me2, troveSide2, other2 := g2.Seats[0], g2.Seats[1], g2.Seats[2]
	effectstest.B12Push(g2, troveSide2.ID, "Trove of Temptation", "Enchantment", troveOfTemptationOracle, 0, 0)
	only := effectstest.B12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g2, 0)
	if err := g2.DeclareAttacker(only, other2.ID); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("the only creature at another opponent = %v, want ErrAttackRequirement", err)
	}
	if err := g2.DeclareAttacker(only, troveSide2.ID); err != nil {
		t.Fatalf("the only creature at Trove's controller: %v", err)
	}
}

// TestTroveOfTemptationMakesATreasureAtYourEndStepOnly: one Treasure at
// its controller's end step, none at an opponent's.
func TestTroveOfTemptationMakesATreasureAtYourEndStepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Trove of Temptation", "Enchantment", troveOfTemptationOracle, 0, 0)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.MonSettle(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Treasures after your end step = %d, want 1", n)
	}
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.MonSettle(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Treasures after an opponent's end step = %d, want still 1", n)
	}
}
