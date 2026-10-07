package lobby

// Tests for ADR 0128's playmat routes: who may set one, what an upload
// may be, and that every seat the person holds follows their playmat.

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/playmats"
)

// A 1×1 PNG.
var playmatPNG = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\rIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82")

type playmatReply struct {
	Error string `json:"error"`
	Path  string `json:"path"`
	Wash  int    `json:"wash"`
}

// newPlaymatStack is the signed-in stack with the playmat store, its
// files in a temp dir, and the lobby's lookup wired as main.go does.
func newPlaymatStack(t *testing.T) (userStack, string) {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	dir := filepath.Join(t.TempDir(), "playmats")
	files, err := playmats.NewFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	var store playmats.Store
	s := newUserStackWith(t, nil, func(c *Config) {
		store = playmats.NewSQLStore(c.Lobby.store.(*SQLStore).db)
		c.Playmats = store
		c.PlaymatFiles = files
	})
	s.lobby.SetPlaymatLookup(func(user uuid.UUID) (string, int) {
		p, err := store.Get(t.Context(), user)
		if err != nil {
			return "", 0
		}
		return p.Path(), p.Wash
	})
	return s, dir
}

func playmatRequest(t *testing.T, s userStack, method, tok, contentType string, body io.Reader) (int, playmatReply) {
	t.Helper()
	req, err := http.NewRequest(method, s.srv.URL+"/me/playmat", body)
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := s.srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out playmatReply
	if len(bytes.TrimSpace(raw)) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return resp.StatusCode, out
}

func uploadPlaymat(t *testing.T, s userStack, tok string, image []byte, wash int) (int, playmatReply) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if wash != 0 {
		_ = mw.WriteField("wash", strconv.Itoa(wash))
	}
	part, _ := mw.CreateFormFile("image", "mat.png")
	_, _ = part.Write(image)
	_ = mw.Close()
	return playmatRequest(t, s, http.MethodPut, tok, mw.FormDataContentType(), &buf)
}

func seatPlaymat(t *testing.T, s userStack, gameID, playerID uuid.UUID) (string, int) {
	t.Helper()
	s.lobby.mu.Lock()
	entry := s.lobby.games[gameID]
	s.lobby.mu.Unlock()
	var p *game.Player
	if entry != nil {
		p = entry.room.Game.PlayerByID(playerID)
	}
	if p == nil {
		t.Fatalf("no player %s in %s", playerID, gameID)
	}
	return p.PlaymatPath, p.PlaymatWash
}

func TestPlaymatNeedsASignedInPerson(t *testing.T) {
	s, _ := newPlaymatStack(t)
	if status, _ := playmatRequest(t, s, http.MethodGet, "", "", nil); status != http.StatusUnauthorized {
		t.Errorf("GET with no session = %d, want 401", status)
	}
	if status, _ := uploadPlaymat(t, s, "", playmatPNG, 0); status != http.StatusUnauthorized {
		t.Errorf("PUT with no session = %d, want 401", status)
	}
}

func TestPlaymatFollowsThePersonOntoEverySeat(t *testing.T) {
	s, dir := newPlaymatStack(t)
	tok, user := signedInSettingsUser(t, s)

	// Seated before they have a playmat.
	first, err := s.lobby.Create("first")
	if err != nil {
		t.Fatal(err)
	}
	_, firstPID, err := s.lobby.JoinAs(first.ID, first.InviteToken, "Alice", DiscordIdentity{}, user)
	if err != nil {
		t.Fatal(err)
	}

	if status, got := playmatRequest(t, s, http.MethodGet, tok, "", nil); status != http.StatusOK || got.Path != "" {
		t.Fatalf("GET before upload = %d %+v", status, got)
	}
	status, got := uploadPlaymat(t, s, tok, playmatPNG, 50)
	if status != http.StatusOK || !strings.HasPrefix(got.Path, "/playmats/") || got.Wash != 50 {
		t.Fatalf("PUT = %d %+v", status, got)
	}
	// The seat they already held follows at once.
	if path, wash := seatPlaymat(t, s, first.ID, firstPID); path != got.Path || wash != 50 {
		t.Errorf("held seat = %q %d, want %q 50", path, wash, got.Path)
	}
	// A seat taken afterwards carries it from the join.
	second, _ := s.lobby.Create("second")
	_, secondPID, err := s.lobby.JoinAs(second.ID, second.InviteToken, "Alice", DiscordIdentity{}, user)
	if err != nil {
		t.Fatal(err)
	}
	if path, _ := seatPlaymat(t, s, second.ID, secondPID); path != got.Path {
		t.Errorf("new seat = %q, want %q", path, got.Path)
	}
	// Anyone with a session can load the image.
	resp := doGet(t, s.srv, got.Path, tok)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || resp.Header.Get("Content-Type") != "image/png" || !bytes.Equal(body, playmatPNG) {
		t.Errorf("GET image = %d %q", resp.StatusCode, resp.Header.Get("Content-Type"))
	}

	// The wash changes on every seat too, and stays in range.
	if status, washed := playmatRequest(t, s, http.MethodPatch, tok, "application/json", strings.NewReader(`{"wash":80}`)); status != http.StatusOK || washed.Wash != 80 {
		t.Fatalf("PATCH = %d %+v", status, washed)
	}
	if _, wash := seatPlaymat(t, s, first.ID, firstPID); wash != 80 {
		t.Errorf("seat wash = %d, want 80", wash)
	}
	if status, _ := playmatRequest(t, s, http.MethodPatch, tok, "application/json", strings.NewReader(`{"wash":10}`)); status != http.StatusBadRequest {
		t.Errorf("PATCH wash 10 = %d, want 400", status)
	}

	// A new upload replaces the old file.
	_, again := uploadPlaymat(t, s, tok, playmatPNG, 0)
	if again.Path == got.Path || again.Wash != playmats.DefaultWash {
		t.Errorf("second upload = %+v", again)
	}
	if _, err := os.Stat(filepath.Join(dir, strings.TrimPrefix(got.Path, "/playmats/"))); !os.IsNotExist(err) {
		t.Errorf("replaced image still on disk: %v", err)
	}

	// Removing it clears every seat.
	if status, cleared := playmatRequest(t, s, http.MethodDelete, tok, "", nil); status != http.StatusOK || cleared.Path != "" {
		t.Fatalf("DELETE = %d %+v", status, cleared)
	}
	if path, _ := seatPlaymat(t, s, second.ID, secondPID); path != "" {
		t.Errorf("seat after DELETE = %q, want none", path)
	}
	if resp := doGet(t, s.srv, again.Path, tok); resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET removed image = %d, want 404", resp.StatusCode)
	}
}

func TestPlaymatRefusesWhatIsNotAnImage(t *testing.T) {
	s, dir := newPlaymatStack(t)
	tok, _ := signedInSettingsUser(t, s)
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`)
	if status, _ := uploadPlaymat(t, s, tok, svg, 0); status != http.StatusBadRequest {
		t.Errorf("PUT svg = %d, want 400", status)
	}
	if status, _ := uploadPlaymat(t, s, tok, playmatPNG, 95); status != http.StatusBadRequest {
		t.Errorf("PUT wash 95 = %d, want 400", status)
	}
	big := append(append([]byte{}, playmatPNG...), make([]byte, playmats.MaxImageBytes)...)
	if status, _ := uploadPlaymat(t, s, tok, big, 0); status != http.StatusRequestEntityTooLarge {
		t.Errorf("PUT over 4 MiB = %d, want 413", status)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("refused uploads left %d files", len(entries))
	}
	if status, _ := playmatRequest(t, s, http.MethodPatch, tok, "application/json", strings.NewReader(`{"wash":50}`)); status != http.StatusNotFound {
		t.Errorf("PATCH with no playmat = %d, want 404", status)
	}
}
