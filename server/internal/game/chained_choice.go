package game

import (
	"slices"

	"github.com/google/uuid"
)

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

// isCardSetPickKind reports whether a prompt carries the choose-cards
// payload — ChooseCards / ChooseMin / ChooseMax.
//
// EIGHT kinds do. This one; CR 502.3's untap_choice (ADR 0070
// Decision 2); #1198's entry_reveal_from_hand (ADR 0013 §5z) and
// ADR 0098's entry_discard_from_hand and entry_sacrifice; and
// #1214's three resolution-time picks (resolution_pick.go). Each
// shares the SHAPE and differs in what the player is being asked, in
// how a bot should score the answer, and — for the three picks — in
// what a seat may SEE of the pool and who inherits the question when
// the chooser leaves.
//
// Answering that with one predicate is the point: every rule about the
// PAYLOAD — the bounds check, the candidate check, the live-zone
// re-check, #1017's set-level Validate, the two prunes that keep an
// open prompt honest — is written once and asks this, while every rule
// about the QUESTION is a row in one of the two tables keyed by kind.
//
// What they do NOT all share is the RESOLVER: entry_reveal_from_hand's
// continuation is a paused replacement event rather than a card's next
// sentence, so it (and ADR 0098's entry_discard_from_hand and
// entry_sacrifice) settles through ResolveEntryCardChoice
// instead of resolveCardSetPick.
func isCardSetPickKind(kind PendingChoiceKind) bool {
	switch kind {
	case PendingChoiceChooseCards, PendingChoiceUntapChoice, PendingChoiceEntryRevealFromHand,
		PendingChoiceEntryDiscardFromHand, PendingChoiceEntrySacrifice, PendingChoiceEntryChoosePermanents, PendingChoiceChooseSource,
		PendingChoiceRingBearer, PendingChoiceProliferate:
		return true
	}
	return isResolutionPickKind(kind)
}

// IsCardSetPickKind is isCardSetPickKind for callers outside this
// package: `protocol` projects the same three fields for every kind
// that carries them, and a hand-kept list there would be the second
// copy this predicate exists to prevent (IsLookAtTopKind's reason).
func IsCardSetPickKind(kind PendingChoiceKind) bool { return isCardSetPickKind(kind) }

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

// ConfirmPrompt is the queue-side description of a
// PendingChoiceConfirm. A struct rather than eight positional
// arguments, because the labels and the two branches are all optional
// in different combinations.
type ConfirmPrompt struct {
	// Chooser answers the prompt. Required.
	Chooser uuid.UUID
	// FromPlayer names whose material the question is ABOUT, the way
	// it already does on a ChooseCardsPrompt. Defaults to Chooser,
	// which is what almost every confirm is — "you may shuffle", "pay
	// 2 life?" — and a cross-table yes/no (Combustible Gearhulk asks
	// its target whether to mill) sets it to the player it is asked
	// of. The difference is invisible until the chooser leaves the
	// game: CR 800.4g reassigns a question about somebody else's
	// material and drops one about the asker's own, and this field is
	// how the departure sweep tells them apart (#961, the one field
	// ADR 0060's amendment named).
	FromPlayer uuid.UUID
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
	// DeclineEndsLoop is PendingChoice.DeclineEndsLoop (#2040): accepting
	// would continue a loop the table is already in.
	DeclineEndsLoop bool
	// Asked, Word and Objects are what a confirm hands its answer beside
	// the carried IDs (QueueConfirmKeyedForEffect, #4291): the asking stack
	// item as data (NewAskedItem), the one word of the card's text, and the
	// objects with the epoch each had. The branches are a registered answer
	// (RegisterConfirmThen); the closure pair and its door,
	// QueueConfirmForEffect, were deleted in #4253's last part once only
	// tests used them.
	Asked   *AskedItem
	Word    string
	Objects []ObjectRef
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
	// Zone is the zone a collector's leg picks from: a choose_cards run
	// (ChooseCardsRunThenKeyedForEffect) reads it as the leg's LiveZone,
	// and an ordered collection (ChoosePermanentsInOrderKeyedForEffect)
	// refuses any but the battlefield. QueueChooseCardsForEffect does not
	// read it: a pick's live re-check is LiveZone (#4361 S7b, the closure
	// pick that read it is gone).
	Zone ZoneKind
	// The closure continuation and set rule (Then, Validate) are gone
	// (#4361 S7b): every pick names a REGISTERED continuation (ThenKey) and,
	// when it needs one, a set rule registered with it
	// (RegisterChooseCardsThenRule), so a table waiting on any choose_cards
	// pick is a restore point.
	// Destination is where the named cards go, when the clause puts
	// them somewhere a chooser wants them (#2680): PendingChoice's
	// ChooseDestination. Zero declares nothing.
	Destination ChooseDestination

	// ThenKey is Then as a REGISTERED continuation
	// (RegisterChooseCardsThen, choose_cards_keyed.go), and Carry the
	// plain data it is handed with the answer (#2040). A prompt queued
	// with a key holds no closure, so a table waiting on it is still a
	// restore point. Without one (#4361 S7b) the pick runs the engine's
	// continuation that does nothing (choose-cards/nothing), with Zone as
	// its live zone.
	ThenKey ChooseCardsThen
	Carry   []uuid.UUID
	// LiveZone opts a pick into the live-zone re-check (#4291): the pick
	// is refused for, and withdrawn when, its candidates are no longer in
	// that zone, and the withdrawal runs the continuation with nothing
	// picked.
	LiveZone ZoneKind
	// CarryWord is the one word of the card's text a keyed continuation
	// is handed (CardsChosen.Word, #4264): plain data on the prompt
	// (PendingChoice.PickMatchArg).
	CarryWord string
	// CarryNums are numbers a keyed continuation is handed
	// (CardsChosen.Nums, #4294): a round of a repeated process, in
	// PendingChoice.PickNums. Asked is the asking stack item as data
	// (NewAskedItem), handed to it as CardsChosen.Asked. Both are ignored
	// without ThenKey.
	CarryNums []int
	Asked     *AskedItem
	// Tail is a registered tail the continuation hands on (CardsChosen.Tail,
	// #4253 slice 8): the rest of a keyword action this pick is the middle
	// of. Ignored without ThenKey.
	Tail KeyedTail

	// promptRun links the queued prompt to the RUN it is one leg of
	// (PendingChoice.promptRun, prompt_run.go). Unexported because it
	// is engine plumbing: two callers set it, the discard prompt
	// (#1027) and ChooseCardsRunThenKeyedForEffect (#1743), and the catalog
	// never builds a run by hand.
	promptRun uuid.UUID
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
	key := p.ThenKey.key
	live := p.LiveZone
	if key == "" {
		// #4361 S7b: the closure pick is gone. A pick with nothing hanging
		// off the answer runs the engine's continuation that does nothing,
		// re-checked against its Zone as the closure pick was.
		key = chooseCardsNothing.key
		if live == "" {
			live = p.Zone
		}
	}
	// #2040: a keyed pick carries no frame, so it stays a restore point
	// (choose_cards_keyed.go).
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:              PendingChoiceChooseCards,
		Chooser:           p.Chooser,
		FromPlayer:        from,
		Count:             hi,
		Source:            p.Source,
		Reason:            p.Question,
		ChooseCards:       append([]uuid.UUID(nil), p.Cards...),
		ChooseMin:         lo,
		ChooseMax:         hi,
		ChooseDestination: p.Destination,
		PickThen:          key,
		PickMatchArg:      p.CarryWord,
		PickNums:          copyInts(p.CarryNums),
		Asked:             copyAskedItem(p.Asked),
		pickTail:          copyKeyedTail(p.Tail),
		OptionCarry:       copyUUIDs(p.Carry),
		ChooseLiveZone:    live,
		// #4290: a keyed leg of a prompted run.
		promptRun: p.promptRun,
	})
}

// chooseCardsNothing is the continuation of a card-set pick with nothing
// hanging off the answer (#4361 S7b): an unkeyed ChooseCardsPrompt, or an
// unkeyed pick over a sideboard. An on-disk key: never renamed, never
// reused.
var chooseCardsNothing = RegisterChooseCardsThen("choose-cards/nothing", func(*Game, CardsChosen) error { return nil })

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
	g.answerConfirmLocked(idx, choice, chooserID, accept, false)
	return nil
}

// answerConfirmLocked is ResolveConfirm's body, shared with the
// automatic answer (ADR 0127 §4): `auto` takes the prompt out of the
// queue without counting it as a player decision. The branch runs as the
// answered prompt's branch, so a prompt it queues is keyed from this
// one's (§2). Caller must hold g.mu.
func (g *Game) answerConfirmLocked(idx int, choice *PendingChoice, chooserID uuid.UUID, accept, auto bool) {
	frame := confirmFrameOf(choice)
	source := choice.Source
	g.answerChoiceLocked(idx, auto)
	if frame == nil {
		// Nothing to run. The prompt is gone either way rather than
		// stuck: a confirm with no frame is a bug in whoever queued
		// it, and refusing the answer would wedge the seat.
		return
	}
	branch := frame.onDecline
	if accept {
		branch = frame.onAccept
	}
	if branch != nil {
		g.runPromptBranchLocked(choice.AutoAnswerKey, func() {
			if err := branch(g); err != nil {
				g.EmitEvent(Event{
					Kind:     EventEffectError,
					Actor:    chooserID,
					Source:   source,
					ErrorMsg: err.Error(),
				})
			}
		})
	}
	g.runStateChecksLocked()
}

// ResolveChooseCards answers a PendingChoiceChooseCards: `picks` are
// the cards the chooser selected.
//
// Validation rejects before dequeuing, so a client that submits an
// illegal set gets an error and can try again rather than losing the
// prompt — the same contract ResolveSearchLibrary documents. A set the
// prompt's Validate hook refuses comes back as ErrChoiceSetRejected,
// with the prompt still open and Then not run. The #544 lesson is that
// every answer the enumerator offers passes this validation: the bounds
// ride on the choice, and the set-level hook is reached through
// ChooseCardsPickLegalLocked, so `legal` asks the engine instead of
// enumerating a superset it cannot see the constraint on.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveChooseCards(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	return g.resolveCardSetPick(PendingChoiceChooseCards, choiceID, chooserID, picks)
}

// resolveCardSetPick is ResolveChooseCards' body, shared with the ONE
// other kind that carries a choose-cards payload and a choose-cards
// continuation: PendingChoiceUntapChoice, CR 502.3's determination
// (untap_choice.go). `kind` is the kind the caller is answering, so a
// client cannot answer one prompt through the other's verb.
//
// One body rather than two, for the reason checkChooseCardsPicksLocked
// is one function: the validation, the dequeue-before-the-branch order
// and the EventEffectError contract are all rules about the QUEUE, not
// about the card, and a second copy of them would drift.
//
// Caller must NOT hold g.mu.
func (g *Game) resolveCardSetPick(kind PendingChoiceKind, choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != kind {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if err := g.checkChooseCardsPicksLocked(choice, picks); err != nil {
		return err
	}
	if isKeyedChooseCards(choice) {
		// #2040: a keyed pick's continuation, looked up by its key.
		keyed := *choice
		g.dequeueChoiceLocked(idx)
		if keyed.Kind == PendingChoiceRevealPick {
			// A run's leg narrates the pick (settlePickLegLocked); a
			// keyed one does too.
			g.emitCardsChosenLocked(chooserID, keyed.Source, picks)
		}
		g.emitChoiceEffectErrorLocked(chooserID, keyed.Source, g.runKeyedChooseCardsLocked(&keyed, picks, true))
		g.runStateChecksLocked()
		return nil
	}
	leg := *choice
	g.dequeueChoiceLocked(idx)
	if !isRunLegPickKind(leg.Kind) {
		return nil
	}
	// #4361 S7b: a reveal pick or a permanent pick asked as one leg of a
	// prompted run (or of none, when nothing waits on it) settles from its
	// own fields, so it holds no closure. #4281: answered under the cause
	// a keyed branch carried, as the closure frame was.
	g.emitChoiceEffectErrorLocked(chooserID, leg.Source, g.withCarriedCauseLocked(carriedCauseOfChoice(&leg), func() error {
		return g.settlePickLegLocked(&leg, picks)
	}))
	g.runStateChecksLocked()
	return nil
}

// isRunLegPickKind is the card-set kinds a prompted run asks as its legs with
// no continuation of their own: a reveal pick (queueRevealPickLocked) and the
// two permanent picks (queuePermanentPickLocked).
func isRunLegPickKind(k PendingChoiceKind) bool {
	return k == PendingChoiceRevealPick || isPermanentPickKind(k)
}

// settlePickLegLocked narrates a run leg's pick and settles the leg (#4361
// S7b, the closure frames revealPickLegThen and permanentPickLegThen were):
// a reveal pick's leg is keyed by its chooser, a permanent pick's by whose
// permanents were chosen (FromPlayer, see PromptedPicks). A pick no run
// waits on (promptRun nil) only narrates.
//
// Caller must hold g.mu.
func (g *Game) settlePickLegLocked(c *PendingChoice, picked []uuid.UUID) error {
	g.emitCardsChosenLocked(c.Chooser, c.Source, picked)
	seat := c.Chooser
	if isPermanentPickKind(c.Kind) {
		seat = c.FromPlayer
	}
	return g.settleRunLegLocked(c.promptRun, seat, append([]uuid.UUID(nil), picked...))
}

// checkChooseCardsPicksLocked is the ONE copy of "would this answer be
// accepted for this choose-cards prompt": bounds, candidates,
// duplicates, the live-zone re-check, then the set-level Validate hook.
// ResolveChooseCards runs it before it dequeues; internal/legal runs it
// through ChooseCardsPickLegalLocked before it OFFERS a set. One
// function rather than two for checkSearchPicksLocked's reason — the
// hook is card text, and a second copy of card text drifts.
//
// It reads and never writes, so it is safe under the read lock.
//
// Caller must hold g.mu (read or write).
func (g *Game) checkChooseCardsPicksLocked(choice *PendingChoice, picks []uuid.UUID) error {
	if choice == nil || !isCardSetPickKind(choice.Kind) {
		return ErrInvalidParam
	}
	if len(picks) < choice.ChooseMin || len(picks) > choice.ChooseMax {
		return ErrInvalidParam
	}
	candidates := make(map[uuid.UUID]bool, len(choice.ChooseCards))
	for _, id := range choice.ChooseCards {
		candidates[id] = true
	}
	// A proliferate names seats as well as permanents (CR 701.34a).
	for _, id := range choice.ChoosePlayers {
		candidates[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(picks))
	for _, id := range picks {
		if !candidates[id] || seen[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		if zone := choice.liveZone(); zone != "" {
			// Re-checked against the LIVE zone, not the frozen
			// candidate list: the prompt is asynchronous and a card
			// can leave between the question and the answer.
			if !g.pickStillInPickZoneLocked(choice, zone, id) {
				return ErrCardNotFound
			}
		}
	}
	// Skipped for an empty pick, so a zero-floor prompt's "choose
	// nothing" stays the answer nothing can refuse (SetRuleFunc).
	//
	// The rule is the one registered with the pick's continuation (#4264),
	// or, for a permanent pick, with its run (#4290).
	var validate func([]Card) bool
	outside := false
	switch {
	case isPermanentPickKind(choice.Kind):
		// #4361 S7b: a keyed run's set rule, read off its key.
		if choice.PickThen != "" {
			validate = lookupPromptRunRule(choice.PickThen)
		}
	case choice.Kind == PendingChoiceUntapChoice && choice.PickThen != "":
		// #4361 S7a: CR 502.3's determination checks the plan it
		// carries as data (untap_choice.go).
		validate = untapChoiceRule(choice)
	case isEntryCardChoiceKind(choice.Kind):
		// The entry choice's own declaration narrows its candidates; it
		// has no set rule.
	case choice.PickThen != "":
		if rule := lookupChooseCardsRule(choice.PickThen); rule != nil {
			// A keyed pick has no zone of its own, so a card the table
			// cannot find is read from the sideboards too.
			validate, outside = rule, true
		}
	}
	if validate == nil || len(picks) == 0 {
		return nil
	}
	chosen := make([]Card, 0, len(picks))
	for _, id := range picks {
		card, ok := g.LookupCardForEffect(id)
		if !ok && outside {
			card, ok = g.LookupOutsideCardForEffect(id)
		}
		if !ok {
			return ErrCardNotFound
		}
		chosen = append(chosen, card)
	}
	if !validate(chosen) {
		return ErrChoiceSetRejected
	}
	return nil
}

// pickStillInPickZoneLocked is the live-zone re-check above on ONE
// card: is `id` still in the zone `choice` picks from?
//
// Split out in #1045 because the ENGINE has to ask it too, of the
// candidates rather than of the answer: pruneCardSetChoicesLocked
// trims an open prompt to the cards a submitted answer would still be
// accepted for. Two readings of "is this candidate still there" is the
// pair that drifts, and the drift is a prompt offering an answer its
// own resolver refuses — which for a floor of one is the #544 wedge.
//
// Caller must hold g.mu (read or write).
func (g *Game) pickStillInPickZoneLocked(choice *PendingChoice, zone ZoneKind, id uuid.UUID) bool {
	if o := choice.AnyPlayerOffer; o != nil && len(o.Objects) > 0 && choice.Kind == PendingChoiceChooseCards {
		// #4361 S7b: a per-object any-player leg binds each object to the
		// object it was when the offer was made.
		c := findBattlefieldCard(g, id)
		if c == nil || !slices.ContainsFunc(o.Objects, func(r ObjectRef) bool { return r.ID == id && r.Epoch == c.ObjectEpoch }) {
			return false
		}
	}
	if o := choice.ordered; o != nil && o.Next >= 0 && o.Next < len(o.Legs) {
		// #4290: a keyed ordered collection binds each candidate to the
		// object it was when the collection began.
		c := findBattlefieldCard(g, id)
		if c == nil || !slices.ContainsFunc(o.Legs[o.Next].Cards, func(r ObjectRef) bool { return r.ID == id && r.Epoch == c.ObjectEpoch }) {
			return false
		}
	}
	if zone == ZoneOutside {
		// ADR 0191: a pick over the chooser's sideboard. No indexed
		// zone holds the card (CR 400.11c), so ask the container.
		return g.pickStillOutsideLocked(choice, id)
	}
	z := g.findCardZoneLocked(id)
	if z == nil || z.Kind != zone {
		return false
	}
	// Per-player zones carry an owner; the shared ones (battlefield,
	// exile, stack) leave it nil.
	return z.Owner == uuid.Nil || z.Owner == choice.FromPlayer
}

// ChooseCardsPickLegalLocked reports whether `picks` is an answer
// ResolveChooseCards would accept for `choice` right now. It is the
// enumerator's read-only window onto the prompt's Validate hook, which
// rides on the unexported continuation frame and stays there for the
// reason SearchPickLegalLocked gives: handing internal/legal the frame
// would make every future continuation field part of the enumerator's
// contract.
//
// Caller must hold g's read lock; internal/legal calls it from inside
// ReadSnapshot, like SearchPickLegalLocked.
func (g *Game) ChooseCardsPickLegalLocked(choice *PendingChoice, picks []uuid.UUID) bool {
	return g.checkChooseCardsPicksLocked(choice, picks) == nil
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

// ChooseDestination is where a choose_cards prompt sends the cards it
// names (#2680, PendingChoice.ChooseDestination). An open vocabulary
// read only by the bot and the client: an unknown value restores as
// itself and means nothing to either.
type ChooseDestination string

const (
	// ChooseOntoBattlefield: the named card is put onto the
	// battlefield ("you may put a land card from your hand onto the
	// battlefield").
	ChooseOntoBattlefield ChooseDestination = "battlefield"
	// ChooseOntoBattlefieldTapped: the same, tapped.
	ChooseOntoBattlefieldTapped ChooseDestination = "battlefield_tapped"
	// ChooseCastFree: the named hand card is cast without paying its
	// mana cost, as the ability resolves (CR 608.2g, #3270). The
	// continuation grants the pass-closed free cast
	// (FreeCastFromHandPermission); the client opens the cast chain
	// the moment the pick is submitted, and the bot prices the pick
	// as the cast it is rather than as a discard.
	ChooseCastFree ChooseDestination = "cast_free"
	// ChooseCastFreeFromLibrary: the same free cast, over a card the
	// chooser looked at on top of their library ("look at the top six
	// cards of your library. You may cast a spell from among them
	// without paying its mana cost", Aetherworks Marvel, #2110). The
	// continuation puts the rest on the bottom in a random order and
	// grants FreeCastFromLibraryPermission over the pick, which stays
	// where it is. The client opens the cast chain from the library
	// pile the moment the pick is submitted; the bot prices the pick
	// as it prices any card taken from its own library.
	ChooseCastFreeFromLibrary ChooseDestination = "cast_free_library"
	// ChooseLibraryBottom: the named cards go on the bottom of their
	// owner's library, in an order asked for next when there is more
	// than one — the London mulligan's cost (ADR 0168). The client says
	// so; the bot prices the pick as the hand it keeps.
	ChooseLibraryBottom ChooseDestination = "library_bottom"
	// ChooseManifest: the named cards from the chooser's own hand are
	// manifested — put onto the battlefield face down as 2/2 creatures
	// (CR 701.40a; Kozilek, the Broken Reality, ADR 0189 decision 3).
	// The client says "Manifest"; the bot prices the pick as cards it
	// gives up, because a face-down 2/2 is not the card.
	ChooseManifest ChooseDestination = "manifest"
	// ChooseSacrifice: the named permanents are sacrificed, together
	// with every other player's picks, once the last player in APNAP
	// order has chosen (CR 101.4, 608.2e; #3664). The client says so;
	// the bot prices the pick as the permanents it gives up.
	ChooseSacrifice ChooseDestination = "sacrifice"
)
