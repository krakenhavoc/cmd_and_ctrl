package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82BroB08ArtifactLeft is the shared "an artifact is put into a graveyard
// from the battlefield" read for Farid, Enterprising Salvager and Sardian
// Avenger: the permanent an EventLTB to a graveyard names, whether it was
// an artifact as it last existed (CR 603.10a), and who controlled it as it
// left. A token counts here; a card narrows it to nontoken itself.
func S82BroB08ArtifactLeft(ev game.Event, g *game.Game) (c game.Card, controller uuid.UUID, ok bool) {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return game.Card{}, uuid.Nil, false
	}
	c, found := g.LookupCardForEffect(ev.CardID)
	if !found || !LeftAsType(ev, c, "artifact") {
		return game.Card{}, uuid.Nil, false
	}
	return c, LeftUnderControlOf(ev, c), true
}

// s82BroB08Rummage is "discard a card, then draw a card" (Charming
// Scoundrel's keyed rummage): the discard is chosen first and the draw
// follows it, drawing even from an empty hand.
func S82BroB08Rummage(g *game.Game, item *game.StackItem) error {
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player: item.Controller,
		Source: item.SourceCardID,
		N:      1,
	}, CharmingScoundrelDraw, nil)
}
