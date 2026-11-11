package game

import (
	"errors"

	"github.com/google/uuid"
)

// copy_choice.go — "you may have this creature enter as a copy of
// any creature on the battlefield" (CR 707.2 + CR 614.1c).
//
// This is the third member of the family entry_choice.go started: a
// replacement effect with a DECISION inside it, applied as the
// permanent enters so that it never exists on the battlefield as
// its own printed self. Everything an ETB trigger sees — its own
// and everyone else's — is the copied characteristics, because the
// copy lands before EventETB is emitted.
//
// The shape mirrors the shockland's pay-life prompt exactly:
//
//	applyReplacementsLocked sees CopySelector != nil
//	  → offerCopyChoiceLocked queues PendingChoiceCopyTarget, bails
//	  → ResolveCopyTarget stamps ev.EntersAsCopyOf and re-enters
//	  → applyResolvedReplacementEventLocked pushes the permanent
//	    with the copied values already on it
//
// # The entry site had to learn to resume
//
// A shockland is PLAYED, and the land-play branch of CastSpell was
// the one battlefield-entry site flagged `entryResumable`. A Clone
// is CAST, so it enters from the stack — and that site could not
// pause, because the generic resume did not reproduce the two
// things stack resolution does that nothing else does: attaching a
// resolved Aura to what it targeted, and queueing evoke's sacrifice
// trigger from the StackItem. `ReplacementEvent.stackItem` carries
// the item across the pause so the resume can do both, which is
// what makes the stack site resumable and this prompt possible.
//
// That also closes a latent hole older than this file: ANY permanent
// spell whose entry drew two applicable replacements would queue the
// CR 616 ordering prompt, bail, and never be pushed — the card sat
// on the stack with its StackMeta entry already deleted. Nothing had
// hit it because no catalog pair had yet lined up on one entry.

// PendingChoiceCopyTarget is the "choose what this permanent enters
// as a copy of" prompt. Answered with the same `{card_ids: []}`
// payload as the other card-grid kinds; an EMPTY list is a legal
// answer and means "don't copy anything", because every printed
// member of this class says "you MAY have this enter as a copy"
// (CR 614.1c). Declining a Clone is a real choice — it enters as a
// 0/0 and dies to the CR 704.5f state-based action, which is what
// the card says happens.
//
// The chooser is the entering permanent's controller. CopyOptions
// carries the candidate set, recomputed against the live board when
// the answer arrives (the prompt is asynchronous, and a creature can
// leave in the meantime).
const PendingChoiceCopyTarget PendingChoiceKind = "copy_target"

// CopySelector turns an ordinary ReplacementEffect into an "enters
// as a copy" effect. Declared on ReplacementEffect.CopySelector; the
// apply-loop branches to this file when it is non-nil, and the
// effect's own Replace is never called (there is nothing about the
// event left for it to rewrite).
type CopySelector struct {
	// Candidates returns the permanents this effect may copy, in
	// board order. Evaluated twice: once to build the prompt, and
	// once when the answer arrives, so a creature that left in
	// response cannot be copied. Returning an empty list means the
	// permanent simply enters as itself.
	//
	// A filter that reads the SPELL the permanent is entering from
	// asks g.EntryCastCountsForEffect(ev) — Mockingbird's "mana value
	// less than or equal to the amount of mana spent to cast this
	// creature" is its ManaSpent (#1735). The event, not the source,
	// is what carries the resolving stack item, so that read needs no
	// new parameter here; for an entry that is not a resolving spell
	// it answers the zero CastCounts.
	Candidates func(ev *ReplacementEvent, g *Game, source *Card) []uuid.UUID

	// Except applies the card's "except" clause to the values that
	// are about to land — Sakashima keeping its own name, Spark
	// Double dropping legendary, Phyrexian Metamorph adding
	// artifact. It receives the event too, so a clause that changes
	// how the permanent ENTERS rather than what it copies (Spark
	// Double's additional +1/+1 counter) can reach
	// ev.AddCounterAtETB. Nil for a plain Clone.
	Except func(ev *ReplacementEvent, v *PrintedValues, g *Game, source *Card)

	// UntilEndOfTurn makes the copy last only until the cleanup step
	// (#1593): "as this artifact enters, you may have it become a copy
	// of any creature on the battlefield until end of turn" (Cursed
	// Mirror). The copy still lands as the permanent enters, before
	// any event; see settleTimedEntryCopyLocked for how it gets its
	// end.
	UntilEndOfTurn bool

	// ExchangeText makes the pick the creature to EXCHANGE TEXT BOXES
	// with instead of the object to copy (ADR 0160 Decision 6): "As
	// Deadpool enters, you may exchange his text box and another
	// creature's" (CR 612.5). The prompt, its answer and its resume are
	// the copy choice's; nothing is copied, Except and UntilEndOfTurn
	// are ignored, and the exchange is registered once the permanent's
	// object identity exists, before EventETB.
	ExchangeText bool

	// ExileWithAnother makes the copy a choice of TWO candidates, both
	// exiled as the permanent enters (The Mimeoplasm, #5457): "you may
	// exile two creature cards from graveyards. If you do, it enters as a
	// copy of one of those cards with a number of additional +1/+1
	// counters on it equal to the power of the other card." The prompt is
	// asked twice over the same paused entry: first the card to copy
	// (CopyUseExileCopied), then the other card among the candidates left
	// (CopyUseExileOther). An empty answer to either, fewer than two
	// candidates, or a pick that has left its graveyard in the meantime is
	// the "may" declined: nothing is exiled and the permanent enters as
	// itself. The other card's power is read where it is, as the two are
	// exiled (PowerAnywhereForEffect), and the counters are part of the
	// entry (AddCounterAtETB).
	ExileWithAnother bool
}

// offerCopyChoiceLocked handles an applicable copy-selector
// replacement. Returns true when a prompt was queued and the caller
// should bail with errReplacementPending.
//
// Returns false — permanent enters as its own printed self, which
// for a Clone means a 0/0 that dies — when:
//
//   - nothing is legal to copy. The prompt would have one answer.
//   - the chooser can't be identified or has left the game. Asking
//     a question nobody can answer wedges the table (the same
//     posture applyReplacementsLocked takes for the CR 616 and
//     "may" replacement prompts).
//   - the entry site can't be resumed. Pausing there would strand
//     the card in its old zone; declining the copy is weaker than
//     printed and never stronger, which is the posture
//     ReplacementEvent.entryResumable exists to enforce.
//
// Caller must hold g.mu.
func (g *Game) offerCopyChoiceLocked(ev *ReplacementEvent, chosen activeReplacement) bool {
	sel := chosen.effect.CopySelector
	if sel == nil || sel.Candidates == nil {
		g.markReplacementAppliedLocked(ev, chosen.id)
		return false
	}
	candidates := sel.Candidates(ev, g, chosen.source)
	chooser := g.copyChooserLocked(ev, chosen)
	// A token creation (Esix, Fractal Bloom) resumes as any "may" on the
	// instruction does; only an ENTRY needs its site to be resumable.
	resumable := ev.entryResumable || ev.Kind == RepEventCreateTokens
	if sel.ExileWithAnother && len(candidates) < 2 {
		// "Exile two": one card is not enough to do it.
		candidates = nil
	}
	if len(candidates) == 0 || !resumable || g.chooserGoneLocked(chooser) || chooser == uuid.Nil {
		g.markReplacementAppliedLocked(ev, chosen.id)
		return false
	}
	g.queueCopyTargetPromptLocked(ev, chosen, chooser, candidates)
	return true
}

// copyChooserLocked resolves who chooses. The effect's Controller
// hook wins (for an entering permanent it reads ev.Actor or the
// card's controller), then the generic affected-player rule.
//
// Caller must hold g.mu.
func (g *Game) copyChooserLocked(ev *ReplacementEvent, chosen activeReplacement) uuid.UUID {
	if chosen.effect.Controller != nil {
		if id := chosen.effect.Controller(ev, g, chosen.source); id != uuid.Nil {
			return id
		}
	}
	return affectedPlayerForEvent(ev, []activeReplacement{chosen}, g)
}

// queueCopyTargetPromptLocked queues the picker and stashes the
// continuation frame ResolveCopyTarget re-enters with.
//
// Caller must hold g.mu.
func (g *Game) queueCopyTargetPromptLocked(
	ev *ReplacementEvent,
	chosen activeReplacement,
	chooser uuid.UUID,
	candidates []uuid.UUID,
) {
	reason := chosen.effect.PromptQuestion
	if reason == "" {
		reason = chosen.effect.Label
	}
	var source uuid.UUID
	if chosen.source != nil {
		source = chosen.source.InstanceID
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:                 PendingChoiceCopyTarget,
		Chooser:              chooser,
		Count:                1,
		Source:               source,
		Reason:               reason,
		CopyOptions:          append([]uuid.UUID(nil), candidates...),
		ExchangesText:        chosen.effect.CopySelector != nil && chosen.effect.CopySelector.ExchangeText,
		CopyUse:              copyUseOf(ev, chosen.effect.CopySelector),
		ReplacementEffectIDs: []ReplacementEffectID{chosen.id},
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: []activeReplacement{chosen},
		},
	})
}

// ResolveCopyTarget answers a PendingChoiceCopyTarget.
//
// `cardID` is the permanent to copy, or uuid.Nil for "I decline" —
// the answer the client sends when the player submits no pick.
// A pick that is no longer on the candidate list (it died, or was
// never legal) degrades to a decline rather than erroring: the board
// moves between the prompt and the answer, and the CR 614 pipeline
// has no way to re-ask.
//
// Either way the apply-loop is re-entered so CR 616.1 picks up
// anything newly applicable, and the settled event is pushed through
// applyResolvedReplacementEventLocked — which is what actually puts
// the permanent onto the battlefield, wearing the copied values.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveCopyTarget(choiceID, chooserID, cardID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceCopyTarget {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	frame := choice.replacementResume
	g.dequeueChoiceLocked(idx)
	if frame == nil || frame.ev == nil || len(frame.applicable) == 0 {
		return ErrInvalidParam
	}

	ev := frame.ev
	chosen := frame.applicable[0]
	// The decision happens once per event whichever way it goes, so
	// the effect is marked applied before anything else — otherwise
	// the next apply-loop iteration gathers it again and re-prompts
	// forever.
	g.markReplacementAppliedLocked(ev, chosen.id)

	if ev.Kind == RepEventCreateTokens {
		return g.resolveTokenCopyChoiceLocked(ev, chosen, cardID)
	}

	exchange := chosen.effect.CopySelector != nil && chosen.effect.CopySelector.ExchangeText
	if sel := chosen.effect.CopySelector; sel != nil && sel.ExileWithAnother {
		if g.copyOfTwoStepLocked(ev, chosen, chooserID, cardID) {
			// The second question is open over the same paused entry.
			return nil
		}
	} else if exchange && cardID != uuid.Nil && g.copyCandidateStillLegalLocked(ev, chosen, cardID) {
		ev.exchangeTextWith = cardID
	} else if cardID != uuid.Nil && g.copyCandidateStillLegalLocked(ev, chosen, cardID) {
		if src, ok := g.copySourceCardLocked(cardID); ok {
			values := SingleFacedCopiableValuesOf(src)
			// #4797: named before Except runs, so an except clause can
			// read what is being copied (CopySourceForEffect) — Superior
			// Spider-Man's "When you do, exile that card".
			ev.copySourceID = cardID
			if sel := chosen.effect.CopySelector; sel != nil && sel.Except != nil {
				sel.Except(ev, &values, g, chosen.source)
			}
			ev.EntersAsCopyOf = &values
			ev.copyUntilEndOfTurn = chosen.effect.CopySelector != nil && chosen.effect.CopySelector.UntilEndOfTurn
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
	// Through the shared finisher the other entry resumes use (#478),
	// rather than returning nil for a CANCELLED entry: a cancelled
	// entry still owes its caller an answer — a search its shuffle, a
	// simultaneous entry (#1322) the rest of its batch.
	return g.finishSettledReplacementLocked(ev, out)
}

// CopyUseTokens is PendingChoice.CopyUse for a copy_target asked inside a
// token creation: "you may instead choose a creature … and create that
// many tokens that are copies of that creature" (Esix, Fractal Bloom).
// The pick is what the TOKENS copy; nothing is entering as a copy. An
// on-disk and wire value, never renamed.
const CopyUseTokens = "tokens"

// CopyUseExileCopied and CopyUseExileOther are PendingChoice.CopyUse for
// the two questions of CopySelector.ExileWithAnother (The Mimeoplasm):
// the card to exile and enter as a copy of, then the other card to
// exile, whose power becomes +1/+1 counters. On-disk and wire values,
// never renamed.
const (
	CopyUseExileCopied = "exile_copied"
	CopyUseExileOther  = "exile_other"
)

// copyUseOf is the CopyUse a copy_target prompt over ev carries.
func copyUseOf(ev *ReplacementEvent, sel *CopySelector) string {
	switch {
	case ev != nil && ev.Kind == RepEventCreateTokens:
		return CopyUseTokens
	case sel != nil && sel.ExileWithAnother && ev != nil && ev.copySourceID != uuid.Nil:
		return CopyUseExileOther
	case sel != nil && sel.ExileWithAnother:
		return CopyUseExileCopied
	}
	return ""
}

// copyOfTwoStepLocked answers one of CopySelector.ExileWithAnother's two
// questions. The first answer names the card to copy on the event
// (copySourceID, which a restore point carries) and asks for the other
// card among the candidates left; it reports true, and the entry stays
// paused. The second answer exiles both cards and stamps the entry: the
// first card's copiable values, the except clause, and +1/+1 counters
// equal to the second card's power. Any decline or stale pick clears the
// event's copy source and reports false with nothing exiled, so the
// permanent enters as itself.
//
// Caller must hold g.mu, with the effect already marked applied.
func (g *Game) copyOfTwoStepLocked(ev *ReplacementEvent, chosen activeReplacement, chooser, cardID uuid.UUID) bool {
	sel := chosen.effect.CopySelector
	first := ev.copySourceID
	if first == uuid.Nil {
		if cardID == uuid.Nil || !g.copyCandidateStillLegalLocked(ev, chosen, cardID) {
			return false
		}
		var rest []uuid.UUID
		for _, id := range sel.Candidates(ev, g, chosen.source) {
			if id != cardID {
				rest = append(rest, id)
			}
		}
		if len(rest) == 0 {
			return false
		}
		ev.copySourceID = cardID
		// The effect is still being decided, so it is not yet applied: a
		// restore point taken on the second question gathers it again by
		// its ID (rebuildReplacementResumeLocked gathers only what has not
		// been applied). Nothing runs the apply-loop for this event while
		// the question is open, and the second answer marks it.
		delete(g.replacementsAppliedThisEvent[ev.ID], chosen.id)
		g.queueCopyTargetPromptLocked(ev, chosen, chooser, rest)
		return true
	}
	ev.copySourceID = uuid.Nil
	if cardID == uuid.Nil || cardID == first ||
		!g.copyCandidateStillLegalLocked(ev, chosen, cardID) ||
		!g.copyCandidateStillLegalLocked(ev, chosen, first) {
		return false
	}
	src, ok := g.copySourceCardLocked(first)
	other, okOther := g.copySourceCardLocked(cardID)
	if !ok || !okOther {
		return false
	}
	values := SingleFacedCopiableValuesOf(src)
	power := g.PowerAnywhereForEffect(other)
	ev.copySourceID = first
	if sel.Except != nil {
		sel.Except(ev, &values, g, chosen.source)
	}
	ev.EntersAsCopyOf = &values
	if power > 0 {
		ev.AddCounterAtETB(CounterPlusOne, power)
	}
	for _, id := range []uuid.UUID{first, cardID} {
		if _, err := g.routeCardToZoneLocked(zoneRoute{CardID: id, Dst: ZoneExile}); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
	}
	return false
}

// resolveTokenCopyChoiceLocked is ResolveCopyTarget for a copy selector
// on a token CREATION (CR 701.7b, CR 614.1a). A legal pick is named on
// the event (CopySourceForEffect) and the effect's own Replace then
// rewrites the instruction from it, as an Optional "yes" would; a
// decline, a stale pick, or a source that has left since the prompt
// leaves the instruction as it was. Either way the apply-loop is
// re-entered and the creation finishes through the resume the "may"
// replacements on a creation use.
//
// Caller must hold g.mu, with the effect already marked applied.
func (g *Game) resolveTokenCopyChoiceLocked(ev *ReplacementEvent, chosen activeReplacement, cardID uuid.UUID) error {
	if cardID != uuid.Nil && g.copyCandidateStillLegalLocked(ev, chosen, cardID) {
		if live, present := g.liveReplacementSourceLocked(chosen); present {
			ev.copySourceID = cardID
			g.runReplaceLocked(ev, live)
			ev.copySourceID = uuid.Nil
		}
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	return g.finishReplacementResumeLocked(ev, out)
}

// copyCandidateStillLegalLocked re-runs the selector against the
// live board. The prompt is asynchronous; the creature the player
// picked may be in a graveyard by the time the answer lands.
//
// Caller must hold g.mu.
func (g *Game) copyCandidateStillLegalLocked(
	ev *ReplacementEvent,
	chosen activeReplacement,
	cardID uuid.UUID,
) bool {
	sel := chosen.effect.CopySelector
	if sel == nil || sel.Candidates == nil {
		return false
	}
	for _, id := range sel.Candidates(ev, g, chosen.source) {
		if id == cardID {
			return true
		}
	}
	return false
}

// copySourceCardLocked is the object a copy_target answer names, as the
// copy reads it: a permanent on the battlefield, or — for "enter as a
// copy of any creature card in a graveyard" (Body Double, Superior
// Spider-Man, #4797) — a card in a graveyard, whose copiable values
// are its printed ones (CR 707.2). The candidate list decides which
// zones a card offers; this only finds the pick where it is. Anywhere
// else (it was exiled in response) it is no longer there to copy, and
// the answer degrades to a decline like any other stale pick.
//
// Caller must hold g.mu.
func (g *Game) copySourceCardLocked(cardID uuid.UUID) (Card, bool) {
	if src, ok := g.battlefieldCardLocked(cardID); ok {
		return *src, true
	}
	z := g.findCardZoneLocked(cardID)
	if z == nil || z.Kind != ZoneGraveyard {
		return Card{}, false
	}
	for _, c := range z.Cards {
		if c.InstanceID == cardID {
			return c, true
		}
	}
	return Card{}, false
}

// CopySourceForEffect is the object this entry is copying, once its
// copy_target prompt has been answered with one: a permanent, or a card
// in a graveyard (#4797). uuid.Nil before the answer and for a decline.
// An except clause reads it to name "that card" (Superior Spider-Man's
// "When you do, exile that card").
func (ev *ReplacementEvent) CopySourceForEffect() uuid.UUID {
	if ev == nil {
		return uuid.Nil
	}
	return ev.copySourceID
}

// battlefieldCardLocked returns a pointer to the named battlefield
// card. Caller must hold g.mu.
func (g *Game) battlefieldCardLocked(cardID uuid.UUID) (*Card, bool) {
	if i := findCardOnBattlefield(g, cardID); i >= 0 {
		return &g.Battlefield.Cards[i], true
	}
	return nil, false
}

// applyEntersAsCopyLocked stamps a settled copy onto the permanent
// that has just been pushed to the battlefield, BEFORE the zone-move
// and ETB events fire. That ordering is the rule, not an
// optimisation: a permanent that enters as a copy never exists on
// the battlefield as its own printed self (CR 707.2), so every ETB
// trigger — its own and every watcher's — must see the copied
// characteristics.
//
// Returns the post-copy card value so the caller can take the
// catalog key off the right identity: the value MoveCard handed back
// is a pre-copy snapshot, and firing the ETB hook with that key
// would run the Clone's (empty) entry rather than the copied card's.
//
// Caller must hold g.mu.
func (g *Game) applyEntersAsCopyLocked(ev *ReplacementEvent, cardID uuid.UUID) (Card, bool) {
	if ev == nil || ev.EntersAsCopyOf == nil {
		return Card{}, false
	}
	dst, ok := g.battlefieldCardLocked(cardID)
	if !ok {
		return Card{}, false
	}
	// The source is read for its card-carried ability slices, which
	// only a token has. A token that left in the meantime just means
	// those come across empty; the copiable values were snapshotted
	// when the player answered and are still right.
	var src Card
	if from, ok := g.battlefieldCardLocked(ev.copySourceID); ok {
		src = *from
	}
	dst.applyCopy(*ev.EntersAsCopyOf, src)
	g.EmitEvent(Event{
		Kind:   EventCopyApplied,
		CardID: cardID,
		Source: ev.copySourceID,
		Actor:  ev.Actor,
	})
	return *dst, true
}
