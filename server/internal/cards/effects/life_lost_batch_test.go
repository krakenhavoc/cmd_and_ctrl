package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2183 — the total life a player lost in ONE simultaneous event
// batch, read by Ob Nixilis, Captive Kingpin's "exactly 1 life".

const obNixilisKingpinOracle = "55b6434b-1542-40fd-b12a-697d40976582"

func obKingpinBoard(t *testing.T) (g *game.Game, me *game.Player, ob uuid.UUID, top uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	advanceToMain(t, g)
	ob = pushCatalogPermanent(g, me.ID, "Ob Nixilis, Captive Kingpin", "Legendary Creature — Demon", obNixilisKingpinOracle, false)
	top = seedLibrary(me, "Top Card")[0]
	return g, me, ob, top
}

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

func settleLifeBoundary(g *game.Game) { g.RunStateChecksForTest() }

func TestObNixilisKingpinOneDrainToTwoOpponentsFiresOnce(t *testing.T) {
	g, me, ob, top := obKingpinBoard(t)
	g.WithWriteLock(func() {
		for _, opp := range g.Seats[1:3] {
			if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1); err != nil {
				t.Fatal(err)
			}
		}
	})
	settleLifeBoundary(g)
	if n := obTriggersWaiting(g, ob); n != 1 {
		t.Fatalf("one drain of exactly 1 to two opponents put %d triggers on the stack, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ob, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	perm := exiledPermission(g, top)
	if perm.Player != me.ID || perm.CastOnly {
		t.Errorf("permission = %+v, want the controller, and play (not cast only)", perm)
	}
}

func TestObNixilisKingpinTwoLifeDoesNotFire(t *testing.T) {
	g, _, ob, top := obKingpinBoard(t)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, -2); err != nil {
			t.Fatal(err)
		}
	})
	settleLifeBoundary(g)
	passPriorityAroundTable(t, g)
	if obTriggersWaiting(g, ob) != 0 || countersOn(g, ob, "+1/+1") != 0 || g.Exile.Contains(top) {
		t.Fatal("a loss of 2 triggered a loss-of-exactly-1 ability")
	}
}

func TestObNixilisKingpinTwoOnePointLossesToOneOpponentAreATotalOfTwo(t *testing.T) {
	g, _, ob, _ := obKingpinBoard(t)
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, -1); err != nil {
				t.Fatal(err)
			}
		}
	})
	settleLifeBoundary(g)
	passPriorityAroundTable(t, g)
	if countersOn(g, ob, "+1/+1") != 0 {
		t.Fatal("two simultaneous losses of 1 are a loss of 2 and must not trigger")
	}
}

func TestObNixilisKingpinDamageAndLifeLossSumInOneBatch(t *testing.T) {
	g, _, ob, _ := obKingpinBoard(t)
	src := pushVanillaCreature(g, g.Seats[0].ID, "Pinger", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, g.Seats[1].ID, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, -1); err != nil {
			t.Fatal(err)
		}
	})
	settleLifeBoundary(g)
	passPriorityAroundTable(t, g)
	if countersOn(g, ob, "+1/+1") != 0 {
		t.Fatal("1 damage and 1 life lost in one batch are 2 life, and must not trigger")
	}
}

func TestObNixilisKingpinOneOpponentAtOneAndAnotherAtTwoStillFiresOnce(t *testing.T) {
	g, _, ob, _ := obKingpinBoard(t)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, -2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[2].ID, -1)
	})
	settleLifeBoundary(g)
	if n := obTriggersWaiting(g, ob); n != 1 {
		t.Fatalf("triggers = %d, want 1: an opponent lost exactly 1", n)
	}
}

func TestObNixilisKingpinYourOwnLossDoesNotCount(t *testing.T) {
	g, me, ob, _ := obKingpinBoard(t)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	settleLifeBoundary(g)
	if n := obTriggersWaiting(g, ob); n != 0 {
		t.Fatalf("your own life loss triggered Ob %d times", n)
	}
}

func TestObNixilisKingpinSeparateBatchesFireSeparately(t *testing.T) {
	g, _, ob, _ := obKingpinBoard(t)
	opp := g.Seats[1].ID
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp, -1) })
	settleLifeBoundary(g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ob, "+1/+1"); got != 1 {
		t.Fatalf("first loss: counters = %d, want 1", got)
	}
	// Play moves on: the next step is a new batch, so the next loss of
	// exactly 1 is a new occurrence, not the tail of the first.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp, -1) })
	settleLifeBoundary(g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ob, "+1/+1"); got != 2 {
		t.Fatalf("second loss: counters = %d, want 2", got)
	}
}

func TestObNixilisKingpinOneCombatSourceDealingOneFires(t *testing.T) {
	g, _, ob, _ := obKingpinBoard(t)
	a := pushVanillaCreature(g, g.Seats[0].ID, "Pinger", 1, 1)
	attackWith(t, g, g.Seats[1].ID, a)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ob, "+1/+1"); got != 1 {
		t.Fatalf("a lone 1-damage combat hit left %d counters, want 1", got)
	}
}

func TestObNixilisKingpinTwoCombatSourcesDealingOneEachIsATotalOfTwo(t *testing.T) {
	g, _, ob, top := obKingpinBoard(t)
	a := pushVanillaCreature(g, g.Seats[0].ID, "Pinger A", 1, 1)
	b := pushVanillaCreature(g, g.Seats[0].ID, "Pinger B", 1, 1)
	attackWith(t, g, g.Seats[1].ID, a, b)
	passPriorityAroundTable(t, g)
	if countersOn(g, ob, "+1/+1") != 0 || g.Exile.Contains(top) {
		t.Fatal("two 1-damage sources in one combat damage step total 2 and must not trigger")
	}
}

func TestObNixilisKingpinCombatSplitAcrossTwoOpponentsFiresOnce(t *testing.T) {
	g, me, ob, _ := obKingpinBoard(t)
	a := pushVanillaCreature(g, me.ID, "Pinger A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Pinger B", 1, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(b, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ob, "+1/+1"); got != 1 {
		t.Fatalf("one hit on each of two opponents gave %d counters, want exactly 1 (once per batch)", got)
	}
}
