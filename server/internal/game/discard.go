package game

import "github.com/google/uuid"

// discard.go is the engine's ONE discard path.
//
// CR 701.8a: "To discard a card, move it from its owner's hand to
// that player's graveyard." Four places in the game package said that
// in Go, in four near-identical loops — move the card, mark it known
// in its new zone, emit EventDiscardCard:
//
//   - the hand-size cleanup discard (CR 514.1, DiscardSelection);
//   - the effect-discard continuation behind QueueDiscardChoiceForEffect
//     (#651/#797) — Mind Rot, looting, every "discard a card" a card
//     asks its controller to choose;
//   - the revealed-hand discard leg of ResolvePendingChoice
//     (Thoughtseize, where somebody ELSE picks);
//   - the random discard (CR 701.8b, DiscardRandomForEffect);
//
// plus the discard component of an additional cost (CR 601.2h,
// payAdditionalCostLocked).
//
// #799: four copies of three lines is four places for the next
// discard-shaped change to land, and #853 was exactly that change —
// no discard opened the CR 614 replacement window, so a discarded
// commander was never offered the command zone (CR 903.9). Folding
// the loops together first meant the window had one place to go, and
// #853 is the four lines of zoneRoute below. Madness (#657) and a
// replaceable discard that knows its cause (#650) are the next two,
// and they land here too.
//
// What the callers keep is what is genuinely theirs: WHICH cards
// (picked, randomly chosen, or named by the client) and what happens
// afterwards (dequeue the prompt, clear the hand-size debt, run the
// rest of the card). What they hand over is the discard itself.

// discardCause names why a card is being discarded. It is the one
// thing about a discard the rules treat differently, and today it
// decides exactly one thing: a COST may not pause (CR 601.2h).
//
// It is deliberately a closed enum rather than a bool: #650's
// replaceable discard needs the same three-way distinction on the
// event ("if you would discard a card" cares what caused it, and
// Library of Leng only replaces a discard that is an effect's
// instruction), so naming them now is naming them once.
type discardCause int

const (
	// discardCauseEffect is a discard an effect instructed — Mind Rot,
	// looting, a random discard, a revealed-hand pick. CR 701.8a, part
	// of a resolving spell or ability (CR 608.2c).
	discardCauseEffect discardCause = iota

	// discardCauseCleanup is the CR 514.1 hand-size discard: a turn-
	// based action in the cleanup step, nobody's effect.
	discardCauseCleanup

	// discardCauseCost is a discard paid as a cost — the additional
	// cost of casting a spell (CR 601.2h) or of activating an ability
	// (CR 602.2b). Costs are paid as one indivisible step, so a cost
	// discard MUST NOT pause on a player prompt; see
	// discardCardsLocked.
	discardCauseCost
)

// discardOptions carries what differs between the discard sites.
type discardOptions struct {
	// cause is why the discard is happening. See discardCause.
	cause discardCause

	// source is the card that caused the discard — the Mind Rot, the
	// spell whose additional cost this is — stamped on the emitted
	// EventDiscardCard so the log and a future "whenever a player
	// discards" payoff can name it. uuid.Nil for a discard with no
	// source card (the cleanup step's hand-size discard).
	source uuid.UUID

	// then is the rest of whatever asked for the discard, run once the
	// WHOLE batch has landed: the prompt's own "then draw two"
	// (#797), the cleanup step's hand-size bookkeeping. Optional.
	then func(g *Game) error
}

// discardCardsLocked discards `cards` from playerID's hand (CR 701.8a),
// emitting one EventDiscardCard per card — the event every "whenever
// you discard a card" trigger and every graveyard payoff in the
// catalog watches.
//
// THROUGH THE EXIT PRIMITIVE (#853). Each card takes
// routeCardToZoneLocked, the one move that opens the CR 614
// replacement window, exactly as destroy, exile, mill, counter and
// bounce have since #539/#851. Before this every discard was a raw
// MoveCard, so a discarded commander was the one way out of a hand
// that never offered CR 903.9's "put it into the command zone
// instead" — and a discard was the one exit a replacement effect could
// not see at all.
//
// SO A DISCARD CAN PAUSE, which is the whole reason this function is
// shaped the way it is. The card whose owner is being asked about the
// command zone has NOT moved yet, and the answer arrives an action
// later. The cards after it in the batch therefore cannot be discarded
// on the next line: they are the route's continuation, and each one
// starts the next from the point the previous one lands. The batch is
// a value carried forward — the `cards` slice, one card shorter each
// time — rather than a shared accumulator, which is what makes an undo
// across the prompt land where a clean run would. Same idiom as
// LoseLifeEachThenForEffect.
//
// Two consequences worth stating, because they are what a caller sees:
//
//   - opts.then runs when the WHOLE batch has landed, not when this
//     function returns. A Mind Rot on a commander returns nil with one
//     card discarded, one prompt open and "then draw a card" still
//     owed.
//   - The discards of one batch are sequential in the log even though
//     CR 701.8a makes them simultaneous. They already were; what is
//     new is that a paused one puts the rest on the far side of a
//     prompt. Mill made the opposite call (#529) because a paused mill
//     must not re-read the top of the library; a discard reads a list
//     that was fixed before the first card moved, so sequencing costs
//     it nothing and buys the honest "then".
//
// A COST MAY NOT PAUSE. CR 601.2h pays a spell's costs as one
// indivisible step and CR 602.2b says the same for an activated
// ability, so discardCauseCost sets zoneRoute.MustSettleNow: the
// window still runs — a discard replacement would still see it — but
// it settles without asking, and CR 903.9 being a "may" means a
// commander pitched to a cost goes to the graveyard. The argument, and
// the other half of the same cost line, is payLifeAsCostLocked.
//
// A card that is no longer in the hand is skipped rather than
// erroring: the answer paths re-check the live zone before they
// dequeue, so an answer that got here was legal when it arrived, and
// a half-applied discard that abandoned its continuation would be the
// worse failure.
//
// Caller must hold g.mu.
func (g *Game) discardCardsLocked(playerID uuid.UUID, cards []uuid.UUID, opts discardOptions) error {
	p := g.playerByIDLocked(playerID)
	for len(cards) > 0 {
		next, rest := cards[0], cards[1:]
		if p == nil || !p.Hand.Contains(next) {
			cards = rest
			continue
		}
		// Paused or not, the rest of the batch belongs to the route's
		// continuation from here: it runs inline the moment this card
		// lands, or from the CR 903.9 resume when its owner answers.
		_, err := g.routeCardToZoneLocked(zoneRoute{
			CardID: next,
			Dst:    ZoneGraveyard,
			// The discarding player's graveyard, named rather than
			// inferred: CR 701.8a moves the card to the hand owner's
			// graveyard, and every hand in this engine holds only its
			// owner's cards.
			DstOwner:      playerID,
			Actor:         playerID,
			Discard:       true,
			Source:        opts.source,
			MustSettleNow: opts.cause == discardCauseCost,
			then: func(g *Game) error {
				return g.discardCardsLocked(playerID, rest, opts)
			},
		})
		return err
	}
	if opts.then == nil {
		return nil
	}
	return opts.then(g)
}
