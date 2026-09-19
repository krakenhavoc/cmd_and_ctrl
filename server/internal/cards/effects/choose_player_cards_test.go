package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_player_cards_test.go — #929, the three cards the
// resolution-time player choice shipped on.
//
// Slithermuse is the conversion (it auto-picked before), Gluntch is
// the chain of three exclusive choices with a sub-choice made by
// somebody who is not the controller, and Skullwinder is the
// cross-seat pair: the controller names an opponent and that opponent
// then picks a card.

const (
	gluntchOracle     = "0222dc7c-459b-4909-a037-72b2eb248599"
	skullwinderOracle = "d3d8dd6e-a80b-4063-a542-ff67ef8dad4f"
)

// answerChoosePlayer answers `chooser`'s open player prompt by NAMING
// the seat, so a test says who it picked rather than guessing an
// index out of the prompt's bot ordering.
func answerChoosePlayer(t *testing.T, g *game.Game, chooser uuid.UUID, pick *game.Player) {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no player prompt for %s: %+v", chooser, g.PendingChoices)
	}
	for i, opt := range c.PickOptions {
		if opt.Label != pick.Name {
			continue
		}
		if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
			t.Fatalf("ResolveOptionPick(%d): %v", i, err)
		}
		return
	}
	t.Fatalf("%q is not offered; options are %+v", pick.Name, c.PickOptions)
}

// --- Slithermuse ---------------------------------------------------

// TestSlithermuseDrawsOffTheChosenOpponent — the controller picks,
// and the draw is measured against THAT opponent rather than against
// whichever one holds the most cards. That difference is the whole
// conversion: the auto-pick always drew the maximum.
func TestSlithermuseDrawsOffTheChosenOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, small, big := g.Seats[0], g.Seats[1], g.Seats[2]
	muse := pushCatalogPermanent(g, me.ID, "Slithermuse",
		"Creature — Elemental", slithermuseOracle, false)
	for i := 0; i < 2; i++ {
		small.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "filler",
			Owner: small.ID, Controller: small.ID})
	}
	for i := 0; i < 6; i++ {
		big.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "filler",
			Owner: big.ID, Controller: big.ID})
	}
	before := len(me.Hand.Cards)
	want := before + (len(small.Hand.Cards) - before)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(muse) })
	passPriorityAroundTable(t, g)

	if latestOptionPickFor(g, me.ID) == nil {
		t.Fatalf("the trigger asks the controller to choose: %+v", g.PendingChoices)
	}
	answerChoosePlayer(t, g, me.ID, small)

	if len(me.Hand.Cards) != want {
		t.Errorf("hand %d, want %d — the difference against the CHOSEN opponent", len(me.Hand.Cards), want)
	}
}

// TestSlithermuseDrawsNothingOffAnOpponentWithFewerCards — "if that
// player has more cards in hand than you" is a real condition, and a
// bad choice is a legal one.
func TestSlithermuseDrawsNothingOffAnOpponentWithFewerCards(t *testing.T) {
	g := newCatalogGame(t)
	me, poor := g.Seats[0], g.Seats[1]
	muse := pushCatalogPermanent(g, me.ID, "Slithermuse",
		"Creature — Elemental", slithermuseOracle, false)
	for len(poor.Hand.Cards) > 0 {
		poor.Hand.Cards = poor.Hand.Cards[:len(poor.Hand.Cards)-1]
	}
	before := len(me.Hand.Cards)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(muse) })
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, poor)

	if len(me.Hand.Cards) != before {
		t.Errorf("hand %d, want %d (no difference to draw)", len(me.Hand.Cards), before)
	}
}

// --- Gluntch, the Bestower -----------------------------------------

// TestGluntchBestowsOnThreeDifferentPlayers is the whole card: three
// player choices that must name three different seats, the chosen
// player's own pick of a creature, and all three payloads landing.
func TestGluntchBestowsOnThreeDifferentPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, second, third := g.Seats[0], g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Gluntch, the Bestower",
		"Legendary Creature — Jellyfish", gluntchOracle, false)
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	secondHand := len(second.Hand.Cards)

	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)

	// First: the controller chooses themselves, then picks which of
	// their creatures takes the counters.
	answerChoosePlayer(t, g, me.ID, me)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the CHOSEN player picks the creature: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := counterCount(g, bear, "+1/+1"); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}

	// Second: a different player draws. The first is not offered.
	sec := latestOptionPickFor(g, me.ID)
	if sec == nil {
		t.Fatalf("the second choice is queued: %+v", g.PendingChoices)
	}
	for _, opt := range sec.PickOptions {
		if opt.Label == me.Name {
			t.Errorf("a SECOND player excludes the first: %+v", sec.PickOptions)
		}
	}
	answerChoosePlayer(t, g, me.ID, second)
	if len(second.Hand.Cards) != secondHand+1 {
		t.Errorf("the second player drew: hand %d, want %d", len(second.Hand.Cards), secondHand+1)
	}

	// Third: a different player again, two Treasures.
	thirdPrompt := latestOptionPickFor(g, me.ID)
	if thirdPrompt == nil {
		t.Fatalf("the third choice is queued: %+v", g.PendingChoices)
	}
	if len(thirdPrompt.PickOptions) != 2 {
		t.Errorf("two seats left of four: %+v", thirdPrompt.PickOptions)
	}
	answerChoosePlayer(t, g, me.ID, third)
	if got := countTreasures(g, third.ID); got != 2 {
		t.Errorf("Treasures for the third player: %d, want 2", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("the chain is finished: %+v", g.PendingChoices)
	}
}

// TestGluntchSkipsTheCounterPickWhenThatPlayerHasNoCreature — the
// clause does as much as it can (CR 608.2) and the rest of the
// trigger still resolves.
func TestGluntchSkipsTheCounterPickWhenThatPlayerHasNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, empty := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Gluntch, the Bestower",
		"Legendary Creature — Jellyfish", gluntchOracle, false)

	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, empty)

	if latestChooseCardsFor(g, empty.ID) != nil {
		t.Errorf("a player with no creature is not prompted: %+v", g.PendingChoices)
	}
	if latestOptionPickFor(g, me.ID) == nil {
		t.Fatalf("the trigger carries on to the second clause: %+v", g.PendingChoices)
	}
}

// --- Skullwinder ---------------------------------------------------

// TestSkullwinderSharesTheRegrowthWithTheChosenOpponent — two prompts
// to two different seats out of one trigger.
func TestSkullwinderSharesTheRegrowthWithTheChosenOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := pushGraveyardCard(g, me, "My Bolt")
	theirs := pushGraveyardCard(g, opp, "Their Bolt")
	pushGraveyardCard(g, other, "Someone Else's Bolt")

	castCatalogSpell(t, g, "Skullwinder", "Creature — Snake", skullwinderOracle, nil)
	// The trigger's target is chosen as it goes on the stack
	// (CR 603.3d), before anything resolves.
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(mine) {
		t.Errorf("the targeted card came back to the controller's hand")
	}
	answerChoosePlayer(t, g, me.ID, opp)

	pick := latestChooseCardsFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the CHOSEN opponent picks their own card: %+v", g.PendingChoices)
	}
	if pick.FromPlayer != opp.ID {
		t.Errorf("the pool is the opponent's own graveyard: FromPlayer %s", pick.FromPlayer)
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != theirs {
		t.Errorf("their graveyard only: %v", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !opp.Hand.Contains(theirs) {
		t.Errorf("the opponent's card came back to their hand")
	}
}

// TestSkullwinderAsksNobodyWhenTheOpponentsGraveyardIsEmpty — the
// second half does as much as it can and asks no unanswerable
// question.
func TestSkullwinderAsksNobodyWhenTheOpponentsGraveyardIsEmpty(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushGraveyardCard(g, me, "My Bolt")

	castCatalogSpell(t, g, "Skullwinder", "Creature — Snake", skullwinderOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)

	if len(g.PendingChoices) != 0 {
		t.Errorf("an empty graveyard is not prompted over: %+v", g.PendingChoices)
	}
}

// --- helpers -------------------------------------------------------

// pushGraveyardCard puts a card straight into a player's graveyard.
func pushGraveyardCard(g *game.Game, p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant",
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// advanceToEndStep walks the cursor to the active player's end step,
// where Gluntch's trigger fires.
func advanceToEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; g.Turn.Step != game.StepEnd; i++ {
		if i >= 12 {
			t.Fatalf("never reached the end step (stuck at %v)", g.Turn.Step)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}
