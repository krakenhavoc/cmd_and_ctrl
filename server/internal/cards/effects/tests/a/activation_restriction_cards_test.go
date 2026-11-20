package testsa

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestCursedTotemStopsCreaturesAndNotArtifacts is the card's whole
// text in two assertions, and the second is the one a mana exemption
// would have broken: Cursed Totem does NOT print "unless they're mana
// abilities", so the Birds is off too.
func TestCursedTotemStopsCreaturesAndNotArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]

	krenko := effectstest.ActPushKrenko(g, me.ID)
	birds := effectstest.B21Push(g, me.ID, "Birds of Paradise", "Creature — Bird", effectstest.ActBirdsOracle, 0, 1, "G")
	solRing := effectstest.B21Push(g, me.ID, "Sol Ring", "Artifact", effectstest.ActSolRingOracle, 0, 0)
	effectstest.B21Push(g, me.ID, "Cursed Totem", "Artifact", effectstest.ActCursedTotemOracle, 0, 0)

	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("Krenko's {T}: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateManaAbility(me.ID, birds, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the Birds' mana ability: err = %v, want ErrCantActivate — Cursed Totem prints no mana exemption", err)
	}
	if err := g.ActivateManaAbility(me.ID, solRing, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("Sol Ring: err = %v, want nil — Cursed Totem names creatures, not artifacts", err)
	}
}

// TestCollectorOupheStopsArtifactsAndNotCreatures is the same clause
// with the type swapped, and the assertion that the two cards did not
// end up sharing a predicate by accident.
func TestCollectorOupheStopsArtifactsAndNotCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]

	solRing := effectstest.B21Push(g, me.ID, "Sol Ring", "Artifact", effectstest.ActSolRingOracle, 0, 0)
	birds := effectstest.B21Push(g, me.ID, "Birds of Paradise", "Creature — Bird", effectstest.ActBirdsOracle, 0, 1, "G")
	effectstest.B21Push(g, me.ID, "Collector Ouphe", "Creature — Ouphe", effectstest.ActCollectorOupheOracle, 2, 2, "G")

	if err := g.ActivateManaAbility(me.ID, solRing, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("Sol Ring: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateManaAbility(me.ID, birds, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("the Birds: err = %v, want nil — the Ouphe names artifacts", err)
	}
}

// TestLinvalaStopsYourOpponentsCreaturesOnly is the asymmetrical
// half, and it is written from the OPPONENT's seat on purpose: the
// clause measures against Linvala's controller, so a version that
// compared "not the activator" would pass every symmetric test and
// fail this one.
func TestLinvalaStopsYourOpponentsCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]

	effectstest.B21Push(g, me.ID, "Linvala, Keeper of Silence", "Legendary Creature — Angel", effectstest.ActLinvalaOracle, 3, 4, "W")
	mine := effectstest.B21Push(g, me.ID, "Birds of Paradise", "Creature — Bird", effectstest.ActBirdsOracle, 0, 1, "G")
	theirs := effectstest.B21Push(g, them.ID, "Birds of Paradise", "Creature — Bird", effectstest.ActBirdsOracle, 0, 1, "G")

	if err := g.ActivateManaAbility(them.ID, theirs, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the opponent's Birds: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("Linvala's controller's own Birds: err = %v, want nil", err)
	}
}

// TestPithingNeedleStopsTheNamedSourceForEveryPlayer is the chosen-
// name card end to end: the prompt, the answer, the ban on the named
// source for EVERY seat (the clause says "sources", not "sources your
// opponents control"), the mana exemption, and the sources it does
// not name.
func TestPithingNeedleStopsTheNamedSourceForEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]

	needle := effectstest.B21Push(g, me.ID, "Pithing Needle", "Artifact", effectstest.ActPithingNeedleOracle, 0, 0)
	// The permanent is seeded straight onto the battlefield, so the
	// entry hook has not run; queue the prompt the way the hook does.
	g.WithWriteLock(func() {
		g.QueueCardNameChoiceForEffect(me.ID, needle, "Pithing Needle — choose a card name")
	})
	effectstest.ActNameFor(t, g, me.ID, "Krenko, Mob Boss")

	myKrenko := effectstest.ActPushKrenko(g, me.ID)
	theirKrenko := effectstest.ActPushKrenko(g, them.ID)
	solRing := effectstest.B21Push(g, me.ID, "Sol Ring", "Artifact", effectstest.ActSolRingOracle, 0, 0)

	if err := g.ActivateCatalogAbility(them.ID, theirKrenko, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the opponent's named source: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, myKrenko, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the Needle controller's OWN named source: err = %v, want ErrCantActivate — the clause names sources, not opponents", err)
	}
	if err := g.ActivateManaAbility(me.ID, solRing, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("an unnamed source: err = %v, want nil", err)
	}

	// And the mana exemption, on a source that IS named. A SECOND Sol
	// Ring, because the first one above is tapped — "already tapped"
	// and "the gate refused it" are both non-nil errors and the test
	// must not be able to pass for the wrong one.
	solRing2 := effectstest.B21Push(g, me.ID, "Sol Ring", "Artifact", effectstest.ActSolRingOracle, 0, 0)
	needle2 := effectstest.B21Push(g, them.ID, "Pithing Needle", "Artifact", effectstest.ActPithingNeedleOracle, 0, 0)
	g.WithWriteLock(func() {
		g.QueueCardNameChoiceForEffect(them.ID, needle2, "Pithing Needle — choose a card name")
	})
	effectstest.ActNameFor(t, g, them.ID, "Sol Ring")
	if err := g.ActivateManaAbility(me.ID, solRing2, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("a NAMED source's mana ability: err = %v, want nil — the clause exempts mana abilities", err)
	}
}

// TestHarshMentorBurnsTheOpponentWhoActivated is the watch: it fires
// for an opponent's non-mana activation and for nothing else.
func TestHarshMentorBurnsTheOpponentWhoActivated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]

	effectstest.B21Push(g, me.ID, "Harsh Mentor", "Creature — Human Cleric", effectstest.ActHarshMentorOracle, 2, 2, "R")
	theirKrenko := effectstest.ActPushKrenko(g, them.ID)
	myKrenko := effectstest.ActPushKrenko(g, me.ID)
	// Llanowar Elves and not Birds of Paradise: "add one mana of any
	// color" opens a mana_pick, an open prompt stops priority passing,
	// and the LAST assertion in this test needs a trigger to resolve.
	theirElves := effectstest.B21Push(g, them.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.ActLlanowarElvesOracle, 1, 1, "G")

	before := them.Life

	// The Mentor's controller activating their own Krenko: nothing.
	if err := g.ActivateCatalogAbility(me.ID, myKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("own Krenko: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 40 || them.Life != before {
		t.Errorf("after the Mentor controller's OWN activation: %d / %d, want 40 / %d — the clause says an opponent", me.Life, them.Life, before)
	}

	// An opponent's mana ability: nothing (CR 605.1a, and the trigger
	// does not watch that kind at all).
	if err := g.ActivateManaAbility(them.ID, theirElves, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("their Llanowar Elves: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != before {
		t.Errorf("after an opponent's MANA ability: %d, want %d — 'if it isn't a mana ability'", them.Life, before)
	}

	// An opponent's non-mana activation: 2 damage to THAT player.
	if err := g.ActivateCatalogAbility(them.ID, theirKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("their Krenko: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != before-2 {
		t.Errorf("after an opponent's activation: %d, want %d", them.Life, before-2)
	}
	if me.Life != 40 {
		t.Errorf("the Mentor's controller took %d — the damage goes to the activator", 40-me.Life)
	}
}
