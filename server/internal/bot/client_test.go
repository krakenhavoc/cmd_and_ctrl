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

	createStatus int
	createMeta   lobby.GameMeta
	createName   string

	listStatus int
	listMeta   []lobby.GameMeta

	gotAuthHeaders []string
}

func newFakeServer(t *testing.T) (*fakeServer, *ServerClient) {
	t.Helper()
	fs := &fakeServer{
		t:            t,
		loginStatus:  http.StatusOK,
		loginToken:   "session-token",
		createStatus: http.StatusCreated,
		listStatus:   http.StatusOK,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/admin/login", fs.handleLogin)
	mux.HandleFunc("/games", fs.handleGames)
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c := NewServerClient(ts.URL, "admin-secret").WithHTTPClient(&http.Client{Timeout: 2 * time.Second})
	return fs, c
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
	w.WriteHeader(fs.loginStatus)
	if fs.loginStatus == http.StatusOK {
		_ = json.NewEncoder(w).Encode(map[string]string{"token": fs.loginToken})
	}
}

func (fs *fakeServer) handleGames(w http.ResponseWriter, r *http.Request) {
	fs.gotAuthHeaders = append(fs.gotAuthHeaders, r.Header.Get("Authorization"))
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
	meta, err := c.CreateGame(context.Background(), "friday-commander")
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
	_, err := c.CreateGame(context.Background(), "x")
	if err == nil || !strings.Contains(err.Error(), "missing invite_token") {
		t.Errorf("want missing-invite-token error, got %v", err)
	}
}

func TestCreateGame_Unauthorized(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.createStatus = http.StatusUnauthorized
	_, err := c.CreateGame(context.Background(), "x")
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("want ErrUnauthorized, got %v", err)
	}
}

func TestCreateGame_ServerError(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.createStatus = http.StatusInternalServerError
	_, err := c.CreateGame(context.Background(), "x")
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

func TestNewServerClient_TrimsTrailingSlash(t *testing.T) {
	c := NewServerClient("http://example:9000/", "tok")
	if c.baseURL != "http://example:9000" {
		t.Errorf("baseURL: got %q", c.baseURL)
	}
}
