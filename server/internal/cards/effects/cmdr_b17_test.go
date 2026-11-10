package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b17_test.go — Commander deck cards, batch 17 (#5368).

const (
	cmdrB17ShalaiOracle = "a0c47ab6-dfb4-46ee-a3f7-9e1521b4bb4b"
	cmdrB17DroverOracle = "b0d60856-80f9-4c64-9768-28a0a690ecb1"
)

func TestCmdrB17ShalaiGivesHexproofToYouWalkersAndOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shalai := pushCatalogPermanent(g, me.ID, "Shalai, Voice of Plenty", "Legendary Creature — Angel", cmdrB17ShalaiOracle, false)
	mine := pushVanillaCreature(g, me.ID, "My Guy", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Guy", 2, 2)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID,
	})
	if !hasPlayerAbility(playerAbilities(g, me), KeywordHexproof) {
		t.Error("you have hexproof")
	}
	if hasPlayerAbility(playerAbilities(g, opp), KeywordHexproof) {
		t.Error("opponents don't")
	}
	for _, c := range []struct {
		id   uuid.UUID
		name string
		want bool
	}{{mine, "other creature", true}, {walker, "planeswalker", true}, {theirs, "their creature", false}, {shalai, "Shalai herself", false}} {
		if got := eotHasAbility(effectiveAbilities(t, g, c.id), "hexproof"); got != c.want {
			t.Errorf("%s hexproof = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestCmdrB17ShalaiPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shalai := pushCatalogPermanent(g, me.ID, "Shalai, Voice of Plenty", "Legendary Creature — Angel", cmdrB17ShalaiOracle, false)
	mine := pushVanillaCreature(g, me.ID, "My Guy", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Guy", 2, 2)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "G"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, shalai, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{shalai: 1, mine: 1, theirs: 0} {
		if got := counterCount(g, id, "+1/+1"); got != want {
			t.Errorf("%s: %d counters, want %d", id, got, want)
		}
	}
}

func TestCmdrB17TwilightDroverGrowsWhenAnyCreatureTokenLeavesAndMakesSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drover := pushCatalogPermanent(g, me.ID, "Twilight Drover", "Creature — Spirit", cmdrB17DroverOracle, false)
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	nontoken := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	if err := g.DestroyPermanentForEffect(nontoken); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, drover, "+1/+1"); got != 0 {
		t.Fatalf("a nontoken creature leaving gave %d counters", got)
	}
	if err := g.DestroyPermanentForEffect(tok); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, drover, "+1/+1"); got != 1 {
		t.Fatalf("a creature token leaving gave %d counters, want 1", got)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, drover, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Spirit"); got != 2 {
		t.Fatalf("want two Spirit tokens, got %d", got)
	}
	// The counter was the cost, and the tokens leaving later feed it again.
	if got := counterCount(g, drover, "+1/+1"); got != 0 {
		t.Errorf("counter should be spent, have %d", got)
	}
}
