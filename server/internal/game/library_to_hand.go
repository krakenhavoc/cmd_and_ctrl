package game

import "github.com/google/uuid"

// library_to_hand.go — "put it into your hand", for cards NAMED out of
// the top of a library (#952).
//
// "Look at / reveal the top N, take a subset, do something with the
// rest" is one of the most common templates in the format — every
// Impulse, every Ringleader, every Commune — and until this file the
// engine had no door for the middle step. Two other doors were being
// borrowed instead, and each is the wrong one for a reason:
//
//   - A library SEARCH (SearchLibraryThenForEffect) emits
//     EventSearchLibrary, shuffles, and prompts over the WHOLE library.
//     A player taking a card off a five-card look would be offered
//     every card in their deck, and Aven Mindcensor would see a search
//     that never happened.
//   - BounceCardsToHandForEffect works, and it works by luck:
//     zoneRoute.CardID finds a card's zone by scan, so "return it to
//     its owner's hand" happens to move a card that was never on the
//     battlefield. Goblin Ringleader shipped on that accident (#952),
//     and an accident is not something the next twenty cards should be
//     built on. "Return to hand" is also observably not this move: a
//     card put into a hand off the top of a library was not returned,
//     and nothing watching a bounce should see it.
//
// So this is the named move, on the same route machinery as every
// other zone change: one simultaneous exit, the CR 614 window open, a
// commander offered CR 903.9, and a continuation that is told what
// actually ARRIVED (CR 400.7) rather than what was asked to move.
//
// It is deliberately NOT a reveal and NOT a prompt. Who may look at
// the cards is the caller's first decision — LookAtTopOfLibraryForEffect
// for "look at" (the looker alone becomes a knower) or
// RevealTopOfLibraryForEffect for "reveal" (the whole table does) —
// and which of them may be taken is the card's text, asked through a
// choose_cards prompt. This file is only the move at the end of that
// sentence. See effects.TakeFromLibraryToHand for the card side.

// takeToHandRoute is the route template for a library-to-hand move.
//
// A sibling of bounceRoute rather than bounceRoute itself, and the
// difference is the Actor: the events carry the player who is taking
// the card, which is the player whose library was read and not
// necessarily the card's owner (a Bribery-shaped effect reads somebody
// else's). bounceRoute has no Actor because a mass bounce has no
// single responsible player.
//
// Destination owner is left nil, so routeDestinationLocked resolves
// "its owner's hand" per card — the only hand a card may be put into.
func takeToHandRoute(player uuid.UUID) zoneRoute {
	return zoneRoute{Dst: ZoneHand, Actor: player}
}

// TakeFromLibraryToHandThenForEffect puts every card of `ids` that is
// still in a LIBRARY into its owner's hand as one simultaneous exit,
// and hands `then` the ones that actually arrived in a hand.
//
// `player` is who is taking them — the player whose library was looked
// at or revealed from. It is stamped on the events and is not read as
// the destination: a card always goes to its OWNER's hand.
//
// Cards that are not in a library are skipped rather than moved. That
// is the difference between this and a bounce, and it is the point of
// the door: the ids come from a look or a reveal taken moments ago,
// and one of them having moved on since (a split second of another
// player's replacement effect, a commander that went elsewhere) means
// it is no longer a card being taken off the top of a library.
//
// `then` is told what LANDED, the CR 400.7 reading every other route
// continuation uses: a card whose owner has left the table has no hand
// to be put into and is not in the list, and neither is one a
// replacement sent somewhere else. It runs exactly once, after the
// last leg settles — including after a leg that paused on a CR 903.9
// prompt, which is why "put the rest on the bottom" belongs inside it
// and not on the next line.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) TakeFromLibraryToHandThenForEffect(player uuid.UUID, ids []uuid.UUID, then func(g *Game, taken []uuid.UUID) error) error {
	return g.routeAllThenLocked(takeToHandRoute(player), g.cardsInALibraryLocked(ids), then)
}

// TakeFromLibraryToHandForEffect is the fire-and-forget form, counting
// what landed. A card that reads the number — or has anything at all to
// do afterwards — uses the Then form above.
//
// Caller must hold g.mu in write mode.
func (g *Game) TakeFromLibraryToHandForEffect(player uuid.UUID, ids []uuid.UUID) int {
	return g.routeAllLocked(takeToHandRoute(player), g.cardsInALibraryLocked(ids))
}

// cardsInALibraryLocked filters `ids` to the cards currently in a zone
// of kind ZoneLibrary, in the order given. Caller must hold g.mu.
func (g *Game) cardsInALibraryLocked(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == ZoneLibrary {
			out = append(out, id)
		}
	}
	return out
}
