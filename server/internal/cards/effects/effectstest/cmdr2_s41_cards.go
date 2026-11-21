package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func Cmdr2S41InExile(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == id {
			return &g.Exile.Cards[i]
		}
	}
	return nil
}
