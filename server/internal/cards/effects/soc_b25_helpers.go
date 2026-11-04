package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b25_helpers.go — shared pieces of Secrets of Strixhaven (sos), batch 25
// (#4456).

// socB25GraveyardHasInstantAndSorcery is "there is an instant card and a
// sorcery card in your graveyard" (Flow State): the two are different cards,
// and a graveyard card has only its printed characteristics.
func socB25GraveyardHasInstantAndSorcery(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return false
	}
	instant, sorcery := false, false
	for _, c := range p.Graveyard.Cards {
		instant = instant || c.IsInstant()
		sorcery = sorcery || c.IsSorcery()
	}
	return instant && sorcery
}

// socB25YouPutACounterOnThisObjectThisTurn is "if you put a counter on this
// creature this turn" (Fractal Tender): a counter of any kind that `you`
// placed on the permanent `id` this turn, read back off the turn's event log.
// Counters put on an earlier object of the same card (before it left and came
// back) do not count (CR 400.7): the scan stops at the card's last departure
// from the battlefield. A removal, a cancel and counters another player put
// are not it (counterPlacedBy reads the placer the event names).
//
// Caller must hold g.mu.
func socB25YouPutACounterOnThisObjectThisTurn(g *game.Game, you, id uuid.UUID) bool {
	events := g.EventsThisTurn()
	for i := len(events) - 1; i >= 0; i-- {
		ev := events[i]
		if ev.Kind == game.EventLTB && ev.CardID == id {
			return false
		}
		if ev.Kind == game.EventCounterPlaced && ev.Target == id &&
			counterPlacementDelta(ev, g) > 0 && counterPlacedBy(ev, you, g) {
			return true
		}
	}
	return false
}
