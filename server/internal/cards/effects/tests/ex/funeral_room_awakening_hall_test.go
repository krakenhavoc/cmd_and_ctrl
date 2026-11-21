package testsex

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const funeralRoomOracle = "a39d541e-86c1-4595-a2e9-f107def5bbc6"

// Casting Awakening Hall returns every creature card (and only those)
// as the Room enters with that door unlocked; Funeral Room drains only
// once its own door is unlocked.
func TestFuneralRoomAwakeningHallDoors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.PushGraveyardPermanent(me, "Grave Bear", "Creature — Bear", "{1}{G}")
	wolf := effectstest.PushGraveyardPermanent(me, "Grave Wolf", "Creature — Wolf", "{2}{G}")
	spell := effectstest.PushGraveyardPermanent(me, "Grave Charm", "Enchantment", "{1}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	c := effectstest.RoomsBCard(me.ID, funeralRoomOracle, "Funeral Room", "{2}{B}", "Awakening Hall", "{6}{B}{B}")

	effectstest.RoomsBCast(t, g, me, c, 1)
	effectstest.RoomsBSettle(t, g, me)
	if !effects.OnBattlefield(g, bear) || !effects.OnBattlefield(g, wolf) {
		t.Fatal("Awakening Hall did not return both creature cards")
	}
	if !me.Graveyard.Contains(spell) || !opp.Graveyard.Contains(theirs) {
		t.Fatal("Awakening Hall returned something that is not a creature card of yours")
	}

	meLife, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.RoomsBSettle(t, g, me)
	if me.Life != meLife || opp.Life != oppLife {
		t.Fatalf("a locked Funeral Room drained: me %d -> %d, opp %d -> %d", meLife, me.Life, oppLife, opp.Life)
	}

	effectstest.RoomsBUnlock(t, g, me, c, game.DoorLeft)
	effectstest.RoomsBSettle(t, g, me)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wolf) })
	effectstest.RoomsBSettle(t, g, me)
	if me.Life != meLife+1 {
		t.Errorf("you gained %d, want 1", me.Life-meLife)
	}
	if opp.Life != oppLife-1 {
		t.Errorf("opponent life %d -> %d, want -1", oppLife, opp.Life)
	}
}
