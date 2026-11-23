package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tokensNamed is every token on the battlefield with that name.
func TokensNamed(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}
