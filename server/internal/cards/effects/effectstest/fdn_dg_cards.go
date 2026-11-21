package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdnDGTokens returns the P/T and colours of every token named `name`
// that `controller` controls.
func FdnDGTokens(g *game.Game, controller uuid.UUID, name string) (sizes [][2]int, colors [][]string) {
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == name && c.IsToken() {
				eff := c.Effective()
				sizes = append(sizes, [2]int{eff.Power, eff.Toughness})
				colors = append(colors, append([]string(nil), eff.Colors...))
			}
		}
	})
	return sizes, colors
}
