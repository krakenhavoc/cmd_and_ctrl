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
//     permission on a library its holder may not look at.
//
// #1035 added the one clause the per-library model cannot hold: "you
// may look at the top card of THEIR library" is per (library, viewer)
// rather than per library, it is granted by a resolution rather than
// printed as a static ability, and so it rides the cast permission
// that comes with it. LibraryTopVisibleToLocked is where the two
// spellings meet, and every caller asks that rather than the enum.
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

// LibraryTopVisibleToLocked reports whether `viewer` may see the top
// card of `owner`'s library right now — THE predicate behind CR
// 401.5's "a player can only play a card from the top of a library
// they can see".
//
// Two spellings, and it is the one place that knows both (#1035):
//
//   - the per-LIBRARY strength above, derived from the permanents its
//     owner controls. "You may look at the top card of your library"
//     is that owner alone; "play with the top card of your library
//     revealed" is everybody.
//   - a CROSS-SEAT look carried by a cast permission —
//     CastPermission.SeesLibraryTop, Xanathar, Guild Kingpin's "you
//     may look at the top card of their library any time". It cannot
//     ride the derivation: the clause is granted by a resolution to
//     one chosen player rather than printed as a static ability, so
//     no permanent's catalog entry can express it, and the strength
//     enum cannot either — it is per library, and this one is per
//     pair.
//
// Both the engine's position check (permissionPositionOKLocked) and
// the view's knower stamp (LibraryTopKnowersLocked) read it, so the
// card the holder may cast is exactly the card the holder can see.
//
// Caller must hold g.mu (read or write).
func (g *Game) LibraryTopVisibleToLocked(owner, viewer uuid.UUID) bool {
	if owner == uuid.Nil || viewer == uuid.Nil {
		return false
	}
	return g.libraryTopVisibleToLocked(g.LibraryTopVisibilityLocked(owner), owner, viewer)
}

// libraryTopVisibleToLocked is LibraryTopVisibleToLocked with the
// per-library strength already derived, so a caller asking about every
// seat in turn walks the battlefield once rather than once per seat.
func (g *Game) libraryTopVisibleToLocked(vis LibraryTopVisibility, owner, viewer uuid.UUID) bool {
	switch vis {
	case LibraryTopRevealed:
		return true
	case LibraryTopOwner:
		if viewer == owner {
			return true
		}
	}
	return g.libraryTopLookGrantedLocked(owner, viewer)
}

// libraryTopLookGrantedLocked reports whether `viewer` holds a live
// cast permission that carries CR 401.5's look over `owner`'s library
// (CastPermission.SeesLibraryTop).
//
// A scan of ONE seat's stored permissions, which is empty in almost
// every game — and it reads the permission slice directly rather than
// going through CastPermissionForLocked, because that function asks
// this one. The recursion stops here: liveness is
// CastPermissionActiveForEffect, which reads a duration and nothing
// else.
//
// Caller must hold g.mu.
func (g *Game) libraryTopLookGrantedLocked(owner, viewer uuid.UUID) bool {
	p := g.playerByIDLocked(viewer)
	if p == nil {
		return false
	}
	for i := range p.CastPermissions {
		perm := &p.CastPermissions[i]
		if !perm.SeesLibraryTop || perm.Zone != ZoneLibrary {
			continue
		}
		if perm.PileOwnerFor(viewer) != owner {
			continue
		}
		if !g.CastPermissionActiveForEffect(perm, viewer) {
			continue
		}
		return true
	}
	return false
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
// LibraryTopVisibleToLocked per seat rather than a second reading of
// the strength enum (#1035): the list the view stamps and the
// predicate the cast path checks are then the same rule, so a seat
// that may cast the top card of somebody else's library is a seat the
// projection has already made a knower of it.
//
// Caller must hold g.mu (read or write).
func (g *Game) LibraryTopKnowersLocked(playerID uuid.UUID) ([]uuid.UUID, uuid.UUID) {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
		return nil, uuid.Nil
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	vis := g.LibraryTopVisibilityLocked(playerID)
	out := make([]uuid.UUID, 0, len(g.Seats))
	if g.libraryTopVisibleToLocked(vis, playerID, playerID) {
		out = append(out, playerID)
	}
	for _, seat := range g.Seats {
		if seat == nil || seat.ID == playerID {
			continue
		}
		if g.libraryTopVisibleToLocked(vis, playerID, seat.ID) {
			out = append(out, seat.ID)
		}
	}
	if len(out) == 0 {
		return nil, uuid.Nil
	}
	return out, top
}
