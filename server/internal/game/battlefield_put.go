package game

import (
	"errors"

	"github.com/google/uuid"
)

// battlefield_put.go — "put a card onto the battlefield" from a hand
// or a library, without casting it, playing it or searching for it
// (#654 for the hand, #745 for the library).
//
// The two moves are one body. #728 shipped the hand version first and
// the library version is the same four steps with a different source
// zone, so rather than a second copy of the entry pipeline the hand
// entry point and the two library entry points all call
// putOntoBattlefieldFromZoneLocked. What differs between a hand and a
// library is only what the caller is allowed to name: a hand move
// refuses a library card and the other way round, so a stale ID can
// never rip a card out of a zone the effect did not print.
//
// The library half is what "reveal the top card of your library. If
// it's a land card, put it onto the battlefield" (Coiling Oracle),
// "you may put any number of permanent cards … from among them onto
// the battlefield" (Genesis Wave) and "reveal cards until you reveal a
// creature card. Put that card onto the battlefield" (Atla Palani)
// needed. Before #745 the only library-to-battlefield move in the
// engine was searchEnterBattlefieldLocked, private to a search, which
// emits EventSearchLibrary and may shuffle — neither of which a reveal
// does.

// ZoneEntryOptions are the modifiers a "put a card onto the
// battlefield" effect applies to the entry it causes. The zero value
// is the printed common case — an untapped permanent under its owner's
// control — so an ordinary caller passes an empty struct, exactly as
// CreateTokenForEffect does with TokenEntryOptions.
type ZoneEntryOptions struct {
	// Controller is who the permanent enters under. uuid.Nil, or a
	// player who has left, means "under its owner's control", which
	// is what almost every card in this family prints ("put a land
	// card from YOUR hand", "reveal the top card of YOUR library").
	// The field exists because the move is generic: Lonis's "put
	// [a card from target opponent's library] onto the battlefield
	// under your control" has nowhere else to say so, and the CR 614
	// pipeline has to know the answer before it runs.
	Controller uuid.UUID

	// Tapped is the putting effect's own "onto the battlefield
	// TAPPED" clause (Arboreal Grazer, Risen Reef). It is SEEDED onto
	// the replacement event rather than OR-ed in after the pipeline,
	// the way SearchLibrarySpec.TappedOnEntry is: the effect's clause
	// and whatever CR 614 adds on top then settle in one field, and no
	// reader downstream can lose one of the two.
	Tapped bool
}

// LibraryEntryOptions are ZoneEntryOptions for a library source, named
// so a library caller reads as one.
type LibraryEntryOptions = ZoneEntryOptions

// PutFromLibraryOntoBattlefieldForEffect puts one card from a library
// onto the battlefield without casting it or searching for it —
// Coiling Oracle's "if it's a land card, put it onto the battlefield",
// Chaos Warp's "if it's a permanent card, they put it onto the
// battlefield".
//
// Everything PutFromHandOntoBattlefieldForEffect's comment says holds
// here, with "library" for "hand": the controller is stamped while the
// card is still in the library, the CR 614 pipeline runs before it
// leaves, the move is followed by EventZoneMove, EventETB and the
// catalog's AsEnters hook, and a land put this way is not a land drop
// (CR 305.4). It is not entryResumable either — no Clone choice and
// no entry prompt, the declared simplification the search and hand
// paths already carry.
//
// It does NOT emit EventSearchLibrary and does NOT shuffle. That is
// the whole difference from a search, and the reason this is not
// SearchLibraryThenForEffect with a predicate: nothing was searched
// for, and a "whenever a player searches their library" payoff must
// not see a Coiling Oracle.
//
// Knowledge: whatever the library had recorded about the card (a
// scry, a look, a reveal) is replaced by the battlefield's own —
// every seat knows a permanent — so no seat is left "knowing" a card
// through a library position it no longer holds.
//
// Returns the entering permanent's ID (the library instance ID; only
// an exile return mints a new one), or uuid.Nil with a nil error
// when a replacement canceled or redirected the entry or the pipeline
// paused. Refuses a card not in a library with ErrCardNotFound and a
// nonpermanent card with ErrInvalidParam (CR 110.4), moving nothing.
//
// Caller must hold g.mu.
func (g *Game) PutFromLibraryOntoBattlefieldForEffect(cardID uuid.UUID, opts LibraryEntryOptions) (uuid.UUID, error) {
	entered, err := g.putOntoBattlefieldFromZoneLocked([]uuid.UUID{cardID}, ZoneLibrary, opts)
	if err != nil || len(entered) == 0 {
		return uuid.Nil, err
	}
	return entered[0], nil
}

// PutCardsFromLibraryOntoBattlefieldForEffect puts several library
// cards onto the battlefield as ONE simultaneous entry — Genesis
// Wave's "any number of permanent cards", Animist's Awakening's "all
// land cards from among them".
//
// # What "simultaneous" buys, and what it does not
//
// The moves themselves are still one card at a time; nothing in the
// engine can observe a move between two others. What CAN be observed
// is the ORDER of the three phases, and that is what this function
// fixes (CR 614.12, CR 603.6a):
//
//  1. Every card's CR 614 pipeline runs first, against the board as
//     it stood before ANY of them entered. A check land ("enters
//     tapped unless you control a Forest") put alongside a Forest
//     does not see that Forest, which is what simultaneous entry
//     means. Entering them one by one would let the second land see
//     the first and come in untapped — stronger than printed.
//  2. Every card that survived its pipeline moves.
//  3. Only then do the zone moves, ETB events and AsEnters hooks
//     fire, so every permanent of the batch is already on the
//     battlefield when the first event is harvested: a Soul Warden
//     put alongside two creatures sees both of them, as it would.
//
// # The declared simplification that remains
//
// The AsEnters hooks (CR 614.12 "as this enters, choose …") run one
// after another in phase 3, after every card has landed, so a choice
// one of them makes can see the rest of the batch on the battlefield.
// No catalog AsEnters choice reads the other permanents entering with
// it today; the first one that does would err in whichever direction
// its choice leans, and should say so on its own card.
//
// A card whose pipeline PAUSES (a CR 616 ordering prompt) stays in the
// library and is not part of the entry, and a card a replacement
// cancels or redirects likewise stays where it was — the same posture
// as the single-card move, and weaker than printed rather than
// stronger. Returns the IDs that entered, in the order given.
//
// Every ID is validated before anything happens: one that is not in
// a library (ErrCardNotFound) or not a permanent card
// (ErrInvalidParam) refuses the whole batch and moves nothing, so a
// caller holding one stale ID cannot half-apply an effect. A repeated
// ID is taken once.
//
// Caller must hold g.mu.
func (g *Game) PutCardsFromLibraryOntoBattlefieldForEffect(ids []uuid.UUID, opts LibraryEntryOptions) ([]uuid.UUID, error) {
	return g.putOntoBattlefieldFromZoneLocked(ids, ZoneLibrary, opts)
}

// pendingPut is one card of a putOntoBattlefieldFromZoneLocked batch
// between its pipeline (phase 1) and its announcement (phase 3).
type pendingPut struct {
	cardID     uuid.UUID
	src        *Zone
	controller uuid.UUID
	out        *ReplacementEvent
	eventID    ReplacementEventID
	entered    bool
}

// putOntoBattlefieldFromZoneLocked is the shared body of the hand and
// library moves. `from` is the only zone kind a named card may be in;
// see PutCardsFromLibraryOntoBattlefieldForEffect for the three phases
// and PutFromHandOntoBattlefieldForEffect for the steps each card owes.
//
// Caller must hold g.mu.
func (g *Game) putOntoBattlefieldFromZoneLocked(ids []uuid.UUID, from ZoneKind, opts ZoneEntryOptions) ([]uuid.UUID, error) {
	// Validate the whole batch before touching any of it.
	batch := make([]*pendingPut, 0, len(ids))
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		src := g.findCardZoneLocked(id)
		if src == nil || src.Kind != from {
			return nil, ErrCardNotFound
		}
		c, ok := g.cardInZoneLocked(src, id)
		if !ok {
			return nil, ErrCardNotFound
		}
		if !c.IsPermanent() {
			return nil, ErrInvalidParam
		}
		controller := opts.Controller
		if controller == uuid.Nil || g.playerByIDLocked(controller) == nil {
			controller = src.Owner
		}
		batch = append(batch, &pendingPut{cardID: id, src: src, controller: controller})
	}
	if len(batch) == 0 {
		return nil, nil
	}

	// Phase 1: stamp every controller, then run every pipeline, all
	// against the pre-entry board. Controllers go first so a
	// replacement consulted for one card of the batch reads the right
	// controller on the others.
	for _, p := range batch {
		for i := range p.src.Cards {
			if p.src.Cards[i].InstanceID == p.cardID {
				p.src.Cards[i].Controller = p.controller
				break
			}
		}
	}
	defer func() {
		for _, p := range batch {
			if p.out != nil {
				g.clearReplacementEventLocked(p.eventID)
			}
		}
	}()
	var firstErr error
	for _, p := range batch {
		ev := &ReplacementEvent{
			Kind:         RepEventMove,
			Actor:        p.controller,
			CardID:       p.cardID,
			OldZone:      from,
			NewZone:      ZoneBattlefield,
			EntersTapped: opts.Tapped,
		}
		out, err := g.applyReplacementsLocked(ev)
		if errors.Is(err, errReplacementPending) {
			// A CR 616 ordering prompt is open. Nothing has moved;
			// this card stays where it is and is not part of the put.
			continue
		}
		if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
			g.clearReplacementEventLocked(ev.ID)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if out == nil || out.Canceled || out.NewZone != ZoneBattlefield {
			// Canceled, or redirected somewhere else by a
			// replacement. There is no generic "put it wherever the
			// pipeline said" helper for these sources, so a redirect
			// is treated as a cancel rather than guessed at — the
			// posture the exile-return path takes.
			g.clearReplacementEventLocked(ev.ID)
			continue
		}
		p.out, p.eventID = out, ev.ID
	}

	// Phase 2: move every card whose pipeline settled on the
	// battlefield.
	entered := make([]uuid.UUID, 0, len(batch))
	for _, p := range batch {
		if p.out == nil {
			continue
		}
		moved, err := MoveCard(p.src, g.Battlefield, p.cardID)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID != moved.InstanceID {
				continue
			}
			g.Battlefield.Cards[i].Controller = p.controller
			// out.EntersTapped carries both inputs: the putting
			// effect's own "tapped" clause, seeded onto the event, and
			// whatever the CR 614 pipeline added on top. The permanent
			// ARRIVES tapped — no tap event fires.
			if p.out.EntersTapped {
				g.Battlefield.Cards[i].Tapped = true
			}
			// The impulse grant is spent by the entry, exactly as it
			// is on the land-play and cast branches.
			g.Battlefield.Cards[i].ExilePlay = ExilePlayPermission{}
			// Whatever a hidden zone had recorded about who could see
			// this card (a scry, a look) is superseded: the
			// battlefield is public, and markCardKnownInZoneLocked
			// below makes every seat a knower.
			g.Battlefield.Cards[i].ClearKnown()
			g.Battlefield.Cards[i].FaceDown = false
			break
		}
		g.markCardKnownInZoneLocked(g.Battlefield, moved.InstanceID)
		for name, n := range p.out.EntersWithCounters {
			_ = g.AddCounterForEffect(moved.InstanceID, name, n)
		}
		p.entered = true
		entered = append(entered, moved.InstanceID)
	}

	// Phase 3: announce. Every permanent of the batch is on the
	// battlefield before the first event, so a watcher among them
	// sees the rest.
	for _, p := range batch {
		if !p.entered {
			continue
		}
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   p.controller,
			CardID:  p.cardID,
			OldZone: from,
			NewZone: ZoneBattlefield,
		})
		g.EmitEvent(Event{Kind: EventETB, Actor: p.controller, CardID: p.cardID})
	}
	for _, p := range batch {
		if !p.entered {
			continue
		}
		if c, ok := g.cardInZoneLocked(g.Battlefield, p.cardID); ok {
			g.fireETBHookLocked(p.cardID, CatalogKey(c))
		}
	}
	if len(entered) == 0 {
		return nil, firstErr
	}
	return entered, firstErr
}
