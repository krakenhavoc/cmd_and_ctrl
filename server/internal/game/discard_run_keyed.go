package game

import (
	"fmt"

	"github.com/google/uuid"
)

// discard_run_keyed.go — "discard <n> cards, then <the rest>" whose rest
// is a KEY rather than a closure (#4256), so a table waiting on the
// discard is a restore point.
//
// PlayerDiscardsThenForEffect keeps the rest of the card in a prompted
// run (prompt_run.go): a closure on Game.promptRuns, linked from the
// prompt by a run id, and the prompt's own continuation is a closure
// too. Neither is written to a restore point, so a table waiting on
// learn's "you may discard a card" could be undone but not saved and
// restored.
//
// The keyed form is for the one-seat shape, which is one prompt and so
// needs no run: the prompt is a keyed choose_cards pick over the
// player's hand (choose_cards_keyed.go) whose registered continuation
// discards the picks through the one discard path (discardCardsLocked)
// and, once they have landed, runs the card's registered rest with what
// was really discarded. What the rest needs is plain data on the
// prompt: the player (Chooser), the card that asked (Source) and the IDs
// the card put there (Carry). The key shares the effect-key namespace
// and is ledgered as a choose-cards continuation ("choose <key>").
//
// What it gives up is the closure form's per-prompt extra, the leg's own
// Then, a closure. The set rule (DiscardPrompt.Validate, "exactly two or
// none") is registered with the continuation instead
// (RegisterDiscardThenRule, #4264). A pick that has left the hand by the
// time the answer comes in is skipped by the discard itself, as every
// discard skips a card no longer in hand.
// A discard the CR 903.9 commander prompt pauses part-way is still a
// closure for as long as that prompt is open, which is the commander
// prompt's own frame and no change here.

// DiscardsDone is what a keyed discard-then's continuation is handed.
type DiscardsDone struct {
	// Player discarded (or was asked to); Source is the card that asked.
	Player, Source uuid.UUID
	// Discarded is what really left the player's hand this way
	// (PromptedDiscards' reading). Empty when nothing was asked (an
	// empty hand, CR 701.9a) or the prompt ended unanswered.
	Discarded PromptedDiscards
	// Carry is the plain data the queuing card put on the prompt.
	Carry []uuid.UUID
}

// DiscardThenFunc is a registered discard-then continuation. It runs
// with g.mu held, may queue prompts of its own, and must capture
// nothing.
type DiscardThenFunc func(g *Game, r DiscardsDone) error

// DiscardThen names a registered discard-then continuation. The key is
// unexported, so the only way to hold one is RegisterDiscardThen.
type DiscardThen struct {
	then ChooseCardsThen
	fn   DiscardThenFunc
}

// Key is the continuation's on-disk key.
func (d DiscardThen) Key() string { return d.then.key }

// RegisterDiscardThen registers a keyed discard-then continuation. Call
// it once, from a package-level var. Panics on a bad or duplicate key.
// The key is an on-disk identity: never renamed, never reused, and
// ledgered as "choose <key>".
func RegisterDiscardThen(key string, fn DiscardThenFunc) DiscardThen {
	return RegisterDiscardThenRule(key, nil, fn)
}

// RegisterDiscardThenRule is RegisterDiscardThen for a discard with a
// set rule over what is chosen (#4264): Thrilling Discovery's "discard
// two cards", which is two or none, never one. The rule is never called
// for an empty pick.
func RegisterDiscardThenRule(key string, rule SetRuleFunc, fn DiscardThenFunc) DiscardThen {
	if fn == nil {
		panic(fmt.Sprintf("game: discard-then continuation %q has no function", key))
	}
	then := RegisterChooseCardsThenRule(key, rule, func(g *Game, r CardsChosen) error {
		done := func(g *Game, landed []uuid.UUID) error {
			return fn(g, DiscardsDone{
				Player:    r.Chooser,
				Source:    r.Source,
				Discarded: PromptedDiscards{{Seat: r.Chooser, Cards: landed}},
				Carry:     r.Carry,
			})
		}
		if !r.Answered {
			// The chooser left with the prompt open: they discarded
			// nothing, and the rest of the card still runs, as the run's
			// withdrawn leg does (settleRunLegLocked).
			return done(g, nil)
		}
		return g.discardCardsLocked(r.Chooser, r.Picked, discardOptions{
			cause:  DiscardCauseEffect,
			source: r.Source,
			then:   done,
		})
	})
	return DiscardThen{then: then, fn: fn}
}

// PlayerDiscardsThenKeyedForEffect is PlayerDiscardsThenForEffect with
// the rest of the card as a registered continuation and `carry` the
// plain data it is handed: learn's "you may discard a card. If you do,
// draw a card", Abandon Attachments' "you may discard a card. If you do,
// draw two cards". The prompt carries no closure, so a table waiting on
// it is a restore point.
//
// The prompt's Then and Validate must be nil: they are closures, and a
// restore would drop them (ErrInvalidParam). A set rule is registered
// with the continuation (RegisterDiscardThenRule). When nothing is asked (an
// empty hand, a discard an opponent's effect can't cause, a player who
// has left) the continuation runs at once with nothing discarded, as
// the closure form's does.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) PlayerDiscardsThenKeyedForEffect(p DiscardPrompt, then DiscardThen, carry []uuid.UUID) error {
	if then.then.key == "" || then.fn == nil {
		return fmt.Errorf("%w: keyed discard-then with no continuation", ErrInvalidParam)
	}
	if p.Then != nil || p.Validate != nil {
		return fmt.Errorf("%w: keyed discard-then %q also sets a closure", ErrInvalidParam, then.then.key)
	}
	p.keyed = then.then
	p.carry = copyUUIDs(carry)
	if g.queueDiscardPromptLocked(p, uuid.Nil) != uuid.Nil {
		return nil
	}
	return then.fn(g, DiscardsDone{Player: p.Player, Source: p.Source, Carry: copyUUIDs(carry)})
}

// discardLegThen is the engine's own continuation for a keyed discard
// LEG (#4290): a prompt of a multi-seat discard, or of a fan-out nothing
// waits on, queued as a keyed choose_cards pick over the seat's hand.
// It discards the picks through the one discard path, and once they
// have landed runs the run's per-seat tail (RegisterPromptRunThenLeg)
// and settles the leg, which runs the run's continuation after the
// last one. With no run (a fire-and-forget fan-out) it only discards.
//
// A leg whose chooser leaves is settled with nothing discarded by the
// departure table before this is reached (defaultDroppedChoiceLocked
// reads the run link first); the !Answered branch is the same rule for
// a leg reached any other way.
var discardLegThen = RegisterChooseCardsThen("choose-cards/discard-leg", func(g *Game, r CardsChosen) error {
	if !r.Answered {
		return g.settleRunLegLocked(r.Run, r.Chooser, nil)
	}
	return g.discardCardsLocked(r.Chooser, r.Picked, discardOptions{
		cause:  DiscardCauseEffect,
		source: r.Source,
		then: func(g *Game, landed []uuid.UUID) error {
			if run := g.promptRuns[r.Run]; run != nil {
				g.runLegTailLocked(run.thenKey, r.Chooser, r.Source, landed, run.carry)
			}
			return g.settleRunLegLocked(r.Run, r.Chooser, landed)
		},
	})
})

// EachPlayerDiscardsThenKeyedForEffect is EachPlayerDiscardsThenForEffect
// with the rest of the card as a registered run continuation
// (RegisterPromptRunThen) and `carry` the plain data it is handed: "each
// opponent discards a card. You draw a card for each card discarded this
// way" (Syphon Mind, with the caster on the carry). Neither the run nor
// its legs hold a closure, so a table waiting on any leg is a restore
// point (#4290).
//
// The prompt's Then and Validate must be nil: they are closures. A
// per-seat tail ("…then mills a card") is registered with the run
// (RegisterPromptRunThenLeg). When nobody is asked the continuation runs
// at once with nothing discarded, as the closure form's does.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) EachPlayerDiscardsThenKeyedForEffect(except uuid.UUID, p DiscardPrompt, then PromptRunThen, carry []uuid.UUID) error {
	return g.discardRunKeyedLocked(g.discardFanOutSeatsLocked(except), p, then, carry)
}

// PlayersDiscardThenKeyedForEffect is PlayersDiscardThenForEffect with
// the rest of the card as a registered run continuation; the seats are
// `players`, in the order given. See EachPlayerDiscardsThenKeyedForEffect.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) PlayersDiscardThenKeyedForEffect(players []uuid.UUID, p DiscardPrompt, then PromptRunThen, carry []uuid.UUID) error {
	return g.discardRunKeyedLocked(players, p, then, carry)
}

// discardRunKeyedLocked is discardRunLocked with a keyed continuation and
// keyed legs.
//
// Caller must hold g.mu.
func (g *Game) discardRunKeyedLocked(asks []uuid.UUID, p DiscardPrompt, then PromptRunThen, carry []uuid.UUID) error {
	if then.key == "" {
		return fmt.Errorf("%w: keyed discard run with no continuation", ErrInvalidParam)
	}
	if p.Then != nil || p.Validate != nil || p.keyed.key != "" {
		return fmt.Errorf("%w: keyed discard run %q also sets a closure", ErrInvalidParam, then.key)
	}
	p.keyedLeg = true
	p.runThen = then.key
	p.carry = copyUUIDs(carry)
	_, err := g.runPromptsKeyedLocked(asks, then, p.Source, RunData{Carry: carry}, func(seat, run uuid.UUID) bool {
		leg := p
		leg.Player = seat
		return g.queueDiscardPromptLocked(leg, run) != uuid.Nil
	})
	return err
}
