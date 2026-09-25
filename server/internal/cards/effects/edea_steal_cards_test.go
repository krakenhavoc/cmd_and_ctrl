package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// edea_steal_cards_test.go pins the Edea deck's steal-and-copy slice
// (E2, #1565).

// e2Zone reports the kind of zone a card is in, or "" when it is
// nowhere.
func e2Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

// e2Steal gives `to` layer-2 control of `id` until end of turn, the
// way Act of Treason does.
func e2Steal(t *testing.T, g *game.Game, id, to uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), id, to, g.UntilEndOfTurnDuration(), "test — steal")
	})
	if got := controllerOf(t, g, id); got != to {
		t.Fatalf("setup: controller %s after the steal, want %s", got, to)
	}
}

// e2Destroy destroys a permanent and settles whatever that triggered.
func e2Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	passPriorityAroundTable(t, g)
}

// --- Edea, Possessed Sorceress ------------------------------------

func e2PushEdea(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b43Catalog(g, owner, "Edea, Possessed Sorceress", "Legendary Creature — Human Warlock",
		edeaPossessedSorceressOracleID, 2, 5)
}

func TestEdeaStealsUntapsAndHastesAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	e2PushEdea(g, me.ID)
	mine := ctrlPushCreature(g, me.ID, "My Bear")
	victim := ctrlPushCreature(g, opp.ID, "Their Bear")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})

	advanceTo(t, g, game.StepBeginCombat)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt at the beginning of combat on Edea's controller's turn")
	}
	if hasID(prompt.PickTargetCards, mine) {
		t.Error("\"a creature an opponent controls\": your own creature is offered")
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s, want Edea's controller %s", got, me.ID)
	}
	if b16Tapped(t, g, victim) {
		t.Error("the stolen creature was not untapped")
	}
	if summoningSickOf(t, g, victim) {
		t.Error("the stolen creature has no haste — it cannot attack")
	}

	advancePastCleanupForTest(t, g)
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Errorf("after cleanup the creature stayed stolen: controller %s, want %s", got, opp.ID)
	}
}

func TestEdeaCombatTriggerIsOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	e2PushEdea(g, opp.ID) // seat 1's Edea, and it is seat 0's turn
	ctrlPushCreature(g, g.Seats[0].ID, "Bear")

	advanceTo(t, g, game.StepBeginCombat)
	if latestPickTarget(g, opp.ID) != nil {
		t.Error("Edea triggered at the beginning of combat on an opponent's turn")
	}
}

func TestEdeaReturnsAStolenCreatureThatDiesToItsOwnerAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// On an opponent's turn, so the priority passes that settle the
	// dies trigger cannot walk into Edea's own beginning-of-combat
	// trigger.
	advanceToMainOf(t, g, 1)
	e2PushEdea(g, me.ID)
	victim := ctrlPushCreature(g, opp.ID, "Their Bear")
	e2Steal(t, g, victim, me.ID)

	hand := len(me.Hand.Cards)
	e2Destroy(t, g, victim)

	if z := e2Zone(g, victim); z != game.ZoneBattlefield {
		t.Fatalf("the stolen creature is in %q, want back on the battlefield", z)
	}
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Errorf("returned under %s, want its owner %s", got, opp.ID)
	}
	if got := len(me.Hand.Cards) - hand; got != 1 {
		t.Errorf("Edea's controller drew %d, want 1", got)
	}
}

func TestEdeaIgnoresACreatureYouOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// On an opponent's turn, so the priority passes that settle the
	// dies trigger cannot walk into Edea's own beginning-of-combat
	// trigger.
	advanceToMainOf(t, g, 1)
	e2PushEdea(g, me.ID)
	mine := ctrlPushCreature(g, me.ID, "My Bear")
	theirs := ctrlPushCreature(g, opp.ID, "Their Bear")

	hand := len(me.Hand.Cards)
	e2Destroy(t, g, mine)
	e2Destroy(t, g, theirs)

	if z := e2Zone(g, mine); z != game.ZoneGraveyard {
		t.Errorf("your own creature is in %q, want the graveyard", z)
	}
	if z := e2Zone(g, theirs); z != game.ZoneGraveyard {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// On an opponent's turn, so the priority passes that settle the
	// dies trigger cannot walk into Edea's own beginning-of-combat
	// trigger.
	advanceToMainOf(t, g, 1)
	e2PushEdea(g, me.ID)
	victim := ctrlPushCreature(g, opp.ID, "Their Bear")
	e2Steal(t, g, victim, me.ID)

	hand := len(me.Hand.Cards)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	// The trigger is waiting; exile the card from the graveyard in
	// response.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(victim) })
	passPriorityAroundTable(t, g)

	if z := e2Zone(g, victim); z != game.ZoneExile {
		t.Errorf("the card is in %q, want it left in exile", z)
	}
	if got := len(me.Hand.Cards) - hand; got != 1 {
		t.Errorf("drew %d, want 1 — the draw does not depend on the return", got)
	}
}
