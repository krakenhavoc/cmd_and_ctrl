package aiseat

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// forced_internal_test.go is an in-package test for forcedAnswer,
// which is the rule the runner applies when a policy's picks keep
// being refused.
//
// It is in-package, and a unit test rather than a game, because the
// third case cannot be reached from a real table. It needs the
// enumerator and the engine to disagree in a window that offers
// neither a pass nor an unconditional answer: the enumerator's
// contract is that everything it offers is legal when offered, so
// producing that disagreement on demand is not something a test can
// do. The first two cases ARE reachable and are covered end to end in
// search_hatch_test.go and observer_test.go; this pins all three
// against the move list, which is the whole input the rule has.

func TestForcedAnswerPrefersThePass(t *testing.T) {
	moves := []legal.Move{
		{Kind: legal.KindChoice, Label: "take Island", AlwaysLegal: true},
		{Kind: legal.KindPass, Label: "Pass priority", AlwaysLegal: true},
	}
	idx, reason, forced := forcedAnswer(moves)
	if idx != 1 {
		t.Errorf("index %d, want the pass at 1", idx)
	}
	if forced != ForcedPass {
		t.Errorf("forced %q, want %q", forced, ForcedPass)
	}
	if reason == "" {
		t.Error("no reason for the log")
	}
}

func TestForcedAnswerFallsToTheAlwaysLegalAnswer(t *testing.T) {
	// A seat owing a pending choice: the choice's answers and nothing
	// else, so there is no pass to reach for (#544).
	moves := []legal.Move{
		{Kind: legal.KindChoice, Label: "take Forest and Island"},
		{Kind: legal.KindChoice, Label: "fail to find", AlwaysLegal: true},
	}
	idx, _, forced := forcedAnswer(moves)
	if idx != 1 {
		t.Errorf("index %d, want the always-legal answer at 1", idx)
	}
	if forced != ForcedAlwaysLegal {
		t.Errorf("forced %q, want %q", forced, ForcedAlwaysLegal)
	}
}

func TestForcedAnswerAdmitsWhenThereIsNothing(t *testing.T) {
	// No pass, nothing marked AlwaysLegal. The runner has run out of
	// answers; saying so is the whole point, because the alternative
	// is a seat that silently stops playing and holds the table.
	moves := []legal.Move{
		{Kind: legal.KindChoice, Label: "take Forest and Island"},
		{Kind: legal.KindChoice, Label: "take Island and Mountain"},
	}
	idx, reason, forced := forcedAnswer(moves)
	if idx != Decline {
		t.Errorf("index %d, want Decline (%d) — there is no move to take", idx, Decline)
	}
	if forced != ForcedNoLegalAnswer {
		t.Errorf("forced %q, want %q", forced, ForcedNoLegalAnswer)
	}
	if reason == "" {
		t.Error("no reason for the log")
	}
}
