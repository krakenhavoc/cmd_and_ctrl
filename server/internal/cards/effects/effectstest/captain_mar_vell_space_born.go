package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// marVellTimingOpen asks the engine whether p may begin to cast a spell of
// the given type line out of their hand right now. It is the question the
// cast gate, the bot's enumerator and the client's castable_here all ask.
func MarVellTimingOpen(g *game.Game, p *game.Player, typeLine string) bool {
	var open bool
	g.WithWriteLock(func() {
		open = g.CastTimingOpenLocked(p.ID, game.Card{
			InstanceID: uuid.New(), Name: "Timing Probe", TypeLine: typeLine,
			Owner: p.ID, Controller: p.ID,
		}, game.ZoneHand, nil)
	})
	return open
}
