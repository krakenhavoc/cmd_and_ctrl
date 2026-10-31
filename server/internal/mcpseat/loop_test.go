package mcpseat

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0193: the agent is told what loop is being recorded or proposed,
// and, when it is the one asked, which steps are its own.
func TestLoopHeadLinesNameTheLoopAndTheChoosersSteps(t *testing.T) {
	me, them := "seat-me", "seat-them"
	v := &protocol.GameView{
		Seats: []protocol.PlayerView{{ID: them, Name: "Ann", Seat: 0}, {ID: me, Name: "Bob", Seat: 1}},
		LoopRecording: &protocol.LoopRecordingView{
			Recorder: them, Status: "proposed", Steps: 3, Name: "Basalt Monolith", Net: "Ann: library -1", Count: 50,
		},
		PendingChoices: []protocol.PendingChoiceView{{
			Chooser: me, Kind: "loop_proposal",
			LoopProposal: &protocol.LoopProposalView{
				Proposer: them, Count: 50,
				Steps: []protocol.LoopStepView{
					{Actor: them, Label: "Ann activates Basalt Monolith"},
					{Actor: them, Label: "Ann passes"},
					{Actor: me, Label: "Bob passes"},
				},
				ChoicePoints: []int{2},
			},
		}},
	}
	got := strings.Join(loopHeadLines(v, me, newViewerNameWrapper(v, me)), "\n")
	for _, want := range []string{"proposes repeating Basalt Monolith 50 more times", "library -1", "*3. you (seat 1) passes", " 1. «Ann» activates"} {
		if !strings.Contains(got, want) {
			t.Errorf("header lacks %q:\n%s", want, got)
		}
	}
	if loopHeadLines(&protocol.GameView{}, me, nameWrapper{}) != nil {
		t.Error("a table with no loop has loop lines")
	}
}

// #4736: an armed recording and a stop that waits for priority are
// named too, so an agent knows its next actions will be steps.
func TestLoopHeadLinesNameAnArmedOrStoppingRecording(t *testing.T) {
	me, them := "seat-me", "seat-them"
	seats := []protocol.PlayerView{{ID: them, Name: "Ann", Seat: 0}, {ID: me, Name: "Bob", Seat: 1}}
	armed := &protocol.GameView{Seats: seats, LoopRecording: &protocol.LoopRecordingView{Recorder: them, Status: "armed"}}
	if got := strings.Join(loopHeadLines(armed, me, newViewerNameWrapper(armed, me)), "\n"); !strings.Contains(got, "will record a loop from their next priority") {
		t.Errorf("armed header %q", got)
	}
	stopping := &protocol.GameView{Seats: seats, LoopRecording: &protocol.LoopRecordingView{Recorder: them, Status: "recording", Steps: 2, Stopping: true}}
	if got := strings.Join(loopHeadLines(stopping, me, newViewerNameWrapper(stopping, me)), "\n"); !strings.Contains(got, "2 steps so far") || !strings.Contains(got, "It ends at their next priority") {
		t.Errorf("stopping header %q", got)
	}
}
