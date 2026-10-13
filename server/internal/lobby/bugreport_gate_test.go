package lobby

// bugreport_gate_test.go — the throttle in front of GitHub (ADR 0017
// §10, #2979): the per-account and daily caps, the per-IP bucket that
// still stands, duplicates becoming +1 comments, the admin kill switch,
// the queue of reports kept instead of filed, and the label.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/bugstore"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

type gateIssue struct {
	title, body string
	labels      []string
}

type gateComment struct {
	number int
	body   string
}

// gateReporter is the fake GitHub for the throttle tests: it hands out
// consecutive issue numbers and records issues and comments, so a test
// can say "one issue, one comment".
type gateReporter struct {
	mu         sync.Mutex
	issues     []gateIssue
	comments   []gateComment
	createErr  error
	commentErr error
}

func (r *gateReporter) CreateIssue(_ context.Context, title, body string, labels []string) (string, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.createErr != nil {
		return "", 0, r.createErr
	}
	r.issues = append(r.issues, gateIssue{title, body, labels})
	n := 100 + len(r.issues)
	return "https://github.com/o/r/issues/" + itoa(n), n, nil
}

func (r *gateReporter) CreateComment(_ context.Context, number int, body string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.commentErr != nil {
		return "", r.commentErr
	}
	r.comments = append(r.comments, gateComment{number, body})
	return "https://github.com/o/r/issues/" + itoa(number) + "#c", nil
}

func (r *gateReporter) counts() (issues, comments int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.issues), len(r.comments)
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

// gateClock is a settable clock for the day and window rollovers.
type gateClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *gateClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *gateClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

type gateStack struct {
	srv   *httptest.Server
	auth  auth.Authenticator
	clock *gateClock
	admin string
}

// newGateStack builds the server. relax=false keeps the per-IP bucket
// real; the other tests relax it so only the gate is under test.
func newGateStack(t *testing.T, rep BugReporter, store *bugstore.Store, lim BugLimits, clock *gateClock, relax bool) *gateStack {
	t.Helper()
	if relax {
		t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	}
	if clock == nil {
		clock = &gateClock{t: time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)}
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := auth.NewMemoryAuthenticator()
	srv := httptest.NewServer(Handler(Config{
		Lobby:       NewLobby(ws.NewRoomManager(log, "")),
		Auth:        a,
		AdminToken:  "shared-admin-token",
		BugReporter: rep,
		BugStore:    store,
		BugLimits:   lim,
		Now:         clock.Now,
		Log:         log,
	}))
	t.Cleanup(srv.Close)
	st := &gateStack{srv: srv, auth: a, clock: clock}
	st.admin = adminSession(t, a)
	return st
}

// player issues a session for a fresh account.
func (s *gateStack) player(t *testing.T) string {
	t.Helper()
	return playerSession(t, s.auth, uuid.New())
}

func (s *gateStack) report(t *testing.T, tok string, payload map[string]any) (int, map[string]any) {
	t.Helper()
	resp := postBugReport(t, s.srv, tok, payload)
	defer func() { _ = resp.Body.Close() }()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func (s *gateStack) adminCall(t *testing.T, tok, method, path string, body any) (int, map[string]any) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, s.srv.URL+path, rd)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func (s *gateStack) queued(t *testing.T) []any {
	t.Helper()
	code, out := s.adminCall(t, s.admin, "GET", "/admin/bugreport", nil)
	if code != http.StatusOK {
		t.Fatalf("GET /admin/bugreport = %d", code)
	}
	q, _ := out["queued"].([]any)
	return q
}

func newTestBugStore(t *testing.T) *bugstore.Store {
	t.Helper()
	return bugstore.New(t.TempDir(), "https://cmd.example.test")
}

func TestBugGateAccountCapQueuesInsteadOfFiling(t *testing.T) {
	rep := &gateReporter{}
	s := newGateStack(t, rep, newTestBugStore(t), BugLimits{AccountMax: 2}, nil, true)
	tok := s.player(t)

	for i, title := range []string{"first problem", "second problem"} {
		if code, _ := s.report(t, tok, map[string]any{"title": title}); code != http.StatusCreated {
			t.Fatalf("report %d = %d, want 201", i, code)
		}
	}
	code, out := s.report(t, tok, map[string]any{"title": "third problem", "description": "it is still broken"})
	if code != http.StatusAccepted {
		t.Fatalf("third report = %d, want 202 (queued)", code)
	}
	if out["status"] != "queued" || out["url"] != nil {
		t.Errorf("queued response = %v; want status queued and no issue url", out)
	}
	if note, _ := out["note"].(string); !strings.Contains(note, "not been filed") {
		t.Errorf("note = %q; it must say the report was not filed", note)
	}
	if n, _ := rep.counts(); n != 2 {
		t.Errorf("GitHub saw %d issues, want 2", n)
	}

	q := s.queued(t)
	if len(q) != 1 {
		t.Fatalf("queued = %d, want 1", len(q))
	}
	item := q[0].(map[string]any)
	if item["reason"] != "account limit" || item["title"] != "[in-app] third problem" || item["report_id"] != out["report_id"] {
		t.Errorf("queued item = %v", item)
	}
	if body, _ := item["body"].(string); !strings.Contains(body, "it is still broken") {
		t.Errorf("queued body lost the description: %q", body)
	}

	// A different account is not held up by this one's window...
	if code, _ := s.report(t, s.player(t), map[string]any{"title": "someone else"}); code != http.StatusCreated {
		t.Errorf("other account = %d, want 201", code)
	}
	// ...and the window slides: an hour later the first account files again.
	s.clock.Advance(61 * time.Minute)
	if code, _ := s.report(t, tok, map[string]any{"title": "fourth problem"}); code != http.StatusCreated {
		t.Errorf("after the window = %d, want 201", code)
	}
}

func TestBugGateDailyCapQueuesAcrossAccounts(t *testing.T) {
	rep := &gateReporter{}
	s := newGateStack(t, rep, newTestBugStore(t), BugLimits{DailyCap: 2}, nil, true)

	for i := 0; i < 2; i++ {
		if code, _ := s.report(t, s.player(t), map[string]any{"title": "distinct " + itoa(i)}); code != http.StatusCreated {
			t.Fatalf("report %d = %d, want 201", i, code)
		}
	}
	code, out := s.report(t, s.player(t), map[string]any{"title": "one too many"})
	if code != http.StatusAccepted || out["status"] != "queued" {
		t.Fatalf("over the cap = %d %v, want 202 queued", code, out)
	}
	if n, _ := rep.counts(); n != 2 {
		t.Errorf("GitHub saw %d issues, want 2", n)
	}
	q := s.queued(t)
	if len(q) != 1 || q[0].(map[string]any)["reason"] != "daily cap" {
		t.Errorf("queue = %v, want one daily-cap entry", q)
	}

	// The cap is per UTC day.
	s.clock.Advance(24 * time.Hour)
	if code, _ := s.report(t, s.player(t), map[string]any{"title": "next day"}); code != http.StatusCreated {
		t.Errorf("next day = %d, want 201", code)
	}
}

func TestBugGatePerIPBucketStillEnforced(t *testing.T) {
	rep := &gateReporter{}
	// Limits far above the burst, so a 429 can only be the IP bucket.
	s := newGateStack(t, rep, newTestBugStore(t), BugLimits{AccountMax: 50, DailyCap: 50}, nil, false)
	tok := s.player(t)
	var codes []int
	for i := 0; i < 5; i++ {
		code, _ := s.report(t, tok, map[string]any{"title": "burst " + itoa(i)})
		codes = append(codes, code)
	}
	if codes[2] != http.StatusCreated || codes[3] != http.StatusTooManyRequests {
		t.Errorf("codes = %v; want the 4th request refused by the IP bucket", codes)
	}
}

func TestBugGateDuplicateTitleBecomesAComment(t *testing.T) {
	rep := &gateReporter{}
	s := newGateStack(t, rep, newTestBugStore(t), BugLimits{}, nil, true)

	code, first := s.report(t, s.player(t), map[string]any{"title": "Stack panel overflows"})
	if code != http.StatusCreated {
		t.Fatalf("first = %d", code)
	}
	code, second := s.report(t, s.player(t), map[string]any{
		"title": "  stack PANEL overflows!! ", "description": "same here, turn 9",
		"context": map[string]any{"game_id": "g-1", "turn": 9, "phase": "main1", "step": "main"},
	})
	if code != http.StatusOK || second["status"] != "commented" {
		t.Fatalf("duplicate = %d %v, want 200 commented", code, second)
	}
	if second["number"] != first["number"] || second["url"] != first["url"] {
		t.Errorf("duplicate pointed at %v, want the first issue %v", second, first)
	}
	issues, comments := rep.counts()
	if issues != 1 || comments != 1 {
		t.Fatalf("issues=%d comments=%d, want 1 and 1", issues, comments)
	}
	c := rep.comments[0]
	if c.number != int(first["number"].(float64)) || !strings.HasPrefix(c.body, "+1") || !strings.Contains(c.body, "same here, turn 9") {
		t.Errorf("comment = %+v", c)
	}

	// The same words as a different kind are a different issue.
	if code, out := s.report(t, s.player(t), map[string]any{"title": "Stack panel overflows", "kind": "idea"}); code != http.StatusCreated {
		t.Errorf("same title as an idea = %d %v, want a new issue", code, out)
	}
	// Outside the window the old issue no longer absorbs it.
	s.clock.Advance(25 * time.Hour)
	if code, _ := s.report(t, s.player(t), map[string]any{"title": "stack panel overflows"}); code != http.StatusCreated {
		t.Errorf("after the window = %d, want a new issue", code)
	}
}

func TestBugGateDuplicateStackBecomesAComment(t *testing.T) {
	rep := &gateReporter{}
	s := newGateStack(t, rep, newTestBugStore(t), BugLimits{}, nil, true)
	logWith := func(game string) []map[string]any {
		return []map[string]any{{"at": 1, "kind": "error", "text": "TypeError: Cannot read properties of undefined (reading 'zone') at game " + game + " line 40"}}
	}
	if code, _ := s.report(t, s.player(t), map[string]any{"title": "board goes blank", "log": logWith(uuid.NewString())}); code != http.StatusCreated {
		t.Fatalf("first failed")
	}
	code, out := s.report(t, s.player(t), map[string]any{"title": "totally different words", "log": logWith(uuid.NewString())})
	if code != http.StatusOK || out["status"] != "commented" {
		t.Fatalf("same error under another title = %d %v, want a +1", code, out)
	}
	// A short, generic error line is not a signature.
	short := []map[string]any{{"at": 1, "kind": "error", "text": "failed"}}
	s.report(t, s.player(t), map[string]any{"title": "generic a", "log": short})
	if code, _ := s.report(t, s.player(t), map[string]any{"title": "generic b", "log": short}); code != http.StatusCreated {
		t.Errorf("generic error merged unrelated reports: %d", code)
	}
}

func TestBugGateDeletedIssueFilesAFreshOne(t *testing.T) {
	rep := &gateReporter{}
	s := newGateStack(t, rep, newTestBugStore(t), BugLimits{}, nil, true)
	s.report(t, s.player(t), map[string]any{"title": "gone soon"})
	rep.commentErr = &statusErr{status: http.StatusNotFound, msg: "not found"}
	code, out := s.report(t, s.player(t), map[string]any{"title": "gone soon"})
	if code != http.StatusCreated || out["status"] != "filed" {
		t.Fatalf("after the issue was deleted = %d %v, want a new issue", code, out)
	}
	if issues, _ := rep.counts(); issues != 2 {
		t.Errorf("issues = %d, want 2", issues)
	}
}

func TestBugGateKillSwitchQueuesAndSurvivesARestart(t *testing.T) {
	rep := &gateReporter{}
	store := newTestBugStore(t)
	s := newGateStack(t, rep, store, BugLimits{}, nil, true)
	tok := s.player(t)

	if code, _ := s.adminCall(t, tok, "PUT", "/admin/bugreport/paused", map[string]bool{"paused": true}); code != http.StatusForbidden {
		t.Fatalf("player pausing = %d, want 403", code)
	}
	if code, _ := s.adminCall(t, tok, "GET", "/admin/bugreport", nil); code != http.StatusForbidden {
		t.Fatalf("player reading the queue = %d, want 403", code)
	}
	if code, out := s.adminCall(t, s.admin, "PUT", "/admin/bugreport/paused", map[string]bool{"paused": true}); code != http.StatusOK || out["paused"] != true {
		t.Fatalf("pause = %d %v", code, out)
	}
	code, out := s.report(t, tok, map[string]any{"title": "while paused"})
	if code != http.StatusAccepted || out["status"] != "queued" {
		t.Fatalf("paused report = %d %v, want 202 queued", code, out)
	}
	if note, _ := out["note"].(string); !strings.Contains(note, "paused") {
		t.Errorf("note = %q", note)
	}
	if issues, comments := rep.counts(); issues != 0 || comments != 0 {
		t.Fatalf("GitHub was written to while paused: %d issues, %d comments", issues, comments)
	}
	if q := s.queued(t); len(q) != 1 || q[0].(map[string]any)["reason"] != "paused" {
		t.Fatalf("queue = %v", q)
	}

	// A restart on the same data dir is still paused: the switch is not
	// a deploy-away-able flag.
	s2 := newGateStack(t, rep, store, BugLimits{}, nil, true)
	if code, out := s2.report(t, s2.player(t), map[string]any{"title": "after restart"}); code != http.StatusAccepted {
		t.Fatalf("after restart = %d %v, want still paused", code, out)
	}
	if _, st := s2.adminCall(t, s2.admin, "GET", "/admin/bugreport", nil); st["paused"] != true {
		t.Errorf("status = %v, want paused", st)
	}

	// Resuming files again, and the earlier reports stay queued.
	s2.adminCall(t, s2.admin, "PUT", "/admin/bugreport/paused", map[string]bool{"paused": false})
	if code, _ := s2.report(t, s2.player(t), map[string]any{"title": "after resume"}); code != http.StatusCreated {
		t.Errorf("after resume = %d, want 201", code)
	}
	if q := s2.queued(t); len(q) != 2 {
		t.Errorf("queue = %d, want the 2 paused reports kept", len(q))
	}
}

func TestBugGateDedupeIndexAndDailyCountSurviveARestart(t *testing.T) {
	rep := &gateReporter{}
	store := newTestBugStore(t)
	clock := &gateClock{t: time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)}
	s := newGateStack(t, rep, store, BugLimits{DailyCap: 1}, clock, true)
	s.report(t, s.player(t), map[string]any{"title": "before the restart"})

	s2 := newGateStack(t, rep, store, BugLimits{DailyCap: 1}, clock, true)
	if code, out := s2.report(t, s2.player(t), map[string]any{"title": "Before the restart"}); code != http.StatusOK || out["status"] != "commented" {
		t.Errorf("duplicate after restart = %d %v, want a +1 on the old issue", code, out)
	}
	if code, out := s2.report(t, s2.player(t), map[string]any{"title": "a new one"}); code != http.StatusAccepted {
		t.Errorf("new report after restart = %d %v, want the daily cap still counted", code, out)
	}
}

func TestBugGateDuplicateStillLandsAfterTheDailyCap(t *testing.T) {
	rep := &gateReporter{}
	s := newGateStack(t, rep, newTestBugStore(t), BugLimits{DailyCap: 1}, nil, true)
	s.report(t, s.player(t), map[string]any{"title": "the one issue"})
	if code, out := s.report(t, s.player(t), map[string]any{"title": "the one issue"}); code != http.StatusOK {
		t.Errorf("+1 at the cap = %d %v, want 200", code, out)
	}
}

func TestBugGateFailedFilingGivesTheAllowanceBack(t *testing.T) {
	rep := &gateReporter{createErr: &statusErr{status: http.StatusBadGateway, msg: "boom"}}
	s := newGateStack(t, rep, newTestBugStore(t), BugLimits{AccountMax: 1, DailyCap: 1}, nil, true)
	tok := s.player(t)
	if code, _ := s.report(t, tok, map[string]any{"title": "outage"}); code != http.StatusFailedDependency {
		t.Fatalf("during the outage = %d, want 424", code)
	}
	rep.createErr = nil
	if code, out := s.report(t, tok, map[string]any{"title": "outage"}); code != http.StatusCreated {
		t.Errorf("after the outage = %d %v; the failed attempt must not have used the account or daily slot", code, out)
	}
}

func TestBugGateWithoutAStoreCannotQueue(t *testing.T) {
	rep := &gateReporter{}
	s := newGateStack(t, rep, nil, BugLimits{AccountMax: 1}, nil, true)
	tok := s.player(t)
	s.report(t, tok, map[string]any{"title": "one"})
	if code, _ := s.report(t, tok, map[string]any{"title": "two"}); code != http.StatusTooManyRequests {
		t.Errorf("over the account cap with nowhere to queue = %d, want 429", code)
	}
	s.adminCall(t, s.admin, "PUT", "/admin/bugreport/paused", map[string]bool{"paused": true})
	if code, _ := s.report(t, s.player(t), map[string]any{"title": "three"}); code != http.StatusServiceUnavailable {
		t.Errorf("paused with nowhere to queue = %d, want 503", code)
	}
	if issues, _ := rep.counts(); issues != 1 {
		t.Errorf("issues = %d, want 1", issues)
	}
}

func TestBugGateEveryFiledIssueCarriesThePlayerLabel(t *testing.T) {
	rep := &gateReporter{}
	s := newGateStack(t, rep, newTestBugStore(t), BugLimits{}, nil, true)
	for _, kind := range []string{"bug", "idea", "question"} {
		if code, _ := s.report(t, s.player(t), map[string]any{"title": "kind " + kind, "kind": kind}); code != http.StatusCreated {
			t.Fatalf("%s = %d", kind, code)
		}
	}
	if len(rep.issues) != 3 {
		t.Fatalf("issues = %d", len(rep.issues))
	}
	for _, is := range rep.issues {
		if len(is.labels) != 2 || is.labels[1] != "player-report" {
			t.Errorf("labels = %v, want [<kind> player-report]", is.labels)
		}
	}
}

func TestBugGateQueuedReportKeepsItsScreenshots(t *testing.T) {
	rep := &gateReporter{}
	store := newTestBugStore(t)
	s := newGateStack(t, rep, store, BugLimits{AccountMax: 1}, nil, true)
	tok := s.player(t)
	s.report(t, tok, map[string]any{"title": "first"})
	resp := postMultipartReport(t, s.srv, tok, map[string]any{"title": "with a picture"}, pngBytes(16))
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("queued multipart = %d, want 202", resp.StatusCode)
	}
	q := s.queued(t)
	if len(q) != 1 {
		t.Fatalf("queue = %d", len(q))
	}
	body := q[0].(map[string]any)["body"].(string)
	if !strings.Contains(body, "https://cmd.example.test/bugreport/att/") {
		t.Errorf("queued body lost its screenshot link:\n%s", body)
	}
}

func TestBugDedupeKeys(t *testing.T) {
	bug := bugKinds["bug"]
	idea := bugKinds["idea"]
	a := bugDedupeKeys(bug, "  Hello,   WORLD!! ", nil)
	b := bugDedupeKeys(bug, "hello world", nil)
	if len(a) != 1 || a[0] != b[0] {
		t.Errorf("normalisation: %v vs %v", a, b)
	}
	if bugDedupeKeys(idea, "hello world", nil)[0] == b[0] {
		t.Error("kind must be part of the key")
	}
	if k := bugDedupeKeys(idea, "x", []bugLogEntry{{Kind: "error", Text: "TypeError: something quite long went wrong"}}); len(k) != 1 {
		t.Errorf("a kind that carries no log must have no stack key: %v", k)
	}
}
