package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_cards_s7b_restore_test.go — #4361 slice S7b. A put from hand with
// a rider on the result or a rule on the picked set was a closure pick; it
// is keyed now (the rider rides the prompt's tail, the rule is registered
// with the put's continuation by PutFromHandRule). Each is captured with
// its prompt open, restored strictly through JSON and answered there.

// Spelunking's Cave rider survives a restore: the restored pick still
// pays the 4 life when the land put is a Cave.
func TestS7bSpelunkingCaveRiderSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cave := handCardForTest(me, "Some Cave", "Land — Cave", "")
	castCatalogSpell(t, g, "Spelunking", "Enchantment", putFromHandSpelunkingOracle, nil)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("Spelunking's land offer is not open")
	}

	r := restoreThroughJSON(t, g)
	rme := r.Seats[0]
	life := rme.Life
	answerChooseCards(t, r, rme.ID, cave)
	if !r.Battlefield.Contains(cave) {
		t.Fatal("the restored pick did not put the Cave onto the battlefield")
	}
	if got := rme.Life - life; got != spelunkingCaveLife {
		t.Errorf("life %+d after the restored pick, want %+d for a Cave", got, spelunkingCaveLife)
	}
}

// Yuna's Decision's "a creature card and/or a land card" is a registered
// set rule now, so a restored offer still refuses two creatures and takes a
// creature and a land.
func TestS7bYunasDecisionPutKeepsItsSetRuleAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bearA := handCardForTest(me, "Bear A", "Creature — Bear", "")
	bearB := handCardForTest(me, "Bear B", "Creature — Bear", "")
	forest := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	source := uuid.New()
	g.WithWriteLock(func() {
		// The put yunasDecisionPutRef asks, as it asks it.
		err := PutFromHandOntoBattlefield{
			Player:   me.ID,
			Match:    Or(Creature(), Land()),
			Optional: true,
			Max:      2,
			Rule:     yunasDecisionPick,
			Label:    "Yuna's Decision — you may put a creature card and/or a land card from your hand onto the battlefield",
		}.Apply(NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: source}))
		if err != nil {
			t.Fatalf("the put: %v", err)
		}
	})
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("the put offer is not open")
	}

	r := restoreThroughJSON(t, g)
	rme := r.Seats[0]
	pick := chooseCardsChoiceFor(r, rme.ID)
	if pick == nil {
		t.Fatal("the restored table has no put offer")
	}
	if err := r.ResolveChooseCards(pick.ID, rme.ID, []uuid.UUID{bearA, bearB}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("two creatures after a restore: %v, want ErrChoiceSetRejected", err)
	}
	if err := r.ResolveChooseCards(pick.ID, rme.ID, []uuid.UUID{bearA, forest}); err != nil {
		t.Fatalf("a creature and a land after a restore: %v", err)
	}
	if !r.Battlefield.Contains(bearA) || !r.Battlefield.Contains(forest) {
		t.Error("the creature and the land did not both enter")
	}
}
