package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_hand_cast.go — "you may cast a spell from your hand without
// paying its mana cost" (#3270), the one offer Buster Sword, Kari Zev's
// Expertise, Rishkar's Expertise, Surtland Elementalist and Great
// Intelligence's Plan share. Each used to queue its own copy of the same
// choose-cards prompt and stamp a keyless one-shot grant, so the pick
// opened nothing: the player had to find the card and cast it by hand,
// at the card's own timing, any time that turn.
//
// The offer is now one prompt, tagged game.ChooseCastFree so the client
// opens the cast chain as soon as it is answered and the bot prices the
// pick as a cast, whose grant is the pass-closed any-timing free cast
// game.FreeCastFromHandPermission describes.

// offerFreeHandCast asks `controller` for a card in their hand that
// matches `pred` (nil means every nonland card) to cast free, or none.
// `name` heads the question and labels the grant. Nothing is queued when
// the hand holds no candidate.
func offerFreeHandCast(ctx *Context, controller uuid.UUID, name, clause string, pred CardPredicate) {
	var candidates []uuid.UUID
	for _, id := range allHandCardIDs(ctx.Game, controller) {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		if pred != nil && !pred(ctx.Game, controller, c) {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return
	}
	source := ctx.Source()
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:     controller,
		Source:      source,
		Question:    name + " — cast " + clause + " from your hand without paying its mana cost?",
		Cards:       candidates,
		Min:         0,
		Max:         1,
		Zone:        game.ZoneHand,
		Destination: game.ChooseCastFree,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			g.FreeCastFromHandForEffect(controller, source, picked[0], name)
			return nil
		},
	})
}
