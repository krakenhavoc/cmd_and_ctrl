package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsbu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/bu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func bushiTenderfootRow() cards.Card {
	return effectstest.FlipRow(cardsbu.BushiTenderfootOracleID, "{W}", []string{"W"}, []string{"Bushido", "Double strike"},
		cards.CardFace{Name: "Bushi Tenderfoot", TypeLine: "Creature — Human Soldier",
			OracleText: "When a creature dealt damage by this creature this turn dies, flip this creature.",
			Power:      "1", Toughness: "1"},
		cards.CardFace{Name: "Kenzo the Hardhearted", TypeLine: "Legendary Creature — Human Samurai",
			OracleText: "Double strike; bushido 2 (Whenever this creature blocks or becomes blocked, it gets +2/+2 until end of turn.)",
			Power:      "3", Toughness: "4"})
}

// Bushi Tenderfoot flips when a creature it damaged this turn dies, and
// only that: a creature it did not damage dying does nothing.
func TestBushiTenderfootFlipsWhenACreatureItDamagedDies(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	bushi := effectstest.FlipEntered(t, g, bushiTenderfootRow(), me)
	damaged := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Damaged", "oracle-damaged", 3, 3))
	other := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Other", "oracle-other", 1, 1))

	// A creature somebody else damaged dies: no flip.
	effectstest.DamageAndSettle(t, g, uuid.Nil, other, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantUnflipped(t, g, bushi, "Bushi Tenderfoot")

	// The Tenderfoot damages a creature; the damage alone flips nothing.
	effectstest.Damage(t, g, bushi, damaged, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantUnflipped(t, g, bushi, "Bushi Tenderfoot")

	// It dies later in the turn (to other damage): flip.
	effectstest.DamageAndSettle(t, g, uuid.Nil, damaged, 2)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.WantFlipped(t, g, bushi, "Kenzo the Hardhearted", []string{"W"}, 1)
	if !game.HasKeyword(&c, "double strike") {
		t.Errorf("Kenzo keywords %v, want double strike", c.Keywords)
	}
	// Bushido 2 (CR 702.45a): becoming blocked gives +2/+2.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: bushi, CardID: bushi, Target: bushi})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.FlipState(t, g, bushi).Effective().Power; p != 5 {
		t.Errorf("Kenzo has power %d after becoming blocked, want 5 (3 + bushido 2)", p)
	}
}

// The damage has to have come from THIS object this turn: a creature
// that was damaged by the Tenderfoot on an earlier turn does not count.
func TestBushiTenderfootIgnoresDamageFromAnEarlierTurn(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	bushi := effectstest.FlipEntered(t, g, bushiTenderfootRow(), me)
	victim := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Victim", "oracle-victim", 1, 4))
	effectstest.Damage(t, g, bushi, victim, 1)
	// Next turn: the damage is gone with the turn, and so is the record.
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.DamageAndSettle(t, g, uuid.Nil, victim, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantUnflipped(t, g, bushi, "Bushi Tenderfoot")
}

// A creature that was blinked after the Tenderfoot damaged it is a new
// object that the Tenderfoot never damaged (CR 400.7).
func TestBushiTenderfootIgnoresADamagedCreatureThatLeftAndCameBack(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	bushi := effectstest.FlipEntered(t, g, bushiTenderfootRow(), me)
	victim := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Victim", "oracle-victim", 1, 4))
	effectstest.Damage(t, g, bushi, victim, 1)
	g.WithWriteLock(func() {
		if err := (effects.BounceToHand{Target: victim}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
		if _, err := g.PutFromHandOntoBattlefieldForEffect(victim, game.HandEntryOptions{}); err != nil {
			t.Fatal(err)
		}
	})
	if !effects.OnBattlefield(g, victim) {
		t.Fatal("the bounced creature did not come back")
	}
	effectstest.DamageAndSettle(t, g, uuid.Nil, victim, 9)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantUnflipped(t, g, bushi, "Bushi Tenderfoot")
}
