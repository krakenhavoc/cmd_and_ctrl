package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82GapB117TwiceXCounters is "This creature enters with twice X
// <kind> counters on it" (Banquet Guests): XCounters' X, doubled. The
// count is arithmetic over the announcement alone (CR 614.1c).
func s82GapB117TwiceXCounters(kind string) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind:  kind,
		Count: func(cast game.CastCounts) int { return 2 * cast.X },
	}
}
