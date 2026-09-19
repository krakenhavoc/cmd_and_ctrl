package game

import (
	"errors"

	"github.com/google/uuid"
)

// entry_choice.go — "as this permanent enters, you may pay N life"
// (the Ravnica shockland cycle, CR 614.1c + CR 118.4).
//
// The shape is a replacement effect with a decision inside it, which
// is the one thing the CR 614 pipeline could not previously express.
// The pipeline is synchronous, but it already knows how to STOP: the
// CR 616 ordering prompt and the CR 614.10 "may" prompt both bail
// with errReplacementPending and resume from a stashed frame. This
// file adds a third bail of the same family — the difference is that
// answering it costs life, and that the replacement applies when the
// player says NO rather than when they say yes.
//
//	"you may pay 2 life. If you don't, it enters tapped."
//	 ^ decline → Replace runs → ev.EntersTapped = true
//	 ^ pay     → Replace never runs → the land enters untapped
//
// Everything here runs PRE-push, so the permanent's entry is what
// gets replaced. The observable difference from the OnETB workaround
// this superseded (enter tapped, then untap) is that there is no
// tapped window, no untap event, and no trip through the stack — the
// shockland tests count tap events to pin exactly that.

// PendingChoiceEntryPayLife is the "as this enters, you may pay N
// life" prompt. Answered with the same `{apply: bool}` payload as
// the other yes/no kinds — `true` pays, `false` declines — and
// routed by kind in the actions dispatcher.
//
// Declared here rather than in pending_choice.go's const block on
// purpose: the kind, its queue path, its resume path and the
// battlefield-entry push it settles into are one mechanism, and
// keeping them in one file is what makes it reviewable.
//
// The chooser is the entering permanent's controller. PayCost
// carries the human-readable payment ("2 life") for the prompt; the
// authoritative number lives on the effect's EntryLifeCost.
const PendingChoiceEntryPayLife PendingChoiceKind = "entry_pay_life"

// offerEntryLifePaymentLocked handles an applicable replacement that
// charges life to decline. Returns true when a prompt was queued and
// the caller should bail with errReplacementPending.
//
// Returns false when no prompt is possible:
//
//   - the payer can't be identified, or can't legally pay (CR 118.4:
//     a player may pay N life only if their life total is at least
//     N). A player at 1 life does not get to pay 2, and asking a
//     question whose only answer is "no" is worse than not asking.
//   - the entry can't be resumed (see ReplacementEvent.entryResumable).
//     Pausing an entry site that has no resume would strand the card
//     in its old zone; taking the un-paid branch instead costs the
//     player a choice they'd usually take, which is the weaker and
//     therefore safer failure.
//
// In every one of those cases the replacement is applied inline,
// unprompted — the player "doesn't pay", so the permanent enters
// tapped.
//
// Caller must hold g.mu.
func (g *Game) offerEntryLifePaymentLocked(ev *ReplacementEvent, chosen activeReplacement) bool {
	cost := chosen.effect.EntryLifeCost
	payer := g.entryLifePayerLocked(ev, chosen)
	p := g.playerByIDLocked(payer)
	if cost > 0 && ev.entryResumable && p != nil && !p.Eliminated && p.Life >= cost {
		g.queueEntryPayLifePromptLocked(ev, chosen, payer, cost)
		return true
	}
	// Unaffordable, unattributable or unresumable: the player
	// "doesn't", so the replacement applies.
	g.markReplacementAppliedLocked(ev, chosen.id)
	if chosen.effect.Replace != nil {
		if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
	}
	return false
}

// entryLifePayerLocked resolves who is asked to pay. The effect's
// own Controller hook wins (a card knows best — for an entering
// permanent it reads ev.Actor / the card's controller), then the
// generic affected-player rule.
//
// Caller must hold g.mu.
func (g *Game) entryLifePayerLocked(ev *ReplacementEvent, chosen activeReplacement) uuid.UUID {
	if chosen.effect.Controller != nil {
		if id := chosen.effect.Controller(ev, g, chosen.source); id != uuid.Nil {
			return id
		}
	}
	return affectedPlayerForEvent(ev, []activeReplacement{chosen}, g)
}

// queueEntryPayLifePromptLocked queues the pay-life prompt and
// stashes the continuation frame the resume path re-enters with.
//
// Caller must hold g.mu.
func (g *Game) queueEntryPayLifePromptLocked(
	ev *ReplacementEvent,
	chosen activeReplacement,
	payer uuid.UUID,
	cost int,
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
		Kind:                 PendingChoiceEntryPayLife,
		Chooser:              payer,
		Count:                1,
		Source:               source,
		Reason:               reason,
		PayCost:              lifeCostString(cost),
		ReplacementEffectIDs: []ReplacementEffectID{chosen.id},
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: []activeReplacement{chosen},
		},
	})
}

// lifeCostString renders a life payment for the prompt ("2 life").
func lifeCostString(n int) string {
	if n <= 0 {
		return ""
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits) + " life"
}

// ResolveEntryPayLife processes a resolve_choice action for a
// PendingChoiceEntryPayLife entry. `pay` is the controller's answer:
//
//	true  — pay the life; the replacement never fires, so the
//	        permanent enters the way it was going to (untapped).
//	false — decline; the replacement fires (enters tapped).
//
// A "pay" the player can no longer afford — life can change between
// the prompt and the answer — degrades to a decline rather than
// paying them into a negative total.
//
// Either way the apply-loop is re-entered so CR 616.1 can pick up
// anything newly applicable, and the settled event is then pushed
// through applyResolvedReplacementEventLocked, which is what
// actually puts the permanent onto the battlefield.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveEntryPayLife(choiceID, chooserID uuid.UUID, pay bool) error {
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
	if choice.Kind != PendingChoiceEntryPayLife {
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
	// The decision is made once per event either way, so the effect
	// is marked applied before anything else — a "pay" must not leave
	// the effect eligible to be gathered again on the next iteration.
	g.markReplacementAppliedLocked(ev, chosen.id)

	paid := false
	if pay {
		cost := chosen.effect.EntryLifeCost
		if p := g.playerByIDLocked(chooserID); cost > 0 && p != nil && p.Life >= cost {
			var source uuid.UUID
			if chosen.source != nil {
				source = chosen.source.InstanceID
			}
			// #793: the cost path. A shockland's "you may pay 2 life"
			// is a cost paid inside a replacement's own resume, so a
			// CR 616 prompt on it would nest one paused pipeline
			// inside another — the permanent entering untapped on the
			// strength of a payment still waiting to be ordered. It
			// settles in one step instead (CR 119.4 still applies the
			// life-loss replacements; see payLifeAsCostLocked).
			if err := g.PayLifeForEffect(source, chooserID, cost); err != nil {
				g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
			} else {
				paid = true
			}
		}
	}
	if !paid && chosen.effect.Replace != nil {
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
	// #478: through the shared finisher the other two resumes use,
	// rather than this path's own copy of it. The copy returned nil for
	// a CANCELLED event, which was harmless while the only entry that
	// could ask for life was a land play with nothing waiting on it —
	// and drops a search's shuffle and its caller's Then now that a
	// fetched shockland can be asked.
	return g.finishSettledReplacementLocked(ev, out)
}

// markReplacementAppliedLocked records that one effect has had its
// shot at this event (CR 614.5), allocating the per-event map when
// the resume path is the first thing to touch it.
//
// Caller must hold g.mu.
func (g *Game) markReplacementAppliedLocked(ev *ReplacementEvent, id ReplacementEffectID) {
	if ev == nil {
		return
	}
	if g.replacementsAppliedThisEvent == nil {
		g.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool)
	}
	if _, ok := g.replacementsAppliedThisEvent[ev.ID]; !ok {
		g.replacementsAppliedThisEvent[ev.ID] = make(map[ReplacementEffectID]bool)
	}
	g.replacementsAppliedThisEvent[ev.ID][id] = true
}

// executeEntryToBattlefieldLocked is the ONE finisher for a settled
// battlefield ENTRY: it performs the move with the settled event's
// payload and then runs what the effect that asked for it still owed.
//
// It started life as the resume half alone, and closed a hole that
// predates the shocklands: every battlefield-ENTRY path bails on
// errReplacementPending and trusts the resume path to finish the job,
// but applyResolvedReplacementEventLocked only ever implemented the
// battlefield-LEAVE case, so a paused entry left the card sitting in
// its old zone forever.
//
// #478 made it the finisher for the INLINE path too, for the three
// effect-side entries that reach it through
// enterBattlefieldThroughPipelineLocked — the library search, the exile
// return and the reanimation. Each of those used to have its own copy
// of the push, which is how an entry that paused and one that did not
// drift apart; there is one copy now, and the resume runs exactly the
// code the unpaused path runs. Five sites feed it:
//
//   - the land-play branch, whose push this reproduces exactly;
//   - (since S16.5) stack resolution, which reproduces it PLUS the two
//     jobs only stack resolution does — the resolved Aura's attach and
//     evoke's sacrifice trigger — off the StackItem the event carries;
//   - (since #478) the search, the exile return and the reanimation,
//     whose extra duties ride on ev.entryTail.
//
// The card is located live rather than trusted from ev.OldZone: the
// prompt is asynchronous, and the answer arrives in a later action. But
// the move is REFUSED when the card has left the zone the window opened
// over — it was milled, drawn or exiled while the question was open, so
// this entry can never happen and moving it from wherever it is now
// would rip it out of a zone nobody asked about. Same rule
// pausedZoneChangeStaleLocked prunes by; this is the backstop for a
// departure the prune does not run at.
//
// Returns the entering permanent's ID — the NEW one when the entry
// minted a new object (CR 400.7) — and uuid.Nil when nothing entered.
//
// Caller must hold g.mu.
func (g *Game) executeEntryToBattlefieldLocked(ev *ReplacementEvent) (entered uuid.UUID, err error) {
	defer func() {
		if tailErr := g.runEntryTailLocked(ev, entered); tailErr != nil && err == nil {
			err = tailErr
		}
	}()
	var (
		srcKind ZoneKind
		moved   Card
	)
	src := g.findCardZoneLocked(ev.CardID)
	switch {
	case src == g.Battlefield:
		// Something already resolved the entry; don't double-push.
		return ev.CardID, nil
	case src == nil:
		// #762: a CREATED TOKEN, staged in Game.enteringTokens while
		// its entry window was open. It is minted rather than moved —
		// a token comes from no zone at all (CR 111.1) — so it is
		// pushed here and its arrival announces EventTokenCreated
		// rather than a zone move. Everything after that point is the
		// entry every other permanent takes, which is the whole point
		// of it being on this path.
		tok, ok := g.takeEnteringTokenLocked(ev.CardID)
		if !ok {
			return uuid.Nil, ErrCardNotFound
		}
		if tok.AttackingTarget != uuid.Nil {
			// CR 506.3c: PUT onto the battlefield attacking, never
			// declared, so the attack-declaration lock-in must not
			// mistake this for a staged declaration and announce it.
			g.noteAttackAnnouncedLocked(tok.InstanceID)
		}
		g.Battlefield.PushTop(tok)
		moved = tok
	default:
		if src.Kind != ev.OldZone {
			return uuid.Nil, nil
		}
		srcKind = src.Kind
		m, err := MoveCard(src, g.Battlefield, ev.CardID)
		if err != nil {
			return uuid.Nil, err
		}
		moved = m
	}
	entered = moved.InstanceID
	if ev.entryTail != nil && ev.entryTail.newObject {
		// CR 400.7, and it happens FIRST: the reset zeroes the tapped
		// flag and the counters, so the settled EntersTapped and
		// EntersWithCounters below have to be applied to the new object
		// rather than to the one that is about to be wiped.
		if newID := g.resetAsNewObjectLocked(entered); newID != uuid.Nil {
			entered = newID
			moved.InstanceID = newID
		}
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID != entered {
			continue
		}
		if ev.Actor != uuid.Nil {
			g.Battlefield.Cards[i].Controller = ev.Actor
		} else if g.Battlefield.Cards[i].Controller == uuid.Nil {
			g.Battlefield.Cards[i].Controller = g.Battlefield.Cards[i].Owner
		}
		if ev.EntersTapped {
			g.Battlefield.Cards[i].Tapped = true
		}
		// CR 400.7d / ADR 0073 §5, the twin of the stamp in
		// resolveTopOfStackLocked: a kicked permanent spell whose
		// entry PAUSED on a prompt is still a kicked permanent when
		// it lands, and its "if it was kicked" trigger is harvested
		// off the EventETB emitted below. Stamped after the CR 400.7
		// reset above, which would otherwise wipe it.
		if ev.stackItem != nil && len(ev.stackItem.Paid.OptionalCosts) > 0 {
			g.Battlefield.Cards[i].PaidOptionalCosts =
				append([]int(nil), ev.stackItem.Paid.OptionalCosts...)
			moved.PaidOptionalCosts = g.Battlefield.Cards[i].PaidOptionalCosts
		}
		break
	}
	g.markCardKnownInZoneLocked(g.Battlefield, entered)
	// CR 707.2 — see the twin call in resolveTopOfStackLocked. The
	// copy lands before the counters and before any event, so an ETB
	// trigger never sees the permanent as its own printed self.
	if copied, ok := g.applyEntersAsCopyLocked(ev, entered); ok {
		moved = copied
	}
	for name, n := range ev.EntersWithCounters {
		_ = g.AddCounterForEffect(entered, name, n)
	}
	// Per-turn land-drop tally. The land branch in CastSpell bumps
	// this on the path where nothing pauses; this branch is the same
	// land play finishing after a prompt, and it was never bumping
	// it — so every shockland played since PR #268, and now every
	// MDFC land back, was a land drop the legal-move enumerator
	// never saw. Found by the MDFC back-face test; the bug is older
	// and applies to the whole pay-life cycle.
	//
	// CR 305.2: a land PLAY is a land drop and nothing else is. The
	// flag says which entry this is (#478) — inferring it from "a land
	// with no StackItem" was true while the land play and stack
	// resolution were the only resumable entries and would count a
	// fetched, reanimated or blinked land now that they are not.
	if ev.landPlay && ev.Actor != uuid.Nil {
		if g.LandsPlayedThisTurn == nil {
			g.LandsPlayedThisTurn = make(map[uuid.UUID]int)
		}
		g.LandsPlayedThisTurn[ev.Actor]++
	}
	if srcKind == "" {
		// #762: a created token arrived from no zone, so the
		// announcement is the creation itself (CR 111.1). One event,
		// never two — a token that also emitted a zone move would be
		// counted twice by everything watching permanents arrive.
		g.EmitEvent(Event{
			Kind:   EventTokenCreated,
			Actor:  ev.Actor,
			Source: ev.Source,
			CardID: entered,
		})
	} else {
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   ev.Actor,
			CardID:  entered,
			OldZone: srcKind,
			NewZone: ZoneBattlefield,
		})
	}
	// S16.5: the two jobs stack resolution does that no other entry
	// site does. Both need the StackItem, which is why carrying it
	// across the pause is what made the stack site resumable at all.
	// The Aura attach goes between the zone-move and the ETB so an
	// ETB trigger already sees the attachment, exactly as the
	// un-paused path orders it.
	if ev.stackItem != nil {
		g.attachResolvedAuraLocked(entered, ev.stackItem)
	}
	g.EmitEvent(Event{
		Kind:   EventETB,
		Actor:  ev.Actor,
		CardID: entered,
	})
	g.fireETBHookLocked(entered, CatalogKey(moved))
	if ev.stackItem != nil {
		g.queueAltCostEntryTriggerLocked(moved, ev.stackItem)
	}
	return entered, nil
}
