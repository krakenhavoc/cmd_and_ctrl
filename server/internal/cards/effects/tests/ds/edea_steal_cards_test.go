package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// e2Steal gives `to` layer-2 control of `id` until end of turn, the
// way Act of Treason does.
func e2Steal(t *testing.T, g *game.Game, id, to uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), id, to, g.UntilEndOfTurnDuration(), "test — steal")
	})
	if got := effectstest.ControllerOf(t, g, id); got != to {
		t.Fatalf("setup: controller %s after the steal, want %s", got, to)
	}
}

// --- Edea, Possessed Sorceress ------------------------------------

func e2PushEdea(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B43Catalog(g, owner, "Edea, Possessed Sorceress", "Legendary Creature — Human Warlock",
		effects.EdeaPossessedSorceressOracleID, 2, 5)
}

func TestEdeaStealsUntapsAndHastesAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	e2PushEdea(g, me.ID)
	mine := effectstest.CtrlPushCreature(g, me.ID, "My Bear")
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt at the beginning of combat on Edea's controller's turn")
	}
	if effectstest.HasID(prompt.PickTargetCards, mine) {
		t.Error("\"a creature an opponent controls\": your own creature is offered")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s, want Edea's controller %s", got, me.ID)
	}
	if effectstest.B16Tapped(t, g, victim) {
		t.Error("the stolen creature was not untapped")
	}
	if effectstest.SummoningSickOf(t, g, victim) {
		t.Error("the stolen creature has no haste — it cannot attack")
	}

	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != opp.ID {
		t.Errorf("after cleanup the creature stayed stolen: controller %s, want %s", got, opp.ID)
	}
}

func TestEdeaCombatTriggerIsOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	e2PushEdea(g, opp.ID) // seat 1's Edea, and it is seat 0's turn
	effectstest.CtrlPushCreature(g, g.Seats[0].ID, "Bear")

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if effectstest.LatestPickTarget(g, opp.ID) != nil {
		t.Error("Edea triggered at the beginning of combat on an opponent's turn")
	}
}

func TestEdeaReturnsAStolenCreatureThatDiesToItsOwnerAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// On an opponent's turn, so the priority passes that settle the
	// dies trigger cannot walk into Edea's own beginning-of-combat
	// trigger.
	effectstest.AdvanceToMainOf(t, g, 1)
	e2PushEdea(g, me.ID)
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	e2Steal(t, g, victim, me.ID)

	hand := len(me.Hand.Cards)
	effectstest.E2Destroy(t, g, victim)

	if z := effectstest.E2Zone(g, victim); z != game.ZoneBattlefield {
		t.Fatalf("the stolen creature is in %q, want back on the battlefield", z)
	}
	if got := effectstest.ControllerOf(t, g, victim); got != opp.ID {
		t.Errorf("returned under %s, want its owner %s", got, opp.ID)
	}
	if got := len(me.Hand.Cards) - hand; got != 1 {
		t.Errorf("Edea's controller drew %d, want 1", got)
	}
}

func TestEdeaIgnoresACreatureYouOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// On an opponent's turn, so the priority passes that settle the
	// dies trigger cannot walk into Edea's own beginning-of-combat
	// trigger.
	effectstest.AdvanceToMainOf(t, g, 1)
	e2PushEdea(g, me.ID)
	mine := effectstest.CtrlPushCreature(g, me.ID, "My Bear")
	theirs := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")

	hand := len(me.Hand.Cards)
	effectstest.E2Destroy(t, g, mine)
	effectstest.E2Destroy(t, g, theirs)

	if z := effectstest.E2Zone(g, mine); z != game.ZoneGraveyard {
		t.Errorf("your own creature is in %q, want the graveyard", z)
	}
	if z := effectstest.E2Zone(g, theirs); z != game.ZoneGraveyard {
		t.Errorf("an opponent's creature they controlled is in %q, want the graveyard", z)
	}
	if got := len(me.Hand.Cards) - hand; got != 0 {
		t.Errorf("drew %d with no stolen creature dying, want 0", got)
	}
}

// TestEdeaDoesNotReturnACardThatLeftTheGraveyard pins CR 400.7: the
// trigger returns the object that died, so one exiled out of the
// graveyard in response stays exiled — and the draw still happens.
func TestEdeaDoesNotReturnACardThatLeftTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// On an opponent's turn, so the priority passes that settle the
	// dies trigger cannot walk into Edea's own beginning-of-combat
	// trigger.
	effectstest.AdvanceToMainOf(t, g, 1)
	e2PushEdea(g, me.ID)
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	e2Steal(t, g, victim, me.ID)

	hand := len(me.Hand.Cards)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	// The trigger is waiting; exile the card from the graveyard in
	// response.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(victim) })
	effectstest.PassPriorityAroundTable(t, g)

	if z := effectstest.E2Zone(g, victim); z != game.ZoneExile {
		t.Errorf("the card is in %q, want it left in exile", z)
	}
	if got := len(me.Hand.Cards) - hand; got != 1 {
		t.Errorf("drew %d, want 1 — the draw does not depend on the return", got)
	}
}
