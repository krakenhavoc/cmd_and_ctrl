package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TMNT Commander, batch 5 (#5011): Leonardo, the Balance and Donatello, the
// Brains. The batch's eight lands were already in the catalog.

const (
	tmcB5LeonardoOracle  = "46ea8d02-b087-44f7-8403-93ffc1d3a8ad"
	tmcB5DonatelloOracle = "1ba2f763-5895-4663-b38f-95591ebb2ef4"
)

func TestTmcB5EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmcB5LeonardoOracle:  "Leonardo, the Balance",
		tmcB5DonatelloOracle: "Donatello, the Brains",
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

func tmcB5Token(t *testing.T, g *game.Game, controller uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(controller, effects.TreasureToken(), 1); err != nil {
			t.Fatalf("create token: %v", err)
		}
	})
}

func tmcB5NoTriggerPrompt(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	if c := effectstest.LatestTriggerPrompt(g, who); c != nil {
		t.Fatalf("an unexpected \"you may\" is open: %+v", c)
	}
}

// --- Leonardo, the Balance ---

func TestTmcB5LeonardoCountersOnceEachTurnAfterADecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	leo := effectstest.B12Push(g, me.ID, "Leonardo, the Balance", "Legendary Creature — Mutant Ninja Turtle", tmcB5LeonardoOracle, 3, 3)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)

	// A "no" does not spend the turn's use.
	tmcB5Token(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HobB7Counters(t, g, bear) != 0 || effectstest.HobB7Counters(t, g, leo) != 0 {
		t.Fatal("a declined trigger put counters")
	}

	tmcB5Token(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HobB7Counters(t, g, bear) != 1 || effectstest.HobB7Counters(t, g, leo) != 1 {
		t.Fatalf("counters: bear %d, Leonardo %d; want one each", effectstest.HobB7Counters(t, g, bear), effectstest.HobB7Counters(t, g, leo))
	}

	// The use is spent: a third token this turn makes no trigger.
	tmcB5Token(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	tmcB5NoTriggerPrompt(t, g, me.ID)
	if effectstest.HobB7Counters(t, g, bear) != 1 {
		t.Error("a second set of counters this turn")
	}
}

func TestTmcB5LeonardoTwoTokensAtOnceCountOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Leonardo, the Balance", "Legendary Creature — Mutant Ninja Turtle", tmcB5LeonardoOracle, 3, 3)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 2); err != nil {
			t.Fatalf("create tokens: %v", err)
		}
	})
	// Both triggers are on the stack; the first "yes" spends the turn's use.
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	for i := 0; i < 4; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.LatestTriggerPrompt(g, me.ID); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
		}
	}
	if got := effectstest.HobB7Counters(t, g, bear); got != 1 {
		t.Errorf("bear counters = %d, want 1 (do this only once each turn)", got)
	}
}

func TestTmcB5LeonardoIgnoresAnOpponentsTokenAndNonTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Leonardo, the Balance", "Legendary Creature — Mutant Ninja Turtle", tmcB5LeonardoOracle, 3, 3)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	tmcB5Token(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	tmcB5NoTriggerPrompt(t, g, me.ID)
	effectstest.B12Push(g, me.ID, "Grizzly", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	tmcB5NoTriggerPrompt(t, g, me.ID)
	if effectstest.HobB7Counters(t, g, bear) != 0 {
		t.Error("counters without a token of yours entering")
	}
}

func TestTmcB5LeonardoGrantsMenaceTrampleAndLifelinkForFiveColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	leo := effectstest.B12Push(g, me.ID, "Leonardo, the Balance", "Legendary Creature — Mutant Ninja Turtle", tmcB5LeonardoOracle, 3, 3)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Their Bear", opp.ID)

	effectstest.FloatMana(t, g, me, "{W}{U}{B}{R}")
	if err := g.ActivateCatalogAbility(me.ID, leo, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without green")
	}
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, leo, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{leo, bear} {
		for _, kw := range []string{"menace", "trample", "lifelink"} {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s has no %s", id, kw)
			}
		}
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "trample") {
		t.Error("an opponent's creature gained trample")
	}
}

// --- Donatello, the Brains ---

func TestTmcB5DonatelloAddsOneMutagenPerCreation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Donatello, the Brains", "Legendary Creature — Mutant Ninja Turtle", tmcB5DonatelloOracle, 2, 4)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 2); err != nil {
			t.Fatalf("create tokens: %v", err)
		}
	})
	if got := effectstest.B16CountNamed(g, "Treasure"); got != 2 {
		t.Errorf("Treasures = %d, want 2", got)
	}
	if got := effectstest.B16CountNamed(g, "Mutagen"); got != 1 {
		t.Fatalf("Mutagens = %d, want 1 for the one instruction", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Mutagen" && (c.Controller != me.ID || !effects.IsToken(c) || !c.IsArtifact()) {
			t.Errorf("the Mutagen is not an artifact token of mine: %+v", c)
		}
	}
	// An opponent's creation gets none.
	tmcB5Token(t, g, opp.ID)
	if got := effectstest.B16CountNamed(g, "Mutagen"); got != 1 {
		t.Errorf("Mutagens = %d after an opponent's token, want 1", got)
	}
}

func TestTmcB5DonatellosMutagenEntersTappedWithATappedCreation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Donatello, the Brains", "Legendary Creature — Mutant Ninja Turtle", tmcB5DonatelloOracle, 2, 4)
	g.WithWriteLock(func() {
		if _, err := g.CreateTokensForEffect(me.ID, effects.TreasureToken(), 1, game.TokenEntryOptions{Tapped: true, Keywords: []string{"haste"}}); err != nil {
			t.Fatalf("create tokens: %v", err)
		}
	})
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Mutagen" {
			continue
		}
		found = true
		if !c.Tapped {
			t.Error("the Mutagen did not enter tapped with a tapped creation")
		}
		if effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "haste") {
			t.Error("the Mutagen got an ability the instruction's tokens were created with")
		}
	}
	if !found {
		t.Fatal("no Mutagen")
	}
}

func TestTmcB5DonatelloAndLeonardoTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Donatello, the Brains", "Legendary Creature — Mutant Ninja Turtle", tmcB5DonatelloOracle, 2, 4)
	leo := effectstest.B12Push(g, me.ID, "Leonardo, the Balance", "Legendary Creature — Mutant Ninja Turtle", tmcB5LeonardoOracle, 3, 3)
	tmcB5Token(t, g, me.ID)
	if got := effectstest.B16CountNamed(g, "Mutagen"); got != 1 {
		t.Fatalf("Mutagens = %d, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	for i := 0; i < 4; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.LatestTriggerPrompt(g, me.ID); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
		}
	}
	if got := effectstest.HobB7Counters(t, g, leo); got != 1 {
		t.Errorf("Leonardo counters = %d, want 1", got)
	}
}
