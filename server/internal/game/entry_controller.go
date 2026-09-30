package game

import (
	"errors"

	"github.com/google/uuid"
)

// entry_controller.go — "This enchantment enters under the control of
// an opponent of your choice" (Captive Audience, Pendant of Prosperity,
// Abby, Merciless Soldier, Xantcha, Sleeper Agent). ADR 0102, #1759.
//
// The clause is a CR 614.1d replacement effect from the permanent
// itself (CR 614.12), and its choice is made before the permanent
// enters (CR 614.12a). So it is asked INSIDE the CR 614 window, the way
// a Clone's copy and a shockland's life are, and never on cast: a
// reanimated, blinked or copied permanent with the clause asks too,
// because the clause is on the permanent and not on the spell.
//
// The shape is copy_choice.go's:
//
//	applyReplacementsLocked sees EntryController != nil
//	  → offerEntryControllerLocked either settles it inline (one
//	    opponent, or an entry that cannot pause) or queues
//	    PendingChoiceEntryController and bails
//	  → ResolveEntryController sets the would-be controller and
//	    re-enters the apply-loop
//	  → the shared finisher lands the permanent under that player
//
// # The would-be controller
//
// ADR 0102 decision 2 (owner decision 2026-09-30). The player a
// permanent is about to enter under has two spellings during the
// window, and every door writes both: ev.Actor, which the landing and
// EnteringPermanentChooser read, and Card.Controller on the entering
// card in its source zone, which Authority of the Consuls, Kismet and
// affectedPlayerForEvent read. setEntryControllerLocked writes both
// and remembers what the card carried, so a cancelled entry is put
// back (restoreEntryControllerLocked).
//
// # CR 616.1b
//
// "If any of the replacement and/or prevention effects would modify
// under whose control an object would enter the battlefield, one of
// them must be chosen." ReplacementEffect.ChangesEntryController is
// that tier, and the apply-loop narrows to it before anything else
// (entryControlTier). With the control change applied first, every
// later effect in the window is gathered against the new controller:
// a Xantcha given to Kismet's controller enters untapped.

// PendingChoiceEntryController is "choose the opponent this permanent
// enters under the control of". Answered with {option_index: N} — the
// same payload option_pick takes, routed by kind for the same reason
// (the meaningful value is zero). The options are seats, carried in
// PickOptions with ChoiceOption.Player set, so the seat prune
// (pruneDepartedSeatOptionsLocked) and the wire projection that
// option_pick already has serve this kind unchanged.
//
// It is a kind of its own rather than an option_pick for
// entry_reveal_from_hand's reason: its continuation is a PAUSED EVENT
// (replacementResume), not a card's next sentence, and it resumes
// through the replacement pipeline's contract rather than the option
// chain's.
const PendingChoiceEntryController PendingChoiceKind = "entry_controller"

// EventEntryControllerChosen records the answer: `Actor` chose,
// `Target` is the player the permanent enters under, `CardID` the
// entering card. Narrated (owner decision 2026-09-30): "Alice chose Bob
// to control Captive Audience." Emitted however the answer was reached
// — a prompt, a forced single opponent, or the default an entry that
// cannot pause takes — because the table sees a permanent land under
// somebody who did not cast it either way, and the line is what says
// why.
const EventEntryControllerChosen EventKind = "entry_controller_chosen"

// ControlPurpose is what giving the permanent away does to the player
// who receives it — the card-side declaration a policy reads to know
// whether an opponent is being punished or helped (#780's
// ColorPurpose, for a seat instead of a colour). It rides the wire as
// `control_purpose`. The engine never reads it.
type ControlPurpose string

const (
	// ControlForHarm — the permanent hurts whoever controls it:
	// Captive Audience, Xantcha. A bot gives it to its strongest
	// opponent.
	ControlForHarm ControlPurpose = "harm"
	// ControlForBenefit — the permanent helps whoever controls it:
	// Pendant of Prosperity. A bot gives it to its weakest opponent.
	ControlForBenefit ControlPurpose = "benefit"
)

// EntryControllerChoice is the declaration on ReplacementEffect: "this
// permanent enters under the control of an opponent of your choice".
// Built by effects.EntersUnderTheControlOfAnOpponentOfYourChoice and by
// nothing else.
//
// The effect's own Replace is never called: the choice is the whole of
// what it does to the event.
type EntryControllerChoice struct {
	// Purpose is the card's declaration of what the gift does; see
	// ControlPurpose.
	Purpose ControlPurpose
	// Question is the prompt's header. Falls back to the effect's
	// PromptQuestion and then its Label.
	Question string
}

// entryControlTier is CR 616.1b: when any gathered effect would change
// under whose control the object enters, only those effects are
// candidates on this pass. The rest are gathered again afterwards
// (CR 616.1f), against the new controller.
//
// Returns the input unchanged when no effect is in the tier, or when
// every effect is.
func entryControlTier(applicable []activeReplacement) []activeReplacement {
	n := 0
	for _, a := range applicable {
		if a.effect.ChangesEntryController {
			n++
		}
	}
	if n == 0 || n == len(applicable) {
		return applicable
	}
	out := make([]activeReplacement, 0, n)
	for _, a := range applicable {
		if a.effect.ChangesEntryController {
			out = append(out, a)
		}
	}
	return out
}

// entryControllerCandidatesLocked is the pool: the chooser's opponents
// still in the game (CR 102.3, CR 800.4a), in turn order starting with
// the seat after the chooser. The order is load-bearing — its first
// entry is the default an entry that cannot pause takes (owner
// decision 2026-09-30).
//
// Caller must hold g.mu.
func (g *Game) entryControllerCandidatesLocked(chooser uuid.UUID) []uuid.UUID {
	start := 0
	for i, p := range g.Seats {
		if p != nil && p.ID == chooser {
			start = i + 1
			break
		}
	}
	var out []uuid.UUID
	for k := 0; k < len(g.Seats); k++ {
		p := g.Seats[(start+k)%len(g.Seats)]
		if p == nil || p.ID == chooser || p.Eliminated {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}

// offerEntryControllerLocked handles an applicable entry-controller
// replacement. Returns true when a prompt was queued and the caller
// should bail with errReplacementPending.
//
// It settles inline — no prompt — in four cases, each an owner
// decision or a rule:
//
//   - no eligible opponent. The effect does nothing and the permanent
//     enters under its would-be controller. A live game cannot reach
//     this: the last player standing has already won.
//   - exactly one eligible opponent. The choice is forced (owner
//     decision 3), which is every two-player game.
//   - the entry cannot pause (mustSettleNow, or no resume). The
//     replacement is mandatory, so it is not skipped the way an
//     optional question is: the first opponent in turn order after the
//     chooser is used (owner decision 4).
//   - the chooser has left the game. Same default, for the same reason.
//
// Caller must hold g.mu.
func (g *Game) offerEntryControllerLocked(ev *ReplacementEvent, chosen activeReplacement) bool {
	chooser := g.entryChoicePlayerLocked(ev, chosen)
	candidates := g.entryControllerCandidatesLocked(chooser)
	g.markReplacementAppliedLocked(ev, chosen.id)
	if len(candidates) == 0 {
		return false
	}
	if len(candidates) == 1 || ev.mustSettleNow || !ev.entryResumable || g.chooserGoneLocked(chooser) || chooser == uuid.Nil {
		g.setEntryControllerLocked(ev, chooser, candidates[0])
		return false
	}
	spec := chosen.effect.EntryController
	question := ""
	var purpose ControlPurpose
	if spec != nil {
		question = spec.Question
		purpose = spec.Purpose
	}
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
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:                 PendingChoiceEntryController,
		Chooser:              chooser,
		FromPlayer:           chooser,
		Count:                1,
		Source:               source,
		Reason:               question,
		PickOptions:          g.seatChoiceOptionsLocked(candidates),
		ControlPurpose:       purpose,
		ReplacementEffectIDs: []ReplacementEffectID{chosen.id},
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: []activeReplacement{chosen},
		},
	})
	if id == uuid.Nil {
		// QueueChoiceForEffect refused the chooser (#864). The default
		// stands rather than a paused entry nobody can resume.
		g.setEntryControllerLocked(ev, chooser, candidates[0])
		return false
	}
	return true
}

// settleEntryControllerByDefaultLocked applies an entry-controller
// effect with no question asked — the path applyFirstGatheredLocked
// takes when the window could not be put to anybody. The first eligible
// opponent after the chooser is used (owner decision 4); with none, the
// effect does nothing. The caller has already marked it applied.
//
// Caller must hold g.mu.
func (g *Game) settleEntryControllerByDefaultLocked(ev *ReplacementEvent, chosen activeReplacement) {
	chooser := g.entryChoicePlayerLocked(ev, chosen)
	if candidates := g.entryControllerCandidatesLocked(chooser); len(candidates) > 0 {
		g.setEntryControllerLocked(ev, chooser, candidates[0])
	}
}

// setEntryControllerLocked makes `player` the would-be controller of
// the entering card: ev.Actor, and Card.Controller on the card in its
// source zone (or on the staged token). ADR 0102 decision 2.
//
// The first call on an event remembers what the card carried, so
// restoreEntryControllerLocked can put it back if the entry does not
// happen. The spell's StackItem.Controller is never touched — the
// spell is still its caster's while it resolves (CR 110.2b's
// distinction).
//
// Caller must hold g.mu.
func (g *Game) setEntryControllerLocked(ev *ReplacementEvent, chooser, player uuid.UUID) {
	if ev == nil || player == uuid.Nil {
		return
	}
	if !ev.entryControllerSet {
		ev.entryControllerSet = true
		ev.entryControllerPrior = g.inZoneControllerLocked(ev)
	}
	ev.Actor = player
	g.stampInZoneControllerLocked(ev, player)
	g.EmitEvent(Event{
		Kind:   EventEntryControllerChosen,
		Actor:  chooser,
		Target: player,
		CardID: ev.CardID,
		Label:  g.seatLabelLocked(player),
	})
}

// inZoneControllerLocked is Card.Controller on the entering card where
// it sits now — its source zone, or the staged token list.
//
// Caller must hold g.mu.
func (g *Game) inZoneControllerLocked(ev *ReplacementEvent) uuid.UUID {
	if z := g.findCardZoneLocked(ev.CardID); z != nil {
		for i := range z.Cards {
			if z.Cards[i].InstanceID == ev.CardID {
				return z.Cards[i].Controller
			}
		}
		return uuid.Nil
	}
	if tok, ok := g.enteringTokenLocked(ev.CardID); ok {
		return tok.Controller
	}
	return uuid.Nil
}

// stampInZoneControllerLocked writes Card.Controller on the entering
// card in its source zone, or on the staged token. A card already on
// the battlefield is left alone: its entry has already happened.
//
// Caller must hold g.mu.
func (g *Game) stampInZoneControllerLocked(ev *ReplacementEvent, player uuid.UUID) {
	if z := g.findCardZoneLocked(ev.CardID); z != nil {
		if z == g.Battlefield {
			return
		}
		for i := range z.Cards {
			if z.Cards[i].InstanceID == ev.CardID {
				z.Cards[i].Controller = player
				return
			}
		}
		return
	}
	for i := range g.enteringTokens {
		if g.enteringTokens[i].InstanceID == ev.CardID {
			g.enteringTokens[i].Controller = player
			return
		}
	}
}

// restoreEntryControllerLocked puts back what the entering card
// carried before an entry-controller effect re-stamped it, when the
// entry did not happen — the window cancelled or redirected it, or its
// prompt was taken away. A no-op for every event no such effect
// touched, and for a card of a simultaneous entry, whose batch restores
// its own priorController (entry_batch.go).
//
// Caller must hold g.mu.
func (g *Game) restoreEntryControllerLocked(ev *ReplacementEvent) {
	if ev == nil || !ev.entryControllerSet {
		return
	}
	if ev.entryTail != nil && ev.entryTail.batch != nil {
		return
	}
	g.stampInZoneControllerLocked(ev, ev.entryControllerPrior)
}

// ResolveEntryController answers a PendingChoiceEntryController.
// `index` is the offset into the prompt's option list; the seat is read
// off the option the chooser picked (#994's rule — the list may have
// been pruned since it was built, so only the option itself is a stable
// answer).
//
// An out-of-range index is refused with the prompt still open. Once
// answered, the apply-loop is re-entered so CR 616.1f gathers anything
// the new controller made applicable, and the settled event goes
// through the shared finisher, which lands the permanent.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveEntryController(choiceID, chooserID uuid.UUID, index int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceEntryController {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if index < 0 || index >= len(choice.PickOptions) {
		return ErrInvalidParam
	}
	seat := choice.PickOptions[index].Player
	if seat == uuid.Nil || g.chooserGoneLocked(seat) {
		return ErrInvalidParam
	}
	frame := choice.replacementResume
	g.dequeueChoiceLocked(idx)
	if frame == nil || frame.ev == nil {
		return ErrInvalidParam
	}
	if g.dropStaleReplacementResumeLocked(frame) {
		return nil
	}
	ev := frame.ev
	g.setEntryControllerLocked(ev, chooserID, seat)
	return g.resumeEntryControllerEventLocked(ev)
}

// resumeEntryControllerEventLocked re-enters the apply-loop for an
// event whose entry-controller question has just been settled, and
// lands it through the shared finisher.
//
// Caller must hold g.mu.
func (g *Game) resumeEntryControllerEventLocked(ev *ReplacementEvent) error {
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	return g.finishSettledReplacementLocked(ev, out)
}

// settleEntryControllerWithoutChoiceLocked finishes an open
// entry-controller prompt whose every offered seat has left the game
// while it was open (pruneDepartedSeatOptionsLocked). The effect has
// nobody left to give the permanent to, so it does nothing and the
// permanent enters under its would-be controller — the no-opponent
// rule, reached late. The prompt must already be out of the queue.
//
// Caller must hold g.mu.
func (g *Game) settleEntryControllerWithoutChoiceLocked(c *PendingChoice) {
	if c == nil || c.replacementResume == nil || c.replacementResume.ev == nil {
		return
	}
	frame := c.replacementResume
	if g.dropStaleReplacementResumeLocked(frame) {
		return
	}
	if err := g.resumeEntryControllerEventLocked(frame.ev); err != nil {
		g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
	}
}
