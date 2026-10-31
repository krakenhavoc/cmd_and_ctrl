package effects

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// klaw_sonic_subjugator_test.go — #4588: "target player reveals a number of
// cards from their hand equal to one plus the number of creature cards in
// your graveyard. You choose one of them. That player discards that card."

const klawOracle = "1c233e08-d5a3-48d7-9074-20393896a480"

// klawSetUp seats Klaw's controller with `dead` creature cards (and one
// noncreature card, which must not count) in their graveyard and an
// opponent holding `hand` cards named "Their Card 1" .. "Their Card N".
func klawSetUp(t *testing.T, dead, hand int) (*game.Game, *game.Player, *game.Player, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	for i := 0; i < dead; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: fmt.Sprintf("Dead Bear %d", i), TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Rite", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	var ids []uuid.UUID
	for i := 1; i <= hand; i++ {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: fmt.Sprintf("Their Card %d", i), TypeLine: "Instant", ManaCost: "{U}", Owner: opp.ID, Controller: opp.ID})
		ids = append(ids, id)
	}
	return g, me, opp, ids
}

// klawCast casts Klaw, lets it enter and targets opp with the trigger.
func klawCast(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	klaw := tdCastFromHand(t, g, game.Card{Name: "Klaw, Sonic Subjugator", TypeLine: "Legendary Creature — Human Rogue Villain",
		OracleID: klawOracle, Power: 2, Toughness: 2}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	return klaw
}

func revealPickFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceRevealPick && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// klawRevealed answers the target's "reveal N cards" prompt and returns the
// controller's reveal pick.
func klawRevealed(t *testing.T, g *game.Game, me, opp *game.Player, n int, reveal ...uuid.UUID) *game.PendingChoice {
	t.Helper()
	ask := chooseCardsChoiceFor(g, opp.ID)
	if ask == nil || ask.ChooseMin != n || ask.ChooseMax != n {
		t.Fatalf("the target chooses exactly %d cards to reveal, got %+v", n, ask)
	}
	if len(ask.ChooseCards) != len(opp.Hand.Cards) {
		t.Fatalf("the target chooses from their whole hand (%d), got %d candidates", len(opp.Hand.Cards), len(ask.ChooseCards))
	}
	answerChooseCards(t, g, opp.ID, reveal...)
	rp := revealPickFor(g, me.ID)
	if rp == nil {
		t.Fatal("no reveal pick for Klaw's controller")
	}
	return rp
}

// The count is one plus the creature cards in the controller's graveyard,
// read as the trigger resolves; the target picks which cards to show.
func TestKlawTargetRevealsOnePlusCreatureCardsOfTheirChoice(t *testing.T) {
	for dead := 0; dead <= 2; dead++ {
		t.Run(fmt.Sprintf("dead=%d", dead), func(t *testing.T) {
			g, me, opp, hand := klawSetUp(t, dead, 5)
			klawCast(t, g, me, opp)
			n := 1 + dead
			// The target chooses the LAST n cards of the hand, not the first.
			reveal := hand[len(hand)-n:]
			rp := klawRevealed(t, g, me, opp, n, reveal...)
			if len(rp.ChooseCards) != n {
				t.Fatalf("revealed %d cards, want %d", len(rp.ChooseCards), n)
			}
			shown := map[uuid.UUID]bool{}
			for _, id := range rp.ChooseCards {
				shown[id] = true
			}
			for _, id := range reveal {
				if !shown[id] {
					t.Errorf("a card the target chose to reveal was not revealed")
				}
			}
		})
	}
}

// A hand no bigger than the count is revealed whole and nobody is asked to
// choose which cards.
func TestKlawRevealsTheWholeHandWhenItIsNoLargerThanTheCount(t *testing.T) {
	g, me, opp, hand := klawSetUp(t, 2, 2)
	klawCast(t, g, me, opp)
	if chooseCardsChoiceFor(g, opp.ID) != nil {
		t.Fatal("the target was asked which cards to reveal with N at least their hand size")
	}
	rp := revealPickFor(g, me.ID)
	if rp == nil || len(rp.ChooseCards) != len(hand) {
		t.Fatalf("want both cards revealed to Klaw's controller, got %+v", rp)
	}
}

// Only the revealed cards are shown; the controller picks one of them and
// only one of them, and the target discards it.
func TestKlawControllerPicksARevealedCardAndItIsDiscarded(t *testing.T) {
	g, me, opp, hand := klawSetUp(t, 1, 4)
	klawCast(t, g, me, opp)
	rp := klawRevealed(t, g, me, opp, 2, hand[0], hand[1])

	for i, id := range hand {
		c, _ := g.LookupCardForEffect(id)
		if known := c.IsKnownTo(me.ID); known != (i < 2) {
			t.Errorf("hand card %d known to Klaw's controller = %v, want %v", i, known, i < 2)
		}
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{hand[3]}); err == nil {
		t.Fatal("the controller picked a card that was not revealed")
	}
	if !opp.Hand.Contains(hand[3]) {
		t.Fatal("a refused pick moved a card")
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{hand[1]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if !opp.Graveyard.Contains(hand[1]) || opp.Hand.Contains(hand[1]) {
		t.Error("the chosen card was not discarded")
	}
	if !opp.Hand.Contains(hand[0]) || opp.Hand.Size() != 3 {
		t.Error("a revealed card that was not chosen left the hand")
	}
	if me.Graveyard.Contains(hand[1]) {
		t.Error("the card went to the wrong graveyard")
	}
}

// An empty hand reveals nothing and asks nothing.
func TestKlawAgainstAnEmptyHandDoesNothing(t *testing.T) {
	g, me, opp, _ := klawSetUp(t, 2, 0)
	klawCast(t, g, me, opp)
	if chooseCardsChoiceFor(g, opp.ID) != nil || revealPickFor(g, me.ID) != nil {
		t.Fatal("an empty hand raised a prompt")
	}
	if opp.Graveyard.Size() != 0 {
		t.Error("something was discarded from an empty hand")
	}
}

// Klaw can target its own controller: they reveal from their own hand and
// choose from it.
func TestKlawCanTargetItsControllerThemselves(t *testing.T) {
	g, me, _, _ := klawSetUp(t, 0, 0)
	me.Hand.Cards = nil
	var mine []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: fmt.Sprintf("My Card %d", i), TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		mine = append(mine, id)
	}
	klawCast(t, g, me, me)
	ask := chooseCardsChoiceFor(g, me.ID)
	if ask == nil || ask.ChooseMin != 1 || ask.ChooseMax != 1 {
		t.Fatalf("want a reveal-one prompt, got %+v", ask)
	}
	answerChooseCards(t, g, me.ID, mine[0])
	rp := revealPickFor(g, me.ID)
	if rp == nil {
		t.Fatal("no reveal pick")
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{mine[0]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if !me.Graveyard.Contains(mine[0]) {
		t.Error("the chosen card was not discarded")
	}
}

// Both prompts are restore points: a strict restore with the reveal
// prompt open, and again with the reveal pick open, finishes the discard.
func TestKlawBothPromptsAreRestorePoints(t *testing.T) {
	g, me, opp, hand := klawSetUp(t, 1, 4)
	klawCast(t, g, me, opp)
	if chooseCardsChoiceFor(g, opp.ID) == nil {
		t.Fatal("no reveal prompt")
	}
	if snap := g.CaptureSnapshot(); !snap.Restorable() {
		t.Fatalf("the table is not a restore point at the reveal: %+v", snap.Continuations)
	}
	g = restoreRoundTrip(t, g, true)
	me, opp = g.PlayerByIDForEffect(me.ID), g.PlayerByIDForEffect(opp.ID)
	answerChooseCards(t, g, opp.ID, hand[2], hand[3])

	if snap := g.CaptureSnapshot(); !snap.Restorable() {
		t.Fatalf("the table is not a restore point at the pick: %+v", snap.Continuations)
	}
	g = restoreRoundTrip(t, g, true)
	me, opp = g.PlayerByIDForEffect(me.ID), g.PlayerByIDForEffect(opp.ID)
	rp := revealPickFor(g, me.ID)
	if rp == nil || len(rp.ChooseCards) != 2 {
		t.Fatalf("the restored pick is %+v, want the two revealed cards", rp)
	}
	if c, _ := g.LookupCardForEffect(hand[0]); c.IsKnownTo(me.ID) {
		t.Error("the restored table shows an unrevealed card")
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{hand[2]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if !opp.Graveyard.Contains(hand[2]) {
		t.Error("the restored pick did not discard the chosen card")
	}
}
