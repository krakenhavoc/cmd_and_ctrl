package game

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// counter_unless_paid.go — CR 118.12's "counter that spell unless its
// controller pays {N}", and the one rule #951 is about: while that
// question is unanswered, the spell it is about must not resolve.
//
// THE BUG THIS FILE IS. `pay_unless` is the single kind ADR 0018 §6
// lets the table walk past, and the reason it gives is entirely
// Rhystic Study's: the trigger has resolved and left the stack, the
// question is addressed to a DIFFERENT player, and neither answer
// reads the step. Ward borrows the same prompt and none of that
// survives the move. What hangs on the answer is whether an object
// still on the stack is countered, so a table that resolves that
// object while the prompt is open has ANSWERED the question by doing
// — for free, against the payer, and with the ward tax skipped
// (#951, reproduced with Diffusion Sliver). CR 117.4 does not let
// the top of the stack resolve while a required action is
// outstanding, and CR 608.2 makes the counter part of the resolution
// that asked.
//
// WHY IT IS NOT A FLAG. #567's amendment had already moved the
// latitude from the KIND to the PROMPT, and #794 established that
// "does this stop the table" has exactly one answer, read by the
// engine gate and by `internal/legal` through the same predicate.
// What #567 left behind was a card-level "please block" switch, and
// that put the rule back in the hands of whoever writes the next
// card — six counter-unless-pays cards shipped before this file and
// all six were wrong in the same way (and the switch itself was
// missed twice more before #997 removed it).
//
// So the prompt records WHAT ITS DECLINE IS ABOUT
// (PendingChoice.GuardsStackItem) and the gate works the rest out
// for itself. That is also what keeps the halt from becoming a
// wedge: the block is a live read of the stack, so the moment the
// guarded object leaves it — countered in response, fizzled, already
// resolved because the prompt was raised too late — the question
// goes back to being background and the table moves on without
// anybody having to answer it.
//
// SCOPE. This is "<consequence for that spell> unless its controller
// pays…": the consequence is a counter by default and a caller's
// effect since #2365 (Divert's retarget). A
// pay-unless whose decline is a draw (Rhystic Study), a Treasure
// (Smothering Tithe) or a token (Kazuul) guards nothing, keeps ADR
// 0018 §6's latitude, and plays exactly as it did.

// CounterUnlessPaidPrompt is the whole of "counter <StackItem> unless
// its controller pays <Cost>".
//
// One struct rather than six positional arguments, matching the
// engine's other prompt specs (ConfirmPrompt, ChooseCardsPrompt), and
// one door rather than six: every card with this text goes through
// QueueCounterUnlessPaidForEffect so the halt cannot be forgotten by
// the seventh.
type CounterUnlessPaidPrompt struct {
	// StackItem is the object the decline counters — the "that
	// spell" of the printed text, and the object the prompt guards
	// while it is on the stack.
	StackItem uuid.UUID

	// Chooser is the player asked to pay. Zero means "its
	// controller", read off the guarded object, which is what every
	// printed counter-unless-pays says and is the reading that
	// survives a control change on the stack.
	//
	// Ward sets it explicitly. CR 702.21a asks the player who cast
	// the spell that targeted the warded permanent, which the
	// trigger captured as the targeting event's actor; that is the
	// same player in every printed case and the trigger's capture is
	// the authority when it is not.
	Chooser uuid.UUID

	// Source attributes the prompt to a card for the client — the
	// warded permanent, or the counterspell doing the asking.
	Source uuid.UUID

	// OnDecline is the consequence of a decline (or an unaffordable
	// "pay"). Nil is the printed "counter that spell", which is every
	// counter-unless-pays card. A non-nil effect makes this the
	// GUARDED PAY-UNLESS (#2365): the same halt, the same live
	// re-check, a different "or else" — Divert's "change the target
	// of target spell … unless that spell's controller pays {2}".
	//
	// It runs only while StackItem is still on the stack, for the
	// reason the counter does: an object that has left has nothing
	// left for the "unless" to protect. It receives the *Game the
	// answer arrives at, which after an undo is not the one that
	// asked, so it must capture no game state.
	OnDecline func(g *Game) error

	// OnPaid is the consequence of a payment that was actually made
	// (#3591): Assimilate Essence's "If they do, you incubate 2". Nil is
	// every other counter-unless-pays card, where paying just lets the
	// object resolve.
	//
	// It is the paid half of the same answer OnDecline is the unpaid
	// half of, so exactly one of the two runs, and it runs under the
	// same stack guard: answerPayUnlessLocked drops BOTH branches when
	// the guarded object has left the stack, because the "unless" then
	// bought off nothing and there is nothing to be paid for. A "Pay"
	// the chooser cannot fund is a decline, not a payment, so it never
	// reaches here. Like OnDecline it receives the *Game the answer
	// arrives at and must capture no game state.
	OnPaid func(g *Game) error

	// Then is OnDecline and OnPaid as REGISTERED branches
	// (RegisterPayUnlessBranches, #4264), and Carry the plain data they
	// are handed (the guarded object rides PayUnlessAnswer.Guards): a
	// prompt queued with one holds no closure, so a table waiting on it
	// is a restore point. With it OnDecline and OnPaid must be nil. A
	// prompt with none of the three, the printed "counter that spell", is
	// keyed by the engine itself (counterGuardedDecline).
	Then  PayUnlessElse
	Carry []uuid.UUID

	// Asked is the asking stack item as data (NewAskedItem), handed to a
	// keyed branch as PayUnlessAnswer.Item (#4264). Nil when the branch
	// reads none.
	Asked *AskedItem

	// Cost is the printed payment ("{1}", "{2}"), parsed by
	// ParseCost and auto-tapped for like any other pay-unless.
	Cost string

	// LifeCost is mandatory fixed life paid alongside Cost in one
	// payment. It is distinct from optional Phyrexian-symbol payments.
	LifeCost int

	// Question is the dialog header.
	Question string

	// Waterbend is set when the payment is a waterbend cost (CR
	// 701.67a) — "Ward—Waterbend {4}" (#1311): the chooser may tap
	// untapped artifacts and creatures they control, each paying {1}
	// of Cost's generic mana. Nil is an ordinary mana payment. Build
	// it with effects.Waterbend(Cost), so Extra repeats Cost.
	//
	// Without it a waterbend ward would be charged as plain mana,
	// which is HARDER to pay than printed — the card stronger than
	// it is, the #259 direction.
	Waterbend *TapPermanentsCost
}

// QueueCounterUnlessPaidForEffect queues the CR 118.12 prompt for
// `p.StackItem` and returns. The ability that asked has finished
// resolving; what is left is the payer's decision, which arrives as
// an ordinary resolve_choice.
//
// An object that is no longer on the stack is not asked about at all:
// there is nothing to counter and so nothing to charge for (CR
// 118.12 — the "unless" buys off a consequence that can no longer
// happen). That is the same no-op the cards used to spell out one at
// a time before reaching for the prompt.
//
// A decline — or a "pay" the chooser cannot fund — counters the
// object if it is still there, and does nothing if it is not. With
// p.OnDecline set it runs that instead, under the same guard.
//
// Caller must hold g.mu. Added in #951.
func (g *Game) QueueCounterUnlessPaidForEffect(p CounterUnlessPaidPrompt) error {
	if p.LifeCost < 0 {
		return ErrInvalidParam
	}
	guarded := g.StackItemForEffect(p.StackItem)
	if guarded == nil {
		return nil
	}
	chooser := p.Chooser
	if chooser == uuid.Nil {
		chooser = guarded.Controller
	}
	stackID := p.StackItem
	// #4264: the registered branches, or the engine's own for the plain
	// "counter that spell", leave the prompt without a closure. A
	// waterbend payment keeps its frame.
	then, carry := p.Then, p.Carry
	if then.key != "" && (p.OnDecline != nil || p.OnPaid != nil) {
		return fmt.Errorf("%w: keyed counter-unless-pays %q also sets a closure", ErrInvalidParam, then.key)
	}
	// A waterbend clause stays a closure unless its kind is registered
	// (RegisterTapCost), which is what lets a prompt carry it as data.
	tap := payTapSpecOf(p.Waterbend)
	keyable := p.Waterbend == nil || tap != nil
	if then.key == "" && p.OnDecline == nil && p.OnPaid == nil && keyable {
		then, carry = counterGuardedDecline, nil
	}
	if then.key != "" && keyable {
		return g.QueuePayUnlessPromptKeyedForEffect(PayUnlessKeyedPrompt{
			Chooser: chooser, Source: p.Source, Cost: p.Cost, Question: p.Question,
			Then: then, Carry: carry, Guards: stackID, LifeCost: p.LifeCost, Asked: p.Asked,
			Tap: tap,
		})
	}
	decline := func(g *Game) error { return g.counterIfStillOnStackLocked(stackID) }
	if p.OnDecline != nil {
		custom := p.OnDecline
		decline = func(g *Game) error {
			if g.StackItemForEffect(stackID) == nil {
				return nil
			}
			return custom(g)
		}
	}
	return g.queuePayUnlessLocked(chooser, p.Source, p.Cost, p.Question,
		decline,
		TurnStep{}, stackID, p.Waterbend, payUnlessExtras{lifeCost: p.LifeCost, onPay: p.OnPaid})
}

// counterGuardedDecline is the printed "counter that spell", as the
// keyed decline of every counter-unless-pays that sets no consequence of
// its own (#4264): the guarded object rides the prompt (GuardsStackItem),
// so the prompt needs no closure and a table waiting on it is a restore
// point. A spell that has left the stack by the answer is not an error.
var counterGuardedDecline = RegisterPayUnlessElse("pay-unless/counter-guarded", func(g *Game, r PayUnlessAnswer) error {
	return g.counterIfStillOnStackLocked(r.Guards)
})

// counterIfStillOnStackLocked counters `stackID` unless it has
// already left the stack, in which case the "unless" clause bought
// off nothing and that is not an error. Caller must hold g.mu.
func (g *Game) counterIfStillOnStackLocked(stackID uuid.UUID) error {
	if g.StackItemForEffect(stackID) == nil {
		return nil
	}
	if err := g.CounterTargetForEffect(stackID); err != nil && !errors.Is(err, ErrCardNotOnStack) {
		return err
	}
	return nil
}

// CounterIfStillOnStackForEffect counters `stackID` unless it has already
// left the stack, in which case the "unless" clause bought off nothing and
// that is not an error. It is the decline a keyed counter-unless-pays names
// when it also has a payment branch of its own (Assimilate Essence's "if
// they do, you incubate 2"): a registered key's branches replace the
// engine's own "counter that spell", so a card that registers a pay branch
// registers this as its decline.
//
// Caller must hold g.mu.
func (g *Game) CounterIfStillOnStackForEffect(stackID uuid.UUID) error {
	return g.counterIfStillOnStackLocked(stackID)
}

// choiceGuardsALiveStackItem reports whether `c` is a prompt whose
// answer still decides the fate of an object on the stack — the live
// half of Game.ChoicePromptBlocksTable (choice_gate.go).
//
// Live, not a snapshot taken when the prompt was queued: the guarded
// object can leave the stack while the question is open (a second
// counterspell underneath the same trigger, a fizzle), and when it
// does the prompt stops blocking on its own. A halt that could only
// be lifted by answering is how a prompt nobody can answer wedges a
// table; this one cannot outlive what it is about.
func (g *Game) choiceGuardsALiveStackItem(c *PendingChoice) bool {
	if g == nil || c == nil || c.GuardsStackItem == uuid.Nil {
		return false
	}
	return g.StackItemForEffect(c.GuardsStackItem) != nil
}
