package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- shared test helpers ------------------------------------------------------

// trekB33Tokens is the tokens named `name` that p controls.
func TrekB33Tokens(g *game.Game, p uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == name && c.Controller == p {
			out = append(out, c)
		}
	}
	return out
}
