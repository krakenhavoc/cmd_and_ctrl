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

	// FaceDown enters the permanent FACE DOWN in this state (CR 708,
	// ADR 0069) — FaceDownManifested for manifest, and later
	// FaceDownCloaked for cloak. The zero value is an ordinary face-up
	// entry, which is every other caller.
	//
	// It changes three things about the entry, and all three are the
	// point of it:
	//
	//   - the CR 110.4 nonpermanent refusal is lifted. Manifest puts
	//     ANY card onto the battlefield face down (CR 701.40a); the
	//     object that arrives is a creature whatever the card is.
	//   - the entry sets the face-down state and its viewers instead
	//     of clearing the flag and marking every seat a knower. Exile
	//     and the battlefield are both PUBLIC zones; this is the one
	//     line that separates a public zone from a public object.
	//   - because the state is set before phase 3 announces,
	//     CatalogKey answers "" for the permanent by the time EventETB
	//     fires and the AsEnters hook runs — so a face-down entry
	//     runs no ETB trigger and no "as enters" choice, which is
	//     CR 708.2a falling out rather than being special-cased.
	FaceDown FaceDownKind
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
// paused. Refuses a card not in a library with ErrCardNotFound, and a
// nonpermanent card (CR 110.4) or a token (CR 111.8: a token that has
// left the battlefield can't come back onto it) with ErrInvalidParam,
// moving nothing.
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
// a library (ErrCardNotFound) or not a permanent card — a nonpermanent
// or a token (ErrInvalidParam) — refuses the whole batch and moves
// nothing, so a
// caller holding one stale ID cannot half-apply an effect. A repeated
// ID is taken once.
//
// Caller must hold g.mu.
func (g *Game) PutCardsFromLibraryOntoBattlefieldForEffect(ids []uuid.UUID, opts LibraryEntryOptions) ([]uuid.UUID, error) {
	return g.putOntoBattlefieldFromZoneLocked(ids, ZoneLibrary, opts)
}

// ManifestForEffect is CR 701.40a: put the top card of playerID's
// library onto the battlefield face down as a 2/2 creature, under
// playerID's control. Returns the manifested permanent's ID, or
// uuid.Nil with a nil error when the library is empty or a
// replacement canceled, redirected or paused the entry.
//
// It is the ordinary library→battlefield entry with
// ZoneEntryOptions.FaceDown set, which is the whole of manifest that
// is not a card: the CR 614 replacement window, the entry pipeline,
// the events and the undo path all come from
// putOntoBattlefieldFromZoneLocked unchanged. What the face-down
// option adds is on that field's doc comment.
//
// The card it manifests need not be a permanent card (CR 701.40a),
// and it is NOT revealed on the way — the controller becomes its only
// knower (CR 708.5), which is why this cannot go through the reveal
// helpers a search uses.
//
// MANIFEST THE MECHANIC IS NOT SHIPPED. This is the primitive that
// proves ADR 0069's object model — the 2/2 projection, the catalog
// suppression, the CR 708.5 viewers rule and the CR 708.9 reveal —
// and the one morph, megamorph, disguise and cloak build on (#95).
// No catalog card calls it yet; the turn-face-up special action
// (CR 116.2g) and the face-down CAST (CR 708.4) are #95's.
//
// Caller must hold g.mu.
func (g *Game) ManifestForEffect(playerID uuid.UUID) (uuid.UUID, error) {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return uuid.Nil, ErrPlayerNotFound
	}
	if p.Library.Size() == 0 {
		// An empty library is not an error and is not a draw, the
		// same posture ExileTopFaceDownForEffect takes.
		return uuid.Nil, nil
	}
	// The library's top is the LAST element — the same read PopTop
	// and ExileTopFaceDownForEffect use.
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	entered, err := g.putOntoBattlefieldFromZoneLocked([]uuid.UUID{top}, ZoneLibrary, ZoneEntryOptions{
		Controller: playerID,
		FaceDown:   FaceDownManifested,
	})
	if err != nil || len(entered) == 0 {
		return uuid.Nil, err
	}
	return entered[0], nil
}

// pendingPut is one card of a putOntoBattlefieldFromZoneLocked batch
// between its pipeline (phase 1) and its announcement (phase 3).
type pendingPut struct {
	cardID     uuid.UUID
	src        *Zone
	controller uuid.UUID
	// priorController is the Controller the card carried in its
	// source zone, put back if it does not enter.
	priorController uuid.UUID
	out             *ReplacementEvent
	eventID         ReplacementEventID
	entered         bool
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
		// CR 110.4: only a permanent card can be put onto the
		// battlefield — unless it is going there FACE DOWN, in which
		// case the object that arrives is a 2/2 creature whatever the
		// card says (CR 701.40a, CR 708.2). Manifesting an instant is
		// legal and common.
		if !c.IsPermanent() && opts.FaceDown == FaceDownNone {
			return nil, ErrInvalidParam
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
			return nil, ErrInvalidParam
		}
		controller := opts.Controller
		if controller == uuid.Nil || g.playerByIDLocked(controller) == nil {
			controller = src.Owner
		}
		batch = append(batch, &pendingPut{cardID: id, src: src, controller: controller, priorController: c.Controller})
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
		// NO entryResumable, and this is THE site the flag's doc
		// comment and the shockland / MDFC-land caveats name: the
		// three phases run every card's pipeline against the
		// pre-entry board and then move them together, and a per-card
		// resume would finish one card's entry after the others had
		// landed, which is the simultaneity this function exists for.
		// So an effect that would pause here (a pay-life entry
		// choice, a Clone pick) takes the un-paid branch instead —
		// weaker than printed, never stronger. See
		// ReplacementEvent.entryResumable and
		// offerEntryLifePaymentLocked.
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
			// Whatever a hidden zone had recorded about who could see
			// this card (a scry, a look) is superseded: the
			// battlefield is public, and markCardKnownInZoneLocked
			// below makes every seat a knower.
			//
			// MoveCard has already cleared the face-down state
			// (CR 400.7, ADR 0069 decision 5), so the former
			// `FaceDown = false` here is gone; a FACE-DOWN entry sets
			// it back, below, instead of marking the table.
			g.Battlefield.Cards[i].ClearKnown()
			break
		}
		if opts.FaceDown != FaceDownNone {
			// CR 708.5: the controller of a face-down permanent may
			// look at it, and nobody else may — so this replaces the
			// public-zone marking rather than adding to it.
			g.applyFaceDownLandingLocked(g.Battlefield, moved.InstanceID, opts.FaceDown)
		} else {
			g.markCardKnownInZoneLocked(g.Battlefield, moved.InstanceID)
		}
		g.applyEntryCountersLocked(moved.InstanceID, p.out.EntersWithCounters)
		p.entered = true
		entered = append(entered, moved.InstanceID)
	}

	// A card that did not enter (canceled, redirected, paused, or its
	// move failed) is still in its source zone wearing the controller
	// phase 1 stamped for the entry. Give it back what it had, so a put
	// under someone other than the owner (Lonis) leaves no trace on a
	// card it never moved.
	for _, p := range batch {
		if p.entered {
			continue
		}
		for i := range p.src.Cards {
			if p.src.Cards[i].InstanceID == p.cardID {
				p.src.Cards[i].Controller = p.priorController
				break
			}
		}
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
		// Nothing moved, so no zone lost a card and no open pick can
		// have been invalidated by this call.
		return nil, firstErr
	}
	// #1069: the batch has left the hand or the library it came from,
	// so an open choose_cards prompt over that zone — a discard prompt
	// whose candidates a Warp World just put onto the battlefield — is
	// trimmed or withdrawn here. One call for the whole batch: the
	// prune is keyed by zone and re-reads every open pick
	// (battlefield_entry.go). After phase 3, so every arrival is
	// announced and every ETB hook has run before a withdrawal's
	// continuation can start something of its own.
	g.pruneChoicesAfterArrivalLocked()
	return entered, firstErr
}
