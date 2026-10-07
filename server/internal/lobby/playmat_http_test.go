package lobby

// Tests for ADR 0128's routes and seat binding: who may call them, the
// upload and link paths, the serving headers, and that the table sees
// a playmat on its owner's seat and on no one else's.

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/playmat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// playmatStack is a user stack with a playmat service over a temp dir,
// wired into both the routes and the lobby, as main.go does.
func playmatStack(t *testing.T) (userStack, string) {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	dir := filepath.Join(t.TempDir(), "playmats")
	var svc *playmat.Service
	s := newUserStackWithDB(t, nil, func(c *Config, d *db.DB) {
		svc = playmat.NewService(playmat.NewFileStore(dir), d.DB, nil)
		c.Playmats = svc
		c.Lobby.SetPlaymats(svc)
	})
	return s, dir
}

func pngOf(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = 30, 90, 160, 255
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// upload sends PUT /me/playmat with data as the "file" part.
func upload(t *testing.T, s userStack, tok string, data []byte) (int, playmatResponse, string) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="file"; filename="mat.png"`)
	h.Set("Content-Type", "image/png") // a claim the server must not rely on
	part, _ := mw.CreatePart(h)
	_, _ = part.Write(data)
	_ = mw.Close()
	return sendPlaymat(t, s, http.MethodPut, "/me/playmat", tok, mw.FormDataContentType(), &body)
}

func sendPlaymat(t *testing.T, s userStack, method, path, tok, ctype string, body io.Reader) (int, playmatResponse, string) {
	t.Helper()
	req, err := http.NewRequest(method, s.srv.URL+path, body)
	if err != nil {
		t.Fatal(err)
	}
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
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
	var out playmatResponse
	_ = json.Unmarshal(raw, &out)
	return resp.StatusCode, out, string(raw)
}

func linkPlaymat(t *testing.T, s userStack, tok, url string) (int, playmatResponse, string) {
	t.Helper()
	raw, _ := json.Marshal(map[string]string{"url": url})
	return sendPlaymat(t, s, http.MethodPost, "/me/playmat/link", tok, "application/json", bytes.NewReader(raw))
}

func getPlaymat(t *testing.T, s userStack, tok string) (int, playmatResponse) {
	t.Helper()
	status, out, _ := sendPlaymat(t, s, http.MethodGet, "/me/playmat", tok, "", nil)
	return status, out
}

func fetchImage(t *testing.T, s userStack, tok, url string) *http.Response {
	t.Helper()
	return doGet(t, s.srv, url, tok)
}

func TestMePlaymatRefusesEveryCallerWhoIsNotAPerson(t *testing.T) {
	s, dir := playmatStack(t)
	meta, err := s.lobby.Create("guests only")
	if err != nil {
		t.Fatal(err)
	}
	_, pid, err := s.lobby.Join(meta.ID, meta.InviteToken, "Guest")
	if err != nil {
		t.Fatal(err)
	}
	spectator, _, err := s.auth.Issue(context.Background(), auth.Principal{
		Role: auth.RoleSpectator, GameID: meta.ID, Name: "Watcher",
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	callers := []struct {
		name string
		tok  string
		want int
	}{
		{"no session", "", http.StatusUnauthorized},
		{"admin token", adminToken(t, s.srv), http.StatusForbidden},
		{"guest seat", playerToken(t, s.auth, meta.ID, pid, "Guest"), http.StatusForbidden},
		{"guest spectator", spectator, http.StatusForbidden},
	}
	for _, c := range callers {
		if status, _ := getPlaymat(t, s, c.tok); status != c.want {
			t.Errorf("GET as %s: %d, want %d", c.name, status, c.want)
		}
		if status, _, _ := upload(t, s, c.tok, pngOf(t, 8, 8)); status != c.want {
			t.Errorf("PUT as %s: %d, want %d", c.name, status, c.want)
		}
		if status, _, _ := linkPlaymat(t, s, c.tok, "https://example.com/a.png"); status != c.want {
			t.Errorf("POST link as %s: %d, want %d", c.name, status, c.want)
		}
		if status, _, _ := sendPlaymat(t, s, http.MethodDelete, "/me/playmat", c.tok, "", nil); status != c.want {
			t.Errorf("DELETE as %s: %d, want %d", c.name, status, c.want)
		}
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 0 {
		t.Errorf("a refused caller left %d files behind", len(ents))
	}
}

// With no database a Discord sign-in carries no user, so every route
// answers 403 and the client hides the control, as for /me/settings.
func TestMePlaymatOnADeploymentWithNoDatabase(t *testing.T) {
	srv, _, _, state := newDiscordTestStack(t)
	tok := identityTokenFromCallback(t, srv, state)
	resp := doGet(t, srv, "/me/playmat", tok)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("GET: %d, want 403", resp.StatusCode)
	}
	resp = postJSON(t, srv, "/me/playmat/link", tok, map[string]string{"url": "https://example.com/a.png"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("POST link: %d, want 403", resp.StatusCode)
	}
}

// A signed-in person on a server with no data directory: the feature
// reports itself disabled, and a write is a 503.
func TestMePlaymatDisabledWithoutADataDir(t *testing.T) {
	s := newMyGamesStack(t) // no Playmats on the Config
	tok, _ := signedInSettingsUser(t, s)
	status, got := getPlaymat(t, s, tok)
	if status != http.StatusOK || got.Enabled || got.URL != "" {
		t.Errorf("GET = %d %+v, want 200 {enabled:false}", status, got)
	}
	if status, _, _ := upload(t, s, tok, pngOf(t, 8, 8)); status != http.StatusServiceUnavailable {
		t.Errorf("PUT: %d, want 503", status)
	}
	if status, _, _ := linkPlaymat(t, s, tok, "https://example.com/a.png"); status != http.StatusServiceUnavailable {
		t.Errorf("POST link: %d, want 503", status)
	}
}

func TestMePlaymatUploadServeReplaceAndRemove(t *testing.T) {
	s, dir := playmatStack(t)
	tok, me := signedInSettingsUser(t, s)

	status, got := getPlaymat(t, s, tok)
	if status != http.StatusOK || !got.Enabled || got.URL != "" {
		t.Fatalf("GET with none = %d %+v", status, got)
	}

	status, first, raw := upload(t, s, tok, pngOf(t, 120, 80))
	if status != http.StatusOK || first.URL == "" || first.Width != 120 || first.Height != 80 {
		t.Fatalf("PUT = %d %s", status, raw)
	}
	if _, cur := getPlaymat(t, s, tok); cur.URL != first.URL {
		t.Errorf("GET after PUT = %+v, want %s", cur, first.URL)
	}

	// The image route: any session, the right headers, our JPEG.
	resp := fetchImage(t, s, tok, first.URL)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET image: %d", resp.StatusCode)
	}
	for h, want := range map[string]string{
		"Content-Type":            "image/jpeg",
		"X-Content-Type-Options":  "nosniff",
		"Content-Security-Policy": "default-src 'none'; sandbox",
	} {
		if got := resp.Header.Get(h); got != want {
			t.Errorf("%s = %q, want %q", h, got, want)
		}
	}
	if cc := resp.Header.Get("Cache-Control"); !strings.Contains(cc, "immutable") || !strings.Contains(cc, "private") {
		t.Errorf("Cache-Control = %q, want private + immutable", cc)
	}
	if _, f, err := image.DecodeConfig(bytes.NewReader(body)); err != nil || f != "jpeg" {
		t.Errorf("served bytes are %q (%v), want a jpeg", f, err)
	}
	// Anyone with a session may fetch it (every player sees every mat),
	// and nobody without one.
	meta, _ := s.lobby.Create("t")
	_, pid, _ := s.lobby.Join(meta.ID, meta.InviteToken, "Guest")
	resp = fetchImage(t, s, playerToken(t, s.auth, meta.ID, pid, "Guest"), first.URL)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("a guest seat fetching the image: %d, want 200", resp.StatusCode)
	}
	resp = fetchImage(t, s, "", first.URL)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no session fetching the image: %d, want 401", resp.StatusCode)
	}

	// A replacement deletes the old file and its URL stops resolving.
	status, second, _ := upload(t, s, tok, pngOf(t, 40, 40))
	if status != http.StatusOK || second.URL == first.URL {
		t.Fatalf("replacement = %d %+v", status, second)
	}
	resp = fetchImage(t, s, tok, first.URL)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("the replaced image: %d, want 404", resp.StatusCode)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 1 {
		t.Errorf("%d files after a replacement, want 1", len(ents))
	}

	// Remove.
	status, gone, _ := sendPlaymat(t, s, http.MethodDelete, "/me/playmat", tok, "", nil)
	if status != http.StatusOK || gone.URL != "" || !gone.Enabled {
		t.Errorf("DELETE = %d %+v", status, gone)
	}
	resp = fetchImage(t, s, tok, second.URL)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("the removed image: %d, want 404", resp.StatusCode)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 0 {
		t.Errorf("%d files after a removal, want 0", len(ents))
	}
	var id *string
	if err := s.db.QueryRow(`SELECT playmat_id FROM users WHERE id = ?`, me.String()).Scan(&id); err != nil || id != nil {
		t.Errorf("users.playmat_id = %v (%v), want NULL", id, err)
	}
	// Removing none is a success.
	if status, _, _ := sendPlaymat(t, s, http.MethodDelete, "/me/playmat", tok, "", nil); status != http.StatusOK {
		t.Errorf("a second DELETE: %d", status)
	}
}

func TestMePlaymatRefusesBadUploads(t *testing.T) {
	s, dir := playmatStack(t)
	tok, _ := signedInSettingsUser(t, s)
	for name, c := range map[string]struct {
		data []byte
		want int
	}{
		"text claiming to be a png": {[]byte("<html>not an image</html>"), http.StatusUnsupportedMediaType},
		"gif":                       {[]byte("GIF89a\x01\x00\x01\x00\x00\x00\x00;"), http.StatusUnsupportedMediaType},
		"over 10 MB":                {make([]byte, playmat.MaxUploadBytes+1), http.StatusRequestEntityTooLarge},
	} {
		if status, _, raw := upload(t, s, tok, c.data); status != c.want {
			t.Errorf("%s: %d (%s), want %d", name, status, raw, c.want)
		}
	}
	// Not multipart at all, and multipart with no "file" part.
	if status, _, _ := sendPlaymat(t, s, http.MethodPut, "/me/playmat", tok, "image/png", bytes.NewReader(pngOf(t, 8, 8))); status != http.StatusBadRequest {
		t.Errorf("a raw body: %d, want 400", status)
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("other", "x")
	_ = mw.Close()
	if status, _, _ := sendPlaymat(t, s, http.MethodPut, "/me/playmat", tok, mw.FormDataContentType(), &body); status != http.StatusBadRequest {
		t.Errorf("no file part: %d, want 400", status)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 0 {
		t.Errorf("a refused upload left %d files", len(ents))
	}
	if _, cur := getPlaymat(t, s, tok); cur.URL != "" {
		t.Error("a refused upload set a playmat")
	}
}

func TestMePlaymatLinkRefusesAddressesTheServerMustNotReach(t *testing.T) {
	s, dir := playmatStack(t)
	tok, _ := signedInSettingsUser(t, s)
	for _, url := range []string{
		"",
		"http://example.com/a.png",
		"https://127.0.0.1/a.png",
		"https://[::1]/a.png",
		"https://10.0.0.1/a.png",
		"https://169.254.169.254/latest/meta-data/",
		"https://user:pw@example.com/a.png",
		"file:///etc/passwd",
	} {
		status, _, raw := linkPlaymat(t, s, tok, url)
		if status != http.StatusUnprocessableEntity {
			t.Errorf("link %q: %d (%s), want 422", url, status, raw)
		}
		if strings.Contains(raw, "169.254") {
			t.Errorf("link %q: the error echoes the address: %s", url, raw)
		}
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 0 {
		t.Errorf("a refused link left %d files", len(ents))
	}
}

// The signed-in person and a guest sit down; the playmat set mid-game
// reaches the owner's seat on the next snapshot and no one else's, and
// every viewer sees the same URLs.
func TestPlaymatReachesTheTableOnTheOwnersSeatOnly(t *testing.T) {
	s, _ := playmatStack(t)
	tok, me := signedInSettingsUser(t, s)
	meta, err := s.lobby.Create("table")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.lobby.JoinAs(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{}, me); err != nil {
		t.Fatalf("JoinAs: %v", err)
	}
	if _, _, err := s.lobby.Join(meta.ID, meta.InviteToken, "Guest"); err != nil {
		t.Fatal(err)
	}
	room := s.lobby.RoomOf(meta.ID)
	seats := func() map[string]string {
		view, _, err := room.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]string{}
		for _, p := range view.Seats {
			out[p.Name] = p.PlaymatURL
		}
		return out
	}
	if got := seats(); got["Alice"] != "" || got["Guest"] != "" {
		t.Fatalf("before any playmat: %v", got)
	}

	// Setting one mid-game reaches the seat on the next snapshot, and
	// nobody else's.
	status, info, _ := upload(t, s, tok, pngOf(t, 50, 50))
	if status != http.StatusOK {
		t.Fatalf("PUT: %d", status)
	}
	got := seats()
	if got["Alice"] != info.URL || got["Guest"] != "" {
		t.Fatalf("after PUT: %v, want Alice on %s and the guest on none", got, info.URL)
	}

	// Every viewer sees the same URL.
	view, _, _ := room.Snapshot()
	var alice, other string
	for _, p := range view.Seats {
		if p.Name == "Alice" {
			alice = p.ID
		} else {
			other = p.ID
		}
	}
	for _, viewer := range []string{alice, other, protocol.SpectatorViewerID, ""} {
		for _, p := range protocol.FilterViewFor(view, viewer).Seats {
			want := ""
			if p.Name == "Alice" {
				want = info.URL
			}
			if p.PlaymatURL != want {
				t.Errorf("viewer %q: seat %s has %q, want %q", viewer, p.Name, p.PlaymatURL, want)
			}
		}
	}

	// Replacing and removing follow.
	_, second, _ := upload(t, s, tok, pngOf(t, 30, 30))
	if got := seats(); got["Alice"] != second.URL {
		t.Errorf("after a replacement: %v", got)
	}
	sendPlaymat(t, s, http.MethodDelete, "/me/playmat", tok, "", nil)
	if got := seats(); got["Alice"] != "" {
		t.Errorf("after a removal: %v", got)
	}
}

// A person who already has a playmat when they sit down brings it.
func TestPlaymatIsStampedWhenTheSeatIsBound(t *testing.T) {
	s, _ := playmatStack(t)
	tok, me := signedInSettingsUser(t, s)
	_, info, _ := upload(t, s, tok, pngOf(t, 50, 50))

	meta, _ := s.lobby.Create("table")
	_, pid, err := s.lobby.JoinAs(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{}, me)
	if err != nil {
		t.Fatal(err)
	}
	view, _, _ := s.lobby.RoomOf(meta.ID).Snapshot()
	for _, p := range view.Seats {
		if p.ID == pid.String() && p.PlaymatURL != info.URL {
			t.Errorf("seat bound with playmat %q, want %q", p.PlaymatURL, info.URL)
		}
	}
	// A guest never gets one, whatever else is true.
	_, gid, _ := s.lobby.Join(meta.ID, meta.InviteToken, "Guest")
	view, _, _ = s.lobby.RoomOf(meta.ID).Snapshot()
	for _, p := range view.Seats {
		if p.ID == gid.String() && p.PlaymatURL != "" {
			t.Errorf("a guest seat has playmat %q", p.PlaymatURL)
		}
	}
}

// Linking a guest seat to an account gives the seat that account's
// playmat, and a seat moved to another account takes the new one's.
func TestLinkingASeatBringsTheAccountsPlaymat(t *testing.T) {
	s, _ := playmatStack(t)
	tok, me := signedInSettingsUser(t, s)
	_, info, _ := upload(t, s, tok, pngOf(t, 50, 50))

	meta, _ := s.lobby.Create("table")
	_, pid, err := s.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	seatMat := func() string {
		view, _, _ := s.lobby.RoomOf(meta.ID).Snapshot()
		for _, p := range view.Seats {
			if p.ID == pid.String() {
				return p.PlaymatURL
			}
		}
		t.Fatal("seat not found")
		return ""
	}
	if got := seatMat(); got != "" {
		t.Fatalf("a guest seat has %q", got)
	}
	if _, _, err := s.lobby.LinkSeat(meta.ID, pid, DiscordIdentity{ID: "discord-99", Username: "alice"}, me); err != nil {
		t.Fatalf("LinkSeat: %v", err)
	}
	// LinkSeat commits through the room, so its own capture carries it.
	if got := seatMat(); got != info.URL {
		t.Errorf("after linking: %q, want %q", got, info.URL)
	}
}

// An admin removes anyone's playmat; nobody else can.
func TestAdminRemovesAPlaymat(t *testing.T) {
	s, dir := playmatStack(t)
	tok, me := signedInSettingsUser(t, s)
	_, info, _ := upload(t, s, tok, pngOf(t, 30, 30))

	path := "/admin/users/" + me.String() + "/playmat"
	for name, who := range map[string]string{"no session": "", "the person themself": tok} {
		status, _, _ := sendPlaymat(t, s, http.MethodDelete, path, who, "", nil)
		if status != http.StatusUnauthorized && status != http.StatusForbidden && status != http.StatusNotFound {
			t.Errorf("%s: %d, want a refusal", name, status)
		}
	}
	if _, cur := getPlaymat(t, s, tok); cur.URL != info.URL {
		t.Fatal("a non-admin removed the playmat")
	}

	admin := adminToken(t, s.srv)
	if status, _, _ := sendPlaymat(t, s, http.MethodDelete, path, admin, "", nil); status != http.StatusNoContent {
		t.Fatalf("admin DELETE: %d, want 204", status)
	}
	if _, cur := getPlaymat(t, s, tok); cur.URL != "" {
		t.Errorf("the playmat survived: %+v", cur)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 0 {
		t.Errorf("%d files left", len(ents))
	}
	if status, _, _ := sendPlaymat(t, s, http.MethodDelete, path, admin, "", nil); status != http.StatusNoContent {
		t.Errorf("removing none: %d, want 204", status)
	}
	if status, _, _ := sendPlaymat(t, s, http.MethodDelete, "/admin/users/"+uuid.NewString()+"/playmat", admin, "", nil); status != http.StatusNotFound {
		t.Errorf("an unknown user: %d, want 404", status)
	}
	if status, _, _ := sendPlaymat(t, s, http.MethodDelete, "/admin/users/not-a-uuid/playmat", admin, "", nil); status != http.StatusBadRequest {
		t.Errorf("a bad id: %d, want 400", status)
	}
}

// A restored table stamps its seats from the accounts again: the mat
// is the account's, not the engine snapshot's.
func TestPlaymatSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	d := openTestDB(t, dir)
	me := uuid.New()
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'Alice', 1, 1)`, me.String()); err != nil {
		t.Fatal(err)
	}
	svc := playmat.NewService(playmat.NewFileStore(filepath.Join(dir, "playmats")), d.DB, nil)
	info, err := svc.SetFromBytes(context.Background(), me, pngOf(t, 40, 40))
	if err != nil {
		t.Fatal(err)
	}

	l, _ := newDurableLobby(t, dir)
	l.SetPlaymats(svc)
	meta, err := l.Create("FNM")
	if err != nil {
		t.Fatal(err)
	}
	_, alice, err := l.JoinAs(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{}, me)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Guest"); err != nil {
		t.Fatal(err)
	}

	// The deploy. A new process has an empty cache and a fresh service.
	l2, _ := newDurableLobby(t, dir)
	l2.SetPlaymats(playmat.NewService(playmat.NewFileStore(filepath.Join(dir, "playmats")), d.DB, nil))
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	view, _, err := l2.RoomOf(meta.ID).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range view.Seats {
		want := ""
		if p.ID == alice.String() {
			want = info.URL
		}
		if p.PlaymatURL != want {
			t.Errorf("restored seat %s has playmat %q, want %q", p.Name, p.PlaymatURL, want)
		}
	}
}

// ADR 0128 amendment: PATCH /me/playmat sets the owner's wash, the
// table sees it on the owner's seat, and a seat taken later carries it.
func TestPlaymatWashReachesTheTable(t *testing.T) {
	s, _ := playmatStack(t)
	tok, me := signedInSettingsUser(t, s)
	patch := func(body string) (int, playmatResponse) {
		status, resp, _ := sendPlaymat(t, s, http.MethodPatch, "/me/playmat", tok, "application/json", strings.NewReader(body))
		return status, resp
	}
	// Before any image: the default, and a wash may still be chosen.
	if status, got := getPlaymat(t, s, tok); status != http.StatusOK || got.Wash != playmat.DefaultWash {
		t.Fatalf("GET before = %d %+v, want wash %d", status, got, playmat.DefaultWash)
	}
	if status, got := patch(`{"wash":40}`); status != http.StatusOK || got.Wash != 40 || got.URL != "" {
		t.Fatalf("PATCH with no image = %d %+v", status, got)
	}
	for _, bad := range []string{`{"wash":10}`, `{"wash":95}`, `{"wash":"x"}`, `{"darkness":50}`} {
		if status, _ := patch(bad); status != http.StatusBadRequest {
			t.Errorf("PATCH %s = %d, want 400", bad, status)
		}
	}

	meta, _ := s.lobby.Create("table")
	if _, _, err := s.lobby.JoinAs(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{}, me); err != nil {
		t.Fatal(err)
	}
	room := s.lobby.RoomOf(meta.ID)
	aliceWash := func() (string, int) {
		view, _, err := room.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range view.Seats {
			if p.Name == "Alice" {
				return p.PlaymatURL, p.PlaymatWash
			}
		}
		return "", -1
	}
	status, info, _ := upload(t, s, tok, pngOf(t, 40, 40))
	if status != http.StatusOK || info.Wash != 40 {
		t.Fatalf("PUT = %d %+v, want the chosen wash back", status, info)
	}
	if url, wash := aliceWash(); url != info.URL || wash != 40 {
		t.Fatalf("after upload the seat has %q %d, want %q 40", url, wash, info.URL)
	}
	if status, got := patch(`{"wash":85}`); status != http.StatusOK || got.Wash != 85 || got.URL != info.URL {
		t.Fatalf("PATCH = %d %+v", status, got)
	}
	if _, wash := aliceWash(); wash != 85 {
		t.Errorf("after PATCH the seat has wash %d, want 85", wash)
	}

	// A seat taken afterwards carries the wash from the join.
	meta2, _ := s.lobby.Create("second")
	if _, _, err := s.lobby.JoinAs(meta2.ID, meta2.InviteToken, "Alice", DiscordIdentity{}, me); err != nil {
		t.Fatal(err)
	}
	view, _, _ := s.lobby.RoomOf(meta2.ID).Snapshot()
	for _, p := range view.Seats {
		if p.Name == "Alice" && p.PlaymatWash != 85 {
			t.Errorf("a new seat has wash %d, want 85", p.PlaymatWash)
		}
	}
}
