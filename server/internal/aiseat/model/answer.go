package model

import (
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
)

// answer.go turns a model's reply into a move, or into nothing
// (#2196).
//
// # The number and the words
//
// The reply carries two names for one move: the number from the list
// and that entry's label copied beside it. They are not equally
// reliable. Measured against qwen3.6:35b-a3b on the probe's window
// (five moves, 0..4), half the replies were `{"index": 5, ...}`: one
// past the end. The model is not counting from one — its in-range
// answers number correctly — it is reaching for a card in its hand
// that the list does not offer and giving it the next free number.
// Once in that sample the words named a listed move ("Cycling
// Skyray", which is 1) and the number did not.
//
// So the label decides when the two disagree, but only when it names
// exactly ONE listed entry:
//
//   - The number names a move and the label agrees, or there is no
//     label: the number (PickIndex, PickUnlabelled).
//   - The number is not a move, or names a different one, and the
//     label names exactly one listed entry: the label
//     (PickLabelRescued, PickLabelCorrected). The label is the model
//     saying what it wants in the move's own words; the number is a
//     pointer it had to look up, and the lookup is the step that was
//     measured going wrong.
//   - The label names no single listed entry — nothing at all, or a
//     line that appears twice (two Mountains) — so it cannot pick
//     between candidates. The number is taken when it names a move
//     (PickLabelMismatch), which is exactly what happened before
//     labels existed, and the window falls back when it does not
//     (FallbackOutOfRange).
//
// Neither name can invent a move. The number is bounds-checked
// against Input.Moves and the label is compared, after normalising
// case and whitespace, with the labels of the entries the model was
// SHOWN and nothing else. A label that is not one of those selects
// nothing. ADR 0033 §1's closed list is the only source of an answer.

// Answer is one reply, resolved against the list the model was shown.
type Answer struct {
	// Parsed is the number the reply wrote. Meaningful only when Err
	// is nil.
	Parsed int
	// Move is the label the reply copied, as written. Empty when it
	// copied none.
	Move string
	// Why is the reply's one-line reason.
	Why string
	// Err is why the reply could not be parsed at all. Fallback is
	// FallbackMalformed whenever it is set.
	Err error

	// Index is the move the reply resolved to: an index into
	// Input.Moves. Meaningful only when Fallback is empty.
	Index int
	// Pick is how it resolved — one of the Pick* constants — and
	// empty when it did not.
	Pick string
	// Fallback is FallbackMalformed or FallbackOutOfRange when the
	// reply named no move, and empty when it named one.
	Fallback string
}

// ResolveAnswer parses a reply and resolves it against shown, the
// entries the prompt listed (Request.Choices), with moves the length
// of Input.Moves. It is the funnel's own resolution, exported so the
// tools that score a raw reply — `boteval probe` — classify it
// exactly as a live seat would.
func ResolveAnswer(text string, shown []Choice, moves int) Answer {
	r, err := parseAnswer(text)
	if err != nil {
		return Answer{Err: err, Fallback: FallbackMalformed}
	}
	a := Answer{Parsed: r.index, Move: r.move, Why: r.why}
	inRange := r.index >= 0 && r.index < moves
	named, unique := labelNames(r.move, shown)
	switch {
	case inRange && normLabel(r.move) == "":
		a.Index, a.Pick = r.index, PickUnlabelled
	case inRange && labelAgrees(r.index, r.move, shown):
		a.Index, a.Pick = r.index, PickIndex
	case unique && inRange:
		a.Index, a.Pick = named, PickLabelCorrected
	case unique:
		a.Index, a.Pick = named, PickLabelRescued
	case inRange:
		a.Index, a.Pick = r.index, PickLabelMismatch
	default:
		// -1 lands here too, deliberately. It is aiseat.Decline on
		// the Go side, but the model was never told that and "none of
		// these" is not one of the things it is being asked. Declining
		// belongs to Layer B, whose answer this falls back to anyway.
		a.Fallback = FallbackOutOfRange
	}
	return a
}

// labelAgrees reports whether label is the shown entry at index i.
// An index that was not shown has no label to agree with.
func labelAgrees(i int, label string, shown []Choice) bool {
	want := normLabel(label)
	for _, c := range shown {
		if c.Index == i {
			return normLabel(c.Label) == want
		}
	}
	return false
}

// labelNames finds the one shown entry label names. ok is false when
// it names none, or more than one (two identical lines, such as two
// untapped Mountains): the words cannot choose between those, so they
// choose nothing.
func labelNames(label string, shown []Choice) (int, bool) {
	want := normLabel(label)
	if want == "" {
		return 0, false
	}
	found, n := 0, 0
	for _, c := range shown {
		if normLabel(c.Label) == want {
			found = c.Index
			n++
		}
	}
	return found, n == 1
}

// normLabel is the comparison form of a label: case folded, runs of
// whitespace collapsed, and the packaging a model may copy along with
// the text taken off — the list's own "N: " prefix and the fallback
// marker. Nothing else is forgiven: a label that differs in a word
// names a different move, or none.
func normLabel(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if i := strings.Index(s, fallbackMarker); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	// The answers note the prompt prints after a label (ADR 0142).
	if i := strings.LastIndex(s, " ("+boardtext.AnswersPrefix); i >= 0 && strings.HasSuffix(s, ")") {
		s = strings.TrimSpace(s[:i])
	}
	if colon := strings.IndexByte(s, ':'); colon > 0 {
		if _, err := strconv.Atoi(s[:colon]); err == nil {
			s = strings.TrimSpace(s[colon+1:])
		}
	}
	return strings.ToLower(s)
}
