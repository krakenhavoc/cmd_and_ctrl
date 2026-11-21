package testsbr

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const captivatingCrewOracle = "1049bc06-83de-4ed6-ae38-f259e5038a95"

// TestCaptivatingCrewStealsUntapsAndHastesAnOpponentsCreature exercises
// every printed clause — Act of Treason's three primitives (#756) off
// an activated ability rather than a spell.
func TestCaptivatingCrewStealsUntapsAndHastesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crew := effectstest.PushCatalogPermanent(g, me.ID, "Captivating Crew", "Creature — Human Pirate", captivatingCrewOracle, false)
	victim := effectstest.PushTappedVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	if err := g.AddManaForEffect(me.ID, me.ID, "{C}{C}{C}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.ActivateCatalogAbility(me.ID, crew, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(victim),
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Errorf("controller %s, want the activator %s", got, me.ID)
	}
	if c, _ := effectstest.BattlefieldCard(g, victim); c.Tapped {
		t.Error("Captivating Crew did not untap the creature it stole")
	}
	effectstest.AssertKeywords(t, g, victim, "haste")
}

// A creature you control is not "a creature an opponent controls".
func TestCaptivatingCrewRefusesACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crew := effectstest.PushCatalogPermanent(g, me.ID, "Captivating Crew", "Creature — Human Pirate", captivatingCrewOracle, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	if err := g.AddManaForEffect(me.ID, me.ID, "{C}{C}{C}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.ActivateCatalogAbility(me.ID, crew, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(mine),
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("targeting my own creature: err = %v, want ErrIllegalTarget", err)
	}
}

// Activate only as a sorcery: refused with something on the stack.
func TestCaptivatingCrewIsSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crew := effectstest.PushCatalogPermanent(g, me.ID, "Captivating Crew", "Creature — Human Pirate", captivatingCrewOracle, false)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	if err := g.AddManaForEffect(me.ID, me.ID, "{C}{C}{C}{C}{R}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	effectstest.CastWithCost(t, g, "Distraction", "Instant", "{C}", "")

	if err := g.ActivateCatalogAbility(me.ID, crew, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(victim),
	}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("with something on the stack: err = %v, want ErrSorcerySpeedRequired", err)
	}
}
