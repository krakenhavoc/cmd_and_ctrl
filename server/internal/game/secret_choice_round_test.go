package game

import (
	"testing"

	"github.com/google/uuid"
)

// secret_choice_round_test.go — #5948 Q3, ADR 0180 amendment 2026-11-24:
// a secret choice round's asks, skips and departures.

var secretRoundResults []VoteResult

var secretRoundTestThen = RegisterVoteThen(testEffectKeyPrefix+"secret-round", func(_ *Game, r VoteResult) error {
	secretRoundResults = append(secretRoundResults, r)
	return nil
})

func startSecretRound(t *testing.T, g *Game, p SecretChoicePrompt) {
	t.Helper()
	secretRoundResults = nil
	p.Then = secretRoundTestThen
	p.Controller = g.Seats[0].ID
	p.Source = departureTestSource(g, p.Controller, "Test Secret")
	g.WithWriteLock(func() {
		if err := g.StartSecretChoicesForEffect(p); err != nil {
			t.Fatalf("StartSecretChoicesForEffect: %v", err)
		}
	})
}

func TestASecretRoundSkipsAnAskWithNothingToOfferAndAVoterWhoLeft(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	startSecretRound(t, g, SecretChoicePrompt{
		Question: "test round",
		Options:  []ChoiceOption{{Label: "x"}, {Label: "y"}},
		Asks: []VoteAsk{
			{Voter: a.ID, Number: &VoteNumber{Min: 1, Max: 5}},
			{Voter: b.ID, Allowed: []int{0}},
			{Voter: c.ID, Number: &VoteNumber{NoMax: true}},
			{Voter: b.ID, Allowed: []int{1}},
		},
	})
	ask := openBallot(g)
	if ask == nil || ask.Kind != PendingChoicePayAmount || ask.PayAmount.Min != 1 || ask.PayAmount.Max != 5 {
		t.Fatalf("the first ask is a number from 1 to 5: %+v", ask)
	}
	if err := g.ResolvePayAmount(ask.ID, a.ID, 6); err == nil {
		t.Fatal("a number above the range was accepted")
	}
	if err := g.ResolvePayAmount(ask.ID, a.ID, 4); err != nil {
		t.Fatal(err)
	}
	ask = openBallot(g)
	if ask == nil || ask.Chooser != b.ID || len(ask.PickOptions) != 1 || ask.PickOptions[0].Label != "x" {
		t.Fatalf("b's first ask offers only x: %+v", ask)
	}
	if err := g.ResolveOptionPick(ask.ID, b.ID, 0); err != nil {
		t.Fatal(err)
	}
	// c leaves with their number unanswered: nothing is recorded for them
	// and the round goes on (CR 800.4a).
	if err := g.Concede(c.ID); err != nil {
		t.Fatal(err)
	}
	ask = openBallot(g)
	if ask == nil || ask.Chooser != b.ID || ask.PickOptions[0].Label != "y" {
		t.Fatalf("after c left, b's second ask is open: %+v", ask)
	}
	if len(secretRoundResults) != 0 {
		t.Fatal("the round finished early")
	}
	if err := g.ResolveOptionPick(ask.ID, b.ID, 0); err != nil {
		t.Fatal(err)
	}
	if len(secretRoundResults) != 1 {
		t.Fatalf("the round finished %d times", len(secretRoundResults))
	}
	r := secretRoundResults[0]
	if got, ok := r.AnswerTo(0); !ok || got.Option != VoteNumberIndex || got.Number != 4 || got.Voter != a.ID {
		t.Errorf("ask 0: %+v", got)
	}
	if got, ok := r.AnswerTo(1); !ok || got.Option != 0 {
		t.Errorf("ask 1: %+v", got)
	}
	if _, ok := r.AnswerTo(2); ok {
		t.Error("the departed voter's ask has an answer")
	}
	if got, ok := r.AnswerTo(3); !ok || got.Option != 1 {
		t.Errorf("ask 3: %+v", got)
	}
	revealed, finished := 0, 0
	for _, ev := range g.Events {
		switch ev.Kind {
		case EventVotesRevealed:
			revealed++
			if ev.Label != SecretChoiceLabel || len(ev.Ballots) != 3 || ev.VoteOptions[ev.Ballots[0].Option] != "4" {
				t.Errorf("the reveal: %+v", ev)
			}
		case EventVotesFinished:
			finished++
		}
	}
	if revealed != 1 || finished != 0 {
		t.Errorf("revealed %d, finished voting %d; want 1 and 0", revealed, finished)
	}
}

func TestASecretRoundWithEveryAskSkippedStillRunsItsContinuation(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	var thing ChoiceOption
	g.WithWriteLock(func() {
		thing = ChoiceOption{Label: "A's Thing", Cards: []uuid.UUID{departureTestSource(g, a.ID, "A's Thing")}}
	})
	// "A creature you control", asked of a player who controls none.
	startSecretRound(t, g, SecretChoicePrompt{
		Question: "test round",
		Options:  []ChoiceOption{thing},
		Asks:     []VoteAsk{{Voter: b.ID, Relation: VoteAskControls}},
	})
	if openBallot(g) != nil {
		t.Fatal("an ask over no permanent was put")
	}
	if len(secretRoundResults) != 1 || len(secretRoundResults[0].Ballots) != 0 {
		t.Fatalf("the continuation ran with %+v", secretRoundResults)
	}
}
