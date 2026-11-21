package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func ZoneOfCard(g *game.Game, id uuid.UUID) game.ZoneKind {
	var k game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			k = z.Kind
		}
	})
	return k
}
