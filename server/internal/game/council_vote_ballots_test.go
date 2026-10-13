package game

import (
	"testing"

	"github.com/google/uuid"
)

// council_vote_ballots_test.go — ADR 0146's amendment of 2026-10-13:
// the readers behind "a choice you voted for" (#2927).

func TestVotedForListsDistinctOptionsInFirstVoteOrder(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	ballots := []VoteBallot{{Voter: a, Option: 2}, {Voter: b, Option: 0}, {Voter: a, Option: 1}, {Voter: a, Option: 2}}
	got := VotedFor(ballots, a)
	if len(got) != 2 || got[0] != 2 || got[1] != 1 {
		t.Errorf("VotedFor = %v, want [2 1]", got)
	}
	if got := VotedFor(ballots, uuid.New()); len(got) != 0 {
		t.Errorf("a player who cast no ballot voted for %v, want nothing", got)
	}
}

func TestVotedForAChoiceSharedAndNot(t *testing.T) {
	you, one, split, none := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	ballots := []VoteBallot{
		{Voter: you, Option: 0},
		{Voter: one, Option: 0},
		{Voter: split, Option: 0}, {Voter: split, Option: 1},
	}
	cases := []struct {
		name          string
		voter         uuid.UUID
		shared, apart bool
	}{
		{"same choice only", one, true, false},
		{"one vote each way (CR 701.38d)", split, true, true},
		{"cast no ballot", none, false, false},
	}
	for _, c := range cases {
		if got := VotedForAChoiceVotedForBy(ballots, c.voter, you); got != c.shared {
			t.Errorf("%s: shared = %v, want %v", c.name, got, c.shared)
		}
		if got := VotedForAChoiceNotVotedForBy(ballots, c.voter, you); got != c.apart {
			t.Errorf("%s: apart = %v, want %v", c.name, got, c.apart)
		}
	}

	// A controller with two votes who took both choices: nobody voted
	// for one they didn't.
	both := []VoteBallot{{Voter: you, Option: 0}, {Voter: you, Option: 1}, {Voter: one, Option: 1}}
	if VotedForAChoiceNotVotedForBy(both, one, you) || !VotedForAChoiceVotedForBy(both, one, you) {
		t.Error("a controller who voted for both choices shares every choice")
	}
	// A controller who cast no ballot voted for no choice.
	absent := []VoteBallot{{Voter: one, Option: 1}}
	if VotedForAChoiceVotedForBy(absent, one, you) || !VotedForAChoiceNotVotedForBy(absent, one, you) {
		t.Error("with no ballot of mine, no one shares a choice and every choice is one I didn't vote for")
	}
}
