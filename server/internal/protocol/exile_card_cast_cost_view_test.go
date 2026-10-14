package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_card_cast_cost_view_test.go — #3536: the hand card's
// additional_cost carries the "exile six cards from your graveyard"
// picker, off the same walk the engine validates against. Five cards in the
// graveyard is a picker with five options for a count of six: the client
// greys the spell on exactly that comparison.

func TestAbhorrentOculusCarriesItsExileCardsPicker(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Abhorrent Oculus", TypeLine: "Creature — Eye", OracleID: "b1a0e129-27f0-4a09-b00d-497e18777306",
			ManaCost: "{2}{U}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if c.AdditionalCost == nil || c.AdditionalCost.ExileCostN != 6 {
		t.Fatalf("additional_cost = %+v, want an exile cost of six even with an empty graveyard", c.AdditionalCost)
	}
	if len(c.AdditionalCost.ExileCostOptions) != 0 {
		t.Errorf("an empty graveyard offers %v", c.AdditionalCost.ExileCostOptions)
	}

	var mine []string
	g.WithWriteLock(func() {
		for range 5 {
			f := game.NewCard("Fuel", me.ID)
			mine = append(mine, f.InstanceID.String())
			me.Graveyard.PushTop(f)
		}
		opp.Graveyard.PushTop(game.NewCard("Their Fuel", opp.ID))
	})
	c = handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	ac := c.AdditionalCost
	if ac.ExileCostN != 6 || len(ac.ExileCostOptions) != 5 {
		t.Fatalf("exile cost n=%d with %d options, want 6 with the five cards of mine", ac.ExileCostN, len(ac.ExileCostOptions))
	}
	got := map[string]bool{}
	for _, s := range ac.ExileCostOptions {
		got[s] = true
	}
	for _, s := range mine {
		if !got[s] {
			t.Errorf("option list %v is missing my card %s", ac.ExileCostOptions, s)
		}
	}
	if ac.ExileCostZone != "graveyard" || ac.ExileCostLabel != "six cards" {
		t.Errorf("zone %q label %q, want graveyard / six cards", ac.ExileCostZone, ac.ExileCostLabel)
	}
}
