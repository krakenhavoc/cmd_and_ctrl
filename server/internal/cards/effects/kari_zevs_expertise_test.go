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

	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no free-cast pick over the hand")
	}
	if hasID(pick.ChooseCards, expensive) {
		t.Error("a mana value 5 spell should not be offered for free")
	}
	if !hasID(pick.ChooseCards, cheapSpell) {
		t.Error("a mana value 1 spell should be offered for free")
	}
}

func TestKariZevsExpertiseFreeCastCanBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	cheapSpell := pushHandCardWithManaCost(me, "Cheap Spell", "Instant", "{1}")

	castCatalogSpell(t, g, "Kari Zev's Expertise", "Sorcery", kariZevsExpertiseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no free-cast pick over the hand")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the free cast: %v", err)
	}
	if !me.Hand.Contains(cheapSpell) {
		t.Error("declining the free cast must leave the card in hand")
	}
}
