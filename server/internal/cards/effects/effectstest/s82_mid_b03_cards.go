package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// midB03Flash casts the card in the active seat's graveyard for its
// flashback cost, strictly checking nothing about mana (the catalog game
// lets a non-strict cast through), and returns the card's ID.
func MidB03Flash(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := SeedGraveyardCard(t, g, name, typeLine, oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: targets,
	}); err != nil {
		t.Fatalf("flashback cast of %s: %v", name, err)
	}
	return id
}

func MidB03Exiled(t *testing.T, g *game.Game, id uuid.UUID, name string) {
	t.Helper()
	if !g.Exile.Contains(id) || g.Seats[g.Turn.ActiveSeat].Graveyard.Contains(id) {
		t.Errorf("the flashed-back %s should be exiled, not back in the graveyard", name)
	}
}

func MidB03Tokens(g *game.Game, controller uuid.UUID, subtype string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.HasSubtype(subtype) {
			out = append(out, c)
		}
	}
	return out
}

func MidB03CantBlock(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return AuraRestrictions(t, g, id)&game.CantBlock != 0
}

func CountTokensNamed2(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.Name == name {
			n++
		}
	}
	return n
}
