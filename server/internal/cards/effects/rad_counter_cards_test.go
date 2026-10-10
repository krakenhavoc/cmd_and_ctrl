package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rad_counter_cards_test.go pins the first rad-counter cards (#2042).
// The counters' own trigger (CR 728.1) is pinned in
// game/rad_counters_test.go; these are the cards that give them.

const (
	radBloatflySwarm     = "fc50e254-71bc-46b3-93e6-c6eb820a3473"
	radNuclearFallout    = "f09388e2-f190-4b48-ba85-3ffc33020392"
	radContaminatedDrink = "ee327c5e-ed29-4de3-bf50-92488f51a3a5"
	radMegatonsFate      = "09a688c1-21f6-4c7d-a6bd-aedc0d76a671"
	radVexingRadgull     = "2b632e67-ff80-4d52-9f44-3bc4bf083ab1"
	radFeralGhoul        = "284db94d-e583-4355-89fd-6c0906f9d17e"
	radGlowingOne        = "b30eae2a-bc1c-45a3-93ae-e1c30d26797c"
)

// radOf is a player's rad counters.
func radOf(p *game.Player) int { return p.Counters[game.CounterRad] }

// radEach asserts every seat's rad counters.
func radEach(t *testing.T, g *game.Game, want int) {
	t.Helper()
	for _, p := range g.Seats {
		if radOf(p) != want {
			t.Errorf("%s has %d rad counters, want %d", p.Name, radOf(p), want)
		}
	}
}

// Bloatfly Swarm: enters with five; 2 damage is prevented, two counters
// come off and every player gets two rad counters. Damage beyond its
// counters removes only what it has.
func TestBloatflySwarmTradesCountersForRad(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pr8Enter(t, g, me, "Bloatfly Swarm", "Creature — Insect Mutant", radBloatflySwarm)
	if n := pr8Counters(g, id, game.CounterPlusOne); n != 5 {
		t.Fatalf("entered with %d counters, want 5", n)
	}
	src := pr7Creature(g, opp.ID, "Shooter", 2, "R")
	pr6Damage(t, g, src, id, 2)
	if pr6Marked(g, id) != 0 || pr8Counters(g, id, game.CounterPlusOne) != 3 {
		t.Fatalf("2 damage: marked %d (want 0), counters %d (want 3)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
	radEach(t, g, 2)
	pr6Damage(t, g, src, id, 9)
	radEach(t, g, 5)
	if pr8Counters(g, id, game.CounterPlusOne) > 0 {
		t.Error("9 damage to three counters should leave none")
	}
}

func TestNuclearFalloutShrinksTwiceXAndGivesX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := pr7Creature(g, opp.ID, "Small", 4, "G")
	big := apaPush(g, me.ID, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Test", Power: 6, Toughness: 6})
	castXSpell(t, g, "Nuclear Fallout", "Sorcery", radNuclearFallout, "{X}{B}{B}", 2, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if findBattlefieldCardForTest(g, small) != nil {
		t.Error("a 4-toughness creature should die to -4/-4")
	}
	if c := findBattlefieldCardForTest(g, big); c == nil || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("the 6/6 should be a 2/2 until end of turn, got %v", c)
	}
	radEach(t, g, 2)
}

func TestContaminatedDrinkDrawsXAndGivesHalfRoundedUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castXSpell(t, g, "Contaminated Drink", "Instant", radContaminatedDrink, "{X}{U}{B}", 3, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if radOf(me) != 2 {
		t.Errorf("rad %d, want 2 (half of 3, rounded up)", radOf(me))
	}
	for _, p := range g.Seats[1:] {
		if radOf(p) != 0 {
			t.Errorf("%s got rad counters; only the caster does", p.Name)
		}
	}
}

func TestMegatonsFateDisarm(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	art := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	b19CastXModal(t, g, "Megaton's Fate", "Sorcery", radMegatonsFate, "{5}{R}", 0, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: art}})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, art) != nil {
		t.Error("the artifact should be destroyed")
	}
	if n := countTreasuresControlledBy(g, me.ID); n != 4 {
		t.Errorf("%d Treasures, want 4", n)
	}
	radEach(t, g, 0)
}

func TestMegatonsFateDetonate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, opp.ID, "A", 8, "G")
	b := apaPush(g, me.ID, me.ID, game.Card{Name: "B", TypeLine: "Creature — Test", Power: 9, Toughness: 9})
	castCatalogSpellWithModes(t, g, "Megaton's Fate", "Sorcery", radMegatonsFate, []int{1})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if findBattlefieldCardForTest(g, a) != nil {
		t.Error("an 8-toughness creature should die to 8 damage")
	}
	if findBattlefieldCardForTest(g, b) == nil {
		t.Error("a 9-toughness creature survives 8 damage")
	}
	radEach(t, g, 4)
}

// radCombatHit fakes one creature's combat damage to a player and
// settles the trigger it fires.
func radCombatHit(t *testing.T, g *game.Game, creature, victim uuid.UUID) {
	t.Helper()
	dealCombatDamageToPlayer(g, creature, victim, 1)
	passPriorityAroundTable(t, g)
}

// Vexing Radgull: two rad counters to a player with none; proliferate
// (the controller's choice) when they have some.
func TestVexingRadgullGivesTwoThenProliferates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gull := apaPush(g, me.ID, me.ID, game.Card{Name: "Vexing Radgull", OracleID: radVexingRadgull, TypeLine: "Creature — Bird Mutant", Power: 1, Toughness: 2})
	radCombatHit(t, g, gull, opp.ID)
	if radOf(opp) != 2 {
		t.Fatalf("rad %d after the first hit, want 2", radOf(opp))
	}
	radCombatHit(t, g, gull, opp.ID)
	answerProliferate(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if radOf(opp) != 3 {
		t.Errorf("rad %d after the second hit, want 3 (proliferated)", radOf(opp))
	}
}

func TestGlowingOneRadOnHitAndLifeOnNonlandMill(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := apaPush(g, me.ID, me.ID, game.Card{Name: "Glowing One", OracleID: radGlowingOne, TypeLine: "Creature — Zombie Mutant", Power: 2, Toughness: 2})
	radCombatHit(t, g, one, opp.ID)
	if radOf(opp) != 4 {
		t.Fatalf("rad %d, want 4", radOf(opp))
	}
	// Two nonland cards and a land on top of the opponent's library.
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre", Owner: opp.ID, Controller: opp.ID})
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(opp.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2 (one per nonland card)", got)
	}
}

func TestFeralGhoulGrowsAndIrradiatesOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ghoul := apaPush(g, me.ID, me.ID, game.Card{Name: "Feral Ghoul", OracleID: radFeralGhoul, TypeLine: "Creature — Zombie Mutant", Power: 2, Toughness: 2})
	other := apaPush(g, me.ID, me.ID, game.Card{Name: "Fodder", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(other, game.DestroyOptions{}); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := pr8Counters(g, ghoul, game.CounterPlusOne); n != 1 {
		t.Fatalf("Ghoul has %d +1/+1 counters, want 1", n)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(ghoul, game.DestroyOptions{}); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if radOf(me) != 0 {
		t.Errorf("the Ghoul's controller got %d rad counters; only opponents do", radOf(me))
	}
	for _, p := range g.Seats[1:] {
		if radOf(p) != 3 {
			t.Errorf("%s has %d rad counters, want 3 (its last power)", p.Name, radOf(p))
		}
	}
}
