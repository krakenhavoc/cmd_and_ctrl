package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const rimefireTorqueOracle = "c33e99c6-189e-4dcf-8b6a-64937dbad361"

func TestRimefireTorqueIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(rimefireTorqueOracle)
	if !ok || spec.Name != "Rimefire Torque" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

func rimefireTorqueCharge(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("Rimefire Torque is not on the battlefield")
	}
	return c.Counters[game.CounterCharge]
}

// Casting the Torque asks for a creature type; a permanent of that type
// entering under your control adds a charge counter, another type does
// not.
func TestRimefireTorqueChargesForTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	torque := effectstest.RfdHandCard(me, "Rimefire Torque", "Artifact", "", rimefireTorqueOracle, 0, 0)
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, torque, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.RfdCast(t, g, "Elvish Mystic", "Creature — Elf Druid", "", 1, 1)
	if got := rimefireTorqueCharge(t, g, torque); got != 1 {
		t.Fatalf("an Elf entering: %d charge counters, want 1", got)
	}
	effectstest.RfdCast(t, g, "Goblin Guide", "Creature — Goblin Scout", "", 2, 2)
	if got := rimefireTorqueCharge(t, g, torque); got != 1 {
		t.Errorf("a Goblin entering: %d charge counters, want 1", got)
	}
}

// Three charge counters and a tap: the next instant is copied. With only
// two the activation is refused.
func TestRimefireTorqueCopiesTheNextInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	torque := effectstest.PushNamedTribePermanent(t, g, me, "Rimefire Torque", "Artifact", rimefireTorqueOracle, "Elf")
	effectstest.SetCounters(g, torque, map[string]int{game.CounterCharge: 2})
	effectstest.ToMainForCost(t, g)
	if err := g.ActivateCatalogAbility(me, torque, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only two charge counters")
	}
	if effectstest.B16Tapped(t, g, torque) {
		t.Fatal("a refused activation tapped the Torque")
	}

	effectstest.SetCounters(g, torque, map[string]int{game.CounterCharge: 3})
	effectstest.B16Activate(t, g, me, torque, 0, game.ActivateAbilityParams{Strict: true})
	if got := rimefireTorqueCharge(t, g, torque); got != 0 {
		t.Errorf("charge counters after activating = %d, want 0", got)
	}
	if !effectstest.B16Tapped(t, g, torque) {
		t.Error("the Torque is tapped as the cost")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want 1", len(g.DelayedTriggers))
	}

	before := effectstest.LifeOf(g, victim)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)
	if got := effectstest.LifeOf(g, victim); got != before-6 {
		t.Errorf("victim life = %d, want %d (the Bolt and its copy)", got, before-6)
	}
}
