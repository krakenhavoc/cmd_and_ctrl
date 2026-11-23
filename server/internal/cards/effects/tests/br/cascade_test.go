package testsbr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCascadeIsWired(t *testing.T) {
	for _, tc := range []struct {
		oracle string
		want   int
	}{
		{effectstest.BloodbraidElfOracle, 1},
		{effectstest.ShardlessAgentOracle, 1},
		{effectstest.MaelstromWandererOracl, 2}, // "cascade, cascade"
		{effectstest.MaelstromNexusOracle, 1},
		{effectstest.ImotiOracle, 2}, // has cascade AND grants it
	} {
		if got := len(game.CatalogTriggers(tc.oracle)); got != tc.want {
			t.Errorf("%s: %d triggered abilities, want %d", tc.oracle, got, tc.want)
		}
	}
}

// Casting a cascader queues the trigger from the STACK, resolves it,
// and offers the hit.
func TestBloodbraidElfCascadesOnCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 3)

	effectstest.CastWithCost(t, g, "Bloodbraid Elf", "Creature — Elf Berserker", "{2}{R}{G}", effectstest.BloodbraidElfOracle)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("cascade offers = %d, want 1", got)
	}
	if n := effectstest.AnswerAllMayCast(t, g, me.ID, false); n != 1 {
		t.Errorf("answered %d offers, want 1", n)
	}
}

// "Cascade, cascade" is two abilities, so the Wanderer makes two
// separate offers.
func TestMaelstromWandererCascadesTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 6)

	effectstest.CastWithCost(t, g, "Maelstrom Wanderer", "Legendary Creature — Elemental", "{5}{G}{U}{R}", effectstest.MaelstromWandererOracl)
	effectstest.PassPriorityAroundTable(t, g)

	total := effectstest.AnswerAllMayCast(t, g, me.ID, false)
	// The two triggers resolve one at a time with a priority pass
	// between them, so drain, pass, drain again.
	effectstest.PassPriorityAroundTable(t, g)
	total += effectstest.AnswerAllMayCast(t, g, me.ID, false)
	if total != 2 {
		t.Errorf("Maelstrom Wanderer made %d cascade offers, want 2", total)
	}
}

// Maelstrom Nexus grants cascade to the FIRST spell of the turn and
// nothing after it.
func TestMaelstromNexusGrantsCascadeToTheFirstSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 6)
	effectstest.PushCatalogPermanent(g, me.ID, "Maelstrom Nexus", "Enchantment", effectstest.MaelstromNexusOracle, false)

	effectstest.CastWithCost(t, g, "First Spell", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("first spell: cascade offers = %d, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	// #730: the cascade offer gated the table while it was open, so
	// the spell that made it is still on the stack until it is
	// answered — and the second cast is sorcery speed.
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.CastWithCost(t, g, "Second Spell", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Errorf("second spell: cascade offers = %d, want 0", got)
	}
}

// Imoti grants cascade on mana value 6 or greater, measured on the
// PRINTED cost.
func TestImotiGrantsCascadeToSixDrops(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 6)
	effectstest.PushCatalogPermanent(g, me.ID, "Imoti, Celebrant of Bounty", "Legendary Creature — Snake Druid", effectstest.ImotiOracle, false)

	effectstest.CastWithCost(t, g, "Five Drop", "Sorcery", "{4}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Errorf("a five-drop got %d cascade offers, want 0", got)
	}

	effectstest.CastWithCost(t, g, "Six Drop", "Sorcery", "{5}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Errorf("a six-drop got %d cascade offers, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
}

// Taking the offer really does cast the card, for free and during the
// cascade trigger's resolution (CR 608.2g, #4243) — the end-to-end claim
// the keyword makes. The cast is strict with an empty pool: the only way
// it succeeds is if the permission really made it free.
func TestCascadeHitIsCastableForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 1)

	effectstest.CastWithCost(t, g, "Shardless Agent", "Artifact Creature — Human Rogue", "{1}{G}{U}", effectstest.ShardlessAgentOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 || !g.ResolutionPaused() {
		t.Fatalf("%d cascade offers, resolution paused %v: want one offer holding the trigger's resolution", got, g.ResolutionPaused())
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, true)
	if g.ResolutionPaused() {
		t.Errorf("the resolution is still held after the cast")
	}
	found := false
	for _, c := range g.Stack.Cards {
		if c.Name == "Cheap Spell" {
			found = true
		}
	}
	if !found {
		t.Fatalf("the cascaded Cheap Spell is not on the stack above the Shardless Agent")
	}
}
