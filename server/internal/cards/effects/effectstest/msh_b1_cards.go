package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mshB1Tokens counts the tokens a player controls that have a subtype.
func MshB1Tokens(g *game.Game, controller uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

// mshB1Token returns the first token a player controls with a subtype.
func MshB1Token(t *testing.T, g *game.Game, controller uuid.UUID, subtype string) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.HasSubtype(subtype) {
			return GaLayered(t, g, c.InstanceID)
		}
	}
	t.Fatalf("no %s token", subtype)
	return game.Card{}
}
