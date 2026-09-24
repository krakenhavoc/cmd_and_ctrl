package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const talonGatesOfMadaraOracle = "8c45bf9d-a017-43bf-9e32-67810a8a217b"

func TestTalonGatesOfMadaraPhasesOutATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	playLandFromHand(t, g, "Talon Gates of Madara", talonGatesOfMadaraOracle)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if !phasedOutInCatalogGame(g, bear) {
		t.Fatal("a creature is a legal target and phases out")
	}
}

// "Up to one" means declining is a legal answer, and nothing phases
// out — the same shape Thassa's end-step blink pins.
func TestTalonGatesOfMadaraETBCanDeclineTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	playLandFromHand(t, g, "Talon Gates of Madara", talonGatesOfMadaraOracle)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt for the up-to-one target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the up-to-one pick: %v", err)
	}
	passPriorityAroundTable(t, g)

	if phasedOutInCatalogGame(g, bear) {
		t.Error("declining the pick still phased out the creature")
	}
}

func TestTalonGatesOfMadaraManaAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Talon Gates of Madara", "Land — Gate", talonGatesOfMadaraOracle)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool = %v, want {C}", got)
	}
	me.ManaPool.EmptyPool()
	b08Untap(g, land)

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("float {1}: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{1}, {T}: Add one mana of any color: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("colour options %+v, want all five", pick)
	}
	riderAnswerManaPicks(t, g, me.ID, "R")
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
}

func TestTalonGatesOfMadaraPutsItselfFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := pushCatalogHandCard(me, "Talon Gates of Madara", "Land — Gate", talonGatesOfMadaraOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("{4}: put this card from your hand onto the battlefield: %v", err)
	}
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("declining the ETB phase-out: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the land is not on the battlefield")
	}
	if me.Hand.Contains(id) {
		t.Error("the land is still in hand")
	}
}
