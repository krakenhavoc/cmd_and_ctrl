package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Slithermuse ---------------------------------------------------

// TestSlithermuseDrawsOffTheChosenOpponent — the controller picks,
// and the draw is measured against THAT opponent rather than against
// whichever one holds the most cards. That difference is the whole
// conversion: the auto-pick always drew the maximum.
func TestSlithermuseDrawsOffTheChosenOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, small, big := g.Seats[0], g.Seats[1], g.Seats[2]
	muse := effectstest.PushCatalogPermanent(g, me.ID, "Slithermuse",
		"Creature — Elemental", effectstest.SlithermuseOracle, false)
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
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatalf("the trigger asks the controller to choose: %+v", g.PendingChoices)
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, small)

	if len(me.Hand.Cards) != want {
		t.Errorf("hand %d, want %d — the difference against the CHOSEN opponent", len(me.Hand.Cards), want)
	}
}

// TestSlithermuseDrawsNothingOffAnOpponentWithFewerCards — "if that
// player has more cards in hand than you" is a real condition, and a
// bad choice is a legal one.
func TestSlithermuseDrawsNothingOffAnOpponentWithFewerCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, poor := g.Seats[0], g.Seats[1]
	muse := effectstest.PushCatalogPermanent(g, me.ID, "Slithermuse",
		"Creature — Elemental", effectstest.SlithermuseOracle, false)
	for len(poor.Hand.Cards) > 0 {
		poor.Hand.Cards = poor.Hand.Cards[:len(poor.Hand.Cards)-1]
	}
	before := len(me.Hand.Cards)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(muse) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, poor)

	if len(me.Hand.Cards) != before {
		t.Errorf("hand %d, want %d (no difference to draw)", len(me.Hand.Cards), before)
	}
}

// --- Gluntch, the Bestower -----------------------------------------

// TestGluntchBestowsOnThreeDifferentPlayers is the whole card: three
// player choices that must name three different seats, the chosen
// player's own pick of a creature, and all three payloads landing.
func TestGluntchBestowsOnThreeDifferentPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, second, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Gluntch, the Bestower",
		"Legendary Creature — Jellyfish", effectstest.GluntchOracle, false)
	bear := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	secondHand := len(second.Hand.Cards)

	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	// First: the controller chooses themselves, then picks which of
	// their creatures takes the counters.
	effectstest.AnswerChoosePlayer(t, g, me.ID, me)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the CHOSEN player picks the creature: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}

	// Second: a different player draws. The first is not offered.
	sec := effectstest.LatestOptionPickFor(g, me.ID)
	if sec == nil {
		t.Fatalf("the second choice is queued: %+v", g.PendingChoices)
	}
	for _, opt := range sec.PickOptions {
		if opt.Label == me.Name {
			t.Errorf("a SECOND player excludes the first: %+v", sec.PickOptions)
		}
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, second)
	if len(second.Hand.Cards) != secondHand+1 {
		t.Errorf("the second player drew: hand %d, want %d", len(second.Hand.Cards), secondHand+1)
	}

	// Third: a different player again, two Treasures.
	thirdPrompt := effectstest.LatestOptionPickFor(g, me.ID)
	if thirdPrompt == nil {
		t.Fatalf("the third choice is queued: %+v", g.PendingChoices)
	}
	if len(thirdPrompt.PickOptions) != 2 {
		t.Errorf("two seats left of four: %+v", thirdPrompt.PickOptions)
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, third)
	if got := effectstest.CountTreasures(g, third.ID); got != 2 {
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
	g := effectstest.NewCatalogGame(t)
	me, empty := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Gluntch, the Bestower",
		"Legendary Creature — Jellyfish", effectstest.GluntchOracle, false)

	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, empty)

	if effectstest.LatestChooseCardsFor(g, empty.ID) != nil {
		t.Errorf("a player with no creature is not prompted: %+v", g.PendingChoices)
	}
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatalf("the trigger carries on to the second clause: %+v", g.PendingChoices)
	}
}

// TestGluntchCarriesOnWhenItsCounterPickIsWithdrawn — #1225, and the
// drop-time twin of the test above.
//
// The chosen player's creature leaves the battlefield while "put two
// +1/+1 counters on a creature you control" is in front of them, so
// the prompt has no legal answer left and is withdrawn (#1045). The
// drop now runs that pick's continuation with nothing picked — which
// is the SAME closure the test above reaches at queue time, with the
// same empty answer — so one printed sentence has one behaviour
// whether the board was empty when the question was asked or emptied
// while it was open. Before #1225 the trigger stopped here and the
// second and third players were never chosen.
func TestGluntchCarriesOnWhenItsCounterPickIsWithdrawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, chosen := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Gluntch, the Bestower",
		"Legendary Creature — Jellyfish", effectstest.GluntchOracle, false)
	bear := effectstest.PushTribalCreature(g, chosen.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, chosen)

	if effectstest.LatestChooseCardsFor(g, chosen.ID) == nil {
		t.Fatalf("the chosen player is asked which creature: %+v", g.PendingChoices)
	}

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bear); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})

	if c := effectstest.LatestChooseCardsFor(g, chosen.ID); c != nil {
		t.Fatalf("the pick survives a board with none of its candidates on it: %+v", c)
	}
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatalf("the trigger stopped at the withdrawn pick — the second and third clauses "+
			"went with it (#1225): %+v", g.PendingChoices)
	}
}

// --- Skullwinder ---------------------------------------------------

// TestSkullwinderSharesTheRegrowthWithTheChosenOpponent — two prompts
// to two different seats out of one trigger.
func TestSkullwinderSharesTheRegrowthWithTheChosenOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.PushGraveyardCard(g, me, "My Bolt")
	theirs := effectstest.PushGraveyardCard(g, opp, "Their Bolt")
	effectstest.PushGraveyardCard(g, other, "Someone Else's Bolt")

	effectstest.CastCatalogSpell(t, g, "Skullwinder", "Creature — Snake", effectstest.SkullwinderOracle, nil)
	// The trigger's target is chosen as it goes on the stack
	// (CR 603.3d), before anything resolves.
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Hand.Contains(mine) {
		t.Errorf("the targeted card came back to the controller's hand")
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)

	pick := effectstest.LatestChooseCardsFor(g, opp.ID)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushGraveyardCard(g, me, "My Bolt")

	effectstest.CastCatalogSpell(t, g, "Skullwinder", "Creature — Snake", effectstest.SkullwinderOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)

	if len(g.PendingChoices) != 0 {
		t.Errorf("an empty graveyard is not prompted over: %+v", g.PendingChoices)
	}
}
