package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_restriction_cards_test.go — #1210's card proofs.
//
// The engine test (game/activation_gate_test.go) pins the gate
// against stub restrictions; these drive the four printed cards
// through the real catalog, because the half of this issue that can
// go wrong per card is which clause the card file reached for: a
// Cursed Totem written with the mana exemption is a blank against the
// deck it exists to beat, and a Linvala written against the activator
// instead of its own controller breaks the moment somebody steals it.

const (
	actKrenkoOracle         = "68418069-f615-40ef-ae0d-764192acae00"
	actSolRingOracle        = "6ad8011d-3471-4369-9d68-b264cc027487"
	actBirdsOracle          = "d3a0b660-358c-41bd-9cd2-41fbf3491b1a"
	actLlanowarElvesOracle  = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
	actCursedTotemOracle    = "6225a704-430a-4f56-ad87-0e8d87f285f5"
	actLinvalaOracle        = "88dadc31-dfac-41b3-bf2a-65fa89e3c16d"
	actCollectorOupheOracle = "0c4bc9ea-a5fd-4f44-96a1-5448eee228c4"
	actPithingNeedleOracle  = "a188fe7e-68de-4c7c-806c-bfe8fc7b44bf"
	actHarshMentorOracle    = "168336e2-d795-4b75-bf21-ce128b0dd7b0"
)

// actPushKrenko seeds a Krenko, Mob Boss — a creature with a NON-mana
// {T} ability, which is the thing an activation restriction is really
// about.
func actPushKrenko(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b21Push(g, owner, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", actKrenkoOracle, 3, 3, "R")
}

// actNameFor answers the open choose_card_name prompt addressed to
// `chooser` with `name`.
func actNameFor(t *testing.T, g *game.Game, chooser uuid.UUID, name string) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoiceCardName || c.Chooser != chooser {
			continue
		}
		if err := g.ResolveCardNameChoice(c.ID, chooser, name); err != nil {
			t.Fatalf("ResolveCardNameChoice(%q): %v", name, err)
		}
		return
	}
	t.Fatalf("no open choose_card_name prompt for %s", chooser)
}

// TestCursedTotemStopsCreaturesAndNotArtifacts is the card's whole
// text in two assertions, and the second is the one a mana exemption
// would have broken: Cursed Totem does NOT print "unless they're mana
// abilities", so the Birds is off too.
func TestCursedTotemStopsCreaturesAndNotArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]

	krenko := actPushKrenko(g, me.ID)
	birds := b21Push(g, me.ID, "Birds of Paradise", "Creature — Bird", actBirdsOracle, 0, 1, "G")
	solRing := b21Push(g, me.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)
	b21Push(g, me.ID, "Cursed Totem", "Artifact", actCursedTotemOracle, 0, 0)

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
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]

	solRing := b21Push(g, me.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)
	birds := b21Push(g, me.ID, "Birds of Paradise", "Creature — Bird", actBirdsOracle, 0, 1, "G")
	b21Push(g, me.ID, "Collector Ouphe", "Creature — Ouphe", actCollectorOupheOracle, 2, 2, "G")

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
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]

	b21Push(g, me.ID, "Linvala, Keeper of Silence", "Legendary Creature — Angel", actLinvalaOracle, 3, 4, "W")
	mine := b21Push(g, me.ID, "Birds of Paradise", "Creature — Bird", actBirdsOracle, 0, 1, "G")
	theirs := b21Push(g, them.ID, "Birds of Paradise", "Creature — Bird", actBirdsOracle, 0, 1, "G")

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
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]

	needle := b21Push(g, me.ID, "Pithing Needle", "Artifact", actPithingNeedleOracle, 0, 0)
	// The permanent is seeded straight onto the battlefield, so the
	// entry hook has not run; queue the prompt the way the hook does.
	g.WithWriteLock(func() {
		g.QueueCardNameChoiceForEffect(me.ID, needle, "Pithing Needle — choose a card name")
	})
	actNameFor(t, g, me.ID, "Krenko, Mob Boss")

	myKrenko := actPushKrenko(g, me.ID)
	theirKrenko := actPushKrenko(g, them.ID)
	solRing := b21Push(g, me.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)

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
	solRing2 := b21Push(g, me.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)
	needle2 := b21Push(g, them.ID, "Pithing Needle", "Artifact", actPithingNeedleOracle, 0, 0)
	g.WithWriteLock(func() {
		g.QueueCardNameChoiceForEffect(them.ID, needle2, "Pithing Needle — choose a card name")
	})
	actNameFor(t, g, them.ID, "Sol Ring")
	if err := g.ActivateManaAbility(me.ID, solRing2, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("a NAMED source's mana ability: err = %v, want nil — the clause exempts mana abilities", err)
	}
}

// TestHarshMentorBurnsTheOpponentWhoActivated is the watch: it fires
// for an opponent's non-mana activation and for nothing else.
func TestHarshMentorBurnsTheOpponentWhoActivated(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]

	b21Push(g, me.ID, "Harsh Mentor", "Creature — Human Cleric", actHarshMentorOracle, 2, 2, "R")
	theirKrenko := actPushKrenko(g, them.ID)
	myKrenko := actPushKrenko(g, me.ID)
	// Llanowar Elves and not Birds of Paradise: "add one mana of any
	// color" opens a mana_pick, an open prompt stops priority passing,
	// and the LAST assertion in this test needs a trigger to resolve.
	theirElves := b21Push(g, them.ID, "Llanowar Elves", "Creature — Elf Druid", actLlanowarElvesOracle, 1, 1, "G")

	before := them.Life

	// The Mentor's controller activating their own Krenko: nothing.
	if err := g.ActivateCatalogAbility(me.ID, myKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("own Krenko: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != 40 || them.Life != before {
		t.Errorf("after the Mentor controller's OWN activation: %d / %d, want 40 / %d — the clause says an opponent", me.Life, them.Life, before)
	}

	// An opponent's mana ability: nothing (CR 605.1a, and the trigger
	// does not watch that kind at all).
	if err := g.ActivateManaAbility(them.ID, theirElves, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("their Llanowar Elves: %v", err)
	}
	passPriorityAroundTable(t, g)
	if them.Life != before {
		t.Errorf("after an opponent's MANA ability: %d, want %d — 'if it isn't a mana ability'", them.Life, before)
	}

	// An opponent's non-mana activation: 2 damage to THAT player.
	if err := g.ActivateCatalogAbility(them.ID, theirKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("their Krenko: %v", err)
	}
	passPriorityAroundTable(t, g)
	if them.Life != before-2 {
		t.Errorf("after an opponent's activation: %d, want %d", them.Life, before-2)
	}
	if me.Life != 40 {
		t.Errorf("the Mentor's controller took %d — the damage goes to the activator", 40-me.Life)
	}
}
