package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b28_helpers.go — shared readers for Secrets of Strixhaven batch 28
// (#4456).

// socB28FirstLifeGainThisTurn is "whenever you gain life for the first time
// each turn": the event is a gain of the source controller's life (a positive
// EventChangeLife) and no earlier gain of theirs happened this turn, the
// turn's log read back to the real turn boundary (g.EventsThisTurn). The
// log is read whether or not the source was on the battlefield for the
// earlier gain: it is the player's first gain of the turn, not the source's
// first sight of one. The gain being asked about is itself excluded.
func socB28FirstLifeGainThisTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	if !YouGainedLife(ev, source, game.Characteristic{}, g) {
		return false
	}
	for _, prev := range g.EventsThisTurn() {
		if prev.Seq >= ev.Seq {
			break
		}
		if prev.Kind == game.EventChangeLife && prev.Target == source.Controller && prev.Amount > 0 {
			return false
		}
	}
	return true
}
