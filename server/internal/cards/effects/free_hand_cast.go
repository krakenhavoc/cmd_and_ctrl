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
// #4250: the cast is made DURING the resolution that offers it (CR
// 608.2g), through the resolution_cast prompt
// (game.OfferResolutionCastForEffect). The offer stamps ONE pass-closed,
// any-timing free permission (game.FreeCastFromHandPermission) over every
// candidate and asks the prompt over them: the first cast spends the
// permission, which is "cast a spell" (one), and the decline lapses it,
// with the cards left in hand. A hand card also has its ordinary cast, so
// the prompt's cast is held to the permission's claim (the engine names
// game.AltCostKeyGrantedCost on it). The prompt is plain data: a table
// waiting on it is a restore point.
//
// The keyed pick (#4246, choose_cards_keyed.go) that came before it stays
// registered under each card's key, so a table saved while waiting on one
// still restores and answers. It is also the offer outside an open
// resolution, which no catalog card reaches.

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

// offerFreeHandCastOver offers the free cast over `candidates`, cards in
// `controller`'s hand: the resolution_cast prompt inside an open
// resolution (#4250), else the keyed pick under `question`. Nothing is
// queued for none.
func offerFreeHandCastOver(g *game.Game, controller, source uuid.UUID, offer freeHandCastOffer, question string, candidates []uuid.UUID) {
	if len(candidates) == 0 {
		return
	}
	perm := g.FreeCastFromHandPermission(controller, source, offer.name)
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       offer.name,
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordHandCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Keep it in your hand",
	}) {
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
// once instead of one picked card. Each card carries its own grant, so
// casting one leaves the rest open (a grant that named all of them would
// be spent by the first cast).
//
// #4250: inside the resolution the casts are the resolution_cast prompt
// over every candidate, each cast its own answer until the player
// declines or none is left (CR 608.2g); the decline lapses the rest, with
// the cards left in hand. `then` is what the resolution does after the
// casts (The Great Synthesis's flicker), run once the prompt is answered;
// the zero value is nothing. Outside a resolution the grants alone are
// the window, closing on the holder's next priority pass.
//
// Returns how many cards were granted. Cards drawn by the same
// resolution are included, so this runs from the draw's continuation.
func grantFreeHandCastsFrom(g *game.Game, controller, source uuid.UUID, name string, pred CardPredicate) int {
	return offerFreeHandCastsOf(g, controller, source, name, freeCastHandCandidates(g, controller, pred),
		func(id uuid.UUID) bool { return g.FreeCastFromHandForEffect(controller, source, id, name) },
		game.ChooseCardsThen{})
}

// offerFreeHandCastsOf grants each of `candidates` with `grant` and asks
// the resolution_cast prompt over the ones granted (grantFreeHandCastsFrom).
func offerFreeHandCastsOf(g *game.Game, controller, source uuid.UUID, name string, candidates []uuid.UUID,
	grant func(uuid.UUID) bool, then game.ChooseCardsThen) int {
	var granted []uuid.UUID
	for _, id := range candidates {
		if grant(id) {
			granted = append(granted, id)
		}
	}
	if !g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       name,
		Cards:        granted,
		Keyword:      game.MayCastKeywordHandCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Cast no more",
		Then:         then,
		Carry:        granted,
	}) && then.Key() != "" {
		// No resolution to hold open: the rest of it runs now.
		g.RunResolutionCastThenForEffect(then, controller, source, nil)
	}
	return len(granted)
}
