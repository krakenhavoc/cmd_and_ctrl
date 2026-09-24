package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_replacements_test.go — #1222 from the card side: the CR 614
// window on the mana PRODUCED, watched through the ordinary
// Spec.Replacements slot.
//
// Two printed cards, deliberately, so the kind is not shaped around
// one of them: Mana Reflection's ×2 and Nyxbloom Ancient's ×3 are the
// same sentence with different arithmetic, and the board with both on
// it is ×6.

const (
	manaReflectionOracle  = "a3a8a044-283d-443e-bc40-c2f826d70c22"
	nyxbloomAncientOracle = "8b610f8f-c8dd-4eeb-bc6e-3bc706d5f63e"
)

// pushLandFor seeds a basic land under `owner`; the synthetic
// CR 305.6 mana ability comes off the type line.
func pushLandFor(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner,
	})
	return id
}

func TestManaProductionReplacementsAreWired(t *testing.T) {
	for _, oracle := range []string{manaReflectionOracle, nyxbloomAncientOracle} {
		if n := len(game.CatalogReplacements(oracle)); n != 1 {
			t.Errorf("%s declared %d replacements, want 1", oracle, n)
		}
	}
}

// --- Mana Reflection -------------------------------------------------

// The headline: "if you tap a permanent for mana, it produces twice as
// much of that mana instead".
func TestManaReflectionDoublesYourLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Mana Reflection", "Enchantment", manaReflectionOracle, false)
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")

	tapForMana(t, g, me.ID, forest)
	if got := len(me.ManaPool); got != 2 {
		t.Errorf("a Mana-Reflected Forest made %d mana, want 2", got)
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "G" {
			t.Errorf("twice as much of THAT mana: got %q, want G", tok.Color)
		}
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("one replacement queued %d prompts, want none", len(g.PendingChoices))
	}
}

// "If YOU tap" is the whole scope clause — an opponent's land is
// untouched.
func TestManaReflectionDoesNotDoubleAnOpponentsLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Mana Reflection", "Enchantment", manaReflectionOracle, false)
	island := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")

	tapForMana(t, g, opp.ID, island)
	if got := len(opp.ManaPool); got != 1 {
		t.Errorf("the opponent's Island made %d mana, want 1", got)
	}
}

// A spell's "Add {B}{B}{B}" is not a tap for mana, so Dark Ritual is
// not doubled (CR 106.12a — the printed condition).
func TestManaReflectionDoesNotDoubleASpellsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushCatalogPermanent(g, me.ID, "Mana Reflection", "Enchantment", manaReflectionOracle, false)

	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(me.ID, src, "{B}{B}{B}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	})
	if got := len(me.ManaPool); got != 3 {
		t.Errorf("a spell's Add {B}{B}{B} made %d mana, want 3 (nothing was tapped)", got)
	}
}

// Two Mana Reflections are ×4 and nobody is asked to order them: two
// objects contributing ONE declared effect is #792's identical-window
// skip.
func TestTwoManaReflectionsQuadrupleWithNoPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Mana Reflection", "Enchantment", manaReflectionOracle, false)
	pushCatalogPermanent(g, me.ID, "Mana Reflection", "Enchantment", manaReflectionOracle, false)
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")

	tapForMana(t, g, me.ID, forest)
	if got := len(me.ManaPool); got != 4 {
		t.Errorf("two Mana Reflections: got %d mana, want 4", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("two copies of one declared effect queued %d prompts, want none (#792)", len(g.PendingChoices))
	}
}

// --- Nyxbloom Ancient ------------------------------------------------

func TestNyxbloomAncientTriplesYourLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Nyxbloom Ancient", "Enchantment Creature — Elemental", nyxbloomAncientOracle, false)
	mountain := pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")

	tapForMana(t, g, me.ID, mountain)
	if got := len(me.ManaPool); got != 3 {
		t.Errorf("a Nyxbloomed Mountain made %d mana, want 3", got)
	}
}

// Both out is ×6, and the CR 616.1 ordering is never put to anybody: a
// production cannot pause (CR 605.3b, ADR 0013 §5ab), and the answer is
// the same in either order.
func TestManaReflectionAndNyxbloomCompose(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Mana Reflection", "Enchantment", manaReflectionOracle, false)
	pushCatalogPermanent(g, me.ID, "Nyxbloom Ancient", "Enchantment Creature — Elemental", nyxbloomAncientOracle, false)
	swamp := pushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")

	tapForMana(t, g, me.ID, swamp)
	if got := len(me.ManaPool); got != 6 {
		t.Errorf("Mana Reflection + Nyxbloom Ancient on one Swamp: got %d mana, want 6", got)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			t.Fatalf("a mana production must never queue a CR 616 ordering prompt")
		}
	}
}

// The auto-tapper plans with the replaced amount: a cast of {G}{G}
// that needs two Forests without Mana Reflection needs one with it.
func TestManaReflectionIsPricedByTheAutoTapper(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")

	cost, err := game.ParseCost("{G}{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if _, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Fatalf("one Forest must not pay {G}{G} without a doubler")
	}
	pushCatalogPermanent(g, me.ID, "Mana Reflection", "Enchantment", manaReflectionOracle, false)
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok {
		t.Fatalf("a Mana-Reflected Forest pays {G}{G}")
	}
	if len(plan) != 1 || plan[0] != forest {
		t.Fatalf("plan: got %v, want just the Forest", plan)
	}
}
