package game

import (
	"strconv"

	"github.com/google/uuid"
)

// option_pick.go — "choose one of the following", asked of a named
// seat while an effect is resolving (CR 608.2).
//
// # The gap this closes (#568)
//
// The queue could already ask a yes/no with two card-supplied branches
// (PendingChoiceConfirm, chained_choice.go) and a card-set pick
// (PendingChoiceChooseCards). What it could not ask is a choice among
// THREE OR MORE named consequences — Torment of Hailfire's "each
// opponent loses 3 life unless that player sacrifices a nonland
// permanent of their choice or discards a card" is the card that
// forced it, and a confirm cannot express it without nesting two
// yes/nos and asking the second question in an order the card does
// not print.
//
// A modal spell's "choose one —" is NOT this kind. That choice is made
// at announce (CR 601.2b), lives on Spec.Modes, and is over by the time
// the spell resolves. This one happens DURING resolution, is frequently
// addressed to somebody other than the controller, and each branch is a
// plain continuation supplied by the card.
//
// # Addressed by seat
//
// Chooser is any seat, not the resolving effect's controller. That is
// the whole of #568: Fact or Fiction's piles are separated by an
// opponent, Torment's choice is made by each opponent, and the engine
// had no prompt shape that could be pointed at another player without
// being a mana payment (pay_unless) or a trigger's own "you may".
//
// FromPlayer names whose material the options are about, and it is
// load-bearing rather than decorative: protocol's redaction pass reads
// it to decide what the CHOOSER is entitled to see when the pool is not
// theirs (see redactChoiceCards in protocol/view.go).
//
// # Two shapes, one kind
//
// An option may carry cards. That is what makes the second half of a
// pile split — "put one pile into your hand and the other into your
// graveyard" — the same prompt as Torment's three-way question rather
// than a kind of its own: two options, each holding a pile, addressed
// to the controller. The first half of a pile split is an ordinary
// PendingChoiceChooseCards addressed to the splitter, and the two are
// joined by #552's chaining, which is why "pile split" needs no kind
// at all.

// PendingChoiceOptionPick is "choose one of the following", addressed
// to any seat, answered with the INDEX of the chosen option.
//
// Answered with {option_index: N}, routed by the choice's KIND rather
// than by the field's presence — the meaningful value of the payload
// is zero (the first option), so there is nothing to route on. The
// same reason loop_shortcut is routed by kind.
//
// Every offered option is accepted: ResolveOptionPick validates the
// index and nothing else, exactly as ResolveConfirm validates nothing
// about the board. The legality lives at QUEUE time — an effect builds
// the option list out of what the chooser can actually do, dropping
// "sacrifice a nonland permanent" for a player who controls none — and
// must offer at least one option that always works, which is the one
// the enumerator marks AlwaysLegal. A prompt whose every branch could
// fail is a prompt a seat can be stuck on (#544).
const PendingChoiceOptionPick PendingChoiceKind = "option_pick"

// NoChoiceIndex is the index a frame's continuation is run with when
// the question could not be put, or could not be kept up: "nobody
// chose". Not an answer a client may send — ResolveOptionPick refuses
// a negative index — and not an offset into anything.
//
// It is the #544 rule in one value. A continuation is the rest of the
// card, so an option pick that ends without an answer still has to run
// it; what it must not do is silently pick a branch on the chooser's
// behalf. Every continuation the engine queues already handles it,
// because the card-side primitive has always had to
// (effects.PickOption's Then is documented as "the index of the chosen
// option, or -1 when no question could be asked"): a pile pick takes
// the first pile, Torment of Hailfire moves on to the next victim, a
// player choice records the absence.
//
// Reached from two places, and #1006 is the second. The first is QUEUE
// time — an empty option list, or a chooser who has already left, so
// nothing is queued and the caller runs its own continuation. The
// second is DROP time: a prompt the engine withdraws unanswered runs
// the frame with this (dropDefault, leave_game.go).
//
// A SEAT continuation (optionPickFrame.thenSeat, #994) spells the same
// absence as uuid.Nil rather than as an index, because that is the
// currency it speaks; runWithNoChoice is the one place both are
// written.
const NoChoiceIndex = -1

// ChoiceOption is one branch of a PendingChoiceOptionPick: the card's
// own words for it, the cards it is about (a pile, or nothing), and
// what it costs the chooser in life.
//
// It is plain data and goes on the wire. What the option MEANS is the
// frame's business, not this struct's — the index comes back and the
// queuing effect decides, which is what keeps the kind reusable
// instead of being a second modal-spell system.
type ChoiceOption struct {
	// Label is the option's printed sentence — "Lose 3 life",
	// "Sacrifice a nonland permanent", "Put this pile into your
	// hand". Required: it is the whole of what the chooser reads.
	Label string

	// Cards are the cards this option is about, in render order.
	// Empty for an option that names no cards, which is most of them.
	// Projected per viewer through the same redaction pass as every
	// other card list on a prompt, so an option over a pool the
	// chooser does not own shows only what that seat may legally see.
	Cards []uuid.UUID

	// LifeCost is the life this branch charges, for the wire and for
	// a bot's pricing — the same declaration ConfirmPrompt.LifeCost
	// makes, and for the same #547 reason: a policy holding only the
	// wire payload otherwise prices "lose 3 life" exactly like
	// "discard a card", and a bot at 3 life picks the life and dies.
	//
	// The engine does NOT deduct it. The branch does.
	LifeCost int

	// Player is the SEAT this option is about, for an option list
	// whose branches are players rather than consequences — every
	// prompt built by choose_player.go. uuid.Nil on every other
	// option, which is all of them (a pile, a Torment branch): the
	// zero value means "this option is not about a seat" and nothing
	// reads it.
	//
	// It exists because an option has to be able to name its SUBJECT
	// (#994). The engine re-checks an open prompt whenever the board
	// moves under it — reassignChoiceLocked prunes the cards a
	// departed player took with them — and it could do that only for
	// options built out of cards. A seat option carried its seat as
	// rendered TEXT in Label, so a departed seat was identifiable
	// only by string-matching a player's name, and a prompt went on
	// offering a player who was no longer a player (CR 800.4a).
	// With the id here, seats prune exactly the way cards do:
	// pruneDepartedSeatOptionsLocked and reassignChoiceLocked are
	// the two paths, and both key on this field.
	//
	// Label is still what the chooser READS; this is what the engine
	// checks. The two are set together by one function
	// (seatChoiceOptionsLocked) so they cannot disagree.
	Player uuid.UUID
}

// cloneChoiceOptions deep-copies an option list. Each option owns a
// card slice, so a shallow copy would let an undo snapshot and the
// live game share a backing array — the same rule every other slice
// on a PendingChoice follows (clone.go).
func cloneChoiceOptions(in []ChoiceOption) []ChoiceOption {
	if len(in) == 0 {
		return nil
	}
	out := make([]ChoiceOption, len(in))
	for i, opt := range in {
		out[i] = opt
		out[i].Cards = copyUUIDs(opt.Cards)
	}
	return out
}

// optionPickFrame is the continuation behind a
// PendingChoiceOptionPick: ONE closure, given the index the chooser
// picked.
//
// One closure rather than one per option, because the branches of a
// three-way question are nearly always written as one switch in the
// card's own package-level function — and because a slice of closures
// would make the continuation census count a number that varies with
// the card rather than with the prompt.
//
// It receives the live *Game (not a captured one) on the same
// undo-safety contract confirmFrame and StackItem.Effect follow, and
// runs with g.mu held — so it may queue the next link of a chain.
type optionPickFrame struct {
	then func(g *Game, index int) error

	// thenSeat is the continuation for an option list whose branches
	// are SEATS (choose_player.go). It receives the chosen option's
	// ChoiceOption.Player instead of its index, and exactly one of the
	// two is ever set.
	//
	// It is a second continuation rather than a `then` that closes over
	// the seat list because an INDEX into a captured slice is not a
	// stable answer any more (#994). The option list of a player prompt
	// is pruned while it is open — a seat that leaves the game comes off
	// it (CR 800.4a) — so index 2 before the prune and index 2 after it
	// are different players, and a closure holding the original slice
	// would record the wrong one. Reading the seat off the option the
	// chooser actually picked cannot drift from what they were shown,
	// because it IS what they were shown.
	thenSeat func(g *Game, seat uuid.UUID) error
}

// runWithNoChoice runs the frame as though nobody chose. Reports the
// error the continuation returned, for the caller to report the way it
// reports every other continuation failure.
//
// The one place "nobody chose" is handed to a frame, so what a question
// that ended without an answer does is one line rather than one line
// per drop path — and so the two continuation shapes cannot drift: an
// index frame is run with NoChoiceIndex, a SEAT frame with uuid.Nil
// (#994's thenSeat), which is the same absence spelled in the currency
// each one speaks. Both are values their continuations already had to
// handle, and neither is an answer a client can send.
func (f *optionPickFrame) runWithNoChoice(g *Game) error {
	switch {
	case f == nil:
		return nil
	case f.thenSeat != nil:
		return f.thenSeat(g, uuid.Nil)
	case f.then != nil:
		return f.then(g, NoChoiceIndex)
	}
	return nil
}

// defaultDroppedChoiceLocked runs an option pick's continuation with
// the no-choice outcome because the prompt has been dropped rather
// than answered. It is the dropDefault action of the departure table
// (choiceDepartureDecisions, leave_game.go) and the CR 800.4f/#544
// pair applied to this kind: the question ends, and the rest of the
// card does not.
//
// Before #1006 the drop took the frame with it, and the effect that
// was paused mid-resolution never finished — Fact or Fiction put
// neither pile anywhere, a Torment of Hailfire stopped at the victim
// who left, a "choose a player" never ran the sentence printed after
// it. QueuePileSplitForEffect and QueueChoosePlayerForEffect were
// already careful about exactly this at QUEUE time; nothing was
// careful about it at drop time.
//
// This is the SAME continuation ResolveOptionPick runs for an answer,
// reached from a withdrawal rather than from one — the #808 shape, and
// the same one declineDepartedChoiceLocked is. It cannot re-queue a
// prompt to a departed seat (QueueChoiceForEffect refuses an
// eliminated chooser, #864), and it takes no branch on the chooser's
// behalf, which is why it needs no claim about whose material the
// options were.
//
// Caller must hold g.mu, and must already have taken the prompt out of
// the queue: the continuation may queue the next link of the chain and
// must not land behind the question it is replacing.
// #1019 adds the SECOND kind to declare dropDefault, and it is the
// same idea in a different currency: a withdrawn sacrifice prompt is
// a seat that sacrificed nothing, and the run waiting on it has to be
// told so — a run whose last leg was dropped silently is a card that
// stops halfway exactly as a dropped option pick was. The two are one
// function rather than two drop paths because the departure table has
// one second column, and a kind settled in two places is a kind
// settled two ways.
//
// Caller must hold g.mu, and must already have taken the prompt out of
// the queue: the continuation may queue the next link of the chain and
// must not land behind the question it is replacing.
func (g *Game) defaultDroppedChoiceLocked(c *PendingChoice) {
	if c == nil {
		return
	}
	var err error
	switch c.Kind {
	case PendingChoiceSacrifice:
		err = g.settleSacrificeRunLegLocked(c.sacrificeRun, c.Chooser, nil)
	default:
		err = c.optionPickResume.runWithNoChoice(g)
	}
	if err != nil {
		g.emitChoiceEffectErrorLocked(c.Chooser, c.Source, err)
	}
}

// OptionPickPrompt is the queue-side description of a
// PendingChoiceOptionPick.
type OptionPickPrompt struct {
	// Chooser answers the prompt. Required, and frequently NOT the
	// resolving effect's controller.
	Chooser uuid.UUID

	// FromPlayer owns the material the options are about. Defaults to
	// Chooser when empty. Set it when the options name cards the
	// chooser does not own — the redaction pass reads it.
	FromPlayer uuid.UUID

	// Source is the card asking.
	Source uuid.UUID

	// Question is the prompt's header — the card's own sentence.
	Question string

	// Options are the branches, in the order the card prints them.
	// The FIRST is the one the enumerator marks always-legal, so it
	// must be a branch that can always be taken.
	Options []ChoiceOption

	// Then receives the index of the chosen option, or NoChoiceIndex
	// when the prompt ended without an answer — dropped because its
	// chooser left the game or because it had no legal answer left
	// (#1006). Runs with g.mu held; may queue further choices, which
	// is how a chain continues.
	//
	// HANDLE NoChoiceIndex. It is the rest of the card running with
	// the question unanswered, and a continuation that indexes
	// straight into a captured slice with it is a card that stops
	// halfway — which is the #544 rule this kind is otherwise careful
	// about.
	Then func(g *Game, index int) error

	// ThenSeat is Then for an option list whose branches are SEATS:
	// it receives the chosen option's Player rather than its index.
	// Set by choose_player.go and by nothing else; a caller that sets
	// both gets this one.
	//
	// Use it for any option list built out of players. An index is not
	// a stable answer for one: a seat that leaves the game is pruned
	// off an OPEN prompt (#994, CR 800.4a), which renumbers everything
	// after it. See optionPickFrame.thenSeat.
	ThenSeat func(g *Game, seat uuid.UUID) error
}

// QueueOptionPickForEffect queues a "choose one of the following" and
// returns its ID, or uuid.Nil when it queued nothing — an empty option
// list, or a chooser who has left the game (QueueChoiceForEffect's own
// CR 800.4a guard).
//
// A caller with more of the effect to run after the answer must check
// the return: nothing was queued means nothing will call Then, and a
// chain that assumed otherwise stops halfway. That is the same
// contract QueueDiscardChoiceForEffect documents, pointed the other
// way — a discard with nothing to pitch still runs Then, because CR
// 701.8a says "as many as you can"; a question with no answers was
// never asked at all.
//
// Caller must hold g.mu.
func (g *Game) QueueOptionPickForEffect(p OptionPickPrompt) uuid.UUID {
	if len(p.Options) == 0 {
		return uuid.Nil
	}
	from := p.FromPlayer
	if from == uuid.Nil {
		from = p.Chooser
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:        PendingChoiceOptionPick,
		Chooser:     p.Chooser,
		FromPlayer:  from,
		Count:       1,
		Source:      p.Source,
		Reason:      p.Question,
		PickOptions: cloneChoiceOptions(p.Options),
		optionPickResume: &optionPickFrame{
			then:     p.Then,
			thenSeat: p.ThenSeat,
		},
	})
}

// ResolveOptionPick answers a PendingChoiceOptionPick: `index` is the
// offset into the prompt's option list.
//
// The choice is dequeued BEFORE the branch runs, for chaining's sake —
// the branch queues the next prompt, and a queue that still held this
// one would hand the client two open modals and the enumerator two
// owed choices for the same seat (ResolveConfirm's reason, verbatim).
//
// An out-of-range index is refused with the prompt still open, so a
// client that sends a stale index can try again. A branch that ERRORS
// emits EventEffectError and the prompt is still gone: a failed
// continuation must not leave an unanswerable prompt in the queue,
// which is the #544 shape.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveOptionPick(choiceID, chooserID uuid.UUID, index int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceOptionPick {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if index < 0 || index >= len(choice.PickOptions) {
		return ErrInvalidParam
	}
	frame := choice.optionPickResume
	source := choice.Source
	// Read off the option the chooser actually picked, BEFORE the
	// dequeue takes the list away. A seat continuation is answered with
	// this rather than with the index (#994): the list can have been
	// pruned since it was built, so the index is only meaningful
	// against the list the chooser was shown, which is this one.
	seat := choice.PickOptions[index].Player
	if frame != nil && frame.thenSeat != nil && seat == uuid.Nil {
		// A seat continuation answered with an option that names no
		// seat is a bug in whoever built the list, and it must not
		// reach the frame: uuid.Nil is how a DROP says nobody was
		// chosen (#1006, runWithNoChoice), so passing it here would
		// report an answer as an absence. Refused before the dequeue,
		// like an out-of-range index, so the prompt stays open.
		return ErrInvalidParam
	}
	g.dequeueChoiceLocked(idx)
	if frame == nil || (frame.then == nil && frame.thenSeat == nil) {
		// Nothing to run. The prompt is gone either way rather than
		// stuck: an option pick with no frame is a bug in whoever
		// queued it, and refusing the answer would wedge the seat.
		return nil
	}
	run := frame.then
	if frame.thenSeat != nil {
		run = func(g *Game, _ int) error { return frame.thenSeat(g, seat) }
	}
	if err := run(g, index); err != nil {
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

// QueuePileSplitForEffect is the two-prompt "separate these cards into
// two piles; somebody else takes one of them" shape (CR 608.2), built
// out of the two kinds above rather than a kind of its own.
//
// Fact or Fiction is the card: "Reveal the top five cards of your
// library. An opponent separates those cards into two piles. Put one
// pile into your hand and the other into your graveyard."
//
// Two chained prompts to two different seats (#552):
//
//  1. PendingChoiceChooseCards to the SPLITTER — "choose the cards for
//     the first pile", floor zero, ceiling all of them. An empty pile
//     is legal and is frequently the right split — "two piles" does
//     not require both to hold a card, and splitting 5-0 is a real
//     Fact or Fiction play.
//  2. PendingChoiceOptionPick to the CHOOSER — two options, each
//     carrying its pile's cards, so the client renders the two piles
//     rather than quoting names into a sentence.
//
// `then` receives the pile the chooser TOOK first and the other one
// second, both in the order they were split. A splitter who has left
// the game, or a card list with nothing in it, runs `then` with
// everything in the first pile and nothing in the second: the reveal
// still happened and the rest of the card still has to resolve.
//
// Caller must hold g.mu.
func (g *Game) QueuePileSplitForEffect(p PileSplitPrompt) {
	if len(p.Cards) == 0 || p.Then == nil {
		if p.Then != nil {
			if err := p.Then(g, nil, nil); err != nil {
				g.emitChoiceEffectErrorLocked(p.Chooser, p.Source, err)
			}
		}
		return
	}
	cards := copyUUIDs(p.Cards)
	chooser := p.Chooser
	owner := p.Owner
	if owner == uuid.Nil {
		owner = chooser
	}
	source := p.Source
	pickQuestion := p.PickQuestion
	then := p.Then
	queued := g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:    p.Splitter,
		FromPlayer: owner,
		Source:     source,
		Question:   p.SplitQuestion,
		Cards:      cards,
		Min:        0,
		Max:        len(cards),
		// No Zone re-check. The cards are revealed where they sit —
		// Fact or Fiction's five are still on top of a library — and
		// the piles are a partition of what was revealed, not a claim
		// about where anything is now. The `then` that acts on them
		// re-checks each card the way every other effect does.
		Then: func(g *Game, picked []uuid.UUID) error {
			return g.queuePilePickLocked(chooser, owner, source, pickQuestion, cards, picked, then)
		},
	})
	if queued == uuid.Nil {
		// The splitter has left the game (CR 800.4a). Nobody can
		// separate the cards, so the whole reveal is one pile and the
		// rest of the card resolves against it.
		if err := then(g, cards, nil); err != nil {
			g.emitChoiceEffectErrorLocked(chooser, source, err)
		}
	}
}

// PileSplitPrompt is the queue-side description of a pile split.
type PileSplitPrompt struct {
	// Splitter separates the cards — the opponent, on every printed
	// card of this family. Required.
	Splitter uuid.UUID
	// Chooser takes one of the two piles. Required; the resolving
	// effect's controller on every printed card of this family.
	Chooser uuid.UUID
	// Owner owns the cards. Defaults to Chooser. Read by the
	// redaction pass to decide what the splitter may see.
	Owner uuid.UUID
	// Source is the card asking.
	Source uuid.UUID
	// SplitQuestion / PickQuestion are the two prompts' headers.
	SplitQuestion, PickQuestion string
	// Cards are the cards being separated, in the order the table saw
	// them revealed.
	Cards []uuid.UUID
	// Then receives the pile the chooser took and the pile they left,
	// in that order. Runs with g.mu held.
	Then func(g *Game, taken, left []uuid.UUID) error
}

// queuePilePickLocked is the second link of a pile split: the chooser
// takes one of the two piles.
//
// A package-level-shaped continuation reading only scalars and frozen
// slices, for the reason delayed triggers give — the closure has to
// resolve against whichever *Game it is handed, which after an undo is
// the restored one and not the one that queued it.
//
// Caller must hold g.mu.
func (g *Game) queuePilePickLocked(
	chooser, owner, source uuid.UUID,
	question string,
	all, first []uuid.UUID,
	then func(g *Game, taken, left []uuid.UUID) error,
) error {
	pileOne, pileTwo := partitionPiles(all, first)
	queued := g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:    chooser,
		FromPlayer: owner,
		Source:     source,
		Question:   question,
		Options: []ChoiceOption{
			{Label: pileLabel(1, pileOne), Cards: pileOne},
			{Label: pileLabel(2, pileTwo), Cards: pileTwo},
		},
		Then: func(g *Game, index int) error {
			if index == 1 {
				return then(g, pileTwo, pileOne)
			}
			return then(g, pileOne, pileTwo)
		},
	})
	if queued != uuid.Nil {
		return nil
	}
	// The chooser has left the game between the split and the pick.
	// Take the first pile so the rest of the card still resolves.
	return then(g, pileOne, pileTwo)
}

// partitionPiles splits `all` into the cards named by `first` and the
// rest, each in `all`'s order. A card named in `first` that is not in
// `all` is ignored — the answer was validated against the candidate
// list before it got here, and this function is the one place the two
// piles are derived, so it does not get to disagree with that.
func partitionPiles(all, first []uuid.UUID) (one, two []uuid.UUID) {
	inFirst := make(map[uuid.UUID]bool, len(first))
	for _, id := range first {
		inFirst[id] = true
	}
	for _, id := range all {
		if inFirst[id] {
			one = append(one, id)
			continue
		}
		two = append(two, id)
	}
	return one, two
}

// pileLabel names a pile for the option button: "Pile 1 (3 cards)".
// The COUNT is on the label because the client renders the pile's
// cards beside it and a zero-card pile would otherwise be a button
// with nothing under it — an empty pile is a legal split and has to
// look deliberate rather than broken.
func pileLabel(n int, cards []uuid.UUID) string {
	switch len(cards) {
	case 0:
		return "Take pile " + strconv.Itoa(n) + " (empty)"
	case 1:
		return "Take pile " + strconv.Itoa(n) + " (1 card)"
	default:
		return "Take pile " + strconv.Itoa(n) + " (" + strconv.Itoa(len(cards)) + " cards)"
	}
}

// emitChoiceEffectErrorLocked reports a continuation that failed while
// a prompt was being answered or skipped. Every resolver in the queue
// does this by hand; the pile split does it from three places, so it
// is one function here rather than three copies.
//
// Caller must hold g.mu.
func (g *Game) emitChoiceEffectErrorLocked(actor, source uuid.UUID, err error) {
	if err == nil {
		return
	}
	g.EmitEvent(Event{
		Kind:     EventEffectError,
		Actor:    actor,
		Source:   source,
		ErrorMsg: err.Error(),
	})
}
