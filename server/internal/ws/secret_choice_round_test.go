package ws

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// secret_choice_round_test.go — #5948 Q3, ADR 0180 amendment 2026-11-24.
// A secret choice round asks different questions of different players, a
// number among them, and keeps every answer off every other seat's wire
// until the last one is in: the same filter, the same single reveal and
// the same undo refusal as a secret vote.

var secretRoundResult game.VoteResult

var secretRoundThen = game.RegisterVoteThen("test/ws-secret-choice-round", func(_ *game.Game, r game.VoteResult) error {
	secretRoundResult = r
	return nil
})

func secretRoundRoom(t *testing.T) (*Room, *game.Game) {
	t.Helper()
	g := seedTestGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	a, b := g.Seats[0].ID, g.Seats[1].ID
	if _, _, err := room.Apply(uuid.Nil, func() error {
		var err error
		g.WithWriteLock(func() {
			err = g.StartSecretChoicesForEffect(game.SecretChoicePrompt{
				Controller: a,
				Question:   "secret test round",
				Options:    []game.ChoiceOption{{Label: "1"}, {Label: "2"}, {Label: "3"}},
				Asks: []game.VoteAsk{
					{Voter: a, Number: &game.VoteNumber{NoMax: true}},
					{Voter: b, Allowed: []int{1, 2}},
				},
				Then: secretRoundThen,
			})
		})
		return err
	}); err != nil {
		t.Fatalf("start round: %v", err)
	}
	return room, g
}

func secretRoundPrompt(t *testing.T, g *game.Game, who uuid.UUID) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.CouncilVote != nil && c.Chooser == who {
			return c
		}
	}
	t.Fatalf("no ask open for %s", who)
	return nil
}

func TestASecretRoundNumberStaysOffEveryOtherSeatUntilTheReveal(t *testing.T) {
	secretRoundResult = game.VoteResult{}
	room, g := secretRoundRoom(t)
	a, b := g.Seats[0].ID, g.Seats[1].ID

	ask := secretRoundPrompt(t, g, a)
	if ask.Kind != game.PendingChoicePayAmount || ask.PayAmount == nil || !ask.PayAmount.NoMax {
		t.Fatalf("a number ask is a pay_amount with no ceiling: %+v", ask)
	}
	if _, _, err := room.Apply(a, func() error { return g.ResolvePayAmount(ask.ID, a, 987654) }); err != nil {
		t.Fatalf("answer the number: %v", err)
	}

	full, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	// The unfiltered view never carries the answer: not as a ballot, not
	// as a number anywhere in the open prompt.
	raw, err := json.Marshal(full.PendingChoices)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "987654") || strings.Contains(string(raw), `"ballots"`) {
		t.Fatalf("the unfiltered prompt carries the number: %s", raw)
	}
	for viewer, want := range map[string]int{a.String(): 1, b.String(): 0, protocol.SpectatorViewerID: 0, "": 0} {
		var cv *protocol.CouncilVoteView
		for _, c := range protocol.FilterViewFor(full, viewer).PendingChoices {
			if c.CouncilVote != nil {
				cv = c.CouncilVote
			}
		}
		if cv == nil || !cv.Secret || !cv.Choice {
			t.Fatalf("viewer %q: no secret choice on the wire: %+v", viewer, cv)
		}
		if len(cv.Ballots) != want {
			t.Errorf("viewer %q sees %d answers, want %d", viewer, len(cv.Ballots), want)
		}
		if want == 1 && (cv.Ballots[0].Option != game.VoteNumberIndex || cv.Ballots[0].Number != 987654) {
			t.Errorf("the chooser sees their own answer as %+v", cv.Ballots[0])
		}
	}
	// b may pick only the options their ask allows.
	pick := secretRoundPrompt(t, g, b)
	var offered []string
	for _, o := range pick.PickOptions {
		offered = append(offered, o.Label)
	}
	if strings.Join(offered, ",") != "2,3" {
		t.Fatalf("b is offered %v, want 2 and 3", offered)
	}
	for _, e := range full.Log {
		if e.Kind == protocol.LogVoteSecret && !strings.Contains(e.Text, "chose in secret") {
			t.Errorf("a secret choice is logged as %q", e.Text)
		}
	}

	v, _, err := room.Apply(b, func() error { return g.ResolveOptionPick(pick.ID, b, 1) })
	if err != nil {
		t.Fatalf("answer the pick: %v", err)
	}
	var reveal string
	for _, e := range v.Log {
		if e.Kind == protocol.LogVotesRevealed {
			reveal = e.Text
		}
	}
	if !strings.Contains(reveal, "Choices revealed") || !strings.Contains(reveal, "chose 987654") || !strings.Contains(reveal, "chose 3") {
		t.Errorf("the reveal reads %q", reveal)
	}
	first, ok := secretRoundResult.AnswerTo(0)
	second, ok2 := secretRoundResult.AnswerTo(1)
	if !ok || !ok2 || first.Number != 987654 || first.Option != game.VoteNumberIndex || second.Option != 2 {
		t.Errorf("the continuation read %+v / %+v", first, second)
	}
	// The answer that revealed the round can't be taken back by its seat.
	if _, _, err := room.Undo(b); !errors.Is(err, ErrUndoPastSecretVote) {
		t.Fatalf("seat undo of the revealing answer: %v, want ErrUndoPastSecretVote", err)
	}
}
