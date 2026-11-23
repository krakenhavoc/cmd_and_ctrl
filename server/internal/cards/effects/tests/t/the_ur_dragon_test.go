package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ur_dragon_test.go — #2802: The Ur-Dragon, ADR 0140's cost form
// with Dragon for Sphinx, and its attack trigger.

const urDragonOracle = "87b22b09-4f6d-4bc5-9cfc-663e4c7c6981"

func TestUrDragonDiscountsDragonSpellsFromTheCommandZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Command.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "The Ur-Dragon", TypeLine: "Legendary Creature — Dragon Avatar",
		ManaCost: "{4}{W}{U}{B}{R}{G}", OracleID: urDragonOracle, IsCommander: true,
		Owner: me.ID, Controller: me.ID,
	})
	dragon := game.Card{
		InstanceID: uuid.New(), Name: "Test Dragon", TypeLine: "Creature — Dragon",
		ManaCost: "{3}{R}{R}", Owner: me.ID, Controller: me.ID,
	}
	if got := priceOf(t, g, me, dragon, game.ZoneHand); got != 4 {
		t.Errorf("a Dragon spell with The Ur-Dragon in my command zone costs %d, want 4", got)
	}
}

// Two Dragons attack: one trigger, two cards drawn, then the offer to
// put a permanent card from hand onto the battlefield.
func TestUrDragonDrawsPerAttackingDragonThenOffersAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ur := effectstest.PushCatalogPermanent(g, me.ID, "The Ur-Dragon", "Legendary Creature — Dragon Avatar", urDragonOracle, false)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Shivan Dragon", "Creature — Dragon", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	forest := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(forest)
	hand := me.Hand.Size()

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{ur, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("two Dragons attacked and I drew %d, want 2", got)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no offer to put a permanent card from hand onto the battlefield")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, forest.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, forest.InstanceID); !ok {
		t.Error("the chosen Forest is not on the battlefield")
	}
}
