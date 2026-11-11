package protocol

import (
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Magmatic Insight
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_land_view_test.go — #5457: "discard a land card" ships the
// cards that may pay it as `discard_options`, off the engine's own walk;
// a plain "discard a card" ships none.

const viewMagmaticInsightOracle = "6a9031b6-3f4d-4a36-90e1-ac77b8518cee"

func TestDiscardLandCostListsTheLandCards(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	land, insight := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		known := map[uuid.UUID]bool{me.ID: true}
		me.Hand.PushTop(game.Card{InstanceID: land, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID, KnownBy: known})
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", Owner: me.ID, Controller: me.ID, KnownBy: known})
		me.Hand.PushTop(game.Card{
			InstanceID: insight, Name: "Magmatic Insight", TypeLine: "Sorcery", OracleID: viewMagmaticInsightOracle,
			ManaCost: "{R}", Owner: me.ID, Controller: me.ID, KnownBy: known,
		})
	})
	ac := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), insight.String()).AdditionalCost
	if ac == nil || ac.DiscardCards != 1 || ac.Label != "Discard a land card" || ac.DiscardOptions == nil {
		t.Fatalf("additional_cost = %+v, want one discard with options", ac)
	}
	if o := ac.DiscardOptions; o.Min != 1 || o.Max != 1 || len(o.Cards) != 1 || o.Cards[0] != land.String() {
		t.Errorf("discard_options = %+v, want the one land card", o)
	}
}
