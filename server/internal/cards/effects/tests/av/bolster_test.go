package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bolster_test.go — bolster (CR 701.39, #3277): the primitive on plain
// creatures, then Optimus Prime, Hero // Optimus Prime, Autobot Leader.

// testBolsterChosen is the creature the bolster's keyed tail
// (testBolsterHeard) was told; Nil when it never ran.
var (
	testBolsterChosen uuid.UUID
	testBolsterHeard  = effects.CardTail("test/4253-bolster-heard", func(_ *effects.Context, o game.TailOutcome) error {
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
		ctx := effects.NewContext(g, item)
		if err := (effects.Bolster{N: n, ThenTail: ctx.Tail(testBolsterHeard)}).Apply(ctx); err != nil {
			t.Fatalf("Bolster: %v", err)
		}
	})
	return testBolsterChosen
}

// CR 701.39a: the creature with the least toughness gets the counters,
// and nobody else.
func TestBolsterPutsCountersOnTheLeastToughness(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	big := effectstest.PushVanillaCreature(g, me.ID, "Big", 4, 4)
	small := effectstest.PushVanillaCreature(g, me.ID, "Small", 3, 1)
	mid := effectstest.PushVanillaCreature(g, me.ID, "Mid", 2, 3)

	chosen := runBolster(t, g, me, 2)

	if chosen != small {
		t.Fatalf("chose %v, want the 3/1", chosen)
	}
	if got := effectstest.BolsterCounters(t, g, small); got != 2 {
		t.Errorf("the 3/1 has %d +1/+1 counters, want 2", got)
	}
	if effectstest.BolsterCounters(t, g, big) != 0 || effectstest.BolsterCounters(t, g, mid) != 0 {
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
	g, me, opp := effectstest.FlipMain(t)
	grown := effectstest.PushVanillaCreature(g, me.ID, "Grown", 1, 1)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain", 1, 3)
	effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(grown, game.CounterPlusOne, 3); err != nil {
			t.Fatal(err)
		}
	})

	chosen := runBolster(t, g, me, 1)

	if chosen != plain {
		t.Fatalf("chose %v, want the 1/3 (the 1/1 is a 4/4 with its counters)", chosen)
	}
	if effectstest.BolsterCounters(t, g, plain) != 1 {
		t.Error("the 1/3 has no counter")
	}
}

// CR 701.39a: tied for least, the controller picks, and only the tied
// creatures are offered.
func TestBolsterTieIsTheControllersChoice(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	a := effectstest.PushVanillaCreature(g, me.ID, "Tie A", 1, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Tie B", 5, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "Tall", 1, 5)
	effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)

	testBolsterChosen = uuid.Nil
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID}
		ctx := effects.NewContext(g, item)
		if err := (effects.Bolster{N: 1, ThenTail: ctx.Tail(testBolsterHeard)}).Apply(ctx); err != nil {
			t.Fatal(err)
		}
	})
	if testBolsterChosen != uuid.Nil {
		t.Fatal("the tail ran before the tie was broken")
	}
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no choose-cards prompt for a tie")
	}
	if len(prompt.ChooseCards) != 2 {
		t.Errorf("offered %d creatures, want only the two tied", len(prompt.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, b)

	if testBolsterChosen != b {
		t.Errorf("the tail was told %v, want the picked creature", testBolsterChosen)
	}
	if effectstest.BolsterCounters(t, g, b) != 1 || effectstest.BolsterCounters(t, g, a) != 0 || effectstest.BolsterCounters(t, g, other) != 0 {
		t.Errorf("counters a=%d b=%d tall=%d, want only the picked tie to have one",
			effectstest.BolsterCounters(t, g, a), effectstest.BolsterCounters(t, g, b), effectstest.BolsterCounters(t, g, other))
	}
}

// Without a creature there is nothing to choose: no counters, no
// prompt, and the clauses that follow ("the chosen creature …") don't
// run.
func TestBolsterWithNoCreatureDoesNothing(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)

	chosen := runBolster(t, g, me, 3)

	if chosen != uuid.Nil {
		t.Errorf("chose %v with no creature", chosen)
	}
	if effectstest.BolsterCounters(t, g, theirs) != 0 {
		t.Error("an opponent's creature was bolstered")
	}
	if len(g.PendingChoices) != 0 {
		t.Error("a prompt was queued with nothing to choose")
	}
}

// --- Optimus Prime, Hero // Optimus Prime, Autobot Leader ------------

func optimusPrimeRow() cards.Card {
	return effectstest.TransformerRow(cardso.OptimusPrimeOracleID,
		effectstest.DisturbFace{Name: "Optimus Prime, Hero", TypeLine: "Legendary Artifact Creature — Robot", Cost: "{3}{U}{R}{W}", Power: "4", Toughness: "8"},
		effectstest.DisturbFace{Name: "Optimus Prime, Autobot Leader", TypeLine: "Legendary Artifact — Vehicle", Cost: "", Power: "6", Toughness: "8"},
		"More Than Meets the Eye {2}{U}{R}{W}\nAt the beginning of each end step, bolster 1.\nWhen Optimus Prime dies, return it to the battlefield converted under its owner's control.",
		"Living metal\nTrample\nWhenever you attack, bolster 2. The chosen creature gains trample until end of turn. When that creature deals combat damage to a player this turn, convert Optimus Prime.",
		"Living metal", "Convert", "Trample")
}

// At the beginning of each end step, bolster 1: the least toughness
// among its controller's creatures gets the counter, on its controller's
// end step and on an opponent's.
func TestOptimusPrimeHeroBolstersAtEachEndStep(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, optimusPrimeRow(), me)
	ally := effectstest.PushVanillaCreature(g, me.ID, "Ally", 1, 2)

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.TfSettle(t, g)
	if got := effectstest.BolsterCounters(t, g, ally); got != 1 {
		t.Fatalf("after my end step the 1/2 has %d counters, want 1", got)
	}
	if got := effectstest.BolsterCounters(t, g, id); got != 0 {
		t.Errorf("Optimus Prime (4/8) was bolstered over the 1/2: %d counters", got)
	}

	// The opponent's end step bolsters too (each end step).
	mine := g.Turn.ActiveSeat
	for i := 0; i < 200 && g.Turn.ActiveSeat == mine; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.TfSettle(t, g)
	if got := effectstest.BolsterCounters(t, g, ally); got != 2 {
		t.Errorf("after the opponent's end step the ally has %d counters, want 2", got)
	}
}

// Alone, Optimus Prime is the least toughness among the creatures its
// controller controls and bolsters itself.
func TestOptimusPrimeHeroBolstersItselfWhenAlone(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, optimusPrimeRow(), me)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.TfSettle(t, g)
	if got := effectstest.BolsterCounters(t, g, id); got != 1 {
		t.Errorf("Optimus Prime has %d counters, want 1", got)
	}
}

// "When Optimus Prime dies, return it to the battlefield converted"
// (CR 712.14a): a new object, back face up, untapped.
func TestOptimusPrimeHeroDiesAndReturnsConverted(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, optimusPrimeRow(), me)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.TfSettle(t, g)
	c := effectstest.WantFace(t, g, id, 1, "Optimus Prime, Autobot Leader")
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
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerConverted(t, g, optimusPrimeRow(), me)
	ally := effectstest.PushVanillaCreature(g, me.ID, "Ally", 2, 2)
	effectstest.Unsick(g, ally)
	bystander := effectstest.PushVanillaCreature(g, me.ID, "Bystander", 1, 6)
	effectstest.Unsick(g, bystander)

	// One declaration of two attackers is one trigger (CR 603.2c).
	effectstest.TransformerAttacks(t, g, id, opp.ID, ally)

	if got := effectstest.BolsterCounters(t, g, ally); got != 2 {
		t.Fatalf("the 2/2 has %d counters, want 2 from bolster 2", got)
	}
	if effectstest.BolsterCounters(t, g, id) != 0 || effectstest.BolsterCounters(t, g, bystander) != 0 {
		t.Error("bolster 2 put counters on a creature other than the least toughness (once)")
	}
	c := effectstest.FlipState(t, g, ally)
	if !game.HasKeyword(&c, "trample") {
		t.Error("the chosen creature did not gain trample")
	}
	if other := effectstest.FlipState(t, g, bystander); game.HasKeyword(&other, "trample") {
		t.Error("a creature that was not chosen gained trample")
	}
	effectstest.WantFace(t, g, id, 1, "Optimus Prime, Autobot Leader")

	// Another creature connecting does not convert it.
	effectstest.EmitCombatDamage(g, id, opp.ID, 6)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 1, "Optimus Prime, Autobot Leader")

	// The chosen creature connecting does (once).
	effectstest.EmitCombatDamage(g, ally, opp.ID, 4)
	effectstest.TfSettle(t, g)
	c = effectstest.WantFace(t, g, id, 0, "Optimus Prime, Hero")
	if c.TurnOvers != 2 {
		t.Errorf("turn-overs = %d, want 2 (converted once to set up, once by the trigger)", c.TurnOvers)
	}
}

// CR 701.28e: a Prime that turned over some other way before the chosen
// creature connects ignores the delayed convert.
func TestOptimusPrimeAutobotLeaderDelayedConvertIsIgnoredAfterAnotherTurnOver(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerConverted(t, g, optimusPrimeRow(), me)
	ally := effectstest.PushVanillaCreature(g, me.ID, "Ally", 2, 2)
	effectstest.Unsick(g, ally)
	effectstest.TransformerAttacks(t, g, id, opp.ID, ally)

	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.WantFace(t, g, id, 0, "Optimus Prime, Hero")
	effectstest.EmitCombatDamage(g, ally, opp.ID, 4)
	effectstest.TfSettle(t, g)
	if c := effectstest.WantFace(t, g, id, 0, "Optimus Prime, Hero"); c.TurnOvers != 2 {
		t.Errorf("turn-overs = %d, want 2: the delayed convert should have been ignored", c.TurnOvers)
	}
}
