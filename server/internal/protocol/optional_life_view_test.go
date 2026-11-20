package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_life_view_test.go — the wire half of ADR 0100's 2026-11-20
// amendment (#5782): an optional cost that pays a fixed amount of life
// ("Buyback—Pay 4 life") names the life, and is marked short when the
// viewer's life total is below it (CR 119.4), so the client greys the
// offer the server would refuse.

const viewOptionalLifeOracle = "test-view-optional-life"

func TestOptionalLifeOfferIsShortBelowTheAmount(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	stubViewOptionalCosts(t, viewOptionalLifeOracle, []game.AdditionalCost{
		{Optional: true, Key: game.BuybackKey, PayLife: 4, Label: "Buyback—Pay 4 life"},
	})
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Life Thing", TypeLine: "Instant", OracleID: viewOptionalLifeOracle,
			ManaCost: "{2}{B}{B}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})

	for _, tc := range []struct {
		life  int
		short bool
	}{
		{life: 40},
		{life: 5},
		{life: 4}, // CR 119.4: exactly the amount may be paid
		{life: 3, short: true},
		{life: 0, short: true},
	} {
		g.WithWriteLock(func() { me.Life = tc.life })
		c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
		if len(c.OptionalCosts) != 1 {
			t.Fatalf("life %d: optional_costs = %+v, want the buyback offer", tc.life, c.OptionalCosts)
		}
		o := c.OptionalCosts[0]
		if o.PayLife != 4 || o.LifeShort != tc.short {
			t.Errorf("life %d: offer = pay_life %d, life_short %v; want 4, %v", tc.life, o.PayLife, o.LifeShort, tc.short)
		}
	}
}
