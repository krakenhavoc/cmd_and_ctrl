package testst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Eternal Wanderer's +1 returns the exiled card at the beginning of
// ITS OWNER's next end step (#1538, DelayedTrigger.TurnOf): an
// opponent's creature exiled on your turn waits out your end step and
// comes back on theirs, under their control, while the delayed ability
// is still yours (CR 603.7d).
func TestEternalWandererPlusOneReturnsOnTheOwnersEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	wanderer := effectstest.PushWanderer(g, me.ID)
	victim := effectstest.PushFlickerCreature(g, opp.ID, "Mulldrifter", effectstest.MulldrifterOracle)
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, wanderer, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ExileHas(g, victim) {
		t.Fatal("the target was not exiled")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed queue = %d, want 1", len(g.DelayedTriggers))
	}
	if dt := g.DelayedTriggers[0]; dt.TurnOf != opp.ID || dt.Controller != me.ID {
		t.Fatalf("delayed trigger TurnOf=%v Controller=%v, want the owner %v and the Wanderer's controller %v",
			dt.TurnOf, dt.Controller, opp.ID, me.ID)
	}

	// The Wanderer controller's own end step is not the owner's.
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ExileHas(g, victim) || len(effectstest.BattlefieldIDsNamed(g, "Mulldrifter")) != 0 {
		t.Fatal("the card came back on the Wanderer controller's end step")
	}

	effectstest.AdvanceToEndStepOf(t, g, opp.Seat)
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.BattlefieldIDsNamed(g, "Mulldrifter")
	if len(back) != 1 {
		t.Fatalf("returned %d Mulldrifters on the owner's end step, want 1", len(back))
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == back[0] && c.Controller != opp.ID {
			t.Errorf("returned under %v's control, want its owner %v", c.Controller, opp.ID)
		}
	}
}

// Your own card is exiled on your own turn, so it comes back at this
// turn's end step exactly as before.
func TestEternalWandererPlusOneOnYourOwnCreatureReturnsThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	wanderer := effectstest.PushWanderer(g, me.ID)
	mine := effectstest.PushFlickerCreature(g, me.ID, "Mulldrifter", effectstest.MulldrifterOracle)
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, wanderer, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.BattlefieldIDsNamed(g, "Mulldrifter")) != 1 {
		t.Fatal("your own creature did not return at this turn's end step")
	}
}

func TestEternalWandererIsCatalogedWhole(t *testing.T) {
	spec, ok := effects.Lookup("20a1671d-e8a4-4cf1-87a7-f2f6319f4b9e")
	if !ok {
		t.Fatal("The Eternal Wanderer is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v caveats %v, want full with none", spec.Completeness, spec.Caveats)
	}
}
