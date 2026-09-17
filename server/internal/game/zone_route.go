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
// sequences itself through it rather than looping, which is the
// opposite of what the mill above does, and discard.go says why.

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

	// FaceDown exiles the card face down (CR 406.3, Necropotence).
	// Face-down exile is the one destination that must NOT mark the
	// table as knowers — see ExileTopFaceDownForEffect.
	FaceDown bool

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

	// Source is the card whose effect asked for the move, stamped on
	// the emitted event. Read only by the Discard leg today, which is
	// the only route whose event has ever carried one; uuid.Nil
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

// routeCardToZoneLocked opens the CR 614 replacement window for a
// move of r.CardID into r.Dst and, once the pipeline settles on a
// destination, performs the move through executeZoneRouteLocked.
//
// Returns paused=true when the pipeline queued a player prompt. In
// that case NOTHING has moved: the card is still sitting in its old
// zone, no event has been emitted, and the move completes from
// applyResolvedReplacementEventLocked when the player answers. Every
// caller has to be able to tolerate that, which is why the bool is
// returned rather than swallowed — the mill loop in particular has to
// know not to pop the same card again.
//
// A caller with something to do AFTER the move hands it over as
// r.then rather than writing it on the next line, and then ignores the
// bool: the continuation runs from the landing either way, inline when
// nothing paused and from the resume when something did. That is how
// a multi-card discard sequences itself (#853).
//
// Caller must hold g.mu.
func (g *Game) routeCardToZoneLocked(r zoneRoute) (paused bool, err error) {
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
		return false, g.runRouteTailLocked(&r)
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
		return false, g.runRouteTailLocked(&r)
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
func (g *Game) executeZoneRouteLocked(ev *ReplacementEvent) error {
	if ev == nil || ev.zoneRoute == nil {
		return ErrInvalidParam
	}
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
		return g.runRouteTailLocked(ev.zoneRoute)
	}
	if r.Actor != uuid.Nil {
		actor = r.Actor
	}
	redirected := dstZone.Kind != r.Dst

	fromBattlefield := src.Kind == ZoneBattlefield
	if fromBattlefield {
		g.snapshotLKILocked(ev.CardID)
	}
	if _, err := MoveCard(src, dstZone, ev.CardID); err != nil {
		return err
	}

	// Destination bookkeeping. All of it is keyed on where the card
	// actually LANDED, so a redirect to the command zone cannot carry
	// a face-down flag or a to-the-bottom instruction with it.
	faceDown := r.FaceDown && !redirected && dstZone.Kind == ZoneExile
	for i := range dstZone.Cards {
		if dstZone.Cards[i].InstanceID != ev.CardID {
			continue
		}
		switch {
		case faceDown:
			// CR 406.3: nobody may look at it, including the player
			// who exiled it. Clearing rather than leaving the set
			// alone matters — a scryed library card has a knower.
			dstZone.Cards[i].FaceDown = true
			dstZone.Cards[i].ClearKnown()
		case dstZone.Kind == ZoneLibrary:
			// A library is a hidden zone (CR 401.2). Whoever could
			// read this card a moment ago cannot now.
			dstZone.Cards[i].FaceDown = false
			dstZone.Cards[i].ClearKnown()
		default:
			// CR 400.7 / 708.2: "face down" belongs to an object in a
			// zone, and a card that changes zones is a new object
			// with no memory of it.
			dstZone.Cards[i].FaceDown = false
		}
		break
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
		// happened whatever the CR 903.9 window did with the
		// destination — a commander put into the command zone instead
		// was still discarded, and Megrim, Containment Construct and
		// the rest of the family still see it. NewZone is where the
		// card really went, so a listener that cares can tell.
		g.EmitEvent(Event{
			Kind:    EventDiscardCard,
			Actor:   actor,
			Source:  r.Source,
			CardID:  ev.CardID,
			OldZone: src.Kind,
			NewZone: dstZone.Kind,
		})
	default:
		kind := EventZoneMove
		if r.Mill && dstZone.Kind == ZoneGraveyard {
			kind = EventMill
		}
		g.EmitEvent(Event{
			Kind:    kind,
			Actor:   actor,
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
	// this one finished. Last, because it may queue the next prompt.
	return g.runRouteTailLocked(ev.zoneRoute)
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
