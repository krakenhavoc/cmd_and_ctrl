package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// coveted_jewel_test.go — #1600: Coveted Jewel's steal trigger, "whenever
// one or more creatures an opponent controls attack you and aren't
// blocked, that player draws three cards and gains control of this
// artifact. Untap it." It fires off the defending player's completed
// block declaration (EventBlockersDeclared, #1279), once per
// declaration however many attackers got through.

// jewelTable seats the Jewel, tapped, under seat 1 of a four-seat game
// whose active player is seat 0.
func jewelTable(t *testing.T) (g *game.Game, me, opp *game.Player, jewel uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	jewel = b12Push(g, opp.ID, "Coveted Jewel", "Artifact", covetedJewelOracle, 0, 0)
	g.WithWriteLock(func() {
		if err := g.TapTargetForEffect(jewel); err != nil {
			t.Fatalf("tap the Jewel: %v", err)
		}
	})
	return g, me, opp, jewel
}

// jewelState is the Jewel's controller and tapped bit, read through a
// fresh layer pass (the control change is a layer-2 effect).
func jewelState(t *testing.T, g *game.Game, jewel uuid.UUID) (controller uuid.UUID, tapped bool) {
	t.Helper()
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == jewel {
				controller, tapped, found = c.Controller, c.Tapped, true
			}
		}
	})
	if !found {
		t.Fatal("the Jewel left the battlefield")
	}
	return controller, tapped
}

// attackInto declares each attacker at its defender, locks the
// declaration in and walks into declare blockers.
func attackInto(t *testing.T, g *game.Game, attacks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for attacker, defender := range attacks {
		if err := g.DeclareAttacker(attacker, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
}

// The whole card: one unblocked attacker, and the attacking player
// draws three, takes the Jewel and untaps it.
func TestCovetedJewelUnblockedAttackerTakesItDrawsThreeAndUntapsIt(t *testing.T) {
	g, me, opp, jewel := jewelTable(t)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	attackInto(t, g, map[uuid.UUID]uuid.UUID{bear: opp.ID})

	if triggerOnStack(g, jewel) == nil {
		t.Fatal("an unblocked attacker at the Jewel's controller put no trigger on the stack")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("attacking player's hand %d → %d, want +3", hand, got)
	}
	controller, tapped := jewelState(t, g, jewel)
	if controller != me.ID {
		t.Errorf("the Jewel is controlled by %s, want the attacking player %s", controller, me.ID)
	}
	if tapped {
		t.Error("\"Untap it\": the Jewel is still tapped")
	}
}

// Blocked means no trigger: the Jewel stays where it is.
func TestCovetedJewelDoesNothingWhenEveryAttackerIsBlocked(t *testing.T) {
	g, me, opp, jewel := jewelTable(t)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	hand := me.Hand.Size()
	attackInto(t, g, map[uuid.UUID]uuid.UUID{bear: opp.ID})
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)

	if n := len(g.PendingTriggers) + triggersOnStackFrom(g, jewel); n != 0 {
		t.Fatalf("a blocked attack triggered the Jewel %d times", n)
	}
	if controller, _ := jewelState(t, g, jewel); controller != opp.ID {
		t.Errorf("the Jewel changed hands on a blocked attack")
	}
	if me.Hand.Size() != hand {
		t.Errorf("the attacking player drew on a blocked attack")
	}
}

// "One or more": three attackers through is ONE trigger and three
// cards, and one blocked attacker beside an unblocked one still
// triggers.
func TestCovetedJewelTriggersOnceForOneOrMoreUnblockedAttackers(t *testing.T) {
	t.Run("three unblocked", func(t *testing.T) {
		g, me, opp, jewel := jewelTable(t)
		attacks := map[uuid.UUID]uuid.UUID{}
		for i := 0; i < 3; i++ {
			attacks[b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)] = opp.ID
		}
		hand := me.Hand.Size()
		attackInto(t, g, attacks)
		if n := triggersOnStackFrom(g, jewel); n != 1 {
			t.Fatalf("%d Jewel triggers for three unblocked attackers, want 1", n)
		}
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != hand+3 {
			t.Errorf("hand %d → %d, want exactly +3", hand, got)
		}
	})
	t.Run("one blocked, one not", func(t *testing.T) {
		g, me, opp, jewel := jewelTable(t)
		blocked := b12Creature(g, me.ID, "Blocked Bear", "Creature — Bear", 2, 2)
		through := b12Creature(g, me.ID, "Free Bear", "Creature — Bear", 2, 2)
		wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
		attackInto(t, g, map[uuid.UUID]uuid.UUID{blocked: opp.ID, through: opp.ID})
		if err := g.DeclareBlocker(wall, blocked); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
		lockInBlocks(t, g)
		if n := triggersOnStackFrom(g, jewel); n != 1 {
			t.Fatalf("%d Jewel triggers with one attacker unblocked, want 1", n)
		}
		passPriorityAroundTable(t, g)
		if controller, _ := jewelState(t, g, jewel); controller != me.ID {
			t.Error("the Jewel did not change hands")
		}
	})
}

// Four seats. Only the active player attacks (CR 506.2), and only an
// attack at the Jewel's controller counts: seat 0 swinging at seat 2
// leaves seat 1's Jewel alone, and seat 0 swinging at both takes it —
// seat 0, not seat 2 or 3, the other opponents of the Jewel's
// controller.
func TestCovetedJewelOnlyTheAttackingPlayerAndOnlyAnAttackAtYou(t *testing.T) {
	t.Run("attack elsewhere", func(t *testing.T) {
		g, me, _, jewel := jewelTable(t)
		third := g.Seats[2]
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		attackInto(t, g, map[uuid.UUID]uuid.UUID{bear: third.ID})
		if n := len(g.PendingTriggers) + triggersOnStackFrom(g, jewel); n != 0 {
			t.Fatalf("an attack at another player triggered the Jewel %d times", n)
		}
	})
	// "Attack YOU": a creature attacking a planeswalker the Jewel's
	// controller controls makes them the defending player, and still
	// is not attacking them.
	t.Run("attack your planeswalker", func(t *testing.T) {
		g, me, opp, jewel := jewelTable(t)
		walker := pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Opposing Walker", TypeLine: "Legendary Planeswalker — Test",
			Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 3},
		})
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		attackInto(t, g, map[uuid.UUID]uuid.UUID{bear: walker})
		if n := len(g.PendingTriggers) + triggersOnStackFrom(g, jewel); n != 0 {
			t.Fatalf("an attack at the controller's planeswalker triggered the Jewel %d times", n)
		}
	})
	t.Run("attack two players", func(t *testing.T) {
		g, me, opp, jewel := jewelTable(t)
		third, fourth := g.Seats[2], g.Seats[3]
		a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
		b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
		hands := map[uuid.UUID]int{me.ID: me.Hand.Size(), third.ID: third.Hand.Size(), fourth.ID: fourth.Hand.Size()}
		attackInto(t, g, map[uuid.UUID]uuid.UUID{a: opp.ID, b: third.ID})
		if n := triggersOnStackFrom(g, jewel); n != 1 {
			t.Fatalf("%d Jewel triggers, want 1", n)
		}
		passPriorityAroundTable(t, g)
		if controller, _ := jewelState(t, g, jewel); controller != me.ID {
			t.Errorf("the Jewel went to %s, want the active player", controller)
		}
		if me.Hand.Size() != hands[me.ID]+3 {
			t.Errorf("active player's hand %d → %d, want +3", hands[me.ID], me.Hand.Size())
		}
		for _, p := range []*game.Player{third, fourth} {
			if p.Hand.Size() != hands[p.ID] {
				t.Errorf("%s drew off the Jewel", p.Name)
			}
		}
	})
}

// The trigger is the Jewel's controller's: once it has changed hands, an
// attack at its OLD controller does nothing to it.
func TestCovetedJewelWatchesItsCurrentController(t *testing.T) {
	g, me, opp, jewel := jewelTable(t)
	g.WithWriteLock(func() {
		g.GainControlForEffect(jewel, jewel, g.Seats[2].ID, game.IndefiniteDuration(), "test")
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	attackInto(t, g, map[uuid.UUID]uuid.UUID{bear: opp.ID})
	if n := len(g.PendingTriggers) + triggersOnStackFrom(g, jewel); n != 0 {
		t.Fatalf("an attack at the Jewel's former controller triggered it %d times", n)
	}
}
