package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// radOf is a player's rad counters.
func RadOf(p *game.Player) int { return p.Counters[game.CounterRad] }
