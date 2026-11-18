package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB03StokeOracle     = "2249001c-08d5-4c0a-86f8-d97519a39f37"
	rprB03GargantuaOracle = "a3925dfd-62b0-44ef-8545-7bee92cad5a2"
	rprB03VanquishOracle  = "4a3cc4f6-037f-461d-a3b7-9f9e24c4b8cf"
)

func TestStokeTheFlamesDealsFourToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	life := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Stoke the Flames", "Instant", rprB03StokeOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent at %d, want %d", got, life-4)
	}
	big := blbB1Push(g, opp.ID, "Their Big", "Creature — Beast", "", 4, 4)
	castCatalogSpell(t, g, "Stoke the Flames", "Instant", rprB03StokeOracle, cardTarget(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a 4/4 survived 4 damage")
	}
}

func TestStokeTheFlamesConvokeTapsCreaturesToPay(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	var tap []uuid.UUID
	for i := 0; i < 4; i++ {
		tap = append(tap, b16Creature(g, me.ID, "Helper", "Creature — Goblin", 1, 1, "R"))
	}
	id := putInHand(me, game.Card{Name: "Stoke the Flames", TypeLine: "Instant", OracleID: rprB03StokeOracle,
		ManaCost: "{2}{R}{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: b16TargetPlayer(opp.ID)}); err == nil {
		t.Fatal("cast {2}{R}{R} with an empty pool and no convoke")
	}
	life := lifeOf(g, opp.ID)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: b16TargetPlayer(opp.ID), TapIDs: tap}); err != nil {
		t.Fatalf("convoke four red creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent at %d, want %d", got, life-4)
	}
}

func TestPhyrexianGargantuaDrawsTwoAndLosesTwoOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand, life := me.Hand.Size(), me.Life
	msc17Play(t, g, me, game.Card{Name: "Phyrexian Gargantua", TypeLine: "Creature — Phyrexian Horror", OracleID: rprB03GargantuaOracle, Power: 4, Toughness: 4})
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestVanquishTheWeakDestroysPowerThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	small := blbB1Push(g, opp.ID, "Small", "Creature — Bear", "", 3, 3)
	castCatalogSpell(t, g, "Vanquish the Weak", "Instant", rprB03VanquishOracle, cardTarget(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a power-3 creature survived")
	}
}

func TestVanquishTheWeakRefusesPowerFour(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	big := blbB1Push(g, opp.ID, "Big", "Creature — Beast", "", 4, 4)
	if err := castCatalogSpellErrStrict(t, g, "Vanquish the Weak", "Instant", rprB03VanquishOracle, cardTarget(big)); err == nil {
		t.Fatal("targeted a power-4 creature")
	}
}
