package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const searOracle = "23b11fdc-61b5-4f5e-ad80-f81b729fc91d"

func TestSearIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(searOracle)
	if !ok || spec.Name != "Sear" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Sear kills a 4-toughness creature and takes four loyalty off a
// planeswalker; a player is not a legal target.
func TestSearDealsFourToACreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMainForCost(t, g)
	ogre := pushVanillaCreature(g, opp.ID, "Ogre", 4, 4)

	sear := rfdHandCard(me, "Sear", "Instant", "{1}{R}", searOracle, 0, 0)
	fillPoolColored(me, "R", 2)
	if err := g.CastSpell(me.ID, sear, game.CastSpellParams{Strict: true, Targets: pr6Player(opp.ID)}); err == nil {
		t.Fatal("Sear targeted a player")
	}
	if err := g.CastSpell(me.ID, sear, game.CastSpellParams{Strict: true, Targets: cardTarget(ogre)}); err != nil {
		t.Fatalf("cast at a creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the 4/4 survived 4 damage")
	}

	walker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 6}, Owner: opp.ID, Controller: opp.ID})
	sear2 := rfdHandCard(me, "Sear", "Instant", "{1}{R}", searOracle, 0, 0)
	fillPoolColored(me, "R", 2)
	if err := g.CastSpell(me.ID, sear2, game.CastSpellParams{Strict: true, Targets: cardTarget(walker)}); err != nil {
		t.Fatalf("cast at a planeswalker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, walker); !ok || c.Counters[game.CounterLoyalty] != 2 {
		got := c.Counters[game.CounterLoyalty]
		t.Errorf("planeswalker loyalty %d, want 2", got)
	}
}
