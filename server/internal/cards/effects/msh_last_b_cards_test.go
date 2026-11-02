package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_last_b_cards_test.go — the last three Marvel Super Heroes cards (#4455).

const (
	mshLastSueOracle    = "e9d20b6d-fa4f-473b-9866-230fd92a1319"
	mshLastWidowOracle  = "f3882587-7aff-451f-a990-e25039e5ef85"
	mshLastHappyOracle  = "557ce9d4-8fe3-4ffc-8e56-b9b73d58d41b"
	mshLastHeroTypeLine = "Creature — Human Hero"
)

func mshLastWalls(g *game.Game, owner uuid.UUID) int {
	return mshB1Tokens(g, owner, "Wall")
}

func TestMshLastSueStormHasLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sue := b12Push(g, me.ID, "Invisible Woman, Sue Storm", "Legendary Creature — Human Hero", mshLastSueOracle, 2, 5)
	if !hasEffectiveKeyword(t, g, sue, "lifelink") {
		t.Fatal("Sue Storm has lifelink")
	}
}

func TestMshLastSueStormMayMakeAWallWhenYouCountersAnotherHero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Invisible Woman, Sue Storm", "Legendary Creature — Human Hero", mshLastSueOracle, 2, 5)
	hero := b12Creature(g, me.ID, "Some Hero", mshLastHeroTypeLine, 2, 2)
	advanceToMain(t, g)
	if err := g.AddCounterBy(me.ID, hero, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := mshLastWalls(g, me.ID); n != 1 {
		t.Fatalf("Walls = %d, want 1", n)
	}
	wall := mshB1Token(t, g, me.ID, "Wall")
	if wall.Power != 0 || wall.Toughness != 4 || !hasEffectiveKeyword(t, g, wall.InstanceID, "defender") {
		t.Errorf("want a 0/4 Wall with defender, got %d/%d", wall.Power, wall.Toughness)
	}
}

func TestMshLastSueStormDecliningMakesNoWall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Invisible Woman, Sue Storm", "Legendary Creature — Human Hero", mshLastSueOracle, 2, 5)
	hero := b12Creature(g, me.ID, "Some Hero", mshLastHeroTypeLine, 2, 2)
	advanceToMain(t, g)
	if err := g.AddCounterBy(me.ID, hero, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := mshLastWalls(g, me.ID); n != 0 {
		t.Fatalf("Walls = %d, want 0", n)
	}
}

func TestMshLastSueStormOneWallForCountersOnSeveralHeroesAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Invisible Woman, Sue Storm", "Legendary Creature — Human Hero", mshLastSueOracle, 2, 5)
	a := b12Creature(g, me.ID, "Hero A", mshLastHeroTypeLine, 2, 2)
	b := b12Creature(g, me.ID, "Hero B", mshLastHeroTypeLine, 2, 2)
	advanceToMain(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.AddCounterBy(me.ID, id, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterBy: %v", err)
		}
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := mshLastWalls(g, me.ID); n != 1 {
		t.Fatalf("Walls = %d, want 1 (one trigger for the batch)", n)
	}
}

func TestMshLastSueStormIgnoresSelfNonHeroesOpponentsAndOtherCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sue := b12Push(g, me.ID, "Invisible Woman, Sue Storm", "Legendary Creature — Human Hero", mshLastSueOracle, 2, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Hero", mshLastHeroTypeLine, 2, 2)
	hero := b12Creature(g, me.ID, "Some Hero", mshLastHeroTypeLine, 2, 2)
	advanceToMain(t, g)
	for _, c := range []struct {
		placer, target uuid.UUID
		kind           string
	}{
		{me.ID, sue, game.CounterPlusOne},    // herself
		{me.ID, bear, game.CounterPlusOne},   // not a Hero
		{me.ID, theirs, game.CounterPlusOne}, // an opponent's Hero
		{me.ID, hero, "charge"},              // not a +1/+1 counter
		{opp.ID, hero, game.CounterPlusOne},  // an opponent put it
	} {
		if err := g.AddCounterBy(c.placer, c.target, c.kind, 1); err != nil {
			t.Fatalf("AddCounterBy: %v", err)
		}
	}
	if c := latestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID); c != nil {
		t.Fatal("no trigger prompt should be open")
	}
	passPriorityAroundTable(t, g)
	if n := mshLastWalls(g, me.ID); n != 0 {
		t.Fatalf("Walls = %d, want 0", n)
	}
}

func TestMshLastVanillasRegisterWithNoAbilities(t *testing.T) {
	for oracle, name := range map[string]string{
		mshLastWidowOracle: "Black Widow, Natasha Romanoff",
		mshLastHappyOracle: "Happy Hogan, Dauntless Driver",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", name)
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q completeness %v", name, spec.Name, spec.Completeness)
		}
		if len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s must have no abilities", name)
		}
	}
}
