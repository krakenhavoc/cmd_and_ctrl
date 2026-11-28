package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b845_helpers.go — shared pieces of Ice Age batch 845 (#5636, S82).

// OldB845BlockedSinceYourLastUpkeep is Wiitigo's "if it has blocked or been
// blocked since your last upkeep": read off the append-only event log, newest
// first. The walk passes the upkeep that is happening now (the first
// EventBeginUpkeep of `controller` it meets), stops at the one before it, and
// stops at the permanent's own entry, so an earlier object that shared the
// instance ID is not counted (CR 400.7). A block event of the permanent
// (EventBlock: it blocks; EventBecomesBlocked: it became blocked) in between
// answers yes. Caller must hold g.mu.
func OldB845BlockedSinceYourLastUpkeep(g *game.Game, id, controller uuid.UUID) bool {
	passedCurrent := false
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		switch ev.Kind {
		case game.EventBeginUpkeep:
			if ev.Actor != controller {
				continue
			}
			if passedCurrent {
				return false
			}
			passedCurrent = true
		case game.EventETB:
			if ev.CardID == id {
				return false
			}
		case game.EventBlock, game.EventBecomesBlocked:
			if ev.CardID == id {
				return true
			}
		}
	}
	return false
}
