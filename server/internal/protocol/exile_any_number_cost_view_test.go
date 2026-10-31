package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_any_number_cost_view_test.go — #4455. An activated ability whose
// cost exiles any number of cards with a mana-symbol floor (Baron Helmut
// Zemo's boast) ships the any-number flag, the colour and the floor in place
// of a count, lists the graveyard's matching cards, and says why it can't be
// activated while the graveyard cannot reach the floor.

func TestActivatedAbilityViewCarriesAnAnyNumberExileCost(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	src := seatExileAbilitySource(g, me.ID, &game.ExileCost{
		AnyNumber: true, Label: "any number of black cards", From: game.ZoneGraveyard,
		Match:       func(c game.Card) bool { return c.HasColor("B") },
		SymbolTotal: &game.ManaSymbolTotal{Color: "B", AtLeast: 15},
	})
	black := func(name, cost string) {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	}
	black("Short A", "{B}{B}{B}{B}")
	black("Short B", "{B}{B}{B}{B}")
	seatGraveyardCardFor(me, "Colourless Rock", "Artifact")
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if !ab.ExileCostAnyNumber || ab.ExileCostN != 0 || ab.ExileCostZone != "graveyard" {
		t.Errorf("shape: any=%v n=%d zone=%q, want any number, no count, graveyard", ab.ExileCostAnyNumber, ab.ExileCostN, ab.ExileCostZone)
	}
	if ab.ExileCostSymbolColor != "B" || ab.ExileCostSymbolMin != 15 {
		t.Errorf("floor = %d %q, want 15 B", ab.ExileCostSymbolMin, ab.ExileCostSymbolColor)
	}
	if len(ab.ExileCostOptions) != 2 {
		t.Errorf("exile_cost_options = %v, want the two black cards", ab.ExileCostOptions)
	}
	if ab.CantActivate == "" {
		t.Error("eight black symbols in the graveyard: the ability should say it can't be activated")
	}

	black("Enough A", "{B}{B}{B}{B}")
	black("Enough B", "{B}{B}{B}{B}")
	g.BumpLayerVersionForTest()
	ab = vehicleView(t, g, src).ActivatedAbilities[0]
	if len(ab.ExileCostOptions) != 4 {
		t.Errorf("exile_cost_options = %v, want the four black cards", ab.ExileCostOptions)
	}
	if ab.CantActivate != "" {
		t.Errorf("sixteen black symbols: CantActivate = %q, want none", ab.CantActivate)
	}
}
