package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b687_helpers.go — shared pieces of the Mercadian Masques creatures of batch 687
// (S82 OLD, #5619).

// ThisBecameBlockedEvent is the When of "whenever this creature becomes blocked": the
// once-per-attacker EventBecomesBlocked that names the source (CR 509.1h).
func ThisBecameBlockedEvent(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventBecomesBlocked && ev.CardID == source.InstanceID
}

// BlockedDefenderDiscardsACard is "defending player discards a card" for a becomes-blocked
// trigger: the event's actor is the defending player (CR 509.3c), who chooses the card
// (CR 701.9a). A defender who has left the game discards nothing.
func BlockedDefenderDiscardsACard(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return DefendingPlayerDiscardsACard(g, item, item.Trigger.Event.Actor)
}
