package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b39TokensNamed lists the battlefield tokens with a name.
func B39TokensNamed(g *game.Game, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.IsToken() {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}
