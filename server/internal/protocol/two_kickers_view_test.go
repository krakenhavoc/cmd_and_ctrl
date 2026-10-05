package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// two_kickers_view_test.go — #2153's view half. "Kicker {R} and/or
// {W}" (CR 702.33b) is two offers that share key "kicker": no new
// field, and the client's existing per-index toggles are the prompt.
// What this pins is that the projection keeps both — one toggle each,
// told apart by index and label — rather than collapsing on the key.

const viewTwoKickersOracle = "test-view-two-kickers"

func TestTwoKickerCostsAreTwoOffers(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	stubViewOptionalCosts(t, viewTwoKickersOracle, []game.AdditionalCost{
		{Optional: true, Key: game.KickerKey, ManaCost: "{R}", Label: "Kicker {R}"},
		{Optional: true, Key: game.KickerKey, ManaCost: "{W}", Label: "Kicker {W}"},
	})
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Two-Kicker Thing", TypeLine: "Creature — Wizard", OracleID: viewTwoKickersOracle,
			ManaCost: "{2}{G}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if len(c.OptionalCosts) != 2 {
		t.Fatalf("optional_costs = %+v, want one offer per kicker", c.OptionalCosts)
	}
	for i, want := range []string{"{R}", "{W}"} {
		o := c.OptionalCosts[i]
		if o.Index != i || o.Key != game.KickerKey || o.ManaCost != want || o.Label != "Kicker "+want || o.MaxTimes != 1 {
			t.Errorf("offer %d = %+v, want index %d, key kicker, %s, once", i, o, i, want)
		}
	}
}
