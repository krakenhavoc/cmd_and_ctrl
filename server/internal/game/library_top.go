package game

import "github.com/google/uuid"

// library_top.go — S42, ADR 0066 decision 5: who can see the top card
// of a library (CR 401.5).
//
// A player may only play a card from the top of their library if they
// can see it, and the printed clauses come in exactly two strengths:
//
//   - "You may look at the top card of your library any time" —
//     private to its owner. Realmwalker, Bolas's Citadel, Mystic
//     Forge, Vizier of the Menagerie.
//   - "Play with the top card of your library revealed" — public.
//     Oracle of Mul Daya, Courser of Kruphix, Future Sight.
//
// Card.KnownBy cannot express either one. It is per CARD, and the top
// of a library is a POSITION: it changes identity on every draw,
// mill, shuffle, scry, search and cast, and keeping KnownBy correct
// would mean stamping the new top after every one of those. That is
// the wrong shape, and it is the shape #765 asked this ADR to avoid.
//
// So the rule is per PLAYER and lives on the position. It is derived
// from the battlefield — like a standing cast permission and for the
// same reason — and it is resolved LAZILY, at the only two places
// that ever need an answer:
//
//   - the view projection, which marks the current top card as known
//     to the players the rule names before the per-viewer redaction
//     runs;
//   - CastPermissionForLocked, which refuses a top-of-library
//     permission on a library nobody may look at.
//
// Lazy resolution is correct by construction, because the question is
// always about whatever is on top NOW. Nothing has to be invalidated,
// and CR 401.6 ("a top card that stops being revealed and is revealed
// again is a new object") falls out: the projection re-derives the
// answer every frame and never carries a stale one.

// LibraryTopVisibility is how far a player's top library card is
// visible. Ordered weakest to strongest so the derivation can take a
// maximum over the permanents a player controls — two sources compose
// to the more permissive one, and a Courser next to a Realmwalker
// reveals the top card to everybody.
type LibraryTopVisibility int

const (
	// LibraryTopHidden is the ordinary state of every library: the
	// top card is hidden information (CR 400.2) and nobody may look.
	// The zero value, so a player who controls nothing gets it.
	LibraryTopHidden LibraryTopVisibility = iota

	// LibraryTopOwner is "you may look at the top card of your
	// library any time" — its owner sees it, nobody else does.
	LibraryTopOwner

	// LibraryTopRevealed is "play with the top card of your library
	// revealed" — every player sees it.
	LibraryTopRevealed
)

// CatalogLibraryTopVisible returns how far a battlefield permanent
// with the given catalog key makes its controller's top library card
// visible. Populated at init from CardDef.LibraryTopVisible.
//
// A derived hook rather than a CR 613 layer, for the reason
// CatalogAdditionalLandPlays and CatalogNoMaxHandSize give: the layer
// engine models characteristics of objects, and "you may look at the
// top card of your library" is not one.
var CatalogLibraryTopVisible func(oracleID string) LibraryTopVisibility

// LibraryTopVisibilityLocked returns how far playerID's top library
// card is currently visible: the most permissive clause among the
// permanents they control.
//
// CatalogAbilityKey rather than CatalogKey — "play with the top card
// of your library revealed" is a static ability, and a Courser of
// Kruphix that has lost its abilities reveals nothing.
//
// Caller must hold g.mu (read or write).
func (g *Game) LibraryTopVisibilityLocked(playerID uuid.UUID) LibraryTopVisibility {
	if g.Battlefield == nil || CatalogLibraryTopVisible == nil || playerID == uuid.Nil {
		return LibraryTopHidden
	}
	out := LibraryTopHidden
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != playerID || c.OracleID == "" {
			continue
		}
		if v := CatalogLibraryTopVisible(CatalogAbilityKey(*c)); v > out {
			out = v
		}
	}
	return out
}

// LibraryTopVisibilityForEffect is LibraryTopVisibilityLocked for
// callers outside a locked frame — the view builder asks it while
// projecting a seat, and the client uses the answer to decide whether
// to offer the top-of-library affordance at all.
func (g *Game) LibraryTopVisibilityForEffect(playerID uuid.UUID) LibraryTopVisibility {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.LibraryTopVisibilityLocked(playerID)
}

// LibraryTopKnowersLocked returns the seats that may currently see
// playerID's top library card, and the card itself.
//
// (nil, uuid.Nil) when the library is empty or nobody may look. The
// owner is always first in the slice when they are in it at all, so a
// caller that only wants "may the owner see it" can read the answer
// off LibraryTopVisibilityLocked instead.
//
// Caller must hold g.mu (read or write).
func (g *Game) LibraryTopKnowersLocked(playerID uuid.UUID) ([]uuid.UUID, uuid.UUID) {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
		return nil, uuid.Nil
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	switch g.LibraryTopVisibilityLocked(playerID) {
	case LibraryTopOwner:
		return []uuid.UUID{playerID}, top
	case LibraryTopRevealed:
		out := make([]uuid.UUID, 0, len(g.Seats))
		out = append(out, playerID)
		for _, seat := range g.Seats {
			if seat != nil && seat.ID != playerID {
				out = append(out, seat.ID)
			}
		}
		return out, top
	}
	return nil, uuid.Nil
}
