package lobby

// bugreport_kind_test.go — report kinds (ADR 0017 §8): the label the
// issue lands with, the artifacts each kind collects, and the rule
// that a label problem may never cost the report.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// statusErr is an upstream error carrying an HTTP status, the way
// github.APIError does. Declared here rather than importing the
// github package — the point of the unexported statusCoder interface
// is that the BugReporter seam stays one method wide.
type statusErr struct {
	status int
	msg    string
}

func (e *statusErr) Error() string   { return e.msg }
func (e *statusErr) StatusCode() int { return e.status }

// labelRejectingReporter refuses any labelled create with the given
// status and accepts the unlabelled retry — GitHub's behaviour when
// the repo has no such label, or the token may not apply it.
type labelRejectingReporter struct {
	recordingReporter
	status int
	calls  int
}

func (r *labelRejectingReporter) CreateIssue(ctx context.Context, title, body string, labels []string) (string, int, error) {
	r.calls++
	if len(labels) > 0 {
		return "", 0, &statusErr{status: r.status, msg: "github: create issue: status 422: Validation Failed"}
	}
	return r.recordingReporter.CreateIssue(ctx, title, body, labels)
}

// countingFailureReporter always fails, counting attempts.
type countingFailureReporter struct {
	err   error
	calls int
}

func (r *countingFailureReporter) CreateIssue(context.Context, string, string, []string) (string, int, error) {
	r.calls++
	return "", 0, r.err
}

func TestBugReportKindSelectsLabel(t *testing.T) {
	for _, tc := range []struct {
		name      string
		payload   map[string]any
		wantLabel string
		wantNoun  string
	}{
		// The field is absent, not empty: this is literally what a
		// client built before the picker existed sends.
		{"omitted kind is a bug", map[string]any{"title": "it broke"}, "bug", "bug"},
		{"explicit bug", map[string]any{"title": "it broke", "kind": "bug"}, "bug", "bug"},
		{"idea is enhancement", map[string]any{"title": "wider stack panel", "kind": "idea"}, "enhancement", "idea"},
		{"question", map[string]any{"title": "why did that die", "kind": "question"}, "question", "question"},
		{"case and space insensitive", map[string]any{"title": "x", "kind": "  Idea "}, "enhancement", "idea"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rep := &recordingReporter{}
			srv, tok := newBugReportStack(t, rep)
			resp := postBugReport(t, srv, tok, tc.payload)
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("status = %d, want 201", resp.StatusCode)
			}
			var out struct {
				Label string `json:"label"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if out.Label != tc.wantLabel {
				t.Errorf("response label = %q, want %q", out.Label, tc.wantLabel)
			}
			if len(rep.labels) != 1 || len(rep.labels[0]) != 1 || rep.labels[0][0] != tc.wantLabel {
				t.Fatalf("labels = %v, want [%s]", rep.labels, tc.wantLabel)
			}
			// The prefix is deliberately NOT varied by kind: every
			// issue this feature has ever filed carries it and people
			// filter on it.
			if !strings.HasPrefix(rep.titles[0], "[in-app] ") {
				t.Errorf("title = %q, want the stable [in-app] prefix", rep.titles[0])
			}
			// The body records the kind too, so the issue stays
			// self-describing if the label is stripped in triage.
			want := "- Kind: " + tc.wantNoun + " (label `" + tc.wantLabel + "`)"
			if !strings.Contains(rep.bodies[0], want) {
				t.Errorf("body missing %q\n----\n%s", want, rep.bodies[0])
			}
		})
	}
}

func TestBugReportRejectsUnknownKind(t *testing.T) {
	rep := &recordingReporter{}
	srv, tok := newBugReportStack(t, rep)
	resp := postBugReport(t, srv, tok, map[string]any{"title": "x", "kind": "wontfix"})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	// The 400 names the options rather than leaving a client guessing.
	b, _ := io.ReadAll(resp.Body)
	for _, want := range []string{"bug", "idea", "question"} {
		if !strings.Contains(string(b), want) {
			t.Errorf("error body %q should list %q", b, want)
		}
	}
	if len(rep.titles) != 0 {
		t.Fatalf("nothing should have been filed, got %d", len(rep.titles))
	}
}

// TestBugReportKindGatesTheClientLog: the log is attached for the
// kinds that can use it and dropped for the one that can't, whatever
// the client sends.
func TestBugReportKindGatesTheClientLog(t *testing.T) {
	entries := []map[string]any{
		{"at": 1789223415250, "kind": "sent", "text": "action cast_spell id=8f2a"},
	}
	for _, tc := range []struct {
		kind    string
		wantLog bool
	}{
		{"bug", true},
		{"question", true},
		{"idea", false},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			rep := &recordingReporter{}
			srv, tok := newBugReportStack(t, rep)
			resp := postBugReport(t, srv, tok, map[string]any{
				"title": "x", "kind": tc.kind, "log": entries,
			})
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("status = %d, want 201", resp.StatusCode)
			}
			got := strings.Contains(rep.bodies[0], "action cast_spell id=8f2a")
			if got != tc.wantLog {
				t.Errorf("log in body = %v, want %v\n----\n%s", got, tc.wantLog, rep.bodies[0])
			}
		})
	}
}

// TestBugReportKindGatesPinnedArtifacts is decision 2 of ADR 0017
// §8: an idea keeps its screenshot and pins nothing, a question pins
// the cheap game log but not the replay, a bug pins both. Pinning a
// tens-of-MiB replay to "the stack panel should be wider" is waste.
func TestBugReportKindGatesPinnedArtifacts(t *testing.T) {
	for _, tc := range []struct {
		kind        string
		wantReplay  bool
		wantGameLog bool
	}{
		{"bug", true, true},
		{"question", false, true},
		{"idea", false, false},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			rep := &recordingReporter{}
			srv, l, a, _, dumpDir := attachStack(t, rep)
			meta, err := l.Create("FNM")
			if err != nil {
				t.Fatalf("create game: %v", err)
			}
			seedGameLog(t, l, meta.ID)
			writeReplay(t, dumpDir, meta.ID, "{\"seq\":1}\n{\"seq\":2}\n")
			tok := playerSession(t, a, meta.ID)

			resp := postMultipartReport(t, srv, tok, map[string]any{
				"title":   "x",
				"kind":    tc.kind,
				"context": map[string]any{"game_id": meta.ID.String(), "turn": 4},
			}, pngBytes(16))
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusCreated {
				b, _ := io.ReadAll(resp.Body)
				t.Fatalf("status = %d body = %s", resp.StatusCode, b)
			}
			body := rep.bodies[0]
			if got := strings.Contains(body, "Pinned replay:"); got != tc.wantReplay {
				t.Errorf("pinned replay = %v, want %v\n----\n%s", got, tc.wantReplay, body)
			}
			if got := strings.Contains(body, "Pinned game log:"); got != tc.wantGameLog {
				t.Errorf("pinned game log = %v, want %v\n----\n%s", got, tc.wantGameLog, body)
			}
			// A kind that wants no replay must not advertise the live
			// route either — that reads as "we tried and failed".
			if got := strings.Contains(body, "/replay` (admin; not pinned"); got && !tc.wantReplay {
				t.Errorf("non-bug kind should not mention the live replay route\n----\n%s", body)
			}
			// Screenshots ride along whatever the kind: a picture is
			// the cheapest artifact and the most useful for an idea.
			if !strings.Contains(body, "**Screenshots**") {
				t.Errorf("screenshots missing for kind %q\n----\n%s", tc.kind, body)
			}
		})
	}
}

// TestBugReportRefilesUnlabelledWhenGitHubRejectsTheLabel is the
// no-silent-failure contract: a label the repo doesn't have costs the
// label, never the report, and it lands in the server log.
func TestBugReportRefilesUnlabelledWhenGitHubRejectsTheLabel(t *testing.T) {
	for _, status := range []int{http.StatusUnprocessableEntity, http.StatusForbidden} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			rep := &labelRejectingReporter{status: status}
			srv, tok, logged := newBugReportStackLogged(t, rep)
			resp := postBugReport(t, srv, tok, map[string]any{"title": "wider panel", "kind": "idea"})
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("status = %d, want 201 — a label must never cost the report", resp.StatusCode)
			}
			var out struct {
				Number int    `json:"number"`
				Label  string `json:"label"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if out.Number != 7 {
				t.Errorf("number = %d, want the filed issue", out.Number)
			}
			// No label is claimed, because none landed.
			if out.Label != "" {
				t.Errorf("label = %q, want empty — the label was rejected", out.Label)
			}
			if rep.calls != 2 {
				t.Errorf("CreateIssue calls = %d, want 2 (labelled, then unlabelled)", rep.calls)
			}
			if len(rep.labels) != 1 || len(rep.labels[0]) != 0 {
				t.Errorf("accepted call carried labels %v, want none", rep.labels)
			}
			if got := logged.String(); !strings.Contains(got, "labels rejected by GitHub") ||
				!strings.Contains(got, "enhancement") || !strings.Contains(got, "level=ERROR") {
				t.Errorf("the failure must be visible in the log, got:\n%s", got)
			}
		})
	}
}

// TestBugReportDoesNotRetryWhenDeliveryFailed: a timeout or a 5xx may
// mean GitHub created the issue anyway, so the unlabelled retry is
// confined to the statuses that mean it definitively did not. A
// duplicate issue is worse than a 502 the reporter can act on.
func TestBugReportDoesNotRetryWhenDeliveryFailed(t *testing.T) {
	for name, err := range map[string]error{
		"no status": errors.New("github: create issue: context deadline exceeded"),
		"upstream 5xx": &statusErr{
			status: http.StatusBadGateway,
			msg:    "github: create issue: status 502",
		},
	} {
		t.Run(name, func(t *testing.T) {
			rep := &countingFailureReporter{err: err}
			srv, tok := newBugReportStack(t, rep)
			resp := postBugReport(t, srv, tok, map[string]any{"title": "x", "kind": "bug"})
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusBadGateway {
				t.Fatalf("status = %d, want 502", resp.StatusCode)
			}
			if rep.calls != 1 {
				t.Errorf("CreateIssue calls = %d, want 1 — no retry on a failed delivery", rep.calls)
			}
		})
	}
}
