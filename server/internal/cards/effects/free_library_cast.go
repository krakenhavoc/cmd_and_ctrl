package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_library_cast.go — "look at the top N cards of your library. You
// may cast a spell from among them without paying its mana cost. Put
// the rest on the bottom of your library in a random order" (Aetherworks
// Marvel, #2110).
//
// It is free_hand_cast.go's pick over cards the player has only LOOKED
// at, and Sunbird's Invocation's finish for a reveal. The differences
// from Sunbird are the two the printed words make:
//
//   - "look at" (CR 701.20), not "reveal": only the looker is marked a
//     knower (LookAtTopOfLibraryForEffect), so the prompt, the pile and
//     the chosen card are on nobody else's wire. The bottoming clears
//     the knowers (PutOnBottomInRandomOrderForEffect), because the
//     looker knows the cards are somewhere in the bottom few, not where.
//   - no mana value cap.
//
// #4250: the cast is made during the resolution (CR 608.2g), through the
// resolution_cast prompt over the nonland cards looked at, where they sit
// in the library: ONE free, any-timing permission
// (game.FreeCastFromLibraryPermission) over all of them, spent by the
// first cast, so one spell. Once the prompt is answered, by the cast or
// the decline, the rest of the looked-at cards still in the library go to
// the bottom in a random order (lookedAtRestToBottom), which is what "put
// the rest" means when the answer was "none". Only the looker sees the
// candidates: the prompt's cards go through the same knower filter as
// every prompt's.
//
// The keyed pick below (#4246) stays registered, so a table saved while
// waiting on one restores and answers; it is also the offer outside an
// open resolution, which no catalog card reaches. Its chosen card keeps
// its place and gets the pass-closed grant there.

// freeLibraryCastOffer is one card's look-and-cast offer: the name that
// heads the question and labels the grant, and the registered
// continuation the pick resumes through (#4246). The looked-at cards ride
// on the prompt as its Carry and the looker is its Chooser, so a table
// waiting on the pick is a restore point. The name is bound to the key
// for freeHandCastOffer's reason.
type freeLibraryCastOffer struct {
	name string
	then game.ChooseCardsThen
}

// registerFreeLibraryCastOffer registers a card's look-and-cast
// continuation under `key`, an on-disk identity (never renamed, never
// reused). Call it once, from a package-level var. The continuation
// captures only `name`, a constant of the card's text. A pick that ended
// unanswered (the looker left the game) does nothing: their library left
// with them.
func registerFreeLibraryCastOffer(key, name string) freeLibraryCastOffer {
	return freeLibraryCastOffer{
		name: name,
		then: game.RegisterChooseCardsThen(key, func(g *game.Game, r game.CardsChosen) error {
			if !r.Answered {
				return nil
			}
			var chosen uuid.UUID
			if len(r.Picked) > 0 {
				chosen = r.Picked[0]
			}
			return finishFreeCastFromLooked(g, r.Chooser, r.Source, name, r.Carry, chosen)
		}),
	}
}

// lookAtTopAndOfferFreeCast looks at the top n cards of `controller`'s
// library and asks which nonland card of them, if any, to cast free. The
// rest go to the bottom in a random order. Nothing is queued when no
// card of them can be cast, but the rest still go to the bottom.
// `allowed`, when not nil, narrows the cards that can be picked (Cosmic
// Cube's mana value cap); the rest still go to the bottom.
func lookAtTopAndOfferFreeCast(g *game.Game, controller, source uuid.UUID, offer freeLibraryCastOffer, n int, allowed func(game.Card) bool) error {
	name := offer.name
	looked := g.LookAtTopOfLibraryForEffect(controller, n)
	if len(looked) == 0 {
		return nil
	}
	var eligible []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() && (allowed == nil || allowed(c)) {
			eligible = append(eligible, id)
		}
	}
	if len(eligible) == 0 {
		return finishFreeCastFromLooked(g, controller, source, name, looked, uuid.Nil)
	}
	perm := g.FreeCastFromLibraryPermission(controller, source, name)
	// Nothing moves on the decline: lookedAtRestToBottom bottoms every
	// looked-at card left, the declined ones with the rest, in one order.
	perm.LapseOnPass = game.LapseStaysInHand
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       name,
		Cards:        eligible,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put them on the bottom",
		Then:         lookedAtRestToBottom,
		Carry:        looked,
	}) {
		return nil
	}
	// Keyed (#4246): the looked-at cards ride on the prompt as its Carry,
	// so a table waiting on the pick is a restore point.
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:     controller,
		Source:      source,
		Question:    name + " — cast a spell from among the top " + strconv.Itoa(len(looked)) + " cards of your library without paying its mana cost?",
		Cards:       eligible,
		Min:         0,
		Max:         1,
		Destination: game.ChooseCastFreeFromLibrary,
		ThenKey:     offer.then,
		Carry:       looked,
	})
	return nil
}

// finishFreeCastFromLooked is what happens once the pick is known: the
// rest of the looked-at cards still in a library go to the bottom in a
// random order, and the chosen card, if any and still in a library, keeps
// its place and gets its free cast there. A keyed pick has no live-zone
// re-check of its own (choose_cards_keyed.go), so the library check here
// is it.
func finishFreeCastFromLooked(g *game.Game, controller, source uuid.UUID, name string, looked []uuid.UUID, chosen uuid.UUID) error {
	var rest []uuid.UUID
	stillThere := false
	for _, id := range cardsStillInALibrary(g, looked) {
		if id == chosen {
			stillThere = true
			continue
		}
		rest = append(rest, id)
	}
	if err := g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneLibrary, rest); err != nil {
		return err
	}
	if stillThere {
		g.FreeCastFromLibraryForEffect(controller, source, chosen, name)
	}
	return nil
}

// lookedAtRestToBottom is the rest of a look-and-cast resolution once
// its resolution_cast prompt is answered (#4250): every looked-at card
// (the prompt's Carry) still in a library goes to the bottom in a random
// order. The card cast is on the stack, so it is not among them. An
// on-disk key, never renamed.
var lookedAtRestToBottom = game.RegisterChooseCardsThen("resolution-cast/looked-at-rest-to-bottom",
	func(g *game.Game, r game.CardsChosen) error {
		return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, cardsStillInALibrary(g, r.Carry))
	})
