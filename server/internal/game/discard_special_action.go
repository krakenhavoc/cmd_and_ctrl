package game

import "github.com/google/uuid"

// discard_special_action.go — #6224, CR 116.2e: "One card (Circling
// Vultures) has the ability "You may discard Circling Vultures any time
// you could cast an instant." Doing so is a special action. A player can
// take such an action any time they have priority."
//
// A special action like foretell or plot (special_action.go): no stack,
// nothing to respond to, and the player keeps priority. It is declared on
// the card (effects.DiscardAnyTimeYouCouldCastAnInstant) and taken on the
// card in its owner's hand.
//
// The window is the printed one, "any time you could cast an instant":
// the player has priority (the dispatcher's requirePriorityHolder, and
// holdsPriorityLocked here for the steps nobody holds it in) and no
// split-second spell is on the stack, under which nobody could cast an
// instant (CR 702.61a). That is narrower than CR 116.2e's summary ("any
// time they have priority") only under split second, where the card's
// own words govern.
//
// The discard is the ordinary one (CR 701.9a): discard triggers see it,
// madness replaces it, and its cause is its own, DiscardCauseSpecialAction
// — not an effect's, so "if an effect causes you to discard" (Library of
// Leng) does not reach it, and not a cost's, so it is not paid as one.

// SpecialActionDiscard is CR 116.2e: discard this card from your hand
// any time you could cast an instant (Circling Vultures).
const SpecialActionDiscard SpecialActionKind = "discard"

// DiscardCauseSpecialAction is a discard taken as a special action
// (CR 116.2e, Circling Vultures): the player's own action, neither an
// effect's instruction nor a cost.
const DiscardCauseSpecialAction DiscardCause = "special_action"

// discardSpecialActionLocked discards the card from its owner's hand.
// Caller must hold g.mu.
func (g *Game) discardSpecialActionLocked(p *Player, cardID uuid.UUID, _ SpecialAction) error {
	return g.discardCardsLocked(p.ID, []uuid.UUID{cardID}, discardOptions{cause: DiscardCauseSpecialAction, source: cardID})
}
