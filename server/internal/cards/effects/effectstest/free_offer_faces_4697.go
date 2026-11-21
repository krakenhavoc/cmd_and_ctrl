package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushTop puts a built card on top of the Envoy controller's library.
func (tb *envoyTable) PushTop(c game.Card) uuid.UUID {
	c.Owner, c.Controller = tb.Me.ID, tb.Me.ID
	tb.G.WithWriteLock(func() { tb.Me.Library.PushTop(c) })
	return c.InstanceID
}
