package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestKariZevsExpertiseStealsUntapsHastesAndOffersAFreeCast exercises
// every printed clause: Act of Treason's three primitives, plus the
// free-cast offer over the hand, capped at mana value 2.
func TestKariZevsExpertiseStealsUntapsHastesAndOffersAFreeCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushTappedVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	cheapSpell := effectstest.PushHandCardWithManaCost(me, "Cheap Spell", "Instant", "{1}")
	expensive := effectstest.PushHandCardWithManaCost(me, "Expensive Spell", "Instant", "{5}")

	effectstest.CastCatalogSpell(t, g, "Kari Zev's Expertise", "Sorcery", effectstest.KariZevsExpertiseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Errorf("controller %s, want the caster %s", got, me.ID)
	}
	if c, _ := effectstest.BattlefieldCard(g, victim); c.Tapped {
		t.Error("Kari Zev's Expertise did not untap the creature it stole")
	}
	effectstest.AssertKeywords(t, g, victim, "haste")

	// #4250: the free cast is offered inside the spell's resolution.
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheapSpell)
	if effectstest.HasID(offer.DiscardOptions, expensive) {
		t.Error("a mana value 5 spell should not be offered for free")
	}
	if !g.Stack.Contains(kariZevsExpertiseOnStack(g)) {
		t.Error("Kari Zev's Expertise left the stack before the free cast was answered")
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, cheapSpell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	if !g.Stack.Contains(cheapSpell) {
		t.Fatal("the free spell is not on the stack")
	}
	effectstest.CastInsideTheResolution(t, g, cheapSpell)
}

// kariZevsExpertiseOnStack is the resolving Expertise's card, still in the
// stack zone while its resolution is held.
func kariZevsExpertiseOnStack(g *game.Game) uuid.UUID {
	for _, c := range g.Stack.Cards {
		if c.OracleID == effectstest.KariZevsExpertiseOracle {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestKariZevsExpertiseFreeCastCanBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	cheapSpell := effectstest.PushHandCardWithManaCost(me, "Cheap Spell", "Instant", "{1}")

	effectstest.CastCatalogSpell(t, g, "Kari Zev's Expertise", "Sorcery", effectstest.KariZevsExpertiseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheapSpell)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if !me.Hand.Contains(cheapSpell) {
		t.Error("declining the free cast must leave the card in hand")
	}
	if g.CastPermissionOnCardByIDForEffect(cheapSpell) != nil || g.ResolutionPaused() {
		t.Error("the decline left the grant, or held the resolution")
	}
}
