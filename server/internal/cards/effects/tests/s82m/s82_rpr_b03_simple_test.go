package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestStokeTheFlamesDealsFourToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Stoke the Flames", "Instant", effectstest.RprB03StokeOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent at %d, want %d", got, life-4)
	}
	big := effectstest.BlbB1Push(g, opp.ID, "Their Big", "Creature — Beast", "", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Stoke the Flames", "Instant", effectstest.RprB03StokeOracle, effectstest.CardTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a 4/4 survived 4 damage")
	}
}

func TestStokeTheFlamesConvokeTapsCreaturesToPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	var tap []uuid.UUID
	for i := 0; i < 4; i++ {
		tap = append(tap, effectstest.B16Creature(g, me.ID, "Helper", "Creature — Goblin", 1, 1, "R"))
	}
	id := effectstest.PutInHand(me, game.Card{Name: "Stoke the Flames", TypeLine: "Instant", OracleID: effectstest.RprB03StokeOracle,
		ManaCost: "{2}{R}{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.B16TargetPlayer(opp.ID)}); err == nil {
		t.Fatal("cast {2}{R}{R} with an empty pool and no convoke")
	}
	life := effectstest.LifeOf(g, opp.ID)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.B16TargetPlayer(opp.ID), TapIDs: tap}); err != nil {
		t.Fatalf("convoke four red creatures: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent at %d, want %d", got, life-4)
	}
}

func TestPhyrexianGargantuaDrawsTwoAndLosesTwoOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand, life := me.Hand.Size(), me.Life
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Phyrexian Gargantua", TypeLine: "Creature — Phyrexian Horror", OracleID: effectstest.RprB03GargantuaOracle, Power: 4, Toughness: 4})
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestVanquishTheWeakDestroysPowerThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	small := effectstest.BlbB1Push(g, opp.ID, "Small", "Creature — Bear", "", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Vanquish the Weak", "Instant", effectstest.RprB03VanquishOracle, effectstest.CardTarget(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a power-3 creature survived")
	}
}

func TestVanquishTheWeakRefusesPowerFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	big := effectstest.BlbB1Push(g, opp.ID, "Big", "Creature — Beast", "", 4, 4)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Vanquish the Weak", "Instant", effectstest.RprB03VanquishOracle, effectstest.CardTarget(big)); err == nil {
		t.Fatal("targeted a power-4 creature")
	}
}
