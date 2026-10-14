package legal_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// sideboard_test.go — ADR 0191 §8: a prompt over the cards a seat owns
// outside the game is answerable by the enumerator (#544), every move it
// offers is one the engine accepts, and it never names a card to a seat
// that may not see it.

func newSideboardTable(t *testing.T, side ...game.Card) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil)}
		for j := 0; j < 20; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if i == 0 {
			deck = append(deck, side...)
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat%d", i), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithSource(nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

func sideCard(name, typeLine string) game.Card {
	c := game.NewCard(name, uuid.Nil)
	c.TypeLine = typeLine
	c.IsSideboard = true
	return c
}

func outsideChoiceMoves(moves []legal.Move) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Kind == legal.KindChoice {
			out = append(out, m)
		}
	}
	return out
}

func TestAPickFromTheSideboardIsEnumeratedAndEveryMoveIsAccepted(t *testing.T) {
	g := newSideboardTable(t,
		sideCard("Alpha Sorcery", "Sorcery"), sideCard("Beta Sorcery", "Sorcery"), sideCard("Gamma Creature", "Creature — Bear"))
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
			Player: me.ID, Question: "Burning Wish — reveal a sorcery card", Max: 1,
			Match: func(c game.Card) bool { return c.IsSorcery() },
		})
	})
	moves := outsideChoiceMoves(legal.EnumerateFor(g, me.ID))
	got := labelsOf(moves)
	for _, want := range []string{"choose nothing", "Alpha Sorcery", "Beta Sorcery"} {
		if !strings.Contains(got, want) {
			t.Errorf("moves lack %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Gamma Creature") {
		t.Errorf("a creature was offered for a sorcery wish:\n%s", got)
	}
	if len(moves) != 3 {
		t.Errorf("%d moves, want choose nothing + 2 sorceries:\n%s", len(moves), got)
	}
	// Only a "may" has an always-legal answer: choosing nothing.
	always := 0
	for _, m := range moves {
		if m.AlwaysLegal {
			always++
		}
	}
	if always != 1 {
		t.Errorf("%d always-legal moves, want 1 (choose nothing)", always)
	}
	dispatchAll(t, g, me.ID, moves)
	// The opponent owes nothing and is offered nothing on it.
	if opp := outsideChoiceMoves(legal.EnumerateFor(g, g.Seats[1].ID)); len(opp) != 0 {
		t.Errorf("the opponent was offered %d moves for my prompt", len(opp))
	}
}

// A mandatory pick (Mastermind's Acquisition) has no "choose nothing"
// and still has an answer.
func TestAMandatoryPickFromTheSideboardHasAnAnswer(t *testing.T) {
	g := newSideboardTable(t, sideCard("Alpha", "Sorcery"), sideCard("Beta", "Instant"))
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(game.OutsidePickPrompt{Player: me.ID, Question: "Mastermind's Acquisition", Min: 1, Max: 1})
	})
	moves := outsideChoiceMoves(legal.EnumerateFor(g, me.ID))
	if len(moves) != 2 || strings.Contains(labelsOf(moves), "choose nothing") {
		t.Fatalf("moves:\n%s", labelsOf(moves))
	}
	dispatchAll(t, g, me.ID, moves)
}

// Extrapolate the Impossible's pair: only sets the engine accepts are
// offered, so a pair that shares a name is not.
func TestAPairPickOffersOnlyPairsWithDifferentNames(t *testing.T) {
	g := newSideboardTable(t,
		sideCard("Twin", "Sorcery"), sideCard("Twin", "Sorcery"), sideCard("Other", "Instant"))
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
			Player: me.ID, Question: "Extrapolate the Impossible", Min: 0, Max: 2,
			Validate: func(picked []game.Card) bool { return len(picked) == 2 && picked[0].Name != picked[1].Name },
		})
	})
	moves := outsideChoiceMoves(legal.EnumerateFor(g, me.ID))
	// choose nothing, Twin+Other, Twin+Other (two different Twins).
	if len(moves) != 3 {
		t.Fatalf("%d moves:\n%s", len(moves), labelsOf(moves))
	}
	for _, m := range moves {
		if strings.Contains(m.Label, "Twin Twin") {
			t.Errorf("a same-name pair was offered: %q", m.Label)
		}
	}
	dispatchAll(t, g, me.ID, moves)
}

// The opponent's half of Extrapolate: a reveal_pick over the two revealed
// cards is answerable by the opponent, who may name them.
func TestTheOpponentsPickAmongRevealedCardsIsEnumerated(t *testing.T) {
	g := newSideboardTable(t, sideCard("Alpha", "Sorcery"), sideCard("Beta", "Instant"))
	me, opp := g.Seats[0], g.Seats[1]
	ids := []uuid.UUID{me.SideboardCards()[0].InstanceID, me.SideboardCards()[1].InstanceID}
	g.WithWriteLock(func() {
		g.RevealOutsideForEffect(game.RevealSpec{Player: me.ID, Reason: "reveal", Cards: ids})
		if _, err := g.RevealPickThenForEffect(game.RevealPickPrompt{
			Chooser: opp.ID, Owner: me.ID, Question: "choose one", Min: 1, Max: 1, Cards: ids,
		}, nil); err != nil {
			t.Fatal(err)
		}
	})
	moves := outsideChoiceMoves(legal.EnumerateFor(g, opp.ID))
	got := labelsOf(moves)
	if len(moves) != 2 || !strings.Contains(got, "Alpha") || !strings.Contains(got, "Beta") {
		t.Fatalf("opponent's moves:\n%s", got)
	}
	dispatchAll(t, g, opp.ID, moves)
}

// A seat is never told the name of a card it may not see: the owner's
// prompt names the cards, and nobody else is offered the prompt at all.
func TestTheSideboardPromptNamesNothingToAnotherSeat(t *testing.T) {
	g := newSideboardTable(t, sideCard("Secret Sorcery", "Sorcery"))
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(game.OutsidePickPrompt{Player: me.ID, Question: "Wish", Max: 1})
	})
	all := labelsOf(legal.EnumerateFor(g, g.Seats[1].ID))
	if strings.Contains(all, "Secret Sorcery") {
		t.Errorf("the opponent's move list names a sideboard card:\n%s", all)
	}
	if !strings.Contains(labelsOf(legal.EnumerateFor(g, me.ID)), "Secret Sorcery") {
		t.Error("the owner's own prompt does not name their card")
	}
}
