package game

import "github.com/google/uuid"

// manifest_dread.go — CR 701.62a, manifest dread (ADR 0082's 2026-10-07
// amendment, #2570).
//
//	"To manifest dread, look at the top two cards of your library. Put
//	 one of those cards onto the battlefield face down as a 2/2 creature
//	 and the other into your graveyard."   — CR 701.62a
//
// Manifest (ManifestForEffect) always takes the top card; the new thing
// here is the CHOICE. It is composed from doors that already exist and
// owns nothing but the order they run in:
//
//   - LookAtTopOfLibraryForEffect — "look at", not reveal: only the
//     looker becomes a knower, exactly as scry, surveil and hideaway do.
//   - A choose_cards prompt over the looked-at pair, answered by the
//     looker alone and hidden from every other seat (the hideaway
//     prompt, so the enumerator and every bot already answer it).
//   - The CR 614 face-down entry, the one manifest uses.
//   - A plain zone move for the other card. It is not a mill (CR 701.17a
//     counts cards off the top) and not a discard, so no mill payoff
//     sees a manifest dread.
//
// # What the continuation is told
//
// Every manifest-dread card that does anything afterwards says "that
// creature" (Experimental Lab, the four Equipment) or "a card you put
// into your graveyard this way" (Paranormal Analyst), so the finished
// action reports both: which permanent entered and which card went to
// the graveyard. Either is uuid.Nil when it did not happen — a library
// of one card has nothing to put into a graveyard, and a replacement
// that sends the manifested card somewhere else leaves nothing to
// report. The continuation runs exactly once, after the last leg
// settles, including after a leg that paused on a CR 614 question or a
// CR 903.9 prompt.
//
// # Order of the two moves
//
// The manifested card enters first and the other card goes to the
// graveyard from the entry's continuation. The rules text puts them in
// one sentence, and nothing observable separates the orders: no card
// reads the graveyard between them, and the entry's own replacement
// window may ask a question that the graveyard card must not jump.
//
// # The event
//
// EventManifestDread is emitted once, after both moves, for "whenever
// you manifest dread". It is not emitted when there was nothing to
// look at (an empty library does nothing, as with every other look).

// ManifestDreadResult is what a finished manifest dread did.
type ManifestDreadResult struct {
	// Player manifested dread.
	Player uuid.UUID
	// Manifested is the permanent that entered face down, or uuid.Nil.
	Manifested uuid.UUID
	// Graveyarded is the card put into Player's graveyard "this way",
	// or uuid.Nil.
	Graveyarded uuid.UUID
}

// ManifestDreadThenForEffect carries out CR 701.62a for `playerID` and
// hands `then` what happened. `source` is the permanent or spell asking
// and names the prompt's source; it may be uuid.Nil.
//
// The prompt is skipped when the library shows one card (nothing to
// choose) and the whole action is skipped, with `then` still told an
// empty result, when it shows none: a "then put counters on that
// creature" continuation must run so it can find it has no creature.
//
// `then` may be nil. It must capture only scalars — an undo across the
// prompt replays it against a restored game — and use the *Game it is
// handed.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ManifestDreadThenForEffect(playerID, source uuid.UUID, then func(g *Game, res ManifestDreadResult) error) error {
	if g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	done := func(g *Game, res ManifestDreadResult) error {
		if then == nil {
			return nil
		}
		return then(g, res)
	}
	looked := g.LookAtTopOfLibraryForEffect(playerID, 2)
	if len(looked) == 0 {
		return done(g, ManifestDreadResult{Player: playerID})
	}
	finish := func(g *Game, picked []uuid.UUID) error {
		return g.finishManifestDreadLocked(playerID, source, looked, picked, done)
	}
	if len(looked) == 1 {
		return finish(g, looked)
	}
	g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:  playerID,
		Source:   source,
		Question: "Manifest dread — put one onto the battlefield face down; the other goes to your graveyard",
		Cards:    looked,
		Min:      1,
		Max:      1,
		Zone:     ZoneLibrary,
		Then:     finish,
	})
	return nil
}

// finishManifestDreadLocked runs the two moves once the card is chosen.
// An answer that names nothing falls back to the top card, the pick
// manifest has always made, so a continuation that was handed an empty
// list cannot leave the pair in the library.
//
// Caller must hold g.mu in write mode.
func (g *Game) finishManifestDreadLocked(playerID, source uuid.UUID, looked, picked []uuid.UUID, done func(*Game, ManifestDreadResult) error) error {
	chosen := looked[0]
	if len(picked) > 0 {
		chosen = picked[0]
	}
	var rest []uuid.UUID
	for _, id := range g.cardsInALibraryLocked(looked) {
		if id != chosen {
			rest = append(rest, id)
		}
	}
	finishAll := func(g *Game, manifested uuid.UUID) error {
		return g.routeAllThenLocked(zoneRoute{Dst: ZoneGraveyard, Actor: playerID, Source: source}, rest,
			func(g *Game, landed []uuid.UUID) error {
				res := ManifestDreadResult{Player: playerID, Manifested: manifested}
				if len(landed) > 0 {
					res.Graveyarded = landed[0]
				}
				g.EmitEvent(Event{
					Kind:   EventManifestDread,
					Actor:  playerID,
					Source: source,
					CardID: res.Manifested,
					Target: res.Graveyarded,
				})
				return done(g, res)
			})
	}
	if len(g.cardsInALibraryLocked([]uuid.UUID{chosen})) == 0 {
		// The chosen card left the library between the look and the
		// answer; nothing enters.
		return finishAll(g, uuid.Nil)
	}
	return g.PutCardsFromLibraryOntoBattlefieldThenForEffect([]uuid.UUID{chosen},
		ZoneEntryOptions{Controller: playerID, FaceDown: FaceDownManifested},
		func(g *Game, entered []uuid.UUID) error {
			manifested := uuid.Nil
			if len(entered) > 0 {
				manifested = entered[0]
			}
			return finishAll(g, manifested)
		})
}
