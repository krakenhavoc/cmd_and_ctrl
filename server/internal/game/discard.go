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
// the loops together first meant the window had one place to go.
// Madness (#657) and a replaceable discard with a cause (#650) are
// the next two, and they land here too.
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
// A card that is no longer in the hand is skipped rather than
// erroring: the answer paths re-check the live zone before they
// dequeue, so an answer that got here was legal when it arrived, and
// a half-applied discard that abandoned its continuation would be the
// worse failure.
//
// Caller must hold g.mu.
func (g *Game) discardCardsLocked(playerID uuid.UUID, cards []uuid.UUID, opts discardOptions) error {
	p := g.playerByIDLocked(playerID)
	for _, id := range cards {
		if p == nil || !p.Hand.Contains(id) {
			continue
		}
		if _, err := MoveCard(p.Hand, p.Graveyard, id); err != nil {
			return err
		}
		g.markCardKnownInZoneLocked(p.Graveyard, id)
		g.EmitEvent(Event{
			Kind:    EventDiscardCard,
			Actor:   playerID,
			Source:  opts.source,
			CardID:  id,
			OldZone: ZoneHand,
			NewZone: ZoneGraveyard,
		})
	}
	if opts.then == nil {
		return nil
	}
	return opts.then(g)
}
