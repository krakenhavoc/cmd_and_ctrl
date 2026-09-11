package lobby

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/appenv"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// newDevStack builds an HTTP stack with an explicit environment and
// feature set, so the gate can be exercised from the outside.
func newDevStack(t *testing.T, env appenv.Env, features appenv.Features, idx *cards.Index) (*httptest.Server, *Lobby, auth.Authenticator) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()
	srv := httptest.NewServer(Handler(Config{
		Lobby:      l,
		Auth:       a,
		AdminToken: "shared-admin-token",
		Cards:      idx,
		Env:        env,
		Features:   features,
	}))
	t.Cleanup(srv.Close)
	return srv, l, a
}

func devIndex(t *testing.T) *cards.Index {
	t.Helper()
	idx := cards.NewIndex()
	idx.Put(cards.Card{
		ID:       uuid.MustParse("77c6fa74-5543-42ac-9ead-0e890b188e99"),
		OracleID: uuid.MustParse("4457ed35-7c10-48c8-9776-456485fdf070"),
		Name:     "Lightning Bolt",
		TypeLine: "Instant",
		ManaCost: "{R}",
		SetCode:  "lea",
	})
	return idx
}

func req(t *testing.T, srv *httptest.Server, method, path, token, body string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	request, err := http.NewRequest(method, srv.URL+path, r)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	resp, err := srv.Client().Do(request)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

// The load-bearing test for this feature. Spawning arbitrary cards is
// an outright cheat on a live table, so production must not merely
// hide the control -- the routes must not exist.
func TestDevRoutesAreAbsentInProduction(t *testing.T) {
	srv, _, a := newDevStack(t, appenv.EnvProd, appenv.LoadFeatures(appenv.EnvProd), devIndex(t))
	tok, _, err := a.Issue(context.Background(), auth.Principal{Role: auth.RoleAdmin}, time.Hour)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	gameID := uuid.New().String()

	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/dev/cards?q=lightning", ""},
		{http.MethodPost, "/games/" + gameID + "/dev/spawn", `{"name":"Lightning Bolt","player_id":"` + uuid.New().String() + `","zone":"battlefield"}`},
	} {
		// Even with a valid ADMIN credential.
		if got := req(t, srv, tc.method, tc.path, tok, tc.body).StatusCode; got != http.StatusNotFound {
			t.Errorf("%s %s with admin token: status %d, want 404", tc.method, tc.path, got)
		}
		// And unauthenticated: 404, never 401. A 401 would confirm
		// the route exists, which is the thing requireDev refuses to
		// do.
		if got := req(t, srv, tc.method, tc.path, "", tc.body).StatusCode; got != http.StatusNotFound {
			t.Errorf("%s %s unauthenticated: status %d, want 404 (401 leaks that the route exists)", tc.method, tc.path, got)
		}
	}
}

// A dev deployment that has switched card_spawn off is the same 404.
func TestDevRoutesAbsentWhenFeatureDisabled(t *testing.T) {
	srv, _, a := newDevStack(t, appenv.EnvDev, appenv.Features{SeatSwap: true}, devIndex(t))
	tok, _, err := a.Issue(context.Background(), auth.Principal{Role: auth.RoleAdmin}, time.Hour)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if got := req(t, srv, http.MethodGet, "/dev/cards?q=lightning", tok, "").StatusCode; got != http.StatusNotFound {
		t.Errorf("status %d, want 404 with card_spawn disabled", got)
	}
}

// On a dev deployment the route exists, so an unauthenticated call
// gets past the env gate and is refused by auth instead.
func TestDevRoutesPresentInDev(t *testing.T) {
	srv, _, a := newDevStack(t, appenv.EnvDev, appenv.LoadFeatures(appenv.EnvDev), devIndex(t))

	if got := req(t, srv, http.MethodGet, "/dev/cards?q=lightning", "", "").StatusCode; got != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want 401 (route should exist on dev)", got)
	}

	tok, _, err := a.Issue(context.Background(), auth.Principal{Role: auth.RoleAdmin}, time.Hour)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	resp := req(t, srv, http.MethodGet, "/dev/cards?q=lightning", tok, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("search status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		Cards []devCardResult `json:"cards"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Cards) != 1 || body.Cards[0].Name != "Lightning Bolt" {
		t.Fatalf("search results = %+v, want one Lightning Bolt", body.Cards)
	}
}

func TestParseSpawnZone(t *testing.T) {
	for _, ok := range []string{"battlefield", "HAND", " graveyard ", "exile", "library", "command"} {
		if _, err := parseSpawnZone(ok); err != nil {
			t.Errorf("parseSpawnZone(%q) = %v, want nil", ok, err)
		}
	}
	// The stack is rejected with its own message rather than the
	// generic list -- asking for it is coherent, just unsupported.
	_, err := parseSpawnZone("stack")
	if err == nil || !strings.Contains(err.Error(), "spawn to hand and cast it") {
		t.Errorf("parseSpawnZone(stack) = %v, want the cast-it-instead hint", err)
	}
	if _, err := parseSpawnZone("nowhere"); err == nil {
		t.Error("parseSpawnZone(nowhere) = nil, want an error")
	}
}

func TestResolveDevCard(t *testing.T) {
	idx := devIndex(t)

	if c, err := resolveDevCard(idx, devSpawnRequest{Name: "lightning bolt"}); err != nil || c.Name != "Lightning Bolt" {
		t.Errorf("by name: %+v %v", c, err)
	}
	if c, err := resolveDevCard(idx, devSpawnRequest{ScryfallID: "77c6fa74-5543-42ac-9ead-0e890b188e99"}); err != nil || c.Name != "Lightning Bolt" {
		t.Errorf("by id: %+v %v", c, err)
	}
	if _, err := resolveDevCard(idx, devSpawnRequest{}); err == nil {
		t.Error("neither id nor name should be an error")
	}
	if _, err := resolveDevCard(idx, devSpawnRequest{Name: "Not A Real Card"}); err == nil {
		t.Error("unknown name should be an error")
	}
}

// SpawnCards refuses a game that has not started: Start deals opening
// hands and would erase the spawn, which reads as the tool being
// broken rather than misused.
func TestSpawnCardsRequiresActiveGame(t *testing.T) {
	_, l, _ := newDevStack(t, appenv.EnvDev, appenv.LoadFeatures(appenv.EnvDev), devIndex(t))
	meta, err := l.Create("test")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = l.SpawnCards(meta.ID, uuid.New(), game.ZoneBattlefield, game.Card{Name: "X"}, 1)
	if err != ErrGameNotActiveForSpawn {
		t.Fatalf("err = %v, want ErrGameNotActiveForSpawn", err)
	}
}
