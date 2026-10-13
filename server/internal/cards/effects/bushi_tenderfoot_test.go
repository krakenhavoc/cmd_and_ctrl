package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func bushiTenderfootRow() cards.Card {
	return flipRow(bushiTenderfootOracleID, "{W}", []string{"W"}, []string{"Bushido", "Double strike"},
		cards.CardFace{Name: "Bushi Tenderfoot", TypeLine: "Creature — Human Soldier",
			OracleText: "When a creature dealt damage by this creature this turn dies, flip this creature.",
			Power:      "1", Toughness: "1"},
		cards.CardFace{Name: "Kenzo the Hardhearted", TypeLine: "Legendary Creature — Human Samurai",
			OracleText: "Double strike; bushido 2 (Whenever this creature blocks or becomes blocked, it gets +2/+2 until end of turn.)",
			Power:      "3", Toughness: "4"})
}

func damage(t *testing.T, g *game.Game, src, victim uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(src, victim, n); err != nil {
			t.Fatal(err)
		}
	})
}

// damageAndSettle deals damage and lets the state-based actions kill
// what it killed, then lets the triggers that follow resolve.
func damageAndSettle(t *testing.T, g *game.Game, src, victim uuid.UUID, n int) {
	t.Helper()
	damage(t, g, src, victim, n)
	settleStateBasedActions(t, g)
	passPriorityAroundTable(t, g)
}

// Bushi Tenderfoot flips when a creature it damaged this turn dies, and
// only that: a creature it did not damage dying does nothing.
func TestBushiTenderfootFlipsWhenACreatureItDamagedDies(t *testing.T) {
	g, me, opp := flipMain(t)
	bushi := flipEntered(t, g, bushiTenderfootRow(), me)
	damaged := apaPush(g, opp.ID, opp.ID, apaCreature("Damaged", "oracle-damaged", 3, 3))
	other := apaPush(g, opp.ID, opp.ID, apaCreature("Other", "oracle-other", 1, 1))

	// A creature somebody else damaged dies: no flip.
	damageAndSettle(t, g, uuid.Nil, other, 1)
	passPriorityAroundTable(t, g)
	wantUnflipped(t, g, bushi, "Bushi Tenderfoot")

	// The Tenderfoot damages a creature; the damage alone flips nothing.
	damage(t, g, bushi, damaged, 1)
	passPriorityAroundTable(t, g)
	wantUnflipped(t, g, bushi, "Bushi Tenderfoot")

	// It dies later in the turn (to other damage): flip.
	damageAndSettle(t, g, uuid.Nil, damaged, 2)
	passPriorityAroundTable(t, g)
	c := wantFlipped(t, g, bushi, "Kenzo the Hardhearted", []string{"W"}, 1)
	if !game.HasKeyword(&c, "double strike") {
		t.Errorf("Kenzo keywords %v, want double strike", c.Keywords)
	}
	// Bushido 2 (CR 702.45a): becoming blocked gives +2/+2.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: bushi, CardID: bushi, Target: bushi})
	})
	passPriorityAroundTable(t, g)
	if p := flipState(t, g, bushi).Effective().Power; p != 5 {
		t.Errorf("Kenzo has power %d after becoming blocked, want 5 (3 + bushido 2)", p)
	}
}

// The damage has to have come from THIS object this turn: a creature
// that was damaged by the Tenderfoot on an earlier turn does not count.
func TestBushiTenderfootIgnoresDamageFromAnEarlierTurn(t *testing.T) {
	g, me, opp := flipMain(t)
	bushi := flipEntered(t, g, bushiTenderfootRow(), me)
	victim := apaPush(g, opp.ID, opp.ID, apaCreature("Victim", "oracle-victim", 1, 4))
	damage(t, g, bushi, victim, 1)
	// Next turn: the damage is gone with the turn, and so is the record.
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	damageAndSettle(t, g, uuid.Nil, victim, 5)
	passPriorityAroundTable(t, g)
	wantUnflipped(t, g, bushi, "Bushi Tenderfoot")
}

// A creature that was blinked after the Tenderfoot damaged it is a new
// object that the Tenderfoot never damaged (CR 400.7).
func TestBushiTenderfootIgnoresADamagedCreatureThatLeftAndCameBack(t *testing.T) {
	g, me, opp := flipMain(t)
	bushi := flipEntered(t, g, bushiTenderfootRow(), me)
	victim := apaPush(g, opp.ID, opp.ID, apaCreature("Victim", "oracle-victim", 1, 4))
	damage(t, g, bushi, victim, 1)
	g.WithWriteLock(func() {
		if err := (BounceToHand{Target: victim}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
		if _, err := g.PutFromHandOntoBattlefieldForEffect(victim, game.HandEntryOptions{}); err != nil {
			t.Fatal(err)
		}
	})
	if !onBattlefield(g, victim) {
		t.Fatal("the bounced creature did not come back")
	}
	damageAndSettle(t, g, uuid.Nil, victim, 9)
	passPriorityAroundTable(t, g)
	wantUnflipped(t, g, bushi, "Bushi Tenderfoot")
}
