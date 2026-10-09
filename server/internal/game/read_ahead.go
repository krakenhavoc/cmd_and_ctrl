package game

import (
	"errors"
	"strconv"

	"github.com/google/uuid"
)

// read_ahead.go — read ahead (CR 702.155), #2123.
//
//	CR 702.155a  "Read ahead" means "Chapter abilities of this Saga
//	             can't trigger the turn it entered the battlefield unless
//	             it has exactly the number of lore counters on it
//	             specified in the chapter symbol of that ability."
//	CR 702.155b  Each Saga with read ahead has the intrinsic abilities
//	             "As this Saga enters, choose a number between one and
//	             this Saga's final chapter number" and "This Saga enters
//	             with the chosen number of lore counters on it."
//	CR 702.155c  Multiple instances of read ahead on the same object are
//	             redundant.
//	CR 714.3b    Those abilities create replacement effects (CR 614.1c).
//
// # The entry choice
//
// A keyword entry replacement, riot's shape (riot.go): the CR 614.12
// look-ahead (entry_lookahead.go) reports whether the permanent would
// have read ahead and its final chapter number as it would exist on the
// battlefield, and the gather adds ONE replacement for it (702.155c).
// Its question is PendingChoiceEntryReadAhead, asked before the Saga
// enters (CR 614.12a); the answer seeds ReplacementEvent.EntersWithCounters
// with that many lore counters, which the landing drains through the
// ordinary counter window — so Doubling Season doubles them, as it
// doubles any "enters with" counters.
//
// So the choice is made however the Saga enters: cast, put onto the
// battlefield by an effect, or a token copy of a read-ahead Saga
// (702.155b is an ability of the permanent, not of the spell).
//
// An entry that cannot pause, or whose chooser has gone, takes 1: the
// lore count a Saga without read ahead enters with (CR 714.3a), and the
// answer that skips nothing. A Saga whose final chapter is 1 is asked
// nothing, because 1 is the only number it could choose.
//
// A Saga the catalog has no chapters for has a final chapter number of
// 0 here (SagaFinalChapter), so there is no number to choose: read
// ahead adds nothing to its entry, and sagaEntersWithLoreCounterLocked
// gives it the one sandbox lore counter every unknown Saga gets.
//
// # The chapter rule
//
// fireSagaChaptersLocked (sagas.go) asks readAheadRestrictsLocked: on
// the turn the Saga entered, a chapter fires only if the Saga now has
// exactly that chapter's number of lore counters. So a Saga that enters
// with three counters fires chapter III and nothing before it, a
// proliferate the same turn fires the chapter it lands on exactly, and
// a Saga doubled past its final chapter fires nothing at all.

// KeywordReadAhead is CR 702.155's canonical token.
const KeywordReadAhead = "read ahead"

// PendingChoiceEntryReadAhead is read ahead's question: "which chapter
// does this Saga start on?" Answered with {option_index: N}, the
// payload option_pick and entry_controller take, routed by kind for
// their reason (the meaningful value is zero): option N is chapter N+1.
// The options are the chapters, in order, carried in PickOptions.
//
// A kind of its own rather than an option_pick for entry_controller's
// reason: its continuation is a paused entry (replacementResume), and
// it resumes through the replacement pipeline, not an option chain.
//
// The chooser is the entering Saga's would-be controller.
const PendingChoiceEntryReadAhead PendingChoiceKind = "entry_read_ahead"

// readAheadReplacementID is read ahead's replacement ID: the fifth
// stride of the self-replacement range, above riot's and unleash's.
// One ID, because the instances are redundant (CR 702.155c).
const readAheadReplacementID = selfReplacementIDBase + 4*MaxCatalogReplacementSlots

// readAheadReplacement is read ahead's entry replacement. Its Replace is
// the default answer, one lore counter; the prompt's answer writes the
// chosen count itself (ResolveEntryReadAhead).
func readAheadReplacement(name string) ReplacementEffect {
	return ReplacementEffect{
		Watches:         []EventKind{EventZoneMove},
		Replace:         addEntryLoreCounter,
		Controller:      entryKeywordController,
		SelfReplacement: true,
		PromptQuestion:  "Read ahead — choose the chapter " + name + " starts on",
		Label:           "Read ahead",
		entryKeyword:    KeywordReadAhead,
	}
}

// addEntryLoreCounter is read ahead's default: the Saga enters with one
// lore counter.
func addEntryLoreCounter(ev *ReplacementEvent, _ *Game, _ *Card) error {
	ev.AddCounterAtETB(CounterLore, 1)
	return nil
}

// gatherReadAheadReplacementLocked appends read ahead's replacement when
// the look-ahead says the entering permanent would have read ahead and a
// final chapter to choose up to.
//
// Caller must hold g.mu (write).
func (g *Game) gatherReadAheadReplacementLocked(ev *ReplacementEvent, applied map[ReplacementEffectID]bool, out []activeReplacement) []activeReplacement {
	if applied[readAheadReplacementID] {
		return out
	}
	la := g.entryLookAheadLocked(ev)
	if !la.readAhead || la.finalChapter <= 0 {
		return out
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return out
	}
	src := entering
	return append(out, activeReplacement{effect: readAheadReplacement(src.Name), source: &src, id: readAheadReplacementID})
}

// RomanChapter renders a chapter number as the Roman numeral a Saga
// prints (CR 714.2a). Past the numerals a printed Saga uses, it falls
// back to the digits.
func RomanChapter(n int) string {
	numerals := []string{"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X"}
	if n >= 1 && n < len(numerals) {
		return numerals[n]
	}
	return strconv.Itoa(n)
}

// readAheadOptions is the prompt's option list: one per chapter, in
// order, so option i is chapter i+1.
func readAheadOptions(final int) []ChoiceOption {
	out := make([]ChoiceOption, 0, final)
	for n := 1; n <= final; n++ {
		out = append(out, ChoiceOption{Label: "Chapter " + RomanChapter(n)})
	}
	return out
}

// offerEntryReadAheadLocked asks read ahead's question, or settles it on
// one lore counter when there is nothing to ask (a final chapter of 1)
// or nobody to ask it of: the entry cannot pause, or its chooser has
// left or cannot be named. Returns true when the prompt was queued and
// the caller must bail.
//
// Caller must hold g.mu.
func (g *Game) offerEntryReadAheadLocked(ev *ReplacementEvent, chosen activeReplacement) bool {
	final := g.entryLookAheadLocked(ev).finalChapter
	chooser := g.entryChoicePlayerLocked(ev, chosen)
	if final <= 1 || ev.mustSettleNow || !g.optionalReplacementResumableLocked(ev) || chooser == uuid.Nil || g.chooserGoneLocked(chooser) {
		g.settleReadAheadOnChapterOneLocked(ev, chosen)
		return false
	}
	var source uuid.UUID
	if chosen.source != nil {
		source = chosen.source.InstanceID
	}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:                 PendingChoiceEntryReadAhead,
		Chooser:              chooser,
		FromPlayer:           chooser,
		Count:                1,
		Source:               source,
		Reason:               chosen.effect.PromptQuestion,
		PickOptions:          readAheadOptions(final),
		ReplacementEffectIDs: []ReplacementEffectID{chosen.id},
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: []activeReplacement{chosen},
		},
	})
	if id == uuid.Nil {
		// QueueChoiceForEffect refused the chooser (#864): the default
		// stands rather than a paused entry nobody can resume.
		g.settleReadAheadOnChapterOneLocked(ev, chosen)
		return false
	}
	return true
}

// settleReadAheadOnChapterOneLocked applies read ahead's default: one
// lore counter.
//
// Caller must hold g.mu.
func (g *Game) settleReadAheadOnChapterOneLocked(ev *ReplacementEvent, chosen activeReplacement) {
	g.markReplacementAppliedLocked(ev, chosen.id)
	if chosen.effect.Replace != nil && !ev.Canceled {
		if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
	}
}

// ResolveEntryReadAhead answers a PendingChoiceEntryReadAhead: `index`
// is the offset into the prompt's chapters, so the Saga enters with
// index+1 lore counters. The replacement is marked applied (CR 614.5),
// the apply-loop is re-entered so CR 616.1f gathers anything left, and
// the settled entry is landed.
//
// An out-of-range index is refused with the prompt still open.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveEntryReadAhead(choiceID, chooserID uuid.UUID, index int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceEntryReadAhead {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if index < 0 || index >= len(choice.PickOptions) {
		return ErrInvalidParam
	}
	frame := choice.replacementResume
	g.dequeueChoiceLocked(idx)
	if frame == nil || frame.ev == nil || len(frame.applicable) == 0 {
		return ErrInvalidParam
	}
	if g.dropStaleReplacementResumeLocked(frame) {
		return nil
	}
	ev := frame.ev
	chosen := frame.applicable[0]
	g.markReplacementAppliedLocked(ev, chosen.id)
	if !ev.Canceled {
		ev.AddCounterAtETB(CounterLore, index+1)
	}
	// Said out loud, as an "as this enters, choose" answer is: "Ian
	// chose chapter III for The Cruelty of Gix".
	g.EmitEvent(Event{
		Kind:   EventOptionChosen,
		Actor:  chooserID,
		CardID: ev.CardID,
		Label:  "chapter " + RomanChapter(index+1),
	})
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

// readAheadRestrictsLocked reports whether CR 702.155a's restriction
// is in force for this Saga right now: it has read ahead, and it
// entered the battlefield this turn. While it is, a chapter triggers
// only if the Saga has exactly that chapter's number of lore counters.
//
// "Entered this turn" is the turn tally (EnteredThisTurn) or, while the
// Saga's "enters with" counters are being put on it, the landing itself:
// those counters land after the move and before EventETB, which is what
// the tally counts.
//
// Caller must hold g.mu.
func (g *Game) readAheadRestrictsLocked(sagaID uuid.UUID) bool {
	if sagaID == uuid.Nil {
		return false
	}
	if g.landingEntryCounters != sagaID && !g.EnteredThisTurn(sagaID) {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	return HasKeyword(findBattlefieldCard(g, sagaID), KeywordReadAhead)
}
