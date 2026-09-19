package game

import (
	"errors"

	"github.com/google/uuid"
)

// zone_route.go is the engine's ONE zone-change path for an
// effect-driven move into a zone other than the battlefield.
//
// Why it exists (#529). CR 903.9 — "if a commander would be put into
// a library, hand, graveyard or exile from anywhere, its owner may
// put it into the command zone instead" — is implemented as a CR 614
// replacement effect (commanderZoneReplacement, builtin_replacements.go).
// A replacement effect only ever fires for a mover that ASKS: it needs
// a RepEventMove pushed through applyReplacementsLocked. Before this
// file exactly two callers did that, both battlefield → graveyard, so
// a commander that was countered, exiled, bounced, tucked or milled
// took a raw MoveCard and its owner was never offered the choice.
// "From anywhere" was honoured from one place.
//
// Patching the six callers would have fixed the six callers. The
// defect is structural: the replacement window was opened by
// individual movers instead of by the primitive they share, so a
// seventh mover reintroduces it for free. So the window moves down
// here, and the movers above are reduced to descriptions of a
// destination.
//
// The battlefield ENTRY side keeps its own path
// (executeEntryToBattlefieldLocked): an entry owes enters-tapped,
// enters-with-counters, enters-as-a-copy and the ETB fire, none of
// which an exit has any use for. This file is the exit half.
//
// #707 folded the sandbox move in. MoveCardByIDAsCommander used to
// keep its own pipeline call, on the grounds that it accepts an
// arbitrary source AND destination — the battlefield included — which
// an exit primitive has no business expressing. True of half of it:
// the half whose destination is the battlefield or the stack is an
// ENTRY and still runs inline over there. The other half IS an exit,
// and keeping its own pipeline call meant keeping its own resume,
// which it never had — a commander moved out of a graveyard, a hand,
// a library or the stack by hand paused on the CR 903.9 prompt and
// the move was simply lost, both answers alike. The exit primitive's
// resume finishes a move from whatever zone the card was in when the
// window opened, so folding the admin exit in gave it one for free.
//
// What this deliberately does not do:
//
//   - routeBattlefieldCardToOwnerGraveyardLocked keeps its own. The
//     destroy / sacrifice / SBA route zeroes battlefield-only state
//     before the move and owes a different fallback when the owner
//     has left the table; it already had a faithful resume of its
//     own before this file existed, and folding the two together
//     would be a second change riding on this one.
//
//   - A batch exit (Farewell, Evacuation — simultaneous.go) that
//     pauses on one card finishes the rest and lets the paused one
//     land when its owner answers. So a commander caught in a wrath
//     leaves the battlefield a moment after the rest of the board
//     rather than simultaneously with it, and its leaves-the-
//     battlefield event falls outside that batch. Holding a whole
//     simultaneous exit open across a player prompt needs a
//     continuation frame the engine does not have yet (#478); a
//     commander that is asked one beat late is a far smaller
//     deviation than a commander that is never asked at all.
//
// #853 folded the DISCARD in — the last exit that still moved a card
// with a raw MoveCard, and therefore the last one that never opened
// the window. It brought the other half of a pausable move with it:
// zoneRoute.then, the continuation a caller with more to do hands over
// instead of writing it on the next line. A multi-card discard
// sequences itself through it rather than looping, and discard.go says
// why.
//
// #866 made that sequencing ONE body (routeAllThenLocked,
// simultaneous.go) shared by every batched exit, and #893 put the mill
// on it too. The mill above is therefore both shapes at once, on one
// implementation: MillToZoneThenForEffect sequences through the
// continuation, because a caller that reads what was milled cannot be
// told while a commander's prompt is open, and the fire-and-forget
// MillToZoneForEffect keeps the loop that proceeds AROUND a paused
// card, because nothing is waiting on its answer and a mill must not
// re-read the top of the library.

// zoneRoute describes one card's motion into a zone, together with
// the bookkeeping that particular route owes beyond the physical
// move. It is the whole input to routeCardToZoneLocked, and it is
// stashed on the ReplacementEvent so a CR 903.9 pause can be
// finished faithfully from the resume path — a tucked commander whose
// owner declines still goes to the BOTTOM of the library if that is
// what the tucking effect said.
type zoneRoute struct {
	// CardID is the card to move. Its current zone is found by scan,
	// so a caller never has to name the source.
	CardID uuid.UUID

	// Dst is the destination zone kind. DstOwner names the seat for a
	// per-player destination; uuid.Nil means "the card's owner",
	// which is what CR 903.9's "its owner's" destinations all want.
	Dst      ZoneKind
	DstOwner uuid.UUID

	// Actor is stamped on the emitted events. uuid.Nil for routes
	// with no single responsible player.
	Actor uuid.UUID

	// ToBottom sends a ZoneLibrary destination to the bottom rather
	// than the top (Condemn, "put it on the bottom of its owner's
	// library").
	ToBottom bool

	// Depth places a ZoneLibrary destination N cards down from the
	// top — Teferi, Hero of Dominaria's "third from the top" is 3.
	// Zero and 1 both mean the top, which is the default every other
	// library route wants. ToBottom wins if both are set.
	//
	// It rides the route rather than being applied by the caller for
	// the same reason ToBottom does: a commander tucked to depth
	// whose owner is asked about the command zone and declines still
	// lands at the right depth, because nothing has moved until the
	// prompt is answered.
	Depth int

	// FaceDown exiles the card face down in this state (CR 406.3a) —
	// FaceDownExiled for Necropotence, FaceDownForetold for foretell
	// (#658). The zero value, FaceDownNone, is an ordinary face-up
	// move.
	//
	// Face-down exile is the one destination that must NOT mark the
	// table as knowers: who may look is the kind's answer, written by
	// applyFaceDownLandingLocked. See ADR 0069 decision 2 and
	// ExileTopFaceDownForEffect.
	FaceDown FaceDownKind

	// Mill flags a mill (CR 701.17) so the completed move emits
	// EventMill rather than EventZoneMove. Only honoured when the
	// move actually lands in a graveyard: a commander redirected to
	// the command zone was never put into a graveyard, so it was
	// never milled, and no mill payoff should see it.
	Mill bool

	// Discard flags a discard (CR 701.8a) so the completed move emits
	// EventDiscardCard rather than EventZoneMove — the single event
	// every "whenever you discard a card" payoff keys on, and the one
	// Syr Konrad's "put into a graveyard from anywhere other than the
	// battlefield" clause counts a hand card by.
	//
	// Unlike Mill it is honoured WHEREVER the card lands (#853). A
	// commander whose owner takes CR 903.9's offer was still moved out
	// of their hand by a discard, so the discard happened and its
	// payoffs see it; what changed is only where the card ended up.
	// A mill is the other way round because CR 701.17a defines the
	// keyword action by its destination ("puts the top N cards of
	// their library into their graveyard") while CR 701.8a defines a
	// discard by its SOURCE ("move it from its owner's hand").
	Discard bool

	// DiscardCause is why the discard is happening — an effect's
	// instruction, a cost, or the cleanup step's turn-based action. It
	// rides onto the RepEventDiscard this route opens, where Library of
	// Leng and the rest of the cause-sensitive family read it, and onto
	// the EventDiscardCard the completed move emits. Meaningless unless
	// Discard is set; the empty value is normalised to
	// DiscardCauseEffect. #650.
	DiscardCause DiscardCause

	// Source is the card whose effect asked for the move, stamped on
	// the emitted event. The Discard leg has always carried one; #931
	// gave the plain move and the mill one too, because surveil's
	// graveyard leg names the card that surveilled and its EventMill
	// carried that before the leg went through this route. uuid.Nil
	// everywhere else leaves the event exactly as it was.
	Source uuid.UUID

	// MustSettleNow forbids this move from pausing on a player prompt:
	// the pipeline applies what it gathered and skips anything that
	// would ask a question rather than queueing one (see
	// ReplacementEvent.mustSettleNow).
	//
	// One thing sets it today: a discard paid as a COST (CR 601.2h /
	// CR 602.2b). Costs are paid as one indivisible step, so a
	// CR 903.9 prompt in the middle would leave a spell on the stack
	// with its cost half paid — the argument payLifeAsCostLocked makes
	// for the other half of the same cost line. A cost discard of a
	// commander therefore goes to the graveyard without asking:
	// CR 903.9 is a "may", and a cost that cannot ask falls back to
	// the ordinary result.
	MustSettleNow bool

	// Countered flags a counterspell, whose completed move emits
	// EventCounterSpell INSTEAD of EventZoneMove (the historic shape
	// — counter watchers key on it and a zone move would double-
	// count). DropStackMeta additionally retires the stack item.
	Countered     bool
	DropStackMeta bool

	// Sacrifice says this battlefield exit is a SACRIFICE (CR 701.17a)
	// rather than a destruction, so the leg announces EventSacrifice
	// while the permanent is still on the battlefield, before the
	// window opens over its move. Only meaningful with
	// ViaBattlefieldLeave, which is the only route a sacrifice takes.
	//
	// It also picks the rule the leg's "this way" answer is read by:
	// routeLegLandedLocked sends a sacrifice to sacrificedThisWayLocked
	// rather than to destroyedThisWayLocked. #910.
	//
	// It is deliberately NOT a flavour of Destruction below, and the
	// template says so by building on battlefieldExitRoute: a sacrifice
	// is not a destruction (CR 701.17a), so the CR 701.19 regeneration
	// built-in must never see one.
	Sacrifice bool

	// ViaBattlefieldLeave says the physical move belongs to
	// executeBattlefieldLeaveLocked rather than to
	// executeZoneRouteLocked — the destroy / sacrifice / SBA exit,
	// which keeps its own mover for the reasons listed at the top of
	// this file and is the one exit in the engine that does.
	//
	// Such a route carries no destination of its own (the event's
	// NewZone is authoritative there, as it is everywhere else); it
	// rides along for `then`, so a caller that has to know what the
	// destruction actually DID gets the same continuation every other
	// exit already had. #815.
	ViaBattlefieldLeave bool

	// Destruction marks this exit as a DESTRUCTION (CR 701.7a) rather
	// than a sacrifice, a legend-rule death, an illegally attached
	// Aura or a zero-counter state-based action, all of which take the
	// same ViaBattlefieldLeave exit. It rides onto
	// ReplacementEvent.Destruction, where the CR 701.19 regeneration
	// built-in reads it. #667.
	//
	// CantBeRegenerated is the rider a destroying effect prints
	// alongside it (CR 701.19c). Meaningless without Destruction.
	Destruction       bool
	CantBeRegenerated bool

	// simultaneousExit carries the pre-move copies for a destroy batch
	// whose current leg may pause. The snapshot has to be active around
	// the physical move on BOTH paths: the inline path and the later
	// replacement-prompt resume. Keeping it only on the caller's stack
	// loses it while the prompt is open, so a watcher that died earlier
	// in the wipe cannot see the resumed leg leave (#815).
	//
	// Read-only after construction. Unexported engine plumbing.
	simultaneousExit []Card

	// AsCommander is the sandbox move_card action's "yes, send this
	// commander back to the command zone" flavour flag (#707). It is
	// NOT a gate on the CR 903.9 built-in — that gate was dropped in
	// #171 and the replacement has been destination-only ever since —
	// and it rides the route only so the breadcrumb on the event
	// (ReplacementEvent.asCommanderMove) survives a pause along with
	// everything else the move was asked for.
	AsCommander bool

	// then is the rest of whatever asked for the move, run once this
	// one has reached a TERMINAL outcome — landed, replaced away
	// (CR 614.10), or asked for a move that was not one. A pause is
	// not terminal: the resume reaches it later, through
	// executeZoneRouteLocked.
	//
	// #853: it exists because a move can PAUSE, and a caller with more
	// to do cannot just do it on the next line. A multi-card discard
	// sequences itself through here — each card's continuation starts
	// the next one and the last one runs the discard's own "then draw
	// two" — so the batch is a value carried forward rather than a
	// shared accumulator, which is what makes an undo across the
	// prompt land where a clean run would. The same idiom lifeTail and
	// damageTail use, and for the same reason.
	//
	// It takes the live *Game rather than capturing one, on the undo-
	// safety contract every other continuation in the engine follows:
	// an undo restores this game's fields in place, so a *Game
	// argument is always the right game and a captured *Player would
	// not be. It runs with g.mu held, so it may start the next move or
	// queue the next prompt itself.
	//
	// Unexported engine plumbing — the catalog never sets one. Cleared
	// THROUGH the pointer as it runs (runRouteTailLocked), which is
	// why cloneReplacementResume gives an undo snapshot its own copy
	// of the route.
	then func(g *Game) error
}

// runRouteTailLocked runs a settled route's continuation exactly once.
// The continuation is cleared before it runs, so a tail that re-enters
// the pipeline on the same route value cannot run itself twice.
//
// Every TERMINAL outcome of a routed move goes through here — landed,
// cancelled, or never a move at all — because a caller sequencing a
// batch through the tail has to be told even when the answer is
// "nothing happened", or it waits forever. A PAUSE is not terminal:
// the resume reaches this function later, through
// executeZoneRouteLocked.
//
// Caller must hold g.mu.
func (g *Game) runRouteTailLocked(r *zoneRoute) error {
	if r == nil || r.then == nil {
		return nil
	}
	then := r.then
	r.then = nil
	return then(g)
}

// abandonZoneRouteLocked is the terminal outcome a paused route reaches
// when its prompt is taken AWAY rather than answered. Two things do
// that, and before #865 both simply discarded the frame:
//
//   - the prompt is DROPPED because its chooser left the game
//     (cleanupStackForEliminatedLocked, CR 800.4a);
//   - the prompt is PRUNED because the card moved by some other route
//     while the question was open, so the move it is asking about can
//     never happen (pruneStaleZoneChangeChoicesLocked / #605, and its
//     answer-path twin dropStaleReplacementResumeLocked).
//
// Despite the name it is the abandon-side dispatcher for EVERY
// replacement event kind, not just a routed exit: every caller reaches
// it with whatever frame the discarded prompt was holding, and the
// switch inside decides what that kind owes. #982's enumeration test
// reads that switch, so a new kind cannot arrive here undecided.
//
// NOTHING MOVES and no event is emitted. A paused route has moved
// nothing (see routeCardToZoneLocked), so the card is still in its old
// zone and the route's own bookkeeping never happened — which is the
// same shape a CR 614.10 cancellation has, and the reason a discard
// abandoned here fires no EventDiscardCard: CR 701.8a defines a
// discard as the move OUT of the hand, and there was none.
//
// WHAT IT OWES IS THE CONTINUATION. #808 made "the player left" a
// terminal outcome of the life and damage tails for exactly this
// reason: a caller sequencing a batch through a continuation has to be
// told even when the answer is "nothing happened", or it waits
// forever. Without this a two-card discard stopped after the first
// card and a wipe stopped after the commander — the rest of the batch,
// and the caller's own "then draw two" / "for each creature destroyed
// this way", were dropped on the floor with the frame.
//
// The leg is reported as NOT LANDED, and it reports itself: every
// reader of a route's outcome reads the live board rather than being
// handed a verdict (destroyedThisWayLocked, landedInZoneLocked), and
// the board still has the card where it was. So an abandoned leg is
// not counted as destroyed, exiled, bounced or discarded, and needs no
// flag to say so.
//
// Caller must hold g.mu, and must already have taken the prompt out of
// g.PendingChoices.
func (g *Game) abandonZoneRouteLocked(frame *replacementResumeFrame) error {
	if frame == nil || frame.ev == nil {
		return nil
	}
	ev := frame.ev
	// The CR 614.5 once-per-event bookkeeping the abandoned event was
	// holding: nothing else will release it now.
	g.clearReplacementEventLocked(ev.ID)
	// One arm per ReplacementEventKind, for the reason
	// finishSettledReplacementLocked's switch carries: #982's
	// enumeration test (replacement_kind_gate_test.go) reads this
	// switch, so a new kind cannot reach the abandon path with nobody
	// deciding what it owes. Naming the kinds that owe NOTHING is half
	// the value — the keyword action was silently one of them until the
	// test said so.
	switch ev.Kind {
	case RepEventCreateTokens:
		// #762: an abandoned CREATION makes nothing, and the rest of
		// the card behind it still has to be told. Nothing is staged
		// yet — the tokens are not minted until the window settles.
		return g.abandonTokenCreationLocked(ev)
	case RepEventKeywordAction:
		// #982: and an abandoned KEYWORD ACTION, which this path was
		// missing. "Scry 2, then draw a card" whose CR 616 ordering
		// prompt is taken away has scried nothing and still owes the
		// draw — the same call finishSettledReplacementLocked makes for
		// a cancelled one. Before this the continuation went out with
		// the frame.
		return g.abandonKeywordActionLocked(ev)
	case RepEventMill:
		// #569: an abandoned MILL has read no cards off the library —
		// the plan is made after the amount settles — and the caller
		// reading "the cards milled this way" still has to be told.
		return g.abandonMillLocked(ev)
	case RepEventMove, RepEventDiscard:
		// #762: an abandoned ENTRY of a CREATED TOKEN leaves the token
		// staged and unentered. It never reached the battlefield, so it
		// never existed (CR 111.1). A no-op for every other entry.
		g.dropEnteringTokenLocked(ev.CardID)
		if err := g.runRouteTailLocked(ev.zoneRoute); err != nil {
			return err
		}
		// #478: a paused ENTRY is abandoned the same way and owes the
		// same answer. A fetch whose entry prompt is taken away — its
		// chooser left, or the card left the library by another route
		// while the question was open — moved nothing, and the search
		// behind it has to be told so, or its shuffle and its caller's
		// Then wait forever. A move carries a route or an entry tail,
		// never both, so this is one call and a no-op for every exit.
		// Since #762 a token creation threads the rest of its batch
		// through the same tail.
		return g.runEntryTailLocked(ev, uuid.Nil)
	case RepEventLife, RepEventDamage:
		// finishDroppedReplacementLocked settles these two itself, with
		// their own tails and an amount of zero, and never reaches
		// here; the prune and the stale-resume paths only ever hold a
		// zone change. Nothing to do either way — the tail has already
		// run, and running it again through the cleared pointer would
		// be a no-op rather than a second payout.
		return nil
	case RepEventDraw, RepEventCounter, RepEventStepTransition:
		// No continuation exists on any of the three, so an abandoned
		// one owes nobody an answer. A cancelled step transition is a
		// SKIP and does move the cursor (CR 500.11), but only when its
		// prompt is ANSWERED — one taken away leaves the step where it
		// was, which a player can always advance.
		return nil
	}
	return nil
}

// routeCardToZoneLocked opens the CR 614 replacement window for a
// move of r.CardID into r.Dst and, once the pipeline settles on a
// destination, performs the move through executeZoneRouteLocked.
//
// Returns paused=true when the pipeline queued a player prompt. In
// that case NOTHING has moved: the card is still sitting in its old
// zone, no event has been emitted, and the move completes from
// applyResolvedReplacementEventLocked when the player answers. Every
// caller has to be able to tolerate that, which is why the bool is
// returned rather than swallowed — ExileTopFaceDownForEffect in
// particular has to know not to read the same top card again.
//
// A caller with something to do AFTER the move hands it over as
// r.then rather than writing it on the next line, and then ignores the
// bool: the continuation runs from the landing either way, inline when
// nothing paused and from the resume when something did. That is how
// a multi-card discard sequences itself (#853).
//
// Caller must hold g.mu.
func (g *Game) routeCardToZoneLocked(r zoneRoute) (paused bool, err error) {
	// #865: every exit of this function but the PAUSE is terminal for
	// the route — it landed, it was cancelled, it was never a move, or
	// it was refused — so the caller's continuation runs from all of
	// them, from here, rather than from each branch remembering to. A
	// pause is not terminal: the resume reaches the tail later, through
	// executeZoneRouteLocked. The landing path has already run the tail
	// through this same pointer by the time this defer fires, and
	// runRouteTailLocked clears the continuation as it runs, so it
	// cannot run twice.
	defer func() {
		if paused {
			return
		}
		if tailErr := g.runRouteTailLocked(&r); tailErr != nil && err == nil {
			err = tailErr
		}
	}()
	src := g.findCardZoneLocked(r.CardID)
	if src == nil {
		return false, ErrCardNotFound
	}
	dstZone, _, err := g.routeDestinationLocked(r.CardID, r.Dst, r.DstOwner)
	if err != nil {
		return false, err
	}
	if dstZone == src {
		// Already there. Historically each mover short-circuited this
		// itself (exile-an-exiled-card, bounce-a-card-in-hand); the
		// pipeline must not fire for a move that isn't one. The
		// caller's continuation still runs: "nothing to do" is an
		// answer, and a batch sequenced through the tail needs it.
		return false, nil
	}

	ev := &ReplacementEvent{
		Kind:            RepEventMove,
		Actor:           r.Actor,
		CardID:          r.CardID,
		OldZone:         src.Kind,
		NewZone:         r.Dst,
		NewZoneOwner:    dstZone.Owner,
		zoneRoute:       &r,
		asCommanderMove: r.AsCommander,
		mustSettleNow:   r.MustSettleNow,
	}
	if r.Discard {
		// #650: a discard is its own event kind, because what a discard
		// replacement watches for is the DISCARD and not the zone move
		// underneath it. Everything else on the event is the same, and
		// every exit site downstream reads the two kinds together
		// (isExitMove).
		ev.Kind = RepEventDiscard
		ev.DiscardPlayer = r.Actor
		ev.DiscardCause = r.DiscardCause
		if ev.DiscardCause == "" {
			ev.DiscardCause = DiscardCauseEffect
		}
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// A prompt is queued and the resume owns the move now. The
		// event's tracking map entry deliberately survives: the
		// resume re-enters the apply-loop with the same ev.ID.
		return true, nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return false, err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		// CR 614.10 with a null replacement: the move simply does not
		// happen. The caller's continuation still runs — see
		// runRouteTailLocked.
		return false, nil
	}
	return false, g.executeZoneRouteLocked(out)
}

// executeZoneRouteLocked performs the physical move for a settled
// RepEventMove carrying a zoneRoute. Shared by the inline path above
// and by the resume path (ResolveOptionalReplacement →
// applyResolvedReplacementEventLocked), so a paused move and an
// unpaused one cannot drift apart.
//
// The event's NewZone / NewZoneOwner are authoritative — a
// replacement may have rewritten them (CR 903.9 to the command zone,
// Stone of Erech's graveyard → exile, Whip of Erebos exiling what
// would leave the battlefield). The route's per-destination flags are
// re-read against the SETTLED destination, not the requested one.
//
// Caller must hold g.mu.
func (g *Game) executeZoneRouteLocked(ev *ReplacementEvent) (err error) {
	if ev == nil || ev.zoneRoute == nil {
		return ErrInvalidParam
	}
	// #865: this function is only ever reached for a route that is
	// going to settle here, so EVERY exit of it is terminal — the move
	// landed, it was not a move at all, or it was refused — and the
	// caller's continuation runs from all of them. A refused move is as
	// terminal as a completed one, and a batch waiting on the tail must
	// not be left waiting; finishBattlefieldLeaveLocked makes the same
	// call for the other exit.
	//
	// #866: a sequenced batch carries its pre-move copies on the route
	// so the leg that runs on the far side of a CR 903.9 prompt is
	// still part of the same simultaneous exit — the same thing
	// finishBattlefieldLeaveLocked does for the destroy route. It is
	// registered FIRST so LIFO runs it LAST: the continuation, like
	// that one's, runs with the batch still open. Empty for every
	// unbatched move, where publishing is a no-op.
	defer g.publishSimultaneousExitLocked(ev.zoneRoute.simultaneousExit)()
	defer func() {
		if tailErr := g.runRouteTailLocked(ev.zoneRoute); tailErr != nil && err == nil {
			err = tailErr
		}
	}()
	r := *ev.zoneRoute
	src := g.findCardZoneLocked(ev.CardID)
	if src == nil {
		return ErrCardNotFound
	}
	dstZone, actor, err := g.routeDestinationLocked(ev.CardID, ev.NewZone, ev.NewZoneOwner)
	if err != nil {
		return err
	}
	if dstZone == src {
		return nil
	}
	if r.Actor != uuid.Nil {
		actor = r.Actor
	}
	redirected := dstZone.Kind != r.Dst

	fromBattlefield := src.Kind == ZoneBattlefield
	if fromBattlefield {
		// LKI, and the CR 400.7 forget — battlefield_exit.go.
		g.battlefieldExitLocked(ev.CardID)
	}
	// The pre-move card, for CR 708.9 below: MoveCard clears the
	// face-down state on the way through (CR 400.7), so "was this a
	// face-down permanent" can only be asked before it runs.
	before, hadBefore := g.cardInZoneLocked(src, ev.CardID)
	if _, err := MoveCard(src, dstZone, ev.CardID); err != nil {
		return err
	}
	// CR 708.9: a face-down PERMANENT that moves to another zone is
	// revealed by its owner. FIRST, before the destination's own
	// knowledge rule below, and that order is the rule: the reveal is
	// what every player SAW, and the destination then decides what
	// they still KNOW. A morph tucked into a library is revealed to
	// the table and then lost in it (CR 401.2); one exiled face down
	// is revealed and then unreadable again. Reveal last would leave
	// every seat able to read a library card by position.
	if hadBefore {
		g.revealFaceDownExitLocked(before)
	}

	// Destination bookkeeping. All of it is keyed on where the card
	// actually LANDED, so a redirect to the command zone cannot carry
	// a face-down flag or a to-the-bottom instruction with it.
	//
	// MoveCard has already cleared the face-down state for every
	// destination (ADR 0069 decision 5), so the only thing left to do
	// here is set it again when the destination IS a face-down state.
	// The two former `FaceDown = false` arms are gone with it.
	faceDown := r.FaceDown != FaceDownNone && !redirected && dstZone.Kind == ZoneExile
	switch {
	case faceDown:
		// Who may look is the kind's answer (CR 406.3 for a plain
		// exile: nobody, the player who exiled it included; CR
		// 702.143d for a foretold card: its owner). Replacing the
		// knowledge set rather than leaving it alone matters — a
		// scryed library card has a knower, and carrying that in
		// would let exactly one seat read a card nobody may.
		g.applyFaceDownLandingLocked(dstZone, ev.CardID, r.FaceDown)
	case dstZone.Kind == ZoneLibrary:
		// A library is a hidden zone (CR 401.2). Whoever could read
		// this card a moment ago cannot now.
		for i := range dstZone.Cards {
			if dstZone.Cards[i].InstanceID == ev.CardID {
				dstZone.Cards[i].ClearKnown()
				break
			}
		}
	}
	if !faceDown {
		g.markCardKnownInZoneLocked(dstZone, ev.CardID)
	}
	if (r.ToBottom || r.Depth > 1) && !redirected && dstZone.Kind == ZoneLibrary {
		c, err := dstZone.Remove(ev.CardID)
		if err != nil {
			return err
		}
		if r.ToBottom {
			dstZone.PushBottom(c)
		} else {
			dstZone.InsertFromTop(c, r.Depth)
		}
	}
	if r.DropStackMeta {
		delete(g.StackMeta, ev.CardID)
		g.recomputeSplitSecondLocked()
	}

	// Events. A counterspell keeps its historic single-event shape;
	// so does a discard; everything else emits a zone move, promoted
	// to EventMill when the card really did land in a graveyard off a
	// mill. One event per route, never two — Syr Konrad and Bloodchief
	// Ascension both watch the whole family and would double-count a
	// move that emitted its own kind AND a zone move.
	switch {
	case r.Countered:
		g.EmitEvent(Event{
			Kind:   EventCounterSpell,
			Target: ev.CardID,
			CardID: ev.CardID,
		})
	case r.Discard:
		// CR 701.8a: the discard is the move OUT of the hand, so it
		// happened whatever the window did with the destination — a
		// commander put into the command zone instead was still
		// discarded, Library of Leng putting it on top of the library
		// was still a discard, and Megrim, Containment Construct and
		// the rest of the family still see it. NewZone is where the
		// card really went, so a listener that cares can tell.
		//
		// #650: it also carries the CAUSE now, so the log can say why
		// and a payoff that cares ("a spell or ability an opponent
		// controls causes you to discard") has it beside the Source.
		g.EmitEvent(Event{
			Kind:         EventDiscardCard,
			Actor:        actor,
			Source:       r.Source,
			CardID:       ev.CardID,
			OldZone:      src.Kind,
			NewZone:      dstZone.Kind,
			DiscardCause: ev.DiscardCause,
		})
	default:
		kind := EventZoneMove
		if r.Mill && dstZone.Kind == ZoneGraveyard {
			kind = EventMill
		}
		g.EmitEvent(Event{
			Kind:    kind,
			Actor:   actor,
			Source:  r.Source,
			CardID:  ev.CardID,
			OldZone: src.Kind,
			NewZone: dstZone.Kind,
		})
	}
	if fromBattlefield {
		g.EmitEvent(Event{
			Kind:    EventLTB,
			Actor:   actor,
			CardID:  ev.CardID,
			NewZone: dstZone.Kind,
		})
		// A permanent leaving the battlefield can invalidate a queued
		// "sacrifice a creature of your choice" prompt (Grave Pact).
		// Same reason executeBattlefieldLeaveLocked re-checks here:
		// the state-check loop does not run while a choice is queued.
		g.pruneSacrificeChoicesLocked()
	}
	// #605: a card that has just landed invalidates any OTHER queued
	// prompt still asking about a move of the same card out of the
	// zone it has now left. Unconditional — an exit from the stack or
	// a graveyard can strand a sibling prompt just as a battlefield
	// one can.
	g.pruneStaleZoneChangeChoicesLocked()
	// #853: and now the rest of whatever asked for this move, with the
	// card already where it is going and its event already emitted, so
	// a continuation that starts the next move or reads the log sees
	// this one finished. Last, because it may queue the next prompt —
	// which is why it is the deferred terminal outcome above rather
	// than a line here.
	return nil
}

// routeDestinationLocked resolves a destination zone kind + owner to
// a concrete *Zone and the actor to stamp on the move's events.
//
// A per-player destination with no named owner resolves to the CARD's
// owner, which is what every CR 903.9 destination means ("its owner's
// graveyard", "its owner's hand"). An unseated owner falls back the
// way the pre-existing movers already did: exile for a graveyard or
// command-zone destination — so the engine never carries a reference
// into a dead player's zone — and ErrPlayerNotFound for hand and
// library, where there is no sensible public substitute.
//
// Caller must hold g.mu.
func (g *Game) routeDestinationLocked(cardID uuid.UUID, dst ZoneKind, dstOwner uuid.UUID) (*Zone, uuid.UUID, error) {
	switch dst {
	case ZoneExile:
		return g.Exile, uuid.Nil, nil
	case ZoneBattlefield, ZoneStack:
		// Not an exit. The entry path owns these.
		return nil, uuid.Nil, ErrZoneNotFound
	}
	owner := dstOwner
	if owner == uuid.Nil {
		card, ok := g.LookupCardForEffect(cardID)
		if !ok {
			return nil, uuid.Nil, ErrCardNotFound
		}
		owner = card.Owner
	}
	p := g.playerByIDLocked(owner)
	if p == nil {
		if dst == ZoneGraveyard || dst == ZoneCommand {
			return g.Exile, uuid.Nil, nil
		}
		return nil, uuid.Nil, ErrPlayerNotFound
	}
	switch dst {
	case ZoneGraveyard:
		return p.Graveyard, p.ID, nil
	case ZoneHand:
		return p.Hand, p.ID, nil
	case ZoneLibrary:
		return p.Library, p.ID, nil
	case ZoneCommand:
		return p.Command, p.ID, nil
	}
	return nil, uuid.Nil, ErrZoneNotFound
}
