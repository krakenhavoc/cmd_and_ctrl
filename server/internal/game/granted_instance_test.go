package game

import (
	"testing"

	"github.com/google/uuid"
)

// granted_instance_test.go is ADR 0066 decision 3 as amended on
// 2026-10-16 (#3938): a card that prints flashback and is granted
// flashback has two flashback abilities (CR 113.2c), and its owner casts
// it with either (CR 702.34a, CR 601.2b). The printed one is claimed as
// "flashback", the granted one as GrantedInstanceKey("flashback"), and
// the spell records "flashback" whichever paid.

// grantFlashbackAtManaCost is Snapcaster Mage's grant over one card in
// me's graveyard: flashback, its cost the card's mana cost.
func grantFlashbackAtManaCost(t *testing.T, g *Game, me *Player, id uuid.UUID) {
	t.Helper()
	c := graveyardCardByID(t, me, id)
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player: me.ID, Zone: ZoneGraveyard, AltCostKey: "flashback", ExileOnResolution: true,
		Label: "Flashback — its mana cost (Test Snapcaster)",
	}, []Card{c})
}

// graveyardOffers is CastOffersForLocked for a card in me's graveyard,
// under the permission a cast of it would use.
func graveyardOffers(t *testing.T, g *Game, me *Player, id uuid.UUID) []*AlternativeCost {
	t.Helper()
	c := graveyardCardByID(t, me, id)
	return g.CastOffersForLocked(me.ID, c, ZoneGraveyard, g.CastPermissionForLocked(me.ID, c, ZoneGraveyard))
}

func graveyardCardByID(t *testing.T, me *Player, id uuid.UUID) Card {
	t.Helper()
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatal("the card is not in the graveyard")
	return Card{}
}

func offerPrices(offers []*AlternativeCost) map[string]string {
	out := map[string]string{}
	for _, ac := range offers {
		if ac == nil {
			out[""] = "printed"
			continue
		}
		out[ac.Key] = ac.ManaCost
	}
	return out
}

func TestGrantedFlashbackBesideAPrintedOneIsASecondOffer(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	flashback := flashbackCost("{2}{R}")
	flashback.ExileOnLeavingStack = true
	withCatalogAlternativeCosts(t, altCostFor(oracle, flashback))
	id := looterInGraveyard(t, g, me, oracle)
	grantFlashbackAtManaCost(t, g, me, id)

	got := offerPrices(graveyardOffers(t, g, me, id))
	want := map[string]string{"flashback": "{2}{R}", GrantedInstanceKey("flashback"): "{R}"}
	if len(got) != len(want) || got["flashback"] != want["flashback"] || got[GrantedInstanceKey("flashback")] != "{R}" {
		t.Fatalf("offers = %v, want %v", got, want)
	}

	// The printed claim still means the printed flashback: one red is
	// not enough for it.
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"})
	var im *InsufficientManaError
	if !errorsAs(err, &im) {
		t.Fatalf("printed flashback on one red: got %v, want *InsufficientManaError", err)
	}

	// The granted one costs the card's mana cost, and the spell is a
	// flashback spell: it records "flashback" and is exiled as it leaves
	// the stack (CR 702.34a).
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: GrantedInstanceKey("flashback")}); err != nil {
		t.Fatalf("granted flashback on one red: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("the granted flashback did not charge {R}: pool %+v", me.ManaPool)
	}
	item := g.StackMeta[id]
	if item == nil || item.AltCost != "flashback" || !item.AltCostExiles {
		t.Fatalf("stack item = %+v, want AltCost flashback and the flashback exile", item)
	}
}

// Two flashback abilities at one price are one choice: the grant's is
// not listed beside a printed one it only repeats.
func TestGrantedFlashbackAtThePrintedPriceIsListedOnce(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, flashbackCost("{R}")))
	id := looterInGraveyard(t, g, me, oracle)
	grantFlashbackAtManaCost(t, g, me, id)

	got := offerPrices(graveyardOffers(t, g, me, id))
	if len(got) != 1 || got["flashback"] != "{R}" {
		t.Fatalf("offers = %v, want the printed flashback alone", got)
	}
}

// A granted flashback on a card that prints none is claimed under the
// bare key, as before; the instance spelling is refused, so each offer
// has one spelling.
func TestGrantedInstanceClaimNeedsAPrintedOffer(t *testing.T) {
	const oracle = "test-plain-sorcery"
	g := newActiveGame(t)
	me := g.Seats[0]
	id := looterInGraveyard(t, g, me, oracle)
	grantFlashbackAtManaCost(t, g, me, id)

	got := offerPrices(graveyardOffers(t, g, me, id))
	if len(got) != 1 || got["flashback"] != "{R}" {
		t.Fatalf("offers = %v, want the granted flashback under its bare key", got)
	}
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: GrantedInstanceKey("flashback")})
	if err != ErrInvalidParam {
		t.Fatalf("instance claim with no printed flashback: got %v, want ErrInvalidParam", err)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("granted flashback under its bare key: %v", err)
	}
}

func TestAltCostKeywordReadsBothSpellings(t *testing.T) {
	for claim, want := range map[string]string{
		"":                              "",
		"flashback":                     "flashback",
		GrantedInstanceKey("flashback"): "flashback",
		GrantedInstanceKey("escape"):    "escape",
		AltCostKeyGrantedCost:           AltCostKeyGrantedCost,
	} {
		if got := AltCostKeyword(claim); got != want {
			t.Errorf("AltCostKeyword(%q) = %q, want %q", claim, got, want)
		}
	}
}

// A granted retrace beside a printed one charges the same: the mana
// cost and a land card discarded. One choice, listed once.
func TestGrantedRetraceBesideAPrintedOneIsListedOnce(t *testing.T) {
	const oracle = "test-retrace"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
		Key: AltCostKeyRetrace, Label: "Retrace", ManaCost: "{R}", FromZone: ZoneGraveyard,
		DiscardFromHand: RetraceDiscardSpec(), PayLabel: "a land card",
	}))
	id := looterInGraveyard(t, g, me, oracle)
	land := NewCard("Mountain", me.ID)
	land.TypeLine = "Basic Land — Mountain"
	me.Hand.PushTop(land)
	c := graveyardCardByID(t, me, id)
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player: me.ID, Zone: ZoneGraveyard, AltCostKey: AltCostKeyRetrace, DiscardLandCard: true,
	}, []Card{c})

	got := offerPrices(graveyardOffers(t, g, me, id))
	if len(got) != 1 || got[AltCostKeyRetrace] != "{R}" {
		t.Fatalf("offers = %v, want the printed retrace alone", got)
	}
}
