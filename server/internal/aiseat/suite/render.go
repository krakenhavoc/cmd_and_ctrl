package suite

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
)

// render.go is the labelling screen.
//
// It prints the exact bytes a model would be shown for one position,
// followed by the move list with its real indices and a marker on
// every move somebody has an opinion about. That is what makes
// labelling affordable: a position is 40 KiB of JSON view that no
// human is going to read, and the same window rendered as the prompt
// is a page.
//
// The prompt is REBUILT from the frozen Input rather than read off
// the harvested trace. Two reasons, and the second is the important
// one. A harvested position may never have reached a model at all
// (the heuristic games that seeded the suite never did), so there is
// no recorded prompt to print. And the decision log deliberately
// stores the static system block once per file rather than once per
// record — it is byte-identical across a game, and repeating it would
// multiply the file by the size of the decklist — so a record's
// Prompt.System is empty for every record but the first. Rebuilding
// through BuildRequest is the only way to get the whole thing, and it
// has the side benefit of showing the prompt TODAY'S code would send,
// which is what a labeller comparing against a live seat wants.

// Render writes the position's header, the model prompt it would
// produce under cfg, and the annotated move list.
func Render(w io.Writer, p Position, cfg model.Config) error {
	var b strings.Builder

	fmt.Fprintf(&b, "=== %s ===\n", p.ID)
	if len(p.Tags) > 0 {
		fmt.Fprintf(&b, "tags      %s\n", strings.Join(p.Tags, ", "))
	}
	if p.Note != "" {
		fmt.Fprintf(&b, "note      %s\n", p.Note)
	}
	if len(p.Gate) > 0 {
		fmt.Fprintf(&b, "gate      %s\n", strings.Join(p.Gate, ", "))
	}
	fmt.Fprintf(&b, "source    game %s seat %d seq %d, turn %d %s\n",
		short(p.Source.GameID), p.Source.Seat, p.Source.Seq, p.Source.Turn, p.Source.Step)
	if p.Source.Reviewer != "" {
		fmt.Fprintf(&b, "reviewed  %s %s\n", p.Source.Reviewer, p.Source.ReviewedAt)
	}
	if p.Source.Log != "" {
		fmt.Fprintf(&b, "log       %s\n", p.Source.Log)
	}
	if !p.Labelled() {
		b.WriteString("label     NONE YET — this position is an unanswered question\n")
	}
	b.WriteString("\n")

	pol := model.New(cfg)
	req, _, verdict := pol.BuildRequest(context.Background(), p.Input)
	if verdict.Absorbed() {
		fmt.Fprintf(&b, "--- PROMPT ---\nLayer A absorbed this window (rule %q → move %d %q); no model call would be made.\n\n",
			verdict.Rule, verdict.Index, moveLabel(p, verdict.Index))
	} else {
		for i, blk := range req.System {
			fmt.Fprintf(&b, "--- SYSTEM[%d] ---\n%s\n\n", i, blk.Text)
		}
		fmt.Fprintf(&b, "--- USER ---\n%s\n\n", req.User)
	}

	fmt.Fprintf(&b, "--- MOVES (%d) ---\n", len(p.Input.Moves))
	for i, mv := range p.Input.Moves {
		var marks []string
		if p.Accepts(i) {
			marks = append(marks, "<- accept")
		}
		if p.Rejects(i) {
			marks = append(marks, "<- reject")
		}
		if h := p.AtCapture.Heuristic; h != nil && h.Index == i {
			marks = append(marks, "<- heuristic")
		}
		if m := p.AtCapture.Model; m != nil && m.Index == i {
			marks = append(marks, "<- model@capture")
		}
		fmt.Fprintf(&b, "%3d: [%s] %s", i, mv.Kind, mv.Label)
		if len(marks) > 0 {
			fmt.Fprintf(&b, "   %s", strings.Join(marks, " "))
		}
		b.WriteString("\n")
	}
	if p.Expected.DeclineOK {
		b.WriteString("     (declining is accepted on this position)\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func moveLabel(p Position, i int) string {
	if i < 0 || i >= len(p.Input.Moves) {
		return ""
	}
	return p.Input.Moves[i].Label
}

func short(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
