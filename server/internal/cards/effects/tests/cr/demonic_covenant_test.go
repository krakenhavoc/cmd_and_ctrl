package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const demonicCovenantOracle = "393b9d9e-9d71-43e7-aede-eac7f85d2365"

func TestDemonicCovenantIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(demonicCovenantOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Demonic Covenant: %+v %v", spec, ok)
	}
}

func pushDemonicCovenant(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, "Demonic Covenant", "Kindred Enchantment — Demon", demonicCovenantOracle, false)
}

func TestDemonicCovenantDemonAttackDrawsAndCostsOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushDemonicCovenant(g, me.ID)
	demon := effectstest.PushDiesCreatureForTest(g, me.ID, "Test Demon", "", "Creature — Demon", 2, 2)
	other := effectstest.PushDiesCreatureForTest(g, me.ID, "Test Bear", "", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(demon, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(other, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1 (one trigger for the batch)", hand, me.Hand.Size())
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
}

func TestDemonicCovenantNonDemonAttackDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushDemonicCovenant(g, me.ID)
	other := effectstest.PushDiesCreatureForTest(g, me.ID, "Test Bear", "", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(other, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d → %d, life %d → %d, want both unchanged", hand, me.Hand.Size(), life, me.Life)
	}
}

// stackLibraryTop puts two cards on top of the library, the first on top.
func stackLibraryTop(g *game.Game, p *game.Player, typeLines ...string) {
	g.WithWriteLock(func() {
		for i := len(typeLines) - 1; i >= 0; i-- {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Stacked", TypeLine: typeLines[i], Owner: p.ID, Controller: p.ID})
		}
	})
}

func demonicCovenantEndStep(t *testing.T, top ...string) (g *game.Game, me *game.Player, covenant uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	covenant = pushDemonicCovenant(g, me.ID)
	stackLibraryTop(g, me, top...)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, covenant
}

func demonTokens(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.HasSubtype("Demon") && c.CurrentPower() == 5 {
			n++
		}
	}
	return n
}

func TestDemonicCovenantEndStepMakesADemonAndKeepsItselfOnMixedTypes(t *testing.T) {
	g, me, covenant := demonicCovenantEndStep(t, "Creature — Bear", "Land")
	if demonTokens(g, me.ID) != 1 {
		t.Errorf("5/5 Demon tokens = %d, want 1", demonTokens(g, me.ID))
	}
	if !effects.OnBattlefield(g, covenant) {
		t.Error("sacrificed although the milled cards share no types")
	}
}

func TestDemonicCovenantSacrificesItselfOnTwoCardsWithTheSameTypes(t *testing.T) {
	g, me, covenant := demonicCovenantEndStep(t, "Basic Land — Forest", "Land")
	if demonTokens(g, me.ID) != 1 {
		t.Errorf("5/5 Demon tokens = %d, want 1", demonTokens(g, me.ID))
	}
	if effects.OnBattlefield(g, covenant) {
		t.Error("not sacrificed although both milled cards are lands")
	}
}

func TestDemonicCovenantSharingOneTypeIsNotSharingAll(t *testing.T) {
	// An artifact creature and a creature share a type but not all of them.
	g, _, covenant := demonicCovenantEndStep(t, "Artifact Creature — Golem", "Creature — Bear")
	if !effects.OnBattlefield(g, covenant) {
		t.Error("sacrificed although the milled cards share only some of their types")
	}
}
