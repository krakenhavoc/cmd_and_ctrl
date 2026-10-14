package ws

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// secret_vote_undo_test.go — ADR 0180 decision 4. The last ballot of a
// secret vote is the commit that reveals every other ballot, so the
// seat that cast it may not take it back and vote again; an undo
// before the reveal restores a vote whose earlier ballots are still
// hidden; the admin may still undo.

var secretUndoThen = game.RegisterVoteThen("test/ws-secret-vote-undo", func(*game.Game, game.VoteResult) error { return nil })

func secretUndoRoom(t *testing.T) (*Room, *game.Game) {
	t.Helper()
	g := seedTestGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	if _, _, err := room.Apply(uuid.Nil, func() error {
		var err error
		g.WithWriteLock(func() {
			err = g.StartVoteForEffect(game.VotePrompt{
				Starting: g.Seats[0].ID, Controller: g.Seats[0].ID,
				Question: "secret test vote",
				Options:  []game.ChoiceOption{{Label: "yes"}, {Label: "no"}},
				Secret:   true,
				Then:     secretUndoThen,
			})
		})
		return err
	}); err != nil {
		t.Fatalf("start vote: %v", err)
	}
	return room, g
}

func castSecretBallot(t *testing.T, room *Room, g *game.Game, voter uuid.UUID, label string) protocol.GameView {
	t.Helper()
	var ballot *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.CouncilVote != nil && c.Chooser == voter {
			ballot = c
		}
	}
	if ballot == nil {
		t.Fatalf("no ballot open for %s", voter)
	}
	for i, o := range ballot.PickOptions {
		if o.Label == label {
			v, _, err := room.Apply(voter, func() error { return g.ResolveOptionPick(ballot.ID, voter, i) })
			if err != nil {
				t.Fatalf("cast %q: %v", label, err)
			}
			return v
		}
	}
	t.Fatalf("ballot offers no %q", label)
	return protocol.GameView{}
}

func TestTheBallotThatRevealsASecretVoteCannotBeUndoneBySeat(t *testing.T) {
	room, g := secretUndoRoom(t)
	a, b := g.Seats[0].ID, g.Seats[1].ID
	castSecretBallot(t, room, g, a, "yes")
	v := castSecretBallot(t, room, g, b, "no")

	if got := blockedFor(t, v, b); got != protocol.UndoBlockedSecretVote {
		t.Fatalf("the last voter's undo badge says %q, want %q", got, protocol.UndoBlockedSecretVote)
	}
	if _, _, err := room.Undo(b); !errors.Is(err, ErrUndoPastSecretVote) {
		t.Fatalf("seat undo of the revealing ballot: %v, want ErrUndoPastSecretVote", err)
	}
	// The admin may still take it back, for mistakes.
	if _, _, err := room.Undo(uuid.Nil); err != nil {
		t.Fatalf("admin undo: %v", err)
	}
}

func TestUndoBeforeTheRevealKeepsTheEarlierBallotsHidden(t *testing.T) {
	room, g := secretUndoRoom(t)
	a, b := g.Seats[0].ID, g.Seats[1].ID
	castSecretBallot(t, room, g, a, "yes")
	// A takes their ballot back and votes again: nothing was revealed.
	if _, _, err := room.Undo(a); err != nil {
		t.Fatalf("undo before the reveal: %v", err)
	}
	castSecretBallot(t, room, g, a, "no")

	// B's view of the restored, re-cast vote carries no ballot of A's;
	// A's carries their own.
	full, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	for viewer, want := range map[string]int{a.String(): 1, b.String(): 0, protocol.SpectatorViewerID: 0, "": 0} {
		var cv *protocol.CouncilVoteView
		for _, c := range protocol.FilterViewFor(full, viewer).PendingChoices {
			if c.CouncilVote != nil {
				cv = c.CouncilVote
			}
		}
		if cv == nil || !cv.Secret {
			t.Fatalf("viewer %q: no secret vote on the wire: %+v", viewer, cv)
		}
		if len(cv.Ballots) != want {
			t.Errorf("viewer %q sees %d ballots, want %d", viewer, len(cv.Ballots), want)
		}
	}
}
