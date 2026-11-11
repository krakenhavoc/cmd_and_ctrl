package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2_s36_helpers.go — shared pieces of the morph batch (batch 36).

// cmdr2S36PermanentYouControlTurnedFaceUp reads a turned-face-up event
// for a permanent you control (the source's controller), the permanent
// named by the event, read after it turned up (CR 708.8). `self` admits
// the source itself.
func cmdr2S36PermanentYouControlTurnedFaceUp(ev game.Event, source *game.Card, g *game.Game, self bool) (game.Card, bool) {
	if ev.Kind != game.EventTurnedFaceUp {
		return game.Card{}, false
	}
	if !self && ev.CardID == source.InstanceID {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != source.Controller {
		return game.Card{}, false
	}
	return c, true
}

// cmdr2S36AnotherCreatureYouControlTurnedFaceUp is "whenever another
// permanent you control is turned face up, if it's a creature".
func cmdr2S36AnotherCreatureYouControlTurnedFaceUp(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := cmdr2S36PermanentYouControlTurnedFaceUp(ev, source, g, false)
	return ok && c.IsCreature()
}
