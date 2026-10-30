package mcpseat

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// loop.go — ADR 0193 at the agent seat (owner decision 9: an MCP seat
// may record and propose a loop, and answers proposals like any seat).
// The moves are the enumerator's (move kind "loop", and the
// resolve_loop_proposal answers); this is the header that tells the
// agent what is being recorded or proposed, since the board does not
// show it.

// loopHeadLines are the decision header's lines about a demonstrated
// loop: one for the table's recording, if any, and the iteration's
// steps when the proposal is the question this seat is answering.
func loopHeadLines(v *protocol.GameView, me string, nw nameWrapper) []string {
	lr := v.LoopRecording
	if lr == nil {
		return nil
	}
	who := "someone"
	for i := range v.Seats {
		if v.Seats[i].ID == lr.Recorder {
			who = nw.apply(v.Seats[i].Name)
		}
	}
	var out []string
	switch lr.Status {
	case "armed":
		out = append(out, fmt.Sprintf("LOOP (CR 732.2): %s will record a loop from their next priority with nothing open. Every action from then on, yours included, is a step.", who))
	case "recording":
		line := fmt.Sprintf("LOOP (CR 732.2): %s is recording a loop, %d steps so far. Every action, yours included, is a step.", who, lr.Steps)
		if lr.Stopping {
			line += " It ends at their next priority with nothing open."
		}
		out = append(out, line)
	case "ready":
		out = append(out, fmt.Sprintf("LOOP (CR 732.2): %s recorded %s, %d steps; each time: %s.", who, nw.apply(lr.Name), lr.Steps, nw.apply(lr.Net)))
	case "proposed":
		out = append(out, fmt.Sprintf("LOOP (CR 732.2): %s proposes repeating %s %d more times; each time: %s.", who, nw.apply(lr.Name), lr.Count, nw.apply(lr.Net)))
	}
	for i := range v.PendingChoices {
		c := &v.PendingChoices[i]
		if c.Chooser != me || c.LoopProposal == nil {
			continue
		}
		mine := map[int]bool{}
		for _, p := range c.LoopProposal.ChoicePoints {
			mine[p] = true
		}
		out = append(out, "The iteration (accept, or shorten to one of your own steps, marked *):")
		for j, s := range c.LoopProposal.Steps {
			mark := " "
			if mine[j] {
				mark = "*"
			}
			out = append(out, fmt.Sprintf("  %s%d. %s", mark, j+1, nw.apply(s.Label)))
		}
	}
	return out
}
