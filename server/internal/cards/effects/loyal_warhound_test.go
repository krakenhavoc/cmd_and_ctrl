package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loyal_warhound_test.go — the clauses of Loyal Warhound that
// aang_batch1_test.go's two tests did not pin, written when the card
// was read against its oracle text and marked CompletenessFull (#1306).

// TestLoyalWarhoundRechecksItsConditionOnResolution — CR 603.4: the
// intervening "if" is checked again as the trigger resolves. The
// opponent is ahead on lands when the Warhound enters, the lands leave
// with the trigger on the stack, and the search does not happen.
func TestLoyalWarhoundRechecksItsConditionOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plains := pushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	a := aangPushLand(g, opp.ID, "Island", false)
	b := aangPushLand(g, opp.ID, "Island", false)

	castAndResolveCreature(t, g, "Loyal Warhound", "Creature — Dog", loyalWarhoundOracle)
	if len(g.PendingTriggers) == 0 && len(g.StackMeta) == 0 {
		t.Fatal("the Warhound's trigger did not fire while an opponent was ahead on lands")
	}
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{a, b} {
			if _, err := game.MoveCard(g.Battlefield, opp.Graveyard, id); err != nil {
				t.Fatalf("MoveCard: %v", err)
			}
		}
	})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(plains) {
		t.Error("the Plains was fetched although nobody had more lands than the Warhound's controller on resolution")
	}
}

// TestLoyalWarhoundTakesOnlyABasicPlains — "a basic Plains card": a
// nonbasic Plains-typed land and a basic of another type are not
// offered.
func TestLoyalWarhoundTakesOnlyABasicPlains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedSearchLibrary(me,
		game.Card{Name: "Hallowed Fountain", TypeLine: "Land — Plains Island"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Snow-Covered Plains", TypeLine: "Basic Snow Land — Plains"},
	)
	aangPushLand(g, opp.ID, "Island", false)

	castCatalogSpell(t, g, "Loyal Warhound", "Creature — Dog", loyalWarhoundOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two basic Plains match (one of them snow), so the search asks which")
	}
	for _, name := range []string{"Hallowed Fountain", "Island"} {
		if searchOptionNamed(g, c, name) != uuid.Nil {
			t.Errorf("%s is offered, but it is not a basic Plains", name)
		}
	}
	answerSearchNamed(t, g, me.ID, "Snow-Covered Plains")
	got := findBattlefieldByName(g, "Snow-Covered Plains")
	if got == uuid.Nil {
		t.Fatal("the chosen basic Plains did not enter")
	}
	if c, _ := battlefieldCard(g, got); !c.Tapped {
		t.Error("the Plains entered untapped; the card says tapped")
	}
}
