package game

import "github.com/google/uuid"

// chained_choice.go — composition for the pending-choice queue: how one
// prompt asks the next one.
//
// # The gap this closes
//
// Every prompt kind before this one was a LEAF. An effect queued a
// choice, the resolver applied the answer, and the stashed continuation
// ran the rest of the effect to completion. Nothing in the queue could
// express "and now ask them this, which depends on what they just
// said" — so Sylvan Library ("choose two cards drawn this turn; for
// each, pay 4 life or put it back") and Ponder's "you may shuffle" were
// both unimplementable, and the Ponder file said so in a comment.
//
// # The mechanism, in one sentence
//
// A choice's continuation runs with g.mu already held, so it may call
// QueueChoiceForEffect itself — the follow-up prompt is appended to the
// same queue, the client's modal picks it up on the next frame, and the
// enumerator offers its answers to a bot the same way. Chaining is
// therefore not a new subsystem: it is the existing continuation
// contract, plus two prompt kinds general enough to be the links.
//
// That is all a chain is. The work in this file is making the two
// general links exist, be answerable by every client (human modal, bot
// enumerator, gamecli), and be honest about what they cost:
//
//   - PendingChoiceConfirm — "do A, or do B": a two-way question whose
//     branches are named by the card. Ponder's "you may shuffle" is the
//     degenerate form with an inert decline.
//   - PendingChoiceChooseCards — "choose N of these cards": a card-set
//     pick whose continuation receives the picks and decides what they
//     mean. Sylvan Library's "choose two cards in your hand drawn this
//     turn" is the first one; the roadmap's "put or pick a card from
//     hand at resolution" family is the rest.
//
// Neither kind knows what it is for. What happens on each answer lives
// entirely in the frame the queuing effect supplies, which is what
// makes them reusable rather than two more bespoke prompts.
//
// # What chaining costs: restorability
//
// A continuation is a closure and cannot be serialised. GameSnapshot
// already counts every choice holding one (ContinuationCensus.
// ChoiceResumeFrames) and Restorable() is Continuations.Empty(), so a
// game with ANY continuation-bearing prompt open — a scry, a search, an
// optional trigger, and now these — writes no restore point until it is
// answered. The two frames below are added to that census, which is the
// thing that must not be forgotten: an uncounted frame would let the
// server write a restore point that silently drops the rest of the
// effect.
//
// Chaining does not introduce that cost, but it does LENGTHEN it: a
// two-link chain holds the window open across two answers instead of
// one. In wall-clock terms that is the difference between one prompt
// and two, on a table that is otherwise blocked on the same player
// anyway — the engine refuses pass_priority while a choice is open, so
// the game was already waiting. It is still a real cost and #544 is the
// reason to say so out loud.
//
// # What a chain must never do
//
// Queue a prompt no enumerator can answer. `legal.EnumerateFor` returns
// ONLY a choice's answers while a seat owes one, so a kind `legal` does
// not know is a bot seat with an empty move list, asleep, holding the
// table — that is #544 and #499. Both kinds below have a case in
// internal/legal/choices.go, and both are answerable without any
// knowledge of the card that queued them.

// PendingChoiceConfirm is the general two-way prompt: "do A, or do B."
//
// Answered with the same {apply: bool} payload the other yes/no kinds
// use — true takes the accept branch, false the decline branch — and
// routed by kind in the actions dispatcher.
//
// It is NOT a fifth spelling of pay_unless / trigger_prompt /
// optional_replacement / may_cast / entry_pay_life. Each of those is
// welded to a specific engine pipeline: the replacement apply-loop, the
// trigger harvester's Build closure, a parsed mana cost. This one is
// welded to nothing. Its two branches are plain continuations supplied
// by whoever queued it, which is exactly what a chain link has to be —
// the second prompt of a chain is not a replacement effect or a
// trigger, it is "the rest of this card, now that I know your answer".
//
// AcceptLabel / DeclineLabel are the card's own words for the two
// branches ("Pay 4 life" / "Put it on top"), because a chained question
// is rarely a yes/no in the player's head. A card that really is asking
// a yes/no leaves them empty and the client renders Yes / No.
const PendingChoiceConfirm PendingChoiceKind = "confirm"

// PendingChoiceChooseCards is the general card-set pick: "choose N of
// these cards", with the continuation deciding what being chosen means.
//
// Answered with {card_ids: []string} — the same payload discard,
// sacrifice, search and copy already use — and routed by kind.
//
// Deliberately not PendingChoiceSearchLibrary with a different
// candidate list, even though the QUESTION is the same shape. Two
// things differ and both are observable: a search is a search (it emits
// EventSearchLibrary, it shuffles, and the wire withholds the option
// list from non-choosers because the SIZE of the match set is hidden
// information about a library), and, for a bot, the sign is inverted —
// the heuristic's search branch scores "take the best", while the cards
// picked here are as often the ones being given up. Reusing the kind
// would have made Sylvan Library pay life to keep its worst two cards.
const PendingChoiceChooseCards PendingChoiceKind = "choose_cards"

// confirmFrame is the continuation pair behind a PendingChoiceConfirm.
// Both callbacks receive the live *Game (not a captured one) on the
// same undo-safety contract payUnlessFrame and StackItem.Effect follow,
// and both run with g.mu held — so either may queue the next link of
// the chain.
//
// Either may be nil. Ponder's "you may shuffle" declines into nothing
// at all, and an effect should not have to write an empty closure to
// say so.
type confirmFrame struct {
	onAccept  func(g *Game) error
	onDecline func(g *Game) error
}

// chooseCardsFrame is the continuation behind a
// PendingChoiceChooseCards. `then` receives the picks in the order the
// chooser submitted them and runs with g.mu held.
//
// zone, when set, is the zone the picks are re-checked against on
// submit. The prompt is asynchronous and the board moves under it — a
// card chosen from hand can be discarded by an opponent's effect before
// the answer arrives — so an unchecked pick would let an effect operate
// on a card that is somewhere else entirely. Zero value means "no zone
// re-check"; the frame's `then` owns the validation in that case.
type chooseCardsFrame struct {
	zone ZoneKind
	then func(g *Game, picked []uuid.UUID) error
}

// ConfirmPrompt is the queue-side description of a
// PendingChoiceConfirm. A struct rather than eight positional
// arguments, because the labels and the two branches are all optional
// in different combinations.
type ConfirmPrompt struct {
	// Chooser answers the prompt. Required.
	Chooser uuid.UUID
	// Source is the card asking. Empty is legal (test harnesses).
	Source uuid.UUID
	// Question is the prompt's header — the card's own sentence.
	Question string
	// AcceptLabel / DeclineLabel name the two branches. Empty renders
	// as Yes / No.
	AcceptLabel, DeclineLabel string
	// LifeCost is the life the accept branch charges, for the wire and
	// for the bot's pricing. The branch itself still performs (and
	// re-checks) the payment — this is a declaration, not a deduction.
	LifeCost int
	// OnAccept / OnDecline are the branches. Either may be nil. Both
	// run with g.mu held and may queue further choices.
	OnAccept, OnDecline func(g *Game) error
}

// QueueConfirmForEffect queues a two-way prompt and returns its ID.
//
// Caller must hold g.mu.
func (g *Game) QueueConfirmForEffect(p ConfirmPrompt) uuid.UUID {
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:         PendingChoiceConfirm,
		Chooser:      p.Chooser,
		FromPlayer:   p.Chooser,
		Count:        1,
		Source:       p.Source,
		Reason:       p.Question,
		AcceptLabel:  p.AcceptLabel,
		DeclineLabel: p.DeclineLabel,
		LifeCost:     p.LifeCost,
		confirmResume: &confirmFrame{
			onAccept:  p.OnAccept,
			onDecline: p.OnDecline,
		},
	})
}

// ChooseCardsPrompt is the queue-side description of a
// PendingChoiceChooseCards.
type ChooseCardsPrompt struct {
	// Chooser answers the prompt. Required.
	Chooser uuid.UUID
	// FromPlayer owns the zone the candidates live in. Defaults to
	// Chooser when empty — the two differ for the same reason they do
	// on a Thoughtseize discard.
	FromPlayer uuid.UUID
	// Source is the card asking.
	Source uuid.UUID
	// Question is the prompt's header.
	Question string
	// Cards are the candidates, in the order the client should show
	// them.
	Cards []uuid.UUID
	// Min / Max bound the pick. Max <= 0 means len(Cards).
	Min, Max int
	// Zone, when set, is re-checked on submit — every pick must still
	// be in FromPlayer's zone of that kind.
	Zone ZoneKind
	// Then receives the picks. Runs with g.mu held; may queue further
	// choices, which is how a chain continues.
	Then func(g *Game, picked []uuid.UUID) error
}

// QueueChooseCardsForEffect queues a card-set pick and returns its ID.
//
// It does NOT short-circuit an empty or forced candidate set. A prompt
// whose only legal answer is "all of them" still goes through the
// queue, because the chain's next link is queued by the resolver: a
// silent shortcut here would have to duplicate that continuation and
// would rot the first time the two diverged. The cost is one click the
// player had no choice about, which is the cheaper mistake.
//
// Caller must hold g.mu.
func (g *Game) QueueChooseCardsForEffect(p ChooseCardsPrompt) uuid.UUID {
	from := p.FromPlayer
	if from == uuid.Nil {
		from = p.Chooser
	}
	hi := p.Max
	if hi <= 0 || hi > len(p.Cards) {
		hi = len(p.Cards)
	}
	lo := p.Min
	if lo < 0 {
		lo = 0
	}
	if lo > hi {
		lo = hi
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:        PendingChoiceChooseCards,
		Chooser:     p.Chooser,
		FromPlayer:  from,
		Count:       hi,
		Source:      p.Source,
		Reason:      p.Question,
		ChooseCards: append([]uuid.UUID(nil), p.Cards...),
		ChooseMin:   lo,
		ChooseMax:   hi,
		chooseCardsResume: &chooseCardsFrame{
			zone: p.Zone,
			then: p.Then,
		},
	})
}

// ResolveConfirm answers a PendingChoiceConfirm. `accept` picks the
// branch; the other one is discarded.
//
// The choice is dequeued BEFORE the branch runs, which is what makes
// chaining work: the branch queues the next prompt, and a queue that
// still held this one would hand the client two open modals and the
// enumerator two owed choices for the same seat.
//
// A branch that errors emits EventEffectError and the prompt is still
// gone. That matches every other resolver here — a failed continuation
// must not leave an unanswerable prompt in the queue, which is the
// #544 shape.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveConfirm(choiceID, chooserID uuid.UUID, accept bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceConfirm {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	frame := choice.confirmResume
	source := choice.Source
	g.dequeueChoiceLocked(idx)
	if frame == nil {
		// Nothing to run. The prompt is gone either way rather than
		// stuck: a confirm with no frame is a bug in whoever queued
		// it, and refusing the answer would wedge the seat.
		return nil
	}
	branch := frame.onDecline
	if accept {
		branch = frame.onAccept
	}
	if branch != nil {
		if err := branch(g); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    chooserID,
				Source:   source,
				ErrorMsg: err.Error(),
			})
		}
	}
	g.runStateChecksLocked()
	return nil
}

// ResolveChooseCards answers a PendingChoiceChooseCards: `picks` are
// the cards the chooser selected.
//
// Validation rejects before dequeuing, so a client that submits an
// illegal set gets an error and can try again rather than losing the
// prompt — the same contract ResolveSearchLibrary documents. The
// difference, and it is the #544 lesson, is that every answer the
// enumerator offers passes this validation: the bounds are on the
// choice itself, so `legal` enumerates exactly the sets that are legal
// rather than a superset it cannot see the constraint on.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveChooseCards(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceChooseCards {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if len(picks) < choice.ChooseMin || len(picks) > choice.ChooseMax {
		return ErrInvalidParam
	}
	candidates := make(map[uuid.UUID]bool, len(choice.ChooseCards))
	for _, id := range choice.ChooseCards {
		candidates[id] = true
	}
	frame := choice.chooseCardsResume
	seen := make(map[uuid.UUID]bool, len(picks))
	for _, id := range picks {
		if !candidates[id] || seen[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		if frame != nil && frame.zone != "" {
			// Re-checked against the LIVE zone, not the frozen
			// candidate list: the prompt is asynchronous and a card
			// can leave between the question and the answer.
			z := g.findCardZoneLocked(id)
			if z == nil || z.Kind != frame.zone {
				return ErrCardNotFound
			}
			// Per-player zones carry an owner; the shared ones
			// (battlefield, exile, stack) leave it nil.
			if z.Owner != uuid.Nil && z.Owner != choice.FromPlayer {
				return ErrCardNotFound
			}
		}
	}
	source := choice.Source
	g.dequeueChoiceLocked(idx)
	if frame == nil || frame.then == nil {
		return nil
	}
	if err := frame.then(g, append([]uuid.UUID(nil), picks...)); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    chooserID,
			Source:   source,
			ErrorMsg: err.Error(),
		})
	}
	g.runStateChecksLocked()
	return nil
}

// findChoiceLocked returns the queue index and entry for a choice ID,
// or (-1, nil). Caller must hold g.mu.
func (g *Game) findChoiceLocked(choiceID uuid.UUID) (int, *PendingChoice) {
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			return i, c
		}
	}
	return -1, nil
}
