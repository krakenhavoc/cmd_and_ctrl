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
	offerFreeHandCastFrom(ctx.Game, controller, ctx.Source(), name, clause, pred)
}

// offerFreeHandCastFrom is offerFreeHandCast for a continuation that has
// the game and the source but no Context: a "draw, then ..." body
// (Cecily, Haunted Mage, #2110).
func offerFreeHandCastFrom(g *game.Game, controller, source uuid.UUID, name, clause string, pred CardPredicate) {
	candidates := freeCastHandCandidates(g, controller, pred)
	if len(candidates) == 0 {
		return
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
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

// freeCastHandCandidates is the hand cards a free cast may name: every
// nonland card (a land is never a spell) that passes `pred`, or all of
// them for a nil predicate.
func freeCastHandCandidates(g *game.Game, controller uuid.UUID, pred CardPredicate) []uuid.UUID {
	var candidates []uuid.UUID
	for _, id := range allHandCardIDs(g, controller) {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		if pred != nil && !pred(g, controller, c) {
			continue
		}
		candidates = append(candidates, id)
	}
	return candidates
}

// grantFreeHandCastsFrom is "you may cast any number of spells from your
// hand without paying their mana costs" (Aetherflux Conduit, #2110):
// the same pass-closed, any-timing free cast, over every candidate at
// once instead of one picked card. There is no prompt, because there is
// no single choice to ask for. Each card carries its own grant, so
// casting one leaves the rest open (a grant that named all of them would
// be spent by the first cast), and the holder's next priority pass ends
// every one of them together: the cards stay in hand.
//
// Returns how many cards were granted. Cards drawn by the same
// resolution are included, so this runs from the draw's continuation.
func grantFreeHandCastsFrom(g *game.Game, controller, source uuid.UUID, name string, pred CardPredicate) int {
	n := 0
	for _, id := range freeCastHandCandidates(g, controller, pred) {
		if g.FreeCastFromHandForEffect(controller, source, id, name) {
			n++
		}
	}
	return n
}
