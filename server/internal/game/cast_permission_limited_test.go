package game

import (
	"testing"

	"github.com/google/uuid"
)

// cast_permission_limited_test.go — #1729: a permission good for one
// spell (CastPermission.CastsLeft), and the caster's choice between two
// permissions over the same card (CastPermissionForClaimLocked). The
// two cards that print them, Court of Locthwain and Locke, Treasure
// Hunter, are pinned end to end in cards/effects.

func seedExile(g *Game, owner uuid.UUID, name, typeLine, manaCost string) Card {
	c := NewCard(name, owner)
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	g.Exile.PushTop(c)
	return c
}

func liveExiled(t *testing.T, g *Game, id uuid.UUID) Card {
	t.Helper()
	c, ok := g.cardInZoneLocked(g.Exile, id)
	if !ok {
		t.Fatalf("card %v is not in exile", id)
	}
	return c
}

// "You may cast a spell from among those cards": one permission over a
// set, spent by the first cast. The second card is not castable by it.
func TestALimitedGrantOpensOneSpellFromItsSet(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	advanceTo(t, g, StepPrecombatMain)
	a := seedGraveyard(them, "Their Instant", "Instant", "{U}")
	b := seedGraveyard(them, "Their Sorcery", "Sorcery", "{R}")
	g.WithWriteLock(func() {
		var cards []Card
		for _, id := range []uuid.UUID{a, b} {
			c, _ := g.LookupCardForEffect(id)
			cards = append(cards, c)
		}
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, CastOnly: true, CastsLeft: 1, Label: "Test Locke",
		}, cards)
	})
	if err := g.CastSpell(me.ID, a, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("the first spell from the set: %v", err)
	}
	if len(me.CastPermissions) != 0 {
		t.Errorf("the one-spell permission survived its cast: %+v", me.CastPermissions)
	}
	if perm := grantOn(g, me.ID, b, ZoneGraveyard); perm.Granted() {
		t.Errorf("a second spell from the set is still open: %+v", perm)
	}
}

// A count of two is spent one cast at a time.
func TestALimitedGrantCountsDown(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	a := seedExile(g, me.ID, "First", "Instant", "{U}")
	b := seedExile(g, me.ID, "Second", "Instant", "{U}")
	c := seedExile(g, me.ID, "Third", "Instant", "{U}")
	g.WithWriteLock(func() {
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player: me.ID, Zone: ZoneExile, CastOnly: true, CastsLeft: 2, Label: "Two of three",
		}, []Card{a, b, c})
	})
	if err := g.CastSpell(me.ID, a.InstanceID, CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("first: %v", err)
	}
	if got := me.CastPermissions; len(got) != 1 || got[0].CastsLeft != 1 {
		t.Fatalf("after one cast the permission is %+v, want one use left", got)
	}
	if err := g.CastSpell(me.ID, b.InstanceID, CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("second: %v", err)
	}
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a third spell was cast under a permission good for two")
	}
}

// Locke's ruling (2025-06-06): "If you cast a spell from among the
// milled cards using another permission, Locke's effect doesn't apply."
// A card whose own text opens its graveyard is cast by that text, and
// the limited permission is still there afterwards.
func TestALimitedGrantIsNotSpentByTheCardsOwnPermission(t *testing.T) {
	const oracle = "test-gravecrawler"
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	crawler := seedGraveyard(me, "Test Gravecrawler", "Creature — Zombie", "{B}")
	other := seedGraveyard(me, "Other", "Instant", "{U}")
	g.WithWriteLock(func() {
		var cards []Card
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == crawler {
				me.Graveyard.Cards[i].OracleID = oracle
			}
		}
		for _, id := range []uuid.UUID{crawler, other} {
			c, _ := g.LookupCardForEffect(id)
			cards = append(cards, c)
		}
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, CastOnly: true, CastsLeft: 1, Label: "Test Locke",
		}, cards)
	})
	if err := g.CastSpell(me.ID, crawler, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("the card's own graveyard cast: %v", err)
	}
	if perm := grantOn(g, me.ID, other, ZoneGraveyard); !perm.Granted() {
		t.Error("a cast the card's own text allowed spent the one-spell permission")
	}
}

// Two permissions over one exiled card, as Court of Locthwain makes
// them: "you may play that card for as long as it remains exiled", and
// on a monarch turn "you may cast a spell from among cards exiled with
// this enchantment without paying its mana cost". The caster picks by
// claiming the free one's offer; a cast that does not claim it pays and
// leaves the free cast for another card.
func TestAClaimChoosesBetweenTwoPermissions(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	a := seedExile(g, me.ID, "Paid", "Instant", "{3}{U}")
	b := seedExile(g, me.ID, "Free", "Instant", "{5}{R}")
	g.WithWriteLock(func() {
		for _, c := range []Card{a, b} {
			g.GrantCastPermissionToCardsForEffect(CastPermission{
				Player: me.ID, Zone: ZoneExile, AnyColor: true, AnyType: true,
				Duration: WhileInZoneDuration(), Label: "Test Court — play it",
			}, []Card{c})
		}
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player: me.ID, Zone: ZoneExile, CastOnly: true, CastsLeft: 1,
			AltCostKey: "test_court", Cost: "{0}", Label: "Test Court — free",
		}, []Card{a, b})
	})

	var offers []*AlternativeCost
	g.WithWriteLock(func() {
		c := liveExiled(t, g, b.InstanceID)
		offers = g.CastOffersForLocked(me.ID, c, ZoneExile, g.CastPermissionForLocked(me.ID, c, ZoneExile))
	})
	claimable := false
	for _, o := range offers {
		if o != nil && o.Key == "test_court" {
			claimable = true
		}
	}
	if !claimable {
		t.Fatalf("the free permission's offer is not listed: %+v", offers)
	}

	// The paid cast first: no claim, so the play permission is used and
	// the free one is untouched.
	if err := g.CastSpell(me.ID, a.InstanceID, CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("paid cast under the play permission: %v", err)
	}
	free := 0
	for _, p := range me.CastPermissions {
		if p.AltCostKey == "test_court" {
			free++
		}
	}
	if free != 1 {
		t.Fatalf("the paid cast spent the free permission (%d left)", free)
	}

	// Then the free one, claimed.
	var price CastPrice
	g.WithWriteLock(func() {
		var err error
		price, err = g.priceCastLocked(me.ID, liveExiled(t, g, b.InstanceID), CastSpellParams{FromZone: "exile", AlternativeCost: "test_court"})
		if err != nil {
			t.Fatalf("price: %v", err)
		}
	})
	if !price.Total.Empty() {
		t.Errorf("the claimed free cast is priced %+v, want nothing", price.Total)
	}
	if err := g.CastSpell(me.ID, b.InstanceID, CastSpellParams{FromZone: "exile", AlternativeCost: "test_court"}); err != nil {
		t.Fatalf("free cast under the claimed permission: %v", err)
	}
	for _, p := range me.CastPermissions {
		if p.AltCostKey == "test_court" {
			t.Errorf("the free permission survived its one cast: %+v", p)
		}
	}
}

// The count is plain data: it survives an undo and a snapshot.
func TestALimitedGrantSurvivesASnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a := seedExile(g, me.ID, "Held", "Instant", "{U}")
	g.WithWriteLock(func() {
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player: me.ID, Zone: ZoneExile, CastOnly: true, CastsLeft: 1,
		}, []Card{a})
	})
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	p := restored.playerByIDLocked(me.ID)
	if p == nil || len(p.CastPermissions) != 1 || p.CastPermissions[0].CastsLeft != 1 {
		t.Fatalf("restored permissions %+v, want one with one cast left", p.CastPermissions)
	}
}
