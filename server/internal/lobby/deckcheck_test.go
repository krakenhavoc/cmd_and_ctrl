package lobby

// Tests for ADR 0095's deck routes: POST /deck-coverage and POST
// /deck-requests. The deck fetcher and GitHub are fakes throughout;
// nothing here reaches the network.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckcoverage"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckrequests"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/github"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

const (
	needyDeckURL = "https://moxfield.com/decks/needy01"
	cleanDeckURL = "https://archidekt.com/decks/424242"
)

// fakeDeckLinks serves canned decks by canonical URL and counts
// fetches.
type fakeDeckLinks struct {
	mu      sync.Mutex
	decks   map[string]fakeDeck
	errs    map[string]error
	fetches map[string]int
}

type fakeDeck struct {
	name    string
	entries []deck.Entry
}

func (f *fakeDeckLinks) fetch(_ context.Context, rawURL string) (string, []deck.Entry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fetches[rawURL]++
	if err, ok := f.errs[rawURL]; ok {
		return "", nil, err
	}
	d, ok := f.decks[rawURL]
	if !ok {
		return "", nil, fmt.Errorf("%w: moxfield", deck.ErrDeckNotFound)
	}
	return d.name, d.entries, nil
}

func (f *fakeDeckLinks) count(rawURL string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.fetches[rawURL]
}

// fakeFiler stands in for GitHub.
type fakeFiler struct {
	mu           sync.Mutex
	rejectLabels bool
	next         int
	issues       map[int]*fakeIssue
	comments     []fakeComment
}

type fakeIssue struct {
	title, body string
	labels      []string
	state       string
}

type fakeComment struct {
	number int
	body   string
}

func (f *fakeFiler) CreateIssue(_ context.Context, title, body string, labels []string) (string, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.rejectLabels && len(labels) > 0 {
		return "", 0, &statusErr{status: http.StatusUnprocessableEntity, msg: "github: create issue: status 422: labels"}
	}
	f.next++
	f.issues[f.next] = &fakeIssue{title: title, body: body, labels: labels, state: "open"}
	return fmt.Sprintf("https://github.com/o/r/issues/%d", f.next), f.next, nil
}

func (f *fakeFiler) GetIssue(_ context.Context, number int) (github.Issue, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	iss, ok := f.issues[number]
	if !ok {
		return github.Issue{}, &statusErr{status: http.StatusNotFound, msg: "github: get issue: status 404"}
	}
	return github.Issue{Number: number, State: iss.state, HTMLURL: fmt.Sprintf("https://github.com/o/r/issues/%d", number)}, nil
}

func (f *fakeFiler) CreateComment(_ context.Context, number int, body string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.comments = append(f.comments, fakeComment{number: number, body: body})
	return fmt.Sprintf("https://github.com/o/r/issues/%d#c%d", number, len(f.comments)), nil
}

func (f *fakeFiler) snapshot() (issues map[int]fakeIssue, comments []fakeComment) {
	f.mu.Lock()
	defer f.mu.Unlock()
	issues = map[int]fakeIssue{}
	for n, iss := range f.issues {
		issues[n] = *iss
	}
	return issues, append([]fakeComment(nil), f.comments...)
}

func (f *fakeFiler) close(number int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.issues[number].state = "closed"
}

type deckStack struct {
	srv    *httptest.Server
	auth   auth.Authenticator
	users  *users.SQLStore
	store  *deckrequests.SQLStore
	filer  *fakeFiler
	source *fakeDeckLinks
	tc     deckcoverage.TestingCards
}

// newDeckStack wires a database-backed stack with a fake deck source
// and a fake GitHub. relax turns the rate limiters off, which every
// test but the limiter ones wants. configure may adjust the Config.
func newDeckStack(t *testing.T, relax bool, configure func(*Config)) *deckStack {
	t.Helper()
	if relax {
		t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	}
	idx, tc := deckcoverage.TestingIndex(t)
	d := openTestDB(t, t.TempDir())
	s := &deckStack{
		auth:  auth.NewMemoryAuthenticator(),
		users: users.NewSQLStore(d, nil),
		store: deckrequests.NewSQLStore(d),
		filer: &fakeFiler{issues: map[int]*fakeIssue{}},
		source: &fakeDeckLinks{
			decks: map[string]fakeDeck{
				needyDeckURL: {name: "Needy Deck", entries: []deck.Entry{
					{Name: tc.Commander, Count: 1, IsCommander: true},
					{Name: tc.Manual, Count: 1},
					{Name: tc.Unreviewed, Count: 1},
					{Name: tc.Caveated, Count: 1},
					{Name: tc.Automated, Count: 1},
					{Name: tc.Basic, Count: 95},
				}},
				cleanDeckURL: {name: "Clean Deck", entries: []deck.Entry{
					{Name: tc.Commander, Count: 1, IsCommander: true},
					{Name: tc.Caveated, Count: 1},
					{Name: tc.Automated, Count: 1},
					{Name: tc.Basic, Count: 97},
				}},
			},
			errs:    map[string]error{},
			fetches: map[string]int{},
		},
		tc: tc,
	}
	cfg := Config{
		Lobby:            NewLobby(ws.NewRoomManager(quietLogger(), "")),
		Auth:             s.auth,
		AdminToken:       "shared-admin-token",
		Cards:            idx,
		Users:            s.users,
		DeckRequests:     s.store,
		DeckRequestFiler: s.filer,
		FetchDeck:        s.source.fetch,
		Log:              quietLogger(),
	}
	if configure != nil {
		configure(&cfg)
	}
	s.srv = httptest.NewServer(Handler(cfg))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *deckStack) token(t *testing.T, p auth.Principal) string {
	t.Helper()
	tok, _, err := s.auth.Issue(context.Background(), p, time.Hour)
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	return tok
}

func (s *deckStack) adminToken(t *testing.T) string {
	return s.token(t, auth.Principal{Role: auth.RoleAdmin})
}

// discordUser signs a person in with Discord and returns their session.
func (s *deckStack) discordUser(t *testing.T, snowflake, name string) string {
	t.Helper()
	u, err := s.users.UpsertFromDiscord(context.Background(),
		discord.User{ID: snowflake, Username: strings.ToLower(name), GlobalName: name}, "", "identify")
	if err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	return s.token(t, auth.Principal{Role: auth.RoleIdentified, UserID: u.ID, DiscordID: snowflake, DiscordGlobalName: name, Name: name})
}

func (s *deckStack) post(t *testing.T, path, tok string, body any) (int, string) {
	t.Helper()
	resp := postJSON(t, s.srv, path, tok, body)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func (s *deckStack) request(t *testing.T, tok string, body any) (int, deckRequestResponse, string) {
	t.Helper()
	code, raw := s.post(t, "/deck-requests", tok, body)
	var out deckRequestResponse
	_ = json.Unmarshal([]byte(raw), &out)
	return code, out, raw
}

// --- POST /deck-coverage ---

func TestDeckCoverageIsPublic(t *testing.T) {
	s := newDeckStack(t, true, nil)
	code, raw := s.post(t, "/deck-coverage", "", map[string]string{"url": needyDeckURL + "/some-slug?x=1"})
	if code != http.StatusOK {
		t.Fatalf("status %d: %s", code, raw)
	}
	var r deckcoverage.Report
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.DeckName != "Needy Deck" || r.Source != "moxfield" || r.DeckKey != "moxfield:needy01" || r.SourceURL != needyDeckURL {
		t.Errorf("header = %q %q %q %q", r.DeckName, r.Source, r.DeckKey, r.SourceURL)
	}
	if r.Counts[deckcoverage.Manual] != 1 || r.Counts[deckcoverage.Unreviewed] != 1 ||
		r.Counts[deckcoverage.Caveats] != 1 || r.Counts[deckcoverage.Automated] != 1 || r.Counts[deckcoverage.NoEffect] != 2 {
		t.Errorf("counts = %v", r.Counts)
	}
	// The fetcher saw the canonical link, never the caller's query.
	if n := s.source.count(needyDeckURL); n != 1 {
		t.Errorf("canonical URL fetched %d times", n)
	}
	// Names, oracle IDs, buckets and caveats only: no art, no oracle
	// text, no image route.
	for _, banned := range []string{"image", "scryfall", "/cards/", ".jpg", ".png", "oracle_text", "Do what this card does"} {
		if strings.Contains(raw, banned) {
			t.Errorf("public report contains %q: %s", banned, raw)
		}
	}
}

func TestDeckCoverageText(t *testing.T) {
	s := newDeckStack(t, true, nil)
	list := fmt.Sprintf("1 %s *CMDR*\n1 %s\n98 %s\n", s.tc.Commander, s.tc.Manual, s.tc.Basic)
	code, raw := s.post(t, "/deck-coverage", "", map[string]string{"text": list})
	if code != http.StatusOK {
		t.Fatalf("status %d: %s", code, raw)
	}
	var r deckcoverage.Report
	_ = json.Unmarshal([]byte(raw), &r)
	if r.Source != "text" || r.DeckKey != "" || r.Counts[deckcoverage.Manual] != 1 {
		t.Errorf("report = %+v", r)
	}
}

func TestDeckCoverageCachesByDeckKey(t *testing.T) {
	s := newDeckStack(t, true, nil)
	for _, u := range []string{needyDeckURL, "https://www.moxfield.com/decks/needy01/slug", needyDeckURL + "?tab=x"} {
		if code, raw := s.post(t, "/deck-coverage", "", map[string]string{"url": u}); code != http.StatusOK {
			t.Fatalf("%s: status %d: %s", u, code, raw)
		}
	}
	if n := s.source.count(needyDeckURL); n != 1 {
		t.Errorf("three checks of one deck fetched it %d times, want 1", n)
	}
}

func TestDeckCoverageIsRateLimitedPerIP(t *testing.T) {
	s := newDeckStack(t, false, nil)
	for i := 0; i < 3; i++ {
		if code, raw := s.post(t, "/deck-coverage", "", map[string]string{"url": needyDeckURL}); code != http.StatusOK {
			t.Fatalf("check %d: status %d: %s", i+1, code, raw)
		}
	}
	if code, raw := s.post(t, "/deck-coverage", "", map[string]string{"url": needyDeckURL}); code != http.StatusTooManyRequests {
		t.Fatalf("fourth check: status %d, want 429: %s", code, raw)
	}
	// The bot's admin session has its own bucket: every guild member
	// behind one loopback address must not share three checks.
	if code, raw := s.post(t, "/deck-coverage", s.adminToken(t), map[string]string{"url": needyDeckURL}); code != http.StatusOK {
		t.Fatalf("admin check after the public bucket ran dry: status %d: %s", code, raw)
	}
}

func TestDeckCoverageErrors(t *testing.T) {
	s := newDeckStack(t, true, nil)
	private := "https://moxfield.com/decks/secret"
	s.source.errs[private] = fmt.Errorf("%w: moxfield", deck.ErrDeckPrivate)
	blocked := "https://moxfield.com/decks/walled"
	s.source.errs[blocked] = fmt.Errorf("%w: moxfield", deck.ErrUpstreamBlocked)

	for _, tc := range []struct {
		body       map[string]string
		wantStatus int
		wantCode   string
	}{
		{map[string]string{"url": "https://example.com/decks/1"}, http.StatusBadRequest, deck.CodeUnknownSource},
		{map[string]string{"url": "https://moxfield.com/decks/nope"}, http.StatusNotFound, deck.CodeDeckNotFound},
		{map[string]string{"url": private}, http.StatusUnprocessableEntity, deck.CodeDeckPrivate},
		{map[string]string{"url": blocked}, http.StatusBadGateway, deck.CodeUpstreamBlocked},
		{map[string]string{}, http.StatusBadRequest, ""},
		{map[string]string{"url": needyDeckURL, "text": "1 Forest"}, http.StatusBadRequest, ""},
	} {
		code, raw := s.post(t, "/deck-coverage", "", tc.body)
		if code != tc.wantStatus {
			t.Errorf("%v: status %d, want %d: %s", tc.body, code, tc.wantStatus, raw)
			continue
		}
		var body struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}
		_ = json.Unmarshal([]byte(raw), &body)
		if body.Error == "" || body.Code != tc.wantCode {
			t.Errorf("%v: body %s, want code %q and a message", tc.body, raw, tc.wantCode)
		}
	}
	if n := s.source.count("https://example.com/decks/1"); n != 0 {
		t.Error("an unsupported host was fetched")
	}
}

// --- POST /deck-requests ---

func TestDeckRequestFilesAnIssue(t *testing.T) {
	s := newDeckStack(t, true, nil)
	tok := s.discordUser(t, "111111111111111111", "Alice")

	code, out, raw := s.request(t, tok, map[string]string{"url": needyDeckURL})
	if code != http.StatusCreated || out.Status != deckRequestFiled || out.IssueNumber != 1 ||
		out.IssueURL != "https://github.com/o/r/issues/1" || out.Report == nil {
		t.Fatalf("status %d: %s", code, raw)
	}

	issues, _ := s.filer.snapshot()
	iss := issues[1]
	if iss.title != "[deck-request] Needy Deck" {
		t.Errorf("title = %q", iss.title)
	}
	if strings.Join(iss.labels, ",") != "enhancement,deck-request" {
		t.Errorf("labels = %v", iss.labels)
	}
	manualOracle := oracleOf(t, out.Report, s.tc.Manual)
	for _, want := range []string{
		needyDeckURL,
		"**Requested by:** Alice",
		"## Cards to add",
		"- [ ] " + s.tc.Manual + " (" + manualOracle + ")",
		"## Cards to review",
		"- [ ] " + s.tc.Unreviewed + " (",
		"## Automated with caveats",
		s.tc.Caveated + " — " + s.tc.CaveatText,
		"Filed from the deck checker on the site",
	} {
		if !strings.Contains(iss.body, want) {
			t.Errorf("issue body lacks %q:\n%s", want, iss.body)
		}
	}
	if strings.Contains(iss.body, "111111111111111111") {
		t.Error("the issue body publishes the requester's snowflake")
	}
	if strings.Contains(iss.body, "- [ ] "+s.tc.Automated) {
		t.Error("an automated card is on the checklist")
	}

	row, err := s.store.Lookup(context.Background(), "moxfield:needy01")
	if err != nil || row.IssueNumber != 1 {
		t.Errorf("row = %+v, %v", row, err)
	}
}

func oracleOf(t *testing.T, r *deckcoverage.Report, name string) string {
	t.Helper()
	for _, c := range r.Cards {
		if c.Name == name {
			return c.OracleID
		}
	}
	t.Fatalf("%s not in the report", name)
	return ""
}

func TestDeckRequestJoinsTheOpenIssueOnce(t *testing.T) {
	s := newDeckStack(t, true, nil)
	alice := s.discordUser(t, "111", "Alice")
	bob := s.discordUser(t, "222", "Bob")

	if code, _, raw := s.request(t, alice, map[string]string{"url": needyDeckURL}); code != http.StatusCreated {
		t.Fatalf("file: %d %s", code, raw)
	}
	code, out, raw := s.request(t, bob, map[string]string{"url": "https://www.moxfield.com/decks/needy01/other-slug"})
	if code != http.StatusOK || out.Status != deckRequestJoined || out.IssueNumber != 1 || out.AlreadyRequested {
		t.Fatalf("join: %d %s", code, raw)
	}
	issues, comments := s.filer.snapshot()
	if len(issues) != 1 {
		t.Errorf("%d issues filed, want 1", len(issues))
	}
	if len(comments) != 1 || comments[0].number != 1 || !strings.Contains(comments[0].body, "Also requested by Bob") ||
		!strings.Contains(comments[0].body, "not in the engine (manual): 1") {
		t.Errorf("comments = %+v", comments)
	}

	// Bob again, and Alice (who filed it): joined, no new comment.
	for _, tok := range []string{bob, alice} {
		code, out, raw = s.request(t, tok, map[string]string{"url": needyDeckURL})
		if code != http.StatusOK || out.Status != deckRequestJoined || !out.AlreadyRequested {
			t.Errorf("repeat: %d %s", code, raw)
		}
	}
	if _, comments = s.filer.snapshot(); len(comments) != 1 {
		t.Errorf("%d comments after repeats, want 1", len(comments))
	}
}

func TestDeckRequestRefilesAClosedIssue(t *testing.T) {
	s := newDeckStack(t, true, nil)
	alice := s.discordUser(t, "111", "Alice")
	if code, _, raw := s.request(t, alice, map[string]string{"url": needyDeckURL}); code != http.StatusCreated {
		t.Fatalf("file: %d %s", code, raw)
	}
	s.filer.close(1)

	// Alice asked on the closed issue; that does not make her a repeat
	// on the new one.
	code, out, raw := s.request(t, alice, map[string]string{"url": needyDeckURL})
	if code != http.StatusCreated || out.Status != deckRequestFiled || out.IssueNumber != 2 {
		t.Fatalf("refile: %d %s", code, raw)
	}
	row, err := s.store.Lookup(context.Background(), "moxfield:needy01")
	if err != nil || row.IssueNumber != 2 || row.IssueURL != "https://github.com/o/r/issues/2" {
		t.Errorf("row not repointed: %+v, %v", row, err)
	}

	// A deleted issue (404) counts as closed too.
	s.filer.mu.Lock()
	delete(s.filer.issues, 2)
	s.filer.mu.Unlock()
	bob := s.discordUser(t, "222", "Bob")
	if code, out, raw := s.request(t, bob, map[string]string{"url": needyDeckURL}); code != http.StatusCreated || out.IssueNumber != 3 {
		t.Fatalf("after a deleted issue: %d %s", code, raw)
	}
}

func TestDeckRequestIsRateLimitedPerRequester(t *testing.T) {
	s := newDeckStack(t, true, nil)
	tok := s.discordUser(t, "333", "Carol")
	now := time.Now().UTC()
	for i, key := range []string{"moxfield:a", "moxfield:b", "archidekt:1"} {
		if err := s.store.RecordAsk(context.Background(), key, "discord:333", now.Add(-time.Duration(3-i)*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	// An ask a day and a bit ago is outside the window.
	_ = s.store.RecordAsk(context.Background(), "moxfield:old", "discord:333", now.Add(-25*time.Hour))

	resp := postJSON(t, s.srv, "/deck-requests", tok, map[string]string{"url": needyDeckURL})
	defer resp.Body.Close()
	var out deckRequestResponse
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if resp.StatusCode != http.StatusTooManyRequests || out.Status != deckRequestRateLimited {
		t.Fatalf("status %d, %+v", resp.StatusCode, out)
	}
	// The oldest ask in the window was three hours ago, so the next
	// one is allowed in about 21 hours.
	if out.RetryAfter < 20*3600 || out.RetryAfter > 21*3600+60 {
		t.Errorf("retry_after = %d s", out.RetryAfter)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Error("no Retry-After header")
	}
	if n := s.source.count(needyDeckURL); n != 0 {
		t.Error("an over-limit request still fetched the deck")
	}
	if issues, _ := s.filer.snapshot(); len(issues) != 0 {
		t.Error("an over-limit request filed an issue")
	}

	// The limit is per person: the bot asking for someone else is fine.
	code, _, raw := s.request(t, s.adminToken(t), map[string]any{
		"url": needyDeckURL, "requester": map[string]string{"discord_id": "444", "display_name": "Dave"},
	})
	if code != http.StatusCreated {
		t.Errorf("another requester: %d %s", code, raw)
	}
}

func TestDeckRequestNothingToAdd(t *testing.T) {
	s := newDeckStack(t, true, nil)
	tok := s.discordUser(t, "111", "Alice")
	code, out, raw := s.request(t, tok, map[string]string{"url": cleanDeckURL})
	if code != http.StatusOK || out.Status != deckRequestNothingToAdd || out.Report == nil || out.IssueURL != "" {
		t.Fatalf("status %d: %s", code, raw)
	}
	if len(out.Report.CardsIn(deckcoverage.Caveats)) != 1 {
		t.Errorf("the caveated card should be listed: %s", raw)
	}
	if issues, _ := s.filer.snapshot(); len(issues) != 0 {
		t.Error("a deck with nothing to add filed an issue")
	}
	// Nothing to add is not an ask.
	if asks, _ := s.store.AsksSince(context.Background(), "discord:111", time.Time{}); len(asks) != 0 {
		t.Errorf("asks = %v", asks)
	}
}

func TestDeckRequestFromTheBot(t *testing.T) {
	s := newDeckStack(t, true, nil)
	admin := s.adminToken(t)

	code, out, raw := s.request(t, admin, map[string]any{
		"url": needyDeckURL, "requester": map[string]string{"discord_id": "555555555", "display_name": "  Eve\nthe @everyone  "},
	})
	if code != http.StatusCreated || out.Status != deckRequestFiled {
		t.Fatalf("status %d: %s", code, raw)
	}
	issues, _ := s.filer.snapshot()
	body := issues[1].body
	if !strings.Contains(body, "Filed by `/c2-deck-req` in Discord") {
		t.Errorf("footer:\n%s", body)
	}
	if !strings.Contains(body, "**Requested by:** Eve the @\u200beveryone") || strings.Contains(body, "555555555") {
		t.Errorf("requester line:\n%s", body)
	}
	asks, _ := s.store.AsksSince(context.Background(), "discord:555555555", time.Time{})
	if len(asks) != 1 {
		t.Errorf("the bot's ask was recorded %d times under the member's key", len(asks))
	}

	for _, bad := range []map[string]any{
		{"url": needyDeckURL},
		{"url": needyDeckURL, "requester": map[string]string{"discord_id": "not-a-snowflake", "display_name": "x"}},
		{"url": needyDeckURL, "requester": map[string]string{"discord_id": "1", "display_name": " "}},
	} {
		if code, _, raw := s.request(t, admin, bad); code != http.StatusBadRequest {
			t.Errorf("%v: %d %s", bad, code, raw)
		}
	}
}

func TestDeckRequestWhoMayAsk(t *testing.T) {
	s := newDeckStack(t, true, nil)
	body := map[string]string{"url": needyDeckURL}

	if code, _, raw := s.request(t, "", body); code != http.StatusUnauthorized {
		t.Errorf("anonymous: %d %s", code, raw)
	}
	guest := s.token(t, auth.Principal{Role: auth.RolePlayer, GameID: uuid.New(), PlayerID: uuid.New(), Name: "Guest"})
	if code, _, raw := s.request(t, guest, body); code != http.StatusForbidden {
		t.Errorf("guest: %d %s", code, raw)
	}
	// A signed-in user may not file on someone else's behalf.
	alice := s.discordUser(t, "111", "Alice")
	if code, _, raw := s.request(t, alice, map[string]any{
		"url": needyDeckURL, "requester": map[string]string{"discord_id": "999", "display_name": "Mallory"},
	}); code != http.StatusForbidden {
		t.Errorf("user naming a requester: %d %s", code, raw)
	}
	// A session that claims a user the database has never seen has no
	// Discord identity to file under.
	ghost := s.token(t, auth.Principal{Role: auth.RoleIdentified, UserID: uuid.New()})
	if code, _, raw := s.request(t, ghost, body); code != http.StatusForbidden {
		t.Errorf("user with no Discord identity: %d %s", code, raw)
	}
	if issues, _ := s.filer.snapshot(); len(issues) != 0 {
		t.Errorf("%d issues filed by refused callers", len(issues))
	}
}

func TestDeckRequestRefusesText(t *testing.T) {
	s := newDeckStack(t, true, nil)
	tok := s.discordUser(t, "111", "Alice")
	code, _, raw := s.request(t, tok, map[string]string{"text": "1 Forest"})
	if code != http.StatusBadRequest || !strings.Contains(raw, "link") {
		t.Errorf("status %d: %s", code, raw)
	}
	if code, _, raw := s.request(t, tok, map[string]string{"url": "https://example.com/decks/1"}); code != http.StatusBadRequest {
		t.Errorf("unsupported host: %d %s", code, raw)
	}
}

func TestDeckRequestOffWithoutAToken(t *testing.T) {
	s := newDeckStack(t, true, func(c *Config) { c.DeckRequestFiler = nil })
	tok := s.discordUser(t, "111", "Alice")
	code, _, raw := s.request(t, tok, map[string]string{"url": needyDeckURL})
	if code != http.StatusServiceUnavailable || !strings.Contains(raw, "CMDCTRL_GITHUB_TOKEN") {
		t.Errorf("status %d: %s", code, raw)
	}
	// The public checker does not need the token.
	if code, raw := s.post(t, "/deck-coverage", "", map[string]string{"url": needyDeckURL}); code != http.StatusOK {
		t.Errorf("coverage with requests off: %d %s", code, raw)
	}
}

func TestDeckRequestOffWithoutADatabase(t *testing.T) {
	s := newDeckStack(t, true, func(c *Config) { c.DeckRequests = nil })
	code, _, raw := s.request(t, s.adminToken(t), map[string]any{
		"url": needyDeckURL, "requester": map[string]string{"discord_id": "1", "display_name": "x"},
	})
	if code != http.StatusServiceUnavailable || !strings.Contains(raw, "CMDCTRL_DATA_DIR") {
		t.Errorf("status %d: %s", code, raw)
	}
}

func TestDeckRequestLabelRejectionStillFiles(t *testing.T) {
	s := newDeckStack(t, true, nil)
	s.filer.rejectLabels = true
	tok := s.discordUser(t, "111", "Alice")
	code, out, raw := s.request(t, tok, map[string]string{"url": needyDeckURL})
	if code != http.StatusCreated || out.Status != deckRequestFiled {
		t.Fatalf("status %d: %s", code, raw)
	}
	issues, _ := s.filer.snapshot()
	if len(issues) != 1 || len(issues[1].labels) != 0 {
		t.Errorf("issues = %+v", issues)
	}
}

// ADR 0017's redaction applies to anything this route publishes: the
// deck name and the requester's name came from outside.
func TestDeckRequestRedacts(t *testing.T) {
	s := newDeckStack(t, true, nil)
	leaky := "https://moxfield.com/decks/leaky"
	s.source.decks[leaky] = fakeDeck{name: "Spicy ?token=supersecretvalue", entries: s.source.decks[needyDeckURL].entries}
	code, _, raw := s.request(t, s.adminToken(t), map[string]any{
		"url": leaky, "requester": map[string]string{"discord_id": "1", "display_name": "me password=hunter2"},
	})
	if code != http.StatusCreated {
		t.Fatalf("status %d: %s", code, raw)
	}
	issues, _ := s.filer.snapshot()
	for _, leak := range []string{"supersecretvalue", "hunter2"} {
		if strings.Contains(issues[1].title, leak) || strings.Contains(issues[1].body, leak) {
			t.Errorf("issue publishes %q:\n%s\n%s", leak, issues[1].title, issues[1].body)
		}
	}
}

func TestDeckReportCacheExpires(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	c := newDeckReportCache(10*time.Minute, 2, func() time.Time { return now })
	r1, r2, r3 := &deckcoverage.Report{DeckName: "1"}, &deckcoverage.Report{DeckName: "2"}, &deckcoverage.Report{DeckName: "3"}

	c.put("a", r1)
	if got, ok := c.get("a"); !ok || got != r1 {
		t.Fatal("fresh entry missed")
	}
	now = now.Add(10 * time.Minute)
	if _, ok := c.get("a"); ok {
		t.Error("entry served at its TTL")
	}

	c.put("a", r1)
	now = now.Add(time.Minute)
	c.put("b", r2)
	c.put("c", r3) // full: evicts the entry closest to expiring, "a"
	if _, ok := c.get("a"); ok {
		t.Error("a full cache kept its oldest entry")
	}
	if _, ok := c.get("b"); !ok {
		t.Error("b evicted")
	}
	if _, ok := c.get("c"); !ok {
		t.Error("c not stored")
	}
}

// A deck name, a commander line and a display name come from outside
// and are published in a public repo: none of them may render as a
// link, an image or a table break in the issue body, and the title —
// which GitHub does not render — must not show escape backslashes.
func TestDeckRequestIssueEscapesMarkdown(t *testing.T) {
	r := &deckcoverage.Report{
		DeckName:   "Aang (Precon)](https://evil.example)![x](https://evil.example/i.png)",
		SourceURL:  "https://moxfield.com/decks/abc",
		DeckKey:    "moxfield:abc",
		Commanders: []string{"Evil | Cell"},
		Counts:     map[deckcoverage.Bucket]int{},
	}
	body := renderDeckRequestIssue(deckRequester{DiscordID: "1", DisplayName: "[me](https://evil.example)"}, true, r)
	for _, bad := range []string{"](https://evil.example)", "![x]", "Evil | Cell", "[me](https"} {
		if strings.Contains(body, bad) {
			t.Errorf("issue body renders %q as markdown:\n%s", bad, body)
		}
	}
	if !strings.Contains(body, "[Aang \\(Precon\\)") {
		t.Errorf("deck name not escaped as expected:\n%s", body)
	}
	if got := deckRequestTitle(r); strings.Contains(got, `\`) || !strings.HasPrefix(got, "[deck-request] Aang (Precon)") {
		t.Errorf("title = %q, want the plain deck name without escapes", got)
	}
}
