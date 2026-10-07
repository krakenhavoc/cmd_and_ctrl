package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// proliferate_choice_test.go — #2525: proliferate is the player's
// choice (CR 701.34a), not an auto-pick.

// castSteadyProgressToPrompt casts Steady Progress and passes until
// its proliferate prompt is open, leaving the draw undrawn.
func castSteadyProgressToPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	castCatalogSpell(t, g, "Steady Progress", "Instant", steadyProgressOracle, nil)
	passPriorityAroundTable(t, g)
	c := proliferatePrompt(g)
	if c == nil {
		t.Fatal("Steady Progress did not ask what to proliferate")
	}
	return c
}

func idSet(ids []uuid.UUID) map[uuid.UUID]bool {
	out := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out
}

// CR 701.34a: the prompt offers every permanent and player that has a
// counter, nothing that has none, and suggests the beneficial pick. The
// draw after "Proliferate." waits for the answer.
func TestProliferateOffersEverythingWithCountersAndSuggestsThePick(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)
	myShrinking := pushCounterCreature(g, me.ID, "Mine Shrinking", game.CounterMinusOne, 1)
	theirs := pushCounterCreature(g, opp.ID, "Theirs", game.CounterMinusOne, 1)
	bare := pushCounterCreature(g, opp.ID, "Bare", "", 0)
	if err := g.AddPlayerCounter(opp.ID, game.CounterPoison, 1); err != nil {
		t.Fatal(err)
	}
	handBefore := len(me.Hand.Cards)

	c := castSteadyProgressToPrompt(t, g)

	if c.Chooser != me.ID {
		t.Errorf("the prompt is addressed to %v, want the proliferating player", c.Chooser)
	}
	offered := idSet(c.ChooseCards)
	for name, id := range map[string]uuid.UUID{"mine": mine, "mine shrinking": myShrinking, "theirs": theirs} {
		if !offered[id] {
			t.Errorf("%s has a counter and was not offered", name)
		}
	}
	if offered[bare] {
		t.Error("a permanent with no counters was offered (CR 701.34a)")
	}
	if len(c.ChoosePlayers) != 1 || c.ChoosePlayers[0] != opp.ID {
		t.Errorf("seats offered = %v, want just the player with poison", c.ChoosePlayers)
	}
	suggested := idSet(c.ChooseSuggested)
	if len(suggested) != 3 || !suggested[mine] || !suggested[theirs] || !suggested[opp.ID] {
		t.Errorf("suggested = %v, want my growing creature, their shrinking one and their poison", c.ChooseSuggested)
	}
	if c.ChooseMin != 0 {
		t.Errorf("floor = %d, want 0: \"any number\" includes none", c.ChooseMin)
	}
	if got := len(me.Hand.Cards); got != handBefore {
		t.Errorf("the card was drawn while the prompt was open: hand %d -> %d", handBefore, got)
	}
}

// The point of the issue: choosing nothing is a legal proliferate, no
// counter is added, and the rest of the card still happens.
func TestProliferateAllowsChoosingNothingAndStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)
	handBefore := len(me.Hand.Cards)

	castSteadyProgressToPrompt(t, g)
	answerProliferate(t, g)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("a counter was added to a permanent that was not chosen: %d", got)
	}
	if proliferatePrompt(g) != nil {
		t.Error("the prompt is still open after it was answered")
	}
	// The spell leaves the hand when it is cast, then the card is drawn.
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand = %d, want %d (cast the spell, drew a card)", got, handBefore+1)
	}
}

// What the old auto-pick could not do: take a counter the engine would
// not suggest. "Put another -1/-1 counter on my own persist creature."
func TestProliferateLetsThePlayerChooseAgainstTheSuggestion(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)
	myShrinking := pushCounterCreature(g, me.ID, "Mine Shrinking", game.CounterMinusOne, 1)
	theirs := pushCounterCreature(g, opp.ID, "Theirs", game.CounterMinusOne, 1)

	castSteadyProgressToPrompt(t, g)
	answerProliferate(t, g, myShrinking)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, myShrinking, game.CounterMinusOne); got != 2 {
		t.Errorf("the chosen -1/-1 creature has %d counters, want 2", got)
	}
	if got := counterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("an unchosen +1/+1 creature gained a counter: %d", got)
	}
	if got := counterCount(g, theirs, game.CounterMinusOne); got != 1 {
		t.Errorf("an unchosen opposing creature gained a counter: %d", got)
	}
}

// Seats are named in the same list as permanents, by player ID, and a
// seat with no counters is not a legal pick.
func TestProliferateNamesPlayersAndRefusesAnythingNotOffered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bare := pushCounterCreature(g, opp.ID, "Bare", "", 0)
	if err := g.AddPlayerCounter(opp.ID, game.CounterPoison, 2); err != nil {
		t.Fatal(err)
	}
	pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)

	c := castSteadyProgressToPrompt(t, g)

	for name, picks := range map[string][]uuid.UUID{
		"a permanent with no counters":  {bare},
		"a seat with no counters":       {me.ID},
		"a duplicate":                   {opp.ID, opp.ID},
		"something that does not exist": {uuid.New()},
	} {
		if err := g.ResolveProliferate(c.ID, c.Chooser, picks); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
	if proliferatePrompt(g) == nil {
		t.Fatal("a refused answer closed the prompt")
	}
	if err := g.ResolveProliferate(c.ID, opp.ID, nil); err == nil {
		t.Error("a seat that is not the chooser answered")
	}

	answerProliferate(t, g, opp.ID)
	if got := opp.Counters[game.CounterPoison]; got != 3 {
		t.Errorf("poison = %d, want 3", got)
	}
	if got := opp.Poison; got != 3 {
		t.Errorf("legacy Player.Poison = %d, want it mirrored to 3", got)
	}
}

// "Proliferate twice" is two choices, each made over the board the
// previous answer left (CR 701.34 with Tekuthal's replacement), and the
// draw waits for the second.
func TestTekuthalProliferatesTwiceIsTwoChoices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTekuthal(g, me.ID)
	a := pushCounterCreature(g, me.ID, "A", game.CounterPlusOne, 1)
	b := pushCounterCreature(g, me.ID, "B", game.CounterPlusOne, 1)
	handBefore := len(me.Hand.Cards)

	first := castSteadyProgressToPrompt(t, g)
	answerProliferate(t, g, a)
	second := proliferatePrompt(g)
	if second == nil {
		t.Fatal("the second proliferate of \"proliferate twice\" asked nothing")
	}
	if second.ID == first.ID {
		t.Error("the second prompt is the first one again")
	}
	if got := len(me.Hand.Cards); got != handBefore {
		t.Errorf("the card was drawn between the two proliferates: hand %d -> %d", handBefore, got)
	}
	answerProliferate(t, g, b)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, a, game.CounterPlusOne); got != 2 {
		t.Errorf("A = %d counters, want 2 (chosen once)", got)
	}
	if got := counterCount(g, b, game.CounterPlusOne); got != 2 {
		t.Errorf("B = %d counters, want 2 (chosen once)", got)
	}
	if proliferatePrompt(g) != nil {
		t.Error("a third prompt opened")
	}
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand = %d, want %d (the draw after both)", got, handBefore+1)
	}
}

// A seat that concedes with a proliferate open gives nothing, and a
// "proliferate twice" does not ask the departed seat a second time.
func TestProliferatePromptIsDroppedWhenItsChooserConcedes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTekuthal(g, me.ID)
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)

	castSteadyProgressToPrompt(t, g)
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if proliferatePrompt(g) != nil {
		t.Error("a proliferate prompt addressed to a departed seat is still open")
	}
	if got := counterCount(g, mine, game.CounterPlusOne); got > 1 {
		t.Errorf("counters were given on behalf of a seat that left: %d", got)
	}
}
