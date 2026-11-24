package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b12_slith_test.go — S82 GAP2 batch 12: Etched Slith, through the
// real combat-damage trigger, the reflexive trigger's target and the
// counter choice.

const gap2B12Slith = "f1bd0d32-7513-4d1d-94aa-d06101a940c0"

func TestGap2B12EtchedSlithIsFull(t *testing.T) {
	spec, ok := effects.Lookup(gap2B12Slith)
	if !ok {
		t.Fatal("Etched Slith is not registered")
	}
	if spec.Name != "Etched Slith" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func gap2B12PushSlith(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushTestPermanent(g, owner, game.Card{
		Name: "Etched Slith", TypeLine: "Artifact Creature — Phyrexian Slith", OracleID: gap2B12Slith, Power: 1, Toughness: 1,
	})
}

func TestGap2B12EtchedSlithRemovesAPlayersCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slith := gap2B12PushSlith(g, me.ID)
	if err := g.AddPlayerCounter(opp.ID, game.CounterPoison, 3); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayerCounter(opp.ID, game.CounterEnergy, 2); err != nil {
		t.Fatal(err)
	}
	effectstest.DealCombatDamageToPlayer(g, slith, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(slith); c.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("the Slith has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no target prompt for the reflexive trigger: %+v", g.PendingChoices)
	}
	if effectstest.HasID(prompt.PickTargetCards, slith) {
		t.Error("the Slith itself was offered: it says another")
	}
	if effectstest.HasID(prompt.PickTargetPlayers, me.ID) {
		t.Error("its controller was offered: it says opponent")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("targeting the opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Kinds are asked in a fixed order: energy, then poison.
	if p := effectstest.AnswerNumber(t, g, me.ID, 0); p.Max != 1 {
		t.Errorf("the energy question allows %d, want 1", p.Max)
	}
	effectstest.AnswerNumber(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Counters[game.CounterPoison] != 2 || opp.Counters[game.CounterEnergy] != 2 {
		t.Errorf("opponent has %d poison and %d energy, want 2 and 2", opp.Counters[game.CounterPoison], opp.Counters[game.CounterEnergy])
	}
}

func TestGap2B12EtchedSlithRemovesAPermanentsCounterOrNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slith := gap2B12PushSlith(g, me.ID)
	walker := effectstest.PushTestPermanent(g, opp.ID, game.Card{
		Name: "Their Walker", TypeLine: "Planeswalker — Test", Counters: map[string]int{game.CounterLoyalty: 4},
	})
	effectstest.DealCombatDamageToPlayer(g, slith, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, walker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(walker); c.Counters[game.CounterLoyalty] != 3 {
		t.Errorf("the planeswalker has %d loyalty, want 3", c.Counters[game.CounterLoyalty])
	}

	// A second hit: declining removes nothing.
	effectstest.DealCombatDamageToPlayer(g, slith, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, walker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(walker); c.Counters[game.CounterLoyalty] != 3 {
		t.Errorf("declining took a counter: loyalty %d, want 3", c.Counters[game.CounterLoyalty])
	}
	if c, _ := g.LookupCardForEffect(slith); c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("the Slith has %d +1/+1 counters after two hits, want 2", c.Counters[game.CounterPlusOne])
	}
}
