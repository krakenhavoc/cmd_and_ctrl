package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const nettlingNuisanceOracle = "9662c579-f414-4c2d-b133-0f5f21d5c28a"

func TestNettlingNuisanceIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(nettlingNuisanceOracle)
	if !ok || spec.Name != "Nettling Nuisance" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func nettlingPirates(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Pirate" && c.IsToken() && c.Controller == owner {
			out = append(out, c)
		}
	}
	return out
}

func TestNettlingNuisanceGivesTheDamagedPlayerAGoadedPirateOncePerPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	nuisance := b12Push(g, me.ID, "Nettling Nuisance", "Creature — Faerie Rogue", nettlingNuisanceOracle, 3, 1)
	faerie := b12Creature(g, me.ID, "Other Faerie", "Creature — Faerie", 2, 2)
	human := b12Creature(g, me.ID, "A Human", "Creature — Human", 2, 2)

	b784AttackEach(t, g,
		[2]uuid.UUID{nuisance, opp.ID},
		[2]uuid.UUID{faerie, opp.ID},
		[2]uuid.UUID{human, third.ID})

	// Two Faeries on one player are one trigger; a Human is none.
	if n := b784Triggers(g, nuisance); n != 1 {
		t.Fatalf("%d triggers, want 1", n)
	}
	passPriorityAroundTable(t, g)

	if n := len(nettlingPirates(g, third.ID)); n != 0 {
		t.Errorf("the Human's victim got %d Pirates, want 0", n)
	}
	if n := len(nettlingPirates(g, me.ID)); n != 0 {
		t.Errorf("the controller got %d Pirates, want 0", n)
	}
	pirates := nettlingPirates(g, opp.ID)
	if len(pirates) != 1 {
		t.Fatalf("the damaged player has %d Pirates, want 1", len(pirates))
	}
	p := pirates[0]
	if p.Power != 4 || p.Toughness != 2 || !p.HasColor("R") {
		t.Errorf("Pirate = %d/%d %v, want a 4/2 red one", p.Power, p.Toughness, p.Colors)
	}
	if !p.IsGoadedBy(me.ID) {
		t.Error("the Pirate is not goaded by Nettling Nuisance's controller")
	}
	if p.Effective().Restrictions&game.CantBlock == 0 {
		t.Error("the Pirate can block")
	}

	// "For the rest of the game": it is still goaded two of its controller's
	// turns later.
	endTurn(t, g)
	endTurn(t, g)
	endTurn(t, g)
	endTurn(t, g)
	got, ok := g.LookupCardForEffect(p.InstanceID)
	if !ok || !got.IsGoadedBy(me.ID) {
		t.Error("the goad ended")
	}
}
