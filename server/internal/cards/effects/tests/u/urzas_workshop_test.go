package testsu

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const urzasWorkshopOracle = "71099427-e110-488f-ab29-7867241fc7f0"

func TestUrzasWorkshopIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(urzasWorkshopOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || len(spec.ManaAbilities) != 2 {
		t.Fatalf("registered %v, spec %+v", ok, spec)
	}
}

// {T}: Add {C} always works.
func TestUrzasWorkshopTapsForOneColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ws := effectstest.SeedPermanentWithOracle(g, me.ID, "Urza's Workshop", "Land — Urza's", urzasWorkshopOracle)
	if err := g.ActivateManaAbility(me.ID, ws, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool = %v, want {C}", me.ManaPool)
	}
}

// Metalcraft needs three artifacts; it then adds {C} per Urza's land you
// control (Workshop counts itself; an opponent's Urza's land does not).
func TestUrzasWorkshopMetalcraftScalesWithUrzasLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ws := effectstest.SeedPermanentWithOracle(g, me.ID, "Urza's Workshop", "Land — Urza's", urzasWorkshopOracle)
	effectstest.SeedPermanentWithOracle(g, me.ID, "Urza's Factory", "Land — Urza's", "")
	effectstest.SeedPermanentWithOracle(g, opp.ID, "Their Urza's Land", "Land — Urza's", "")
	effectstest.SeedPermanentWithOracle(g, me.ID, "Sol Ring", "Artifact", "")
	effectstest.SeedPermanentWithOracle(g, me.ID, "Ornithopter", "Artifact Creature — Thopter", "")

	if err := g.ActivateManaAbility(me.ID, ws, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("err = %v, want ErrConditionNotMet with two artifacts", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Fatalf("a refused activation paid %v", me.ManaPool)
	}
	effectstest.SeedPermanentWithOracle(g, me.ID, "Mox", "Artifact", "")
	if err := g.ActivateManaAbility(me.ID, ws, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility with metalcraft: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool = %v, want {C}{C} (Workshop and Factory)", me.ManaPool)
	}
}
