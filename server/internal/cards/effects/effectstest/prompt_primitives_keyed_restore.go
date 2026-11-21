package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bolsterProbeTail is a test tail: a shield counter on the creature it is
// told, so a test sees the tail ran and on what.
var BolsterProbeTail = game.RegisterTail("test/bolster-probe", func(g *game.Game, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := g.AddCounterForEffect(id, game.CounterShield, 1); err != nil {
			return err
		}
	}
	return nil
})
