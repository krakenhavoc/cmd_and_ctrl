package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// vote_test.go — ADR 0146: how the bot casts a ballot.

func ballotInput(t *testing.T, vote protocol.CouncilVoteView, options []protocol.PickOptionView, cards ...protocol.CardView) aiseat.Input {
	t.Helper()
	const choiceID = "choice-ballot"
	vote.Voter = seatID(0).String()
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(cards...),
		withChoice(protocol.PendingChoiceView{
			ID:          choiceID,
			Kind:        "option_pick",
			Chooser:     seatID(0).String(),
			Reason:      "a vote",
			PickOptions: options,
			CouncilVote: &vote,
		}))
	var moves []legal.Move
	for i, o := range options {
		moves = append(moves, choiceMove(t, 0, choiceID, o.Label, map[string]any{"option_index": i}))
	}
	return input(0, v, moves...)
}

func words(labels ...string) []protocol.PickOptionView {
	out := make([]protocol.PickOptionView, len(labels))
	for i, l := range labels {
		out[i] = protocol.PickOptionView{Label: l}
	}
	return out
}

// The vote's controller votes for what the card says it wants; an
// opponent votes for what the card says an opponent wants.
func TestHeuristicVotesForItsSidesWord(t *testing.T) {
	vote := protocol.CouncilVoteView{
		Options: []string{"time", "knowledge"}, Tally: []int{0, 0}, Offered: []int{0, 1},
		ForController: []int{2, 1}, ForOpponents: []int{0, 1},
	}
	vote.Controller = seatID(0).String()
	in := ballotInput(t, vote, words("time", "knowledge"))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "time" {
		t.Errorf("the caster's bot voted %q, want time", got)
	}
	vote.Controller = seatID(1).String()
	in = ballotInput(t, vote, words("time", "knowledge"))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "knowledge" {
		t.Errorf("an opponent's bot voted %q, want knowledge", got)
	}
}

// An optional extra vote is cast, not declined.
func TestHeuristicCastsItsExtraVote(t *testing.T) {
	vote := protocol.CouncilVoteView{
		Options: []string{"evidence", "bribery"}, Tally: []int{0, 1}, Offered: []int{0, 1, -1},
		ForController: []int{1, 2}, ForOpponents: []int{1, 0}, Optional: true,
	}
	vote.Controller = seatID(0).String()
	in := ballotInput(t, vote, words("evidence", "bribery", "Don't vote again"))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "bribery" {
		t.Errorf("the bot answered its extra vote with %q, want bribery", got)
	}
}

// An object vote: the most valuable permanent the bot doesn't control,
// never its own.
func TestHeuristicVotesToExileTheirBestPermanent(t *testing.T) {
	small := creature("c-small", 1, "Small", 1, 1)
	big := creature("c-big", 1, "Big", 6, 6)
	mine := creature("c-mine", 0, "Mine", 8, 8)
	vote := protocol.CouncilVoteView{
		Options: []string{"Small", "Big", "Mine"}, Tally: []int{0, 0, 0}, Offered: []int{0, 1, 2},
	}
	vote.Controller = seatID(1).String()
	opts := []protocol.PickOptionView{
		{Label: "Small", Cards: []protocol.CardView{small}},
		{Label: "Big", Cards: []protocol.CardView{big}},
		{Label: "Mine", Cards: []protocol.CardView{mine}},
	}
	in := ballotInput(t, vote, opts, small, big, mine)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Big" {
		t.Errorf("the bot voted for %q, want their biggest permanent", got)
	}
}
