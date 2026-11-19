package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bnsB42OpponentDealtNoncombatDamage is the condition of "whenever an
// opponent is dealt noncombat damage" (Wildfire Elemental): a positive
// non-combat EventDealDamage whose target is a player other than the
// source's controller.
func bnsB42OpponentDealtNoncombatDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return !ev.Combat && ev.Amount > 0 && ev.Target != source.Controller &&
		g.PlayerByIDForEffect(ev.Target) != nil
}

// bnsB42TokensInBatch is how many tokens `controller` had enter the
// battlefield in the event batch `batch` (CR 603.2c: "one or more tokens"
// is one trigger, and "that many" counts the batch).
func bnsB42TokensInBatch(g *game.Game, controller uuid.UUID, batch uint64) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventETB || ev.Batch != batch || ev.CardID == uuid.Nil {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && c.IsToken() && c.Controller == controller {
			n++
		}
	}
	return n
}
