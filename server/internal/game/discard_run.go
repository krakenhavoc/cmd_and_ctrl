package game

import (
	"github.com/google/uuid"
)

// discard_run.go — the CONTINUATION of a prompted discard
// (#1027, CR 701.8a, ADR 0013 §5y).
//
// QueueDiscardChoiceForEffect does not discard anything. It queues a
// QUESTION over the player's own hand and returns the prompt's ID;
// the cards leave the hand when that player answers, which is one or
// more actions later, and they can leave LATER STILL than that — a
// discarded commander's CR 903.9 prompt pauses the batch mid-way (ADR
// 0013 §5g). So every clause a card wrote on the next line ran before
// the discard: Archon of Cruelty's "sacrifices …, discards a card,
// AND LOSES 3 LIFE. You draw a card and gain 3 life" ran its last
// three clauses while two prompts were still open, which is the wrong
// printed order and an observable one.
//
// A RUN is the fix, and it is #1019's run exactly — the mechanic is
// shared (prompt_run.go) and this file is the discard's three entry
// points, its "this way" reading and the vocabulary a card writes its
// clause in. ONE RUN IS ONE PRINTED INSTRUCTION: "target player
// discards two cards" over one seat, "each opponent discards a card"
// over three, however many prompts that takes. Its continuation runs
// ONCE, after the LAST prompt has settled and the cards it named have
// finished moving, and is told per seat what was really discarded.
//
// Three properties, and each of them is a bug that was live before:
//
//   - It waits for every asked seat. Syphon Mind's "you draw a card
//     for each card discarded this way" drew one card per ANSWER, off
//     each prompt's own Then, so the caster's cards arrived spread
//     across the table's decisions instead of once at the end; and the
//     card had to hand-roll an empty-hand skip to stop drawing for a
//     seat that discarded nothing.
//   - It waits for the MOVE, not for the answer, and it counts what
//     CR 701.8a counts. `landed` is discardedThisWayLocked's answer:
//     the card left the HAND, wherever a replacement then sent it. A
//     madness card exiled instead of binned (CR 702.35a, #1008) was
//     discarded; a Library of Leng card put on top of the library was
//     discarded; a leg the CR 614 window cancelled was not.
//   - A prompt the engine WITHDRAWS settles its leg with nothing
//     discarded rather than stranding the run. That is #1016's
//     dropDefault at a third kind — and the first one whose PendingChoice
//     Kind (choose_cards) is shared with prompts that are no run's leg,
//     which is why defaultDroppedChoiceLocked branches on the run LINK.
//
// NOT A RUN: a discard paid as a COST. CR 601.2h pays a spell's costs
// as one indivisible step and the discard settles without asking
// (zoneRoute.MustSettleNow, ADR 0013 §5g), so there is no prompt to
// wait for and nothing for a continuation to buy. discard_cost.go and
// additional_cost.go stay where they are.

// PromptedDiscards is what a prompted discard's continuation is
// handed: one entry per seat the run ASKED, in the order the seats
// were asked (APNAP from the active player for a fan-out), carrying
// the cards that really left that seat's hand.
//
// A seat that was skipped at queue time — an empty hand, so CR 701.8a's
// "as many as you can" asked it nothing, or a seat that had already
// left the game — has no entry at all. `Discarded` reads the same for
// it as for a seat that was asked and discarded nothing, which is
// right: both discarded nothing. A caller that needs to tell them
// apart is asking about the QUESTION rather than about the discard,
// and should not be reading this.
type PromptedDiscards []SeatCards

// By returns the cards `seat` discarded this way, or nil.
func (d PromptedDiscards) By(seat uuid.UUID) []uuid.UUID { return seatCardsBy(d, seat) }

// Discarded reports whether `seat` discarded at least one card this
// way — "if you discard a card this way, …" written as a card prints
// it.
func (d PromptedDiscards) Discarded(seat uuid.UUID) bool { return len(seatCardsBy(d, seat)) > 0 }

// Cards flattens the run into every card discarded this way, in ask
// order — "for each card discarded this way".
func (d PromptedDiscards) Cards() []uuid.UUID { return seatCardsFlat(d) }

// Count is how many cards were discarded this way, across every seat.
// Syphon Mind draws that many.
func (d PromptedDiscards) Count() int { return seatCardsCount(d) }

// PlayerDiscardsThenForEffect is QueueDiscardChoiceForEffect with the
// rest of the card attached: "target opponent … discards a card, and
// loses 3 life" (Archon of Cruelty), "discard two cards, then draw
// three".
//
// `p.Player` is the seat asked. One prompt goes up, however many cards
// the instruction asks for — a discard of N is one question over the
// hand, not N questions — so this run has exactly one leg.
//
// `then` runs once that leg has settled AND the cards it named have
// finished moving, and it runs even when nothing was discarded,
// including when the hand was empty and no prompt went up: a
// continuation is the rest of a card that is paused mid-resolution
// (#544, #1006), and one that is silently never called is a card that
// stops halfway.
//
// Note what this does NOT replace: `p.Then` is this PROMPT's own
// "then", the rest of its own sentence, and a card may set both — see
// DiscardPrompt.Then. The prompt's runs first.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) PlayerDiscardsThenForEffect(p DiscardPrompt, then func(g *Game, discarded PromptedDiscards) error) error {
	_, err := g.discardRunLocked([]uuid.UUID{p.Player}, p, then)
	return err
}

// PlayersDiscardThenForEffect is the GENERAL form: one prompt per
// entry in `players`, in the order given, with one continuation over
// the lot — the seats being a spell's targets rather than the whole
// table.
//
// `p.Player` is IGNORED; the seats come from `players`. Everything
// else on the prompt (the count, the "up to", the set-level Validate,
// the per-leg Then, the question) is the template each seat is asked
// with, which is what makes "each opponent discards TWO cards" one
// call. A caller that names the same seat twice gets two prompts and
// one entry in the answer.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) PlayersDiscardThenForEffect(
	players []uuid.UUID,
	p DiscardPrompt,
	then func(g *Game, discarded PromptedDiscards) error,
) error {
	_, err := g.discardRunLocked(players, p, then)
	return err
}

// EachPlayerDiscardsThenForEffect is the APNAP fan-out: "each player
// discards a card", "each other player discards a card", "each
// opponent discards a card. You draw a card for each card discarded
// this way" (Syphon Mind).
//
// `except` is the player who does NOT discard ("each OTHER player" /
// "each opponent"), or uuid.Nil when everyone does. Seats are walked
// from the active player, which is the order their entries appear in
// the answer. `p.Player` is ignored, as it is for the general form.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) EachPlayerDiscardsThenForEffect(
	except uuid.UUID,
	p DiscardPrompt,
	then func(g *Game, discarded PromptedDiscards) error,
) error {
	_, err := g.discardRunLocked(g.discardFanOutSeatsLocked(except), p, then)
	return err
}

// EachPlayerDiscardsForEffect is the fan-out with NOTHING waiting on
// it — Liliana of the Veil's "+1: each player discards a card",
// Burglar Rat's "each opponent discards a card".
//
// EachPlayerDiscardsThenForEffect without the continuation, and the
// loop five catalog callers used to write out. Returns the number of
// prompts queued, which is a count of QUESTIONS and not of discards:
// no clause may be gated on it (the payout lint's `discardPromptVerb`
// row says so).
//
// Caller must hold g.mu.
func (g *Game) EachPlayerDiscardsForEffect(except uuid.UUID, p DiscardPrompt) int {
	queued, _ := g.discardRunLocked(g.discardFanOutSeatsLocked(except), p, nil)
	return queued
}

// discardFanOutSeatsLocked is "each player", APNAP from the active
// player, minus `except` and minus the seats that have left — the ask
// list both fan-out entry points build, written once so they cannot
// walk the table two ways.
//
// Caller must hold g.mu.
func (g *Game) discardFanOutSeatsLocked(except uuid.UUID) []uuid.UUID {
	var seats []uuid.UUID
	n := len(g.Seats)
	if n == 0 {
		return nil
	}
	start := g.Turn.ActiveSeat
	for i := 0; i < n; i++ {
		p := g.Seats[(start+i)%n]
		if p == nil || p.Eliminated || p.ID == except {
			continue
		}
		seats = append(seats, p.ID)
	}
	return seats
}

// discardRunLocked is the discard verb's half of a prompted run: the
// shared body (runPromptsLocked) with the discard prompt as its queue
// step and PromptedDiscards as the name its continuation reads the
// answer under.
//
// Caller must hold g.mu.
func (g *Game) discardRunLocked(
	asks []uuid.UUID,
	p DiscardPrompt,
	then func(g *Game, discarded PromptedDiscards) error,
) (int, error) {
	var landedThen func(*Game, []SeatCards) error
	if then != nil {
		landedThen = func(g *Game, landed []SeatCards) error {
			return then(g, PromptedDiscards(landed))
		}
	}
	return g.runPromptsLocked(asks, landedThen, func(seat, run uuid.UUID) bool {
		// CR 701.8a's "as many as you can": a seat with an empty hand
		// is asked nothing and gets no leg, and a seat that has left
		// the game can never answer. Both are
		// queueDiscardPromptLocked's own rules, unchanged — the seat
		// template is copied per ask so one run's prompts cannot share
		// a mutated struct.
		leg := p
		leg.Player = seat
		return g.queueDiscardPromptLocked(leg, run) != uuid.Nil
	})
}
