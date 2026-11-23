package testscr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dollmakersShopOracle = "e633412b-e36b-4ec5-b0d0-7cb24c7503f4"

// TestDollmakersShopIgnoresAToyAttackingAlone casts the left half for
// real: a Toy attacking by itself is not a non-Toy creature attacking.
func TestDollmakersShopIgnoresAToyAttackingAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	room := effectstest.RoomCardC(me.ID, dollmakersShopOracle, "Dollmaker's Shop", "{1}{W}", "Porcelain Gallery", "{4}{W}{W}", "W")
	effectstest.CastRoomC(t, g, me.ID, room, 0)
	effectstest.SettleOrdering(t, g, me.ID)
	toy := effectstest.B16Creature(g, me.ID, "Toy", "Artifact Creature — Toy", 1, 1, "W")
	effectstest.DeclareAttack(t, g, opp.ID, toy)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Toy"); n != 1 {
		t.Fatalf("a Toy attacking alone made a Toy: %d Toys, want 1", n)
	}
}

// TestDollmakersShopMakesOneToyPerPlayerAttacked: two Bears and a Toy
// attacking one player are one trigger (CR 603.2c), so one new Toy; and
// the locked Porcelain Gallery does nothing.
func TestDollmakersShopMakesOneToyPerPlayerAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	room := effectstest.RoomCardC(me.ID, dollmakersShopOracle, "Dollmaker's Shop", "{1}{W}", "Porcelain Gallery", "{4}{W}{W}", "W")
	effectstest.CastRoomC(t, g, me.ID, room, 0)
	effectstest.SettleOrdering(t, g, me.ID)
	bear1 := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bear2 := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	toy := effectstest.B16Creature(g, me.ID, "Toy", "Artifact Creature — Toy", 1, 1, "W")
	if p, _ := effectstest.SizeOfC(t, g, bear1); p != 2 {
		t.Fatalf("the locked Porcelain Gallery resized a Bear to %d power", p)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear1, bear2, toy)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Toy"); n != 2 {
		t.Fatalf("%d Toys after the attack, want 2 (the one I had and one made)", n)
	}
}

// TestPorcelainGallerySetsBasePTToTheCreatureCount casts the right half:
// every creature you control is N/N for the N creatures you control
// (layer 7b, a SET), an opponent's creatures are untouched, and the
// other door unlocking later changes nothing about it.
func TestPorcelainGallerySetsBasePTToTheCreatureCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B16Creature(g, me.ID, "Ogre", "Creature — Ogre", 5, 5)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	room := effectstest.RoomCardC(me.ID, dollmakersShopOracle, "Dollmaker's Shop", "{1}{W}", "Porcelain Gallery", "{4}{W}{W}", "W")
	effectstest.CastRoomC(t, g, me.ID, room, 1)
	effectstest.SettleOrdering(t, g, me.ID)
	if p, tough := effectstest.SizeOfC(t, g, a); p != 2 || tough != 2 {
		t.Errorf("Bear: %d/%d, want 2/2 (two creatures)", p, tough)
	}
	if p, tough := effectstest.SizeOfC(t, g, b); p != 2 || tough != 2 {
		t.Errorf("Ogre: %d/%d, want 2/2 (base set, not added)", p, tough)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	if p, tough := effectstest.SizeOfC(t, g, a); p != 3 || tough != 3 {
		t.Errorf("three creatures: Bear is %d/%d, want 3/3", p, tough)
	}
	if p, _ := effectstest.SizeOfC(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature is %d power, want its own 2", p)
	}
	effectstest.UnlockDoorC(t, g, me.ID, room.InstanceID, game.DoorLeft)
	effectstest.SettleOrdering(t, g, me.ID)
	if p, _ := effectstest.SizeOfC(t, g, a); p != 3 {
		t.Errorf("after unlocking the shop the Bear is %d power, want 3", p)
	}
}
