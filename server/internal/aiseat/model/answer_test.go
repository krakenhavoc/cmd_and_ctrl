package model

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// answer_test.go pins #2196: a reply carries a number and the label
// beside it, and when they disagree the label decides only if it
// names exactly one listed move. Neither can select a move that was
// not shown.

// probeWindow is the shape of the window that found the bug: the
// probe's turn-4 upkeep, five moves, two of them the same line.
func probeWindow() []Choice {
	return []Choice{
		{0, "Pass priority"},
		{1, "Scrounging Skyray: Cycling {2} ({2}, Discard this card: Draw a card.)"},
		{2, "Mountain: Add {R}"},
		{3, "Mountain: Add {R}"},
		{4, "Steam Vents: Add {U} or {R}"},
	}
}

func TestResolveAnswer(t *testing.T) {
	skyray := "Scrounging Skyray: Cycling {2} ({2}, Discard this card: Draw a card.)"
	cases := []struct {
		name     string
		reply    string
		index    int
		pick     string
		fallback string
	}{
		{name: "number and label agree", reply: `{"index": 1, "move": "` + skyray + `", "why": "dig"}`, index: 1, pick: PickIndex},
		{name: "a bare number", reply: `4`, index: 4, pick: PickUnlabelled},
		{name: "the old shape, no label", reply: `{"index": 0, "why": "hold"}`, index: 0, pick: PickUnlabelled},

		// The measured failure: one past the end, with the words
		// naming a listed move.
		{name: "out of range, rescued by a unique label", reply: `{"index": 5, "move": "` + skyray + `", "why": "Cycling Skyray"}`, index: 1, pick: PickLabelRescued},
		{name: "negative, rescued by a unique label", reply: `{"index": -1, "move": "Pass priority"}`, index: 0, pick: PickLabelRescued},
		{name: "case and whitespace are forgiven", reply: `{"index": 9, "move": "  steam vents:   add {u} OR {r} "}`, index: 4, pick: PickLabelRescued},
		{name: "the list's own number prefix is forgiven", reply: `{"index": 7, "move": "4: Steam Vents: Add {U} or {R}"}`, index: 4, pick: PickLabelRescued},
		{name: "the fallback marker copied along is forgiven", reply: `{"index": 7, "move": "Pass priority   ` + fallbackMarker + `"}`, index: 0, pick: PickLabelRescued},

		// The label names a listed move the number does not: the
		// label wins, because it is unique.
		{name: "a valid number, a different unique label", reply: `{"index": 4, "move": "Pass priority"}`, index: 0, pick: PickLabelCorrected},

		// The label cannot choose, so the number decides when it can.
		{name: "a valid number, a label that names nothing", reply: `{"index": 4, "move": "Cast Big Score"}`, index: 4, pick: PickLabelMismatch},
		{name: "a valid number, a label on two lines that is not this one", reply: `{"index": 4, "move": "Mountain: Add {R}"}`, index: 4, pick: PickLabelMismatch},
		{name: "a valid number on one of two identical lines", reply: `{"index": 3, "move": "Mountain: Add {R}"}`, index: 3, pick: PickIndex},

		// …and the window falls back when it cannot.
		{name: "out of range with a label that names nothing", reply: `{"index": 5, "move": "Cast Big Score", "why": "Cast Big Score"}`, fallback: FallbackOutOfRange},
		{name: "out of range with an ambiguous label", reply: `{"index": 5, "move": "Mountain: Add {R}"}`, fallback: FallbackOutOfRange},
		{name: "out of range with no label", reply: `{"index": 5, "why": "Cast Big Score"}`, fallback: FallbackOutOfRange},
		{name: "a near miss is not a match", reply: `{"index": 5, "move": "Scrounging Skyray: Cycling {2}"}`, fallback: FallbackOutOfRange},
		{name: "not an answer", reply: `I would cast Big Score.`, fallback: FallbackMalformed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := ResolveAnswer(tc.reply, probeWindow(), 5)
			if a.Fallback != tc.fallback {
				t.Fatalf("fallback = %q, want %q (answer %+v)", a.Fallback, tc.fallback, a)
			}
			if tc.fallback != "" {
				if a.Pick != "" {
					t.Errorf("a reply that named no move has pick %q", a.Pick)
				}
				return
			}
			if a.Index != tc.index || a.Pick != tc.pick {
				t.Errorf("resolved to %d by %q, want %d by %q", a.Index, a.Pick, tc.index, tc.pick)
			}
		})
	}
}

// A label is looked up among the SHOWN entries only. A real move the
// cap dropped from the prompt cannot be reached by its words, because
// the model was never shown them.
func TestALabelNeverSelectsAMoveThatWasNotShown(t *testing.T) {
	shown := []Choice{{0, "Pass priority"}, {2, "Cast Bear"}}
	// Move 1 exists ("Cast Lightning Bolt") but was not listed.
	a := ResolveAnswer(`{"index": 9, "move": "Cast Lightning Bolt"}`, shown, 3)
	if a.Fallback != FallbackOutOfRange {
		t.Errorf("an unshown label resolved: %+v", a)
	}
}

// --- through the funnel, with a FakeClient ---------------------------

func decideTraced(t *testing.T, p *Policy, in aiseat.Input) (aiseat.Decision, aiseat.Trace) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	d, tr, err := p.DecideTraced(ctx, in)
	if err != nil {
		t.Fatalf("DecideTraced: %v", err)
	}
	return d, tr
}

// The measured failure, end to end: a number one past the end of the
// list and a label that names a listed move. The move lands, it is
// Layer C's, and it is counted as rescued.
func TestAnOutOfRangeIndexIsRescuedByItsLabel(t *testing.T) {
	b := &stubB{index: 0, reason: "pass"}
	p := testPolicy(t, AlwaysText(`{"index": 3, "move": "Cast Bear", "why": "develop"}`), b, nil)
	d, tr := decideTraced(t, p, castWindow())
	if d.Index != 2 {
		t.Fatalf("index = %d, want 2 — the move the label names", d.Index)
	}
	if !strings.Contains(d.Reason, "develop") {
		t.Errorf("reason = %q, want the model's own words", d.Reason)
	}
	st := p.Stats()
	if st.ByLayer[LayerC] != 1 || st.ByPick[PickLabelRescued] != 1 {
		t.Errorf("layers %v, picks %v; want one Layer C window, label-rescued", st.ByLayer, st.ByPick)
	}
	if n := st.ByFallback[FallbackOutOfRange]; n != 0 {
		t.Errorf("a rescued reply was also counted out of range (%d)", n)
	}
	if tr.ParsedIndex == nil || *tr.ParsedIndex != 3 {
		t.Errorf("ParsedIndex = %v, want what the model wrote, 3", tr.ParsedIndex)
	}
	if tr.ModelIndex == nil || *tr.ModelIndex != 2 || tr.Pick != PickLabelRescued || tr.ParsedMove != "Cast Bear" {
		t.Errorf("trace = model index %v, pick %q, move %q; want 2, %q, %q",
			tr.ModelIndex, tr.Pick, tr.ParsedMove, PickLabelRescued, "Cast Bear")
	}
	if ps := p.PolicyStats(); ps.ByPick[PickLabelRescued] != 1 {
		t.Errorf("PolicyStats.ByPick = %v; the projection lost the pick", ps.ByPick)
	}
}

// A valid number and a label that uniquely names a DIFFERENT listed
// move: the label wins. It is the model's statement of intent in the
// move's own words; the number is a lookup, and the lookup is the
// step that was measured going wrong.
func TestALabelThatUniquelyNamesAnotherMoveOverridesTheNumber(t *testing.T) {
	b := &stubB{index: 0, reason: "pass"}
	p := testPolicy(t, AlwaysText(`{"index": 1, "move": "Cast Bear", "why": "develop"}`), b, nil)
	d, tr := decideTraced(t, p, castWindow())
	if d.Index != 2 {
		t.Fatalf("index = %d, want 2 — the label names Cast Bear, uniquely", d.Index)
	}
	if st := p.Stats(); st.ByPick[PickLabelCorrected] != 1 {
		t.Errorf("picks = %v, want one %q", st.ByPick, PickLabelCorrected)
	}
	if tr.Pick != PickLabelCorrected {
		t.Errorf("trace pick = %q", tr.Pick)
	}
}

// A valid number and a label that names nothing listed: the number is
// taken, exactly as before labels existed, and the disagreement is
// counted.
func TestALabelThatNamesNothingLeavesAValidNumberAlone(t *testing.T) {
	b := &stubB{index: 0, reason: "pass"}
	p := testPolicy(t, AlwaysText(`{"index": 1, "move": "Cast Big Score", "why": "cards"}`), b, nil)
	d, _ := decideTraced(t, p, castWindow())
	if d.Index != 1 {
		t.Fatalf("index = %d, want the model's 1", d.Index)
	}
	if st := p.Stats(); st.ByPick[PickLabelMismatch] != 1 || st.ByLayer[LayerC] != 1 {
		t.Errorf("picks %v, layers %v", st.ByPick, st.ByLayer)
	}
}

// An out-of-range number with a label that names no listed move is
// what it always was: not a move. Layer B plays.
func TestAnUnknownLabelFallsBack(t *testing.T) {
	b := &stubB{index: 0, reason: "the heuristic's pick"}
	p := testPolicy(t, AlwaysText(`{"index": 3, "move": "Cast Big Score", "why": "cards"}`), b, nil)
	d, tr := decideTraced(t, p, castWindow())
	if d.Index != 0 || d.Reason != "the heuristic's pick" {
		t.Fatalf("decision = %+v, want Layer B's", d)
	}
	st := p.Stats()
	if st.ByFallback[FallbackOutOfRange] != 1 || st.ByLayer[LayerB] != 1 {
		t.Errorf("fallbacks %v, layers %v", st.ByFallback, st.ByLayer)
	}
	if len(st.ByPick) != 0 {
		t.Errorf("a reply that named no move was counted as a pick: %v", st.ByPick)
	}
	if tr.ModelIndex != nil || tr.Pick != "" {
		t.Errorf("trace says the reply resolved: model index %v, pick %q", tr.ModelIndex, tr.Pick)
	}
	if tr.ParsedMove != "Cast Big Score" {
		t.Errorf("trace lost the label the model wrote: %q", tr.ParsedMove)
	}
}

// Request.Choices is the shown list and nothing else, and the schema
// the OpenAI transport builds from it enumerates exactly those
// indices. With a cap, the dropped moves' indices are absent.
func TestTheSchemaEnumIsTheShownIndicesOnly(t *testing.T) {
	var got Request
	fake := &FakeClient{Reply: func(_ int, req Request) (Response, error) {
		got = req
		return Response{Text: `{"index": 0, "move": "Pass priority", "why": "hold"}`}, nil
	}}
	b := &stubB{index: 5, reason: "the heuristic's pick"}
	p := testPolicy(t, fake, b, func(c *Config) { c.MaxCandidates = 3 })

	in := castWindow()
	for _, l := range []string{"Cast Ogre", "Cast Drake", "Cast Wurm"} {
		in.Moves = append(in.Moves, mv(legal.KindCast, l, ""))
	}
	// Six moves, a cap of three: the pass, Layer B's pick (5), and
	// one more by enumeration order.
	_, _ = decideTraced(t, p, in)

	if len(got.Choices) != 3 {
		t.Fatalf("Choices = %+v, want the three shown moves", got.Choices)
	}
	listed := map[int]bool{}
	for _, c := range got.Choices {
		listed[c.Index] = true
		if c.Label != in.Moves[c.Index].Label {
			t.Errorf("choice %d carries label %q, want %q", c.Index, c.Label, in.Moves[c.Index].Label)
		}
		if !strings.Contains(got.User, c.Label) {
			t.Errorf("choice %d %q is not in the prompt the model saw", c.Index, c.Label)
		}
	}
	if !listed[0] || !listed[5] {
		t.Errorf("the pass and Layer B's pick must be among the choices: %+v", got.Choices)
	}

	raw, err := json.Marshal(choiceFormat(got.Choices))
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		JSONSchema struct {
			Schema struct {
				Properties struct {
					Index struct {
						Enum []int `json:"enum"`
					} `json:"index"`
				} `json:"properties"`
			} `json:"schema"`
		} `json:"json_schema"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	enum := f.JSONSchema.Schema.Properties.Index.Enum
	if len(enum) != len(got.Choices) {
		t.Fatalf("enum = %v, want exactly the %d shown indices", enum, len(got.Choices))
	}
	for i, n := range enum {
		if n != got.Choices[i].Index {
			t.Errorf("enum[%d] = %d, want %d", i, n, got.Choices[i].Index)
		}
		if n < 0 || n >= len(in.Moves) {
			t.Errorf("enum carries %d, which is not a move", n)
		}
	}
}
