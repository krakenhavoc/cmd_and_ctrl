package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc16TokensNamed lists the tokens on the battlefield with the given name.
func Msc16TokensNamed(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}
