package protocol

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// impulse_exile_alt_cost_view_test.go — #2767. A card exiled under a
// plain "you may play it this turn" permission (Breeches) is castable
// for any alternative cost it prints that is not bound to another zone.
// The report was a Cyclonic Rift whose overload never reached the
// picker: the view's offer list is the engine's
// (game.CastOffersForLocked), so the permission has to list the card's
// overload and withhold its flashback.

func TestImpulseExiledCardCarriesItsUnboundAlternativeCosts(t *testing.T) {
	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-view-exiled-overloadable"

	prevAlts := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id != oracle {
			return nil
		}
		return []game.AlternativeCost{
			{Key: "overload", Label: "Overload {6}{U}", ManaCost: "{6}{U}", ClearsTargets: true},
			{Key: "flashback", Label: "Flashback {2}{R}", ManaCost: "{2}{R}", FromZone: game.ZoneGraveyard},
		}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prevAlts })

	rift := game.NewCard("Test Rift", me.ID)
	rift.TypeLine = "Instant"
	rift.ManaCost = "{1}{U}"
	rift.OracleID = oracle
	id := exileWithGrant(t, g, rift, game.CastPermission{Player: me.ID})

	mine := cardInZone(ViewOfGameFor(g, me.ID.String()).Exile, id)
	if mine == nil {
		t.Fatal("the exiled card is in the holder's view")
	}
	if len(mine.AlternativeCosts) != 1 || mine.AlternativeCosts[0].Key != "overload" {
		t.Fatalf("alternative_costs = %+v, want overload only (flashback is graveyard-bound)", mine.AlternativeCosts)
	}
	if mine.AlternativeCostRequired {
		t.Error("the permission charges the printed cost, so the printed cost is still on the menu")
	}
	var printed, overload bool
	for _, p := range mine.CastPrices {
		switch p.AlternativeCost {
		case "":
			printed = p.Cost == "{1}{U}"
		case "overload":
			overload = p.Cost == "{6}{U}"
		}
	}
	if !printed || !overload {
		t.Errorf("cast_prices = %+v, want the printed {1}{U} and overload {6}{U}", mine.CastPrices)
	}

	// The bystander is handed none of the holder's answer.
	theirs := cardInZone(ViewOfGameFor(g, opp.ID.String()).Exile, id)
	if theirs == nil || len(theirs.AlternativeCosts) != 0 {
		t.Errorf("a bystander was handed the holder's offers: %+v", theirs)
	}
}

// A permission that names a price of its own ("without paying its mana
// cost") takes no second alternative cost (CR 118.9a).
func TestExiledCardUnderAPricedPermissionListsNoOtherAlternativeCost(t *testing.T) {
	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me := g.Seats[0]
	const oracle = "test-view-exiled-overloadable-free"

	prevAlts := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id != oracle {
			return nil
		}
		return []game.AlternativeCost{
			{Key: "overload", Label: "Overload {6}{U}", ManaCost: "{6}{U}", ClearsTargets: true},
		}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prevAlts })

	free := game.NewCard("Test Rift", me.ID)
	free.TypeLine = "Instant"
	free.ManaCost = "{1}{U}"
	free.OracleID = oracle
	id := exileWithGrant(t, g, free, game.CastPermission{Player: me.ID, Cost: "{0}"})
	priced := cardInZone(ViewOfGameFor(g, me.ID.String()).Exile, id)
	if priced == nil {
		t.Fatal("the free-cast card is in the holder's view")
	}
	if len(priced.AlternativeCosts) != 0 {
		t.Errorf("a permission with its own price listed %+v", priced.AlternativeCosts)
	}
}
