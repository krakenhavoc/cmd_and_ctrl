package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const funeralRoomOracle = "a39d541e-86c1-4595-a2e9-f107def5bbc6"

// Casting Awakening Hall returns every creature card (and only those)
// as the Room enters with that door unlocked; Funeral Room drains only
// once its own door is unlocked.
func TestFuneralRoomAwakeningHallDoors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := pushGraveyardPermanent(me, "Grave Bear", "Creature — Bear", "{1}{G}")
	wolf := pushGraveyardPermanent(me, "Grave Wolf", "Creature — Wolf", "{2}{G}")
	spell := pushGraveyardPermanent(me, "Grave Charm", "Enchantment", "{1}")
	theirs := pushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	c := roomsBCard(me.ID, funeralRoomOracle, "Funeral Room", "{2}{B}", "Awakening Hall", "{6}{B}{B}")

	roomsBCast(t, g, me, c, 1)
	roomsBSettle(t, g, me)
	if !onBattlefield(g, bear) || !onBattlefield(g, wolf) {
		t.Fatal("Awakening Hall did not return both creature cards")
	}
	if !me.Graveyard.Contains(spell) || !opp.Graveyard.Contains(theirs) {
		t.Fatal("Awakening Hall returned something that is not a creature card of yours")
	}

	meLife, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	roomsBSettle(t, g, me)
	if me.Life != meLife || opp.Life != oppLife {
		t.Fatalf("a locked Funeral Room drained: me %d -> %d, opp %d -> %d", meLife, me.Life, oppLife, opp.Life)
	}

	roomsBUnlock(t, g, me, c, game.DoorLeft)
	roomsBSettle(t, g, me)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wolf) })
	roomsBSettle(t, g, me)
	if me.Life != meLife+1 {
		t.Errorf("you gained %d, want 1", me.Life-meLife)
	}
	if opp.Life != oppLife-1 {
		t.Errorf("opponent life %d -> %d, want -1", oppLife, opp.Life)
	}
}
