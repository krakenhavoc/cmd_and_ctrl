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
// What this deliberately does not do:
//
//   - MoveCardByIDAsCommander keeps its own pipeline call. It is the
//     admin / sandbox move and accepts an arbitrary source AND
//     destination, the battlefield included, which an exit primitive
//     has no business expressing.
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

	// Countered flags a counterspell, whose completed move emits
	// EventCounterSpell INSTEAD of EventZoneMove (the historic shape
	// — counter watchers key on it and a zone move would double-
	// count). DropStackMeta additionally retires the stack item.
	Countered     bool
	DropStackMeta bool
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
		// pipeline must not fire for a move that isn't one.
		return false, nil
	}

	ev := &ReplacementEvent{
		Kind:         RepEventMove,
		Actor:        r.Actor,
		CardID:       r.CardID,
		OldZone:      src.Kind,
		NewZone:      r.Dst,
		NewZoneOwner: dstZone.Owner,
		zoneRoute:    &r,
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
		return nil
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
	// everything else emits a zone move, promoted to EventMill when
	// the card really did land in a graveyard off a mill.
	switch {
	case r.Countered:
		g.EmitEvent(Event{
			Kind:   EventCounterSpell,
			Target: ev.CardID,
			CardID: ev.CardID,
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
