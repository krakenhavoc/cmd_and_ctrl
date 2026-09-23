package game

import (
	"errors"

	"github.com/google/uuid"
)

// entry_reveal.go — "as this permanent enters, you may reveal an
// Island or Swamp card from your hand. If you don't, it enters
// tapped." (the ten reveal-lands, CR 614.1c + CR 701.20.)
//
// entry_choice.go's sibling, and deliberately its mirror image: the
// same replacement with a decision inside it, the same pre-push
// pause, the same resume — with a CARD where the shockland has a
// number. That difference is the whole seam. The CR 614 apply-loop
// could stop and ask three questions before this file, and all three
// are yes/no-shaped:
//
//	Optional      — "do you apply this?"            (CR 614.10)
//	EntryLifeCost — "do you pay 2 life?"            (the shocklands)
//	CopySelector  — "which permanent do you copy?"  (CR 707.2, Clone)
//
// None of them can name a card in a HAND. `Optional` cannot name
// anything, and a copy selector names a permanent on the battlefield
// and rewrites the entering object with the answer. So the twelve
// cards on the "Reveal-from-hand entry choice" seam row waited, and
// the batch-02 census entry named the gap exactly: *a card choice
// inside a replacement effect*. See ADR 0013 §5z.
//
//	"you may reveal an Island or Swamp card from your hand.
//	 If you don't, this land enters tapped."
//	 ^ decline / cannot → Replace runs → ev.EntersTapped = true
//	 ^ reveal           → Replace never runs → the land enters untapped
//
// Everything here runs PRE-push, so what is replaced is the
// permanent's entry itself: no tapped window, no untap event and no
// trip through the stack, which is the property the shockland tests
// count tap events to pin and this file's tests count the same way.
//
// # Nothing moves
//
// A reveal is not a zone change (CR 701.20b). The named card is still
// in the revealer's hand when the land finishes entering, and it is
// not a cost — which is why this is not the discard row (Mox Diamond,
// docs/engine-seams.md) and why the bot's sign is the opposite of a
// choose_cards pick's: naming a card here costs nothing at all.

// PendingChoiceEntryRevealFromHand is the "as this enters, you may
// reveal a card from your hand" prompt.
//
// It carries the CHOOSE-CARDS payload — {card_ids}, ChooseMin /
// ChooseMax, and a chooseCardsFrame pinned to ZoneHand — so
// checkChooseCardsPicksLocked (#1017) stays the one copy of "would
// this answer be accepted": bounds, candidacy, duplicates, the live
// zone re-check and the set-level Validate hook, shared with the
// submit path and with the enumerator through
// ChooseCardsPickLegalLocked. It is the third member of
// isCardSetPickKind after choose_cards (#74) and #826's untap_choice.
//
// It is a separate KIND rather than a choose_cards with a replacement
// frame bolted on, for untap_choice's three reasons with the third
// again deciding:
//
//   - The continuation is not a card's next sentence, it is a PAUSED
//     EVENT. resolveCardSetPick dequeues and runs the frame's `then`;
//     this kind has to re-enter the apply-loop and settle the entry
//     through finishSettledReplacementLocked, which is the
//     replacement pipeline's resume contract and not the chain's.
//   - The sentence is different. "Choose cards" is not "reveal a card
//     to keep this untapped", and the kind is what the client renders.
//   - The SIGN is inverted for a bot. The heuristic's choose_cards
//     branch scores an answer by what it does NOT name (#798) —
//     every choose_cards prompt over a bot's own hand is a card being
//     given up. A card named here is revealed and KEPT, so scored
//     through that branch "reveal nothing" would win every time and
//     all ten lands would enter tapped forever.
//
// Declared here rather than in pending_choice.go's const block for
// entry_pay_life's reason: the kind, its queue path, its resume path
// and the battlefield-entry push it settles into are one mechanism.
const PendingChoiceEntryRevealFromHand PendingChoiceKind = "entry_reveal_from_hand"

// EntryHandReveal is the declaration on ReplacementEffect: "as this
// permanent enters, you may reveal <Min..Max cards matching Matches>
// from your hand; if you don't, <Replace>."
//
// The inversion is EntryLifeCost's and it is why this is not
// Optional: an Optional "yes" APPLIES the replacement, and here
// revealing is what AVOIDS it.
type EntryHandReveal struct {
	// Matches is the clause's filter — "an Island or Swamp card".
	// Nil admits every card in hand.
	//
	// A function of the CARD and of nothing else, which is
	// ChooseCardsPrompt.Validate's rule for its reason: the predicate
	// runs on the submit path under the write lock and, through
	// ChooseCardsPickLegalLocked, inside legal.EnumerateFor under the
	// READ lock against whichever clone the enumerator was handed. A
	// *Game argument would put the whole *ForEffect mutation surface
	// one call away from a read-locked caller with only a comment in
	// the way. It must not write through the Card it is handed (the
	// maps are shared with the live card) and must not capture a
	// *Game or a pointer into a zone.
	//
	// It reads a card in a HAND, where no layer has been applied, so
	// it sees printed characteristics. That is correct for every card
	// in the family: "an Island or Swamp card" is a type-line test,
	// and a land-type grant on the battlefield says nothing about a
	// card in somebody's hand.
	Matches func(c Card) bool

	// Min and Max bound the reveal. Every printed card today is
	// 0..1 — "you MAY reveal A card" — and the zero floor is load
	// bearing rather than incidental: it makes "reveal nothing" an
	// answer the resolver can never refuse and the enumerator's
	// AlwaysLegal move, so this prompt cannot be the #544 wedge even
	// with every candidate gone from the hand.
	//
	// Max <= 0 is read as 1.
	Min, Max int

	// Question is the prompt's header — the card's own sentence.
	// Falls back to ReplacementEffect.PromptQuestion and then Label.
	Question string

	// Then, when set, runs after the cards have been revealed and
	// before the apply-loop is re-entered, with the revealed IDs in
	// the order the chooser submitted them. It is the door a card
	// that does something MORE with the revealed cards comes through;
	// no catalog card needs it today (see ADR 0013 §5z on why exile
	// is not a flag here), and it is never called for a decline.
	//
	// Runs with g.mu held.
	Then func(g *Game, revealed []uuid.UUID) error
}

// offerEntryHandRevealLocked handles an applicable replacement whose
// decline condition is "you didn't reveal". Returns true when a
// prompt was queued and the caller should bail with
// errReplacementPending.
//
// Returns false when no prompt is possible, and the three cases are
// offerEntryLifePaymentLocked's one for one:
//
//   - nothing in the revealer's hand matches the clause. Asking a
//     question whose only answer is "no" is worse than not asking,
//     and an empty hand is the ABSENCE of the question rather than a
//     refusal of it.
//   - the entry can't be resumed (ReplacementEvent.entryResumable).
//     Pausing an entry site with no resume would strand the card in
//     its old zone; taking the un-revealed branch instead costs the
//     player a choice they'd usually take, which is the weaker and
//     therefore safer failure.
//   - the revealer can't be identified, or has left the game
//     (CR 800.4a).
//
// In every one of those the replacement is applied inline,
// unprompted — the player "doesn't reveal", so the permanent enters
// tapped.
//
// Caller must hold g.mu.
func (g *Game) offerEntryHandRevealLocked(ev *ReplacementEvent, chosen activeReplacement) bool {
	spec := chosen.effect.EntryHandReveal
	revealer := g.entryChoicePlayerLocked(ev, chosen)
	if spec != nil && ev.entryResumable && !g.chooserGoneLocked(revealer) {
		if candidates := g.handCardsMatchingLocked(revealer, spec.Matches); len(candidates) > 0 {
			if g.queueEntryHandRevealPromptLocked(ev, chosen, revealer, candidates) {
				return true
			}
		}
	}
	// Nothing to show, nobody to ask, or an entry with nothing to
	// resume it: the player "doesn't", so the replacement applies.
	g.markReplacementAppliedLocked(ev, chosen.id)
	if chosen.effect.Replace != nil {
		if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
	}
	return false
}

// handCardsMatchingLocked is every card in `player`'s hand the
// clause admits, in hand order.
//
// The candidate list is FROZEN here, the way every other card-set
// pick's is, and re-checked against the live zone on submit
// (pickStillInPickZoneLocked). Nothing can move a card out of that
// hand while the prompt is open — an open choice stops priority — so
// the re-check is the backstop rather than the mechanism.
//
// Caller must hold g.mu.
func (g *Game) handCardsMatchingLocked(player uuid.UUID, matches func(Card) bool) []uuid.UUID {
	p := g.playerByIDLocked(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, len(p.Hand.Cards))
	for _, c := range p.Hand.Cards {
		if matches != nil && !matches(c) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// queueEntryHandRevealPromptLocked queues the pick and stashes the
// continuation frame the resume path re-enters with. Reports whether
// anything was queued — QueueChoiceForEffect refuses a chooser who
// has left the game (#864), and a caller that believed it would leave
// the entry paused forever.
//
// Caller must hold g.mu.
func (g *Game) queueEntryHandRevealPromptLocked(
	ev *ReplacementEvent,
	chosen activeReplacement,
	revealer uuid.UUID,
	candidates []uuid.UUID,
) bool {
	spec := chosen.effect.EntryHandReveal
	hi := spec.Max
	if hi <= 0 {
		hi = 1
	}
	if hi > len(candidates) {
		hi = len(candidates)
	}
	lo := spec.Min
	if lo < 0 {
		lo = 0
	}
	if lo > hi {
		lo = hi
	}
	question := spec.Question
	if question == "" {
		question = chosen.effect.PromptQuestion
	}
	if question == "" {
		question = chosen.effect.Label
	}
	var source uuid.UUID
	if chosen.source != nil {
		source = chosen.source.InstanceID
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:    PendingChoiceEntryRevealFromHand,
		Chooser: revealer,
		// FromPlayer is the revealer too — the hand is their own.
		// It is what filterPendingChoices and the departure sweep
		// read to tell a question about somebody else's material from
		// one about the asker's own (#961), and every printed card in
		// this family reveals from the controller's own hand.
		FromPlayer:           revealer,
		Count:                hi,
		Source:               source,
		Reason:               question,
		ChooseCards:          append([]uuid.UUID(nil), candidates...),
		ChooseMin:            lo,
		ChooseMax:            hi,
		ReplacementEffectIDs: []ReplacementEffectID{chosen.id},
		chooseCardsResume: &chooseCardsFrame{
			// The zone, and no `then`: the continuation of this
			// prompt is a paused replacement event, and it is run by
			// ResolveEntryRevealFromHand rather than by
			// resolveCardSetPick. The frame is here for the live-zone
			// re-check and so that ChooseCardsPickLegalLocked reaches
			// the same rule the resolver does.
			zone: ZoneHand,
		},
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: []activeReplacement{chosen},
		},
	}) != uuid.Nil
}

// ResolveEntryRevealFromHand processes a resolve_choice action for a
// PendingChoiceEntryRevealFromHand entry. `picks` are the cards the
// controller is revealing:
//
//	one or more — they are revealed to the table; the replacement
//	              never fires, so the permanent enters the way it was
//	              going to (untapped).
//	none        — declined; the replacement fires (enters tapped).
//
// Validation runs BEFORE the dequeue, so a client that submits a card
// that has left the hand gets an error and can try again rather than
// losing the prompt — ResolveChooseCards' contract, through the same
// function. The empty answer skips the re-check entirely and is
// therefore the one answer that can never be refused.
//
// Either way the apply-loop is re-entered so CR 616.1 can pick up
// anything newly applicable, and the settled event is then pushed
// through the shared finisher, which is what actually puts the
// permanent onto the battlefield.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveEntryRevealFromHand(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceEntryRevealFromHand {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if err := g.checkChooseCardsPicksLocked(choice, picks); err != nil {
		return err
	}
	frame := choice.replacementResume
	reason := choice.Reason
	source := choice.Source
	g.dequeueChoiceLocked(idx)
	if frame == nil || frame.ev == nil || len(frame.applicable) == 0 {
		return ErrInvalidParam
	}

	ev := frame.ev
	chosen := frame.applicable[0]
	// The decision is made once per event either way, so the effect
	// is marked applied before anything else — a reveal must not
	// leave the effect eligible to be gathered again on the next
	// iteration (CR 614.5).
	g.markReplacementAppliedLocked(ev, chosen.id)

	if len(picks) > 0 {
		// CR 701.20: the table sees them and is entitled to remember.
		// Through the one reveal primitive, so the knower set and the
		// grouped EventRevealCards run are the ones every other
		// reveal in the engine produces — the other seats learn what
		// was shown here, through the log, because the picker itself
		// is the owner's alone.
		g.RevealForEffect(RevealSpec{
			Player: chooserID,
			Source: source,
			Reason: reason,
			Cards:  picks,
		})
		if then := chosen.effect.EntryHandReveal.Then; then != nil {
			if err := then(g, append([]uuid.UUID(nil), picks...)); err != nil {
				g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
			}
		}
	} else if chosen.effect.Replace != nil {
		if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
	}

	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// Another branch-point; the chain keeps unrolling.
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	// The shared finisher the other entry resumes use (#478), so a
	// fetched reveal-land's search still gets its shuffle and its
	// caller's Then.
	return g.finishSettledReplacementLocked(ev, out)
}
