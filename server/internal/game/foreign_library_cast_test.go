package game

import (
	"errors"
	"testing"
)

// foreign_library_cast_test.go — #1035. #1022 made a cast permission
// reach a card in another seat's GRAVEYARD; the library deliberately
// got none of it, because it could not work: permissionPositionOKLocked
// enforced CR 401.5's "the top card of your library" against the
// PERMISSION HOLDER's own pile, so a permission over somebody else's
// library never matched the card and opened nothing — silently, and
// before any of the three surfaces could ask about it.
//
// The position rule is about the library the CARD is in. Xanathar,
// Guild Kingpin's "you may look at the top card of their library any
// time, you may play the top card of their library" is the printed
// shape, and it is one stored ScopeStanding permission that names a
// seat: ZoneOwner for the pile, SeesLibraryTop for CR 401.5's look.

// xanatharGrant gives `holder` Xanathar's clause over `victim`'s
// library for this turn.
func xanatharGrant(g *Game, holder, victim *Player) {
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(CastPermission{
			Player:           holder.ID,
			Zone:             ZoneLibrary,
			Scope:            ScopeStanding,
			ZoneOwner:        victim.ID,
			TopOfLibraryOnly: true,
			SeesLibraryTop:   true,
			AnyColor:         true,
			Label:            "Play the top card of their library (Xanathar, Guild Kingpin)",
		})
	})
}

// The permission is the key: with one, the top card of the other
// seat's library is castable and reaches the stack.
func TestCastingFromAnotherSeatsLibraryTopUnderAPermission(t *testing.T) {
	g := newActiveGame(t)
	victim, holder := g.Seats[0], g.Seats[1]
	withCatalogLibraryTop(t, func(string) LibraryTopVisibility { return LibraryTopHidden })
	seedLibraryTop(victim, "Buried", "Instant")
	top := seedLibraryTop(victim, "On Top", "Instant")
	xanatharGrant(g, holder, victim)

	if perm := grantOn(g, holder.ID, top, ZoneLibrary); !perm.Granted() {
		t.Fatalf("the grant does not open the top of the named seat's library")
	}
	if err := g.CastSpell(holder.ID, top, CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("CastSpell from the other seat's library: %v", err)
	}
	if victim.Library.Contains(top) {
		t.Errorf("the card is still in its owner's library")
	}
	if !g.Stack.Contains(top) {
		t.Fatalf("the card did not reach the stack")
	}
	for _, c := range g.Stack.Cards {
		// CR 608.1: the caster controls the spell; CR 400.3 leaves the
		// owner alone, so it goes to the victim's graveyard afterwards.
		if c.InstanceID == top && c.Owner != victim.ID {
			t.Errorf("owner = %s, want the library's owner", c.Owner)
		}
	}
}

// And ONLY that card. CR 401.5 opens a position, not a pile, so the
// card under the top is as unreachable as it ever was — and so is
// every card once the permission is gone.
func TestOnlyTheTopOfAnotherSeatsLibraryIsOpen(t *testing.T) {
	g := newActiveGame(t)
	victim, holder := g.Seats[0], g.Seats[1]
	withCatalogLibraryTop(t, func(string) LibraryTopVisibility { return LibraryTopHidden })
	buried := seedLibraryTop(victim, "Buried", "Instant")
	top := seedLibraryTop(victim, "On Top", "Instant")
	xanatharGrant(g, holder, victim)

	if perm := grantOn(g, holder.ID, buried, ZoneLibrary); perm.Granted() {
		t.Errorf("a card below the top was open: %+v", perm)
	}
	if err := g.CastSpell(holder.ID, buried, CastSpellParams{FromZone: "library"}); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("CastSpell of a buried card = %v, want ErrCardNotFound", err)
	}
	// The victim's own top card is not theirs to play either: the
	// grant is the holder's, and CR 401.5 needs one.
	if perm := grantOn(g, victim.ID, top, ZoneLibrary); perm.Granted() {
		t.Errorf("the library's owner got a permission they were never granted: %+v", perm)
	}
}

// The position check follows the CARD. This is the bug #1035 was filed
// for, in one assertion: the holder's OWN library has a different card
// on top, and reading that pile instead of the victim's is what made
// the permission open nothing.
func TestTheLibraryPositionCheckReadsTheCardsOwnLibrary(t *testing.T) {
	g := newActiveGame(t)
	victim, holder := g.Seats[0], g.Seats[1]
	withCatalogLibraryTop(t, func(string) LibraryTopVisibility { return LibraryTopHidden })
	theirTop := seedLibraryTop(victim, "Their Top", "Instant")
	myTop := seedLibraryTop(holder, "My Top", "Instant")
	xanatharGrant(g, holder, victim)

	if perm := grantOn(g, holder.ID, theirTop, ZoneLibrary); !perm.Granted() {
		t.Errorf("the position check read the holder's own library, not the card's")
	}
	// And it does not spill the other way: the grant names the
	// victim's pile, so the holder's own top card is not in it.
	if perm := grantOn(g, holder.ID, myTop, ZoneLibrary); perm.Granted() {
		t.Errorf("a grant over another seat's library opened the holder's own: %+v", perm)
	}
}

// A STANDING permission with no seat named is "your library", exactly
// as a Breach is "your graveyard" (#1022). Two Coursers on one table
// must not play lands off each other's revealed top card — the rule
// that was true by accident while the position check read the holder's
// own pile.
func TestAStandingLibraryPermissionStopsAtItsHoldersOwnLibrary(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	const oracle = "test-standing-courser"
	withCatalogCastPermissions(t, func(key string) []CastPermission {
		if key != oracle {
			return nil
		}
		return []CastPermission{{Zone: ZoneLibrary, Scope: ScopeStanding, TopOfLibraryOnly: true}}
	})
	// BOTH seats reveal their top card, so the visibility half cannot
	// be what refuses the cross-seat read.
	withCatalogLibraryTop(t, func(key string) LibraryTopVisibility {
		if key == oracle {
			return LibraryTopRevealed
		}
		return LibraryTopHidden
	})
	for _, seat := range []*Player{me, them} {
		source := permanentFor(g, seat, "Test Courser", "Creature", "{1}{G}{G}")
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == source {
					g.Battlefield.Cards[i].OracleID = oracle
				}
			}
		})
	}
	mine := seedLibraryTop(me, "My Top", "Instant")
	theirs := seedLibraryTop(them, "Their Top", "Instant")

	if perm := grantOn(g, me.ID, mine, ZoneLibrary); !perm.Granted() {
		t.Errorf("the Courser controller's own library top is not open")
	}
	if perm := grantOn(g, me.ID, theirs, ZoneLibrary); perm.Granted() {
		t.Errorf("a standing \"your library\" permission reached an opponent's library: %+v", perm)
	}
}

// CR 401.5's other half. A card you cannot see is a card you cannot
// play, and a cross-seat grant carries its own look because no
// permanent's static ability can declare one about somebody else's
// library. Without it the grant opens nothing.
func TestACrossSeatLibraryGrantWithoutTheLookOpensNothing(t *testing.T) {
	g := newActiveGame(t)
	victim, holder := g.Seats[0], g.Seats[1]
	withCatalogLibraryTop(t, func(string) LibraryTopVisibility { return LibraryTopHidden })
	top := seedLibraryTop(victim, "On Top", "Instant")

	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(CastPermission{
			Player: holder.ID, Zone: ZoneLibrary, Scope: ScopeStanding,
			ZoneOwner: victim.ID, TopOfLibraryOnly: true,
		})
	})
	if perm := grantOn(g, holder.ID, top, ZoneLibrary); perm.Granted() {
		t.Errorf("a grant with no look clause opened a library nobody may see into: %+v", perm)
	}
	if err := g.CastSpell(holder.ID, top, CastSpellParams{FromZone: "library"}); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("CastSpell without the look = %v, want ErrCardNotFound", err)
	}
}

// The look and the cast are one rule, so the seat that may cast the
// card is a seat the projection has already made a knower of it.
func TestTheGrantedLookReachesTheKnowerSet(t *testing.T) {
	g := newActiveGame(t)
	victim, holder := g.Seats[0], g.Seats[1]
	withCatalogLibraryTop(t, func(string) LibraryTopVisibility { return LibraryTopHidden })
	top := seedLibraryTop(victim, "On Top", "Instant")
	xanatharGrant(g, holder, victim)

	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.LibraryTopVisibleToLocked(victim.ID, holder.ID) {
		t.Errorf("the holder of a look grant cannot see the top card they may play")
	}
	if g.LibraryTopVisibleToLocked(victim.ID, victim.ID) {
		t.Errorf("Xanathar's clause is a look for its holder, not a reveal to the table")
	}
	knowers, gotTop := g.LibraryTopKnowersLocked(victim.ID)
	if gotTop != top {
		t.Fatalf("knower set names card %s, want the top card %s", gotTop, top)
	}
	if len(knowers) != 1 || knowers[0] != holder.ID {
		t.Errorf("knowers = %v, want just the grant holder %s", knowers, holder.ID)
	}
}

// A ScopeCards permission over a card in somebody else's pile survives
// the CR 400.7 sweep: the named object is looked for WHERE IT IS, not
// in the pile the holder owns (#1035, found closing #1022).
func TestAForeignScopeCardsPermissionSurvivesTheSweep(t *testing.T) {
	g := newActiveGame(t)
	owner, holder := g.Seats[0], g.Seats[1]
	id := seedGraveyard(owner, "Their Dead Spell", "Instant", "{1}{U}")

	g.mu.Lock()
	ok := g.GrantCastPermissionOverCardForEffect(id, CastPermission{
		Player: holder.ID, Zone: ZoneGraveyard, Scope: ScopeCards,
		Duration: IndefiniteDuration(),
	})
	g.mu.Unlock()
	if !ok {
		t.Fatalf("GrantCastPermissionOverCardForEffect: card not found")
	}

	g.mu.Lock()
	g.sweepCastPermissionsLocked(true)
	g.mu.Unlock()

	if perm := grantOn(g, holder.ID, id, ZoneGraveyard); !perm.Granted() {
		t.Errorf("the cleanup sweep dropped a live permission over another seat's graveyard card")
	}

}
