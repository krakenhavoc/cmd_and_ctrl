package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// titania_proud_pummeler_test.go — Titania, Proud Pummeler (#4455, seam
// #4722). The engine half of melee (CR 702.121) is game/melee_test.go;
// these tests play the real catalog entry.

const titaniaOracle = "6d51f00e-46a5-4064-a0c8-e840c66dbf2b"

func titaniaBoard(t *testing.T) (g *game.Game, titania, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titania = b12Push(g, me.ID, "Titania, Proud Pummeler", "Legendary Creature — Human Villain", titaniaOracle, 3, 3)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	return
}

func meleeInstances(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := 0
	for _, a := range effectiveAbilities(t, g, id) {
		if a == game.KeywordMelee {
			n++
		}
	}
	return n
}

// attackSplit declares each attacker at its own defender and advances
// into the combat damage step, so every melee trigger has resolved.
func attackSplit(t *testing.T, g *game.Game, declarations map[uuid.UUID]uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for attacker, defender := range declarations {
		if err := g.DeclareAttacker(attacker, defender); err != nil {
			t.Fatalf("DeclareAttacker %s: %v", attacker, err)
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
}

func TestTitaniaIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(titaniaOracle)
	if !ok {
		t.Fatal("Titania, Proud Pummeler is not registered")
	}
	if spec.Name != "Titania, Proud Pummeler" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func TestTitaniaHasMeleeHerselfAndGivesItToOtherCreaturesYouControlOnly(t *testing.T) {
	g, titania, bear := titaniaBoard(t)
	assertKeywords(t, g, titania, "first strike")
	if n := meleeInstances(t, g, titania); n != 1 {
		t.Errorf("Titania has %d instances of melee, want 1 (her own; the grant says other)", n)
	}
	if n := meleeInstances(t, g, bear); n != 1 {
		t.Errorf("another creature I control has %d instances of melee, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" && meleeInstances(t, g, c.InstanceID) != 0 {
			t.Error("an opponent's creature has melee")
		}
	}
}

func TestTitaniaAttackingAloneAtOneOpponentGetsOne(t *testing.T) {
	g, titania, _ := titaniaBoard(t)
	attackSplit(t, g, map[uuid.UUID]uuid.UUID{titania: g.Seats[1].ID})
	if got := hobA4Power(t, g, titania); got != 4 {
		t.Errorf("Titania has power %d, want 4 (3 base + one opponent attacked)", got)
	}
}

func TestTitaniaAndAnotherCreatureCountEveryOpponentAttacked(t *testing.T) {
	g, titania, bear := titaniaBoard(t)
	attackSplit(t, g, map[uuid.UUID]uuid.UUID{titania: g.Seats[1].ID, bear: g.Seats[2].ID})
	// Two opponents were attacked this combat, by two different
	// creatures: each melee counts both.
	if got := hobA4Power(t, g, titania); got != 5 {
		t.Errorf("Titania has power %d, want 5 (3 base + two opponents)", got)
	}
	if got := hobA4Power(t, g, bear); got != 4 {
		t.Errorf("the granted-melee Bear has power %d, want 4 (2 base + two opponents)", got)
	}
}

func TestTitaniaTwoCreaturesAtTheSameOpponentCountThemOnce(t *testing.T) {
	g, titania, bear := titaniaBoard(t)
	attackSplit(t, g, map[uuid.UUID]uuid.UUID{titania: g.Seats[1].ID, bear: g.Seats[1].ID})
	if got := hobA4Power(t, g, titania); got != 4 {
		t.Errorf("Titania has power %d, want 4 (one opponent, attacked twice)", got)
	}
	if got := hobA4Power(t, g, bear); got != 3 {
		t.Errorf("the Bear has power %d, want 3", got)
	}
}

func TestTitaniaGivesACreatureThatPrintsMeleeTwoInstances(t *testing.T) {
	g, _, _ := titaniaBoard(t)
	me := g.Seats[0]
	brawler := b12Creature(g, me.ID, "Brawler", "Creature — Human Warrior", 2, 2)
	battlefieldCardPtr(t, g, brawler).Keywords = []string{game.KeywordMelee}
	if n := meleeInstances(t, g, brawler); n != 2 {
		t.Fatalf("a creature that prints melee has %d instances under Titania, want 2 (CR 702.121b)", n)
	}
	attackSplit(t, g, map[uuid.UUID]uuid.UUID{brawler: g.Seats[1].ID})
	// Two separate triggers, each +1/+1 for the one opponent attacked.
	if got := hobA4Power(t, g, brawler); got != 4 {
		t.Errorf("Brawler has power %d, want 4 (2 base + two instances x one opponent)", got)
	}
}

func TestTitaniaLeavingTakesTheGrantAway(t *testing.T) {
	g, titania, bear := titaniaBoard(t)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(titania) })
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(titania) {
		t.Fatal("Titania did not leave")
	}
	if n := meleeInstances(t, g, bear); n != 0 {
		t.Errorf("the Bear still has %d instances of melee", n)
	}
	attackSplit(t, g, map[uuid.UUID]uuid.UUID{bear: g.Seats[1].ID})
	if got := hobA4Power(t, g, bear); got != 2 {
		t.Errorf("the Bear has power %d, want 2 (no melee)", got)
	}
}
