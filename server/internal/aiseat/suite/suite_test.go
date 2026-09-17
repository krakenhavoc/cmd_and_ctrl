package suite

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// suite_test.go is the heuristic regression gate, and it runs on
// EVERY CI run rather than behind AISEAT_GAME_TESTS.
//
// That is the whole reason the suite stores frozen Inputs: re-deciding
// a window is pure computation over JSON — no engine, no goroutines,
// no wall clock — so twenty-odd real positions cost milliseconds. A
// gated position here is a pinned invariant ("do not chump-block at
// 40 life") or a pinned bug, and it can never come back quietly.

const positionsDir = "testdata/positions"

// gatePolicy is the tier the gate applies to. Model tiers need an
// endpoint and run through `boteval suite run`.
const gatePolicy = string(tiers.Heuristic)

func loadSuite(t *testing.T) []Position {
	t.Helper()
	positions, err := Load(positionsDir)
	if err != nil {
		t.Fatalf("load the suite: %v", err)
	}
	if len(positions) == 0 {
		t.Fatalf("no positions in %s", positionsDir)
	}
	return positions
}

// TestEveryPositionStillLoads is the stale-matcher alarm. Load derives
// every accept and reject index by MATCHING against the frozen move
// list, so a change to legal's labels or kinds that silently
// re-pointed a label at a different move fails here, loudly, naming
// the position — which is the property the whole design rests on.
func TestEveryPositionStillLoads(t *testing.T) {
	positions := loadSuite(t)
	gated, labelled := 0, 0
	tags := map[string]int{}
	for _, p := range positions {
		if p.Note == "" {
			t.Errorf("%s: a position with no note is a label nobody can review", p.ID)
		}
		if len(p.Tags) == 0 {
			t.Errorf("%s: no tags", p.ID)
		}
		if p.Source.Reviewer == "" {
			t.Errorf("%s: no reviewer in source", p.ID)
		}
		if p.Labelled() {
			labelled++
		}
		if p.Gates(gatePolicy) {
			gated++
			if !p.Labelled() {
				t.Errorf("%s is gated but carries no label; a gate on an unanswered question always passes", p.ID)
			}
		}
		for _, tag := range p.Tags {
			tags[tag]++
		}
	}
	t.Logf("%d positions, %d labelled, %d gated on %q, tags %v", len(positions), labelled, gated, gatePolicy, tags)
	// ADR 0052 §3: the suite is a heuristic regression gate, which it
	// is not if nothing is gated.
	if gated < 5 {
		t.Errorf("only %d positions are gated on %q; the suite is meant to pin at least a handful of invariants", gated, gatePolicy)
	}
}

// TestHeuristicOnTheLabelledSuite is the gate itself.
func TestHeuristicOnTheLabelledSuite(t *testing.T) {
	positions := loadSuite(t)
	pol, err := tiers.New(tiers.Heuristic, tiers.Options{})
	if err != nil {
		t.Fatalf("build the heuristic tier: %v", err)
	}
	started := time.Now()
	rep := Run(context.Background(), positions, pol, RunOptions{MaxThink: 2 * time.Second})
	t.Logf("ran in %v\n%s", time.Since(started), rep.Markdown())

	if rep.Policy != gatePolicy {
		t.Fatalf("the heuristic tier reports its name as %q; the gate keys on %q", rep.Policy, gatePolicy)
	}
	if fails := rep.GateFailures(gatePolicy); len(fails) > 0 {
		t.Errorf("%d gated position(s) failed:", len(fails))
		for _, f := range fails {
			t.Errorf("  %s", f)
		}
	}
	// Nothing here should reach a model, error or time out: the
	// heuristic tier is Layer A over Layer B and nothing else.
	if rep.Errors+rep.Timeouts+rep.Malformed+rep.OutOfRange > 0 {
		t.Errorf("the heuristic tier produced failures: %+v", rep)
	}
}

// TestFakeModelFailuresAreClassifiedByTheFunnel pins the three
// transport failures the harness exists to measure. Each one is a
// model tier wired to a fake client, and the suite must report the
// funnel's OWN classification — not "the heuristic underneath
// answered, so that counts as agreement", which is exactly how a
// broken endpoint would hide.
func TestFakeModelFailuresAreClassifiedByTheFunnel(t *testing.T) {
	positions := loadSuite(t)

	cases := []struct {
		name   string
		client model.Client
		count  func(Report) int
		want   Outcome
	}{
		{"malformed", model.AlwaysText("nope"), func(r Report) int { return r.Malformed }, OutcomeMalformed},
		{"out-of-range", model.AlwaysIndex(999), func(r Report) int { return r.OutOfRange }, OutcomeOutOfRange},
		{"timeout", model.NeverAnswers(), func(r Report) int { return r.Timeouts }, OutcomeTimeout},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// A tiny call budget, so the timeout case costs
			// milliseconds rather than the tier's real deadline.
			cfg := model.DefaultConfig()
			cfg.Reserve, cfg.MinBudget, cfg.MaxCall = time.Millisecond, time.Millisecond, 10*time.Millisecond
			pol, err := tiers.New(tiers.Assisted, tiers.Options{Client: tc.client, Config: &cfg})
			if err != nil {
				t.Fatalf("build the assisted tier: %v", err)
			}
			rep := Run(context.Background(), positions, pol, RunOptions{MaxThink: 100 * time.Millisecond, Parallel: 8})
			if got := tc.count(rep); got != rep.Labelled {
				t.Errorf("%d/%d positions came back %s; want all of them\n%s", got, rep.Labelled, tc.want, rep.Markdown())
			}
			if rep.Agree != 0 {
				t.Errorf("a model that never answered agreed on %d positions; the funnel's fallback is being counted as the model's answer", rep.Agree)
			}
		})
	}
}

// TestAStaleMatcherFailsLoad is the control for the design's central
// claim. The fixture's accept matcher names a move that is not in its
// frozen list; Load must refuse the whole directory and say which
// position is wrong.
func TestAStaleMatcherFailsLoad(t *testing.T) {
	_, err := Load("testdata/stale")
	if err == nil {
		t.Fatal("a position whose matcher matches nothing loaded cleanly")
	}
	for _, want := range []string{"stale-label", "STALE"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not mention %q: %v", want, err)
		}
	}
}

// TestUnlabelledPositionsAreSkipped: a harvested position is a
// question, not an answer, and counting it either way would corrupt
// the agreement number.
func TestUnlabelledPositionsAreSkipped(t *testing.T) {
	p := Position{ID: "fresh", V: PositionVersion, Input: twoMoveInput()}
	if err := p.prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	pol, err := tiers.New(tiers.Heuristic, tiers.Options{})
	if err != nil {
		t.Fatalf("tier: %v", err)
	}
	rep := Run(context.Background(), []Position{p}, pol, RunOptions{})
	if rep.Labelled != 0 || rep.Agree != 0 || rep.Positions != 1 {
		t.Fatalf("unlabelled position was graded: %+v", rep)
	}
	if rep.Results[0].Outcome != OutcomeSkipped {
		t.Fatalf("outcome %q, want %q", rep.Results[0].Outcome, OutcomeSkipped)
	}
}

// TestDeclineIsGradedAgainstTheLabel: declining is right on some
// windows and a stalled seat on others, and only the label knows
// which.
func TestDeclineIsGradedAgainstTheLabel(t *testing.T) {
	for _, ok := range []bool{true, false} {
		p := Position{
			ID: "d", V: PositionVersion, Input: twoMoveInput(),
			Expected: Expected{Accept: []Matcher{{Label: "Alpha"}}, DeclineOK: ok},
		}
		if err := p.prepare(); err != nil {
			t.Fatalf("prepare: %v", err)
		}
		rep := Run(context.Background(), []Position{p}, declinePolicy{}, RunOptions{})
		want := OutcomeDecline
		if ok {
			want = OutcomeAgree
		}
		if got := rep.Results[0].Outcome; got != want {
			t.Errorf("decline_ok=%v gave %q, want %q", ok, got, want)
		}
	}
}

func TestDeclineOnlyPositionIsLabelled(t *testing.T) {
	p := Position{
		ID: "decline-only", V: PositionVersion, Input: twoMoveInput(),
		Expected: Expected{DeclineOK: true},
	}
	if err := p.prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	rep := Run(context.Background(), []Position{p}, declinePolicy{}, RunOptions{})
	if rep.Labelled != 1 || rep.Agree != 1 || rep.Results[0].Outcome != OutcomeAgree {
		t.Fatalf("decline-only label was not graded: %+v", rep)
	}
}

// TestMatchersNameMovesRatherThanIndices covers the four matcher
// forms, including the params subset — the one that distinguishes two
// moves sharing a label.
func TestMatchersNameMovesRatherThanIndices(t *testing.T) {
	moves := []legal.Move{
		{Kind: legal.KindPass, Type: "pass_priority", Label: "Pass priority"},
		{Kind: legal.KindChoice, Type: "resolve_choice", Label: "Send it?: yes", Params: json.RawMessage(`{"choice_id":"x","apply":true}`)},
		{Kind: legal.KindChoice, Type: "resolve_choice", Label: "Send it?: no", Params: json.RawMessage(`{"choice_id":"x","apply":false}`)},
	}
	in := aiseat.Input{Moves: moves}
	cases := []struct {
		name    string
		matcher Matcher
		want    []int
	}{
		{"label", Matcher{Label: "Pass priority"}, []int{0}},
		{"label_re", Matcher{LabelRe: `^Send it\?: `}, []int{1, 2}},
		{"kind", Matcher{Kind: legal.KindChoice}, []int{1, 2}},
		{"type+params", Matcher{Type: "resolve_choice", ParamsSubset: json.RawMessage(`{"apply":true}`)}, []int{1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := Position{ID: "m", V: PositionVersion, Input: in, Expected: Expected{Accept: []Matcher{tc.matcher}}}
			if err := p.prepare(); err != nil {
				t.Fatalf("prepare: %v", err)
			}
			if got := p.Accept(); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("matched %v, want %v", got, tc.want)
			}
		})
	}

	// A matcher that sets two forms, or none, is a labelling mistake
	// and is refused rather than guessed at.
	for _, bad := range []Matcher{{}, {Label: "Pass priority", Kind: legal.KindPass}, {ParamsSubset: json.RawMessage(`{}`)}} {
		p := Position{ID: "bad", V: PositionVersion, Input: in, Expected: Expected{Accept: []Matcher{bad}}}
		if err := p.prepare(); err == nil {
			t.Errorf("matcher %+v was accepted", bad)
		}
	}
}

// TestPositionFilesRoundTrip pins the on-disk envelope against
// Position's own tags: a field added to one and not the other would
// silently stop being written, and the suite would lose labels.
func TestPositionFilesRoundTrip(t *testing.T) {
	positions := loadSuite(t)
	p := positions[0]
	dir := t.TempDir()
	path, err := p.Write(dir)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	back, err := LoadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	p.path, back.path = "", ""
	if !reflect.DeepEqual(p, back) {
		t.Fatalf("a position did not survive a round trip through the file format")
	}

	// The envelope must carry every JSON field Position declares.
	want := jsonTagSet(reflect.TypeOf(Position{}))
	got := jsonTagSet(reflect.TypeOf(envelope{}))
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Position writes %v but the envelope writes %v", want, got)
	}

	// And the frozen input is one line, which is what keeps a 25 KiB
	// board view out of the reviewer's way.
	blob, err := os.ReadFile(path) //nolint:gosec // t.TempDir
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	for _, line := range strings.Split(string(blob), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), `"input":`) {
			return
		}
	}
	t.Error(`the file has no single-line "input" key`)
}

func jsonTagSet(t reflect.Type) map[string]bool {
	out := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		out[strings.Split(tag, ",")[0]] = true
	}
	return out
}

// TestRenderPrintsThePromptAndTheMarkers is the labelling screen's
// smoke test. It also pins the reason Render rebuilds the prompt
// instead of reading one off the trace: these positions were played
// by heuristic seats and never saw a model, so there is no recorded
// prompt to print.
func TestRenderPrintsThePromptAndTheMarkers(t *testing.T) {
	positions := loadSuite(t)
	var p Position
	for _, cand := range positions {
		if cand.ID == "never-bolt-yourself" {
			p = cand
		}
	}
	if p.ID == "" {
		t.Skip("the fixture this test reads was renamed")
	}
	var sb strings.Builder
	if err := Render(&sb, p, model.DefaultConfig()); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := sb.String()
	for _, want := range []string{
		"=== never-bolt-yourself ===",
		"--- SYSTEM[0] ---",
		"--- USER ---",
		"--- MOVES (",
		"<- accept",
		"<- reject",
		"<- heuristic",
		p.Note,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("render is missing %q", want)
		}
	}
}

// --- helpers ---------------------------------------------------------

func twoMoveInput() aiseat.Input {
	return aiseat.Input{Moves: []legal.Move{
		{Kind: legal.KindCast, Type: "cast_spell", Label: "Alpha"},
		{Kind: legal.KindPass, Type: "pass_priority", Label: "Pass priority", AlwaysLegal: true},
	}}
}

// declinePolicy always declines, so the decline path can be graded
// without a board that produces one.
type declinePolicy struct{}

func (declinePolicy) Name() string { return "decliner" }

func (declinePolicy) Decide(context.Context, aiseat.Input) (aiseat.Decision, error) {
	return aiseat.Decision{Index: aiseat.Decline, Reason: "nothing here"}, nil
}
