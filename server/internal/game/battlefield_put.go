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

	// PutByEnteringPlayer records each card as put onto the battlefield
	// by the player it enters under, not by the controller of the
	// resolving spell or ability (#1997, last_turn_actions.go): the put
	// of an instruction that names who puts ("each player puts", "they
	// put it onto the battlefield", CR 110.2a). Set for a manifest or a
	// cloak, and for a batch begun inside PutByEnteringPlayerForEffect.
	PutByEnteringPlayer bool

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

	// EntersWithCounters are counters the putting effect says the
	// permanent enters with (Dodecapod's discard replacement: "put it
	// onto the battlefield with two +1/+1 counters on it", ADR 0153).
	// SEEDED onto every card's entry event, as Tapped is, so the
	// counters are the entry's own and the entry's CR 614 window and
	// the counter placement windows after it see them. Read-only: the
	// batch copies it onto each event. nil for every other caller.
	EntersWithCounters map[string]int

	// EntersWithCountersIfType are the putting effect's per-card
	// clauses: "each of them enters with an additional +1/+1 counter
	// on it if it's a creature and an additional loyalty counter on it
	// if it's a planeswalker" (Semester's End, #4193). Each member of
	// the batch is judged by what it would be on the battlefield
	// (CR 614.12) once its own CR 614 window has settled, and gets
	// only the clauses it matches; see entry_counters_if_type.go.
	// Read-only, nil for every other caller.
	EntersWithCountersIfType []EntryCountersIfType

	// AttachedTo is the host the putting effect names — "return … to
	// the battlefield attached to enchanted creature" (Mantle of the
	// Ancients, #3158). Seeded onto every card's entry event: each card
	// enters attached to it, and an Aura is not asked CR 303.4f's
	// question. An Aura that can't legally enchant it stays where it was
	// (CR 303.4i); an Equipment is attached and left to CR 704.5n. The
	// zero value is every other put, whose Auras choose their hosts as
	// they enter (aura_host.go).
	AttachedTo TargetRef
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

// PutFromCommandZoneOntoBattlefieldForEffect puts a card from a
// command zone onto the battlefield without casting it — commander
// ninjutsu's "Put this card onto the battlefield from your hand or the
// command zone tapped and attacking" (CR 702.49c, #1278).
//
// It is the hand door with the command zone for the hand: one body
// (putOntoBattlefieldFromZoneLocked), the controller stamped in the
// source zone, the CR 614 window before the card leaves, then
// EventZoneMove, EventETB and the AsEnters hook. Before it, nothing in
// the engine moved a commander OUT of the command zone by any route
// but casting it, so what the door has to say is mostly what a
// non-cast exit does NOT do:
//
//   - No commander tax, and nothing written towards the next one.
//     CR 903.8 taxes CASTING a commander from the command zone; the
//     tally (Player.CommanderCasts) is bumped by the cast path alone,
//     and this is not a cast (CR 702.49c puts the card, and CR 601
//     never starts). The card keeps its instance ID for the same
//     reason — the tally is keyed by it, and a new ID would forget
//     every earlier cast.
//   - No CR 903.9 prompt. That replacement is about a commander being
//     put into a library, hand, graveyard or exile; an arrival on the
//     battlefield is none of the four, and commanderZoneReplacement's
//     AppliesTo says so on the destination. It is still the replacement
//     window that runs — the ordinary CR 614 entry window, so an
//     "enters tapped" or a Clone-style question on the card is asked —
//     and the commander designation (Card.IsCommander) rides the card
//     untouched, so when the permanent LEAVES the battlefield later the
//     shared exit primitive (routeCardToZoneLocked, zone_route.go)
//     offers CR 903.9 exactly as it would for a cast one.
//   - Colour identity is not consulted. CR 903.4 is a deck-construction
//     and mana rule; it says nothing about which card may leave the
//     command zone.
//
// Returns the entering permanent's ID, or uuid.Nil with a nil error
// when a replacement canceled, redirected or paused the entry. Refuses
// a card not in a command zone with ErrCardNotFound, and a
// nonpermanent card (CR 110.4) with ErrInvalidParam, moving nothing.
//
// Caller must hold g.mu.
func (g *Game) PutFromCommandZoneOntoBattlefieldForEffect(cardID uuid.UUID, opts ZoneEntryOptions) (uuid.UUID, error) {
	entered, err := g.putOntoBattlefieldFromZoneLocked([]uuid.UUID{cardID}, ZoneCommand, opts)
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
// PutCardsFromLibraryOntoBattlefieldKeyedForEffect. See entry_batch.go.
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
// This is the library door of manifest (CR 701.40a), one card from the
// top. It is the primitive ADR 0069's object model was proved on — the
// 2/2 projection, the catalog suppression, the CR 708.5 viewers rule
// and the CR 708.9 reveal — and manifest, manifest dread (ManifestDread,
// manifest_dread.go) and cloak build on it. The turn-face-up special
// action (CR 116.2g) and the face-down CAST (CR 708.4) live in
// face_down.go. Manifesting from a HAND is ManifestFromHandsKeyedForEffect.
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

// ManifestFromHandsKeyedForEffect is "<player> manifests <cards> from
// their hand" (CR 701.40a; Kozilek, the Broken Reality, Scroll of Fate,
// ADR 0189 decision 3): every card in `ids` is turned face down and put
// onto the battlefield as a 2/2 creature under its OWNER's control, and
// `then` (a registered tail, #4253 slice 6) is told the permanents that
// entered as TailOutcome.IDs once the entry is complete. (The closure form,
// ManifestFromHandsThenForEffect, was deleted in #4253's cleanup.)
//
// CR 701.40e's "one at a time" covers only a LIBRARY. Cards named from a
// hand are chosen up front and enter together, as Selvala's Stampede's do:
// one PutCardsFromHandOntoBattlefieldKeyedForEffect batch with FaceDown
// set, so every CR 614 window sees the board as it stood before any card
// arrived. Controller is left unset, so each card enters under its owner;
// the ids may belong to different players' hands (Kozilek's two targets).
//
// Every ID must be in its owner's hand, or the whole call is refused with
// ErrCardNotFound and nothing moves (the batch's own validation). A
// nonpermanent card is legal: the object that arrives is a 2/2 creature
// whatever the card says (CR 701.40a, CR 708.2), and CR 701.40g keeps an
// instant or sorcery face down when it is turned up.
//
// Caller must hold g.mu.
func (g *Game) ManifestFromHandsKeyedForEffect(ids []uuid.UUID, then KeyedTail) error {
	return g.PutCardsFromHandOntoBattlefieldKeyedForEffect(ids, HandEntryOptions{FaceDown: FaceDownManifested}, then)
}

// putOntoBattlefieldFromZoneLocked is the shared body of the hand and
// library moves' synchronous doors: one simultaneous entry
// (entry_batch.go) of cards that are all in zones of kind `from`,
// reporting what entered on return.
//
// When a card's window asks a question (#1322) the batch waits for the
// answer, so this returns (nil, nil) and the cards land when it
// arrives. A caller with something to do with the entered permanents
// uses a keyed door instead —
// PutCardsFromLibraryOntoBattlefieldKeyedForEffect,
// PutFromHandOntoBattlefieldKeyedForEffect or
// PutOntoBattlefieldTogetherKeyedForEffect.
//
// Caller must hold g.mu.
func (g *Game) putOntoBattlefieldFromZoneLocked(ids []uuid.UUID, from ZoneKind, opts ZoneEntryOptions) ([]uuid.UUID, error) {
	// #4253 slice 3: what entered is read off the finished batch rather
	// than handed to a closure over a local, so a batch paused on a
	// question (a put shockland's "pay 2 life?") holds no closure and is
	// a restore point.
	b := &entryBatch{opts: opts}
	err := g.startPreparedEntryBatchLocked(batchEntriesFrom(ids, from), b)
	if !b.done {
		return nil, err
	}
	return b.landed, err
}

// batchEntriesFrom names every ID as put from one zone kind.
func batchEntriesFrom(ids []uuid.UUID, from ZoneKind) []BatchEntry {
	out := make([]BatchEntry, 0, len(ids))
	for _, id := range ids {
		out = append(out, BatchEntry{CardID: id, From: from})
	}
	return out
}

// PutCardsFromLibraryOntoBattlefieldKeyedForEffect is
// PutCardsFromLibraryOntoBattlefieldForEffect with the rest of the
// effect as a continuation, as data (#4253 slice 3): `then` is a
// registered tail (RegisterTail), told the permanents that entered as
// TailOutcome.IDs once the entry is complete — at once when no card's
// window asked a question, and from the last answer when one did
// (#1322). It runs exactly once, including when the batch is refused,
// which is what lets a caller put "the rest" of a library pile away
// without ever doing it while a card of the pile is still waiting to
// enter. A card's question asked on the way (a shockland's) is a
// restore point. (The closure form,
// PutCardsFromLibraryOntoBattlefieldThenForEffect, was deleted in
// #4361 S10 once only tests called it.)
//
// Caller must hold g.mu.
func (g *Game) PutCardsFromLibraryOntoBattlefieldKeyedForEffect(ids []uuid.UUID, opts LibraryEntryOptions, then KeyedTail) error {
	return g.startPreparedEntryBatchLocked(batchEntriesFrom(ids, ZoneLibrary), &entryBatch{keyed: then, opts: opts})
}

// PutOntoBattlefieldTogetherKeyedForEffect puts every named card onto the
// battlefield as ONE simultaneous entry, from whichever zones they are
// in, and then runs `then`, a registered tail (#4253 slice 3), told the
// permanents that entered as TailOutcome.IDs — Sword of
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
// Every question is a restore point: the continuation is data. (The
// closure form, PutOntoBattlefieldTogetherThenForEffect, was deleted in
// #4253's last part once only tests called it.)
//
// Caller must hold g.mu.
func (g *Game) PutOntoBattlefieldTogetherKeyedForEffect(cards []BatchEntry, opts ZoneEntryOptions, then KeyedTail) error {
	return g.startPreparedEntryBatchLocked(cards, &entryBatch{keyed: then, opts: opts})
}

// PutCardsFromHandOntoBattlefieldKeyedForEffect puts every card in `ids`
// from its owner's hand onto the battlefield as ONE simultaneous entry
// ("put up to N permanent cards from your hand onto the battlefield",
// Selvala's Stampede; ADR 0146 amendment 2026-10-13), and tells `then`,
// a registered tail (#4253 slice 6), which permanents entered once the
// entry is complete, as TailOutcome.IDs. The hand twin of
// PutCardsFromLibraryOntoBattlefieldKeyedForEffect, on the same batch
// (entry_batch.go): every CR 614 window is evaluated against the board
// as it stood before any card entered, every ETB fires after all have
// landed, and a card whose window stops to ask a question waits for the
// answer, which is a restore point. `then` runs exactly once, including
// when the batch is refused. (The closure form,
// PutCardsFromHandOntoBattlefieldThenForEffect, was deleted in #4361
// S10 once only tests called it.)
//
// Caller must hold g.mu.
func (g *Game) PutCardsFromHandOntoBattlefieldKeyedForEffect(ids []uuid.UUID, opts HandEntryOptions, then KeyedTail) error {
	return g.startPreparedEntryBatchLocked(batchEntriesFrom(ids, ZoneHand), &entryBatch{keyed: then, opts: opts})
}

// PutFromHandOntoBattlefieldKeyedForEffect is the single-card hand door
// with the continuation as data (#4253 slice 6): `then` is told the
// entering permanent as TailOutcome.IDs, which is empty when nothing
// entered.
//
// Caller must hold g.mu.
func (g *Game) PutFromHandOntoBattlefieldKeyedForEffect(cardID uuid.UUID, opts HandEntryOptions, then KeyedTail) error {
	return g.startPreparedEntryBatchLocked([]BatchEntry{{CardID: cardID, From: ZoneHand}}, &entryBatch{keyed: then, opts: opts})
}
