package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82SncB23DiscardedInBatch lists the cards `you` discarded in the
// triggering event's batch that are in your graveyard now and have not left
// it since they arrived (CR 400.7): "whenever you discard one or more
// cards, … them". conspiracyTheoristDiscardedInBatch's walk without its
// nonland filter.
func S82SncB23DiscardedInBatch(g *game.Game, you uuid.UUID, trigger game.Event) []uuid.UUID {
	events := g.EventsThisTurn()
	var out []uuid.UUID
	for i, ev := range events {
		if ev.Kind != game.EventDiscardCard || ev.Batch != trigger.Batch || ev.Actor != you || ev.CardID == uuid.Nil {
			continue
		}
		if trigger.Batch == 0 && ev.Seq != trigger.Seq {
			continue
		}
		if !conspiracyTheoristInGraveyardOf(g, you, ev.CardID) {
			continue
		}
		left := false
		for _, later := range events[i+1:] {
			if later.CardID == ev.CardID && later.Kind == game.EventZoneMove && later.OldZone == game.ZoneGraveyard {
				left = true
				break
			}
		}
		if !left && !slices.Contains(out, ev.CardID) {
			out = append(out, ev.CardID)
		}
	}
	return out
}
