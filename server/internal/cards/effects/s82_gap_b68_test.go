package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b68_test.go — Fallout lands and sorceries (GAP batch 68, #5948),
// played through the real catalog entries.

const (
	gapB68DiamondCity   = "88e29d50-1680-495d-be84-b92b4c9e636f"
	gapB68NiptonLottery = "8af798a9-2198-4f39-aa6d-5420a0f1203c"
	gapB68Ravages       = "265236d1-d96c-4424-835e-a0fe91c41170"
	gapB68SingleCombat  = "822ef47e-3f1b-40a6-8c5c-fea3d7deba5b"
)

func TestGapB68CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB68DiamondCity: "Diamond City", gapB68NiptonLottery: "The Nipton Lottery",
		gapB68Ravages: "Ravages of War", gapB68SingleCombat: "Single Combat",
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

// --- Ravages of War -------------------------------------------------------

func TestGapB68RavagesDestroysEveryLandAndNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Ravages of War", "Sorcery", gapB68Ravages, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a land survived")
	}
	if !me.Graveyard.Contains(mine) || !opp.Graveyard.Contains(theirs) {
		t.Error("a destroyed land did not go to its owner's graveyard")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a nonland permanent was destroyed")
	}
}

// --- The Nipton Lottery ---------------------------------------------------

func TestGapB68NiptonLotteryTakesOneCreatureUntappedWithHasteAndDestroysTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Push(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	b := b12Push(g, opp.ID, "B", "Creature — Bear", "", 2, 2)
	c := b12Push(g, opp.ID, "C", "Creature — Bear", "", 2, 2)
	land := pushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	all := []uuid.UUID{a, b, c}
	for _, id := range all {
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
	}
	castCatalogSpell(t, g, "The Nipton Lottery", "Sorcery", gapB68NiptonLottery, nil)
	passPriorityAroundTable(t, g)
	var survivors []uuid.UUID
	for _, id := range all {
		if g.Battlefield.Contains(id) {
			survivors = append(survivors, id)
		}
	}
	if len(survivors) != 1 {
		t.Fatalf("%d creatures survived, want exactly the one chosen at random", len(survivors))
	}
	chosen := layeredCard(t, g, survivors[0])
	if chosen.Controller != me.ID {
		t.Errorf("the chosen creature is controlled by %v, want me", chosen.Controller)
	}
	if chosen.Tapped {
		t.Error("the chosen creature is still tapped")
	}
	if !hasEffectiveKeyword(t, g, survivors[0], "haste") {
		t.Error("the chosen creature has no haste")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
}

func TestGapB68NiptonLotteryWithNoCreatureDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	land := pushCatalogPermanent(g, g.Seats[1].ID, "Island", "Basic Land — Island", "", false)
	castCatalogSpell(t, g, "The Nipton Lottery", "Sorcery", gapB68NiptonLottery, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
}

// --- Single Combat --------------------------------------------------------

func TestGapB68SingleCombatEachPlayerKeepsOneCreatureOrPlaneswalkerAndNobodyCastsThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keepMe := b12Push(g, me.ID, "Keep", "Creature — Bear", "", 3, 3)
	loseMe := b12Push(g, me.ID, "Lose", "Creature — Bear", "", 1, 1)
	walker := pushCatalogPermanent(g, me.ID, "Walker", "Legendary Planeswalker — Test", "", false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(walker, game.CounterLoyalty, 3) })
	lone := b12Push(g, opp.ID, "Lone", "Creature — Bear", "", 2, 2)
	rock := pushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	castCatalogSpell(t, g, "Single Combat", "Sorcery", gapB68SingleCombat, nil)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatalf("I was not asked: %+v", g.PendingChoices)
	}
	if len(c.ChooseCards) != 3 {
		t.Errorf("%d candidates, want my two creatures and my planeswalker", len(c.ChooseCards))
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID) != nil {
		t.Error("the opponent with one creature was asked")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{keepMe}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"the other creature": loseMe, "the planeswalker": walker} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was not sacrificed", name)
		}
	}
	if !g.Battlefield.Contains(keepMe) || !g.Battlefield.Contains(lone) || !g.Battlefield.Contains(rock) {
		t.Error("a kept creature or a noncreature permanent was sacrificed")
	}
	// The ban: creature and planeswalker spells, every player, through my next turn.
	toMainForCost(t, g)
	for _, p := range []*game.Player{me, opp} {
		bear := handCardFull(p, "Test Bear", "Creature — Bear", "", "", nil)
		if err := g.CastSpell(p.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
			t.Errorf("%s cast a creature spell under Single Combat", p.Name)
		}
	}
}

// --- Diamond City ---------------------------------------------------------

func TestGapB68DiamondCityEntersShieldedAndMovesTheCounterOnlyAfterTwoCreaturesEntered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	city := b31PlayLand(t, g, "Diamond City", "Land", gapB68DiamondCity)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, city, game.CounterShield); got != 1 {
		t.Fatalf("%d shield counters on entry, want 1", got)
	}
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	if err := g.ActivateCatalogAbility(me.ID, city, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err == nil {
		t.Fatal("the move was activated before two creatures entered this turn")
	}
	for i := 0; i < 2; i++ {
		castRingCreature(t, g, "Entrant", "Creature — Bear", "", 1, 1)
		passPriorityAroundTable(t, g)
	}
	if err := g.ActivateCatalogAbility(me.ID, city, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("move the counter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterShield); got != 1 {
		t.Errorf("the creature has %d shield counters, want 1", got)
	}
	if got := countersOn(g, city, game.CounterShield); got != 0 {
		t.Errorf("the land kept %d shield counters, want 0", got)
	}
}
