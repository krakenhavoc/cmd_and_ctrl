package lobby

// decks_save_test.go covers ADR 0112 §3 items 4, 5 and 8 (Delivery PR
// 2): saving a checked deck with POST /me/decks, requesting a saved
// deck's missing cards with POST /deck-requests {deck_id}, and
// saveToLibrary's own-seat guard. The deck fetcher and GitHub are the
// fakes deckcheck_test.go builds; nothing here reaches the network.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckcoverage"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// person signs a person in with Discord and returns their session and
// user ID.
func (s *deckStack) person(t *testing.T, snowflake, name string) (string, uuid.UUID) {
	t.Helper()
	u, err := s.users.UpsertFromDiscord(context.Background(),
		discord.User{ID: snowflake, Username: strings.ToLower(name), GlobalName: name}, "", "identify")
	if err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	tok := s.token(t, auth.Principal{Role: auth.RoleIdentified, UserID: u.ID, DiscordID: snowflake, DiscordGlobalName: name, Name: name})
	return tok, u.ID
}

func (s *deckStack) save(t *testing.T, tok string, body any) (int, saveDeckResponse, string) {
	t.Helper()
	code, raw := s.post(t, "/me/decks", tok, body)
	var out saveDeckResponse
	_ = json.Unmarshal([]byte(raw), &out)
	return code, out, raw
}

// needyText is the needy deck as a pasted list.
func needyText(tc deckcoverage.TestingCards) string {
	return fmt.Sprintf("1 %s *CMDR*\n1 %s\n1 %s\n1 %s\n1 %s\n95 %s\n",
		tc.Commander, tc.Manual, tc.Unreviewed, tc.Caveated, tc.Automated, tc.Basic)
}

// --- POST /me/decks ---

func TestSaveDeckFromAPastedList(t *testing.T) {
	s := newDeckStack(t, true, nil)
	tok, owner := s.person(t, "111", "Alice")
	text := needyText(s.tc)

	code, out, raw := s.save(t, tok, map[string]string{"text": text})
	if code != http.StatusCreated || out.Replaced {
		t.Fatalf("save: %d %s", code, raw)
	}
	// The reply is the GET /me/decks entry, coverage included.
	if out.Deck.Name != s.tc.Commander || out.Deck.CardCount != 100 || out.Deck.SourceURL != "" ||
		len(out.Deck.Commanders) != 1 || out.Deck.Commanders[0] != s.tc.Commander {
		t.Errorf("deck = %+v", out.Deck)
	}
	if out.Deck.Coverage == nil || out.Deck.Coverage.Counts[deckcoverage.Manual] != 1 || out.Deck.Coverage.Resolved == 0 {
		t.Errorf("coverage = %+v", out.Deck.Coverage)
	}

	decks, err := s.lib.List(context.Background(), owner)
	if err != nil || len(decks) != 1 {
		t.Fatalf("library = %v, %v", decks, err)
	}
	if d := decks[0]; d.ID.String() != out.Deck.ID || d.SourceFormat != "text" || d.SourceText != strings.TrimSpace(text) || d.SourceURL != "" {
		t.Errorf("saved row = %+v", d)
	}
	if n := s.source.count(needyDeckURL); n != 0 {
		t.Errorf("a pasted list fetched %d times", n)
	}
}

// A save after a check fetches nothing: the check's cache keeps the
// fetched list beside the report. The link is saved as the list that
// was fetched, with the canonical link beside it.
func TestSaveDeckFromALinkAfterACheckFetchesNothing(t *testing.T) {
	s := newDeckStack(t, true, nil)
	tok, owner := s.person(t, "111", "Alice")

	if code, raw := s.post(t, "/deck-coverage", "", map[string]string{"url": needyDeckURL}); code != http.StatusOK {
		t.Fatalf("check: %d %s", code, raw)
	}
	code, out, raw := s.save(t, tok, map[string]string{"url": "https://www.moxfield.com/decks/needy01/slug?x=1"})
	if code != http.StatusCreated {
		t.Fatalf("save: %d %s", code, raw)
	}
	if n := s.source.count(needyDeckURL); n != 1 {
		t.Errorf("check then save fetched %d times, want 1", n)
	}
	if out.Deck.Name != "Needy Deck" || out.Deck.SourceURL != needyDeckURL || out.Deck.CardCount != 100 {
		t.Errorf("deck = %+v", out.Deck)
	}

	saved, err := s.lib.Get(context.Background(), uuid.MustParse(out.Deck.ID))
	if err != nil || saved.OwnerID != owner {
		t.Fatalf("Get: %+v %v", saved, err)
	}
	// The stored list reads back as the deck that was fetched: the
	// same cards as the pasted needy list, so the same list key.
	code, stored := s.post(t, "/deck-coverage", "", map[string]string{"text": saved.SourceText})
	code2, pasted := s.post(t, "/deck-coverage", "", map[string]string{"text": needyText(s.tc)})
	var a, b deckcoverage.Report
	_ = json.Unmarshal([]byte(stored), &a)
	_ = json.Unmarshal([]byte(pasted), &b)
	if code != http.StatusOK || code2 != http.StatusOK || a.DeckKey == "" || a.DeckKey != b.DeckKey || a.Counts[deckcoverage.Manual] != 1 {
		t.Errorf("stored list = %q\nreport %s\nwant the key of %s", saved.SourceText, stored, pasted)
	}
	if saved.SourceFormat != "text" {
		t.Errorf("source_format = %q", saved.SourceFormat)
	}

	// A save first and a check second fetch once, too: the save fills
	// the cache.
	if code, _, raw := s.save(t, tok, map[string]string{"url": cleanDeckURL}); code != http.StatusCreated {
		t.Fatalf("save clean: %d %s", code, raw)
	}
	if code, raw := s.post(t, "/deck-coverage", "", map[string]string{"url": cleanDeckURL}); code != http.StatusOK {
		t.Fatalf("check clean: %d %s", code, raw)
	}
	if n := s.source.count(cleanDeckURL); n != 1 {
		t.Errorf("save then check fetched %d times, want 1", n)
	}
}

// A save that must fetch spends one token from POST /deck-coverage's
// per-IP bucket, so saving cannot get around the fetch limit. A save
// served from the cache, and a pasted list, spend none.
func TestSaveDeckCacheMissSpendsADeckCoverageToken(t *testing.T) {
	s := newDeckStack(t, false, nil)
	tok, _ := s.person(t, "111", "Alice")

	// The save fetches: one token.
	if code, _, raw := s.save(t, tok, map[string]string{"url": needyDeckURL}); code != http.StatusCreated {
		t.Fatalf("save: %d %s", code, raw)
	}
	// The bucket holds three. The save took one, so two checks pass
	// and the third is refused.
	for i := 0; i < 2; i++ {
		if code, raw := s.post(t, "/deck-coverage", "", map[string]string{"url": cleanDeckURL}); code != http.StatusOK {
			t.Fatalf("check %d: %d %s", i+1, code, raw)
		}
	}
	if code, raw := s.post(t, "/deck-coverage", "", map[string]string{"url": cleanDeckURL}); code != http.StatusTooManyRequests {
		t.Fatalf("third check: %d, want 429: %s", code, raw)
	}

	// The bucket is empty. A save that would fetch is refused before
	// it fetches...
	archie := "https://archidekt.com/decks/777"
	s.source.decks[archie] = s.source.decks[cleanDeckURL]
	code, _, raw := s.save(t, tok, map[string]string{"url": archie})
	if code != http.StatusTooManyRequests || !strings.Contains(raw, "error") {
		t.Fatalf("save past the fetch limit: %d %s", code, raw)
	}
	if n := s.source.count(archie); n != 0 {
		t.Errorf("a refused save fetched %d times", n)
	}
	// ...a save of a deck the cache holds still goes through...
	if code, out, raw := s.save(t, tok, map[string]string{"url": cleanDeckURL}); code != http.StatusCreated || out.Deck.Name != "Clean Deck" {
		t.Errorf("cached save: %d %s", code, raw)
	}
	// ...and so does a pasted list, which fetches nothing.
	if code, _, raw := s.save(t, tok, map[string]string{"text": needyText(s.tc), "name": "Pasted"}); code != http.StatusCreated {
		t.Errorf("pasted save: %d %s", code, raw)
	}
}

// A name already in the library replaces that deck, and the reply
// says so.
func TestSaveDeckNameReuseReplaces(t *testing.T) {
	s := newDeckStack(t, true, nil)
	tok, owner := s.person(t, "111", "Alice")

	code, first, raw := s.save(t, tok, map[string]string{"text": needyText(s.tc), "name": "  Weekend  "})
	if code != http.StatusCreated || first.Replaced || first.Deck.Name != "Weekend" {
		t.Fatalf("first: %d %s", code, raw)
	}
	code, second, raw := s.save(t, tok, map[string]string{"url": cleanDeckURL, "name": "Weekend"})
	if code != http.StatusOK || !second.Replaced || second.Deck.ID != first.Deck.ID {
		t.Fatalf("second: %d %s", code, raw)
	}
	if second.Deck.SourceURL != cleanDeckURL || second.Deck.Coverage == nil || second.Deck.Coverage.Counts[deckcoverage.Manual] != 0 {
		t.Errorf("replaced deck = %+v", second.Deck)
	}
	if n, _ := s.lib.Count(context.Background(), owner); n != 1 {
		t.Errorf("library holds %d decks, want 1", n)
	}
	// Saving the list again under its link's name clears nothing it
	// should not: a pasted save clears the link.
	code, third, raw := s.save(t, tok, map[string]string{"text": needyText(s.tc), "name": "Weekend"})
	if code != http.StatusOK || !third.Replaced || third.Deck.SourceURL != "" {
		t.Errorf("third: %d %s", code, raw)
	}
}

// A new deck past the 200 cap is a 409 that names the cap; replacing a
// deck at the cap is never refused.
func TestSaveDeckPastTheCap(t *testing.T) {
	s := newDeckStack(t, true, nil)
	tok, owner := s.person(t, "111", "Alice")
	for i := range decklibrary.MaxDecks {
		if _, err := s.lib.Upsert(context.Background(), owner, fmt.Sprintf("Deck %03d", i), "text", needyText(s.tc), nil, 100); err != nil {
			t.Fatal(err)
		}
	}
	code, _, raw := s.save(t, tok, map[string]string{"text": needyText(s.tc), "name": "One more"})
	if code != http.StatusConflict {
		t.Fatalf("past the cap: %d %s", code, raw)
	}
	var body struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	_ = json.Unmarshal([]byte(raw), &body)
	if body.Error != "Your deck library is full (200 decks). Delete one below to save this one." || body.Code != "library_full" {
		t.Errorf("409 body = %s", raw)
	}
	if code, out, raw := s.save(t, tok, map[string]string{"text": needyText(s.tc), "name": "Deck 007"}); code != http.StatusOK || !out.Replaced {
		t.Errorf("replace at the cap: %d %s", code, raw)
	}
	if n, _ := s.lib.Count(context.Background(), owner); n != decklibrary.MaxDecks {
		t.Errorf("count = %d", n)
	}
}

// Without a name: the fetched deck's name, then the first commander,
// then "Untitled deck".
func TestSaveDeckNameFallbacks(t *testing.T) {
	s := newDeckStack(t, true, nil)
	tok, _ := s.person(t, "111", "Alice")

	if _, out, raw := s.save(t, tok, map[string]string{"url": needyDeckURL}); out.Deck.Name != "Needy Deck" {
		t.Errorf("link: %s", raw)
	}
	if _, out, raw := s.save(t, tok, map[string]string{"text": needyText(s.tc)}); out.Deck.Name != s.tc.Commander {
		t.Errorf("pasted: %s", raw)
	}
	// A commander typed in lower case is saved under its card name.
	lower := strings.Replace(needyText(s.tc), s.tc.Commander, strings.ToLower(s.tc.Commander), 1)
	if _, out, raw := s.save(t, tok, map[string]string{"text": lower, "name": "Lower"}); len(out.Deck.Commanders) != 1 || out.Deck.Commanders[0] != s.tc.Commander {
		t.Errorf("lower-case commander: %s", raw)
	}
	noCommander := fmt.Sprintf("1 %s\n99 %s\n", s.tc.Manual, s.tc.Basic)
	if _, out, raw := s.save(t, tok, map[string]string{"text": noCommander}); out.Deck.Name != "Untitled deck" {
		t.Errorf("no commander: %s", raw)
	}
}

// Who may save: a signed-in person, in either mode. Anything else is
// one 403 (never a 401, which signs the browser out), and no
// credential at all is auth.Middleware's 401.
func TestSaveDeckCallers(t *testing.T) {
	s := newDeckStack(t, true, nil)
	body := map[string]string{"text": needyText(s.tc)}

	if code, _, raw := s.save(t, "", body); code != http.StatusUnauthorized {
		t.Errorf("anonymous: %d %s", code, raw)
	}
	guest := s.token(t, auth.Principal{Role: auth.RolePlayer, GameID: uuid.New(), PlayerID: uuid.New(), Name: "Guest"})
	if code, _, raw := s.save(t, guest, body); code != http.StatusForbidden {
		t.Errorf("guest: %d %s", code, raw)
	}
	if code, _, raw := s.save(t, s.adminToken(t), body); code != http.StatusForbidden {
		t.Errorf("admin token: %d %s", code, raw)
	}
	_, owner := s.person(t, "111", "Alice")
	seat := s.token(t, auth.Principal{Role: auth.RolePlayer, GameID: uuid.New(), PlayerID: uuid.New(), UserID: owner, DiscordID: "111"})
	if code, _, raw := s.save(t, seat, body); code != http.StatusCreated {
		t.Errorf("a signed-in seat: %d %s", code, raw)
	}
}

func TestSaveDeckBodyShape(t *testing.T) {
	s := newDeckStack(t, true, nil)
	tok, owner := s.person(t, "111", "Alice")
	blocked := "https://moxfield.com/decks/walled"
	s.source.errs[blocked] = fmt.Errorf("%w: moxfield", deck.ErrUpstreamBlocked)

	for _, bad := range []map[string]string{
		{},
		{"url": " ", "text": " "},
		{"url": needyDeckURL, "text": needyText(s.tc)},
		{"text": needyText(s.tc), "name": strings.Repeat("x", 101)},
		{"text": fmt.Sprintf("%d %s\n", deckcoverage.MaxListCopies+1, s.tc.Basic)},
		{"url": "https://example.com/decks/1"},
	} {
		if code, _, raw := s.save(t, tok, bad); code != http.StatusBadRequest {
			t.Errorf("%v: %d %s", bad, code, raw)
		}
	}
	if code, _, raw := s.save(t, tok, map[string]any{"text": needyText(s.tc), "player_id": "x"}); code != http.StatusBadRequest {
		t.Errorf("unknown field: %d %s", code, raw)
	}

	// A Moxfield link fails with the same paste_list hint as the check.
	code, _, raw := s.save(t, tok, map[string]string{"url": blocked})
	var body struct {
		Hint string `json:"hint"`
		Code string `json:"code"`
	}
	_ = json.Unmarshal([]byte(raw), &body)
	if code != http.StatusFailedDependency || body.Hint != deckFetchHintPasteList || body.Code != deck.CodeUpstreamBlocked {
		t.Errorf("blocked Moxfield: %d %s", code, raw)
	}
	if n, _ := s.lib.Count(context.Background(), owner); n != 0 {
		t.Errorf("refused saves left %d decks", n)
	}
}

// The save rides the /me/decks* bucket.
func TestSaveDeckIsThrottled(t *testing.T) {
	s := newDeckStack(t, false, nil)
	tok, _ := s.person(t, "111", "Alice")
	throttled := false
	for i := range 12 {
		code, _, _ := s.save(t, tok, map[string]string{"text": needyText(s.tc), "name": fmt.Sprintf("D%d", i)})
		if code == http.StatusTooManyRequests {
			throttled = true
		}
	}
	if !throttled {
		t.Error("12 rapid saves never hit 429")
	}
}

// --- POST /deck-requests {deck_id} ---

// A saved deck with a link keys as that link: it joins the issue a
// request from the link filed, and its report comes from the stored
// list, never a fetch.
func TestDeckRequestFromASavedLinkDeck(t *testing.T) {
	s := newDeckStack(t, true, nil)
	alice, aliceID := s.person(t, "111", "Alice")
	bob, _ := s.person(t, "222", "Bob")

	saved, err := s.lib.UpsertFromLink(context.Background(), aliceID, "My needy deck", "text", needyText(s.tc), needyDeckURL, []string{s.tc.Commander}, 100)
	if err != nil {
		t.Fatal(err)
	}
	// The deck site is down: the request must not need it.
	s.source.errs[needyDeckURL] = fmt.Errorf("%w: moxfield", deck.ErrUpstreamBlocked)

	code, out, raw := s.request(t, alice, map[string]string{"deck_id": saved.ID.String()})
	if code != http.StatusCreated || out.Status != deckRequestFiled || out.Report == nil {
		t.Fatalf("file: %d %s", code, raw)
	}
	if out.Report.DeckKey != "moxfield:needy01" || out.Report.SourceURL != needyDeckURL || out.Report.Counts[deckcoverage.Manual] != 1 {
		t.Errorf("report = %+v", out.Report)
	}
	if n := s.source.count(needyDeckURL); n != 0 {
		t.Errorf("a saved deck's request fetched %d times", n)
	}
	if row, err := s.store.Lookup(context.Background(), "moxfield:needy01"); err != nil || row.IssueNumber != out.IssueNumber {
		t.Errorf("row = %+v, %v", row, err)
	}

	// Bob asks from the link once the site is back: he joins.
	delete(s.source.errs, needyDeckURL)
	code, out, raw = s.request(t, bob, map[string]string{"url": needyDeckURL})
	if code != http.StatusOK || out.Status != deckRequestJoined || out.IssueNumber != 1 {
		t.Errorf("join from the link: %d %s", code, raw)
	}
}

// Any other saved deck keys as its list: the same cards pasted join
// its issue.
func TestDeckRequestFromASavedPastedDeck(t *testing.T) {
	s := newDeckStack(t, true, nil)
	alice, aliceID := s.person(t, "111", "Alice")
	bob, _ := s.person(t, "222", "Bob")

	saved, err := s.lib.Upsert(context.Background(), aliceID, "Pasted", "text", needyText(s.tc), []string{s.tc.Commander}, 100)
	if err != nil {
		t.Fatal(err)
	}
	code, out, raw := s.request(t, alice, map[string]string{"deck_id": saved.ID.String()})
	if code != http.StatusCreated || out.Status != deckRequestFiled || !deckcoverage.IsListKey(out.Report.DeckKey) {
		t.Fatalf("file: %d %s", code, raw)
	}
	key := out.Report.DeckKey
	// The same cards from another export: sections, other order.
	other := fmt.Sprintf("Commander\n1 %s (TST) 1\nDeck\n95 %s\n1 %s\n1 %s\n1 %s\n1 %s\n",
		s.tc.Commander, s.tc.Basic, s.tc.Automated, s.tc.Caveated, s.tc.Unreviewed, s.tc.Manual)
	code, out, raw = s.request(t, bob, map[string]string{"text": other})
	if code != http.StatusOK || out.Status != deckRequestJoined || out.Report.DeckKey != key {
		t.Errorf("join from a paste: %d %s", code, raw)
	}
}

func TestDeckRequestDeckIDRules(t *testing.T) {
	s := newDeckStack(t, true, nil)
	alice, aliceID := s.person(t, "111", "Alice")
	mallory, _ := s.person(t, "999", "Mallory")
	saved, err := s.lib.Upsert(context.Background(), aliceID, "Mine", "text", needyText(s.tc), []string{s.tc.Commander}, 100)
	if err != nil {
		t.Fatal(err)
	}
	id := saved.ID.String()

	// Someone else's deck, and no deck at all, are the same 404.
	for _, other := range []string{id, uuid.NewString(), "not-a-uuid"} {
		tok := alice
		if other == id {
			tok = mallory
		}
		if code, _, raw := s.request(t, tok, map[string]string{"deck_id": other}); code != http.StatusNotFound {
			t.Errorf("deck_id %q: %d %s", other, code, raw)
		}
	}
	// One input only.
	for _, bad := range []map[string]string{
		{"deck_id": id, "url": needyDeckURL},
		{"deck_id": id, "text": needyText(s.tc)},
	} {
		if code, _, raw := s.request(t, alice, bad); code != http.StatusBadRequest {
			t.Errorf("%v: %d %s", bad, code, raw)
		}
	}
	// The bot has no library: the token's requester path refuses it.
	if code, _, raw := s.request(t, s.adminToken(t), map[string]any{
		"deck_id": id, "requester": map[string]string{"discord_id": "111", "display_name": "Alice"},
	}); code != http.StatusBadRequest {
		t.Errorf("token with deck_id: %d %s", code, raw)
	}
	if issues, _ := s.filer.snapshot(); len(issues) != 0 {
		t.Errorf("%d issues filed by refused requests", len(issues))
	}
}

// deck_id is one more way to name a deck, not a way around the limit:
// three asks per requester per 24 hours, whatever the input.
func TestDeckRequestDeckIDIsRateLimitedPerRequester(t *testing.T) {
	s := newDeckStack(t, true, nil)
	tok, owner := s.person(t, "333", "Carol")
	saved, err := s.lib.Upsert(context.Background(), owner, "Mine", "text", needyText(s.tc), []string{s.tc.Commander}, 100)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	for i, key := range []string{"moxfield:a", "moxfield:b", "archidekt:1"} {
		if err := s.store.RecordAsk(context.Background(), key, "discord:333", now.Add(-time.Duration(3-i)*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	code, out, raw := s.request(t, tok, map[string]string{"deck_id": saved.ID.String()})
	if code != http.StatusTooManyRequests || out.Status != deckRequestRateLimited || out.RetryAfter < 20*3600 {
		t.Fatalf("over the limit: %d %s", code, raw)
	}
	if issues, _ := s.filer.snapshot(); len(issues) != 0 {
		t.Error("an over-limit request filed an issue")
	}
}

// --- saveToLibrary's own-seat guard (ADR 0112 §3 item 8) ---

// An admin seating someone else's deck seats it and saves nothing; the
// guard is recordLastDeck's: the caller's own seat at this game.
func TestSaveToLibrarySkipsSomeoneElsesSeat(t *testing.T) {
	st := newDeckLibraryStack(t, nil)
	owner := mustLibraryUser(t, st.db, "Admin")
	game, mine, theirs := uuid.New(), uuid.New(), uuid.New()
	list := &deck.List{Name: "Their Deck", Commanders: []cards.Card{{Name: "Test Commander"}}}
	p := auth.Principal{Role: auth.RolePlayer, UserID: owner, GameID: game, PlayerID: mine, DiscordID: "111"}
	c := Config{DeckLibrary: st.library}

	for _, tc := range []struct {
		name           string
		game, playerID uuid.UUID
	}{
		{"another seat at this game", game, theirs},
		{"a seat at another game", uuid.New(), mine},
	} {
		got, err := saveToLibrary(context.Background(), c, p, tc.game, tc.playerID, "", "text", "1 Test Commander\n", list, []string{"Test Commander"})
		if got != "" || err != nil {
			t.Errorf("%s: saveToLibrary = %q, %v; want nothing saved", tc.name, got, err)
		}
	}
	if decks, _ := st.library.List(context.Background(), owner); len(decks) != 0 {
		t.Fatalf("library after seating other seats: %d decks, want 0", len(decks))
	}

	// Their own seat saves, as before.
	got, err := saveToLibrary(context.Background(), c, p, game, mine, "", "text", "1 Test Commander\n", list, []string{"Test Commander"})
	if got == "" || err != nil {
		t.Errorf("own seat: saveToLibrary = %q, %v; want a deck id", got, err)
	}
}
