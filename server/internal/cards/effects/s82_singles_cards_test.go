package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_singles_cards_test.go — three S82 singles: Giant Badger (8ED,
// #5607), Kemba's Skyguard (DDF, #5578) and Hellkite Overlord (DRB,
// #5589).

const (
	s82sBadgerOracle   = "a4109138-5d69-4510-a228-7c1f88c0c8c0"
	s82sSkyguardOracle = "4c8167af-2eb0-4f0a-a1b9-b61b65bac261"
	s82sOverlordOracle = "1d05e441-84d4-482f-95a8-7842e776cad2"
)

func TestS82SinglesAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82sBadgerOracle:   "Giant Badger",
		s82sSkyguardOracle: "Kemba's Skyguard",
		s82sOverlordOracle: "Hellkite Overlord",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Giant Badger --------------------------------------------------------------

// s82sBadgerTable puts a Giant Badger under the second seat and a 3/3
// attacker under the first, and walks to the declare attackers step.
func s82sBadgerTable(t *testing.T) (g *game.Game, me, opp *game.Player, badger, atk uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	badger = b12Push(g, opp.ID, "Giant Badger", "Creature — Badger", s82sBadgerOracle, 2, 2)
	atk = b12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	return g, me, opp, badger, atk
}

func TestS82sGiantBadgerGetsTwoTwoWhenItBlocks(t *testing.T) {
	g, _, _, badger, atk := s82sBadgerTable(t)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(badger, atk); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	if n := triggersOnStackFrom(g, badger); n != 1 {
		t.Fatalf("%d block triggers on the stack, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, badger), effectiveToughness(t, g, badger); p != 4 || tough != 4 {
		t.Errorf("the blocking Badger is %d/%d, want 4/4", p, tough)
	}
	// The 4/4 survives the 3/3 and kills it.
	advanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(badger) {
		t.Error("the pumped Badger died to a 3/3")
	}
	if g.Battlefield.Contains(atk) {
		t.Error("the 3/3 attacker survived 4 damage")
	}
}

func TestS82sGiantBadgerDoesNothingWhenItDoesNotBlock(t *testing.T) {
	g, _, _, badger, _ := s82sBadgerTable(t)
	advanceTo(t, g, game.StepDeclareBlockers)
	lockInBlocks(t, g)
	if n := triggersOnStackFrom(g, badger); n != 0 {
		t.Fatalf("%d triggers from a Badger that did not block", n)
	}
	if p, tough := effectivePower(t, g, badger), effectiveToughness(t, g, badger); p != 2 || tough != 2 {
		t.Errorf("the idle Badger is %d/%d, want 2/2", p, tough)
	}
}

// Attacking is not blocking.
func TestS82sGiantBadgerDoesNotGrowWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	badger := b12Push(g, me.ID, "Giant Badger", "Creature — Badger", s82sBadgerOracle, 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(badger, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if n := triggersOnStackFrom(g, badger); n != 0 {
		t.Fatalf("%d triggers from an attacking Badger", n)
	}
	if p := effectivePower(t, g, badger); p != 2 {
		t.Errorf("the attacking Badger has power %d, want 2", p)
	}
}

// --- Kemba's Skyguard ----------------------------------------------------------

func TestS82sKembasSkyguardGainsTwoLifeAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	id := msc17Play(t, g, me, game.Card{
		Name: "Kemba's Skyguard", TypeLine: "Creature — Cat Knight", OracleID: s82sSkyguardOracle,
		Power: 2, Toughness: 2,
	})
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Skyguard did not resolve onto the battlefield")
	}
	if me.Life != before+2 {
		t.Errorf("life %d, want %d", me.Life, before+2)
	}
	if !effectiveAbilitiesContain(t, g, id, "flying") {
		t.Error("Kemba's Skyguard should have flying")
	}
}

// --- Hellkite Overlord ---------------------------------------------------------

func TestS82sHellkiteOverlordHasFlyingTrampleAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Hellkite Overlord", "Creature — Dragon", s82sOverlordOracle, 8, 8)
	for _, kw := range []string{"flying", "trample", "haste"} {
		if !effectiveAbilitiesContain(t, g, id, kw) {
			t.Errorf("Hellkite Overlord should have %s", kw)
		}
	}
}

// {R} is +1/+0 each time it resolves, and it must be paid.
func TestS82sHellkiteOverlordFirebreathes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Hellkite Overlord", "Creature — Dragon", s82sOverlordOracle, 8, 8)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the {R} pump activated without mana")
	}
	b06AddMana(me, "R", "R")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 10 || tough != 8 {
		t.Errorf("after two pumps the Overlord is %d/%d, want 10/8", p, tough)
	}
}

// {B}{G} puts a shield on it that replaces the next destruction, and the
// shield costs both colours.
func TestS82sHellkiteOverlordRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Hellkite Overlord", "Creature — Dragon", s82sOverlordOracle, 8, 8)
	advanceToMain(t, g)
	b06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("regenerated for {B} alone")
	}
	b06AddMana(me, "G")
	b16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{})
	if n := regenShieldsOn(t, g, id); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	destroyIt(t, g, id)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the shield should replace the destruction")
	}
	if !b16Tapped(t, g, id) {
		t.Error("a regenerated creature is tapped")
	}
	if n := regenShieldsOn(t, g, id); n != 0 {
		t.Errorf("shields = %d after one destruction, want 0", n)
	}
}
