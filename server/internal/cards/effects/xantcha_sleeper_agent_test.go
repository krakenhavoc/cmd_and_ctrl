package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// xantcha_sleeper_agent_test.go — Xantcha, Sleeper Agent (ADR 0106 §1
// and §2, #1793). The entry is ADR 0102's and is pinned in
// entry_controller_test.go; the attack restriction is PR 5's and is
// pinned in attack_target_restrictions_test.go. This file pins the card
// as printed: its row is an any-player row, "you" is the activator, and
// "Xantcha's controller" is read off the permanent.

const xantchaOracle = "0f0f3712-8d13-41a5-b332-2ab34e48d79d"

// xantchaItem is the stack item Xantcha's ability just put on the
// stack.
func xantchaItem(t *testing.T, g *game.Game, xantcha uuid.UUID) *game.StackItem {
	t.Helper()
	for _, it := range g.StackMeta {
		if it.SourceCardID == xantcha && it.Kind == game.StackItemActivated {
			return it
		}
	}
	t.Fatalf("no Xantcha ability on the stack")
	return nil
}

func TestXantchaAnyPlayerMayActivate(t *testing.T) {
	g := newCatalogGame(t)
	owner, controller, other := g.Seats[0], g.Seats[1], g.Seats[2]
	xantcha := pushStolenForTest(g, owner.ID, controller.ID, "Xantcha, Sleeper Agent", xantchaOracle)

	rows := game.ActivatedAbilitiesForCard(*findBattlefieldCardForTest(g, xantcha))
	if len(rows) != 1 || !rows[0].AnyPlayer {
		t.Fatalf("Xantcha's rows = %+v, want one any-player row", rows)
	}
	if rows[0].Purpose != (game.Purpose{Draws: 1, ControllerLosesLife: 2}) {
		t.Errorf("purpose = %+v", rows[0].Purpose)
	}

	// A player who neither owns nor controls it activates it, paying
	// {3} out of THEIR pool (CR 602.1a). The ability is theirs on the
	// stack (CR 113.8), so they draw (CR 109.5), and Xantcha's
	// controller loses the life.
	other.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	ctrlLife, otherLife := controller.Life, other.Life
	otherHand, ctrlHand := other.Hand.Size(), controller.Hand.Size()
	if err := g.ActivateCatalogAbility(other.ID, xantcha, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("a third player activates Xantcha: %v", err)
	}
	if n := len(other.ManaPool); n != 0 {
		t.Errorf("activator's pool holds %d after paying {3}, want 0", n)
	}
	if it := xantchaItem(t, g, xantcha); it.Controller != other.ID {
		t.Errorf("stack item controller = %v, want the activator %v", it.Controller, other.ID)
	}
	passPriorityAroundTable(t, g)
	if controller.Life != ctrlLife-2 {
		t.Errorf("controller life %d, want %d", controller.Life, ctrlLife-2)
	}
	if other.Life != otherLife {
		t.Errorf("activator lost life: %d, want %d", other.Life, otherLife)
	}
	if other.Hand.Size() != otherHand+1 {
		t.Errorf("activator hand %d, want %d", other.Hand.Size(), otherHand+1)
	}
	if controller.Hand.Size() != ctrlHand {
		t.Errorf("controller drew: hand %d, want %d", controller.Hand.Size(), ctrlHand)
	}

	// The 2018-07-13 ruling: Xantcha's controller may activate it too,
	// and then loses the life AND draws.
	controller.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	ctrlLife, ctrlHand = controller.Life, controller.Hand.Size()
	if err := g.ActivateCatalogAbility(controller.ID, xantcha, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Xantcha's controller activates it: %v", err)
	}
	passPriorityAroundTable(t, g)
	if controller.Life != ctrlLife-2 || controller.Hand.Size() != ctrlHand+1 {
		t.Errorf("controller after own activation: life %d hand %d, want %d and %d",
			controller.Life, controller.Hand.Size(), ctrlLife-2, ctrlHand+1)
	}

	// Its owner, who does not control it, is just another player.
	owner.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(owner.ID, xantcha, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Xantcha's owner activates it: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// TestXantchaUnpaidCostIsTheActivatorsToPay pins CR 602.1a the other
// way: the controller's mana does not pay for another player's
// activation.
func TestXantchaUnpaidCostIsTheActivatorsToPay(t *testing.T) {
	g := newCatalogGame(t)
	owner, controller, other := g.Seats[0], g.Seats[1], g.Seats[2]
	xantcha := pushStolenForTest(g, owner.ID, controller.ID, "Xantcha, Sleeper Agent", xantchaOracle)
	controller.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	err := g.ActivateCatalogAbility(other.ID, xantcha, 0, game.ActivateAbilityParams{Strict: true})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("activator with an empty pool: err = %v, want InsufficientManaError", err)
	}
	if len(controller.ManaPool) != 3 {
		t.Errorf("the controller's pool paid for somebody else's activation")
	}
}

// TestXantchaControllerIsReadWhenItHasLeft pins CR 608.2h: with
// Xantcha gone before the ability resolves, "Xantcha's controller" is
// the player who controlled it as it last existed.
func TestXantchaControllerIsReadWhenItHasLeft(t *testing.T) {
	g := newCatalogGame(t)
	owner, controller, other := g.Seats[0], g.Seats[1], g.Seats[2]
	xantcha := pushStolenForTest(g, owner.ID, controller.ID, "Xantcha, Sleeper Agent", xantchaOracle)
	if err := g.ActivateCatalogAbility(other.ID, xantcha, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	var derr error
	g.WithWriteLock(func() { derr = g.DestroyPermanentForEffect(xantcha) })
	if err := derr; err != nil {
		t.Fatalf("destroy Xantcha: %v", err)
	}
	ctrlLife, otherHand := controller.Life, other.Hand.Size()
	passPriorityAroundTable(t, g)
	if controller.Life != ctrlLife-2 {
		t.Errorf("last-known controller life %d, want %d", controller.Life, ctrlLife-2)
	}
	if other.Hand.Size() != otherHand+1 {
		t.Errorf("activator hand %d, want %d", other.Hand.Size(), otherHand+1)
	}
}
