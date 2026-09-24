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
// #853 is the four lines of zoneRoute below. #650 was the next one and
// landed here too: the route now opens a RepEventDiscard carrying the
// cause, so Library of Leng, madness (#657) and the Obstinate Baloth
// family have a discard to key on rather than an anonymous move.
//
// What the callers keep is what is genuinely theirs: WHICH cards
// (picked, randomly chosen, or named by the client) and what happens
// afterwards (dequeue the prompt, clear the hand-size debt, run the
// rest of the card). What they hand over is the discard itself.

// DiscardCause names why a card is being discarded. It is the one
// thing about a discard the rules treat differently, and since #650 it
// decides two: a COST may not pause (CR 601.2h), and a replacement
// effect can read it off the event ("if an EFFECT causes you to
// discard a card" — Library of Leng).
//
// A closed enum rather than a bool, and a STRING one so it reads in
// the event log and on the wire without a translation table. ADR 0013
// §10a is the argument for these three and against the
// voluntary/involuntary framing #160 was written around: the rules
// have no such thing as a voluntary discard.
type DiscardCause string

const (
	// DiscardCauseEffect is a discard an effect instructed — Mind Rot,
	// looting, a random discard, a revealed-hand pick. CR 701.8a, part
	// of a resolving spell or ability (CR 608.2c). This is the only
	// cause Library of Leng replaces.
	DiscardCauseEffect DiscardCause = "effect"

	// DiscardCauseCleanup is the CR 514.1 hand-size discard: a turn-
	// based action in the cleanup step (CR 703.1), nobody's effect.
	DiscardCauseCleanup DiscardCause = "cleanup"

	// DiscardCauseCost is a discard paid as a cost — the additional
	// cost of casting a spell (CR 601.2h), the cost of activating an
	// ability (CR 602.2b), or the CR 118.12 "unless you discard"
	// branch of a resolving spell. Costs are paid as one indivisible
	// step, so a cost discard MUST NOT pause on a player prompt; see
	// discardCardsLocked. Costs are also not effects (Gatherer ruling,
	// 2004-10-04), which is the other half of why Library of Leng
	// leaves one alone.
	DiscardCauseCost DiscardCause = "cost"
)

// discardOptions carries what differs between the discard sites.
type discardOptions struct {
	// cause is why the discard is happening. See DiscardCause. The
	// zero value ("") is normalised to DiscardCauseEffect by the
	// route, which is the cause every caller that forgets to say means.
	cause DiscardCause

	// source is the card that caused the discard — the Mind Rot, the
	// spell whose additional cost this is — stamped on the emitted
	// EventDiscardCard so the log and a future "whenever a player
	// discards" payoff can name it. uuid.Nil for a discard with no
	// source card (the cleanup step's hand-size discard).
	source uuid.UUID

	// then is the rest of whatever asked for the discard, run once the
	// WHOLE batch has landed: the prompt's own "then draw two"
	// (#797), the cleanup step's hand-size bookkeeping, a prompted
	// discard run's leg settle (#1027). Optional.
	//
	// `landed` is the cards of this batch that were really discarded
	// — discardedThisWayLocked's answer, in batch order. #1027 gave it
	// the argument, because "you draw a card for each card discarded
	// this way" (Syphon Mind) cannot be read back off the hand on the
	// next line: the batch may have paused on a CR 903.9 prompt, a
	// madness card went to exile rather than to a graveyard, and the
	// CR 614 window may have cancelled a leg outright.
	then func(g *Game, landed []uuid.UUID) error

	// commanderAnswers are the CR 903.9 answers a COST discard's
	// commanders' owners gave before the payment (#1397,
	// cost_commander_choice.go), keyed by card. Only a cost discard
	// sets it: every other discard can pause, and asks at the move.
	commanderAnswers map[uuid.UUID]bool
}

// discardedThisWayLocked reports whether a settled discard of cardID
// counts as a DISCARD — the question "for each card discarded this
// way" (Syphon Mind's draw) is asking.
//
// CR 701.8a: "To discard a card, move it from its owner's hand to
// that player's graveyard." The discard is that MOVE OUT of the hand,
// and nothing replaces the discard itself — a replacement rewrites
// where the card goes. So the answer is "is it still in the hand":
//
//   - anywhere else. Discarded. A graveyard is the printed
//     destination; EXILE is madness (CR 702.35a, #657 — "that player
//     discards it, BUT exiles it instead", so the discard happened and
//     only the destination was rewritten) or a Rest in Peace; the top
//     of a library is Library of Leng; the command zone is CR 903.9
//     taking the offer. All of them replaced the destination of a
//     discard that had already happened, which is the reading ADR 0013
//     §5g gave `zoneRoute.Discard` — honoured wherever the card lands,
//     because CR 701.8a defines a discard by its SOURCE.
//   - still in the hand. NOT discarded — the CR 614 window cancelled
//     the move outright, or its prompt was abandoned (ADR 0013 §5j),
//     and nothing ever left.
//
// This is sacrificedThisWayLocked's shape with the hand in place of
// the battlefield, and the two rules have the same form because
// CR 701.8a and CR 701.17a do: both name the keyword action as a move
// OUT of a zone. Destroy is the odd one out (CR 701.7a defines it by
// the graveyard it arrives in — destroyedThisWayLocked).
//
// Read off the live hand rather than off the settled event, for the
// reason destroyedThisWayLocked gives: it is the reading that is still
// true after an undo rewinds into an open prompt.
//
// A player who has LEFT the game reads as "not discarded". CR 800.4a
// took their hand out of the game, so nothing they were asked about
// moved for any reason the asking card can pay out on.
//
// Caller must hold g.mu.
func (g *Game) discardedThisWayLocked(playerID, cardID uuid.UUID) bool {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Eliminated || p.Hand == nil {
		return false
	}
	return !p.Hand.Contains(cardID)
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
//     function returns, and is handed the cards that were really
//     discarded (discardedThisWayLocked). A Mind Rot on a commander
//     returns nil with one card discarded, one prompt open and "then
//     draw a card" still owed.
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
// ability, so DiscardCauseCost sets zoneRoute.MustSettleNow: the
// window still runs — a discard replacement would still see it — but
// it settles without asking. A commander pitched to a cost is still
// offered CR 903.9: its owner is asked BEFORE the payment begins
// (#1397, cost_commander_choice.go) and opts.commanderAnswers carries
// the answer onto the move. The argument for settling, and the other
// half of the same cost line, is payLifeAsCostLocked.
//
// A card that is no longer in the hand is skipped rather than
// erroring: the answer paths re-check the live zone before they
// dequeue, so an answer that got here was legal when it arrived, and
// a half-applied discard that abandoned its continuation would be the
// worse failure.
//
// Caller must hold g.mu.
func (g *Game) discardCardsLocked(playerID uuid.UUID, cards []uuid.UUID, opts discardOptions) error {
	return g.discardBatchLocked(playerID, cards, nil, opts)
}

// discardBatchLocked is discardCardsLocked with the LANDED list it has
// accumulated so far — the value the batch carries forward alongside
// the shrinking `cards` slice, and for the same reason: a fresh slice
// per leg rather than a shared accumulator, so an undo across the
// CR 903.9 prompt has nothing half-written to put back and a replayed
// answer builds the same list a clean run would. #1027.
//
// Caller must hold g.mu.
func (g *Game) discardBatchLocked(playerID uuid.UUID, cards, landed []uuid.UUID, opts discardOptions) error {
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
			DiscardCause:  opts.cause,
			Source:        opts.source,
			Cause:         discardMoveCause(opts.cause, playerID),
			MustSettleNow: opts.cause == DiscardCauseCost,
			// #1397: a cost discard's CR 903.9 answer, asked before
			// the payment (cost_commander_choice.go). Nil everywhere
			// else, which is "unasked".
			commanderAnswer: commanderAnswerFor(opts.commanderAnswers, next),
			then: func(g *Game) error {
				// Asked once this leg has reached a TERMINAL outcome,
				// which is the only moment the answer is stable: the
				// card has landed wherever the window sent it, or the
				// window cancelled the move and it is still in hand.
				grown := landed
				if g.discardedThisWayLocked(playerID, next) {
					grown = make([]uuid.UUID, 0, len(landed)+1)
					grown = append(append(grown, landed...), next)
				}
				return g.discardBatchLocked(playerID, rest, grown, opts)
			},
		})
		return err
	}
	if opts.then == nil {
		return nil
	}
	return opts.then(g, landed)
}
