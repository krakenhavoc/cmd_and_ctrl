package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m3cb1Push seeds a permanent, catalog row or not, untapped and able to
// act.
func M3cb1Push(g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	return PushBattlefieldCardWithTimestamp(g, c)
}
