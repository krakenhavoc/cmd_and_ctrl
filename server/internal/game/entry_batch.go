package game

import (
	"errors"

	"github.com/google/uuid"
)

// entry_batch.go — one SIMULTANEOUS battlefield entry of several cards,
// which can stop to ask a question for any of them (#1322) and can take
// its cards from more than one zone (#1324). ADR 0061 amendment
// 2026-09-23.
//
// # Why the batch could not pause
//
// "Put … onto the battlefield" for several cards at once — Genesis
// Wave, Warp World, Sword of Hearth and Home's "put both cards" — has to
// run every card's CR 614 window against the board as it stood before
// ANY of them entered (CR 614.12), land them all, and only then
// announce them (CR 603.6a: "all permanents on the battlefield
// (including the newcomers) are checked"). The single-card resume
// (executeEntryToBattlefieldLocked) finishes ONE card's entry, so a
// card of the batch that paused would land after its siblings had
// already been announced. So the batch was built without
// entryResumable, and every question an entry could ask took its
// un-asked branch: a shockland or a pay-life MDFC back entered tapped, a
// reveal-land entered tapped, a Clone copied nothing.
//
// # What makes it resumable
//
// CR 614.12a says the choice is made before the permanent enters, and
// nothing says the choices for a simultaneous entry are made at the same
// instant — only that all of them are made before any of the permanents
// enter. So the batch is a value that walks its cards' windows ONE AT A
// TIME and stops when one asks. The paused card's event carries the
// batch on its entryTail, inside the replacementResume frame every
// paused event already uses. When the answer settles that event, the
// resume does not land the card: it hands the settled event back to the
// batch, which records it and walks on to the next card. Only when every
// window has settled does anything move — the same three phases the
// batch always had, now on the far side of however many questions.
//
// CR 614.12b falls out of the order: a player asked to pay 2 life for
// the first of two shocklands pays it at once, so the second question
// is asked against the life total the first payment left, and "combined
// costs that are not payable" cannot be chosen.
//
// # Undo
//
// The batch is mutable across the pause — its cursor, the settled events
// behind it — and the resume frame that holds it is cloned into every
// undo snapshot. cloneReplacementResume gives the snapshot its own copy
// of the batch (cloneEntryBatch), for the reason it gives the entry
// tail one: sharing it would let the live game's resume walk the
// snapshot's cursor on, so undoing an answer and giving it again would
// skip the rest of the batch.

// BatchEntry names one card of a simultaneous entry and the zone it is
// put from. From is ZoneHand, ZoneLibrary, ZoneExile or ZoneCommand; a
// card put from exile returns as a NEW OBJECT (CR 400.7) exactly as
// ReturnFromExileToBattlefieldForEffect's does, so the entered ID the
// continuation is told about is the new one.
//
// A card put from the COMMAND ZONE (#1278, commander ninjutsu,
// CR 702.49c) keeps its ID, as a hand or library card does, and for a
// reason the other two do not have: Player.CommanderCasts is keyed by
// the commander's instance ID, so minting a new one on the way out
// would silently reset its CR 903.8 tax the next time it is cast.
type BatchEntry struct {
	CardID uuid.UUID
	From   ZoneKind
}

// entryBatch is a simultaneous entry in progress. Unexported engine
// plumbing, carried on entryTail across a paused window.
type entryBatch struct {
	members []entryBatchMember
	// next is the first member whose CR 614 window has not run.
	next int
	// firstErr is the first error a member's window or move reported.
	// The batch carries on past it (one stale card must not strand the
	// rest) and reports it once everything is done.
	firstErr error
	// then is the caller's continuation, run exactly once when the
	// batch reaches its end. Cleared as it runs.
	then func(g *Game, entered []uuid.UUID) error
	opts ZoneEntryOptions
}

// entryBatchMember is one card of an entryBatch.
type entryBatchMember struct {
	cardID uuid.UUID
	from   ZoneKind
	// controller is who it enters under, stamped on the card in its
	// source zone before the first window opens.
	controller uuid.UUID
	// priorController is what the card carried in its source zone,
	// put back if it does not enter.
	priorController uuid.UUID
	// out is the settled event, nil when the member does not enter
	// (cancelled, redirected, errored, or its prompt taken away).
	out *ReplacementEvent
}

// cloneEntryBatch deep-copies a batch for an undo snapshot. The settled
// events are copied too, with their own EntersWithCounters maps, so
// nothing the live game's landing does to an event can reach the
// snapshot's copy of it. Their tails are copied with the batch pointer
// cleared, as recordEntryBatchMemberLocked leaves them.
func cloneEntryBatch(b *entryBatch) *entryBatch {
	if b == nil {
		return nil
	}
	out := *b
	out.members = make([]entryBatchMember, len(b.members))
	for i, m := range b.members {
		out.members[i] = m
		if m.out != nil {
			ev := *m.out
			if len(m.out.EntersWithCounters) > 0 {
				ev.EntersWithCounters = make(map[string]int, len(m.out.EntersWithCounters))
				for k, v := range m.out.EntersWithCounters {
					ev.EntersWithCounters[k] = v
				}
			}
			if m.out.entryTail != nil {
				t := *m.out.entryTail
				t.batch = nil
				ev.entryTail = &t
			}
			out.members[i].out = &ev
		}
	}
	return &out
}

// PutOntoBattlefieldTogetherThenForEffect puts every named card onto the
// battlefield as ONE simultaneous entry, from whichever zones they are
// in, and then runs `then` with the permanents that entered — Sword of
// Hearth and Home's "exile up to one target creature you own, then
// search your library for a basic land card. Put both cards onto the
// battlefield under your control" (#1324).
//
// It is the batch the hand and library "put" doors already were, with
// the source zone per card rather than per call. Two entries one after
// the other are observably different from this one: CR 603.6a checks
// every permanent on the battlefield, newcomers included, against the
// event that put them there, so with the creature and the land entering
// together a landfall creature sees the land and an intervening-if
// ("if an opponent controls more lands than you") reads the board with
// both of them on it.
//
// Every card's CR 614 window can ask its question (a shockland's life,
// a Clone's copy, a reveal-land's reveal), one card at a time, before
// anything moves; see the file comment. `then` runs exactly once, on
// every outcome but a pause — immediately when nothing asked, and from
// the answer to the last question when something did. It is told
// uuid.Nil-free IDs in the order given, with an exiled card's NEW ID.
// Like every continuation, it takes the live *Game.
//
// Every entry is validated before anything happens, as the single-zone
// batch does: a card not in the zone it is named from
// (ErrCardNotFound), a zone this door does not put from, a nonpermanent
// card or a token (ErrInvalidParam) refuses the whole batch, and `then`
// is still told nothing entered. A repeated ID is taken once.
//
// Caller must hold g.mu.
func (g *Game) PutOntoBattlefieldTogetherThenForEffect(cards []BatchEntry, opts ZoneEntryOptions, then func(g *Game, entered []uuid.UUID) error) error {
	return g.startEntryBatchLocked(cards, opts, then)
}

// startEntryBatchLocked validates a batch, stamps its controllers and
// walks its windows.
//
// Caller must hold g.mu.
func (g *Game) startEntryBatchLocked(cards []BatchEntry, opts ZoneEntryOptions, then func(g *Game, entered []uuid.UUID) error) error {
	b := &entryBatch{then: then, opts: opts}
	refuse := func(err error) error {
		if tailErr := g.finishEntryBatchThenLocked(b, nil); tailErr != nil {
			return errors.Join(err, tailErr)
		}
		return err
	}
	seen := make(map[uuid.UUID]bool, len(cards))
	for _, e := range cards {
		if seen[e.CardID] {
			continue
		}
		seen[e.CardID] = true
		switch e.From {
		case ZoneHand, ZoneLibrary, ZoneExile, ZoneCommand:
		default:
			return refuse(ErrInvalidParam)
		}
		src := g.findCardZoneLocked(e.CardID)
		if src == nil || src.Kind != e.From {
			return refuse(ErrCardNotFound)
		}
		c, ok := g.cardInZoneLocked(src, e.CardID)
		if !ok {
			return refuse(ErrCardNotFound)
		}
		// CR 110.4: only a permanent card can be put onto the
		// battlefield — unless it is going there FACE DOWN, in which
		// case the object that arrives is a 2/2 creature whatever the
		// card says (CR 701.40a, CR 708.2). Manifesting an instant is
		// legal and common.
		if !c.IsPermanent() && opts.FaceDown == FaceDownNone {
			return refuse(ErrInvalidParam)
		}
		if c.IsToken() {
			// CR 111.8: a token that has left the battlefield can't
			// move to another zone or come back onto the battlefield,
			// and CR 108.2: it is not a "card" at all. Since #596 the
			// CR 704.5d sweep removes it at the next state-based
			// check, so a token tucked into a library (Chaos Warp) is
			// only there for the window before that; putting it back
			// inside that window would undo the removal that tucked
			// it. Refused like a nonpermanent.
			return refuse(ErrInvalidParam)
		}
		controller := opts.Controller
		if controller == uuid.Nil || g.playerByIDLocked(controller) == nil {
			controller = src.Owner
		}
		b.members = append(b.members, entryBatchMember{
			cardID:          e.CardID,
			from:            e.From,
			controller:      controller,
			priorController: c.Controller,
		})
	}
	if len(b.members) == 0 {
		return g.finishEntryBatchThenLocked(b, nil)
	}
	// Stamp every controller before the first window opens, so a
	// replacement consulted for one card of the batch reads the right
	// controller on the others (Authority of the Consuls asks whose
	// permanent is entering).
	for _, m := range b.members {
		g.setControllerInZoneLocked(m.cardID, m.from, m.controller)
	}
	return g.continueEntryBatchLocked(b)
}

// setControllerInZoneLocked stamps Controller on a card still in a zone
// of kind `from`, located live. A card that has left that zone is left
// alone.
//
// Caller must hold g.mu.
func (g *Game) setControllerInZoneLocked(cardID uuid.UUID, from ZoneKind, controller uuid.UUID) {
	z := g.findCardZoneLocked(cardID)
	if z == nil || z.Kind != from {
		return
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == cardID {
			z.Cards[i].Controller = controller
			return
		}
	}
}

// continueEntryBatchLocked opens the CR 614 window of every member that
// has not had one yet, in order, against the pre-entry board. A window
// that asks a question stops the walk: the question's resume hands the
// settled event back through resumeEntryBatchLocked, which calls this
// again. When the last window has settled, the batch lands.
//
// Caller must hold g.mu.
func (g *Game) continueEntryBatchLocked(b *entryBatch) error {
	for b.next < len(b.members) {
		i := b.next
		b.next++
		m := b.members[i]
		ev := &ReplacementEvent{
			Kind:         RepEventMove,
			Actor:        m.controller,
			CardID:       m.cardID,
			OldZone:      m.from,
			NewZone:      ZoneBattlefield,
			EntersTapped: b.opts.Tapped,
			// #1227: and the attacking clause, seeded here for the
			// reason EntersTapped is. See ZoneEntryOptions.Attacking.
			EntersAttacking: b.opts.Attacking,
			// ADR 0082 decision 3: the face-down state rides the
			// EVENT rather than a local, so both battlefield-entry
			// doors carry the same fact in the same field and a
			// replacement effect inspecting the entry sees it.
			FaceDown:       b.opts.FaceDown,
			FaceDownListed: b.opts.FaceDownListed,
			// #1322: the batch can be resumed, because what it still
			// owes — the other cards' windows, the landing, the
			// announcement and the caller's continuation — rides on
			// the tail. So a shockland put by Genesis Wave is asked.
			entryResumable: true,
			entryTail: &entryTail{
				// CR 400.7: a card put from exile returns as a new
				// object, as ReturnFromExileToBattlefieldForEffect's
				// does. Hand and library entries keep their IDs.
				newObject:  m.from == ZoneExile,
				batch:      b,
				batchIndex: i,
			},
		}
		out, err := g.applyReplacementsLocked(ev)
		if errors.Is(err, errReplacementPending) {
			// A prompt is queued and its resume owns the batch now.
			// The event's CR 614.5 tracking survives for the resume,
			// which re-enters the apply-loop with the same ev.ID.
			return nil
		}
		g.clearReplacementEventLocked(ev.ID)
		if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
			if b.firstErr == nil {
				b.firstErr = err
			}
			out = nil
		}
		g.recordEntryBatchMemberLocked(b, i, out)
	}
	return g.landEntryBatchLocked(b)
}

// recordEntryBatchMemberLocked stores a member's settled event. A
// cancelled entry, and one a replacement redirected somewhere other
// than the battlefield, is recorded as not entering: there is no
// generic "put it wherever the pipeline said" helper for these sources,
// so a redirect is treated as a cancel rather than guessed at — the
// posture every effect-side entry takes.
func (g *Game) recordEntryBatchMemberLocked(b *entryBatch, i int, out *ReplacementEvent) {
	if out == nil || out.Canceled || out.NewZone != ZoneBattlefield {
		b.members[i].out = nil
		return
	}
	if out.entryTail != nil {
		// The batch is finished with this event's tail as a route
		// back to it; nothing may re-enter the batch through it.
		out.entryTail.batch = nil
	}
	b.members[i].out = out
}

// resumeEntryBatchLocked is how a paused member's settled event gets
// back to its batch — from applyResolvedReplacementEventLocked with the
// settled event when the window settled on an entry, and from
// runEntryTailLocked with nil when it was cancelled or its prompt was
// taken away (the card left its zone, the chooser left the game).
//
// Caller must hold g.mu.
func (g *Game) resumeEntryBatchLocked(b *entryBatch, i int, out *ReplacementEvent) error {
	if b == nil || i < 0 || i >= len(b.members) {
		return nil
	}
	g.recordEntryBatchMemberLocked(b, i, out)
	return g.continueEntryBatchLocked(b)
}

// takeEntryBatch detaches the batch an event's tail is carrying, so
// exactly one path can hand the event back to it.
func (ev *ReplacementEvent) takeEntryBatch() (*entryBatch, int, bool) {
	if ev == nil || ev.entryTail == nil || ev.entryTail.batch == nil {
		return nil, 0, false
	}
	b, i := ev.entryTail.batch, ev.entryTail.batchIndex
	ev.entryTail.batch = nil
	return b, i, true
}

// landEntryBatchLocked is the batch's second and third phases: land
// every member whose window settled on the battlefield, then announce
// them all, then run their AsEnters hooks — so every permanent of the
// batch is on the battlefield when the first event is harvested
// (CR 603.6a) — and finally tell the caller.
//
// Each member is located LIVE: the batch may have waited on a question,
// and a card that left its zone meanwhile (milled, drawn, exiled) is
// not part of the entry. landEntryLocked refuses a card no longer in
// the zone its window opened over.
//
// Caller must hold g.mu.
func (g *Game) landEntryBatchLocked(b *entryBatch) error {
	landings := make([]entryLanding, 0, len(b.members))
	landed := make([]bool, len(b.members))
	for i, m := range b.members {
		if m.out == nil {
			continue
		}
		l, ok, err := g.landEntryLocked(m.out)
		if err != nil {
			if b.firstErr == nil {
				b.firstErr = err
			}
			continue
		}
		if !ok {
			continue
		}
		landed[i] = true
		landings = append(landings, l)
	}
	// A card that did not enter is still in its source zone wearing
	// the controller the batch stamped for the entry. Give it back what
	// it had, so a put under someone other than the owner (Lonis)
	// leaves no trace on a card it never moved.
	for i, m := range b.members {
		if !landed[i] {
			g.setControllerInZoneLocked(m.cardID, m.from, m.priorController)
		}
	}
	for _, l := range landings {
		g.announceEntryLocked(l)
	}
	for _, l := range landings {
		g.runEntryHooksLocked(l)
	}
	entered := make([]uuid.UUID, 0, len(landings))
	for _, l := range landings {
		entered = append(entered, l.entered)
	}
	if len(landings) > 0 {
		// #1069: the batch has left the zones it came from, so an open
		// choose_cards prompt over one of them — a discard prompt whose
		// candidates a Warp World just put onto the battlefield — is
		// trimmed or withdrawn here. One call for the whole batch: the
		// prune is keyed by zone and re-reads every open pick
		// (battlefield_entry.go). After the hooks, so every arrival is
		// announced and every ETB hook has run before a withdrawal's
		// continuation can start something of its own.
		g.pruneChoicesAfterArrivalLocked()
	}
	return g.finishEntryBatchThenLocked(b, entered)
}

// finishEntryBatchThenLocked runs the caller's continuation exactly once
// and reports the batch's first error ahead of the continuation's.
//
// Caller must hold g.mu.
func (g *Game) finishEntryBatchThenLocked(b *entryBatch, entered []uuid.UUID) error {
	then := b.then
	b.then = nil
	var thenErr error
	if then != nil {
		thenErr = then(g, entered)
	}
	if b.firstErr != nil {
		return b.firstErr
	}
	return thenErr
}
