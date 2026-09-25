package bot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
)

// fakeServer spins up an httptest.Server with /admin/login +
// /games routes stubbed out. Individual tests control the
// behaviour via the handler hooks.
type fakeServer struct {
	t *testing.T

	loginStatus int
	loginToken  string
	loginCalls  int

	createStatus int
	createMeta   lobby.GameMeta
	createName   string

	listStatus int
	listMeta   []lobby.GameMeta

	getStatus int
	getMeta   lobby.GameMeta

	archiveStatus int
	archiveMeta   lobby.GameMeta

	// creatorStatus/creatorIsCreator drive GET /games/{id}/creator
	// (#1098). gotCreatorDiscordIDs records the ?discord_id= query
	// value seen on each call, so tests can confirm the invoker's
	// snowflake — not someone else's — was asked about.
	creatorStatus        int
	creatorIsCreator     bool
	gotCreatorDiscordIDs []string

	// requireBearer, when non-empty, makes /games reject any other
	// Authorization value with 401 — lets the token-cache tests
	// simulate a server-side session expiry / rotation.
	requireBearer string

	gotAuthHeaders []string

	// deckCoverageStatus/deckCoverageBody drive POST /deck-coverage.
	// deckCoverageBody is marshalled verbatim, so it can be a
	// DeckCoverageReport (2xx) or an error map (non-2xx).
	deckCoverageStatus  int
	deckCoverageBody    any
	gotDeckCoverageURL  string
	gotDeckCoverageText string

	// deckRequestStatus/deckRequestBody drive POST /deck-requests,
	// the same way. gotDeckRequest records the decoded request body.
	deckRequestStatus int
	deckRequestBody   any
	gotDeckRequest    struct {
		URL       string        `json:"url"`
		Text      string        `json:"text"`
		Requester DeckRequester `json:"requester"`
	}
}

func newFakeServer(t *testing.T) (*fakeServer, *ServerClient) {
	t.Helper()
	fs := &fakeServer{
		t:                  t,
		loginStatus:        http.StatusOK,
		loginToken:         "session-token",
		createStatus:       http.StatusCreated,
		listStatus:         http.StatusOK,
		getStatus:          http.StatusOK,
		archiveStatus:      http.StatusOK,
		creatorStatus:      http.StatusOK,
		deckCoverageStatus: http.StatusOK,
		deckRequestStatus:  http.StatusCreated,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/admin/login", fs.handleLogin)
	mux.HandleFunc("/games", fs.handleGames)
	mux.HandleFunc("GET /games/{id}", fs.handleGetGame)
	mux.HandleFunc("POST /games/{id}/archive", fs.handleArchiveGame)
	mux.HandleFunc("GET /games/{id}/creator", fs.handleIsCreator)
	mux.HandleFunc("POST /deck-coverage", fs.handleDeckCoverage)
	mux.HandleFunc("POST /deck-requests", fs.handleDeckRequests)
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c := NewServerClient(ts.URL, "admin-secret").WithHTTPClient(&http.Client{Timeout: 2 * time.Second})
	return fs, c
}

func (fs *fakeServer) handleDeckCoverage(w http.ResponseWriter, r *http.Request) {
	fs.gotAuthHeaders = append(fs.gotAuthHeaders, r.Header.Get("Authorization"))
	if fs.requireBearer != "" && r.Header.Get("Authorization") != "Bearer "+fs.requireBearer {
		http.Error(w, "stale session", http.StatusUnauthorized)
		return
	}
	var body struct {
		URL  string `json:"url"`
		Text string `json:"text"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	fs.gotDeckCoverageURL = body.URL
	fs.gotDeckCoverageText = body.Text
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(fs.deckCoverageStatus)
	if fs.deckCoverageBody != nil {
		_ = json.NewEncoder(w).Encode(fs.deckCoverageBody)
	}
}

func (fs *fakeServer) handleDeckRequests(w http.ResponseWriter, r *http.Request) {
	fs.gotAuthHeaders = append(fs.gotAuthHeaders, r.Header.Get("Authorization"))
	if fs.requireBearer != "" && r.Header.Get("Authorization") != "Bearer "+fs.requireBearer {
		http.Error(w, "stale session", http.StatusUnauthorized)
		return
	}
	_ = json.NewDecoder(r.Body).Decode(&fs.gotDeckRequest)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(fs.deckRequestStatus)
	if fs.deckRequestBody != nil {
		_ = json.NewEncoder(w).Encode(fs.deckRequestBody)
	}
}

func (fs *fakeServer) handleIsCreator(w http.ResponseWriter, r *http.Request) {
	fs.gotAuthHeaders = append(fs.gotAuthHeaders, r.Header.Get("Authorization"))
	if fs.requireBearer != "" && r.Header.Get("Authorization") != "Bearer "+fs.requireBearer {
		http.Error(w, "stale session", http.StatusUnauthorized)
		return
	}
	fs.gotCreatorDiscordIDs = append(fs.gotCreatorDiscordIDs, r.URL.Query().Get("discord_id"))
	w.WriteHeader(fs.creatorStatus)
	if fs.creatorStatus == http.StatusOK {
		_ = json.NewEncoder(w).Encode(map[string]bool{"is_creator": fs.creatorIsCreator})
	}
}

func (fs *fakeServer) handleGetGame(w http.ResponseWriter, r *http.Request) {
	fs.gotAuthHeaders = append(fs.gotAuthHeaders, r.Header.Get("Authorization"))
	if fs.requireBearer != "" && r.Header.Get("Authorization") != "Bearer "+fs.requireBearer {
		http.Error(w, "stale session", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(fs.getStatus)
	if fs.getStatus == http.StatusOK {
		_ = json.NewEncoder(w).Encode(fs.getMeta)
	}
}

func (fs *fakeServer) handleArchiveGame(w http.ResponseWriter, r *http.Request) {
	fs.gotAuthHeaders = append(fs.gotAuthHeaders, r.Header.Get("Authorization"))
	if fs.requireBearer != "" && r.Header.Get("Authorization") != "Bearer "+fs.requireBearer {
		http.Error(w, "stale session", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(fs.archiveStatus)
	if fs.archiveStatus == http.StatusOK {
		_ = json.NewEncoder(w).Encode(fs.archiveMeta)
	}
}

func (fs *fakeServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if body.Token != "admin-secret" && fs.loginStatus == http.StatusOK {
		fs.t.Errorf("unexpected admin token: got %q", body.Token)
	}
	fs.loginCalls++
	w.WriteHeader(fs.loginStatus)
	if fs.loginStatus == http.StatusOK {
		_ = json.NewEncoder(w).Encode(map[string]string{"token": fs.loginToken})
	}
}

func (fs *fakeServer) handleGames(w http.ResponseWriter, r *http.Request) {
	fs.gotAuthHeaders = append(fs.gotAuthHeaders, r.Header.Get("Authorization"))
	if fs.requireBearer != "" && r.Header.Get("Authorization") != "Bearer "+fs.requireBearer {
		http.Error(w, "stale session", http.StatusUnauthorized)
		return
	}
	switch r.Method {
	case http.MethodPost:
		var body struct {
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		fs.createName = body.Name
		w.WriteHeader(fs.createStatus)
		if fs.createStatus == http.StatusCreated {
			_ = json.NewEncoder(w).Encode(fs.createMeta)
		}
	case http.MethodGet:
		w.WriteHeader(fs.listStatus)
		if fs.listStatus == http.StatusOK {
			_ = json.NewEncoder(w).Encode(listResponse{Games: fs.listMeta})
		}
	default:
		http.Error(w, "method", http.StatusMethodNotAllowed)
	}
}

func TestLogin_Success(t *testing.T) {
	_, c := newFakeServer(t)
	tok, err := c.Login(context.Background())
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if tok != "session-token" {
		t.Errorf("token: got %q", tok)
	}
}

func TestLogin_Unauthorized(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.loginStatus = http.StatusUnauthorized
	_, err := c.Login(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("want ErrUnauthorized, got %v", err)
	}
}

func TestLogin_NetworkError(t *testing.T) {
	c := NewServerClient("http://127.0.0.1:1", "admin").WithHTTPClient(&http.Client{Timeout: 200 * time.Millisecond})
	_, err := c.Login(context.Background())
	if !errors.Is(err, ErrServerUnreachable) {
		t.Errorf("want ErrServerUnreachable, got %v", err)
	}
}

func TestLogin_NoToken(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.loginToken = "" // server returns {"token": ""}
	_, err := c.Login(context.Background())
	if err == nil || !strings.Contains(err.Error(), "missing token") {
		t.Errorf("want missing-token error, got %v", err)
	}
}

func TestCreateGame_Success(t *testing.T) {
	fs, c := newFakeServer(t)
	id := uuid.New()
	fs.createMeta = lobby.GameMeta{
		ID:          id,
		Name:        "friday-commander",
		InviteToken: "invite-xyz",
		State:       "lobby",
	}
	meta, err := c.CreateGame(context.Background(), "friday-commander", "")
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if meta.ID != id {
		t.Errorf("id: got %v", meta.ID)
	}
	if meta.InviteToken != "invite-xyz" {
		t.Errorf("invite_token: got %q", meta.InviteToken)
	}
	if fs.createName != "friday-commander" {
		t.Errorf("server saw name %q", fs.createName)
	}
	if got := fs.gotAuthHeaders[0]; got != "Bearer session-token" {
		t.Errorf("auth header: got %q", got)
	}
}

func TestCreateGame_MissingInviteToken(t *testing.T) {
	fs, c := newFakeServer(t)
	// Token-less response — something's wrong on the server side.
	fs.createMeta = lobby.GameMeta{ID: uuid.New(), Name: "x", State: "lobby"}
	_, err := c.CreateGame(context.Background(), "x", "")
	if err == nil || !strings.Contains(err.Error(), "missing invite_token") {
		t.Errorf("want missing-invite-token error, got %v", err)
	}
}

func TestCreateGame_Unauthorized(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.createStatus = http.StatusUnauthorized
	_, err := c.CreateGame(context.Background(), "x", "")
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("want ErrUnauthorized, got %v", err)
	}
}

func TestCreateGame_ServerError(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.createStatus = http.StatusInternalServerError
	_, err := c.CreateGame(context.Background(), "x", "")
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Errorf("want 500-carrying error, got %v", err)
	}
}

func TestListGames_Success(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.listMeta = []lobby.GameMeta{
		{ID: uuid.New(), Name: "a", State: "lobby"},
		{ID: uuid.New(), Name: "b", State: "active"},
	}
	got, err := c.ListGames(context.Background())
	if err != nil {
		t.Fatalf("ListGames: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 games, got %d", len(got))
	}
	if got[0].Name != "a" || got[1].Name != "b" {
		t.Errorf("names: %q %q", got[0].Name, got[1].Name)
	}
}

func TestListGames_Unauthorized(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.listStatus = http.StatusUnauthorized
	_, err := c.ListGames(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("want ErrUnauthorized, got %v", err)
	}
}

func TestGetGame_Success(t *testing.T) {
	fs, c := newFakeServer(t)
	id := uuid.New()
	archivedAt := time.Now().UTC()
	fs.getMeta = lobby.GameMeta{ID: id, Name: "friday", State: "lobby", ArchivedAt: &archivedAt}

	meta, err := c.GetGame(context.Background(), id)
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if meta.ID != id || meta.Name != "friday" {
		t.Errorf("meta: got %+v", meta)
	}
	// GetGame must surface ArchivedAt — this is how /c2-end tells an
	// already-archived game apart from an active one; ListGames
	// would have filtered it out entirely.
	if !meta.Archived() {
		t.Error("expected Archived() == true; GetGame must not lose archived_at")
	}
}

func TestGetGame_NotFound(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.getStatus = http.StatusNotFound
	_, err := c.GetGame(context.Background(), uuid.New())
	if !errors.Is(err, ErrGameNotFound) {
		t.Errorf("want ErrGameNotFound, got %v", err)
	}
}

func TestGetGame_Unauthorized(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.getStatus = http.StatusUnauthorized
	_, err := c.GetGame(context.Background(), uuid.New())
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("want ErrUnauthorized, got %v", err)
	}
}

func TestGetGame_ServerError(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.getStatus = http.StatusInternalServerError
	_, err := c.GetGame(context.Background(), uuid.New())
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Errorf("want 500-carrying error, got %v", err)
	}
}

func TestArchiveGame_Success(t *testing.T) {
	fs, c := newFakeServer(t)
	id := uuid.New()
	archivedAt := time.Now().UTC()
	fs.archiveMeta = lobby.GameMeta{ID: id, Name: "friday", ArchivedAt: &archivedAt}

	meta, err := c.ArchiveGame(context.Background(), id)
	if err != nil {
		t.Fatalf("ArchiveGame: %v", err)
	}
	if !meta.Archived() {
		t.Error("expected the returned meta to carry archived_at")
	}
	if got := fs.gotAuthHeaders[len(fs.gotAuthHeaders)-1]; got != "Bearer session-token" {
		t.Errorf("auth header: got %q", got)
	}
}

func TestArchiveGame_NotFound(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.archiveStatus = http.StatusNotFound
	_, err := c.ArchiveGame(context.Background(), uuid.New())
	if !errors.Is(err, ErrGameNotFound) {
		t.Errorf("want ErrGameNotFound, got %v", err)
	}
}

func TestArchiveGame_Unauthorized(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.archiveStatus = http.StatusUnauthorized
	_, err := c.ArchiveGame(context.Background(), uuid.New())
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("want ErrUnauthorized, got %v", err)
	}
}

func TestArchiveGame_ServerError(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.archiveStatus = http.StatusInternalServerError
	_, err := c.ArchiveGame(context.Background(), uuid.New())
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Errorf("want 500-carrying error, got %v", err)
	}
}

func TestIsCreator_True(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.creatorIsCreator = true
	id := uuid.New()

	got, err := c.IsCreator(context.Background(), id, "discord-42")
	if err != nil {
		t.Fatalf("IsCreator: %v", err)
	}
	if !got {
		t.Error("want true")
	}
	if len(fs.gotCreatorDiscordIDs) != 1 || fs.gotCreatorDiscordIDs[0] != "discord-42" {
		t.Errorf("discord_id sent to server: got %v, want [discord-42]", fs.gotCreatorDiscordIDs)
	}
}

func TestIsCreator_False(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.creatorIsCreator = false

	got, err := c.IsCreator(context.Background(), uuid.New(), "discord-42")
	if err != nil {
		t.Fatalf("IsCreator: %v", err)
	}
	if got {
		t.Error("want false")
	}
}

func TestIsCreator_NotFound(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.creatorStatus = http.StatusNotFound
	_, err := c.IsCreator(context.Background(), uuid.New(), "discord-42")
	if !errors.Is(err, ErrGameNotFound) {
		t.Errorf("want ErrGameNotFound, got %v", err)
	}
}

func TestIsCreator_Unauthorized(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.creatorStatus = http.StatusUnauthorized
	_, err := c.IsCreator(context.Background(), uuid.New(), "discord-42")
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("want ErrUnauthorized, got %v", err)
	}
}

func TestNewServerClient_TrimsTrailingSlash(t *testing.T) {
	c := NewServerClient("http://example:9000/", "tok")
	if c.baseURL != "http://example:9000" {
		t.Errorf("baseURL: got %q", c.baseURL)
	}
}

// TestSessionTokenCachedAcrossCommands proves consecutive commands
// reuse one admin session instead of minting a fresh 12h session per
// slash command.
func TestSessionTokenCachedAcrossCommands(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.createMeta = lobby.GameMeta{ID: uuid.New(), InviteToken: "inv"}

	if _, err := c.ListGames(context.Background()); err != nil {
		t.Fatalf("ListGames #1: %v", err)
	}
	if _, err := c.ListGames(context.Background()); err != nil {
		t.Fatalf("ListGames #2: %v", err)
	}
	if _, err := c.CreateGame(context.Background(), "FNM", ""); err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if fs.loginCalls != 1 {
		t.Errorf("login calls: got %d, want 1 (token must be cached)", fs.loginCalls)
	}
}

// TestSessionTokenReloginOnce401 covers the refresh path: when the
// server rejects the cached token (expiry / restart), the client
// re-logins exactly once and retries with the fresh bearer.
func TestSessionTokenReloginOnce401(t *testing.T) {
	fs, c := newFakeServer(t)

	// Warm the cache with the original session.
	if _, err := c.ListGames(context.Background()); err != nil {
		t.Fatalf("warm ListGames: %v", err)
	}
	if fs.loginCalls != 1 {
		t.Fatalf("warm login calls: got %d, want 1", fs.loginCalls)
	}

	// Server-side rotation: only a NEW token is accepted from now on.
	fs.loginToken = "rotated-token"
	fs.requireBearer = "rotated-token"

	if _, err := c.ListGames(context.Background()); err != nil {
		t.Fatalf("ListGames after rotation: %v", err)
	}
	if fs.loginCalls != 2 {
		t.Errorf("login calls after 401 retry: got %d, want 2", fs.loginCalls)
	}
	n := len(fs.gotAuthHeaders)
	if n < 2 || fs.gotAuthHeaders[n-1] != "Bearer rotated-token" {
		t.Errorf("retry did not carry the fresh bearer; headers: %v", fs.gotAuthHeaders)
	}

	// And a wholly-dead admin token still surfaces ErrUnauthorized
	// (retry happens once, not forever).
	fs.loginToken = "another"
	fs.requireBearer = "never-matches"
	if _, err := c.ListGames(context.Background()); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("persistent 401: got %v, want ErrUnauthorized", err)
	}
}

// --- DeckCoverage / DeckRequest (ADR 0095 §4, #1631) ---

func TestDeckCoverage_Success(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckCoverageBody = DeckCoverageReport{
		DeckName:   "Needy Deck",
		Source:     "moxfield",
		SourceURL:  "https://moxfield.com/decks/AbC123",
		DeckKey:    "moxfield:AbC123",
		Commanders: []string{"Atraxa, Praetors' Voice"},
		Counts:     map[DeckCoverageBucket]int{BucketManual: 12, BucketUnreviewed: 1, BucketCaveats: 6, BucketAutomated: 40, BucketNoEffect: 29},
		Cards: []DeckCoverageCard{
			{Name: "Doubling Season", OracleID: "oid-1", Count: 1, Bucket: BucketManual},
		},
		Unknown:    []string{},
		Violations: []DeckViolation{},
	}

	report, err := c.DeckCoverage(context.Background(), DeckSource{URL: "https://moxfield.com/decks/AbC123"})
	if err != nil {
		t.Fatalf("DeckCoverage: %v", err)
	}
	if report.DeckName != "Needy Deck" || report.Counts[BucketManual] != 12 {
		t.Errorf("report: got %+v", report)
	}
	if fs.gotDeckCoverageURL != "https://moxfield.com/decks/AbC123" {
		t.Errorf("server saw url %q", fs.gotDeckCoverageURL)
	}
	if got := fs.gotAuthHeaders[0]; got != "Bearer session-token" {
		t.Errorf("deck-coverage must carry the admin session: got %q", got)
	}
}

func TestDeckCoverage_FetchError(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckCoverageStatus = http.StatusUnprocessableEntity
	fs.deckCoverageBody = map[string]any{
		"error":      "That deck is private. Make it public or unlisted on the deck site, or paste the list as text.",
		"code":       "deck_private",
		"violations": []map[string]string{{"code": "deck_private", "card": "https://moxfield.com/decks/x", "message": "private"}},
	}

	_, err := c.DeckCoverage(context.Background(), DeckSource{URL: "https://moxfield.com/decks/x"})
	var apiErr *DeckAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *DeckAPIError, got %v (%T)", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("status: got %d", apiErr.StatusCode)
	}
	if apiErr.Code != "deck_private" {
		t.Errorf("code: got %q", apiErr.Code)
	}
	if !strings.Contains(apiErr.Message, "private") {
		t.Errorf("message: got %q", apiErr.Message)
	}
	if len(apiErr.Violations) != 1 || apiErr.Violations[0].Code != "deck_private" {
		t.Errorf("violations: got %+v", apiErr.Violations)
	}
}

func TestDeckCoverage_Unauthorized(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckCoverageStatus = http.StatusUnauthorized
	_, err := c.DeckCoverage(context.Background(), DeckSource{URL: "https://moxfield.com/decks/x"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("want ErrUnauthorized, got %v", err)
	}
}

func TestDeckCoverage_ServiceUnavailable(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckCoverageStatus = http.StatusServiceUnavailable
	fs.deckCoverageBody = map[string]any{"error": "card index not loaded"}
	_, err := c.DeckCoverage(context.Background(), DeckSource{URL: "https://moxfield.com/decks/x"})
	var apiErr *DeckAPIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("want a 503 *DeckAPIError, got %v", err)
	}
}

func TestDeckRequest_Filed(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckRequestStatus = http.StatusCreated
	fs.deckRequestBody = DeckRequestResult{
		Status: DeckRequestFiled, IssueURL: "https://github.com/krakenhavoc/cmd_and_ctrl/issues/1700", IssueNumber: 1700,
		Report: &DeckCoverageReport{Counts: map[DeckCoverageBucket]int{BucketManual: 12}},
	}

	result, err := c.DeckRequest(context.Background(), DeckSource{URL: "https://moxfield.com/decks/AbC123"}, DeckRequester{DiscordID: "42", DisplayName: "Alice"})
	if err != nil {
		t.Fatalf("DeckRequest: %v", err)
	}
	if result.Status != DeckRequestFiled || result.IssueNumber != 1700 {
		t.Errorf("result: got %+v", result)
	}
	if fs.gotDeckRequest.URL != "https://moxfield.com/decks/AbC123" {
		t.Errorf("server saw url %q", fs.gotDeckRequest.URL)
	}
	if fs.gotDeckRequest.Requester.DiscordID != "42" || fs.gotDeckRequest.Requester.DisplayName != "Alice" {
		t.Errorf("server saw requester %+v", fs.gotDeckRequest.Requester)
	}
}

func TestDeckRequest_RateLimited(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckRequestStatus = http.StatusTooManyRequests
	fs.deckRequestBody = DeckRequestResult{Status: DeckRequestRateLimited, RetryAfter: 3600}

	result, err := c.DeckRequest(context.Background(), DeckSource{URL: "https://moxfield.com/decks/x"}, DeckRequester{})
	if err != nil {
		t.Fatalf("DeckRequest: %v (rate_limited is a result, not an error)", err)
	}
	if result.Status != DeckRequestRateLimited || result.RetryAfter != 3600 {
		t.Errorf("result: got %+v", result)
	}
}

func TestDeckRequest_IPLimiterShapeIsAnError(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckRequestStatus = http.StatusTooManyRequests
	fs.deckRequestBody = map[string]string{"error": "too many requests"}

	_, err := c.DeckRequest(context.Background(), DeckSource{URL: "https://moxfield.com/decks/x"}, DeckRequester{})
	var apiErr *DeckAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *DeckAPIError for a status-less 429, got %v", err)
	}
}

func TestDeckRequest_ServiceUnavailable(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckRequestStatus = http.StatusServiceUnavailable
	fs.deckRequestBody = map[string]string{"error": "deck requests need CMDCTRL_GITHUB_TOKEN"}

	_, err := c.DeckRequest(context.Background(), DeckSource{URL: "https://moxfield.com/decks/x"}, DeckRequester{})
	var apiErr *DeckAPIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("want a 503 *DeckAPIError, got %v", err)
	}
}

func TestDeckRequest_Unauthorized(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckRequestStatus = http.StatusUnauthorized
	_, err := c.DeckRequest(context.Background(), DeckSource{URL: "https://moxfield.com/decks/x"}, DeckRequester{})
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("want ErrUnauthorized, got %v", err)
	}
}

// A pasted list goes up as {text}, with no url (ADR 0095, amendment
// 2026-09-25), on both routes.
func TestDeckRoutes_SendAPastedListAsText(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckCoverageBody = DeckCoverageReport{Source: "text", DeckKey: "list:0123456789abcdef"}
	list := "1 Sol Ring\n1 Arcane Signet"
	report, err := c.DeckCoverage(context.Background(), DeckSource{Text: list})
	if err != nil || report.DeckKey != "list:0123456789abcdef" {
		t.Fatalf("DeckCoverage: %+v, %v", report, err)
	}
	if fs.gotDeckCoverageText != list || fs.gotDeckCoverageURL != "" {
		t.Errorf("server saw url %q text %q", fs.gotDeckCoverageURL, fs.gotDeckCoverageText)
	}

	fs.deckRequestBody = DeckRequestResult{Status: DeckRequestFiled, IssueNumber: 9}
	if _, err := c.DeckRequest(context.Background(), DeckSource{Text: list}, DeckRequester{DiscordID: "42", DisplayName: "Alice"}); err != nil {
		t.Fatalf("DeckRequest: %v", err)
	}
	if fs.gotDeckRequest.Text != list || fs.gotDeckRequest.URL != "" || fs.gotDeckRequest.Requester.DiscordID != "42" {
		t.Errorf("server saw %+v", fs.gotDeckRequest)
	}
}

func TestDeckCoverage_MoxfieldHint(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckCoverageStatus = http.StatusBadGateway
	fs.deckCoverageBody = map[string]any{
		"error": "Moxfield blocks our server. On Moxfield, open the deck → Export → Copy plain text, then paste the list here instead.",
		"code":  "upstream_blocked",
		"hint":  "paste_list",
	}
	_, err := c.DeckCoverage(context.Background(), DeckSource{URL: "https://moxfield.com/decks/x"})
	var apiErr *DeckAPIError
	if !errors.As(err, &apiErr) || apiErr.Hint != DeckHintPasteList || apiErr.Code != "upstream_blocked" {
		t.Fatalf("want a hinted *DeckAPIError, got %#v", err)
	}
}
