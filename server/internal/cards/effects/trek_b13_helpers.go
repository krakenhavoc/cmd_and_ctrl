package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// trek_b13_helpers.go — Star Trek, batch 13 (#4973). Its own file, so the
// batch PRs landing in parallel do not meet in a shared one.

// CountersIfCastFrom is "If you do, it enters with a <kind> counter on
// it" after a card's own "You may cast this card from your <zone>"
// (Storied Veteran): a CR 614.1c clause read off the resolving spell's
// cast record (CastCounts.CastFrom), so it composes with every other
// entry clause. A cast from anywhere else, a copy of the spell (which was
// not cast, CR 707.10) and a permanent put onto the battlefield without
// being cast enter with none.
func CountersIfCastFrom(zone game.ZoneKind, kind string, n int) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind: kind,
		Count: func(cast game.CastCounts) int {
			if cast.CastFrom != zone {
				return 0
			}
			return n
		},
	}
}
