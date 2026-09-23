package game

import (
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

	// Attacking is the putting effect's own "onto the battlefield
	// ATTACKING" clause (CR 506.3c) — ninjutsu's "Put this card onto
	// the battlefield from your hand tapped and attacking"
	// (CR 702.49a, #1227). It names the player, planeswalker or battle
	// the permanent attacks, with Card.AttackingTarget's own overloaded
	// domain (attack_target.go); the zero value is every other caller,
	// which enters not attacking.
	//
	// SEEDED onto the replacement event exactly as Tapped is, and for
	// the same reason: the entry's one settled record carries it, so
	// no reader downstream can lose it and a CR 616 resume that holds
	// only the event still knows the entry was an attacking one.
	//
	// The permanent was never DECLARED as an attacker, so it fires no
	// "whenever ~ attacks" trigger and nothing watching the declaration
	// sees it. That whole rule is stampEntryAttackerLocked's, shared
	// with the token door, and DeclareAttackerWith is deliberately NOT
	// the route: it refuses outside StepDeclareAttackers, it taps, it
	// checks summoning sickness and it emits EventAttack — wrong on all
	// four counts for a CR 506.3c entry.
	Attacking uuid.UUID

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

	// FaceDownListed is the body the putting effect LISTED for a
	// face-down entry (CR 708.2) — Cybership's "They're 2/2 Cyberman
	// artifact creatures". nil is CR 708.2a's default 2/2, which is
	// what manifest and cloak get. Ignored without FaceDown. #1270.
	FaceDownListed *FaceDownListing
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
// (CR 305.4). Since #1322 the entry can pause for its own question — a
// shockland's life, a Clone's copy, a reveal-land's reveal — in which
// case this returns uuid.Nil and the card lands when the answer
// arrives; see entry_batch.go.
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
// # A question pauses the whole batch (#1322)
//
// A card whose window asks something — a shockland's "you may pay 2
// life", a Clone's "copy what?", a reveal-land's reveal, a CR 616
// ordering prompt — stops phase 1 there. Nothing has moved; the
// answer's resume hands the settled event back to the batch, which
// opens the next card's window, and the batch lands when the last one
// settles (CR 614.12a: the choice is made before the permanent
// enters). This door then returns nil with a nil error, and a caller
// that needs to know what entered uses
// PutCardsFromLibraryOntoBattlefieldThenForEffect. See entry_batch.go.
//
// # The declared simplification that remains
//
// The AsEnters hooks that are ETB-hook prompts rather than CR 614
// windows ("as this enters, choose a creature type", which
// creature_type_choice.go queues from the hook) run one after another
// in phase 3, after every card has landed, so a choice one of them
// makes can see the rest of the batch on the battlefield. No catalog
// AsEnters choice reads the other permanents entering with it today;
// the first one that does would err in whichever direction its choice
// leans, and should say so on its own card.
//
// A card a replacement cancels or redirects stays where it was — the
// same posture as the single-card move. Returns the IDs that entered,
// in the order given.
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

// putOntoBattlefieldFromZoneLocked is the shared body of the hand and
// library moves' synchronous doors: one simultaneous entry
// (entry_batch.go) of cards that are all in zones of kind `from`,
// reporting what entered on return.
//
// When a card's window asks a question (#1322) the batch waits for the
// answer, so this returns (nil, nil) and the cards land when it
// arrives. A caller with something to do with the entered permanents
// uses a Then door instead —
// PutCardsFromLibraryOntoBattlefieldThenForEffect,
// PutFromHandOntoBattlefieldThenForEffect or
// PutOntoBattlefieldTogetherThenForEffect.
//
// Caller must hold g.mu.
func (g *Game) putOntoBattlefieldFromZoneLocked(ids []uuid.UUID, from ZoneKind, opts ZoneEntryOptions) ([]uuid.UUID, error) {
	var entered []uuid.UUID
	err := g.startEntryBatchLocked(batchEntriesFrom(ids, from), opts, func(_ *Game, in []uuid.UUID) error {
		entered = in
		return nil
	})
	return entered, err
}

// batchEntriesFrom names every ID as put from one zone kind.
func batchEntriesFrom(ids []uuid.UUID, from ZoneKind) []BatchEntry {
	out := make([]BatchEntry, 0, len(ids))
	for _, id := range ids {
		out = append(out, BatchEntry{CardID: id, From: from})
	}
	return out
}

// PutCardsFromLibraryOntoBattlefieldThenForEffect is
// PutCardsFromLibraryOntoBattlefieldForEffect with the rest of the
// effect as a continuation: `then` is told the permanents that entered
// once the entry is complete — at once when no card's window asked a
// question, and from the last answer when one did (#1322). It runs
// exactly once, including when the batch is refused, which is what lets
// a caller put "the rest" of a library pile away without ever doing it
// while a card of the pile is still waiting to enter.
//
// Caller must hold g.mu.
func (g *Game) PutCardsFromLibraryOntoBattlefieldThenForEffect(ids []uuid.UUID, opts LibraryEntryOptions, then func(g *Game, entered []uuid.UUID) error) error {
	return g.startEntryBatchLocked(batchEntriesFrom(ids, ZoneLibrary), opts, then)
}

// PutFromHandOntoBattlefieldThenForEffect is the hand door with a
// continuation, told the entering permanent's ID (uuid.Nil when nothing
// entered) once the entry is complete. See
// PutCardsFromLibraryOntoBattlefieldThenForEffect for when it runs.
//
// Caller must hold g.mu.
func (g *Game) PutFromHandOntoBattlefieldThenForEffect(cardID uuid.UUID, opts HandEntryOptions, then func(g *Game, entered uuid.UUID) error) error {
	return g.startEntryBatchLocked([]BatchEntry{{CardID: cardID, From: ZoneHand}}, opts, func(g *Game, in []uuid.UUID) error {
		if then == nil {
			return nil
		}
		entered := uuid.Nil
		if len(in) > 0 {
			entered = in[0]
		}
		return then(g, entered)
	})
}
