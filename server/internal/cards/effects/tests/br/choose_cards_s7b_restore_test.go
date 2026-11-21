package testsbr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/so"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cave := effectstest.HandCardForTest(me, "Some Cave", "Land — Cave", "")
	effectstest.CastCatalogSpell(t, g, "Spelunking", "Enchantment", effectstest.PutFromHandSpelunkingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("Spelunking's land offer is not open")
	}

	r := effectstest.RestoreThroughJSON(t, g)
	rme := r.Seats[0]
	life := rme.Life
	effectstest.AnswerChooseCards(t, r, rme.ID, cave)
	if !r.Battlefield.Contains(cave) {
		t.Fatal("the restored pick did not put the Cave onto the battlefield")
	}
	if got := rme.Life - life; got != cardsso.SpelunkingCaveLife {
		t.Errorf("life %+d after the restored pick, want %+d for a Cave", got, cardsso.SpelunkingCaveLife)
	}
}

// Yuna's Decision's "a creature card and/or a land card" is a registered
// set rule now, so a restored offer still refuses two creatures and takes a
// creature and a land.
func TestS7bYunasDecisionPutKeepsItsSetRuleAcrossARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bearA := effectstest.HandCardForTest(me, "Bear A", "Creature — Bear", "")
	bearB := effectstest.HandCardForTest(me, "Bear B", "Creature — Bear", "")
	forest := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	source := uuid.New()
	g.WithWriteLock(func() {
		// The put yunasDecisionPutRef asks, as it asks it.
		err := effects.PutFromHandOntoBattlefield{
			Player:   me.ID,
			Match:    effects.Or(effects.Creature(), effects.Land()),
			Optional: true,
			Max:      2,
			Rule:     effects.YunasDecisionPick,
			Label:    "Yuna's Decision — you may put a creature card and/or a land card from your hand onto the battlefield",
		}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: source}))
		if err != nil {
			t.Fatalf("the put: %v", err)
		}
	})
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("the put offer is not open")
	}

	r := effectstest.RestoreThroughJSON(t, g)
	rme := r.Seats[0]
	pick := effectstest.ChooseCardsChoiceFor(r, rme.ID)
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
