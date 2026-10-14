package game

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// entry_card_choice.go — a CARD choice inside an entry replacement
// (CR 614.1a / 614.1c / 614.12a / 614.13). Three printed shapes, one
// mechanism:
//
//	"As this land enters, you may reveal an Island or Swamp card from
//	 your hand. If you don't, this land enters tapped."   (the ten
//	 reveal-lands, #1198, ADR 0013 §5z)
//	"If this artifact would enter, you may discard a land card instead.
//	 If you do, put this artifact onto the battlefield. If you don't,
//	 put it into its owner's graveyard."                   (Mox Diamond,
//	 #1744, ADR 0098)
//	"If this land would enter, sacrifice a Forest instead. If you do,
//	 put this land onto the battlefield. If you don't, put it into its
//	 owner's graveyard."                                   (Heart of
//	 Yavimaya and six more, ADR 0098 Decision 11)
//
// entry_choice.go's sibling, and deliberately its mirror image: the
// same replacement with a decision inside it, the same pre-push pause,
// the same resume — with a CARD where the shockland has a number. The
// declaration's Replace is always the "you didn't" branch (enters
// tapped; goes to the graveyard); naming cards is what AVOIDS it. That
// inversion is EntryLifeCost's, and it is why declineIsReplace
// (replacements.go) lists both.
//
// ADR 0170 adds a fourth shape: a choice of OTHER permanents that the
// engine does nothing to. The declaration's Then acts on them instead:
//
//	"As this creature enters, put a phylactery counter on an artifact
//	 you control."                                         (Phylactery
//	 Lich, #1893)
//	"As Sin enters, remove all counters from any number of artifacts,
//	 creatures, and enchantments. Sin enters with X +1/+1 counters on
//	 it, where X is twice the number of counters removed this way."
//	                                                       (Sin,
//	 Unending Cataclysm)
//
// Everything here runs PRE-push, so what is replaced is the
// permanent's entry itself: a Mox Diamond that is not paid for never
// enters (its 2008-05-01 ruling), triggers nothing that watches an
// entry, and is never tappable in response.
//
// # What the action does
//
//   - Reveal moves nothing (CR 701.20b). The card is still in hand when
//     the land finishes entering.
//   - Discard goes through discardCardsLocked with DiscardCauseEffect
//     (CR 701.9a; Library of Leng's 2004-10-04 ruling: its replacement
//     applies "any time a spell or ability has you discard as part of
//     its effect"). So RepEventDiscard opens and EventDiscardCard fires.
//   - Sacrifice goes through SacrificeAllThenForEffect: one simultaneous
//     exit, EventSacrifice for each permanent.
//
// A discard or a sacrifice can PAUSE — Library of Leng's "may", an
// ordering between two discard replacements, CR 903.9 on a commander.
// The paused ENTRY is carried across that pause in a frozen copy
// (frozenEntryChoice), cloned again each time the continuation runs,
// so an undo into the inner prompt replays the entry rather than
// sharing it with the live game (ADR 0098 Decision 3).

// PendingChoiceEntryRevealFromHand is the "as this enters, you may
// reveal a card from your hand" prompt (#1198).
//
// The three entry kinds carry the CHOOSE-CARDS payload — {card_ids},
// ChooseMin / ChooseMax, and a chooseCardsFrame pinned to a zone — so
// checkChooseCardsPicksLocked (#1017) stays the one copy of "would
// this answer be accepted", shared with the submit path and with the
// enumerator through ChooseCardsPickLegalLocked. All three are
// isCardSetPickKind members.
//
// They are separate KINDS rather than one kind with a flag, because
// the SIGN differs for a bot: a revealed card is kept, so any reveal
// beats none; a discarded or sacrificed card is spent. The client's
// sentence differs too, and the kind is what it renders from.
const PendingChoiceEntryRevealFromHand PendingChoiceKind = "entry_reveal_from_hand"

// PendingChoiceEntryDiscardFromHand is "if this would enter, you may
// discard <a card> instead" (Mox Diamond, ADR 0098). Its floor is
// zero: an empty answer is the decline, which puts the permanent into
// its owner's graveyard.
const PendingChoiceEntryDiscardFromHand PendingChoiceKind = "entry_discard_from_hand"

// PendingChoiceEntrySacrifice is "if this would enter, sacrifice
// <N permanents> instead" (Heart of Yavimaya, Lotus Vale, ADR 0098
// Decision 11). It is not a "may": its floor and ceiling are both N,
// and the only way not to sacrifice is to be unable to, which the
// offer settles inline without asking. Its candidates are public
// permanents, so it is not redacted.
const PendingChoiceEntrySacrifice PendingChoiceKind = "entry_sacrifice"

// PendingChoiceEntryChoosePermanents is "as this enters, <do something
// to> <permanents you choose>" (ADR 0170, #1893): Phylactery Lich's
// artifact, Sin's any number of artifacts, creatures and enchantments.
// The engine does nothing to what is named; the declaration's Then
// does, before the permanent is on the battlefield. Its candidates are
// public permanents, so it is not redacted, and ChooseSuggested carries
// the declaration's recommended answer for the client and the bot.
const PendingChoiceEntryChoosePermanents PendingChoiceKind = "entry_choose_permanents"

// isEntryCardChoiceKind reports the four kinds whose continuation is
// a paused CR 614 ENTRY. The stale-candidate prune and the chained
// resolver both leave them alone: dropping one would strand the entry
// it is pausing, and none of them can need the prune (see
// pruneCardSetChoicesLocked).
func isEntryCardChoiceKind(kind PendingChoiceKind) bool {
	switch kind {
	case PendingChoiceEntryRevealFromHand, PendingChoiceEntryDiscardFromHand, PendingChoiceEntrySacrifice,
		PendingChoiceEntryChoosePermanents:
		return true
	}
	return false
}

// EntryCardAction is what an EntryCardChoice does with the cards it
// names. The zero value is a reveal, so the ten reveal-lands declare
// nothing new.
type EntryCardAction string

const (
	// EntryCardReveal shows a card from the chooser's hand
	// (CR 701.20). Nothing moves.
	EntryCardReveal EntryCardAction = ""
	// EntryCardDiscard discards cards from the chooser's hand
	// (CR 701.9a), as an effect.
	EntryCardDiscard EntryCardAction = "discard"
	// EntryCardSacrifice sacrifices permanents the chooser controls
	// (CR 701.21a), as an effect.
	EntryCardSacrifice EntryCardAction = "sacrifice"
	// EntryCardChoose names permanents on the battlefield and does
	// nothing to them itself (ADR 0170). Then is the whole effect
	// ("put a phylactery counter on", "remove all counters from").
	// There is no "if you don't": Replace is normally nil, and an empty
	// answer (or no candidates) skips Then.
	EntryCardChoose EntryCardAction = "choose"
)

// EntryCardChoice is the declaration on ReplacementEffect: "as this
// permanent enters (or: if it would enter), <Action> <Min..Max cards
// matching Matches>; if you don't, <Replace>."
//
// The inversion is EntryLifeCost's and it is why this is not
// Optional: an Optional "yes" APPLIES the replacement, and here naming
// cards is what AVOIDS it.
type EntryCardChoice struct {
	// Action is what happens to the named cards. The zero value is a
	// reveal.
	Action EntryCardAction

	// Matches is the clause's filter — "an Island or Swamp card",
	// "a land card", "an untapped Mountain". Nil admits every
	// candidate.
	//
	// A function of the CARD and of nothing else, which is
	// ChooseCardsPrompt.Validate's rule for its reason: the predicate
	// runs on the submit path under the write lock and, through
	// ChooseCardsPickLegalLocked, inside legal.EnumerateFor under the
	// READ lock. It must not write through the Card it is handed and
	// must not capture a *Game or a pointer into a zone.
	//
	// For a reveal or a discard it reads a card in a HAND, where no
	// layer has been applied, so it sees printed characteristics. For
	// a sacrifice it reads a permanent, so Card.IsLand / HasSubtype
	// answer from its effective characteristics: an Urborg-made Swamp
	// is a Swamp for Lake of the Dead.
	Matches func(c Card) bool

	// Min and Max bound the answer. A reveal or a discard is "you MAY
	// … a card": 0..1, and the zero floor makes the empty answer the
	// enumerator's AlwaysLegal move. A sacrifice is not a "may": Min
	// == Max == N, and with fewer than Min candidates the offer applies
	// Replace without asking (Lotus Vale with one untapped land).
	//
	// Max <= 0 is read as 1.
	Min, Max int

	// Question is the prompt's header — the card's own sentence.
	// Falls back to ReplacementEffect.PromptQuestion and then Label.
	Question string

	// Then, when set, runs after the action has happened and before
	// the apply-loop is re-entered, with the cards it named in the
	// order the chooser submitted them. Never called for a decline.
	// For a discard or a sacrifice it is handed the cards that really
	// left (discardedThisWayLocked / sacrificedThisWayLocked).
	//
	// `ev` is the paused ENTRY (ADR 0170). Then may change how the
	// permanent enters (ev.AddCounterAtETB for "it enters with X +1/+1
	// counters"), and since nothing has seen the permanent yet, those
	// counters are on it as it arrives (CR 614.1c, 614.12).
	//
	// Then runs while the entry is paused, so whatever it does to OTHER
	// objects must settle inline: use the MustSettleNow forms
	// (AddCounterByMustSettleNowForEffect). A primitive that can pause
	// here would let the entry finish before its own prompt is
	// answered.
	//
	// Runs with g.mu held.
	Then func(g *Game, ev *ReplacementEvent, picked []uuid.UUID) error

	// AnyController widens an EntryCardChoose's candidates from the
	// permanents the chooser controls to every permanent on the
	// battlefield ("any number of artifacts, creatures, and
	// enchantments"). Only meaningful for EntryCardChoose.
	AnyController bool

	// SuggestKey, when set, names an EntryChoiceSuggestion filed with
	// RegisterEntryChoiceSuggestion: the declaration's recommended
	// answer to an EntryCardChoose prompt. It rides the prompt as
	// ChooseSuggested (proliferate's field, #2525): the client
	// pre-selects it, the enumerator always offers it, and the
	// heuristic bot answers with it. A default, never a rule.
	//
	// A KEY rather than a func (state triggers' reason, ADR 0107 §1):
	// the declaration gains no closure for the ADR 0041 ratchet.
	SuggestKey string

	// AnyNumber makes this a sacrifice of "any number of" matching
	// permanents, zero included (CR 702.82a, devour): the floor is Min
	// (normally 0), the ceiling is however many candidates exist, and
	// "if you do" does not apply — whatever really left is what counts.
	// An empty answer, or no candidates, declines without running
	// Replace beyond whatever the declaration set (devour sets none).
	// Only meaningful for EntryCardSacrifice.
	AnyNumber bool

	// Devour is CR 702.82a's N: when positive, every creature really
	// sacrificed adds N +1/+1 counters to the ENTRY (ev.AddCounterAtETB,
	// so Doubling Season and Hardened Scales see them) and one to
	// ReplacementEvent.EntersDevoured, which the landing copies onto
	// Card.Devoured for CR 702.82b. It is data, not a hook, so the
	// declaration carries no closure a restore point would have to
	// rebuild. Only meaningful with AnyNumber.
	Devour int

	// DevourDraw and DevourLife are what a creature's own "for each
	// creature it devoured" ability pays per sacrificed creature
	// (Skullmulcher draws one, Marrow Chomper gains two). They do
	// nothing in the engine, whose trigger reads Card.Devoured; they
	// are on the declaration so the prompt can say what each creature
	// is worth (PendingChoice.DevourOffer) to a seat that cannot read
	// the entering card's text. Only meaningful with Devour.
	DevourDraw, DevourLife int
}

// zone is where the choice's candidates live.
func (s *EntryCardChoice) zone() ZoneKind {
	if s.Action == EntryCardSacrifice || s.Action == EntryCardChoose {
		return ZoneBattlefield
	}
	return ZoneHand
}

// kind is the prompt kind the choice queues.
func (s *EntryCardChoice) kind() PendingChoiceKind {
	switch s.Action {
	case EntryCardDiscard:
		return PendingChoiceEntryDiscardFromHand
	case EntryCardSacrifice:
		return PendingChoiceEntrySacrifice
	case EntryCardChoose:
		return PendingChoiceEntryChoosePermanents
	}
	return PendingChoiceEntryRevealFromHand
}

// bounds normalises Min and Max against the candidate count.
func (s *EntryCardChoice) bounds(candidates int) (lo, hi int) {
	hi = s.Max
	if hi <= 0 {
		hi = 1
	}
	lo = s.Min
	if lo < 0 {
		lo = 0
	}
	if (s.Action == EntryCardSacrifice || s.Action == EntryCardChoose) && s.AnyNumber {
		return lo, candidates
	}
	if s.Action == EntryCardSacrifice {
		// Not a "may": the count is fixed, and a shortfall is decided
		// by the offer before any prompt exists.
		return lo, hi
	}
	if hi > candidates {
		hi = candidates
	}
	if lo > hi {
		lo = hi
	}
	return lo, hi
}

// offerEntryCardChoiceLocked handles an applicable replacement whose
// decline condition is "you didn't <reveal / discard / sacrifice>".
// Returns true when a prompt was queued and the caller should bail
// with errReplacementPending.
//
// Returns false when no prompt is possible, and applies Replace — the
// "you didn't" branch — inline. The cases are
// offerEntryLifePaymentLocked's one for one:
//
//   - nothing the clause admits, or (for a sacrifice) fewer than it
//     needs. Asking a question whose only answer is "no" is worse than
//     not asking.
//   - the entry can't be resumed (ReplacementEvent.entryResumable) —
//     since #1322 only the sandbox move_card verb. ADR 0098 owner
//     decision 6: a Mox Diamond dragged onto the battlefield goes to
//     its owner's graveyard, as a reveal-land dragged there enters
//     tapped.
//   - the chooser can't be identified, or has left the game
//     (CR 800.4a).
//
// Caller must hold g.mu.
func (g *Game) offerEntryCardChoiceLocked(ev *ReplacementEvent, chosen activeReplacement) bool {
	spec := chosen.effect.EntryCardChoice
	chooser := g.entryChoicePlayerLocked(ev, chosen)
	if spec != nil && ev.entryResumable && !g.chooserGoneLocked(chooser) {
		candidates := g.entryChoiceCandidatesLocked(ev, spec, chooser)
		lo, _ := spec.bounds(len(candidates))
		if len(candidates) > 0 && len(candidates) >= lo {
			if g.queueEntryCardChoicePromptLocked(ev, chosen, chooser, candidates) {
				return true
			}
		}
	}
	g.markReplacementAppliedLocked(ev, chosen.id)
	g.runEntryChoiceDeclineLocked(ev, chosen)
	return false
}

// runEntryChoiceDeclineLocked runs the effect's Replace — the "you
// didn't" branch — and logs rather than returns its error, the way
// every inline Replace in the apply-loop does.
//
// Caller must hold g.mu.
func (g *Game) runEntryChoiceDeclineLocked(ev *ReplacementEvent, chosen activeReplacement) {
	if chosen.effect.Replace == nil {
		return
	}
	if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
		g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
	}
}

// entryChoiceCandidatesLocked is every card the clause admits, in zone
// order: the chooser's hand for a reveal or a discard, the permanents
// the chooser controls for a sacrifice or a choose, and every permanent
// for a choose with AnyController (ADR 0170).
//
// CR 614.13a: "You can't choose the object that will become that
// permanent or any other object entering the battlefield at the same
// time as that object." So the entering card and every other member of
// its simultaneous entry are excluded. (A land put from a hand together
// with Mox Diamond cannot be the land it discards.)
//
// The list is FROZEN here and re-checked against the live zone on
// submit (pickStillInPickZoneLocked). Nothing can move a candidate
// while the prompt is open — an open choice stops priority — so the
// re-check is the backstop rather than the mechanism.
//
// Caller must hold g.mu.
func (g *Game) entryChoiceCandidatesLocked(ev *ReplacementEvent, spec *EntryCardChoice, chooser uuid.UUID) []uuid.UUID {
	entering := map[uuid.UUID]bool{ev.CardID: true}
	if ev.entryTail != nil && ev.entryTail.batch != nil {
		for _, m := range ev.entryTail.batch.members {
			entering[m.cardID] = true
		}
	}
	var pool []Card
	if spec.zone() == ZoneBattlefield {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == chooser || (spec.Action == EntryCardChoose && spec.AnyController) {
				pool = append(pool, c)
			}
		}
	} else if p := g.playerByIDLocked(chooser); p != nil && p.Hand != nil {
		pool = p.Hand.Cards
	}
	out := make([]uuid.UUID, 0, len(pool))
	for _, c := range pool {
		if entering[c.InstanceID] {
			continue
		}
		if spec.Matches != nil && !spec.Matches(c) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// queueEntryCardChoicePromptLocked queues the pick and stashes the
// continuation frame the resume path re-enters with. Reports whether
// anything was queued — QueueChoiceForEffect refuses a chooser who has
// left the game (#864), and a caller that believed it would leave the
// entry paused forever.
//
// Caller must hold g.mu.
func (g *Game) queueEntryCardChoicePromptLocked(
	ev *ReplacementEvent,
	chosen activeReplacement,
	chooser uuid.UUID,
	candidates []uuid.UUID,
) bool {
	spec := chosen.effect.EntryCardChoice
	lo, hi := spec.bounds(len(candidates))
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
	var suggested []uuid.UUID
	if spec.Action == EntryCardChoose {
		if suggest, ok := EntryChoiceSuggestionFor(spec.SuggestKey); ok {
			suggested = keepCandidates(suggest(g, chooser, append([]uuid.UUID(nil), candidates...)), candidates)
		}
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:    spec.kind(),
		Chooser: chooser,
		// FromPlayer is the chooser too — the hand, or the permanents,
		// are their own. It is what filterPendingChoices and the
		// departure sweep read to tell a question about somebody
		// else's material from one about the asker's own (#961).
		FromPlayer:           chooser,
		Count:                hi,
		Source:               source,
		Reason:               question,
		ChooseCards:          append([]uuid.UUID(nil), candidates...),
		ChooseMin:            lo,
		ChooseMax:            hi,
		ChooseSuggested:      suggested,
		ReplacementEffectIDs: []ReplacementEffectID{chosen.id},
		chooseCardsResume: &chooseCardsFrame{
			// The zone, and no `then`: the continuation of this
			// prompt is a paused replacement event, and it is run by
			// ResolveEntryCardChoice rather than by
			// resolveCardSetPick. The frame is here for the live-zone
			// re-check and so that ChooseCardsPickLegalLocked reaches
			// the same rule the resolver does.
			zone: spec.zone(),
		},
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: []activeReplacement{chosen},
		},
	}) != uuid.Nil
}

// EntryChoiceSuggestion is an EntryCardChoose declaration's recommended
// answer: a subset of `candidates` for `chooser`. Called once, as the
// prompt is queued, under the write lock. It may read the game and must
// not change it.
type EntryChoiceSuggestion func(g *Game, chooser uuid.UUID, candidates []uuid.UUID) []uuid.UUID

var entryChoiceSuggestions struct {
	sync.RWMutex
	m map[string]EntryChoiceSuggestion
}

// RegisterEntryChoiceSuggestion files `fn` under `key` for an
// EntryCardChoice.SuggestKey to name. Call it at init, from a card's
// declaration. Panics on an empty key, a nil func or a key already
// taken, as RegisterStateCondition does.
func RegisterEntryChoiceSuggestion(key string, fn EntryChoiceSuggestion) {
	if key == "" {
		panic("game: an entry-choice suggestion needs a key")
	}
	if fn == nil {
		panic(fmt.Sprintf("game: entry-choice suggestion %q has no function", key))
	}
	entryChoiceSuggestions.Lock()
	defer entryChoiceSuggestions.Unlock()
	if entryChoiceSuggestions.m == nil {
		entryChoiceSuggestions.m = make(map[string]EntryChoiceSuggestion)
	}
	if _, dup := entryChoiceSuggestions.m[key]; dup {
		panic(fmt.Sprintf("game: entry-choice suggestion %q is already registered", key))
	}
	entryChoiceSuggestions.m[key] = fn
}

// EntryChoiceSuggestionFor is the suggestion registered under key.
// False for an empty or unknown key, which suggests nothing.
func EntryChoiceSuggestionFor(key string) (EntryChoiceSuggestion, bool) {
	if key == "" {
		return nil, false
	}
	entryChoiceSuggestions.RLock()
	defer entryChoiceSuggestions.RUnlock()
	fn, ok := entryChoiceSuggestions.m[key]
	return fn, ok
}

// keepCandidates is `ids` with every member that is not in `candidates`
// dropped, and duplicates dropped, in the order `ids` gave. A
// declaration's suggestion can only recommend what the prompt offers.
func keepCandidates(ids, candidates []uuid.UUID) []uuid.UUID {
	in := make(map[uuid.UUID]bool, len(candidates))
	for _, id := range candidates {
		in[id] = true
	}
	var out []uuid.UUID
	for _, id := range ids {
		if in[id] {
			out = append(out, id)
			delete(in, id)
		}
	}
	return out
}

// ResolveEntryRevealFromHand is ResolveEntryCardChoice under the name
// #1198 gave it.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveEntryRevealFromHand(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	return g.ResolveEntryCardChoice(choiceID, chooserID, picks)
}

// ResolveEntryCardChoice processes a resolve_choice action for any of
// the four entry card-choice kinds. `picks` are the cards the chooser
// names:
//
//	none        — declined; the replacement fires (enters tapped, or
//	              goes to its owner's graveyard). Legal only where the
//	              floor is zero.
//	one or more — revealed, discarded or sacrificed; the replacement
//	              fires only if a discard or sacrifice did not really
//	              happen (ADR 0098 owner decision 5: "if you do" is
//	              "the card left").
//
// Validation runs BEFORE the dequeue, so a client that submits a card
// that has left the zone gets an error and can try again rather than
// losing the prompt — ResolveChooseCards' contract, through the same
// function.
//
// Either way the apply-loop is re-entered so CR 616.1f can pick up
// anything newly applicable, and the settled event is then pushed
// through the shared finisher, which puts the permanent onto the
// battlefield or — if Replace redirected it — wherever it now goes.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveEntryCardChoice(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if !isEntryCardChoiceKind(choice.Kind) {
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
	spec := chosen.effect.EntryCardChoice
	// The decision is made once per event either way, so the effect
	// is marked applied before anything else — an answer must not
	// leave the effect eligible to be gathered again on the next
	// iteration (CR 614.5).
	g.markReplacementAppliedLocked(ev, chosen.id)

	switch {
	case len(picks) == 0:
		g.runEntryChoiceDeclineLocked(ev, chosen)
	case spec != nil && spec.Action == EntryCardDiscard:
		frozen := frozenEntryChoice(ev, chosen)
		return g.discardCardsLocked(chooserID, append([]uuid.UUID(nil), picks...), discardOptions{
			cause:  DiscardCauseEffect,
			source: source,
			then: func(g *Game, landed []uuid.UUID) error {
				return g.resumeEntryAfterActionLocked(frozen, len(landed) == len(picks), landed)
			},
		})
	case spec != nil && spec.Action == EntryCardSacrifice:
		frozen := frozenEntryChoice(ev, chosen)
		return g.SacrificeAllThenForEffect(source, append([]uuid.UUID(nil), picks...), func(g *Game, sacrificed []uuid.UUID) error {
			return g.resumeEntryAfterActionLocked(frozen, len(sacrificed) == len(picks), sacrificed)
		})
	case spec != nil && spec.Action == EntryCardChoose:
		// ADR 0170: nothing happens to the named permanents except what
		// Then does to them, inline, before the entry resumes.
		g.runEntryThenLocked(spec, ev, picks)
	default:
		// CR 701.20: the table sees them and is entitled to remember.
		// Through the one reveal primitive, so the knower set and the
		// grouped EventRevealCards run are the ones every other
		// reveal in the engine produces.
		g.RevealForEffect(RevealSpec{
			Player: chooserID,
			Source: source,
			Reason: reason,
			Cards:  picks,
		})
		if spec != nil {
			g.runEntryThenLocked(spec, ev, picks)
		}
	}
	return g.continueEntryAfterChoiceLocked(ev)
}

// frozenEntryChoice takes an immutable copy of a paused entry and the
// effect that paused it, for a continuation that runs after an action
// that can itself pause (a discard, a sacrifice).
//
// The copy is made with cloneReplacementResume, the clone every undo
// snapshot gives a resume frame, and it is cloned AGAIN each time the
// continuation runs (thaw). So the continuation never shares a mutable
// event with the game it is running in: an undo into Library of Leng's
// prompt re-answers against a fresh copy, which is the undo contract
// every continuation in the engine keeps (ADR 0098 Decision 3). It adds
// no field anywhere — the closure rides the discard's or sacrifice's
// existing route continuation.
func frozenEntryChoice(ev *ReplacementEvent, chosen activeReplacement) *replacementResumeFrame {
	return cloneReplacementResume(&replacementResumeFrame{
		ev:         ev,
		applicable: []activeReplacement{chosen},
	})
}

// thawEntryChoiceLocked returns a fresh copy of a frozen entry and its
// effect. The effect's source is re-read live: the frozen pointer
// points into a zone slice that the discard or the sacrifice may have
// shifted, and a stale *Card is the one thing a continuation must
// never dereference.
//
// Caller must hold g.mu.
func (g *Game) thawEntryChoiceLocked(frozen *replacementResumeFrame) (*ReplacementEvent, activeReplacement) {
	f := cloneReplacementResume(frozen)
	chosen := f.applicable[0]
	if chosen.source != nil {
		if c, ok := g.LookupCardForEffect(chosen.source.InstanceID); ok {
			live := c
			chosen.source = &live
		}
	}
	return f.ev, chosen
}

// resumeEntryAfterActionLocked is the continuation of a discard or a
// sacrifice an entry choice started. `done` is "if you do": every named
// card really left its zone. When it did not, Replace runs — the "you
// didn't" branch. Then the entry resumes from a fresh copy of the frozen
// event.
//
// Caller must hold g.mu.
func (g *Game) resumeEntryAfterActionLocked(frozen *replacementResumeFrame, done bool, moved []uuid.UUID) error {
	ev, chosen := g.thawEntryChoiceLocked(frozen)
	spec := chosen.effect.EntryCardChoice
	if spec != nil && spec.AnyNumber {
		// "Any number of": nothing to fail at, and no "if you don't"
		// branch. What really left is the count.
		done = true
	}
	if !done {
		g.runEntryChoiceDeclineLocked(ev, chosen)
	} else if spec != nil {
		g.runEntryThenLocked(spec, ev, moved)
	}
	return g.continueEntryAfterChoiceLocked(ev)
}

// continueEntryAfterChoiceLocked re-enters the apply-loop for an entry
// whose own question has been answered, and settles it through the
// shared finisher — the one the other entry resumes use (#478), so a
// fetched reveal-land's search still gets its shuffle and its caller's
// Then, and a redirected Mox Diamond is moved by
// moveRedirectedEntryLocked.
//
// Caller must hold g.mu.
func (g *Game) continueEntryAfterChoiceLocked(ev *ReplacementEvent) error {
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
	return g.finishSettledReplacementLocked(ev, out)
}

// applyEntersDevouredLocked is the battlefield landing's half of
// ReplacementEvent.EntersDevoured (CR 702.82a): the permanent that has
// just arrived records how many creatures it devoured, for CR 702.82b.
//
// Caller must hold g.mu in write mode.
func (g *Game) applyEntersDevouredLocked(cardID uuid.UUID, n int) {
	if n <= 0 {
		return
	}
	if c := findBattlefieldCard(g, cardID); c != nil {
		c.Devoured = n
	}
}

// DevouredBy returns how many creatures the permanent devoured as it
// entered (CR 702.82b): zero for a permanent that is not on the
// battlefield or devoured nothing.
//
// Caller must hold either lock.
func (g *Game) DevouredBy(sourceID uuid.UUID) int {
	if c := findBattlefieldCard(g, sourceID); c != nil {
		return c.Devoured
	}
	return 0
}

// runEntryThenLocked runs a choice's Then and applies its devour with
// the cards that really moved. Errors are logged, not returned, like every
// other inline hook of the apply-loop. Devour is skipped when
// nothing moved.
//
// Caller must hold g.mu.
func (g *Game) runEntryThenLocked(spec *EntryCardChoice, ev *ReplacementEvent, moved []uuid.UUID) {
	if spec.Then != nil {
		if err := spec.Then(g, ev, append([]uuid.UUID(nil), moved...)); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
	}
	if spec.Devour > 0 && len(moved) > 0 {
		ev.EntersDevoured += len(moved)
		ev.AddCounterAtETB(CounterPlusOne, spec.Devour*len(moved))
	}
}
