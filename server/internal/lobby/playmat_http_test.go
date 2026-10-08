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
	"strconv"
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

// upload sends PUT /me/playmats/1 with data as the "file" part.
func upload(t *testing.T, s userStack, tok string, data []byte) (int, playmatResponse, string) {
	t.Helper()
	return uploadTo(t, s, tok, 1, data)
}

// uploadTo sends PUT /me/playmats/{slot} with data as the "file" part.
func uploadTo(t *testing.T, s userStack, tok string, slot int, data []byte) (int, playmatResponse, string) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="file"; filename="mat.png"`)
	h.Set("Content-Type", "image/png") // a claim the server must not rely on
	part, _ := mw.CreatePart(h)
	_, _ = part.Write(data)
	_ = mw.Close()
	return sendPlaymat(t, s, http.MethodPut, "/me/playmats/"+strconv.Itoa(slot), tok, mw.FormDataContentType(), &body)
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
	return linkPlaymatTo(t, s, tok, 1, url)
}

func linkPlaymatTo(t *testing.T, s userStack, tok string, slot int, url string) (int, playmatResponse, string) {
	t.Helper()
	raw, _ := json.Marshal(map[string]string{"url": url})
	return sendPlaymat(t, s, http.MethodPost, "/me/playmats/"+strconv.Itoa(slot)+"/link", tok, "application/json", bytes.NewReader(raw))
}

// The active slot's fields, for the tests that keep one mat.
func (r playmatResponse) active() *playmatSlotBody {
	if r.Active == nil {
		return nil
	}
	return r.slot(*r.Active)
}

func (r playmatResponse) slot(n int) *playmatSlotBody {
	for i := range r.Slots {
		if r.Slots[i].Slot == n {
			return &r.Slots[i]
		}
	}
	return nil
}

func (r playmatResponse) URL() string {
	if a := r.active(); a != nil {
		return a.URL
	}
	return ""
}

func (r playmatResponse) Width() int {
	if a := r.active(); a != nil {
		return a.Width
	}
	return 0
}

func (r playmatResponse) Height() int {
	if a := r.active(); a != nil {
		return a.Height
	}
	return 0
}

func getPlaymat(t *testing.T, s userStack, tok string) (int, playmatResponse) {
	t.Helper()
	status, out, _ := sendPlaymat(t, s, http.MethodGet, "/me/playmats", tok, "", nil)
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
		if status, _, _ := sendPlaymat(t, s, http.MethodDelete, "/me/playmats/1", c.tok, "", nil); status != c.want {
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
	resp := doGet(t, srv, "/me/playmats", tok)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("GET: %d, want 403", resp.StatusCode)
	}
	resp = postJSON(t, srv, "/me/playmats/1/link", tok, map[string]string{"url": "https://example.com/a.png"})
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
	if status != http.StatusOK || got.Enabled || got.URL() != "" {
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
	if status != http.StatusOK || !got.Enabled || got.URL() != "" {
		t.Fatalf("GET with none = %d %+v", status, got)
	}

	status, first, raw := upload(t, s, tok, pngOf(t, 120, 80))
	if status != http.StatusOK || first.URL() == "" || first.Width() != 120 || first.Height() != 80 {
		t.Fatalf("PUT = %d %s", status, raw)
	}
	if _, cur := getPlaymat(t, s, tok); cur.URL() != first.URL() {
		t.Errorf("GET after PUT = %+v, want %s", cur, first.URL())
	}

	// The image route: any session, the right headers, our JPEG.
	resp := fetchImage(t, s, tok, first.URL())
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
	resp = fetchImage(t, s, playerToken(t, s.auth, meta.ID, pid, "Guest"), first.URL())
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("a guest seat fetching the image: %d, want 200", resp.StatusCode)
	}
	resp = fetchImage(t, s, "", first.URL())
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no session fetching the image: %d, want 401", resp.StatusCode)
	}

	// A replacement deletes the old file and its URL stops resolving.
	status, second, _ := upload(t, s, tok, pngOf(t, 40, 40))
	if status != http.StatusOK || second.URL() == first.URL() {
		t.Fatalf("replacement = %d %+v", status, second)
	}
	resp = fetchImage(t, s, tok, first.URL())
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("the replaced image: %d, want 404", resp.StatusCode)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 1 {
		t.Errorf("%d files after a replacement, want 1", len(ents))
	}

	// Remove.
	status, gone, _ := sendPlaymat(t, s, http.MethodDelete, "/me/playmats/1", tok, "", nil)
	if status != http.StatusOK || gone.URL() != "" || !gone.Enabled {
		t.Errorf("DELETE = %d %+v", status, gone)
	}
	resp = fetchImage(t, s, tok, second.URL())
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
	if status, _, _ := sendPlaymat(t, s, http.MethodDelete, "/me/playmats/1", tok, "", nil); status != http.StatusOK {
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
	if status, _, _ := sendPlaymat(t, s, http.MethodPut, "/me/playmats/1", tok, "image/png", bytes.NewReader(pngOf(t, 8, 8))); status != http.StatusBadRequest {
		t.Errorf("a raw body: %d, want 400", status)
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("other", "x")
	_ = mw.Close()
	if status, _, _ := sendPlaymat(t, s, http.MethodPut, "/me/playmats/1", tok, mw.FormDataContentType(), &body); status != http.StatusBadRequest {
		t.Errorf("no file part: %d, want 400", status)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 0 {
		t.Errorf("a refused upload left %d files", len(ents))
	}
	if _, cur := getPlaymat(t, s, tok); cur.URL() != "" {
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
	if got["Alice"] != info.URL() || got["Guest"] != "" {
		t.Fatalf("after PUT: %v, want Alice on %s and the guest on none", got, info.URL())
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
				want = info.URL()
			}
			if p.PlaymatURL != want {
				t.Errorf("viewer %q: seat %s has %q, want %q", viewer, p.Name, p.PlaymatURL, want)
			}
		}
	}

	// Replacing and removing follow.
	_, second, _ := upload(t, s, tok, pngOf(t, 30, 30))
	if got := seats(); got["Alice"] != second.URL() {
		t.Errorf("after a replacement: %v", got)
	}
	sendPlaymat(t, s, http.MethodDelete, "/me/playmats/1", tok, "", nil)
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
		if p.ID == pid.String() && p.PlaymatURL != info.URL() {
			t.Errorf("seat bound with playmat %q, want %q", p.PlaymatURL, info.URL())
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
	if got := seatMat(); got != info.URL() {
		t.Errorf("after linking: %q, want %q", got, info.URL())
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
	if _, cur := getPlaymat(t, s, tok); cur.URL() != info.URL() {
		t.Fatal("a non-admin removed the playmat")
	}

	admin := adminToken(t, s.srv)
	if status, _, _ := sendPlaymat(t, s, http.MethodDelete, path, admin, "", nil); status != http.StatusNoContent {
		t.Fatalf("admin DELETE: %d, want 204", status)
	}
	if _, cur := getPlaymat(t, s, tok); cur.URL() != "" {
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
	info, err := svc.SetFromBytes(context.Background(), me, 1, pngOf(t, 40, 40))
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

// ADR 0128 amendment: PATCH /me/playmats sets the owner's wash, the
// table sees it on the owner's seat, and a seat taken later carries it.
func TestPlaymatWashReachesTheTable(t *testing.T) {
	s, _ := playmatStack(t)
	tok, me := signedInSettingsUser(t, s)
	patch := func(body string) (int, playmatResponse) {
		status, resp, _ := sendPlaymat(t, s, http.MethodPatch, "/me/playmats", tok, "application/json", strings.NewReader(body))
		return status, resp
	}
	// Before any image: the default, and a wash may still be chosen.
	if status, got := getPlaymat(t, s, tok); status != http.StatusOK || got.Wash != playmat.DefaultWash {
		t.Fatalf("GET before = %d %+v, want wash %d", status, got, playmat.DefaultWash)
	}
	if status, got := patch(`{"wash":40}`); status != http.StatusOK || got.Wash != 40 || got.URL() != "" {
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
	if url, wash := aliceWash(); url != info.URL() || wash != 40 {
		t.Fatalf("after upload the seat has %q %d, want %q 40", url, wash, info.URL())
	}
	if status, got := patch(`{"wash":85}`); status != http.StatusOK || got.Wash != 85 || got.URL() != info.URL() {
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

// ---- three slots, activation and the best-size fit (ADR 0128 §11) ----

func activatePlaymat(t *testing.T, s userStack, tok, body string) (int, playmatResponse, string) {
	t.Helper()
	return sendPlaymat(t, s, http.MethodPut, "/me/playmats/active", tok, "application/json", strings.NewReader(body))
}

func fitPlaymat(t *testing.T, s userStack, tok string, slot int, body string) (int, playmatResponse, string) {
	t.Helper()
	return sendPlaymat(t, s, http.MethodPost, "/me/playmats/"+strconv.Itoa(slot)+"/fit", tok, "application/json", strings.NewReader(body))
}

func deletePlaymat(t *testing.T, s userStack, tok string, slot int) (int, playmatResponse, string) {
	t.Helper()
	return sendPlaymat(t, s, http.MethodDelete, "/me/playmats/"+strconv.Itoa(slot), tok, "", nil)
}

func patchWash(t *testing.T, s userStack, tok, body string) int {
	t.Helper()
	st, _, _ := sendPlaymat(t, s, http.MethodPatch, "/me/playmats", tok, "application/json", strings.NewReader(body))
	return st
}

// seatPlaymat is the URL the table shows on the person's seat.
func seatPlaymat(t *testing.T, s userStack, gameID uuid.UUID, name string) string {
	t.Helper()
	view, _, err := s.lobby.RoomOf(gameID).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range view.Seats {
		if p.Name == name {
			return p.PlaymatURL
		}
	}
	t.Fatalf("no seat named %s", name)
	return ""
}

// seatedAlice signs a person in and seats them at a new table.
func seatedAlice(t *testing.T, s userStack) (tok string, gameID, user uuid.UUID) {
	t.Helper()
	tok, me := signedInSettingsUser(t, s)
	meta, err := s.lobby.Create("table")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.lobby.JoinAs(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{}, me); err != nil {
		t.Fatal(err)
	}
	return tok, meta.ID, me
}

func TestMePlaymatsRefuseEveryCallerWhoIsNotAPersonOnTheSlotRoutes(t *testing.T) {
	s, dir := playmatStack(t)
	meta, _ := s.lobby.Create("guests only")
	_, pid, _ := s.lobby.Join(meta.ID, meta.InviteToken, "Guest")
	callers := []struct {
		name string
		tok  string
		want int
	}{
		{"no session", "", http.StatusUnauthorized},
		{"admin token", adminToken(t, s.srv), http.StatusForbidden},
		{"guest seat", playerToken(t, s.auth, meta.ID, pid, "Guest"), http.StatusForbidden},
	}
	for _, c := range callers {
		status := func(st int, _ playmatResponse, _ string) int { return st }
		listStatus, _ := getPlaymat(t, s, c.tok)
		for name, got := range map[string]int{
			"PUT slot 2":   status(uploadTo(t, s, c.tok, 2, pngOf(t, 8, 8))),
			"POST link":    status(linkPlaymatTo(t, s, c.tok, 3, "https://example.com/a.png")),
			"POST fit":     status(fitPlaymat(t, s, c.tok, 1, `{"x":0,"y":0}`)),
			"DELETE slot":  status(deletePlaymat(t, s, c.tok, 2)),
			"PUT active":   status(activatePlaymat(t, s, c.tok, `{"slot":1}`)),
			"PATCH wash":   patchWash(t, s, c.tok, `{"wash":50}`),
			"GET the list": listStatus,
		} {
			if got != c.want {
				t.Errorf("%s as %s: %d, want %d", name, c.name, got, c.want)
			}
		}
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 0 {
		t.Errorf("a refused caller left %d files behind", len(ents))
	}
}

func TestMePlaymatsOnADeploymentWithNoDatabaseOrNoDataDir(t *testing.T) {
	srv, _, _, state := newDiscordTestStack(t)
	tok := identityTokenFromCallback(t, srv, state)
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/me/playmats/active", strings.NewReader(`{"slot":1}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("PUT active with no database: %d, want 403", resp.StatusCode)
	}
	resp = postJSON(t, srv, "/me/playmats/1/fit", tok, map[string]int{"x": 0, "y": 0})
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("POST fit with no database: %d, want 403", resp.StatusCode)
	}

	// A person on a server with no data directory: 503 on every write.
	s := newMyGamesStack(t)
	utok, _ := signedInSettingsUser(t, s)
	if st, _, _ := activatePlaymat(t, s, utok, `{"slot":null}`); st != http.StatusServiceUnavailable {
		t.Errorf("PUT active with no data dir: %d, want 503", st)
	}
	if st, _, _ := fitPlaymat(t, s, utok, 1, `{"x":0,"y":0}`); st != http.StatusServiceUnavailable {
		t.Errorf("POST fit with no data dir: %d, want 503", st)
	}
	if st, _, _ := deletePlaymat(t, s, utok, 1); st != http.StatusServiceUnavailable {
		t.Errorf("DELETE with no data dir: %d, want 503", st)
	}
}

func TestMePlaymatsSlotsOutsideOneToThreeAreA400(t *testing.T) {
	s, dir := playmatStack(t)
	tok, _ := signedInSettingsUser(t, s)
	for _, slot := range []string{"0", "4", "-1", "99", "abc", "1.5"} {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		part, _ := mw.CreateFormFile("file", "m.png")
		_, _ = part.Write(pngOf(t, 8, 8))
		_ = mw.Close()
		put, _, _ := sendPlaymat(t, s, http.MethodPut, "/me/playmats/"+slot, tok, mw.FormDataContentType(), &body)
		link, _, _ := sendPlaymat(t, s, http.MethodPost, "/me/playmats/"+slot+"/link", tok, "application/json", strings.NewReader(`{"url":"https://example.com/a.png"}`))
		fit, _, _ := sendPlaymat(t, s, http.MethodPost, "/me/playmats/"+slot+"/fit", tok, "application/json", strings.NewReader(`{"x":0,"y":0}`))
		del, _, _ := sendPlaymat(t, s, http.MethodDelete, "/me/playmats/"+slot, tok, "", nil)
		for what, got := range map[string]int{"PUT": put, "POST link": link, "POST fit": fit, "DELETE": del} {
			if got != http.StatusBadRequest {
				t.Errorf("%s slot %q: %d, want 400", what, slot, got)
			}
		}
	}
	for body, why := range map[string]string{`{"slot":4}`: "4", `{"slot":0}`: "0 (null is how to show none)", `{"slot":"one"}`: `"one"`} {
		if st, _, _ := activatePlaymat(t, s, tok, body); st != http.StatusBadRequest {
			t.Errorf("PUT active %s: %d, want 400", why, st)
		}
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 0 {
		t.Errorf("a refused slot left %d files", len(ents))
	}
}

func TestMePlaymatsListsThreeSlotsWithTheirFitAndSuggestion(t *testing.T) {
	s, _ := playmatStack(t)
	tok, _ := signedInSettingsUser(t, s)

	status, empty := getPlaymat(t, s, tok)
	if status != http.StatusOK || !empty.Enabled || len(empty.Slots) != 0 || empty.Active != nil ||
		empty.MaxSlots != 3 || empty.IdealWidth != playmat.IdealWidth || empty.IdealHeight != playmat.IdealHeight {
		t.Fatalf("empty list = %d %+v", status, empty)
	}

	// 1: wrong shape, small; 2: the ideal; 3: wrong shape, large.
	if st, _, raw := uploadTo(t, s, tok, 1, pngOf(t, 600, 800)); st != http.StatusOK {
		t.Fatalf("PUT 1: %d %s", st, raw)
	}
	if st, _, raw := uploadTo(t, s, tok, 2, pngOf(t, 2400, 1400)); st != http.StatusOK {
		t.Fatalf("PUT 2: %d %s", st, raw)
	}
	st, list, raw := uploadTo(t, s, tok, 3, pngOf(t, 2560, 1800))
	if st != http.StatusOK {
		t.Fatalf("PUT 3: %d %s", st, raw)
	}
	if list.Slot != 3 {
		t.Errorf("the response names slot %d as the one written, want 3", list.Slot)
	}
	if len(list.Slots) != 3 || list.Active == nil || *list.Active != 1 {
		t.Fatalf("list = %+v, want three slots with slot 1 active (the first mat saved)", list)
	}
	one, two, three := list.slot(1), list.slot(2), list.slot(3)
	if one.Fits || one.Suggestion == nil || !one.Suggestion.Smaller || one.Suggestion.Crop.Width != 600 || one.Suggestion.Crop.Height != 350 {
		t.Errorf("slot 1 = %+v %+v", one, one.Suggestion)
	}
	if !two.Fits || two.Suggestion != nil {
		t.Errorf("slot 2 = %+v: the ideal size gets no prompt", two)
	}
	if three.Fits || three.Suggestion == nil || three.Suggestion.Smaller ||
		three.Suggestion.TargetWidth != playmat.IdealWidth || three.Suggestion.TargetHeight != playmat.IdealHeight ||
		three.Suggestion.Crop.Width != 2560 || three.Suggestion.Crop.Height != 1493 {
		t.Errorf("slot 3 = %+v %+v", three, three.Suggestion)
	}
	if three.Width != 2560 || three.Height != 1800 {
		t.Errorf("slot 3 size = %dx%d", three.Width, three.Height)
	}
}

func TestMePlaymatsActivateRemoveAndTheTableFollow(t *testing.T) {
	s, dir := playmatStack(t)
	tok, gid, me := seatedAlice(t, s)
	_, a, _ := uploadTo(t, s, tok, 1, pngOf(t, 40, 40))
	_, b, _ := uploadTo(t, s, tok, 2, pngOf(t, 50, 50))
	one, two := a.slot(1).URL, b.slot(2).URL
	if got := seatPlaymat(t, s, gid, "Alice"); got != one {
		t.Fatalf("seat shows %q, want slot 1's %q (the first mat saved)", got, one)
	}

	// Activating another slot reaches the seat at once.
	status, resp, _ := activatePlaymat(t, s, tok, `{"slot":2}`)
	if status != http.StatusOK || resp.Active == nil || *resp.Active != 2 {
		t.Fatalf("activate 2 = %d %+v", status, resp)
	}
	if got := seatPlaymat(t, s, gid, "Alice"); got != two {
		t.Errorf("after activating 2 the seat shows %q, want %q", got, two)
	}
	// Activating an empty slot is a 404 and changes nothing.
	if st, _, _ := activatePlaymat(t, s, tok, `{"slot":3}`); st != http.StatusNotFound {
		t.Errorf("activate empty slot: %d, want 404", st)
	}
	if got := seatPlaymat(t, s, gid, "Alice"); got != two {
		t.Errorf("a refused activation changed the seat: %q", got)
	}
	// Show none: the mats stay saved, and so do their files.
	status, resp, _ = activatePlaymat(t, s, tok, `{"slot":null}`)
	if status != http.StatusOK || resp.Active != nil || len(resp.Slots) != 2 {
		t.Fatalf("activate null = %d %+v", status, resp)
	}
	if got := seatPlaymat(t, s, gid, "Alice"); got != "" {
		t.Errorf("showing none, the seat shows %q", got)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 2 {
		t.Errorf("%d files after switching, want both kept", len(ents))
	}
	// Back on, then remove the ACTIVE slot: the seat clears, the other
	// mat survives, and none is active.
	_, _, _ = activatePlaymat(t, s, tok, `{"slot":1}`)
	if got := seatPlaymat(t, s, gid, "Alice"); got != one {
		t.Fatalf("activate 1 did not reach the seat: %q", got)
	}
	status, resp, _ = deletePlaymat(t, s, tok, 1)
	if status != http.StatusOK || resp.Active != nil || len(resp.Slots) != 1 || resp.Slots[0].Slot != 2 {
		t.Fatalf("DELETE active slot = %d %+v", status, resp)
	}
	if got := seatPlaymat(t, s, gid, "Alice"); got != "" {
		t.Errorf("after removing the active mat the seat shows %q", got)
	}
	var active *string
	if err := s.db.QueryRow(`SELECT playmat_id FROM users WHERE id = ?`, me.String()).Scan(&active); err != nil || active != nil {
		t.Errorf("users.playmat_id = %v (%v), want NULL", active, err)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 1 {
		t.Errorf("%d files after removing a slot, want 1", len(ents))
	}
	// Removing an INACTIVE slot leaves the table alone.
	_, _, _ = activatePlaymat(t, s, tok, `{"slot":2}`)
	_, _, _ = uploadTo(t, s, tok, 3, pngOf(t, 30, 30))
	if st, _, _ := deletePlaymat(t, s, tok, 3); st != http.StatusOK {
		t.Errorf("DELETE inactive slot: %d", st)
	}
	if got := seatPlaymat(t, s, gid, "Alice"); got != two {
		t.Errorf("removing an inactive slot changed the seat: %q", got)
	}
}

func TestMePlaymatsReplacingTheActiveSlotReachesTheTableAndAnInactiveOneDoesNot(t *testing.T) {
	s, _ := playmatStack(t)
	tok, gid, _ := seatedAlice(t, s)
	_, a, _ := uploadTo(t, s, tok, 1, pngOf(t, 40, 40))
	first := a.slot(1).URL
	_, _, _ = uploadTo(t, s, tok, 2, pngOf(t, 41, 41))
	if got := seatPlaymat(t, s, gid, "Alice"); got != first {
		t.Fatalf("saving into slot 2 changed the table: %q", got)
	}
	_, b, _ := uploadTo(t, s, tok, 1, pngOf(t, 42, 42))
	if got := seatPlaymat(t, s, gid, "Alice"); got != b.slot(1).URL || got == first {
		t.Errorf("replacing the active slot: seat shows %q, want %q", got, b.slot(1).URL)
	}
}

func TestMePlaymatsFitRoute(t *testing.T) {
	s, dir := playmatStack(t)
	tok, gid, _ := seatedAlice(t, s)
	_, orig, _ := uploadTo(t, s, tok, 1, pngOf(t, 2560, 1800))
	before := orig.slot(1)
	if before.Suggestion == nil {
		t.Fatal("a 2560x1800 mat has no suggestion")
	}
	_, _, _ = uploadTo(t, s, tok, 2, pngOf(t, 2400, 1400))

	// A crop outside the image, a malformed body and an empty slot.
	for _, body := range []string{`{"x":0,"y":308}`, `{"x":-1,"y":0}`, `{"x":1,"y":0}`, `{"x":"a","y":0}`, `{"x":0,"y":0,"z":1}`, `not json`} {
		if st, _, raw := fitPlaymat(t, s, tok, 1, body); st != http.StatusBadRequest {
			t.Errorf("fit %s: %d (%s), want 400", body, st, raw)
		}
	}
	if st, _, _ := fitPlaymat(t, s, tok, 3, `{"x":0,"y":0}`); st != http.StatusNotFound {
		t.Errorf("fit an empty slot: %d, want 404", st)
	}
	if st, _, _ := fitPlaymat(t, s, tok, 2, `{"x":0,"y":0}`); st != http.StatusConflict {
		t.Errorf("fit a mat that already fits: %d, want 409", st)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 2 {
		t.Fatalf("%d files after refused fits, want 2", len(ents))
	}
	if got := seatPlaymat(t, s, gid, "Alice"); got != before.URL {
		t.Fatalf("a refused fit changed the table: %q", got)
	}

	// The fit itself: slot 1 is the active slot, so the table re-stamps.
	status, fitted, raw := fitPlaymat(t, s, tok, 1, `{"x":0,"y":307}`)
	if status != http.StatusOK {
		t.Fatalf("fit = %d %s", status, raw)
	}
	after := fitted.slot(1)
	if fitted.Slot != 1 || after.Width != playmat.IdealWidth || after.Height != playmat.IdealHeight || !after.Fits || after.Suggestion != nil {
		t.Errorf("fitted slot = %+v", after)
	}
	if after.URL == before.URL {
		t.Error("the fitted mat kept its URL, so a cache would serve the old image")
	}
	if got := seatPlaymat(t, s, gid, "Alice"); got != after.URL {
		t.Errorf("the table shows %q after a fit, want %q", got, after.URL)
	}
	resp := fetchImage(t, s, tok, before.URL)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("the pre-fit file: %d, want 404 (deleted)", resp.StatusCode)
	}
	resp = fetchImage(t, s, tok, after.URL)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(body)); err != nil || cfg.Width != playmat.IdealWidth || cfg.Height != playmat.IdealHeight {
		t.Errorf("the served file is %v (%v), want %dx%d", cfg, err, playmat.IdealWidth, playmat.IdealHeight)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 2 {
		t.Errorf("%d files after a fit, want 2 (the old one deleted, the new one written)", len(ents))
	}
	// Fitting twice is a 409: the result is already the ideal shape.
	if st, _, _ := fitPlaymat(t, s, tok, 1, `{"x":0,"y":0}`); st != http.StatusConflict {
		t.Errorf("a second fit: %d, want 409", st)
	}
}

func TestMePlaymatsFittingAnInactiveSlotLeavesTheTableAlone(t *testing.T) {
	s, _ := playmatStack(t)
	tok, gid, _ := seatedAlice(t, s)
	_, a, _ := uploadTo(t, s, tok, 1, pngOf(t, 2400, 1400))
	_, _, _ = uploadTo(t, s, tok, 2, pngOf(t, 2560, 1800))
	if st, _, raw := fitPlaymat(t, s, tok, 2, `{"x":0,"y":0}`); st != http.StatusOK {
		t.Fatalf("fit = %d %s", st, raw)
	}
	if got := seatPlaymat(t, s, gid, "Alice"); got != a.slot(1).URL {
		t.Errorf("fitting slot 2 changed the table: %q", got)
	}
}

// The wash is one per account, shared by every slot, and is not lost
// when the active mat changes.
func TestMePlaymatsWashIsPerAccountNotPerSlot(t *testing.T) {
	s, _ := playmatStack(t)
	tok, gid, _ := seatedAlice(t, s)
	_, _, _ = uploadTo(t, s, tok, 1, pngOf(t, 40, 40))
	_, _, _ = uploadTo(t, s, tok, 2, pngOf(t, 41, 41))
	if st := patchWash(t, s, tok, `{"wash":77}`); st != http.StatusOK {
		t.Fatalf("PATCH: %d", st)
	}
	_, _, _ = activatePlaymat(t, s, tok, `{"slot":2}`)
	view, _, _ := s.lobby.RoomOf(gid).Snapshot()
	for _, p := range view.Seats {
		if p.Name == "Alice" && p.PlaymatWash != 77 {
			t.Errorf("after switching mats the seat's wash is %d, want 77", p.PlaymatWash)
		}
	}
	if _, got := getPlaymat(t, s, tok); got.Wash != 77 {
		t.Errorf("GET wash = %d, want 77", got.Wash)
	}
}

// The v1 singular routes are gone: the client moved with the server.
func TestTheV1SingularPlaymatRoutesAreGone(t *testing.T) {
	s, _ := playmatStack(t)
	tok, _ := signedInSettingsUser(t, s)
	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/me/playmat"},
		{http.MethodPut, "/me/playmat"},
		{http.MethodDelete, "/me/playmat"},
		{http.MethodPatch, "/me/playmat"},
		{http.MethodPost, "/me/playmat/link"},
	} {
		st, _, _ := sendPlaymat(t, s, c.method, c.path, tok, "application/json", strings.NewReader(`{}`))
		if st != http.StatusNotFound && st != http.StatusMethodNotAllowed {
			t.Errorf("%s %s: %d, want 404 or 405", c.method, c.path, st)
		}
	}
}
