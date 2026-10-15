package protocol

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loop_view.go — the wire half of ADR 0193 (CR 732.2): the table's view
// of a demonstrated loop, and the loop_proposal prompt's summary.
// Public and identical for every seat: CR 732.1a asks that each player
// understand the shortcut being built, and nothing here names a card in
// a hidden zone (a step's label names only public cards, game.
// LoopPublicName).

// LoopRecordingView is GameView.loop_recording: who is recording a
// loop, where it is, how many steps it has, and what one iteration
// changes once it closed.
type LoopRecordingView struct {
	Recorder string `json:"recorder"`
	// Status is "recording", "ready" or "proposed".
	Status string `json:"status"`
	Steps  int    `json:"steps"`
	// Name is the loop as the log names it ("Basalt Monolith ⇄ Rings
	// of Brighthearth").
	Name string `json:"name,omitempty"`
	// Net is one iteration's net change, once the recording closed
	// ("Luke: +1 {C}, library -1").
	Net string `json:"net,omitempty"`
	// MaxIterations is the cap on a proposal's count (owner decision 2).
	MaxIterations int `json:"max_iterations"`
	// Count and End describe an outstanding proposal.
	Count int            `json:"count,omitempty"`
	End   *LoopPointView `json:"end,omitempty"`
	// ShortenedBy is the seat that last shortened it.
	ShortenedBy string `json:"shortened_by,omitempty"`
	// Iteration is the recorded steps, once the recording closed: what
	// the recorder's propose modal lists.
	Iteration []LoopStepView `json:"iteration,omitempty"`
}

// LoopPointView is a place in a proposed sequence: before step `step`
// (0-based) of iteration `iteration` (1-based). An unshortened proposal
// ends at {count, steps}.
type LoopPointView struct {
	Iteration int `json:"iteration"`
	Step      int `json:"step"`
}

// LoopProposalView populates a loop_proposal prompt: the proposal the
// chooser is answering, the iteration's steps, and which of them are
// the chooser's own choices — the places CR 732.2b lets them shorten to.
type LoopProposalView struct {
	Proposer string         `json:"proposer"`
	Name     string         `json:"name,omitempty"`
	Count    int            `json:"count"`
	End      LoopPointView  `json:"end"`
	Net      string         `json:"net,omitempty"`
	Steps    []LoopStepView `json:"steps"`
	// ChoicePoints is the indices of Steps the chooser may shorten to.
	ChoicePoints []int `json:"choice_points,omitempty"`
}

// LoopStepView is one recorded step as the table reads it.
type LoopStepView struct {
	Actor string `json:"actor,omitempty"`
	Label string `json:"label"`
}

// viewOfLoopRecording projects the game's recording, or nil. Caller
// holds the read lock ViewOfGame holds.
func viewOfLoopRecording(g *game.Game) *LoopRecordingView {
	r := g.LoopRecordingLocked()
	if r == nil {
		return nil
	}
	v := &LoopRecordingView{
		Recorder:      r.Recorder.String(),
		Status:        string(r.Status),
		Steps:         len(r.Steps),
		Name:          r.Name(),
		MaxIterations: game.MaxLoopShortcutIterations,
	}
	if r.Status != game.LoopStatusRecording {
		v.Net = g.LoopNetTextLocked(r, 1)
		v.Iteration = loopStepViews(r)
	}
	if p := r.Proposal; p != nil {
		v.Count = p.Count
		v.End = &LoopPointView{Iteration: p.End.Iteration, Step: p.End.Step}
		v.ShortenedBy = uuidStringOrEmpty(p.ShortenedBy)
	}
	return v
}

// viewOfLoopProposal is the summary a loop_proposal prompt to chooser
// carries, or nil when no proposal stands. Caller holds the read lock.
func viewOfLoopProposal(g *game.Game, chooser uuid.UUID) *LoopProposalView {
	r := g.LoopRecordingLocked()
	if r == nil || r.Proposal == nil {
		return nil
	}
	v := &LoopProposalView{
		Proposer:     r.Recorder.String(),
		Name:         r.Name(),
		Count:        r.Proposal.Count,
		End:          LoopPointView{Iteration: r.Proposal.End.Iteration, Step: r.Proposal.End.Step},
		Net:          g.LoopNetTextLocked(r, 1),
		ChoicePoints: r.LoopChoicePoints(chooser),
		Steps:        loopStepViews(r),
	}
	return v
}

// loopStepViews is the recording's steps as the table reads them.
func loopStepViews(r *game.LoopRecording) []LoopStepView {
	out := make([]LoopStepView, len(r.Steps))
	for i, s := range r.Steps {
		out[i] = LoopStepView{Actor: uuidStringOrEmpty(s.Actor), Label: s.Label}
	}
	return out
}
