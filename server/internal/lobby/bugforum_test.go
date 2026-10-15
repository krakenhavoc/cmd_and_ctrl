package lobby

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/bugstore"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/github"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// The real client is what main hands the mirror.
var _ BugIssueReader = (*github.Client)(nil)

// fakeForum is the Discord half: it records every call and answers from
// canned errors, so nothing here touches a network.
type fakeForum struct {
	mu         sync.Mutex
	created    []discord.BugThread
	createErrs []error // consumed one per CreateThread call, nil past the end
	threads    []discord.ForumThread
	listErr    error
	notes      map[string][]string
	archived   []string
	archiveErr error
	calls      int
}

func (f *fakeForum) CreateThread(_ context.Context, _ *http.Client, t discord.BugThread) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if len(f.createErrs) > 0 {
		err := f.createErrs[0]
		f.createErrs = f.createErrs[1:]
		if err != nil {
			return "", err
		}
	}
	f.created = append(f.created, t)
	return "9000", nil
}

func (f *fakeForum) ActiveThreads(context.Context, *http.Client) ([]discord.ForumThread, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]discord.ForumThread(nil), f.threads...), f.listErr
}

func (f *fakeForum) PostNote(_ context.Context, _ *http.Client, id, text string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.notes == nil {
		f.notes = map[string][]string{}
	}
	f.notes[id] = append(f.notes[id], text)
	return nil
}

func (f *fakeForum) Archive(_ context.Context, _ *http.Client, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.archiveErr != nil {
		return f.archiveErr
	}
	f.archived = append(f.archived, id)
	return nil
}

func (f *fakeForum) createdThreads() []discord.BugThread {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]discord.BugThread(nil), f.created...)
}

// fakeIssues answers issue state from a table. A number with an entry in
// errs answers that error instead.
type fakeIssues struct {
	state map[int]string
	errs  map[int]error
}

func (f fakeIssues) GetIssue(_ context.Context, n int) (github.Issue, error) {
	if err := f.errs[n]; err != nil {
		return github.Issue{}, err
	}
	return github.Issue{Number: n, State: f.state[n]}, nil
}

func quietBugForum(api bugForumAPI, issues BugIssueReader) *BugForum {
	f := newBugForum(api, issues, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	f.wait = func(time.Duration) {}
	return f
}

// forumStack is the HTTP stack with the mirror wired. The lobby is
// returned so a test can seat the reporter at a table.
func forumStack(t *testing.T, forum *BugForum) (*httptest.Server, *Lobby, *auth.MemoryAuthenticator) {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	l := NewLobby(ws.NewRoomManager(log, ""))
	a := auth.NewMemoryAuthenticator()
	srv := httptest.NewServer(Handler(Config{
		Lobby: l, Auth: a, AdminToken: "shared-admin-token",
		BugReporter: &recordingReporter{}, BugForum: forum,
		Log: log,
	}))
	t.Cleanup(srv.Close)
	return srv, l, a
}

func sessionFor(t *testing.T, a *auth.MemoryAuthenticator, p auth.Principal) string {
	t.Helper()
	tok, _, err := a.Issue(context.Background(), p, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func lobbyReporter() auth.Principal {
	return auth.Principal{Role: auth.RoleIdentified, DiscordID: "501", DiscordUsername: "alice", Name: "Alice"}
}

func fileReport(t *testing.T, srv *httptest.Server, tok string, payload map[string]any) int {
	t.Helper()
	resp := postBugReport(t, srv, tok, payload)
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}

func TestBugForumMirrorsAFiledReport(t *testing.T) {
	api := &fakeForum{}
	forum := quietBugForum(api, fakeIssues{})
	srv, _, a := forumStack(t, forum)

	code := fileReport(t, srv, sessionFor(t, a, lobbyReporter()), map[string]any{
		"title":       "Deck page crashes",
		"description": "Opening it fails. token=abcdef1234567890 is in my url",
		"context":     map[string]any{"page": "decks"},
	})
	if code != http.StatusCreated {
		t.Fatalf("status %d, want 201", code)
	}
	forum.Wait()

	got := api.createdThreads()
	if len(got) != 1 {
		t.Fatalf("threads = %+v, want one", got)
	}
	th := got[0]
	if th.IssueNumber != 7 || th.IssueURL != "https://github.com/o/r/issues/7" || th.Kind != "bug" {
		t.Errorf("issue/kind = %d %q %q", th.IssueNumber, th.IssueURL, th.Kind)
	}
	if th.Title != "Deck page crashes" {
		t.Errorf("title = %q: the thread carries the report's title, not the tracker's [in-app] prefix", th.Title)
	}
	if th.ReporterID != "501" {
		t.Errorf("reporter = %q, want the Discord id", th.ReporterID)
	}
	if strings.Contains(th.Description, "abcdef1234567890") {
		t.Errorf("the thread carries a credential: %q", th.Description)
	}
}

func TestBugForumGuestReporterIsNotMentioned(t *testing.T) {
	api := &fakeForum{}
	forum := quietBugForum(api, fakeIssues{})
	srv, _, a := forumStack(t, forum)
	guest := auth.Principal{Role: auth.RoleIdentified, Name: "Guest"}
	if code := fileReport(t, srv, sessionFor(t, a, guest), map[string]any{"title": "x"}); code != http.StatusCreated {
		t.Fatalf("status %d", code)
	}
	forum.Wait()
	if got := api.createdThreads(); len(got) != 1 || got[0].ReporterID != "" {
		t.Fatalf("threads = %+v, want one with no reporter to mention", got)
	}
}

func TestBugForumNeverMirrorsAReportFromAPrivateTable(t *testing.T) {
	for _, tc := range []struct {
		name   string
		setup  func(l *Lobby) (gameID uuid.UUID)
		useCtx bool // report names the game in its context; otherwise only the session is bound to it
		want   int
	}{
		{"an ordinary invite-only table", func(l *Lobby) uuid.UUID { return makeTable(t, l, false, false) }, true, 0},
		{"the same table, found from the session alone", func(l *Lobby) uuid.UUID { return makeTable(t, l, false, false) }, false, 0},
		{"a table the host put on the #lfg board", func(l *Lobby) uuid.UUID { return makeTable(t, l, true, false) }, true, 1},
		{"a practice table, even with the lfg flag", func(l *Lobby) uuid.UUID { return makeTable(t, l, true, true) }, true, 0},
		{"a table the lobby no longer holds", func(*Lobby) uuid.UUID { return uuid.New() }, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := &fakeForum{}
			forum := quietBugForum(api, fakeIssues{})
			srv, l, a := forumStack(t, forum)
			id := tc.setup(l)
			p := lobbyReporter()
			p.GameID = id
			payload := map[string]any{"title": "Stack is wrong"}
			if tc.useCtx {
				payload["context"] = map[string]any{"game_id": id.String(), "turn": 3}
			}
			if code := fileReport(t, srv, sessionFor(t, a, p), payload); code != http.StatusCreated {
				t.Fatalf("status %d: the report itself is always filed", code)
			}
			forum.Wait()
			if got := len(api.createdThreads()); got != tc.want {
				t.Errorf("threads = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestBugForumRefusesAGameIDItCannotRead(t *testing.T) {
	api := &fakeForum{}
	forum := quietBugForum(api, fakeIssues{})
	srv, _, a := forumStack(t, forum)
	code := fileReport(t, srv, sessionFor(t, a, lobbyReporter()), map[string]any{
		"title": "x", "context": map[string]any{"game_id": "not-a-uuid"},
	})
	if code != http.StatusCreated {
		t.Fatalf("status %d", code)
	}
	forum.Wait()
	if got := len(api.createdThreads()); got != 0 {
		t.Errorf("threads = %d: a game id that does not parse cannot be shown to be public", got)
	}
}

// makeTable opens a table and sets the flags a host's #lfg post and the
// practice route would.
func makeTable(t *testing.T, l *Lobby, lfg, practice bool) uuid.UUID {
	t.Helper()
	meta, err := l.Create("Table")
	if err != nil {
		t.Fatal(err)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.games[meta.ID]
	e.meta.LFG = lfg
	e.meta.Practice = practice
	return meta.ID
}

func TestBugForumOffByDefault(t *testing.T) {
	// A nil mirror is the feature off, and a report is filed as before.
	srv, _, a := forumStack(t, nil)
	if code := fileReport(t, srv, sessionFor(t, a, lobbyReporter()), map[string]any{"title": "x"}); code != http.StatusCreated {
		t.Fatalf("status %d", code)
	}
	var f *BugForum
	f.Mirror(bugMirror{Number: 1})
	f.Sweep(context.Background())
	f.Start(context.Background())
	f.Wait()
}

func TestBugForumDiscordFailureNeverFailsTheReport(t *testing.T) {
	api := &fakeForum{createErrs: []error{discord.ErrBugForumForbidden}}
	forum := quietBugForum(api, fakeIssues{})
	srv, _, a := forumStack(t, forum)
	if code := fileReport(t, srv, sessionFor(t, a, lobbyReporter()), map[string]any{"title": "x"}); code != http.StatusCreated {
		t.Fatalf("status %d: a Discord outage must not cost the reporter their report", code)
	}
	forum.Wait()
	if api.calls != 1 || len(api.createdThreads()) != 0 {
		t.Errorf("calls %d, threads %d: a 403 is not retried", api.calls, len(api.createdThreads()))
	}
}

func TestBugForumRetriesOnlyTransientFailures(t *testing.T) {
	api := &fakeForum{createErrs: []error{&discord.RateLimitError{RetryAfter: time.Second}, discord.ErrVoiceUnavailable}}
	forum := quietBugForum(api, fakeIssues{})
	forum.Mirror(bugMirror{Number: 3, URL: "https://x/3"})
	forum.Wait()
	if api.calls != 3 || len(api.createdThreads()) != 1 {
		t.Errorf("calls %d, threads %d: want two transient failures retried into one thread", api.calls, len(api.createdThreads()))
	}

	api = &fakeForum{createErrs: []error{discord.ErrVoiceUnavailable, discord.ErrVoiceUnavailable, discord.ErrVoiceUnavailable, nil}}
	forum = quietBugForum(api, fakeIssues{})
	forum.Mirror(bugMirror{Number: 4})
	forum.Wait()
	if api.calls != bugForumMaxAttempts || len(api.createdThreads()) != 0 {
		t.Errorf("calls %d, threads %d: want the attempts capped at %d", api.calls, len(api.createdThreads()), bugForumMaxAttempts)
	}
}

func TestBugForumDropsAMirrorWhenTooManyAreInFlight(t *testing.T) {
	release := make(chan struct{})
	api := &blockingForum{release: release}
	forum := quietBugForum(api, fakeIssues{})
	for i := 0; i < bugForumMaxInFlight+3; i++ {
		forum.Mirror(bugMirror{Number: i + 1})
	}
	// Wait for the slots to be held before releasing them.
	deadline := time.Now().Add(5 * time.Second)
	for api.started() < bugForumMaxInFlight && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	close(release)
	forum.Wait()
	if got := api.started(); got != bugForumMaxInFlight {
		t.Errorf("%d creates started, want exactly the %d slots", got, bugForumMaxInFlight)
	}
}

type blockingForum struct {
	fakeForum
	release chan struct{}
	n       int
	nmu     sync.Mutex
}

func (b *blockingForum) CreateThread(ctx context.Context, c *http.Client, t discord.BugThread) (string, error) {
	b.nmu.Lock()
	b.n++
	b.nmu.Unlock()
	<-b.release
	return b.fakeForum.CreateThread(ctx, c, t)
}

func (b *blockingForum) started() int {
	b.nmu.Lock()
	defer b.nmu.Unlock()
	return b.n
}

func TestBugForumSweepClosesOnlyClosedAndRemovedIssues(t *testing.T) {
	api := &fakeForum{threads: []discord.ForumThread{
		{ID: "1", Name: "#7 still open"},
		{ID: "2", Name: "#8 fixed"},
		{ID: "3", Name: "#9 deleted spam"},
		{ID: "4", Name: "#10 github is down"},
	}}
	issues := fakeIssues{
		state: map[int]string{7: "open", 8: "closed"},
		errs:  map[int]error{9: &statusErr{status: 404, msg: "not found"}, 10: &statusErr{status: 502, msg: "bad gateway"}},
	}
	forum := quietBugForum(api, issues)
	forum.Sweep(context.Background())

	if got := strings.Join(api.archived, ","); got != "2,3" {
		t.Fatalf("archived = %q, want 2,3", got)
	}
	if n := api.notes["2"]; len(n) != 1 || !strings.Contains(n[0], "#8 is closed") {
		t.Errorf("note for 2 = %v", n)
	}
	if n := api.notes["3"]; len(n) != 1 || !strings.Contains(n[0], "#9 was removed") {
		t.Errorf("note for 3 = %v", n)
	}
	if len(api.notes["1"])+len(api.notes["4"]) != 0 {
		t.Errorf("an open issue, or one GitHub could not answer for, must be left alone: %v", api.notes)
	}
}

func TestBugForumSweepNotesOnceWhenArchiveFails(t *testing.T) {
	api := &fakeForum{threads: []discord.ForumThread{{ID: "2", Name: "#8 fixed"}}, archiveErr: discord.ErrBugForumForbidden}
	forum := quietBugForum(api, fakeIssues{state: map[int]string{8: "closed"}})
	forum.Sweep(context.Background())
	forum.Sweep(context.Background())
	if n := len(api.notes["2"]); n != 1 {
		t.Fatalf("the closing note was posted %d times, want once", n)
	}
	api.archiveErr = nil
	forum.Sweep(context.Background())
	if len(api.archived) != 1 || api.archived[0] != "2" || len(api.notes["2"]) != 1 {
		t.Errorf("archived %v, notes %v: the archive is retried without a second note", api.archived, api.notes)
	}
}

func TestBugForumSweepSurvivesADiscordListingError(t *testing.T) {
	api := &fakeForum{listErr: errors.New("boom")}
	forum := quietBugForum(api, fakeIssues{})
	forum.Sweep(context.Background())
	if len(api.archived) != 0 {
		t.Errorf("archived %v on a failed listing", api.archived)
	}
}

func TestBugForumSweepChecksAtMostBugForumMaxChecks(t *testing.T) {
	api := &fakeForum{}
	state := map[int]string{}
	for i := 1; i <= bugForumMaxChecks+10; i++ {
		api.threads = append(api.threads, discord.ForumThread{ID: uuid.NewString(), Name: "#" + strconv.Itoa(i) + " t"})
		state[i] = "closed"
	}
	forum := quietBugForum(api, fakeIssues{state: state})
	forum.Sweep(context.Background())
	if len(api.archived) != bugForumMaxChecks {
		t.Errorf("archived %d, want %d (the cap on GitHub reads in one sweep)", len(api.archived), bugForumMaxChecks)
	}
}

func TestMirrorBugReportCarriesTheScreenshotLinks(t *testing.T) {
	api := &fakeForum{}
	forum := quietBugForum(api, fakeIssues{})
	in := bugIssue{Desc: "see picture", Images: []bugstore.Image{{Name: "a.png", URL: "https://cmd.example/a.png"}, {Name: "b.png", URL: "https://cmd.example/b.png"}}}
	mirrorBugReport(Config{BugForum: forum}, lobbyReporter(), nil, bugKinds["bug"], "Picture", in, 12, "https://x/12")
	forum.Wait()
	got := api.createdThreads()
	if len(got) != 1 || len(got[0].Images) != 2 || got[0].Images[1] != "https://cmd.example/b.png" || got[0].IssueNumber != 12 {
		t.Fatalf("threads = %+v", got)
	}
}

func TestBugForumSweepLeavesAReopenedThreadAloneUntilItsIssueReopens(t *testing.T) {
	api := &fakeForum{threads: []discord.ForumThread{{ID: "2", Name: "#8 fixed"}}}
	issues := fakeIssues{state: map[int]string{8: "closed"}}
	forum := quietBugForum(api, issues)
	forum.Sweep(context.Background())
	// The reporter answers late, so Discord lists the thread as open
	// again while the issue is still closed: no second note, no re-archive.
	forum.Sweep(context.Background())
	forum.Sweep(context.Background())
	if len(api.notes["2"]) != 1 || len(api.archived) != 1 {
		t.Fatalf("notes %v archived %v: a live conversation on a closed issue must not be closed over and over", api.notes, api.archived)
	}
	// The issue is reopened and closed again: that is a new closing.
	issues.state[8] = "open"
	forum.Sweep(context.Background())
	issues.state[8] = "closed"
	forum.Sweep(context.Background())
	if len(api.notes["2"]) != 2 || len(api.archived) != 2 {
		t.Fatalf("notes %v archived %v: want a second closing after the reopen", api.notes, api.archived)
	}
}
