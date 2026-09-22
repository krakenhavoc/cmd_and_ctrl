package lobby

// decklibrary_http_test.go covers ADR 0051 decision 7 (S34 sub-PR 5):
// POST /games/{id}/decks saving or updating a library row for a
// signed-in caller, POST /games/{id}/decks/{deck_id} seating one back
// without re-pasting, and GET /me/decks. The upload's own pre-existing
// behaviour (guests, validation, rate limits, ...) is covered by
// http_test.go; this file is additive.

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// deckLibraryStack is a full HTTP stack with a real, migrated
// decklibrary.SQLStore wired in (an in-memory Users store would do,
// but decks.owner_id is a real foreign key, so a users row has to
// exist for every owner a test uses — see mustLibraryUser).
type deckLibraryStack struct {
	srv     *httptest.Server
	lobby   *Lobby
	auth    *auth.MemoryAuthenticator
	db      *db.DB
	library decklibrary.Store
}

func newDeckLibraryStack(t *testing.T, idx *cards.Index) *deckLibraryStack {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()
	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&WSAuthorizer{Auth: a})

	dbase, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = dbase.Close() })
	library := decklibrary.NewSQLStore(dbase)

	cfg := Config{
		Lobby:       l,
		Auth:        a,
		AdminToken:  "shared-admin-token",
		Cards:       idx,
		DeckLibrary: library,
		Log:         log,
	}
	mux := http.NewServeMux()
	mux.Handle("/", Handler(cfg))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &deckLibraryStack{srv: srv, lobby: l, auth: a, db: dbase, library: library}
}

// mustLibraryUser inserts a users row directly — decks.owner_id is a
// real foreign key from migration 0005 — and returns its id.
func mustLibraryUser(t *testing.T, dbase *db.DB, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := dbase.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, ?, ?)`,
		id.String(), name, time.Now().UnixMilli(), time.Now().UnixMilli()); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

// signedInToken mints a RolePlayer session that also carries a
// UserID, the way S34's Discord-linked join flow would, without
// standing up the OAuth round trip. Mirrors the pattern
// server/internal/auth/hmac_test.go already uses for the same reason.
func signedInToken(t *testing.T, a auth.Authenticator, gameID, playerID, userID uuid.UUID) string {
	t.Helper()
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role:     auth.RolePlayer,
		GameID:   gameID,
		PlayerID: playerID,
		UserID:   userID,
	}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	return tok
}

func TestUploadDeckSavesToLibraryForSignedInUser(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	st := newDeckLibraryStack(t, idx)
	owner := mustLibraryUser(t, st.db, "Alice")

	meta, _ := st.lobby.Create("FNM")
	_, playerID, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	tok := signedInToken(t, st.auth, meta.ID, playerID, owner)

	source := "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"
	resp := postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks", tok,
		uploadDeckRequest{Format: "text", Source: source, PlayerID: playerID})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
	resp.Body.Close()

	decks, err := st.library.List(context.Background(), owner)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(decks) != 1 {
		t.Fatalf("library decks = %d, want 1", len(decks))
	}
	d := decks[0]
	// A plain-text paste has no name of its own; the fallback is the
	// first commander's name (saveToLibrary / libraryFallbackName).
	if d.Name != "Test Commander" {
		t.Errorf("name = %q, want the commander's name as a fallback", d.Name)
	}
	if d.SourceFormat != "text" || d.SourceText != source {
		t.Errorf("stored source = %q/%q, want text/%q", d.SourceFormat, d.SourceText, source)
	}
	if d.CardCount != 100 || len(d.Commanders) != 1 || d.Commanders[0] != "Test Commander" {
		t.Errorf("deck = %+v", d)
	}
}

func TestUploadDeckReuploadWithSameNameUpdatesInPlace(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	st := newDeckLibraryStack(t, idx)
	owner := mustLibraryUser(t, st.db, "Alice")

	meta, _ := st.lobby.Create("FNM")
	_, playerID, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	tok := signedInToken(t, st.auth, meta.ID, playerID, owner)

	// Moxfield JSON carries an explicit name, so both uploads share one
	// without depending on the fallback-name path.
	first := `{"name":"Atraxa Superfriends","boards":{"commanders":{"cards":{"c1":{"quantity":1,"card":{"name":"Test Commander"}}}},"mainboard":{"cards":{"m1":{"quantity":99,"card":{"name":"Plains"}}}}}}`
	resp := postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks", tok,
		uploadDeckRequest{Format: "moxfield", Source: first, PlayerID: playerID})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("first upload: got %d (body=%s)", resp.StatusCode, body)
	}
	resp.Body.Close()

	decksAfterFirst, err := st.library.List(context.Background(), owner)
	if err != nil || len(decksAfterFirst) != 1 {
		t.Fatalf("after first upload: %v decks, err=%v, want 1", decksAfterFirst, err)
	}
	firstID := decksAfterFirst[0].ID

	// Re-upload the same-named deck. Card-for-card identical (99 Plains
	// stays legal) but a different Moxfield-internal card id, so the
	// bytes differ and updating in place is visible in source_text
	// without the deck itself changing.
	second := `{"name":"Atraxa Superfriends","boards":{"commanders":{"cards":{"c1":{"quantity":1,"card":{"name":"Test Commander"}}}},"mainboard":{"cards":{"m2":{"quantity":99,"card":{"name":"Plains"}}}}}}`
	resp = postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks", tok,
		uploadDeckRequest{Format: "moxfield", Source: second, PlayerID: playerID})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("second upload: got %d (body=%s)", resp.StatusCode, body)
	}
	resp.Body.Close()

	decksAfterSecond, err := st.library.List(context.Background(), owner)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(decksAfterSecond) != 1 {
		t.Fatalf("after second upload: %d decks, want 1 (update rule inserted a duplicate)", len(decksAfterSecond))
	}
	if decksAfterSecond[0].ID != firstID {
		t.Errorf("update rule: same owner+name got a new id (%s vs %s)", decksAfterSecond[0].ID, firstID)
	}
	if decksAfterSecond[0].SourceText != second {
		t.Errorf("source_text not updated: got %q", decksAfterSecond[0].SourceText)
	}
}

func TestUploadDeckDoesNotSaveForGuest(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	st := newDeckLibraryStack(t, idx)

	meta, _ := st.lobby.Create("FNM")
	resp := postJSON(t, st.srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Guest"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	source := "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"
	resp = postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{Format: "text", Source: source, PlayerID: joined.PlayerID})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
	resp.Body.Close()

	// A guest session has a zero UserID, so nothing was saved to the
	// library and the seat's deck_id stayed empty. SeatInfo.DeckID
	// isn't on the wire (json:"-"), so read it back via the *Lobby
	// directly rather than through the HTTP response.
	after, err := st.lobby.Get(meta.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(after.Players) != 1 || after.Players[0].DeckID != "" {
		t.Errorf("guest seat deck_id = %q, want empty", after.Players[0].DeckID)
	}
}

func TestSeatLibraryDeckHappyPathAndOwnership(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	st := newDeckLibraryStack(t, idx)
	alice := mustLibraryUser(t, st.db, "Alice")
	bob := mustLibraryUser(t, st.db, "Bob")

	source := "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"
	saved, err := st.library.Upsert(context.Background(), alice, "Alice's Deck", "text", source, []string{"Test Commander"}, 100)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	meta, _ := st.lobby.Create("FNM")
	_, aliceSeat, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	_, bobSeat, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Bob")
	aliceTok := signedInToken(t, st.auth, meta.ID, aliceSeat, alice)
	bobTok := signedInToken(t, st.auth, meta.ID, bobSeat, bob)

	// Bob cannot seat Alice's deck.
	resp := postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks/"+saved.ID.String(), bobTok, struct{}{})
	if resp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("bob seating alice's deck: got %d, want 403 (body=%s)", resp.StatusCode, body)
	}
	resp.Body.Close()

	// Alice can.
	resp = postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks/"+saved.ID.String(), aliceTok, struct{}{})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("alice seating her own deck: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
	var out uploadDeckResponse
	_ = json.NewDecoder(resp.Body).Decode(&out)
	resp.Body.Close()
	if out.DeckName != "Alice's Deck" || out.DeckID != saved.ID.String() || out.CardCount != 100 {
		t.Errorf("response = %+v", out)
	}

	// A made-up id is a 404, not a 403 or 500.
	resp = postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks/"+uuid.New().String(), aliceTok, struct{}{})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown deck id: got %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestSeatLibraryDeckSurfacesRevalidationFailure(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	st := newDeckLibraryStack(t, idx)
	alice := mustLibraryUser(t, st.db, "Alice")

	// Save a deck that names a card THIS build's index does not have —
	// as if the catalog moved between when it was saved and when it is
	// seated. Bypasses the upload pipeline's own resolve step on
	// purpose (the point of this test is what happens at SEAT time).
	source := "Commander:\n1 Nonexistent Card\nMainboard:\n99 Plains\n"
	saved, err := st.library.Upsert(context.Background(), alice, "Stale Deck", "text", source, []string{"Nonexistent Card"}, 100)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	meta, _ := st.lobby.Create("FNM")
	_, aliceSeat, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	tok := signedInToken(t, st.auth, meta.ID, aliceSeat, alice)

	resp := postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks/"+saved.ID.String(), tok, struct{}{})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("got %d, want 422 (body=%s)", resp.StatusCode, body)
	}
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	resp.Body.Close()
	violations, _ := out["violations"].([]any)
	if len(violations) == 0 {
		t.Fatalf("422 body has no violations: %v", out)
	}
	first, _ := violations[0].(map[string]any)
	if first["code"] != "unknown_card" {
		t.Errorf("violation code = %v, want unknown_card", first["code"])
	}
}

func TestMyDecksAuth(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	st := newDeckLibraryStack(t, idx)
	alice := mustLibraryUser(t, st.db, "Alice")

	// No credential at all: the auth middleware's own 401.
	resp := doGet(t, st.srv, "/me/decks", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no credential: got %d, want 401", resp.StatusCode)
	}
	resp.Body.Close()

	// A guest RolePlayer session: valid credential, zero UserID.
	meta, _ := st.lobby.Create("FNM")
	joinResp := postJSON(t, st.srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Guest"})
	var joined sessionResponse
	_ = json.NewDecoder(joinResp.Body).Decode(&joined)
	joinResp.Body.Close()

	resp = doGet(t, st.srv, "/me/decks", joined.Token)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("guest session: got %d, want 401", resp.StatusCode)
	}
	resp.Body.Close()

	// An admin session: valid credential, also zero UserID.
	adminResp := postJSON(t, st.srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var adminSession sessionResponse
	_ = json.NewDecoder(adminResp.Body).Decode(&adminSession)
	adminResp.Body.Close()

	resp = doGet(t, st.srv, "/me/decks", adminSession.Token)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("admin session: got %d, want 401", resp.StatusCode)
	}
	resp.Body.Close()

	// A signed-in player sees their own decks, newest updated first.
	if _, err := st.library.Upsert(context.Background(), alice, "Older", "text", "x", []string{"A"}, 100); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	// SQLStore truncates updated_at to the MILLISECOND and the listing
	// breaks a tie with `ORDER BY updated_at DESC, id` — a random UUID.
	// Two upserts inside one millisecond therefore come back in an
	// arbitrary order, and this assertion is about which is newer. The
	// sleep is what makes "Newer" actually newer; the flake it removes
	// was latent (it surfaced when an unrelated package's init got
	// slower) and the store-side gap it papers over is filed
	// separately.
	time.Sleep(2 * time.Millisecond)
	if _, err := st.library.Upsert(context.Background(), alice, "Newer", "text", "y", []string{"B"}, 99); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	_, aliceSeat, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	tok := signedInToken(t, st.auth, meta.ID, aliceSeat, alice)

	resp = doGet(t, st.srv, "/me/decks", tok)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("signed-in: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
	var out myDecksResponse
	_ = json.NewDecoder(resp.Body).Decode(&out)
	resp.Body.Close()
	if len(out.Decks) != 2 {
		t.Fatalf("decks = %d, want 2", len(out.Decks))
	}
	if out.Decks[0].Name != "Newer" || out.Decks[1].Name != "Older" {
		t.Errorf("order = [%s, %s], want [Newer, Older]", out.Decks[0].Name, out.Decks[1].Name)
	}
	if out.Decks[0].CardCount != 99 || len(out.Decks[0].Commanders) != 1 || out.Decks[0].Commanders[0] != "B" {
		t.Errorf("decks[0] = %+v", out.Decks[0])
	}
}

// TestDeckLibraryHandlersFallBackCleanlyWithNoStore pins the
// defensive path: a deployment with no database wires
// decklibrary.NoStore (see newDeckLibraryStore in cmd/server), and the
// two routes that touch it must not panic or 500 even in the
// unreachable-in-production case of a principal that somehow still
// carries a UserID.
func TestDeckLibraryHandlersFallBackCleanlyWithNoStore(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()
	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&WSAuthorizer{Auth: a})

	cfg := Config{Lobby: l, Auth: a, Cards: idx, Log: log} // DeckLibrary left nil
	mux := http.NewServeMux()
	mux.Handle("/", Handler(cfg))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	meta, _ := l.Create("FNM")
	_, playerID, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	tok := signedInToken(t, a, meta.ID, playerID, uuid.New())

	resp := doGet(t, srv, "/me/decks", tok)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET /me/decks with no store: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
	var out myDecksResponse
	_ = json.NewDecoder(resp.Body).Decode(&out)
	resp.Body.Close()
	if len(out.Decks) != 0 {
		t.Errorf("decks with no store = %v, want empty", out.Decks)
	}

	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks/"+uuid.New().String(), tok, struct{}{})
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("seat-from-library with no store: got %d, want 503", resp.StatusCode)
	}
	resp.Body.Close()

	// The upload path still works for a "signed-in" caller with no
	// store configured — saveToLibrary's nil check, not a crash.
	source := "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", tok,
		uploadDeckRequest{Format: "text", Source: source, PlayerID: playerID})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload with no store: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
	resp.Body.Close()
}

// --- saveToLibrary gating (white-box; see the function's own doc
// comment for the three deliberate gates) ---

func TestSaveToLibrarySkipsAPrebuiltCatalogPick(t *testing.T) {
	st := newDeckLibraryStack(t, nil)
	owner := mustLibraryUser(t, st.db, "Alice")
	list := &deck.List{Name: "", Commanders: []cards.Card{{Name: "Test Commander"}}}
	p := auth.Principal{Role: auth.RolePlayer, UserID: owner}

	got := saveToLibrary(context.Background(), Config{DeckLibrary: st.library}, p, "izzet-aggro", "text", "1 Test Commander\n", list, []string{"Test Commander"})
	if got != "" {
		t.Errorf("catalog pick: saveToLibrary = %q, want empty", got)
	}
	decks, err := st.library.List(context.Background(), owner)
	if err != nil || len(decks) != 0 {
		t.Errorf("library after catalog pick: %v (err=%v), want empty", decks, err)
	}
}

func TestSaveToLibrarySkipsURLFormat(t *testing.T) {
	st := newDeckLibraryStack(t, nil)
	owner := mustLibraryUser(t, st.db, "Alice")
	list := &deck.List{Name: "URL Deck", Commanders: []cards.Card{{Name: "Test Commander"}}}
	p := auth.Principal{Role: auth.RolePlayer, UserID: owner}

	got := saveToLibrary(context.Background(), Config{DeckLibrary: st.library}, p, "", "url", "https://moxfield.com/decks/abc", list, []string{"Test Commander"})
	if got != "" {
		t.Errorf("url format: saveToLibrary = %q, want empty", got)
	}
	decks, err := st.library.List(context.Background(), owner)
	if err != nil || len(decks) != 0 {
		t.Errorf("library after url upload: %v (err=%v), want empty", decks, err)
	}
}

func TestSaveToLibrarySkipsGuest(t *testing.T) {
	st := newDeckLibraryStack(t, nil)
	list := &deck.List{Name: "Guest Deck", Commanders: []cards.Card{{Name: "Test Commander"}}}
	p := auth.Principal{Role: auth.RolePlayer} // zero UserID

	got := saveToLibrary(context.Background(), Config{DeckLibrary: st.library}, p, "", "text", "1 Test Commander\n", list, []string{"Test Commander"})
	if got != "" {
		t.Errorf("guest: saveToLibrary = %q, want empty", got)
	}
}

func TestSaveToLibraryUsesCommanderFallbackName(t *testing.T) {
	st := newDeckLibraryStack(t, nil)
	owner := mustLibraryUser(t, st.db, "Alice")
	// No name — deck.ParseText never sets one.
	list := &deck.List{Commanders: []cards.Card{{Name: "Krenko, Mob Boss"}}}
	p := auth.Principal{Role: auth.RolePlayer, UserID: owner}

	got := saveToLibrary(context.Background(), Config{DeckLibrary: st.library}, p, "", "text", "1 Krenko, Mob Boss\n", list, []string{"Krenko, Mob Boss"})
	if got == "" {
		t.Fatal("saveToLibrary returned no id")
	}
	saved, err := st.library.Get(context.Background(), uuid.MustParse(got))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if saved.Name != "Krenko, Mob Boss" {
		t.Errorf("name = %q, want the commander's name", saved.Name)
	}
}

func TestDetectDeckFormat(t *testing.T) {
	cases := map[string]string{
		"1 Sol Ring\n":                   "text",
		"  \n{\"name\":\"x\"}":           "moxfield",
		"https://moxfield.com/decks/abc": "url",
		"http://archidekt.com/decks/123": "url",
		"":                               "text",
	}
	for source, want := range cases {
		if got := detectDeckFormat(source); got != want {
			t.Errorf("detectDeckFormat(%q) = %q, want %q", source, got, want)
		}
	}
}
