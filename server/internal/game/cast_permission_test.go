package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// cast_permission_test.go — S42 / ADR 0066. The MODEL, against stubbed
// catalog hooks; the eight cards are pinned end to end in
// cards/effects/granted_permissions_test.go.
//
// What is worth a test here is exactly what the ADR argued about: the
// CR 400.7 identity check, the two scopes, the fact that a permission
// opens a zone without opening the sorcery-speed gate, the library
// top's position rule, and that the whole thing survives an undo.

// withCatalogCastPermissions stubs the standing-permission hook for
// one test.
func withCatalogCastPermissions(t *testing.T, fn func(oracleID string) []CastPermission) {
	t.Helper()
	prev := CatalogCastPermissions
	CatalogCastPermissions = fn
	t.Cleanup(func() { CatalogCastPermissions = prev })
}

// withCatalogLibraryTop stubs the CR 401.5 visibility hook.
func withCatalogLibraryTop(t *testing.T, fn func(oracleID string) LibraryTopVisibility) {
	t.Helper()
	prev := CatalogLibraryTopVisible
	CatalogLibraryTopVisible = fn
	t.Cleanup(func() { CatalogLibraryTopVisible = prev })
}

// grantOn reads the permission that opens `card` out of `zone` for
// `seat`, or nil.
func grantOn(g *Game, seat, card uuid.UUID, zone ZoneKind) *CastPermission {
	c, ok := g.LookupCardForEffect(card)
	if !ok {
		return nil
	}
	return g.CastPermissionForLocked(seat, c, zone)
}

// seedGraveyard puts a card in p's graveyard and returns its ID.
func seedGraveyard(p *Player, name, typeLine, manaCost string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// --- CR 400.7 -------------------------------------------------------

// The identity check, and the reason Card.ObjectEpoch exists: a
// permission names an OBJECT, so a card that leaves the zone and comes
// back is a new object with nothing granted to it — by any route, not
// just by being cast.
func TestGrantEndsWhenTheCardBecomesANewObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedGraveyard(me, "Granted Sorcery", "Sorcery", "{R}")
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(id, CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, AltCostKey: "flashback",
		})
	})
	if perm := grantOn(g, me.ID, id, ZoneGraveyard); !perm.Granted() {
		t.Fatalf("setup: no permission on the graveyard card")
	}

	// Out to exile and back: same instance, new object.
	g.WithWriteLock(func() {
		if _, err := MoveCard(me.Graveyard, g.Exile, id); err != nil {
			t.Fatalf("graveyard -> exile: %v", err)
		}
		if _, err := MoveCard(g.Exile, me.Graveyard, id); err != nil {
			t.Fatalf("exile -> graveyard: %v", err)
		}
	})
	if perm := grantOn(g, me.ID, id, ZoneGraveyard); perm.Granted() {
		t.Errorf("the permission survived two zone changes (CR 400.7): %+v", perm)
	}
}

// The scope check that makes Past in Flames different from Underworld
// Breach: a card named by a ScopeCards permission is covered, and one
// that was not is not — the set does not re-derive itself.
func TestScopeCardsNamesObjectsRatherThanAQuery(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	named := seedGraveyard(me, "Named", "Instant", "{U}")

	g.WithWriteLock(func() {
		c, _ := g.LookupCardForEffect(named)
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, AltCostKey: "flashback",
		}, []Card{c})
	})
	later := seedGraveyard(me, "Later", "Instant", "{U}")

	if perm := grantOn(g, me.ID, named, ZoneGraveyard); !perm.Granted() {
		t.Errorf("the named card is not covered")
	}
	if perm := grantOn(g, me.ID, later, ZoneGraveyard); perm.Granted() {
		t.Errorf("a card added after the lock is covered: %+v", perm)
	}
}

// The standing half: derived from the battlefield, so it covers
// whatever matches right now and stops the instant the source leaves.
func TestScopeStandingFollowsItsSource(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastPermissions(t, func(id string) []CastPermission {
		if id != "test-breach" {
			return nil
		}
		return []CastPermission{{
			Zone: ZoneGraveyard, Scope: ScopeStanding, WhileInZone: true,
			Filter: PermissionFilter{NonLandOnly: true}, AltCostKey: "escape",
		}}
	})
	source := permanentFor(g, me, "Test Breach", "Enchantment", "{1}{R}")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == source {
				g.Battlefield.Cards[i].OracleID = "test-breach"
			}
		}
	})

	spell := seedGraveyard(me, "Dead Spell", "Instant", "{U}")
	land := seedGraveyard(me, "Dead Land", "Basic Land — Island", "")
	if perm := grantOn(g, me.ID, spell, ZoneGraveyard); !perm.Granted() {
		t.Errorf("a nonland card in the graveyard has no escape")
	}
	if perm := grantOn(g, me.ID, land, ZoneGraveyard); perm.Granted() {
		t.Errorf("a land gained escape from a nonland-only filter")
	}

	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, source); err != nil {
			t.Fatalf("remove the source: %v", err)
		}
	})
	if perm := grantOn(g, me.ID, spell, ZoneGraveyard); perm.Granted() {
		t.Errorf("the standing permission outlived its source: %+v", perm)
	}
}

// --- timing ---------------------------------------------------------

// A grant opens a ZONE, not the stack: a sorcery in the graveyard is
// still a sorcery, and a permission that says nothing about timing
// changes nothing about it (ADR 0066 decision 6).
func TestGrantedCastStillObeysSorceryTiming(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	id := seedGraveyard(me, "Granted Sorcery", "Sorcery", "{R}")
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(id, CastPermission{Player: me.ID, Zone: ZoneGraveyard})
	})

	err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"})
	if !errors.Is(err, ErrSorcerySpeedRequired) {
		t.Fatalf("granted graveyard cast in the upkeep: got %v, want ErrSorcerySpeedRequired", err)
	}

	// The same permission with TimingFlash opens it, which is the
	// field madness (#657) will set.
	g.WithWriteLock(func() {
		me.CastPermissions = nil
		g.GrantCastPermissionOverCardForEffect(id, CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, Timing: TimingFlash,
		})
	})
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("granted cast under TimingFlash: %v", err)
	}
}

// A permission that PRICES its zone must be paid for, exactly as a
// printed zone-bound offer must (CR 118.9). Otherwise a Snapcaster'd
// card would be castable out of the graveyard for its printed cost —
// strictly better than the card.
func TestAGrantedPriceMustBeClaimed(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := seedGraveyard(me, "Granted Sorcery", "Sorcery", "{R}")
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(id, CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, AltCostKey: "flashback", ExileOnResolution: true,
		})
	})
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"}); !errors.Is(err, ErrCastCostRequired) {
		t.Fatalf("unclaimed granted cost: got %v, want ErrCastCostRequired", err)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("claimed granted cost: %v", err)
	}
	// CR 702.34a rides the STACK ITEM, because the catalog has never
	// heard of this card's flashback.
	item := g.StackMeta[id]
	if item == nil || !item.AltCostExiles {
		t.Fatalf("the granted flashback did not stamp its exile clause on the stack item: %+v", item)
	}
}

// A permission must never take away a path the card already prints.
// Gravecrawler's own text opens the graveyard for free; an Underworld
// Breach on the table ADDS escape and must not start charging for the
// free cast — a permission that made its controller strictly worse off
// would be wrong in both directions at once.
func TestAPermissionNeverRemovesThePrintedPath(t *testing.T) {
	const oracle = "test-gravecrawler"
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogCastPermissions(t, func(id string) []CastPermission {
		if id != "test-breach" {
			return nil
		}
		return []CastPermission{{
			Zone: ZoneGraveyard, Scope: ScopeStanding, WhileInZone: true,
			Filter: PermissionFilter{NonLandOnly: true}, AltCostKey: "escape",
		}}
	})
	source := permanentFor(g, me, "Test Breach", "Enchantment", "{1}{R}")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == source {
				g.Battlefield.Cards[i].OracleID = "test-breach"
			}
		}
	})
	crawler := seedGraveyard(me, "Test Gravecrawler", "Creature — Zombie", "{B}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == crawler {
				me.Graveyard.Cards[i].OracleID = oracle
			}
		}
	})

	// The printed path: no claim, no price.
	if err := g.CastSpell(me.ID, crawler, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("the card's own free graveyard cast was refused under a permission: %v", err)
	}
}

// --- the library top ------------------------------------------------

// CR 401.5: the permission opens the top card and nothing else, and
// only while the player may see it.
func TestLibraryTopPermissionNeedsTheTopAndTheLook(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	withCatalogCastPermissions(t, func(id string) []CastPermission {
		if id != "test-citadel" {
			return nil
		}
		return []CastPermission{{
			Zone: ZoneLibrary, Scope: ScopeStanding, WhileInZone: true, TopOfLibraryOnly: true,
		}}
	})
	source := permanentFor(g, me, "Test Citadel", "Artifact", "{3}")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == source {
				g.Battlefield.Cards[i].OracleID = "test-citadel"
			}
		}
	})
	buried := seedLibraryTop(me, "Buried", "Instant")
	top := seedLibraryTop(me, "On Top", "Instant")

	// No visibility declared yet: a card you cannot see is a card you
	// cannot play.
	withCatalogLibraryTop(t, func(string) LibraryTopVisibility { return LibraryTopHidden })
	if perm := grantOn(g, me.ID, top, ZoneLibrary); perm.Granted() {
		t.Errorf("the top card was playable with nobody able to look at it")
	}

	withCatalogLibraryTop(t, func(id string) LibraryTopVisibility {
		if id == "test-citadel" {
			return LibraryTopOwner
		}
		return LibraryTopHidden
	})
	if perm := grantOn(g, me.ID, top, ZoneLibrary); !perm.Granted() {
		t.Errorf("the top card is not playable with the look declared")
	}
	if perm := grantOn(g, me.ID, buried, ZoneLibrary); perm.Granted() {
		t.Errorf("a card below the top was playable: %+v", perm)
	}
	// And it is the OWNER's permission, not the table's.
	if perm := grantOn(g, g.Seats[1].ID, top, ZoneLibrary); perm.Granted() {
		t.Errorf("an opponent could play off someone else's library")
	}
}

// The visibility rule itself: per player, derived, and composing to
// the more permissive of two sources.
func TestLibraryTopVisibilityIsPerPlayerAndComposes(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withCatalogLibraryTop(t, func(id string) LibraryTopVisibility {
		switch id {
		case "test-look":
			return LibraryTopOwner
		case "test-reveal":
			return LibraryTopRevealed
		}
		return LibraryTopHidden
	})
	seedLibraryTop(me, "Mine", "Instant")

	if got := g.LibraryTopVisibilityForEffect(me.ID); got != LibraryTopHidden {
		t.Errorf("a player controlling nothing = %v, want hidden", got)
	}

	look := permanentFor(g, me, "Looker", "Artifact", "{1}")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == look {
				g.Battlefield.Cards[i].OracleID = "test-look"
			}
		}
	})
	if got := g.LibraryTopVisibilityForEffect(me.ID); got != LibraryTopOwner {
		t.Errorf("with a looker out = %v, want owner-only", got)
	}
	if got := g.LibraryTopVisibilityForEffect(them.ID); got != LibraryTopHidden {
		t.Errorf("the rule leaked to another seat: %v", got)
	}
	knowers, _ := g.LibraryTopKnowersLocked(me.ID)
	if len(knowers) != 1 || knowers[0] != me.ID {
		t.Errorf("owner-only knowers = %v, want just the owner", knowers)
	}

	// A revealed source alongside it wins — the two compose to the
	// more permissive.
	reveal := permanentFor(g, me, "Revealer", "Enchantment", "{2}")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == reveal {
				g.Battlefield.Cards[i].OracleID = "test-reveal"
			}
		}
	})
	if got := g.LibraryTopVisibilityForEffect(me.ID); got != LibraryTopRevealed {
		t.Errorf("look + reveal = %v, want revealed", got)
	}
	if knowers, _ := g.LibraryTopKnowersLocked(me.ID); len(knowers) != len(g.Seats) {
		t.Errorf("revealed knowers = %d seats, want all %d", len(knowers), len(g.Seats))
	}
}

// --- undo and snapshot ----------------------------------------------

// Permissions are data, classified `carried`. An undo that rewinds
// past a Snapcaster must take the flashback back; one that rewinds to
// a point where it was still owed must restore it intact.
func TestCastPermissionSurvivesCloneAndRestore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedGraveyard(me, "Granted Sorcery", "Sorcery", "{R}")
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(id, CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, AltCostKey: "flashback", UntilTurn: 7,
		})
	})

	clone := g.Clone()
	if perm := grantOn(clone, me.ID, id, ZoneGraveyard); !perm.Granted() || perm.AltCostKey != "flashback" {
		t.Fatalf("the clone lost the permission: %+v", perm)
	}
	// Independent backing arrays: mutating the clone must not reach
	// the original.
	clone.WithWriteLock(func() { clone.Seats[0].CastPermissions = nil })
	if perm := grantOn(g, me.ID, id, ZoneGraveyard); !perm.Granted() {
		t.Errorf("clearing the clone cleared the original")
	}

	// And through the persisted snapshot.
	snap := g.CaptureSnapshot()
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	perm := grantOn(restored, me.ID, id, ZoneGraveyard)
	if !perm.Granted() || perm.UntilTurn != 7 || perm.AltCostKey != "flashback" {
		t.Fatalf("the restored game lost the permission: %+v", perm)
	}
	// The epoch has to survive with it, or the restore would revive
	// every permission ever granted against this card.
	c, _ := restored.LookupCardForEffect(id)
	live, _ := g.LookupCardForEffect(id)
	if c.ObjectEpoch != live.ObjectEpoch {
		t.Errorf("restored epoch = %d, want %d", c.ObjectEpoch, live.ObjectEpoch)
	}
}

// The cleanup sweep is hygiene, not correctness — but it does have to
// drop what it says it drops, or the slice grows for the length of the
// game.
func TestCleanupSweepsSpentPermissions(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	thisTurn := seedGraveyard(me, "This Turn", "Instant", "{U}")
	nextTurn := seedGraveyard(me, "Next Turn", "Instant", "{U}")
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(thisTurn, CastPermission{Player: me.ID, Zone: ZoneGraveyard})
		g.GrantCastPermissionOverCardForEffect(nextTurn, CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, UntilTurn: g.Turn.Number + 1,
		})
		g.clearExpiredCastPermissionsLocked()
	})
	if len(me.CastPermissions) != 1 {
		t.Fatalf("permissions after the sweep = %d, want 1", len(me.CastPermissions))
	}
	if perm := grantOn(g, me.ID, nextTurn, ZoneGraveyard); !perm.Granted() {
		t.Errorf("the sweep reaped a permission with a later window")
	}
}
