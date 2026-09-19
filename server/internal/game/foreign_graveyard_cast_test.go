package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// foreign_graveyard_cast_test.go — #1022. ADR 0066 makes a ScopeCards
// CastPermission a statement about an OBJECT ("you may cast that
// card"), and nothing in it says the object has to sit in the holder's
// own zone — Wrexial's "you may cast target instant or sorcery card
// from that player's graveyard" is the printed shape.
//
// castSourceZoneLocked resolved `from_zone: "graveyard"` to the
// caster's own pile and nothing else, so the permission was answerable
// by CastPermissionForLocked, invisible to the view (the bug #1022 was
// filed for) and unreachable by the cast path, which returned
// ErrCardNotFound. These are the two halves of the rule that replaced
// that scoping: a permission is the key, and the ONLY key.

// deadSpellIn drops a castable instant into a player's graveyard.
func deadSpellIn(p *Player, name string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{1}{U}"
	c.KnownBy = map[uuid.UUID]bool{p.ID: true}
	id := c.InstanceID
	p.Graveyard.PushTop(c)
	return id
}

// The permission is the key: with one, the cast comes out of the other
// seat's graveyard and lands on the stack under the caster's control.
func TestCastingFromAnotherSeatsGraveyardUnderAPermission(t *testing.T) {
	g := newActiveGame(t)
	owner, caster := g.Seats[0], g.Seats[1]
	id := deadSpellIn(owner, "Their Dead Spell")

	g.mu.Lock()
	ok := g.GrantCastPermissionOverCardForEffect(id, CastPermission{
		Player: caster.ID, Zone: ZoneGraveyard, Scope: ScopeCards,
	})
	g.mu.Unlock()
	if !ok {
		t.Fatalf("GrantCastPermissionOverCardForEffect: card not found")
	}

	if err := g.CastSpell(caster.ID, id, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("CastSpell from the other seat's graveyard: %v", err)
	}
	if owner.Graveyard.Contains(id) {
		t.Errorf("the card is still in its owner's graveyard")
	}
	if !g.Stack.Contains(id) {
		t.Fatalf("the card did not reach the stack")
	}
	for _, c := range g.Stack.Cards {
		if c.InstanceID != id {
			continue
		}
		// CR 608.1: the caster controls the spell; CR 400.3: the owner
		// is unchanged, so it goes back to the owner's graveyard.
		if c.Owner != owner.ID {
			t.Errorf("owner = %s, want the graveyard's owner", c.Owner)
		}
	}
}

// And the only key: without one, the card in somebody else's graveyard
// is as unreachable as it was before #1022 — including a card whose own
// text opens the graveyard, because flashback, escape and Gravecrawler
// all print "your graveyard".
func TestCastingFromAnotherSeatsGraveyardWithoutAPermission(t *testing.T) {
	g := newActiveGame(t)
	owner, caster := g.Seats[0], g.Seats[1]
	id := deadSpellIn(owner, "Their Dead Spell")

	err := g.CastSpell(caster.ID, id, CastSpellParams{FromZone: "graveyard"})
	if !errors.Is(err, ErrCardNotFound) {
		t.Fatalf("CastSpell without a permission = %v, want ErrCardNotFound", err)
	}
	if !owner.Graveyard.Contains(id) {
		t.Errorf("a refused cast moved the card anyway")
	}
}

// A STANDING permission is a permanent's printed text, and every one of
// them says "your graveyard" — the filter has no ownership clause. The
// scoping used to be implicit in every caller looking only at its own
// pile; now that the lookup above can reach another seat's, it is
// written down in CastPermissionForLocked.
func TestAStandingGraveyardPermissionStopsAtItsHoldersOwnYard(t *testing.T) {
	g := newActiveGame(t)
	owner, holder := g.Seats[0], g.Seats[1]
	const oracle = "test-standing-breach-scope"

	prev := CatalogCastPermissions
	CatalogCastPermissions = func(key string) []CastPermission {
		if key == oracle {
			return []CastPermission{{
				Zone: ZoneGraveyard, Scope: ScopeStanding,
				Filter: PermissionFilter{NonLandOnly: true}, AltCostKey: "escape",
			}}
		}
		return nil
	}
	t.Cleanup(func() { CatalogCastPermissions = prev })

	breach := NewCard("Test Breach", holder.ID)
	breach.TypeLine = "Enchantment"
	breach.OracleID = oracle
	breach.Controller = holder.ID
	g.Battlefield.PushTop(breach)

	theirs := deadSpellIn(owner, "Their Dead Spell")
	mine := deadSpellIn(holder, "My Dead Spell")

	g.mu.Lock()
	defer g.mu.Unlock()
	var theirCard, myCard Card
	for _, c := range owner.Graveyard.Cards {
		if c.InstanceID == theirs {
			theirCard = c
		}
	}
	for _, c := range holder.Graveyard.Cards {
		if c.InstanceID == mine {
			myCard = c
		}
	}
	if g.CastPermissionForLocked(holder.ID, myCard, ZoneGraveyard) == nil {
		t.Errorf("the Breach controller's own graveyard card is not covered")
	}
	if perm := g.CastPermissionForLocked(holder.ID, theirCard, ZoneGraveyard); perm != nil {
		t.Errorf("a standing \"your graveyard\" permission reached an opponent's graveyard: %+v", perm)
	}
}
