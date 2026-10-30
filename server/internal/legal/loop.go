package legal

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loop.go — ADR 0193 §8 (CR 732.2): the demonstrated-loop moves.
//
//   - The four recording verbs (KindLoop) go to an AGENT seat that
//     holds priority: owner decision 9 lets an MCP seat record and
//     propose a loop, and bots never do. A bot seat is never offered
//     them, which is stronger than ranking them last, and a browser's
//     list does not carry them either: the client drives the dock
//     control from GameView.loop_recording, and a move kind the client
//     does not know would stop its smart autopass on every window.
//     The server accepts loop_record_start from any seat in the game
//     and arms it when the seat cannot start now (#4736); an agent is
//     still offered it only in its priority windows, so a record move
//     never turns a window it has nothing else to do in into a
//     decision.
//   - The loop_proposal prompt is answered by every seat. Accept is
//     first and always legal, so a bot takes it; a bot seat is offered
//     nothing else (owner decision 9: bot opponents always accept).

// KindLoop marks the recording verbs: record, stop, cancel, propose.
const KindLoop Kind = "loop"

// Wire action types of the loop verbs (actions.TypeLoop*).
const (
	TypeLoopRecordStart     = "loop_record_start"
	TypeLoopRecordStop      = "loop_record_stop"
	TypeLoopRecordCancel    = "loop_record_cancel"
	TypeLoopPropose         = "loop_propose"
	TypeResolveLoopProposal = "resolve_loop_proposal"
)

// LoopProposeCounts is the bounded set of counts the enumerator offers
// for loop_propose (ADR 0193 §8); a person may send any count up to the
// cap.
var LoopProposeCounts = []int{1, 10, 100, game.MaxLoopShortcutIterations}

type loopProposeParams struct {
	Count int `json:"count"`
}

type loopAnswerParams struct {
	ChoiceID string          `json:"choice_id"`
	Accept   bool            `json:"accept,omitempty"`
	Shorten  *game.LoopPoint `json:"shorten,omitempty"`
}

// loopMoves offers an agent seat that holds priority the recording
// verbs that apply now. Called from run's priority branch.
func (e *enumerator) loopMoves() {
	if !e.p.Agent || e.p.IsBot {
		return
	}
	e.leave()
	g := e.g
	r := g.LoopRecordingLocked()
	add := func(typ, label string, params any) {
		m := Move{Type: typ, Player: e.seat, Kind: KindLoop, Label: label}
		if params != nil {
			m.Params = mustJSON(params)
		}
		e.add(m)
	}
	if r == nil {
		if len(g.PendingChoices) == 0 && len(g.PendingTriggers) == 0 {
			add(TypeLoopRecordStart, "Record a loop (CR 732.2a): every action from now is one iteration", nil)
		} else {
			add(TypeLoopRecordStart, "Record a loop (CR 732.2a): every action from your next priority with nothing open is one iteration", nil)
		}
		return
	}
	if r.Recorder != e.seat {
		return
	}
	switch r.Status {
	case game.LoopStatusRecording:
		if len(r.Steps) > 0 && !r.StopAtPriority {
			add(TypeLoopRecordStop, fmt.Sprintf("Stop recording the loop (%d steps)", len(r.Steps)), nil)
		}
	case game.LoopStatusReady:
		if len(g.PendingChoices) == 0 {
			for _, n := range LoopProposeCounts {
				add(TypeLoopPropose, fmt.Sprintf("Propose repeating %s %d more times", r.Name(), n), loopProposeParams{Count: n})
			}
		}
	}
	add(TypeLoopRecordCancel, "Cancel the loop recording", nil)
}

// loopProposalMoves answers a loop_proposal prompt: accept, then — for
// a seat that is not a bot — shorten to their first choice point in
// iteration 1, when that is before the current end.
func (e *enumerator) loopProposalMoves(c *game.PendingChoice) {
	r := e.g.LoopRecordingLocked()
	if r == nil || r.Proposal == nil {
		return
	}
	e.add(Move{
		Type:        TypeResolveLoopProposal,
		Player:      e.seat,
		Kind:        KindChoice,
		Label:       fmt.Sprintf("Accept repeating %s %d more times", r.Name(), r.Proposal.Count),
		Params:      mustJSON(loopAnswerParams{ChoiceID: c.ID.String(), Accept: true}),
		AlwaysLegal: true,
	})
	if e.p.IsBot {
		return
	}
	points := r.LoopChoicePoints(e.seat)
	if len(points) == 0 {
		return
	}
	at := game.LoopPoint{Iteration: 1, Step: points[0]}
	end := r.Proposal.End
	if end.Iteration < at.Iteration || (end.Iteration == at.Iteration && end.Step < at.Step) {
		return
	}
	e.add(Move{
		Type:   TypeResolveLoopProposal,
		Player: e.seat,
		Kind:   KindChoice,
		Label:  fmt.Sprintf("Shorten the loop: stop in iteration 1 before “%s”", r.Steps[at.Step].Label),
		Params: mustJSON(loopAnswerParams{ChoiceID: c.ID.String(), Shorten: &at}),
	})
}
