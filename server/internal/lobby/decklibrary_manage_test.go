package lobby

// decklibrary_manage_test.go covers ADR 0110 section 6: coverage on
// read, the per-deck report, rename, delete, the 200-deck cap's message
// and link imports saved as fetched.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckcoverage"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
)

const manageDeckText = "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"

func doJSON(t *testing.T, srv *httptest.Server, method, path, token string, body any) *http.Response {
	t.Helper()
	var rd io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		rd = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, srv.URL+path, rd)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// manageSetup returns a stack, an owner, and the owner's session token.
func manageSetup(t *testing.T, opts ...func(*Config)) (*deckLibraryStack, uuid.UUID, string, *cards.Index) {
	t.Helper()
	// The library routes share a 1/s, burst-5 bucket per client; the
	// throttle has its own test.
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	idx := buildMinimalDeckIndex(t)
	st := newDeckLibraryStack(t, idx, opts...)
	owner := mustLibraryUser(t, st.db, "Alice")
	meta, _ := st.lobby.Create("FNM")
	_, pid, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	return st, owner, signedInToken(t, st.auth, meta.ID, pid, owner), idx
}

func saveText(t *testing.T, st *deckLibraryStack, owner uuid.UUID, name string) decklibrary.Deck {
	t.Helper()
	d, err := st.library.Upsert(context.Background(), owner, name, "text", manageDeckText, []string{"Test Commander"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func guestToken(t *testing.T, st *deckLibraryStack) string {
	t.Helper()
	meta, _ := st.lobby.Create("G")
	_, pid, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Guest")
	tok, _, err := st.auth.Issue(context.Background(), auth.Principal{Role: auth.RolePlayer, GameID: meta.ID, PlayerID: pid}, 3600e9)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestMyDecksCarriesCoverageMatchingBuild(t *testing.T) {
	st, owner, tok, idx := manageSetup(t)
	d := saveText(t, st, owner, "Mono White")

	resp := doGet(t, st.srv, "/me/decks", tok)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var got myDecksResponse
	_ = json.NewDecoder(resp.Body).Decode(&got)
	if len(got.Decks) != 1 || got.Decks[0].Coverage == nil {
		t.Fatalf("decks = %+v", got.Decks)
	}
	entries, _ := deck.ParseText(manageDeckText)
	want, err := deckcoverage.Build(idx, deckcoverage.Deck{Name: d.Name, Source: "text", Entries: entries})
	if err != nil {
		t.Fatal(err)
	}
	n, m := want.AsPrinted()
	cov := got.Decks[0].Coverage
	if cov.AsPrinted != n || cov.Resolved != m || m == 0 {
		t.Errorf("coverage as_printed/resolved = %d/%d, Build says %d/%d", cov.AsPrinted, cov.Resolved, n, m)
	}
	nc, size := want.AsPrintedCopies()
	if cov.AsPrintedCopies != nc || cov.DeckSize != size || cov.UnknownCopies != want.UnknownCopies || size == 0 {
		t.Errorf("coverage copies = %d of %d (+%d unknown), Build says %d of %d (+%d)",
			cov.AsPrintedCopies, cov.DeckSize, cov.UnknownCopies, nc, size, want.UnknownCopies)
	}
	for _, b := range deckcoverage.Buckets {
		if cov.Copies[b] != want.Copies[b] {
			t.Errorf("copies bucket %s = %d, want %d", b, cov.Copies[b], want.Copies[b])
		}
		if cov.Counts[b] != want.Counts[b] {
			t.Errorf("bucket %s = %d, want %d", b, cov.Counts[b], want.Counts[b])
		}
	}
}

func TestMyDeckCoverageReport(t *testing.T) {
	st, owner, tok, _ := manageSetup(t)
	d := saveText(t, st, owner, "Mono White")
	other := mustLibraryUser(t, st.db, "Bob")
	theirs := saveText(t, st, other, "Bob's")

	resp := doGet(t, st.srv, "/me/decks/"+d.ID.String()+"/coverage", tok)
	var rep deckcoverage.Report
	_ = json.NewDecoder(resp.Body).Decode(&rep)
	resp.Body.Close()
	if resp.StatusCode != 200 || rep.DeckName != "Mono White" || len(rep.Cards) != 2 {
		t.Fatalf("report: %d %+v", resp.StatusCode, rep)
	}

	// Someone else's deck is a 404, never 403 or 200.
	resp = doGet(t, st.srv, "/me/decks/"+theirs.ID.String()+"/coverage", tok)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("other's deck: %d, want 404", resp.StatusCode)
	}
	resp = doGet(t, st.srv, "/me/decks/not-a-uuid/coverage", tok)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("bad id: %d, want 404", resp.StatusCode)
	}
}

func TestLibraryRoutesRefuseCallersWithoutAUser(t *testing.T) {
	st, owner, _, _ := manageSetup(t)
	d := saveText(t, st, owner, "Mono White")
	id := d.ID.String()
	for name, tok := range map[string]string{"guest": guestToken(t, st), "admin": adminToken(t, st.srv)} {
		for _, c := range []struct {
			method, path string
			want         int
			body         any
		}{
			{"GET", "/me/decks", 403, nil},
			{"GET", "/me/decks/" + id + "/coverage", 403, nil},
			{"PATCH", "/me/decks/" + id, 403, map[string]string{"name": "x"}},
			{"DELETE", "/me/decks/" + id, 403, nil},
		} {
			resp := doJSON(t, st.srv, c.method, c.path, tok, c.body)
			resp.Body.Close()
			if resp.StatusCode != c.want {
				t.Errorf("%s %s %s: %d, want %d", name, c.method, c.path, resp.StatusCode, c.want)
			}
		}
	}
	// No credential at all: 401 from the middleware.
	for _, m := range []string{"PATCH", "DELETE"} {
		resp := doJSON(t, st.srv, m, "/me/decks/"+id, "", map[string]string{"name": "x"})
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("anonymous %s: %d, want 401", m, resp.StatusCode)
		}
	}
	// The deck is untouched.
	if got, err := st.library.Get(context.Background(), d.ID); err != nil || got.Name != "Mono White" {
		t.Errorf("deck after refused calls: %+v %v", got, err)
	}
}

func TestRenameDeck(t *testing.T) {
	st, owner, tok, _ := manageSetup(t)
	a := saveText(t, st, owner, "Alpha")
	saveText(t, st, owner, "Beta")
	other := mustLibraryUser(t, st.db, "Bob")
	theirs := saveText(t, st, other, "Bob's")

	resp := doJSON(t, st.srv, "PATCH", "/me/decks/"+a.ID.String(), tok, map[string]string{"name": "  Gamma  "})
	var info myDeckInfo
	_ = json.NewDecoder(resp.Body).Decode(&info)
	resp.Body.Close()
	if resp.StatusCode != 200 || info.Name != "Gamma" {
		t.Fatalf("rename: %d %+v", resp.StatusCode, info)
	}

	for _, c := range []struct {
		name string
		id   uuid.UUID
		body map[string]string
		want int
	}{
		{"conflict", a.ID, map[string]string{"name": "Beta"}, 409},
		{"blank", a.ID, map[string]string{"name": "   "}, 400},
		{"too long", a.ID, map[string]string{"name": string(bytes.Repeat([]byte("x"), 101))}, 400},
		{"not yours", theirs.ID, map[string]string{"name": "Mine"}, 404},
		{"missing", uuid.New(), map[string]string{"name": "Mine"}, 404},
	} {
		resp := doJSON(t, st.srv, "PATCH", "/me/decks/"+c.id.String(), tok, c.body)
		resp.Body.Close()
		if resp.StatusCode != c.want {
			t.Errorf("%s: %d, want %d", c.name, resp.StatusCode, c.want)
		}
	}
	if got, _ := st.library.Get(context.Background(), theirs.ID); got.Name != "Bob's" {
		t.Errorf("someone else's deck was renamed: %q", got.Name)
	}
}

func TestDeleteDeckClearsTheSeat(t *testing.T) {
	st, owner, tok, _ := manageSetup(t)
	meta, _ := st.lobby.Create("T")
	_, pid, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	seatTok := signedInToken(t, st.auth, meta.ID, pid, owner)
	resp := postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks", seatTok,
		uploadDeckRequest{Format: "text", Source: manageDeckText, PlayerID: pid})
	resp.Body.Close()
	decks, _ := st.library.List(context.Background(), owner)
	if len(decks) != 1 {
		t.Fatalf("decks = %d", len(decks))
	}
	other := mustLibraryUser(t, st.db, "Bob")
	theirs := saveText(t, st, other, "Bob's")

	resp = doJSON(t, st.srv, "DELETE", "/me/decks/"+theirs.ID.String(), tok, nil)
	resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("delete other's deck: %d, want 404", resp.StatusCode)
	}
	if _, err := st.library.Get(context.Background(), theirs.ID); err != nil {
		t.Errorf("someone else's deck was deleted: %v", err)
	}

	resp = doJSON(t, st.srv, "DELETE", "/me/decks/"+decks[0].ID.String(), tok, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: %d", resp.StatusCode)
	}
	if _, err := st.library.Get(context.Background(), decks[0].ID); err != decklibrary.ErrNotFound {
		t.Errorf("deck still there: %v", err)
	}
	got, _ := st.lobby.Get(meta.ID)
	for _, p := range got.Players {
		if p.PlayerID == pid && p.DeckID != "" {
			t.Errorf("seat still points at the deleted deck: %q", p.DeckID)
		}
	}
	resp = doJSON(t, st.srv, "DELETE", "/me/decks/"+decks[0].ID.String(), tok, nil)
	resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("second delete: %d, want 404", resp.StatusCode)
	}
}

func TestUploadSaysWhenTheLibraryIsFull(t *testing.T) {
	st, owner, _, _ := manageSetup(t)
	for i := range decklibrary.MaxDecks {
		saveText(t, st, owner, fmt.Sprintf("Deck %03d", i))
	}
	meta, _ := st.lobby.Create("T")
	_, pid, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	tok := signedInToken(t, st.auth, meta.ID, pid, owner)

	resp := postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks", tok,
		uploadDeckRequest{Format: "text", Source: manageDeckText, PlayerID: pid})
	defer resp.Body.Close()
	var body uploadDeckResponse
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if resp.StatusCode != 200 {
		t.Fatalf("upload: %d", resp.StatusCode)
	}
	if body.LibraryNote == "" || body.DeckID != "" {
		t.Errorf("want a library note and no deck id, got %+v", body)
	}
	if n, _ := st.library.Count(context.Background(), owner); n != decklibrary.MaxDecks {
		t.Errorf("count = %d", n)
	}
}

func TestLinkImportIsSavedAsFetchedAndReseatedWithoutTheNetwork(t *testing.T) {
	hits := 0
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		fmt.Fprint(w, `{"name":"Linked Deck","boards":{"commanders":{"cards":{"c1":{"quantity":1,"card":{"name":"Test Commander"}}}},"mainboard":{"cards":{"m1":{"quantity":99,"card":{"name":"Plains"}}}}}}`)
	}))
	t.Cleanup(stub.Close)
	deck.TestingSetMoxfieldAPIHost(t, stub.URL)

	st, owner, _, _ := manageSetup(t, func(c *Config) { c.DeckHTTPClient = stub.Client() })
	meta, _ := st.lobby.Create("T")
	_, pid, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	tok := signedInToken(t, st.auth, meta.ID, pid, owner)

	resp := postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks", tok,
		uploadDeckRequest{Format: "url", Source: "https://www.moxfield.com/decks/xyz789?x=1", PlayerID: pid})
	var up uploadDeckResponse
	_ = json.NewDecoder(resp.Body).Decode(&up)
	resp.Body.Close()
	if resp.StatusCode != 200 || hits != 1 {
		t.Fatalf("import: %d, upstream hits %d", resp.StatusCode, hits)
	}
	decks, _ := st.library.List(context.Background(), owner)
	if len(decks) != 1 {
		t.Fatalf("decks = %d, want 1", len(decks))
	}
	saved := decks[0]
	if saved.Name != "Linked Deck" || saved.SourceFormat != "text" || saved.SourceURL != "https://moxfield.com/decks/xyz789" || saved.CardCount != 100 {
		t.Errorf("saved = %+v", saved)
	}
	if up.DeckID != "" {
		t.Errorf("response deck_id = %q (reserved for pre-built decks)", up.DeckID)
	}

	// The listing carries the link, and re-seating never fetches.
	resp = doGet(t, st.srv, "/me/decks", tok)
	var list myDecksResponse
	_ = json.NewDecoder(resp.Body).Decode(&list)
	resp.Body.Close()
	if list.Decks[0].SourceURL != saved.SourceURL {
		t.Errorf("listed source_url = %q", list.Decks[0].SourceURL)
	}
	stub.Close() // any fetch from here fails
	resp = postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks/"+saved.ID.String(), tok, nil)
	resp.Body.Close()
	if resp.StatusCode != 200 || hits != 1 {
		t.Errorf("reseat: %d, upstream hits %d, want 200 and 1", resp.StatusCode, hits)
	}
}

func TestLibraryReadsAreThrottled(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	st := newDeckLibraryStack(t, idx) // no relax knob
	owner := mustLibraryUser(t, st.db, "Alice")
	meta, _ := st.lobby.Create("T")
	_, pid, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	tok := signedInToken(t, st.auth, meta.ID, pid, owner)
	throttled := false
	for range 12 {
		resp := doGet(t, st.srv, "/me/decks", tok)
		resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests {
			throttled = true
		}
	}
	if !throttled {
		t.Error("12 rapid reads never hit 429")
	}
}
