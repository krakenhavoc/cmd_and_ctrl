package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bolster_test.go — bolster (CR 701.39, #3277): the primitive on plain
// creatures, then Optimus Prime, Hero // Optimus Prime, Autobot Leader.

// testBolsterChosen is the creature the bolster's keyed tail
// (testBolsterHeard) was told; Nil when it never ran.
var (
	testBolsterChosen uuid.UUID
	testBolsterHeard  = CardTail("test/4253-bolster-heard", func(_ *Context, o game.TailOutcome) error {
		if len(o.IDs) == 1 {
			testBolsterChosen = o.IDs[0]
		}
		return nil
	})
)

// runBolster resolves Bolster for the player as if an ability of theirs
// did, and returns the chosen creature its tail was told (Nil when it
// never ran).
func runBolster(t *testing.T, g *game.Game, me *game.Player, n int) uuid.UUID {
	t.Helper()
	testBolsterChosen = uuid.Nil
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID}
		ctx := NewContext(g, item)
		if err := (Bolster{N: n, ThenTail: ctx.Tail(testBolsterHeard)}).Apply(ctx); err != nil {
			t.Fatalf("Bolster: %v", err)
		}
	})
	return testBolsterChosen
}

func bolsterCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return flipState(t, g, id).Counters[game.CounterPlusOne]
}

// CR 701.39a: the creature with the least toughness gets the counters,
// and nobody else.
func TestBolsterPutsCountersOnTheLeastToughness(t *testing.T) {
	g, me, _ := flipMain(t)
	big := pushVanillaCreature(g, me.ID, "Big", 4, 4)
	small := pushVanillaCreature(g, me.ID, "Small", 3, 1)
	mid := pushVanillaCreature(g, me.ID, "Mid", 2, 3)

	chosen := runBolster(t, g, me, 2)

	if chosen != small {
		t.Fatalf("chose %v, want the 3/1", chosen)
	}
	if got := bolsterCounters(t, g, small); got != 2 {
		t.Errorf("the 3/1 has %d +1/+1 counters, want 2", got)
	}
	if bolsterCounters(t, g, big) != 0 || bolsterCounters(t, g, mid) != 0 {
		t.Error("a creature that wasn't the least toughness got counters")
	}
	if len(g.PendingChoices) != 0 {
		t.Error("a single candidate was still asked about")
	}
}

// Toughness is read after counters and layers, and only creatures the
// bolstering player controls count: an opponent's 0-toughness-ish
// creature is not a candidate.
func TestBolsterReadsEffectiveToughnessAmongYourCreatures(t *testing.T) {
	g, me, opp := flipMain(t)
	grown := pushVanillaCreature(g, me.ID, "Grown", 1, 1)
	plain := pushVanillaCreature(g, me.ID, "Plain", 1, 3)
	pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(grown, game.CounterPlusOne, 3); err != nil {
			t.Fatal(err)
		}
	})

	chosen := runBolster(t, g, me, 1)

	if chosen != plain {
		t.Fatalf("chose %v, want the 1/3 (the 1/1 is a 4/4 with its counters)", chosen)
	}
	if bolsterCounters(t, g, plain) != 1 {
		t.Error("the 1/3 has no counter")
	}
}

// CR 701.39a: tied for least, the controller picks, and only the tied
// creatures are offered.
func TestBolsterTieIsTheControllersChoice(t *testing.T) {
	g, me, opp := flipMain(t)
	a := pushVanillaCreature(g, me.ID, "Tie A", 1, 2)
	b := pushVanillaCreature(g, me.ID, "Tie B", 5, 2)
	other := pushVanillaCreature(g, me.ID, "Tall", 1, 5)
	pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)

	testBolsterChosen = uuid.Nil
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID}
		ctx := NewContext(g, item)
		if err := (Bolster{N: 1, ThenTail: ctx.Tail(testBolsterHeard)}).Apply(ctx); err != nil {
			t.Fatal(err)
		}
	})
	if testBolsterChosen != uuid.Nil {
		t.Fatal("the tail ran before the tie was broken")
	}
	prompt := chooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no choose-cards prompt for a tie")
	}
	if len(prompt.ChooseCards) != 2 {
		t.Errorf("offered %d creatures, want only the two tied", len(prompt.ChooseCards))
	}
	answerChooseCards(t, g, me.ID, b)

	if testBolsterChosen != b {
		t.Errorf("the tail was told %v, want the picked creature", testBolsterChosen)
	}
	if bolsterCounters(t, g, b) != 1 || bolsterCounters(t, g, a) != 0 || bolsterCounters(t, g, other) != 0 {
		t.Errorf("counters a=%d b=%d tall=%d, want only the picked tie to have one",
			bolsterCounters(t, g, a), bolsterCounters(t, g, b), bolsterCounters(t, g, other))
	}
}

// Without a creature there is nothing to choose: no counters, no
// prompt, and the clauses that follow ("the chosen creature …") don't
// run.
func TestBolsterWithNoCreatureDoesNothing(t *testing.T) {
	g, me, opp := flipMain(t)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)

	chosen := runBolster(t, g, me, 3)

	if chosen != uuid.Nil {
		t.Errorf("chose %v with no creature", chosen)
	}
	if bolsterCounters(t, g, theirs) != 0 {
		t.Error("an opponent's creature was bolstered")
	}
	if len(g.PendingChoices) != 0 {
		t.Error("a prompt was queued with nothing to choose")
	}
}

// --- Optimus Prime, Hero // Optimus Prime, Autobot Leader ------------

func optimusPrimeRow() cards.Card {
	return transformerRow(optimusPrimeOracleID,
		disturbFace{"Optimus Prime, Hero", "Legendary Artifact Creature — Robot", "{3}{U}{R}{W}", "4", "8"},
		disturbFace{"Optimus Prime, Autobot Leader", "Legendary Artifact — Vehicle", "", "6", "8"},
		"More Than Meets the Eye {2}{U}{R}{W}\nAt the beginning of each end step, bolster 1.\nWhen Optimus Prime dies, return it to the battlefield converted under its owner's control.",
		"Living metal\nTrample\nWhenever you attack, bolster 2. The chosen creature gains trample until end of turn. When that creature deals combat damage to a player this turn, convert Optimus Prime.",
		"Living metal", "Convert", "Trample")
}

// At the beginning of each end step, bolster 1: the least toughness
// among its controller's creatures gets the counter, on its controller's
// end step and on an opponent's.
func TestOptimusPrimeHeroBolstersAtEachEndStep(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, optimusPrimeRow(), me)
	ally := pushVanillaCreature(g, me.ID, "Ally", 1, 2)

	advanceTo(t, g, game.StepEnd)
	tfSettle(t, g)
	if got := bolsterCounters(t, g, ally); got != 1 {
		t.Fatalf("after my end step the 1/2 has %d counters, want 1", got)
	}
	if got := bolsterCounters(t, g, id); got != 0 {
		t.Errorf("Optimus Prime (4/8) was bolstered over the 1/2: %d counters", got)
	}

	// The opponent's end step bolsters too (each end step).
	mine := g.Turn.ActiveSeat
	for i := 0; i < 200 && g.Turn.ActiveSeat == mine; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	advanceTo(t, g, game.StepEnd)
	tfSettle(t, g)
	if got := bolsterCounters(t, g, ally); got != 2 {
		t.Errorf("after the opponent's end step the ally has %d counters, want 2", got)
	}
}

// Alone, Optimus Prime is the least toughness among the creatures its
// controller controls and bolsters itself.
func TestOptimusPrimeHeroBolstersItselfWhenAlone(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, optimusPrimeRow(), me)
	advanceTo(t, g, game.StepEnd)
	tfSettle(t, g)
	if got := bolsterCounters(t, g, id); got != 1 {
		t.Errorf("Optimus Prime has %d counters, want 1", got)
	}
}

// "When Optimus Prime dies, return it to the battlefield converted"
// (CR 712.14a): a new object, back face up, untapped.
func TestOptimusPrimeHeroDiesAndReturnsConverted(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, optimusPrimeRow(), me)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	tfSettle(t, g)
	c := wantFace(t, g, id, 1, "Optimus Prime, Autobot Leader")
	if c.Tapped {
		t.Error("it returned tapped")
	}
	if c.TurnOvers != 0 {
		t.Errorf("turn-overs = %d, want 0: it is a new object, not a convert", c.TurnOvers)
	}
}

// Whenever you attack, bolster 2; the chosen creature gains trample and
// converts Optimus Prime when it connects with a player.
func TestOptimusPrimeAutobotLeaderBolstersTramplesAndConverts(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerConverted(t, g, optimusPrimeRow(), me)
	ally := pushVanillaCreature(g, me.ID, "Ally", 2, 2)
	unsick(g, ally)
	bystander := pushVanillaCreature(g, me.ID, "Bystander", 1, 6)
	unsick(g, bystander)

	// One declaration of two attackers is one trigger (CR 603.2c).
	transformerAttacks(t, g, id, opp.ID, ally)

	if got := bolsterCounters(t, g, ally); got != 2 {
		t.Fatalf("the 2/2 has %d counters, want 2 from bolster 2", got)
	}
	if bolsterCounters(t, g, id) != 0 || bolsterCounters(t, g, bystander) != 0 {
		t.Error("bolster 2 put counters on a creature other than the least toughness (once)")
	}
	c := flipState(t, g, ally)
	if !game.HasKeyword(&c, "trample") {
		t.Error("the chosen creature did not gain trample")
	}
	if other := flipState(t, g, bystander); game.HasKeyword(&other, "trample") {
		t.Error("a creature that was not chosen gained trample")
	}
	wantFace(t, g, id, 1, "Optimus Prime, Autobot Leader")

	// Another creature connecting does not convert it.
	emitCombatDamage(g, id, opp.ID, 6)
	tfSettle(t, g)
	wantFace(t, g, id, 1, "Optimus Prime, Autobot Leader")

	// The chosen creature connecting does (once).
	emitCombatDamage(g, ally, opp.ID, 4)
	tfSettle(t, g)
	c = wantFace(t, g, id, 0, "Optimus Prime, Hero")
	if c.TurnOvers != 2 {
		t.Errorf("turn-overs = %d, want 2 (converted once to set up, once by the trigger)", c.TurnOvers)
	}
}

// CR 701.28e: a Prime that turned over some other way before the chosen
// creature connects ignores the delayed convert.
func TestOptimusPrimeAutobotLeaderDelayedConvertIsIgnoredAfterAnotherTurnOver(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerConverted(t, g, optimusPrimeRow(), me)
	ally := pushVanillaCreature(g, me.ID, "Ally", 2, 2)
	unsick(g, ally)
	transformerAttacks(t, g, id, opp.ID, ally)

	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	wantFace(t, g, id, 0, "Optimus Prime, Hero")
	emitCombatDamage(g, ally, opp.ID, 4)
	tfSettle(t, g)
	if c := wantFace(t, g, id, 0, "Optimus Prime, Hero"); c.TurnOvers != 2 {
		t.Errorf("turn-overs = %d, want 2: the delayed convert should have been ignored", c.TurnOvers)
	}
}
