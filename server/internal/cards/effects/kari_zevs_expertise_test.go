package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const kariZevsExpertiseOracle = "e0de0121-6185-45c5-9d88-a415b07347e0"

// pushHandCardWithManaCost seeds a card into a player's hand with a
// real mana value, for the free-cast pick's mana-value ceiling.
func pushHandCardWithManaCost(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// TestKariZevsExpertiseStealsUntapsHastesAndOffersAFreeCast exercises
// every printed clause: Act of Treason's three primitives, plus the
// free-cast offer over the hand, capped at mana value 2.
func TestKariZevsExpertiseStealsUntapsHastesAndOffersAFreeCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushTappedVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	cheapSpell := pushHandCardWithManaCost(me, "Cheap Spell", "Instant", "{1}")
	expensive := pushHandCardWithManaCost(me, "Expensive Spell", "Instant", "{5}")

	castCatalogSpell(t, g, "Kari Zev's Expertise", "Sorcery", kariZevsExpertiseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, victim); got != me.ID {
		t.Errorf("controller %s, want the caster %s", got, me.ID)
	}
	if c, _ := battlefieldCard(g, victim); c.Tapped {
		t.Error("Kari Zev's Expertise did not untap the creature it stole")
	}
	assertKeywords(t, g, victim, "haste")

	// #4250: the free cast is offered inside the spell's resolution.
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheapSpell)
	if hasID(offer.DiscardOptions, expensive) {
		t.Error("a mana value 5 spell should not be offered for free")
	}
	if !g.Stack.Contains(kariZevsExpertiseOnStack(g)) {
		t.Error("Kari Zev's Expertise left the stack before the free cast was answered")
	}
	if err := castThroughResolutionCast(t, g, me.ID, cheapSpell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	if !g.Stack.Contains(cheapSpell) {
		t.Fatal("the free spell is not on the stack")
	}
	castInsideTheResolution(t, g, cheapSpell)
}

// kariZevsExpertiseOnStack is the resolving Expertise's card, still in the
// stack zone while its resolution is held.
func kariZevsExpertiseOnStack(g *game.Game) uuid.UUID {
	for _, c := range g.Stack.Cards {
		if c.OracleID == kariZevsExpertiseOracle {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestKariZevsExpertiseFreeCastCanBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	cheapSpell := pushHandCardWithManaCost(me, "Cheap Spell", "Instant", "{1}")

	castCatalogSpell(t, g, "Kari Zev's Expertise", "Sorcery", kariZevsExpertiseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheapSpell)
	declineResolutionCast(t, g, me.ID)
	if !me.Hand.Contains(cheapSpell) {
		t.Error("declining the free cast must leave the card in hand")
	}
	if g.CastPermissionOnCardByIDForEffect(cheapSpell) != nil || g.ResolutionPaused() {
		t.Error("the decline left the grant, or held the resolution")
	}
}
