package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const heroesInAHalfShellOracle = "9f2f12cc-0adb-482d-bbce-05e4ce8850c4"

// heroesTable seeds Heroes in a Half Shell (5/5) for seat 0 beside a
// 1/1 Turtle, a 1/1 Human Ninja and a 2/2 Bear, all able to attack.
func heroesTable(t *testing.T) (g *game.Game, me *game.Player, heroes, turtle, ninja, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	advanceToMain(t, g)
	heroes = pushCatalogPermanent(g, me.ID, "Heroes in a Half Shell", "Legendary Creature — Mutant Ninja Turtle", heroesInAHalfShellOracle, false)
	battlefieldCardPtr(t, g, heroes).Power, battlefieldCardPtr(t, g, heroes).Toughness = 5, 5
	turtle = b12Creature(g, me.ID, "Test Turtle", "Creature — Turtle", 1, 1)
	ninja = b12Creature(g, me.ID, "Test Ninja", "Creature — Human Ninja", 1, 1)
	bear = b12Creature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	return g, me, heroes, turtle, ninja, bear
}

func TestHeroesInAHalfShellIsFullWithItsKeywords(t *testing.T) {
	spec, ok := Lookup(heroesInAHalfShellOracle)
	if !ok {
		t.Fatal("Heroes in a Half Shell is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	for _, kw := range []string{"vigilance", "menace", "trample", "haste"} {
		if !slices.Contains(spec.PrintedKeywords, kw) {
			t.Errorf("missing printed keyword %q in %v", kw, spec.PrintedKeywords)
		}
	}
}

func TestHeroesInAHalfShellCountersEachMutantNinjaTurtleThatHitAndDrawsOnce(t *testing.T) {
	g, me, heroes, turtle, ninja, bear := heroesTable(t)
	hand := me.Hand.Size()
	attackWith(t, g, g.Seats[1].ID, heroes, turtle, ninja, bear)
	mscM12Settle(t, g)
	if got := hobA4Power(t, g, heroes); got != 6 {
		t.Errorf("Heroes power %d, want 6 (one counter)", got)
	}
	if got := hobA4Power(t, g, turtle); got != 2 {
		t.Errorf("Turtle power %d, want 2 (one counter)", got)
	}
	if got := hobA4Power(t, g, ninja); got != 2 {
		t.Errorf("Ninja power %d, want 2 (one counter)", got)
	}
	if got := hobA4Power(t, g, bear); got != 2 {
		t.Errorf("Bear power %d, want 2 (not a Mutant, Ninja or Turtle)", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want exactly one card for the batch", hand, me.Hand.Size())
	}
}

// The ruling: one trigger per player damaged, and each puts counters only
// on the creatures that dealt combat damage to that player.
func TestHeroesInAHalfShellTriggersPerPlayerWithThatPlayersCreatures(t *testing.T) {
	g, me, heroes, turtle, ninja, _ := heroesTable(t)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	for id, defender := range map[uuid.UUID]uuid.UUID{heroes: g.Seats[1].ID, turtle: g.Seats[1].ID, ninja: g.Seats[2].ID} {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
	mscM12Settle(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want a card for each player hit", hand, me.Hand.Size())
	}
	for name, want := range map[uuid.UUID]int{heroes: 6, turtle: 2, ninja: 2} {
		if got := hobA4Power(t, g, name); got != want {
			t.Errorf("power %d, want %d (one counter each, from its own player's trigger)", got, want)
		}
	}
}

// A first-strike step and the regular step are two occurrences: the
// second trigger's creatures are only the ones that hit in that step.
func TestHeroesInAHalfShellFirstStrikeAndRegularDamageAreSeparateTriggers(t *testing.T) {
	g, me, _, turtle, ninja, _ := heroesTable(t)
	battlefieldCardPtr(t, g, ninja).Keywords = []string{"first strike"}
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{turtle, ninja} {
		if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepFirstStrikeDamage)
	mscM12Settle(t, g)
	if got := hobA4Power(t, g, ninja); got != 2 {
		t.Errorf("after first strike: Ninja power %d, want 2", got)
	}
	if got := hobA4Power(t, g, turtle); got != 1 {
		t.Errorf("after first strike: Turtle power %d, want 1 (it has not dealt damage yet)", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("after first strike: hand %d -> %d, want one card", hand, me.Hand.Size())
	}
	advanceTo(t, g, game.StepCombatDamage)
	mscM12Settle(t, g)
	if got := hobA4Power(t, g, ninja); got != 2 {
		t.Errorf("after regular damage: Ninja power %d, want 2 (it dealt no damage in this step)", got)
	}
	if got := hobA4Power(t, g, turtle); got != 2 {
		t.Errorf("after regular damage: Turtle power %d, want 2", got)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("after regular damage: hand %d -> %d, want two cards", hand, me.Hand.Size())
	}
}

func TestHeroesInAHalfShellIgnoresOtherCreaturesAndNoncombatDamage(t *testing.T) {
	g, me, _, turtle, _, bear := heroesTable(t)
	hand := me.Hand.Size()
	attackWith(t, g, g.Seats[1].ID, bear)
	mscM12Settle(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(turtle, g.Seats[1].ID, 1) })
	mscM12Settle(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw", hand, me.Hand.Size())
	}
	if got := hobA4Power(t, g, bear); got != 2 {
		t.Errorf("Bear power %d, want 2", got)
	}
	if got := hobA4Power(t, g, turtle); got != 1 {
		t.Errorf("Turtle power %d, want 1", got)
	}
}

// An opponent's Turtles are not "you control".
func TestHeroesInAHalfShellIgnoresAnOpponentsTurtles(t *testing.T) {
	g, me, _, _, _, _ := heroesTable(t)
	hand := me.Hand.Size()
	theirs := b12Creature(g, g.Seats[1].ID, "Their Turtle", "Creature — Turtle", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, g.Seats[2].ID, 1) })
	mscM12Settle(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw", hand, me.Hand.Size())
	}
}

func TestHeroesInAHalfShellGivesNoCounterToACreatureThatLeftButStillDraws(t *testing.T) {
	g, me, heroes, turtle, _, _ := heroesTable(t)
	hand := me.Hand.Size()
	attackWith(t, g, g.Seats[1].ID, heroes, turtle)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(turtle) })
	mscM12Settle(t, g)
	if got := hobA4Power(t, g, heroes); got != 6 {
		t.Errorf("Heroes power %d, want 6", got)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want the draw plus the bounced Turtle", hand, me.Hand.Size())
	}
}
