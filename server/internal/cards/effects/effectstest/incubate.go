package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// incubatorsOf returns every Incubator token on the battlefield that
// `controller` controls, whichever face is up.
func IncubatorsOf(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.TokenKey == game.TokenKey("incubator") && c.Controller == controller {
				out = append(out, c)
			}
		}
	})
	return out
}
