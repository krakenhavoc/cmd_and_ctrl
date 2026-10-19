package lobby

// mydecks_update_test.go covers #2036: updating a saved deck from its
// link, a new link or a pasted list. The deck fetcher and GitHub are the
// fakes deckcheck_test.go builds; nothing here reaches the network.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

const secondEngine = "Testing Second Engine"

// updateStack is newDeckStack with one more uncatalogued card, so an
// update can add a missing card the first version did not have.
func newUpdateStack(t *testing.T, configure func(*Config)) *deckStack {
	t.Helper()
	return newDeckStack(t, true, func(c *Config) {
		c.Cards.Put(cards.Card{
			ID: uuid.New(), OracleID: uuid.New(), Name: secondEngine, TypeLine: "Artifact",
			OracleText: "At the beginning of your end step, do another thing no catalog card does.",
			Legalities: map[string]string{"commander": "legal"},
		})
		if configure != nil {
			configure(c)
		}
	})
}

// saveNeedy saves the needy deck as a pasted list and returns its id.
func (s *deckStack) saveNeedy(t *testing.T, tok string) string {
	t.Helper()
	code, out, raw := s.save(t, tok, map[string]string{"text": needyText(s.tc), "name": "My Deck"})
	if code != http.StatusCreated {
		t.Fatalf("save: %d %s", code, raw)
	}
	return out.Deck.ID
}

// needyPlusSecond is the needy list with the second uncatalogued card
// swapped in for one Forest.
func needyPlusSecond(tc deckcoverage.TestingCards) string {
	return fmt.Sprintf("1 %s *CMDR*\n1 %s\n1 %s\n1 %s\n1 %s\n1 %s\n94 %s\n",
		tc.Commander, tc.Manual, tc.Unreviewed, tc.Caveated, tc.Automated, secondEngine, tc.Basic)
}

func (s *deckStack) preview(t *testing.T, tok, deckID string, body any) (int, updatePreviewResponse, string) {
	t.Helper()
	code, raw := s.post(t, "/me/decks/"+deckID+"/update", tok, body)
	var out updatePreviewResponse
	_ = json.Unmarshal([]byte(raw), &out)
	return code, out, raw
}

func (s *deckStack) confirm(t *testing.T, tok, deckID string, body any) (int, updateConfirmResponse, string) {
	t.Helper()
	code, raw := s.post(t, "/me/decks/"+deckID+"/update/confirm", tok, body)
	var out updateConfirmResponse
	_ = json.Unmarshal([]byte(raw), &out)
	return code, out, raw
}

func (s *deckStack) deck(t *testing.T, id string) decklibrary.Deck {
	t.Helper()
	d, err := s.lib.Get(context.Background(), uuid.MustParse(id))
	if err != nil {
		t.Fatalf("Get %s: %v", id, err)
	}
	return d
}

// --- preview ---

func TestUpdatePreviewFromAPastedListShowsTheDiffAndWritesNothing(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)
	before := s.deck(t, id)

	code, out, raw := s.preview(t, tok, id, map[string]string{"list": needyPlusSecond(s.tc)})
	if code != http.StatusOK {
		t.Fatalf("preview: %d %s", code, raw)
	}
	if out.PreviewID == "" || out.Source != "list" || out.NoChange {
		t.Errorf("preview = %+v", out)
	}
	d := out.Diff
	if len(d.Added) != 1 || d.Added[0].Name != secondEngine || d.Added[0].Count != 1 {
		t.Errorf("added = %+v", d.Added)
	}
	if len(d.Removed) != 0 {
		t.Errorf("removed = %+v", d.Removed)
	}
	if len(d.Changed) != 1 || d.Changed[0].Name != s.tc.Basic || d.Changed[0].From != 95 || d.Changed[0].To != 94 {
		t.Errorf("changed = %+v", d.Changed)
	}
	if d.CommanderChanged {
		t.Errorf("commander flagged as changed: %+v", d)
	}
	if out.CoverageBefore == nil || out.CoverageAfter == nil ||
		out.CoverageBefore.Counts[deckcoverage.Manual] != 1 || out.CoverageAfter.Counts[deckcoverage.Manual] != 2 {
		t.Errorf("coverage before/after = %+v / %+v", out.CoverageBefore, out.CoverageAfter)
	}
	if out.Report == nil || out.Report.Counts[deckcoverage.Manual] != 2 {
		t.Errorf("report = %+v", out.Report)
	}

	// Nothing was written, filed or fetched.
	if after := s.deck(t, id); after.SourceText != before.SourceText || !after.UpdatedAt.Equal(before.UpdatedAt) {
		t.Errorf("preview changed the deck: %+v", after)
	}
	if issues, comments := s.filer.snapshot(); len(issues) != 0 || len(comments) != 0 {
		t.Errorf("preview filed %v / %v", issues, comments)
	}
}

func TestUpdatePreviewReportsACommanderChangeAndViolations(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)

	// A 60-card list with a different commander: legal to preview, with
	// the Commander violations listed and the commander change flagged.
	list := fmt.Sprintf("1 %s *CMDR*\n59 %s\n", s.tc.Vanilla, s.tc.Basic)
	code, out, raw := s.preview(t, tok, id, map[string]string{"list": list})
	if code != http.StatusOK {
		t.Fatalf("preview: %d %s", code, raw)
	}
	if !out.Diff.CommanderChanged || len(out.Diff.CommandersBefore) != 1 || out.Diff.CommandersAfter[0] != s.tc.Vanilla {
		t.Errorf("commander diff = %+v", out.Diff)
	}
	if out.Report == nil || len(out.Report.Violations) == 0 {
		t.Errorf("violations missing: %+v", out.Report)
	}
}

func TestUpdatePreviewRelinkBypassesTheCacheAndSeesTheNewList(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	code, saved, raw := s.save(t, tok, map[string]string{"url": cleanDeckURL})
	if code != http.StatusCreated {
		t.Fatalf("save: %d %s", code, raw)
	}
	id := saved.Deck.ID
	if n := s.source.count(cleanDeckURL); n != 1 {
		t.Fatalf("save fetched %d times", n)
	}

	// The deck is edited on the site; the 10-minute cache still holds
	// the old fetch.
	s.source.mu.Lock()
	s.source.decks[cleanDeckURL] = fakeDeck{name: "Clean Deck", entries: []deck.Entry{
		{Name: s.tc.Commander, Count: 1, IsCommander: true},
		{Name: s.tc.Automated, Count: 1},
		{Name: s.tc.Vanilla, Count: 1},
		{Name: s.tc.Basic, Count: 97},
	}}
	s.source.mu.Unlock()

	code, out, raw := s.preview(t, tok, id, map[string]string{"source": "relink"})
	if code != http.StatusOK {
		t.Fatalf("relink preview: %d %s", code, raw)
	}
	if n := s.source.count(cleanDeckURL); n != 2 {
		t.Errorf("relink fetched %d times in all, want 2 (the cache must be skipped)", n)
	}
	if out.Source != "relink" || out.SourceURL != cleanDeckURL {
		t.Errorf("source = %q %q", out.Source, out.SourceURL)
	}
	if len(out.Diff.Added) != 1 || out.Diff.Added[0].Name != s.tc.Vanilla ||
		len(out.Diff.Removed) != 1 || out.Diff.Removed[0].Name != s.tc.Caveated {
		t.Errorf("diff = %+v", out.Diff)
	}
}

func TestUpdateRelinkNeedsALink(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok) // a pasted deck: no link

	code, raw := s.post(t, "/me/decks/"+id+"/update", tok, map[string]string{"source": "relink"})
	if code != http.StatusBadRequest || !strings.Contains(raw, "no link") {
		t.Errorf("relink on a pasted deck: %d %s", code, raw)
	}
}

func TestUpdateRejectsBadInput(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)

	cases := []struct {
		name string
		body any
		code int
	}{
		{"nothing", map[string]string{}, http.StatusBadRequest},
		{"two inputs", map[string]string{"url": cleanDeckURL, "list": "1 Forest"}, http.StatusBadRequest},
		{"unknown source value", map[string]string{"source": "refetch"}, http.StatusBadRequest},
		{"unknown field", map[string]string{"text": "1 Forest"}, http.StatusBadRequest},
		{"unsupported host", map[string]string{"url": "https://example.com/decks/12"}, http.StatusBadRequest},
		{"loopback", map[string]string{"url": "http://127.0.0.1:8080/decks/1"}, http.StatusBadRequest},
		{"metadata address", map[string]string{"url": "http://169.254.169.254/latest/meta-data"}, http.StatusBadRequest},
		{"link past the length cap", map[string]string{"url": "https://archidekt.com/decks/1?" + strings.Repeat("a", maxUpdateURLLen)}, http.StatusBadRequest},
		{"list with no cards", map[string]string{"list": "   \n\n"}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, raw := s.post(t, "/me/decks/"+id+"/update", tok, tc.body)
			if code != tc.code {
				t.Errorf("%d %s, want %d", code, raw, tc.code)
			}
		})
	}
	// None of the links reached the fetcher.
	for _, u := range []string{"https://example.com/decks/12", "http://127.0.0.1:8080/decks/1", "http://169.254.169.254/latest/meta-data"} {
		if n := s.source.count(u); n != 0 {
			t.Errorf("%s was fetched %d times", u, n)
		}
	}
	// A list over the copy bound is refused as the first import refuses it.
	code, raw := s.post(t, "/me/decks/"+id+"/update", tok,
		map[string]string{"list": fmt.Sprintf("%d %s\n", deckcoverage.MaxListCopies+1, s.tc.Basic)})
	if code != http.StatusBadRequest {
		t.Errorf("oversize list: %d %s", code, raw)
	}
	// A link the site cannot find is the site's answer, passed on.
	code, raw = s.post(t, "/me/decks/"+id+"/update", tok, map[string]string{"url": "https://archidekt.com/decks/9999"})
	if code != http.StatusNotFound {
		t.Errorf("missing deck: %d %s", code, raw)
	}
	// A deck with no cards in it is not an update.
	s.source.mu.Lock()
	s.source.decks["https://archidekt.com/decks/777"] = fakeDeck{name: "Empty", entries: nil}
	s.source.mu.Unlock()
	code, raw = s.post(t, "/me/decks/"+id+"/update", tok, map[string]string{"url": "https://archidekt.com/decks/777"})
	if code != http.StatusUnprocessableEntity {
		t.Errorf("empty deck: %d %s", code, raw)
	}
}

func TestUpdateIsOwnerOnly(t *testing.T) {
	s := newUpdateStack(t, nil)
	alice, _ := s.person(t, "111", "Alice")
	bob, _ := s.person(t, "222", "Bob")
	id := s.saveNeedy(t, alice)
	before := s.deck(t, id)

	_, out, _ := s.preview(t, alice, id, map[string]string{"list": needyPlusSecond(s.tc)})

	for name, tok := range map[string]string{"another person": bob} {
		code, raw := s.post(t, "/me/decks/"+id+"/update", tok, map[string]string{"list": needyPlusSecond(s.tc)})
		if code != http.StatusNotFound {
			t.Errorf("%s previewing: %d %s, want 404", name, code, raw)
		}
		// Bob holds Alice's preview id and still cannot use it.
		code, raw = s.post(t, "/me/decks/"+id+"/update/confirm", tok, map[string]any{"preview_id": out.PreviewID})
		if code != http.StatusNotFound {
			t.Errorf("%s confirming: %d %s, want 404", name, code, raw)
		}
	}
	// No session, and an admin session (not a person), cannot either.
	for _, path := range []string{"/update", "/update/confirm"} {
		if code, _ := s.post(t, "/me/decks/"+id+path, "", map[string]string{"list": "1 Forest"}); code != http.StatusUnauthorized {
			t.Errorf("no session on %s: %d, want 401", path, code)
		}
		if code, _ := s.post(t, "/me/decks/"+id+path, s.adminToken(t), map[string]string{"list": "1 Forest"}); code != http.StatusForbidden {
			t.Errorf("admin on %s: %d, want 403", path, code)
		}
	}
	// Alice's preview survived all of that and is still hers.
	if code, _, raw := s.confirm(t, alice, id, map[string]any{"preview_id": out.PreviewID}); code != http.StatusOK {
		t.Fatalf("alice confirming: %d %s", code, raw)
	}
	if s.deck(t, id).SourceText == before.SourceText {
		t.Error("the owner's own update did not land")
	}
	// A deck id that does not exist is a 404.
	if code, _ := s.post(t, "/me/decks/"+uuid.NewString()+"/update", alice, map[string]string{"list": "1 Forest"}); code != http.StatusNotFound {
		t.Errorf("unknown deck: %d, want 404", code)
	}
}

// --- confirm ---

func TestUpdateConfirmCommitsThePreviewedListAndKeepsTheDeckIdentity(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, owner := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)
	before := s.deck(t, id)

	_, prev, _ := s.preview(t, tok, id, map[string]string{"list": needyPlusSecond(s.tc)})
	code, out, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID})
	if code != http.StatusOK {
		t.Fatalf("confirm: %d %s", code, raw)
	}
	if out.Request != nil {
		t.Errorf("a request was made unasked: %+v", out.Request)
	}
	got := s.deck(t, id)
	if got.ID != before.ID || got.Name != "My Deck" || got.OwnerID != owner || !got.CreatedAt.Equal(before.CreatedAt) {
		t.Errorf("identity changed: %+v", got)
	}
	if got.CardCount != 100 || len(got.Commanders) != 1 || got.SourceFormat != "text" || got.SourceURL != "" {
		t.Errorf("row = %+v", got)
	}
	if !strings.Contains(got.SourceText, secondEngine) || !got.UpdatedAt.After(before.UpdatedAt) {
		t.Errorf("list not replaced: %q", got.SourceText)
	}
	if out.Deck.ID != id || out.Deck.Coverage == nil || out.Deck.Coverage.Counts[deckcoverage.Manual] != 2 {
		t.Errorf("response deck = %+v", out.Deck)
	}
	// The stored text reads back to the same cards.
	entries, err := libraryDeckEntries(got)
	if err != nil {
		t.Fatal(err)
	}
	if d := diffEntries(Config{Cards: nil}, entries, entries); !d.empty() {
		t.Errorf("self diff = %+v", d)
	}
	var seen int
	for _, e := range entries {
		if e.Name == secondEngine && e.Count == 1 {
			seen++
		}
		if e.Name == s.tc.Commander && !e.IsCommander {
			t.Errorf("the commander lost its flag: %+v", e)
		}
	}
	if seen != 1 {
		t.Errorf("entries = %+v", entries)
	}

	// A preview confirms once.
	if code, _, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID}); code != http.StatusGone {
		t.Errorf("second confirm: %d %s, want 410", code, raw)
	}
	// A made-up one, and none at all.
	if code, _, _ := s.confirm(t, tok, id, map[string]any{"preview_id": "nope"}); code != http.StatusGone {
		t.Errorf("unknown preview: %d, want 410", code)
	}
	if code, _, _ := s.confirm(t, tok, id, map[string]any{}); code != http.StatusBadRequest {
		t.Errorf("no preview id: %d, want 400", code)
	}
}

func TestUpdateConfirmCommitsWhatWasPreviewedNotAFreshFetch(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	code, saved, raw := s.save(t, tok, map[string]string{"url": cleanDeckURL})
	if code != http.StatusCreated {
		t.Fatalf("save: %d %s", code, raw)
	}
	id := saved.Deck.ID
	_, prev, _ := s.preview(t, tok, id, map[string]string{"source": "relink"})
	fetches := s.source.count(cleanDeckURL)

	// The site changes between the preview and the confirm.
	s.source.mu.Lock()
	s.source.decks[cleanDeckURL] = fakeDeck{name: "Changed", entries: []deck.Entry{{Name: s.tc.Commander, Count: 1, IsCommander: true}, {Name: s.tc.Basic, Count: 99}}}
	s.source.mu.Unlock()

	if code, _, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID}); code != http.StatusOK {
		t.Fatalf("confirm: %d %s", code, raw)
	}
	if n := s.source.count(cleanDeckURL); n != fetches {
		t.Errorf("confirm fetched again (%d -> %d)", fetches, n)
	}
	got := s.deck(t, id)
	if !strings.Contains(got.SourceText, s.tc.Caveated) || got.SourceURL != cleanDeckURL {
		t.Errorf("saved %q (link %q), want the previewed list and the same link", got.SourceText, got.SourceURL)
	}
}

func TestUpdateConfirmRefusesAPreviewOfAnOlderList(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)
	_, prev, _ := s.preview(t, tok, id, map[string]string{"list": needyPlusSecond(s.tc)})

	// The deck is replaced in the meantime (a re-save by name).
	if code, _, raw := s.save(t, tok, map[string]string{"text": fmt.Sprintf("1 %s *CMDR*\n99 %s\n", s.tc.Commander, s.tc.Basic), "name": "My Deck"}); code != http.StatusOK {
		t.Fatalf("re-save: %d %s", code, raw)
	}
	code, _, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID})
	if code != http.StatusConflict {
		t.Fatalf("stale confirm: %d %s, want 409", code, raw)
	}
	if strings.Contains(s.deck(t, id).SourceText, secondEngine) {
		t.Error("a stale preview overwrote the newer list")
	}
}

func TestUpdateExpiredPreviewIsGone(t *testing.T) {
	var d *deckCheck
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	_, owner := s.person(t, "333", "Carol")
	_ = tok

	// The cache itself: a preview confirms before its expiry and not after.
	d = newDeckCheck()
	now := time.Now().UTC()
	d.updates.nowFunc = func() time.Time { return now }
	deckID := uuid.New()
	p := &pendingUpdate{owner: owner, deckID: deckID}
	if err := d.updates.put(p); err != nil {
		t.Fatal(err)
	}
	if got := d.updates.take(p.id, owner, deckID); got != p {
		t.Fatalf("take before expiry = %v", got)
	}
	q := &pendingUpdate{owner: owner, deckID: deckID}
	_ = d.updates.put(q)
	now = now.Add(updatePreviewTTL + time.Second)
	if got := d.updates.take(q.id, owner, deckID); got != nil {
		t.Errorf("take after expiry = %v", got)
	}
	if d.updates.len() != 0 {
		t.Errorf("expired preview kept: %d", d.updates.len())
	}
}

func TestUpdatePreviewsAreBoundedAndOnePerDeck(t *testing.T) {
	d := newDeckCheck()
	owner := uuid.New()
	deckID := uuid.New()
	first := &pendingUpdate{owner: owner, deckID: deckID}
	_ = d.updates.put(first)
	second := &pendingUpdate{owner: owner, deckID: deckID}
	_ = d.updates.put(second)
	if d.updates.take(first.id, owner, deckID) != nil {
		t.Error("a newer preview of the same deck should replace the older one")
	}
	if d.updates.take(second.id, owner, deckID) == nil {
		t.Error("the newer preview is gone")
	}
	for i := 0; i < updatePreviewMax*2; i++ {
		_ = d.updates.put(&pendingUpdate{owner: owner, deckID: uuid.New()})
	}
	if n := d.updates.len(); n > updatePreviewMax {
		t.Errorf("held %d previews, bound is %d", n, updatePreviewMax)
	}
}

func TestUpdateFromANewLinkReplacesTheLinkAndAPastedListClearsIt(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	code, saved, raw := s.save(t, tok, map[string]string{"url": needyDeckURL, "name": "Linked"})
	if code != http.StatusCreated {
		t.Fatalf("save: %d %s", code, raw)
	}
	id := saved.Deck.ID

	_, prev, _ := s.preview(t, tok, id, map[string]string{"url": cleanDeckURL + "/slug?x=1"})
	if prev.Source != "url" || prev.SourceURL != cleanDeckURL || prev.LinkCleared {
		t.Errorf("new-link preview = %+v", prev)
	}
	if code, _, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID}); code != http.StatusOK {
		t.Fatalf("confirm: %d %s", code, raw)
	}
	if got := s.deck(t, id); got.SourceURL != cleanDeckURL || got.Name != "Linked" {
		t.Errorf("after a new link: %+v", got)
	}

	_, prev, _ = s.preview(t, tok, id, map[string]string{"list": needyText(s.tc)})
	if prev.SourceURL != "" || !prev.LinkCleared {
		t.Errorf("pasted-list preview should say the link goes: %+v", prev)
	}
	if code, _, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID}); code != http.StatusOK {
		t.Fatalf("confirm: %d %s", code, raw)
	}
	if got := s.deck(t, id); got.SourceURL != "" {
		t.Errorf("a pasted list kept the link: %+v", got)
	}
}

func TestUpdateKeepsThePrintingChoicesOfUnchangedCards(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, owner := s.person(t, "111", "Alice")
	// A deck whose Forests and commander carry chosen printings.
	text := fmt.Sprintf("Commander:\n1 %s (TST) 1\nMainboard:\n1 %s\n1 %s\n1 %s\n1 %s\n95 %s (TST) 7\n",
		s.tc.Commander, s.tc.Manual, s.tc.Unreviewed, s.tc.Caveated, s.tc.Automated, s.tc.Basic)
	saved, err := s.lib.Upsert(context.Background(), owner, "Printed", "text", text, []string{s.tc.Commander}, 100)
	if err != nil {
		t.Fatal(err)
	}
	id := saved.ID.String()

	// The pasted update names no printings and changes the Forest count;
	// it also brings a card the old list did not have.
	list := fmt.Sprintf("1 %s *CMDR*\n1 %s\n1 %s\n1 %s\n1 %s\n1 %s\n94 %s\n",
		s.tc.Commander, s.tc.Manual, s.tc.Unreviewed, s.tc.Caveated, s.tc.Automated, secondEngine, s.tc.Basic)
	_, prev, raw := s.preview(t, tok, id, map[string]string{"list": list})
	if prev.PrintingsKept != 2 {
		t.Fatalf("printings kept = %d, want 2 (commander and Forest): %s", prev.PrintingsKept, raw)
	}
	if code, _, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID}); code != http.StatusOK {
		t.Fatalf("confirm: %d %s", code, raw)
	}
	got := s.deck(t, id)
	for _, want := range []string{
		fmt.Sprintf("1 %s (TST) 1", s.tc.Commander),
		fmt.Sprintf("94 %s (TST) 7", s.tc.Basic),
		fmt.Sprintf("1 %s\n", secondEngine), // no printing invented for a new card
	} {
		if !strings.Contains(got.SourceText, want) {
			t.Errorf("stored list lacks %q:\n%s", want, got.SourceText)
		}
	}
	if strings.Contains(got.SourceText, secondEngine+" (") {
		t.Errorf("a new card was given a printing:\n%s", got.SourceText)
	}
}

func TestCarryPrintingsKeepsDistinctBasicLandPrintingsInOrder(t *testing.T) {
	old := []deck.Entry{
		{Name: "Forest", Count: 5, SetCode: "AAA", CollectorNo: "1"},
		{Name: "Forest", Count: 5, SetCode: "BBB", CollectorNo: "2"},
		{Name: "Sol Ring", Count: 1},
	}
	fresh := []deck.Entry{
		{Name: "forest", Count: 4},
		{Name: "Forest", Count: 6},
		{Name: "Forest", Count: 1},
		{Name: "Sol Ring", Count: 1, SetCode: "NEW", CollectorNo: "9"},
	}
	got, kept := carryPrintings(Config{}, old, fresh)
	if kept != 2 || got[0].SetCode != "AAA" || got[1].SetCode != "BBB" || got[2].SetCode != "" {
		t.Errorf("carried = %+v (kept %d)", got, kept)
	}
	// The old row had no printing: the new source's own is kept.
	if got[3].SetCode != "NEW" {
		t.Errorf("a printing from the new source was dropped: %+v", got[3])
	}
	if fresh[0].SetCode != "" {
		t.Error("carryPrintings modified its input")
	}
}

func TestEntriesTextRoundTripsThroughTheParser(t *testing.T) {
	in := []deck.Entry{
		{Name: "Testing Commander", Count: 1, IsCommander: true, SetCode: "TST", CollectorNo: "1"},
		{Name: "Forest", Count: 30, SetCode: "TST", CollectorNo: "7"},
		{Name: "Forest", Count: 9},
		{Name: "Fire // Ice", Count: 1},
		{Name: "Sideboard Card", Count: 2, IsSideboard: true},
	}
	out, err := deck.ParseText(entriesTextKeepingPrintings(Config{}, in))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != len(in) {
		t.Fatalf("round trip = %+v", out)
	}
	for i := range in {
		if out[i].Name != in[i].Name || out[i].Count != in[i].Count || out[i].IsCommander != in[i].IsCommander ||
			out[i].IsSideboard != in[i].IsSideboard || !strings.EqualFold(out[i].SetCode, in[i].SetCode) || out[i].CollectorNo != in[i].CollectorNo {
			t.Errorf("row %d: %+v became %+v", i, in[i], out[i])
		}
	}
}

func TestDiffEntriesIgnoresAReorderAndCountsOnlyRealChanges(t *testing.T) {
	a := []deck.Entry{{Name: "Forest", Count: 10}, {Name: "Sol Ring", Count: 1}, {Name: "Bear", Count: 1, IsSideboard: true}}
	b := []deck.Entry{{Name: "sol ring", Count: 1}, {Name: "Forest", Count: 10}, {Name: "Bear", Count: 1, IsSideboard: true}}
	if d := diffEntries(Config{}, a, b); !d.empty() || d.Unchanged != 3 {
		t.Errorf("a reorder and a case change is a diff: %+v", d)
	}
	// A card moved to the sideboard is a removal and an addition.
	c := []deck.Entry{{Name: "Forest", Count: 10}, {Name: "Sol Ring", Count: 1, IsSideboard: true}, {Name: "Bear", Count: 1, IsSideboard: true}}
	d := diffEntries(Config{}, a, c)
	if len(d.Added) != 1 || !d.Added[0].Sideboard || len(d.Removed) != 1 || d.Removed[0].Sideboard {
		t.Errorf("sideboard move = %+v", d)
	}
}

// --- the live table ---

// A deck seated at a table keeps the cards it was seated with: the
// update changes the library row and nothing at the table, and the next
// game sits the new list.
func TestUpdateLeavesATableSeatedWithTheDeckUntouched(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	idx.Put(cards.Card{ID: uuid.New(), Name: "Wastes", TypeLine: "Basic Land", ColorIdentity: []string{},
		Legalities: map[string]string{"commander": "legal"}})
	st := newDeckLibraryStack(t, idx)
	alice := mustLibraryUser(t, st.db, "Alice")

	saved, err := st.library.Upsert(context.Background(), alice, "Alice's Deck", "text",
		"Commander:\n1 Test Commander\nMainboard:\n99 Plains\n", []string{"Test Commander"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	meta, _ := st.lobby.Create("FNM")
	_, seat, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	tok := signedInToken(t, st.auth, meta.ID, seat, alice)
	resp := postJSON(t, st.srv, "/games/"+meta.ID.String()+"/decks/"+saved.ID.String(), tok, struct{}{})
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("seating: %d", resp.StatusCode)
	}

	countIn := func(gameID uuid.UUID, name string) int {
		view, _, err := st.lobby.RoomOf(gameID).Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, p := range view.Seats {
			for _, c := range append(append([]protocol.CardView{}, p.Library.Cards...), p.Hand.Cards...) {
				if c.Name == name {
					n++
				}
			}
		}
		return n
	}
	if countIn(meta.ID, "Plains") == 0 || countIn(meta.ID, "Wastes") != 0 {
		t.Fatalf("seated deck is not the Plains list: plains %d wastes %d", countIn(meta.ID, "Plains"), countIn(meta.ID, "Wastes"))
	}

	// The preview knows a table holds the deck.
	var prev updatePreviewResponse
	postDecode(t, st, "/me/decks/"+saved.ID.String()+"/update", tok,
		map[string]string{"list": "1 Test Commander *CMDR*\n50 Plains\n49 Wastes\n"}, http.StatusOK, &prev)
	if prev.LiveTables != 1 {
		t.Errorf("live tables = %d, want 1", prev.LiveTables)
	}
	postDecode(t, st, "/me/decks/"+saved.ID.String()+"/update/confirm", tok,
		map[string]any{"preview_id": prev.PreviewID}, http.StatusOK, nil)

	// The library row changed; the table did not.
	if got, _ := st.library.Get(context.Background(), saved.ID); !strings.Contains(got.SourceText, "Wastes") {
		t.Fatalf("library row not updated: %q", got.SourceText)
	}
	if countIn(meta.ID, "Wastes") != 0 || countIn(meta.ID, "Plains") == 0 {
		t.Errorf("the seated deck changed under the table: plains %d wastes %d", countIn(meta.ID, "Plains"), countIn(meta.ID, "Wastes"))
	}
	if m, _ := st.lobby.Get(meta.ID); len(m.Players) == 0 || m.Players[0].DeckID != saved.ID.String() {
		t.Errorf("the seat lost its deck link: %+v", m.Players)
	}

	// The next game seats the new list.
	meta2, _ := st.lobby.Create("FNM 2")
	_, seat2, _ := st.lobby.Join(meta2.ID, meta2.InviteToken, "Alice")
	tok2 := signedInToken(t, st.auth, meta2.ID, seat2, alice)
	resp = postJSON(t, st.srv, "/games/"+meta2.ID.String()+"/decks/"+saved.ID.String(), tok2, struct{}{})
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("seating the updated deck: %d", resp.StatusCode)
	}
	if countIn(meta2.ID, "Wastes") == 0 {
		t.Error("the next game did not get the updated list")
	}
}

// postDecode posts body, requires want, and decodes the answer into out.
func postDecode(t *testing.T, st *deckLibraryStack, path, tok string, body any, want int, out any) {
	t.Helper()
	resp := postJSON(t, st.srv, path, tok, body)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != want {
		t.Fatalf("POST %s: %d %s, want %d", path, resp.StatusCode, raw, want)
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			t.Fatalf("decode %s: %v", raw, err)
		}
	}
}

// --- deck requests ---

func (s *deckStack) requestSaved(t *testing.T, tok, deckID string) deckRequestResponse {
	t.Helper()
	code, out, raw := s.request(t, tok, map[string]string{"deck_id": deckID})
	if code != http.StatusOK && code != http.StatusCreated {
		t.Fatalf("request: %d %s", code, raw)
	}
	return out
}

// An update that adds one missing card to a deck whose missing cards were
// already requested extends that request with only the new card.
func TestUpdateRequestExtendsTheOpenIssueWithOnlyTheNewCards(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)
	first := s.requestSaved(t, tok, id)
	if first.Status != deckRequestFiled || first.IssueNumber != 1 {
		t.Fatalf("first request = %+v", first)
	}

	_, prev, raw := s.preview(t, tok, id, map[string]string{"list": needyPlusSecond(s.tc)})
	if !prev.Request.Available || len(prev.Request.NewCards) != 1 || prev.Request.NewCards[0] != secondEngine || prev.Request.AlreadyRequested != 2 {
		t.Fatalf("request preview = %+v\n%s", prev.Request, raw)
	}
	code, out, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID, "request_missing": true})
	if code != http.StatusOK || out.Request == nil {
		t.Fatalf("confirm: %d %s", code, raw)
	}
	if out.Request.Status != updateRequestJoined || out.Request.IssueNumber != 1 || len(out.Request.NewCards) != 1 || out.Request.AlreadyRequested != 2 {
		t.Errorf("request = %+v", out.Request)
	}

	issues, comments := s.filer.snapshot()
	if len(issues) != 1 {
		t.Fatalf("issues = %d, want the one already open", len(issues))
	}
	if len(comments) != 1 || comments[0].number != 1 {
		t.Fatalf("comments = %+v", comments)
	}
	c := comments[0].body
	if !strings.Contains(c, secondEngine) || strings.Contains(c, s.tc.Manual) || strings.Contains(c, s.tc.Unreviewed) {
		t.Errorf("the comment should list only the new card:\n%s", c)
	}
	if strings.Contains(c, "111") {
		t.Errorf("the comment names the requester's id:\n%s", c)
	}

	// The updated deck is now tracked by that issue: asking again from
	// the deck joins it rather than filing a second.
	again := s.requestSaved(t, tok, id)
	if again.Status != deckRequestJoined || again.IssueNumber != 1 {
		t.Errorf("request after the update = %+v", again)
	}
	if issues, _ := s.filer.snapshot(); len(issues) != 1 {
		t.Errorf("issues after the second ask = %d", len(issues))
	}

	// Updating again to a list with nothing new asks for nothing.
	_, prev2, _ := s.preview(t, tok, id, map[string]string{"list": needyPlusSecond(s.tc) + "\n"})
	_ = prev2
	_, prev3, _ := s.preview(t, tok, id, map[string]string{"list": strings.Replace(needyPlusSecond(s.tc), "94 "+s.tc.Basic, "93 "+s.tc.Basic+"\n1 "+s.tc.Vanilla, 1)})
	if len(prev3.Request.NewCards) != 0 {
		t.Errorf("an update that misses nothing new offers %v", prev3.Request.NewCards)
	}
	code, out, _ = s.confirm(t, tok, id, map[string]any{"preview_id": prev3.PreviewID, "request_missing": true})
	if code != http.StatusOK || out.Request == nil || out.Request.Status != updateRequestNothingNew {
		t.Errorf("nothing-new confirm = %d %+v", code, out.Request)
	}
	if _, comments := s.filer.snapshot(); len(comments) != 1 {
		t.Errorf("a nothing-new update commented: %d", len(comments))
	}
}

// A link deck keeps its deck key through a re-pull, so the same issue
// is extended in place.
func TestUpdateRelinkRequestJoinsTheSameIssue(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	code, saved, raw := s.save(t, tok, map[string]string{"url": needyDeckURL})
	if code != http.StatusCreated {
		t.Fatalf("save: %d %s", code, raw)
	}
	id := saved.Deck.ID
	if first := s.requestSaved(t, tok, id); first.Status != deckRequestFiled {
		t.Fatalf("first request = %+v", first)
	}

	s.source.mu.Lock()
	d := s.source.decks[needyDeckURL]
	d.entries = append([]deck.Entry{{Name: secondEngine, Count: 1}}, d.entries[:len(d.entries)-1]...)
	d.entries = append(d.entries, deck.Entry{Name: s.tc.Basic, Count: 94})
	s.source.decks[needyDeckURL] = d
	s.source.mu.Unlock()

	_, prev, raw := s.preview(t, tok, id, map[string]string{"source": "relink"})
	if len(prev.Request.NewCards) != 1 || prev.Request.NewCards[0] != secondEngine {
		t.Fatalf("new cards = %+v\n%s", prev.Request, raw)
	}
	_, out, _ := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID, "request_missing": true})
	if out.Request == nil || out.Request.Status != updateRequestJoined || out.Request.IssueNumber != 1 {
		t.Fatalf("request = %+v", out.Request)
	}
	issues, comments := s.filer.snapshot()
	if len(issues) != 1 || len(comments) != 1 || strings.Contains(comments[0].body, s.tc.Manual) {
		t.Errorf("issues %d, comments %+v", len(issues), comments)
	}
}

// With no earlier request, or one that was closed, a fresh issue lists
// the cards that are new to it.
func TestUpdateRequestFilesANewIssueWhenNothingIsOpen(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)

	// No earlier request: every missing card is new.
	_, prev, _ := s.preview(t, tok, id, map[string]string{"list": needyPlusSecond(s.tc)})
	if len(prev.Request.NewCards) != 3 || prev.Request.AlreadyRequested != 0 {
		t.Fatalf("request preview = %+v", prev.Request)
	}
	_, out, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID, "request_missing": true})
	if out.Request == nil || out.Request.Status != updateRequestFiled || out.Request.IssueNumber != 1 {
		t.Fatalf("request = %+v\n%s", out.Request, raw)
	}
	issues, _ := s.filer.snapshot()
	body := issues[1].body
	for _, want := range []string{s.tc.Manual, s.tc.Unreviewed, secondEngine} {
		if !strings.Contains(body, "- [ ] "+want) {
			t.Errorf("issue lacks %q:\n%s", want, body)
		}
	}
	if !strings.HasPrefix(issues[1].title, deckRequestTitlePrefix) {
		t.Errorf("title = %q", issues[1].title)
	}
}

// The earlier issue is closed (its cards landed); the deck then gains one
// more missing card. A new issue is filed, with only that card.
func TestUpdateRequestAfterAClosedIssueFilesOnlyTheNewCard(t *testing.T) {
	var idx *cards.Index
	s := newUpdateStack(t, func(c *Config) { idx = c.Cards })
	idx.Put(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Testing Third Engine", TypeLine: "Artifact",
		OracleText: "Do a third thing no catalog card does.", Legalities: map[string]string{"commander": "legal"}})
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)
	if first := s.requestSaved(t, tok, id); first.Status != deckRequestFiled {
		t.Fatalf("first request = %+v", first)
	}
	s.filer.close(1)

	list := needyText(s.tc)
	list = strings.Replace(list, "95 "+s.tc.Basic, "94 "+s.tc.Basic+"\n1 Testing Third Engine", 1)
	_, prev, _ := s.preview(t, tok, id, map[string]string{"list": list})
	if len(prev.Request.NewCards) != 1 || prev.Request.NewCards[0] != "Testing Third Engine" {
		t.Fatalf("request preview = %+v", prev.Request)
	}
	_, out, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID, "request_missing": true})
	if out.Request == nil || out.Request.Status != updateRequestFiled || out.Request.IssueNumber != 2 {
		t.Fatalf("request = %+v\n%s", out.Request, raw)
	}
	issues, _ := s.filer.snapshot()
	body := issues[2].body
	if !strings.Contains(body, "- [ ] Testing Third Engine") || strings.Contains(body, "- [ ] "+s.tc.Manual) {
		t.Errorf("the new issue should list only the new card:\n%s", body)
	}
}

func TestUpdateRequestIsOptInAndNeverBlocksTheUpdate(t *testing.T) {
	s := newUpdateStack(t, nil)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)

	// Not asked: nothing is filed even though cards are new.
	_, prev, _ := s.preview(t, tok, id, map[string]string{"list": needyPlusSecond(s.tc)})
	if _, out, _ := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID}); out.Request != nil {
		t.Errorf("request = %+v", out.Request)
	}
	if issues, _ := s.filer.snapshot(); len(issues) != 0 {
		t.Errorf("filed without being asked: %v", issues)
	}

	// Over the daily limit: the update lands, the request is refused, and
	// the answer says when to try again.
	for range deckRequestLimit {
		if err := s.store.RecordAsk(context.Background(), "x:"+uuid.NewString(), "discord:111", time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	list := fmt.Sprintf("1 %s *CMDR*\n1 %s\n98 %s\n", s.tc.Commander, secondEngine, s.tc.Basic)
	_, prev, _ = s.preview(t, tok, id, map[string]string{"list": list})
	code, out, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID, "request_missing": true})
	if code != http.StatusOK || out.Request == nil || out.Request.Status != updateRequestRateLimited || out.Request.RetryAfter <= 0 {
		t.Fatalf("limited confirm = %d %+v\n%s", code, out.Request, raw)
	}
	if !strings.Contains(s.deck(t, id).SourceText, secondEngine) {
		t.Error("the update did not land when the request was refused")
	}
}

func TestUpdateRequestWithoutAFilerSaysSoAndStillUpdates(t *testing.T) {
	s := newUpdateStack(t, func(c *Config) { c.DeckRequestFiler = nil })
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)
	_, prev, _ := s.preview(t, tok, id, map[string]string{"list": needyPlusSecond(s.tc)})
	if prev.Request.Available || prev.Request.Reason == "" {
		t.Errorf("request preview = %+v", prev.Request)
	}
	code, out, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID, "request_missing": true})
	if code != http.StatusOK || out.Request == nil || out.Request.Status != updateRequestUnavailable {
		t.Fatalf("confirm = %d %+v\n%s", code, out.Request, raw)
	}
	if !strings.Contains(s.deck(t, id).SourceText, secondEngine) {
		t.Error("the update did not land")
	}
}

func TestUpdateRequestNeedsADiscordIdentity(t *testing.T) {
	s := newUpdateStack(t, nil)
	// A person with a users row and no Discord identity.
	other := mustLibraryUser(t, s.dbase, "Nobody")
	noDiscord := s.token(t, auth.Principal{Role: auth.RoleIdentified, UserID: other, Name: "Nobody"})
	saved, err := s.lib.Upsert(context.Background(), other, "Mine", "text", needyText(s.tc), []string{s.tc.Commander}, 100)
	if err != nil {
		t.Fatal(err)
	}
	_, prev, _ := s.preview(t, noDiscord, saved.ID.String(), map[string]string{"list": needyPlusSecond(s.tc)})
	if prev.Request.Available || !strings.Contains(prev.Request.Reason, "Discord") {
		t.Errorf("request preview = %+v", prev.Request)
	}
	_, out, _ := s.confirm(t, noDiscord, saved.ID.String(), map[string]any{"preview_id": prev.PreviewID, "request_missing": true})
	if out.Request == nil || out.Request.Status != updateRequestUnavailable {
		t.Errorf("request = %+v", out.Request)
	}
}

// --- limits ---

func TestUpdateRoutesRideTheLibraryRateLimit(t *testing.T) {
	s := newDeckStack(t, false, nil)
	tok, _ := s.person(t, "111", "Alice")
	limited := 0
	for range 20 {
		code, _ := s.post(t, "/me/decks/"+uuid.NewString()+"/update", tok, map[string]string{"list": "1 Forest"})
		if code == http.StatusTooManyRequests {
			limited++
		}
	}
	if limited == 0 {
		t.Error("20 rapid updates were never limited")
	}
}

func TestUpdateLinkSpendsTheFetchBucket(t *testing.T) {
	s := newDeckStack(t, false, nil)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)
	// POST /deck-coverage's bucket holds three. Every link update spends
	// one, cache or not, so the fourth is refused before it fetches (the
	// library bucket allows five calls in a burst: the save, then these).
	for i := 1; i <= 3; i++ {
		if code, raw := s.post(t, "/me/decks/"+id+"/update", tok, map[string]string{"url": cleanDeckURL}); code != http.StatusOK {
			t.Fatalf("update %d: %d %s", i, code, raw)
		}
	}
	code, raw := s.post(t, "/me/decks/"+id+"/update", tok, map[string]string{"url": cleanDeckURL})
	if code != http.StatusTooManyRequests {
		t.Fatalf("fourth link update: %d %s, want 429", code, raw)
	}
	if n := s.source.count(cleanDeckURL); n != 3 {
		t.Errorf("fetched %d times, want 3 (every update skips the cache; the refused one fetches nothing)", n)
	}
}
