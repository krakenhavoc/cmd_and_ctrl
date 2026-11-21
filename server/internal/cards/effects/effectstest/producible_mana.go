package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// producible is CR 106.7's answer for one battlefield card, read the
// way the derivations read it.
func Producible(g *game.Game, id uuid.UUID) []string {
	var out []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out = g.ProducibleManaLocked(c)
				return
			}
		}
	})
	return out
}
