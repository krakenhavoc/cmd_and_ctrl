package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// obTriggersWaiting counts Ob's triggers waiting on the stack.
func obTriggersWaiting(g *game.Game, ob uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == ob {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == ob {
			n++
		}
	}
	return n
}

func TestObNixilisKingpinOneDrainToTwoOpponentsFiresOnce(t *testing.T) {
	g, me, ob, top := effectstest.ObKingpinBoard(t)
	g.WithWriteLock(func() {
		for _, opp := range g.Seats[1:3] {
			if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1); err != nil {
				t.Fatal(err)
			}
		}
	})
	effectstest.SettleLifeBoundary(g)
	if n := obTriggersWaiting(g, ob); n != 1 {
		t.Fatalf("one drain of exactly 1 to two opponents put %d triggers on the stack, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ob, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	perm := effectstest.ExiledPermission(g, top)
	if perm.Player != me.ID || perm.CastOnly {
		t.Errorf("permission = %+v, want the controller, and play (not cast only)", perm)
	}
}

func TestObNixilisKingpinTwoLifeDoesNotFire(t *testing.T) {
	g, _, ob, top := effectstest.ObKingpinBoard(t)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, -2); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.SettleLifeBoundary(g)
	effectstest.PassPriorityAroundTable(t, g)
	if obTriggersWaiting(g, ob) != 0 || effectstest.CountersOn(g, ob, "+1/+1") != 0 || g.Exile.Contains(top) {
		t.Fatal("a loss of 2 triggered a loss-of-exactly-1 ability")
	}
}

func TestObNixilisKingpinTwoOnePointLossesToOneOpponentAreATotalOfTwo(t *testing.T) {
	g, _, ob, _ := effectstest.ObKingpinBoard(t)
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, -1); err != nil {
				t.Fatal(err)
			}
		}
	})
	effectstest.SettleLifeBoundary(g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, ob, "+1/+1") != 0 {
		t.Fatal("two simultaneous losses of 1 are a loss of 2 and must not trigger")
	}
}

func TestObNixilisKingpinDamageAndLifeLossSumInOneBatch(t *testing.T) {
	g, _, ob, _ := effectstest.ObKingpinBoard(t)
	src := effectstest.PushVanillaCreature(g, g.Seats[0].ID, "Pinger", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, g.Seats[1].ID, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, -1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.SettleLifeBoundary(g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, ob, "+1/+1") != 0 {
		t.Fatal("1 damage and 1 life lost in one batch are 2 life, and must not trigger")
	}
}

func TestObNixilisKingpinOneOpponentAtOneAndAnotherAtTwoStillFiresOnce(t *testing.T) {
	g, _, ob, _ := effectstest.ObKingpinBoard(t)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, -2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[2].ID, -1)
	})
	effectstest.SettleLifeBoundary(g)
	if n := obTriggersWaiting(g, ob); n != 1 {
		t.Fatalf("triggers = %d, want 1: an opponent lost exactly 1", n)
	}
}

func TestObNixilisKingpinYourOwnLossDoesNotCount(t *testing.T) {
	g, me, ob, _ := effectstest.ObKingpinBoard(t)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	effectstest.SettleLifeBoundary(g)
	if n := obTriggersWaiting(g, ob); n != 0 {
		t.Fatalf("your own life loss triggered Ob %d times", n)
	}
}

func TestObNixilisKingpinSeparateBatchesFireSeparately(t *testing.T) {
	g, _, ob, _ := effectstest.ObKingpinBoard(t)
	opp := g.Seats[1].ID
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp, -1) })
	effectstest.SettleLifeBoundary(g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ob, "+1/+1"); got != 1 {
		t.Fatalf("first loss: counters = %d, want 1", got)
	}
	// Play moves on: the next step is a new batch, so the next loss of
	// exactly 1 is a new occurrence, not the tail of the first.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp, -1) })
	effectstest.SettleLifeBoundary(g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ob, "+1/+1"); got != 2 {
		t.Fatalf("second loss: counters = %d, want 2", got)
	}
}

func TestObNixilisKingpinOneCombatSourceDealingOneFires(t *testing.T) {
	g, _, ob, _ := effectstest.ObKingpinBoard(t)
	a := effectstest.PushVanillaCreature(g, g.Seats[0].ID, "Pinger", 1, 1)
	effectstest.AttackWith(t, g, g.Seats[1].ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ob, "+1/+1"); got != 1 {
		t.Fatalf("a lone 1-damage combat hit left %d counters, want 1", got)
	}
}

func TestObNixilisKingpinTwoCombatSourcesDealingOneEachIsATotalOfTwo(t *testing.T) {
	g, _, ob, top := effectstest.ObKingpinBoard(t)
	a := effectstest.PushVanillaCreature(g, g.Seats[0].ID, "Pinger A", 1, 1)
	b := effectstest.PushVanillaCreature(g, g.Seats[0].ID, "Pinger B", 1, 1)
	effectstest.AttackWith(t, g, g.Seats[1].ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, ob, "+1/+1") != 0 || g.Exile.Contains(top) {
		t.Fatal("two 1-damage sources in one combat damage step total 2 and must not trigger")
	}
}

func TestObNixilisKingpinCombatSplitAcrossTwoOpponentsFiresOnce(t *testing.T) {
	g, me, ob, _ := effectstest.ObKingpinBoard(t)
	a := effectstest.PushVanillaCreature(g, me.ID, "Pinger A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "Pinger B", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(b, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ob, "+1/+1"); got != 1 {
		t.Fatalf("one hit on each of two opponents gave %d counters, want exactly 1 (once per batch)", got)
	}
}
