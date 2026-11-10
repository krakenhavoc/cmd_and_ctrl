package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b50_helpers.go — Standard set leftovers (#4452), batch 50.

// timesDescendedThisTurn is how many times `player` descended this turn
// (CR 700.11): once for each permanent card of theirs put into their
// graveyard from anywhere that was not already in a graveyard. descendedThisTurn's
// walk, counted instead of stopped at the first. Caller holds g.mu.
func timesDescendedThisTurn(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard || ev.CardID == uuid.Nil {
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if ok && c.Owner == player && isPermanentCard(c) {
			n++
		}
	}
	return n
}
