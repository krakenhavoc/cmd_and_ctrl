package mcpseat

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ballotMoves is two answers to one prompt, the first marked AlwaysLegal
// as the enumerator marks the first option of every option_pick.
func ballotMoves(choiceID string, player uuid.UUID) []legal.Move {
	var out []legal.Move
	for i, label := range []string{"Council's Judgment: vote for Alpha Bear", "Council's Judgment: vote for Beta Bear"} {
		out = append(out, legal.Move{
			Type:        legal.TypeResolveChoice,
			Kind:        legal.KindChoice,
			Player:      player,
			Label:       label,
			AlwaysLegal: i == 0,
			Params:      mustJSON(map[string]any{"choice_id": choiceID, "option_index": i}),
		})
	}
	return out
}

// #3669: every vote option is equally legal, so the line must not tag
// the first one "always legal". The first option of a plain option_pick
// keeps its tag.
func TestMoveLineDoesNotTagAVoteOptionAlwaysLegal(t *testing.T) {
	v := &protocol.GameView{PendingChoices: []protocol.PendingChoiceView{
		{ID: "vote", Kind: "option_pick", CouncilVote: &protocol.CouncilVoteView{}},
		{ID: "plain", Kind: "option_pick"},
	}}
	vote := ballotMoves("vote", uuid.Nil)
	plain := ballotMoves("plain", uuid.Nil)
	if !vote[0].AlwaysLegal {
		t.Fatal("fixture: the first ballot option should carry the flag")
	}
	for i, m := range vote {
		if got := moveLine(i, m, v, "", newNameWrapper(v)); strings.Contains(got, "always legal") {
			t.Errorf("vote option %d says always legal: %q", i, got)
		}
	}
	if got := moveLine(0, plain[0], v, "", newNameWrapper(v)); !strings.Contains(got, "always legal") {
		t.Errorf("a plain option_pick's first option lost its tag: %q", got)
	}
}

// #3669: the tag is gone from the text, but the first option is still
// the move the seat may take after three refusals, so a mandatory vote
// cannot stall the seat.
func TestAVoteCanStillBeAnsweredAfterThreeRefusals(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	v := activeView(f)
	v.PendingChoices = []protocol.PendingChoiceView{{
		ID: "vote", Kind: "option_pick", Chooser: f.playerID.String(),
		CouncilVote: &protocol.CouncilVoteView{},
	}}
	f.mu.Lock()
	f.onAction = func(protocol.ActionPayload) string { return "error" }
	f.mu.Unlock()
	f.setState(v, ballotMoves("vote", f.playerID), false)
	tok, window := decisionWindow(t, s)
	if strings.Contains(window, "always legal") {
		t.Errorf("the vote window tags an option always legal:\n%s", window)
	}
	for i := 0; i < maxRejections; i++ {
		if r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: 1}); !strings.Contains(r.Text, "status: rejected") {
			t.Fatalf("rejection %d: %s", i, r.Text)
		}
	}
	sent := f.actionCount()
	if r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: 1}); !r.IsError {
		t.Fatalf("a fourth try at a plain ballot option went through: %s", r.Text)
	}
	if f.actionCount() != sent {
		t.Error("the refused move was sent")
	}
	// The escape is the first option.
	r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: 0})
	if f.actionCount() != sent+1 {
		t.Errorf("the first vote option was not sent after three refusals: %s", r.Text)
	}
}
