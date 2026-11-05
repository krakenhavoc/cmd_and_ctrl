package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// april_oneil_live_on_the_scene_test.go — April O'Neil, Live on the
// Scene (TMNT Commander, #5011), played through the real catalog entry.

const tmcB1AprilOracle = "4900c157-8d9f-4f92-aaca-5246b6e2832e"

func TestTmcB1AprilIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(tmcB1AprilOracle)
	if !ok {
		t.Fatal("April O'Neil, Live on the Scene is not registered")
	}
	if spec.Name != "April O'Neil, Live on the Scene" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func tmcB1PushApril(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "April O'Neil, Live on the Scene", "Legendary Creature — Human Detective", tmcB1AprilOracle, "{1}{U}", 2, 1, "U")
}

// tmcB1CastCreature casts a vanilla {1} creature with `typeLine` from
// seat 0's hand strictly and resolves everything.
func tmcB1CastCreature(t *testing.T, g *game.Game, typeLine string) {
	t.Helper()
	me := g.Seats[0]
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if _, err := msc17Cast(t, g, me, game.Card{
		Name: "Test Creature", TypeLine: typeLine, ManaCost: "{1}", Power: 1, Toughness: 1,
	}, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("strict cast of %q: %v", typeLine, err)
	}
	b27Settle(t, g, me.ID, uuid.Nil)
}

func tmcB1Clues(g *game.Game, controller uuid.UUID) int {
	return tmtB1Count(g, controller, "Clue")
}

// One Clue for each of a Mutant, a Ninja and a Turtle; one Clue (not
// three) for a Mutant Ninja Turtle; and it works.
func TestTmcB1AprilInvestigatesForEachTypeOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tmcB1PushApril(g, me.ID)
	for i, typeLine := range []string{"Creature — Mutant", "Creature — Ninja", "Creature — Turtle", "Creature — Mutant Ninja Turtle"} {
		tmcB1CastCreature(t, g, typeLine)
		if got := tmcB1Clues(g, me.ID); got != i+1 {
			t.Fatalf("after %q: %d Clues, want %d", typeLine, got, i+1)
		}
	}
	clue := tmtB1Named(g, "Clue")[0]
	hand := me.Hand.Size()
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, clue, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack the Clue: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after cracking the Clue, want %d", got, hand+1)
	}
}

// A catalog Mutant cast for real: Crustacean Commando enters and April
// investigates beside its own Mutagen trigger.
func TestTmcB1AprilSeesACatalogMutant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tmcB1PushApril(g, me.ID)
	castCatalogSpell(t, g, "Crustacean Commando", "Creature — Crab Mutant Soldier", tmtMuCommandoOracle, nil)
	b27Settle(t, g, me.ID, uuid.Nil)
	if got := tmcB1Clues(g, me.ID); got != 1 {
		t.Errorf("%d Clues after a Mutant entered, want 1", got)
	}
}

func TestTmcB1AprilIgnoresOtherTypesAndItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if _, err := msc17Cast(t, g, me, game.Card{
		Name: "April O'Neil, Live on the Scene", TypeLine: "Legendary Creature — Human Detective",
		OracleID: tmcB1AprilOracle, ManaCost: "{1}{U}", Power: 2, Toughness: 1, Colors: []string{"U"},
	}, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("strict cast of April: %v", err)
	}
	passPriorityAroundTable(t, g)
	tmcB1CastCreature(t, g, "Creature — Human Rat")
	if got := tmcB1Clues(g, me.ID); got != 0 {
		t.Errorf("%d Clues from April and a Human Rat, want 0", got)
	}
}

// An opponent's April does not see my Turtle.
func TestTmcB1AprilIgnoresAnotherPlayersTurtle(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	tmcB1PushApril(g, opp.ID)
	tmcB1CastCreature(t, g, "Creature — Turtle")
	if got := tmcB1Clues(g, opp.ID); got != 0 {
		t.Errorf("the opponent's April made %d Clues from my Turtle, want 0", got)
	}
	if got := tmcB1Clues(g, g.Seats[0].ID); got != 0 {
		t.Errorf("I got %d Clues, want 0", got)
	}
}
