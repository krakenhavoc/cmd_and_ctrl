package roadmap

import (
	"fmt"
	"strings"
)

// seams.go — docs/engine-seams.md's open table, generated from the
// registry.
//
// The doc is otherwise hand-written: its intro and its Closed seams
// list are history a person wrote, and nothing here touches them. The
// one machine-owned part is the block between the two markers, the
// same fence coverage/census.go draws around the roadmap census.
//
// Unlike the census, this block is a function of the REGISTRY alone —
// no catalog count, no clock — so it changes only in the PR that edits
// registry.go, and TestSeamsTableIsCurrent fails that PR until it has
// run -update. There is no CI refresh step and none is needed.

// Generated-block delimiters for the open-seams table.
const (
	SeamsBeginMarker = "<!-- BEGIN GENERATED OPEN SEAMS — generated from server/internal/roadmap/registry.go; regenerate with: go test ./internal/roadmap/ -update -->"
	SeamsEndMarker   = "<!-- END GENERATED OPEN SEAMS -->"
)

// SeamsDocPath is engine-seams.md relative to this package's directory
// (go test's working directory).
const SeamsDocPath = "../../../docs/engine-seams.md"

// OpenSeams returns the seam items that are not implemented, in
// registry order.
func OpenSeams() []Item {
	var out []Item
	for _, it := range items {
		if it.Kind == KindSeam && it.Status != StatusImplemented {
			out = append(out, it)
		}
	}
	return out
}

// SeamsBlock renders the open-seams table, markers included.
// Deterministic: same registry, same bytes.
func SeamsBlock() string {
	var b strings.Builder
	b.WriteString(SeamsBeginMarker + "\n\n")
	b.WriteString("| Seam | Status | What is missing | Cards waiting | Unblocks | Tracked |\n")
	b.WriteString("|---|---|---|---|---:|---|\n")
	for _, it := range OpenSeams() {
		waiting := "—"
		if len(it.Waiting) > 0 {
			waiting = strings.Join(it.Waiting, "; ")
		}
		unblocks := "—"
		if it.Unblocks > 0 {
			unblocks = fmt.Sprint(it.Unblocks)
		}
		tracked := it.Tracked
		if tracked == "" && it.Issue > 0 {
			tracked = fmt.Sprintf("#%d", it.Issue)
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n",
			cell(it.Name), it.Status, cell(it.EngineNotes), cell(waiting), unblocks, cell(tracked))
	}
	b.WriteString("\n" + SeamsEndMarker)
	return b.String()
}

// cell makes text safe inside one markdown table cell: a newline ends
// the row and a bare pipe ends the cell, even inside a code span.
func cell(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.ReplaceAll(s, "|", `\|`)
}

// SpliceSeams replaces the generated block inside doc. ok is false when
// the markers are missing or out of order.
func SpliceSeams(doc, block string) (string, bool) {
	start := strings.Index(doc, SeamsBeginMarker)
	end := strings.Index(doc, SeamsEndMarker)
	if start < 0 || end < start {
		return doc, false
	}
	return doc[:start] + block + doc[end+len(SeamsEndMarker):], true
}

// ExtractSeams returns the generated block currently in doc, markers
// included.
func ExtractSeams(doc string) (string, bool) {
	start := strings.Index(doc, SeamsBeginMarker)
	end := strings.Index(doc, SeamsEndMarker)
	if start < 0 || end < start {
		return "", false
	}
	return doc[start : end+len(SeamsEndMarker)], true
}
