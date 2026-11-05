package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// put_blocking_becomes_blocked_test.go — #4956. A "whenever this
// creature becomes blocked" ability (afflict, CR 702.130a) triggers when
// a creature put onto the battlefield as a blocker blocks an attacker
// that was unblocked at that time (CR 509.3c), and not when the
// attacker was already blocked.

// putCatSoldierBlocking creates one 1/1 token under controller's
// control that's blocking attacker (CR 506.3d), then runs the priority
// boundary so any trigger it caused is put on the stack.
func putCatSoldierBlocking(t *testing.T, g *game.Game, controller, attacker uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokensBlockingForEffect(controller, game.Card{
			Name: "Cat Soldier", TypeLine: "Token Creature — Cat Soldier", Power: 1, Toughness: 1,
		}, 1, attacker); err != nil {
			t.Fatalf("CreateTokensBlockingForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
}

func TestAfflictTriggersWhenABlockingTokenBlocksAnUnblockedAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eternal := b12Push(g, me.ID, "Eternal of Harsh Truths", "Creature — Zombie Cleric", eternalOfHarshTruthsOracle, 1, 3)
	unblockedAttackIntoBlockers(t, g, eternal)

	// No blocks: Eternal is an unblocked creature, and its draw resolves.
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := triggersOnStackFrom(g, eternal); n != 0 {
		t.Fatalf("the draw trigger should have resolved, %d left", n)
	}
	life := opp.Life

	putCatSoldierBlocking(t, g, opp.ID, eternal)
	if n := triggersOnStackFrom(g, eternal); n != 1 {
		t.Fatalf("the token blocks an unblocked attacker, so afflict triggers (CR 509.3c): %d triggers", n)
	}
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("afflict 2: defender lost %d", got)
	}
}

func TestAfflictDoesNotTriggerAgainWhenABlockingTokenJoinsABlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eternal := b12Push(g, me.ID, "Eternal of Harsh Truths", "Creature — Zombie Cleric", eternalOfHarshTruthsOracle, 1, 3)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	unblockedAttackIntoBlockers(t, g, eternal)

	if err := g.DeclareBlocker(wall, eternal); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	life := opp.Life

	putCatSoldierBlocking(t, g, opp.ID, eternal)
	if n := triggersOnStackFrom(g, eternal); n != 0 {
		t.Fatalf("Eternal was already blocked, so it does not become blocked again: %d triggers", n)
	}
	if opp.Life != life {
		t.Errorf("no second afflict: %d → %d", life, opp.Life)
	}
}
