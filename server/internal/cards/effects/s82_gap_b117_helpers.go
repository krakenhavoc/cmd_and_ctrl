package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82GapB117TwiceXCounters is "This creature enters with twice X
// <kind> counters on it" (Banquet Guests): XCounters' X, doubled. The
// count is arithmetic over the announcement alone (CR 614.1c).
func S82GapB117TwiceXCounters(kind string) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind:  kind,
		Count: func(cast game.CastCounts) int { return 2 * cast.X },
	}
}

// APlayerAttacksYou is "whenever a player attacks you": an attack declaration
// by a player other than the source's controller, aimed at that controller
// (Archnemesis, Mirkwood Trapper).
func APlayerAttacksYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventAttack || ev.Actor == source.Controller {
		return false
	}
	return AttackedPlayer(g, ev) == source.Controller
}
