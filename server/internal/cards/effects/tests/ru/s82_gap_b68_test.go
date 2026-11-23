package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Ravages of War -------------------------------------------------------

func TestGapB68RavagesDestroysEveryLandAndNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ravages of War", "Sorcery", gapB68Ravages, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Push(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	b := effectstest.B12Push(g, opp.ID, "B", "Creature — Bear", "", 2, 2)
	c := effectstest.B12Push(g, opp.ID, "C", "Creature — Bear", "", 2, 2)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	all := []uuid.UUID{a, b, c}
	for _, id := range all {
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
	}
	effectstest.CastCatalogSpell(t, g, "The Nipton Lottery", "Sorcery", gapB68NiptonLottery, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var survivors []uuid.UUID
	for _, id := range all {
		if g.Battlefield.Contains(id) {
			survivors = append(survivors, id)
		}
	}
	if len(survivors) != 1 {
		t.Fatalf("%d creatures survived, want exactly the one chosen at random", len(survivors))
	}
	chosen := effectstest.LayeredCard(t, g, survivors[0])
	if chosen.Controller != me.ID {
		t.Errorf("the chosen creature is controlled by %v, want me", chosen.Controller)
	}
	if chosen.Tapped {
		t.Error("the chosen creature is still tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, survivors[0], "haste") {
		t.Error("the chosen creature has no haste")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
}

func TestGapB68NiptonLotteryWithNoCreatureDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Island", "Basic Land — Island", "", false)
	effectstest.CastCatalogSpell(t, g, "The Nipton Lottery", "Sorcery", gapB68NiptonLottery, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
}

// --- Single Combat --------------------------------------------------------

func TestGapB68SingleCombatEachPlayerKeepsOneCreatureOrPlaneswalkerAndNobodyCastsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keepMe := effectstest.B12Push(g, me.ID, "Keep", "Creature — Bear", "", 3, 3)
	loseMe := effectstest.B12Push(g, me.ID, "Lose", "Creature — Bear", "", 1, 1)
	walker := effectstest.PushCatalogPermanent(g, me.ID, "Walker", "Legendary Planeswalker — Test", "", false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(walker, game.CounterLoyalty, 3) })
	lone := effectstest.B12Push(g, opp.ID, "Lone", "Creature — Bear", "", 2, 2)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	effectstest.CastCatalogSpell(t, g, "Single Combat", "Sorcery", gapB68SingleCombat, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatalf("I was not asked: %+v", g.PendingChoices)
	}
	if len(c.ChooseCards) != 3 {
		t.Errorf("%d candidates, want my two creatures and my planeswalker", len(c.ChooseCards))
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID) != nil {
		t.Error("the opponent with one creature was asked")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{keepMe}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"the other creature": loseMe, "the planeswalker": walker} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was not sacrificed", name)
		}
	}
	if !g.Battlefield.Contains(keepMe) || !g.Battlefield.Contains(lone) || !g.Battlefield.Contains(rock) {
		t.Error("a kept creature or a noncreature permanent was sacrificed")
	}
	// The ban: creature and planeswalker spells, every player, through my next turn.
	effectstest.ToMainForCost(t, g)
	for _, p := range []*game.Player{me, opp} {
		bear := effectstest.HandCardFull(p, "Test Bear", "Creature — Bear", "", "", nil)
		if err := g.CastSpell(p.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
			t.Errorf("%s cast a creature spell under Single Combat", p.Name)
		}
	}
}

// --- Diamond City ---------------------------------------------------------

func TestGapB68DiamondCityEntersShieldedAndMovesTheCounterOnlyAfterTwoCreaturesEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	city := effectstest.B31PlayLand(t, g, "Diamond City", "Land", gapB68DiamondCity)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, city, game.CounterShield); got != 1 {
		t.Fatalf("%d shield counters on entry, want 1", got)
	}
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	if err := g.ActivateCatalogAbility(me.ID, city, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err == nil {
		t.Fatal("the move was activated before two creatures entered this turn")
	}
	for i := 0; i < 2; i++ {
		effectstest.CastRingCreature(t, g, "Entrant", "Creature — Bear", "", 1, 1)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if err := g.ActivateCatalogAbility(me.ID, city, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("move the counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterShield); got != 1 {
		t.Errorf("the creature has %d shield counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, city, game.CounterShield); got != 0 {
		t.Errorf("the land kept %d shield counters, want 0", got)
	}
}
