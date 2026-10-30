package game

import (
	"errors"
	"fmt"

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
// put from. From is ZoneHand, ZoneLibrary, ZoneGraveyard, ZoneExile or
// ZoneCommand; a card put from exile returns as a NEW OBJECT (CR 400.7)
// exactly as ReturnFromExileToBattlefieldForEffect's does, so the
// entered ID the continuation is told about is the new one.
//
// A card put from a GRAVEYARD (#1867: Reveillark, Pull, Replenish)
// keeps its ID, as the single-card reanimation door
// (returnFromGraveyardFaceLocked) does: it is a new object all the
// same, and the battlefield landing re-mints its object epoch as it
// does for every card that arrives there. Before #1867 every
// multi-card return from a graveyard was a loop over that single-card
// door, so the first card was announced before the second moved and
// only the first saw the second enter.
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
	// exchangeSource / exchangeSacrifice name the exit paired with this
	// one-card entry (Goblin Welder). Both windows settle before either moves.
	exchangeSource, exchangeSacrifice uuid.UUID
	source                            uuid.UUID
	members                           []entryBatchMember
	// next is the first member whose CR 614 window has not run.
	next int
	// firstErr is the first error a member's window or move reported.
	// The batch carries on past it (one stale card must not strand the
	// rest) and reports it once everything is done. Held as data
	// (#4361 S10): its message and the engine sentinel it wraps.
	firstErr entryBatchErr
	// keyed is the caller's continuation as data (#4253 slice 3), run
	// exactly once when the batch reaches its end: a registered tail,
	// told the permanents that entered as TailOutcome.IDs. Cleared as it
	// runs. (#4361 S17 deleted the closure form, `then`, with its last
	// writer, a closure discard batch's leg that became this entry.)
	keyed KeyedTail
	// discard is the rest of a keyed discard batch whose leg became this
	// entry (#4253 slice 9, discard_entry.go), run after the entry
	// instead of a caller's continuation.
	discard *discardEntryRun
	// attraction is the rest of an "open N Attractions" whose open
	// this entry is (#4361 S10, attractions.go), run after the entry
	// instead of a caller's continuation.
	attraction *attractionOpenRun
	// report is a synchronous token door's result (tokenTail.report),
	// told the tokens that entered together. Tokens entering together
	// (enterCreatedTokensTogetherLocked) leave the staging area as the
	// batch finishes: a member that is a token and did not enter never
	// existed (CR 111.1).
	report *[]uuid.UUID
	opts   ZoneEntryOptions
	// done and landed are set as the batch finishes: what entered. A
	// synchronous door reads them on its next line instead of handing
	// the batch a closure over a local (putOntoBattlefieldFromZoneLocked),
	// so a batch paused on a question holds no closure. A batch that
	// paused returns before it is done, and the door reports nothing, as
	// it always did.
	done   bool
	landed []uuid.UUID
}

// entryBatchMember is one card of an entryBatch.
type entryBatchMember struct {
	cardID uuid.UUID
	from   ZoneKind
	// Tokens carry their own entry clause and come from no zone.
	token      bool
	tokenEntry TokenEntryOptions
	attacking  uuid.UUID
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

// startEntryBatchLocked validates a batch, stamps its controllers and
// walks its windows.
//
// Caller must hold g.mu.
func (g *Game) startEntryBatchLocked(cards []BatchEntry, opts ZoneEntryOptions) error {
	return g.startPreparedEntryBatchLocked(cards, &entryBatch{opts: opts})
}

func (g *Game) startPreparedEntryBatchLocked(cards []BatchEntry, b *entryBatch) error {
	// ADR 0210: an unsound overwrite is a programming error, caught by
	// the first test that puts a card with it.
	if p := b.opts.EntersAs.Problem(); p != "" {
		panic(fmt.Sprintf("game: %s (%q)", p, b.opts.EntersAs.Label))
	}
	// #1997: a batch begun inside PutByEnteringPlayerForEffect, or a
	// manifest or cloak (the player who manifests puts the card, and it
	// enters under them, CR 701.40a, CR 701.58a), is put by the player
	// each card enters under. Captured now: a later member may land
	// after a pause, when the scope has ended.
	if g.putByEntering > 0 || b.opts.FaceDown == FaceDownManifested || b.opts.FaceDown == FaceDownCloaked {
		b.opts.PutByEnteringPlayer = true
	}
	opts := b.opts
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
		case ZoneHand, ZoneLibrary, ZoneGraveyard, ZoneExile, ZoneCommand:
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
		// "Under its owner's control" is the CARD's owner. The zone's
		// owner is the same player for a hand, library, graveyard or
		// command zone, but exile is shared and owned by nobody, so a
		// card returned from it would carry no controller while its
		// CR 614 window is open (#1872) — and Authority of the Consuls
		// asks whose permanent is entering.
		controller := opts.Controller
		if controller == uuid.Nil || g.playerByIDLocked(controller) == nil {
			controller = c.Owner
			if controller == uuid.Nil {
				controller = src.Owner
			}
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
			Source:       b.source,
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
			// ADR 0210: "they are <types>", seeded for the reason
			// EntersTapped is. Shared read-only, never written through.
			entersAs: b.opts.EntersAs,
			// #1997: who put it, for Arboria.
			putByEntering: b.opts.PutByEnteringPlayer,
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
		// ADR 0153: the putting effect's own "with N counters" clause,
		// copied onto this card's event (AddCounterAtETB sums into the
		// event's own map, so the options map is never written).
		for name, n := range b.opts.EntersWithCounters {
			ev.AddCounterAtETB(name, n)
		}
		// #3158: the putting effect's own "attached to …" clause.
		if b.opts.AttachedTo.Kind != "" {
			ev.auraHost, ev.auraHostFixed = b.opts.AttachedTo, true
		}
		if m.token {
			ev.EntersTapped = m.tokenEntry.Tapped
			ev.EntersAttacking = m.attacking
			for name, n := range m.tokenEntry.Counters {
				ev.AddCounterAtETB(name, n)
			}
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
			b.firstErr.record(err)
			out = nil
		}
		g.recordEntryBatchMemberLocked(b, i, out)
	}
	return g.landEntryBatchLocked(b)
}

// recordEntryBatchMemberLocked stores a member's settled event. A
// cancelled entry is recorded as not entering. One a replacement
// REDIRECTED somewhere other than the battlefield (a Mox Diamond whose
// controller discarded nothing) is kept: the batch moves it there when
// it lands (landEntryBatchLocked, ADR 0098 Decision 4) rather than
// leaving it where it was.
func (g *Game) recordEntryBatchMemberLocked(b *entryBatch, i int, out *ReplacementEvent) {
	if out == nil || out.Canceled {
		b.members[i].out = nil
		return
	}
	if out.entryTail != nil {
		// The batch is finished with this event's tail as a route
		// back to it; nothing may re-enter the batch through it.
		out.entryTail.batch = nil
	}
	// #4193: the putting effect's per-card "if it's a <type>" counters,
	// judged now that this member's window has settled (so after any
	// copy it was told to enter as) and before anything lands.
	g.seedEntryCountersIfTypeLocked(b.opts.EntersWithCountersIfType, out)
	g.seedEntryTappedIfTypeLocked(b.opts.EntersTappedIfType, out)
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
	if b.exchangeSacrifice != uuid.Nil {
		if len(g.sacrificeableLocked(g.unprotectedFromSacrificeLocked([]uuid.UUID{b.exchangeSacrifice}))) == 0 {
			b.exchangeSacrifice = uuid.Nil
			return g.landEntryBatchLocked(b)
		}
		r := sacrificeRoute(b.exchangeSource)
		// A sacrifice is a routed zone move, not destruction. This path
		// needs the exit's landing/announcement split for the paired entry.
		r.ViaBattlefieldLeave = false
		r.Dst = ZoneGraveyard
		if err := g.announceSacrificeLocked(b.exchangeSacrifice, b.exchangeSource, MoveCause{}); err != nil {
			return err
		}
		if len(b.members) == 1 && b.members[0].out != nil {
			r.simultaneousEntry = settledExchangeEntryOf(b.members[0].out)
		}
		return g.routeLegLocked(r, b.exchangeSacrifice, nil)
	}
	landings := make([]entryLanding, 0, len(b.members))
	landed := make([]bool, len(b.members))
	// ADR 0098 Decision 4: a member whose window was REDIRECTED never
	// enters (CR 614.6). It goes where its window sent it, as part of
	// the same event and before any member that does enter is
	// announced. Its controller stamp is put back first, so the card
	// arrives in its new zone carrying nothing the entry gave it.
	for _, m := range b.members {
		if m.out == nil || m.out.NewZone == ZoneBattlefield {
			continue
		}
		g.setControllerInZoneLocked(m.cardID, m.from, m.priorController)
		if err := g.moveRedirectedEntryLocked(m.out); err != nil {
			b.firstErr.record(err)
		}
	}
	for i, m := range b.members {
		if m.out == nil || m.out.NewZone != ZoneBattlefield {
			continue
		}
		l, ok, err := g.landEntryLocked(m.out)
		if err != nil {
			b.firstErr.record(err)
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
	// ADR 0109 §8: one simultaneous entry is ONE entry for the world
	// rule's "shortest amount of time" (CR 704.5k) — every card it
	// announces shares an ordinal.
	// #3862: an "as this enters" clause that changes other objects is
	// part of the entry, before any of it is announced.
	g.runAsEntersOthersLocked(landings)
	release := g.beginSimultaneousEntryLocked()
	for _, l := range landings {
		g.announceEntryLocked(l)
	}
	release()
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
	keyed, discard, attraction, report := b.keyed, b.discard, b.attraction, b.report
	b.keyed, b.discard, b.attraction, b.report = KeyedTail{}, nil, nil, nil
	b.done, b.landed = true, entered
	// Tokens entering together: every one that did not land never
	// existed (CR 111.1), and every one that did is on the battlefield
	// now, so none is staged any longer.
	for _, m := range b.members {
		if m.token {
			g.dropEnteringTokenLocked(m.cardID)
		}
	}
	if report != nil {
		*report = entered
	}
	var thenErr error
	switch {
	case discard != nil:
		thenErr = g.finishDiscardEntryLocked(discard)
	case attraction != nil:
		thenErr = g.finishAttractionOpenLocked(attraction, entered)
	default:
		thenErr = g.runKeyedTailLocked(keyed, entered, 0)
	}
	if err := b.firstErr.err(); err != nil {
		return err
	}
	return thenErr
}

// entryBatchErr is an error a batch member reported, as data (#4361 S10):
// its message, and the engine sentinel it wraps (an index into
// entryBatchErrSentinels, 0 for none), so the batch reaches no interface.
// The zero value is "no error".
type entryBatchErr struct {
	msg      string
	sentinel int
	set      bool
}

// entryBatchErrSentinels are the sentinels a batch's error keeps its
// identity for: what a member's window, landing or redirected move
// reports. Index 0 is "none". Never written to disk, so the order is
// free to change.
var entryBatchErrSentinels = []error{
	nil,
	ErrCardNotFound,
	ErrZoneNotFound,
	ErrZoneEmpty,
	ErrPlayerNotFound,
	ErrInvalidParam,
	ErrIllegalTarget,
	ErrPlayerEliminated,
	ErrGameNotActive,
	ErrNotACreature,
	ErrReplacementIterationExceeded,
}

// record keeps err when it is the batch's first.
func (e *entryBatchErr) record(err error) {
	if err == nil || e.set {
		return
	}
	*e = entryBatchErr{msg: err.Error(), set: true}
	for i, s := range entryBatchErrSentinels {
		if s != nil && errors.Is(err, s) {
			e.sentinel = i
			break
		}
	}
}

// err is the recorded error, nil when none was: it says what the
// original said and errors.Is still finds its sentinel.
func (e entryBatchErr) err() error {
	if !e.set {
		return nil
	}
	return recordedEntryErr{msg: e.msg, sentinel: entryBatchErrSentinels[e.sentinel]}
}

// recordedEntryErr is entryBatchErr as an error.
type recordedEntryErr struct {
	msg      string
	sentinel error
}

func (e recordedEntryErr) Error() string { return e.msg }
func (e recordedEntryErr) Unwrap() error { return e.sentinel }
