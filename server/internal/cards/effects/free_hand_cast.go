package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_hand_cast.go — "you may cast a spell from your hand without
// paying its mana cost" (#3270), the one offer Buster Sword, Cecily,
// Haunted Mage, Kari Zev's Expertise, Rishkar's Expertise, Surtland
// Elementalist and Great Intelligence's Plan share. Each used to queue its
// own copy of the same choose-cards prompt and stamp a keyless one-shot
// grant, so the pick opened nothing: the player had to find the card and
// cast it by hand, at the card's own timing, any time that turn.
//
// The offer is now one prompt, tagged game.ChooseCastFree so the client
// opens the cast chain as soon as it is answered and the bot prices the
// pick as a cast, whose grant is the pass-closed any-timing free cast
// game.FreeCastFromHandPermission describes.
//
// The pick is KEYED (#4246, choose_cards_keyed.go): its continuation is a
// registered key, one per card, and what the answer needs rides on the
// prompt as plain data. A table waiting on the offer is a restore point.

// freeHandCastOffer is one card's free-cast offer: the name that heads the
// question and labels the grant, and the registered continuation the pick
// resumes through. The prompt carries the chooser (who casts), the source
// (the card that asked) and the candidates; the name is bound to the key.
//
// The name is not read back off the source card when the answer comes in,
// because the source may be gone by then: a copy of Rishkar's Expertise
// has ceased to exist (CR 707.10a), and so has a token Surtland
// Elementalist that died (CR 704.5d). One key per card keeps the label
// whatever became of the source.
type freeHandCastOffer struct {
	name string
	then game.ChooseCardsThen
}

// registerFreeHandCastOffer registers a card's free-cast continuation
// under `key`, an on-disk identity: never renamed, never reused, and
// ledgered in game/testdata/effect_keys.txt. Call it once, from a
// package-level var. The continuation captures only `name`, a constant of
// the card's printed text, never game state.
func registerFreeHandCastOffer(key, name string) freeHandCastOffer {
	return freeHandCastOffer{
		name: name,
		then: game.RegisterChooseCardsThen(key, func(g *game.Game, r game.CardsChosen) error {
			grantPickedFreeHandCast(g, r, name)
			return nil
		}),
	}
}

// grantPickedFreeHandCast is a free-cast offer's answer: the pass-closed,
// any-timing {0} cast over the card picked, if it is still in the
// chooser's hand. A keyed pick has no live-zone re-check of its own
// (choose_cards_keyed.go), so this is it: a card that left the hand
// between the question and the answer is granted nothing.
func grantPickedFreeHandCast(g *game.Game, r game.CardsChosen, name string) {
	if !r.Answered || len(r.Picked) == 0 {
		return
	}
	p := g.PlayerByIDForEffect(r.Chooser)
	if p == nil || p.Hand == nil || !p.Hand.Contains(r.Picked[0]) {
		return
	}
	g.FreeCastFromHandForEffect(r.Chooser, r.Source, r.Picked[0], name)
}

// offerFreeHandCast asks `controller` for a card in their hand that
// matches `pred` (nil means every nonland card) to cast free, or none.
// The offer's name heads the question and labels the grant. Nothing is
// queued when the hand holds no candidate.
func offerFreeHandCast(ctx *Context, controller uuid.UUID, offer freeHandCastOffer, clause string, pred CardPredicate) {
	offerFreeHandCastFrom(ctx.Game, controller, ctx.Source(), offer, clause, pred)
}

// offerFreeHandCastFrom is offerFreeHandCast with the game and the
// source handed in rather than read off a Context.
func offerFreeHandCastFrom(g *game.Game, controller, source uuid.UUID, offer freeHandCastOffer, clause string, pred CardPredicate) {
	offerFreeHandCastOver(g, controller, source, offer,
		offer.name+" — cast "+clause+" from your hand without paying its mana cost?",
		freeCastHandCandidates(g, controller, pred))
}

// offerFreeHandCastOver queues the keyed offer over `candidates`, cards in
// `controller`'s hand, under `question`. Nothing is queued for none.
func offerFreeHandCastOver(g *game.Game, controller, source uuid.UUID, offer freeHandCastOffer, question string, candidates []uuid.UUID) {
	if len(candidates) == 0 {
		return
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:     controller,
		Source:      source,
		Question:    question,
		Cards:       candidates,
		Min:         0,
		Max:         1,
		Destination: game.ChooseCastFree,
		ThenKey:     offer.then,
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
