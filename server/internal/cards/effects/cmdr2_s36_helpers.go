package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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

// cmdr2S36YourPermanentTurnedFaceUpFromGraveyard is "whenever a permanent
// you control is turned face up" for a card watching from its owner's
// graveyard (Deathmist Raptor). A card in a graveyard has no controller
// (CR 108.4), so "you" is its owner.
func cmdr2S36YourPermanentTurnedFaceUpFromGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventTurnedFaceUp {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Owner
}

// cmdr2S36ReturnThisFaceUpOrFaceDown asks "face up or face down?" and
// returns the source card from its owner's graveyard that way. A card
// that left the graveyard meanwhile is a new object (CR 400.7) and
// nothing happens.
func cmdr2S36ReturnThisFaceUpOrFaceDown(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return PickOption{
		Question: "Return it to the battlefield face up or face down?",
		Options: []game.ChoiceOption{
			{Label: "Face up"},
			{Label: "Face down"},
		},
		ThenKey: cmdr2S36FaceUpOrFaceDownAnswer,
		Carry:   []uuid.UUID{item.SourceCardID},
	}.Apply(NewContext(g, item))
}

// cmdr2S36FaceUpOrFaceDownAnswer is the answer, keyed (#4253). An
// on-disk key: never renamed, never reused.
var cmdr2S36FaceUpOrFaceDownAnswer = OptionPickThen("option-pick/return-this-face-up-or-face-down", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	id := r.Carry[0]
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	switch r.Index {
	case 0:
		return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
	case 1:
		_, err := ctx.Game.ReturnFromGraveyardFaceDownForEffect(id, uuid.Nil, false, nil)
		if errors.Is(err, game.ErrCardNotFound) {
			return nil
		}
		return err
	}
	return nil
})
