package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b256_helpers.go — shared vocabulary for S82 OLD batch 256
// (Commander 2016, #5537).

// S82OldB256CardPutIntoAnOpponentsGraveyard is "a card is put into an
// opponent's graveyard from anywhere" for the source's controller (Vulturous
// Zombie): a card (never a token, CR 108.2b) arriving in a graveyard owned by
// another player, by a move, a mill, a discard or a countered spell. A card
// that merely changes graveyard does not arrive.
func S82OldB256CardPutIntoAnOpponentsGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	for _, seat := range g.Seats {
		if seat == nil || seat.ID == source.Controller {
			continue
		}
		if (GraveyardArrivals{}).arrival(ev, g, seat.ID) {
			return true
		}
	}
	if ev.Kind != game.EventCounterSpell || ev.CardID == uuid.Nil || ev.Target != ev.CardID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.IsToken() || c.Owner == source.Controller {
		return false
	}
	z := g.FindCardZoneForEffect(ev.CardID)
	return z != nil && z.Kind == game.ZoneGraveyard
}

// S82OldB256InstantSorceryCardsInAllGraveyards counts the instant and sorcery
// cards in every player's graveyard (Academy Elite).
func S82OldB256InstantSorceryCardsInAllGraveyards(g *game.Game) int {
	n := 0
	for _, seat := range g.Seats {
		if seat != nil {
			n += CmdrB17InstantSorceryCardsInGraveyard(g, seat.ID)
		}
	}
	return n
}
