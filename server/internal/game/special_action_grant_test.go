package game

import "testing"

// special_action_grant_test.go — #1391, a special action granted to
// another card (ADR 0062 amendment 2026-09-24). The card itself,
// Fblthp, Lost on the Range, is tested end to end against the real
// catalog in cards/effects. What is pinned here, with stubbed hooks,
// are the two engine rules that a single legendary card cannot show:
// two grants make one offer, and the price of a granted offer is asked
// about the zone the card is really in.

const (
	grantSourceOracle = "test-grant-source"
	grantedCardOracle = "test-granted-card"
)

// withSpecialActionGrants stubs the grant hook for one test.
func withSpecialActionGrants(t *testing.T, fn func(oracleID string) []SpecialActionGrant) {
	t.Helper()
	prev := CatalogSpecialActionGrants
	CatalogSpecialActionGrants = fn
	t.Cleanup(func() { CatalogSpecialActionGrants = prev })
}

// withFblthpStyleGrant makes grantSourceOracle a Fblthp-shaped
// permanent that is not legendary, so a test can control two.
func withFblthpStyleGrant(t *testing.T) {
	t.Helper()
	withSpecialActionGrants(t, func(id string) []SpecialActionGrant {
		if id != grantSourceOracle {
			return nil
		}
		return []SpecialActionGrant{{Kind: SpecialActionPlot, Zone: ZoneLibrary, Nonland: true, CostIsManaCost: true}}
	})
}

// seedGrantedTop puts a card on top of p's library and returns it.
func seedGrantedTop(p *Player, name, typeLine, manaCost string) Card {
	c := NewCard(name, p.ID)
	c.OracleID = grantedCardOracle
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	p.Library.PushTop(c)
	return c
}

// Two grant sources make one widening and one mana-cost offer, not
// two identical rows the player would have to tell apart.
func TestTwoGrantsMakeOneOffer(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withFblthpStyleGrant(t)
	modifierSource(t, g, me, "Grant Source A", grantSourceOracle)
	modifierSource(t, g, me, "Grant Source B", grantSourceOracle)
	top := seedGrantedTop(me, "Bears", "Creature — Bear", "{1}{G}")

	var offers []SpecialAction
	g.ReadSnapshot(func() { offers = g.SpecialActionsOfferedLocked(me.ID, top, ZoneLibrary) })
	if len(offers) != 1 || offers[0].Cost != "{1}{G}" {
		t.Fatalf("offers under two grants: got %+v, want one Plot {1}{G}", offers)
	}
}

// The cost query of a granted offer names the zone the card is in, so
// a special-action cost modifier that cares where the action is taken
// from sees the library, not the hand.
func TestAGrantedOfferIsPricedFromItsOwnZone(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	withFblthpStyleGrant(t)
	withCatalogCostModifiers(t, modifiersFor("test-library-discount", CostModifier{
		Kind:           CostReduction,
		SpecialActions: true,
		Label:          "Special actions from your library cost {1} less.",
		AppliesTo:      func(q CostQuery) bool { return q.FromZone == ZoneLibrary },
		Amount:         fixed(1),
	}))
	modifierSource(t, g, me, "Grant Source", grantSourceOracle)
	modifierSource(t, g, me, "Library Discount", "test-library-discount")
	top := seedGrantedTop(me, "Bears", "Creature — Bear", "{1}{G}")

	me.ManaPool.AddMana(ManaToken{Color: "G"})
	if err := g.PerformSpecialAction(me.ID, top.InstanceID, SpecialActionPlot, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("plot from the library for {G} after the discount: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool: %d tokens left, want 0", len(me.ManaPool))
	}
}
