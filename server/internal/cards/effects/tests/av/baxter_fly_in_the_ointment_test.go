package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// baxter_fly_in_the_ointment_test.go — Baxter, Fly in the Ointment
// (TMNT Commander, #5011), played through the real catalog entry.

const tmcB1BaxterOracle = "94551fbd-a72c-4a1f-be01-7ba1c7c03eed"

func TestTmcB1BaxterIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(tmcB1BaxterOracle)
	if !ok {
		t.Fatal("Baxter, Fly in the Ointment is not registered")
	}
	if spec.Name != "Baxter, Fly in the Ointment" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func tmcB1BaxterCard() game.Card {
	return game.Card{
		Name: "Baxter, Fly in the Ointment", TypeLine: "Legendary Creature — Insect Mutant Scientist",
		OracleID: tmcB1BaxterOracle, ManaCost: "{3}{U}", Power: 2, Toughness: 2, Colors: []string{"U"},
	}
}

// tmcB1CreatureWithCounters seeds a creature carrying `counters`.
func tmcB1CreatureWithCounters(g *game.Game, owner uuid.UUID, name string, counters map[string]int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner, Counters: counters,
	})
}

func tmcB1Flies(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return effectstest.EffectiveAbilitiesContain(t, g, id, "flying")
}

// Entering: my creatures with any counter fly; mine without one, and an
// opponent's with one, do not. Baxter himself has none and does not.
func TestTmcB1BaxterEntersGrantsFlyingToCreaturesWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plus := tmcB1CreatureWithCounters(g, me.ID, "Plus Bear", map[string]int{game.CounterPlusOne: 1})
	shield := tmcB1CreatureWithCounters(g, me.ID, "Shield Bear", map[string]int{game.CounterShield: 1})
	bare := tmcB1CreatureWithCounters(g, me.ID, "Bare Bear", nil)
	theirs := tmcB1CreatureWithCounters(g, opp.ID, "Their Bear", map[string]int{game.CounterPlusOne: 1})
	effectstest.AdvanceToMain(t, g)
	for _, c := range []string{"U", "C", "C", "C"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	baxter, err := effectstest.Msc17Cast(t, g, me, tmcB1BaxterCard(), game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("strict cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]bool{plus: true, shield: true, bare: false, theirs: false, baxter: false} {
		if got := tmcB1Flies(t, g, id); got != want {
			t.Errorf("%s flies %v, want %v", effectstest.TmtB1Card(t, g, id).Name, got, want)
		}
	}
	// The set was locked as the trigger resolved (CR 611.2c): a counter
	// Baxter gets from a draw afterwards does not make him fly.
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, baxter); n != 1 {
		t.Fatalf("%d counters on Baxter after a draw, want 1", n)
	}
	if tmcB1Flies(t, g, baxter) {
		t.Error("Baxter flies from a counter he got after the trigger resolved")
	}
}

func TestTmcB1BaxterStrictCastRefusedWithoutTheMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for _, c := range []string{"U", "C", "C"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	if _, err := effectstest.Msc17Cast(t, g, me, tmcB1BaxterCard(), game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a strict cast with three mana for {3}{U} was accepted")
	}
}

// Each card I draw puts a +1/+1 counter on him; an opponent's draw does
// not. Attacking with counters on him makes him fly, until end of turn.
func TestTmcB1BaxterGrowsOnDrawsAndFliesWhenHeAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baxter := effectstest.B27Push(g, me.ID, "Baxter, Fly in the Ointment", "Legendary Creature — Insect Mutant Scientist", tmcB1BaxterOracle, "{3}{U}", 2, 2, "U")
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.PlusOnes(g, baxter); n != 2 {
		t.Fatalf("%d counters after two draws, want 2", n)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("opponent draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, baxter); n != 2 {
		t.Fatalf("an opponent's draw grew him to %d counters", n)
	}
	if tmcB1Flies(t, g, baxter) {
		t.Fatal("he flies before attacking")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(baxter, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !tmcB1Flies(t, g, baxter) {
		t.Fatal("he does not fly after attacking with counters on him")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	if tmcB1Flies(t, g, baxter) {
		t.Error("flying outlasted the turn")
	}
}
