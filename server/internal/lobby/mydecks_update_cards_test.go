package lobby

// mydecks_update_cards_test.go covers #4135: the cards each deck request
// listed are remembered (migration 0028), so a deck update de-duplicates
// against what an issue actually lists and not against a guess from the
// old version's key.

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckcoverage"
)

const thirdEngine = "Testing Third Engine"

// newCardsStack is newUpdateStack with a third uncatalogued card.
func newCardsStack(t *testing.T) *deckStack {
	t.Helper()
	var idx *cards.Index
	s := newUpdateStack(t, func(c *Config) { idx = c.Cards })
	idx.Put(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: thirdEngine, TypeLine: "Artifact",
		OracleText: "Do a third thing no catalog card does.", Legalities: map[string]string{"commander": "legal"}})
	return s
}

// withSecond is the needy list plus the second uncatalogued card; the
// vanilla card swaps a basic for something that is not missing, so two
// lists with the same missing cards still differ (and key differently).
func withSecond(tc deckcoverage.TestingCards, vanilla bool) string {
	if !vanilla {
		return needyPlusSecond(tc)
	}
	return strings.Replace(needyPlusSecond(tc), "94 "+tc.Basic, "93 "+tc.Basic+"\n1 "+tc.Vanilla, 1)
}

func withSecondAndThird(tc deckcoverage.TestingCards) string {
	return fmt.Sprintf("1 %s *CMDR*\n1 %s\n1 %s\n1 %s\n1 %s\n1 %s\n1 %s\n93 %s\n",
		tc.Commander, tc.Manual, tc.Unreviewed, tc.Caveated, tc.Automated, secondEngine, thirdEngine, tc.Basic)
}

func sortedNames(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func sameNames(a, b []string) bool {
	a, b = sortedNames(a), sortedNames(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// A filed request records the cards its issue lists.
func TestADeckRequestRecordsTheCardsItsIssueListed(t *testing.T) {
	s := newCardsStack(t)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)
	if first := s.requestSaved(t, tok, id); first.Status != deckRequestFiled {
		t.Fatalf("request = %+v", first)
	}
	key := s.requestKey(t)
	got, err := s.store.CardsOf(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, c := range got {
		if c.IssueNumber != 1 {
			t.Errorf("%s is recorded under issue %d, want 1", c.Name, c.IssueNumber)
		}
		names = append(names, c.Name)
	}
	if want := []string{s.tc.Manual, s.tc.Unreviewed}; !sameNames(names, want) {
		t.Errorf("recorded cards = %v, want %v", names, want)
	}
}

// requestKey is the deck key of the only request filed so far.
func (s *deckStack) requestKey(t *testing.T) string {
	t.Helper()
	var key string
	if err := s.dbase.DB.QueryRow(`SELECT deck_key FROM deck_requests`).Scan(&key); err != nil {
		t.Fatalf("reading the request's deck key: %v", err)
	}
	return key
}

// The player declines to request at one update and updates again: the
// cards the first issue listed are not offered again, and the card they
// declined is.
func TestDecliningAnUpdateThenUpdatingAgainDoesNotReofferListedCards(t *testing.T) {
	s := newCardsStack(t)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)
	if first := s.requestSaved(t, tok, id); first.Status != deckRequestFiled || first.IssueNumber != 1 {
		t.Fatalf("first request = %+v", first)
	}

	// Update 1 adds a missing card, and the player does not request it.
	_, prev, raw := s.preview(t, tok, id, map[string]string{"list": withSecond(s.tc, false)})
	if prev.Request.AlreadyRequested != 2 || !sameNames(prev.Request.NewCards, []string{secondEngine}) {
		t.Fatalf("update 1 preview = %+v\n%s", prev.Request, raw)
	}
	if code, out, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID}); code != 200 || out.Request != nil {
		t.Fatalf("update 1 confirm = %d %+v\n%s", code, out.Request, raw)
	}

	// Update 2 changes the list again. The first issue's two cards are
	// still on that issue; the declined one was never asked for.
	_, prev, raw = s.preview(t, tok, id, map[string]string{"list": withSecond(s.tc, true)})
	if prev.Request.AlreadyRequested != 2 {
		t.Errorf("update 2 offers cards the first issue listed again: %+v\n%s", prev.Request, raw)
	}
	if !sameNames(prev.Request.NewCards, []string{secondEngine}) {
		t.Errorf("update 2 new cards = %v, want just the declined %q", prev.Request.NewCards, secondEngine)
	}
	_, out, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID, "request_missing": true})
	if out.Request == nil || out.Request.Status != updateRequestJoined || out.Request.IssueNumber != 1 {
		t.Fatalf("update 2 request = %+v\n%s", out.Request, raw)
	}
	_, comments := s.filer.snapshot()
	if len(comments) != 1 {
		t.Fatalf("comments = %+v", comments)
	}
	body := comments[0].body
	if !strings.Contains(body, "- [ ] "+secondEngine) || strings.Contains(body, "- [ ] "+s.tc.Manual) || strings.Contains(body, "- [ ] "+s.tc.Unreviewed) {
		t.Errorf("the comment should list only the declined card:\n%s", body)
	}
}

// A comment never repeats a card an earlier comment or the issue
// already listed, across several updates.
func TestAnUpdateCommentNeverRepeatsACardTheIssueListed(t *testing.T) {
	s := newCardsStack(t)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)
	s.requestSaved(t, tok, id)

	step := func(list string, wantNew []string) string {
		t.Helper()
		_, prev, raw := s.preview(t, tok, id, map[string]string{"list": list})
		if !sameNames(prev.Request.NewCards, wantNew) {
			t.Fatalf("new cards = %v, want %v\n%s", prev.Request.NewCards, wantNew, raw)
		}
		_, out, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID, "request_missing": true})
		if out.Request == nil || out.Request.Status != updateRequestJoined {
			t.Fatalf("request = %+v\n%s", out.Request, raw)
		}
		_, comments := s.filer.snapshot()
		return comments[len(comments)-1].body
	}

	second := step(withSecond(s.tc, false), []string{secondEngine})
	third := step(withSecondAndThird(s.tc), []string{thirdEngine})
	for _, repeated := range []string{s.tc.Manual, s.tc.Unreviewed, secondEngine} {
		if strings.Contains(third, "- [ ] "+repeated) {
			t.Errorf("the second comment repeats %q:\n%s", repeated, third)
		}
	}
	if !strings.Contains(second, "- [ ] "+secondEngine) || !strings.Contains(third, "- [ ] "+thirdEngine) {
		t.Errorf("comments lost their own card:\n%s\n---\n%s", second, third)
	}

	// The same list again asks for nothing.
	_, prev, _ := s.preview(t, tok, id, map[string]string{"list": withSecondAndThird(s.tc) + "\n"})
	if len(prev.Request.NewCards) != 0 || prev.Request.AlreadyRequested != 4 {
		t.Errorf("a repeat offers %+v", prev.Request)
	}
}

// A link keeps its deck key through a re-pull, so the old rule read the
// declined update's missing card as already requested and never asked
// for it. The recorded cards know it was not.
func TestADeclinedLinkUpdateIsStillOfferedAtTheNextOne(t *testing.T) {
	s := newCardsStack(t)
	tok, _ := s.person(t, "111", "Alice")
	code, saved, raw := s.save(t, tok, map[string]string{"url": needyDeckURL})
	if code != 201 {
		t.Fatalf("save: %d %s", code, raw)
	}
	id := saved.Deck.ID
	if first := s.requestSaved(t, tok, id); first.Status != deckRequestFiled {
		t.Fatalf("first request = %+v", first)
	}
	addCard := func(name string) {
		s.source.mu.Lock()
		defer s.source.mu.Unlock()
		d := s.source.decks[needyDeckURL]
		d.entries = append([]deck.Entry{{Name: name, Count: 1}}, d.entries[:len(d.entries)-1]...)
		d.entries = append(d.entries, deck.Entry{Name: s.tc.Basic, Count: 1})
		s.source.decks[needyDeckURL] = d
	}

	addCard(secondEngine)
	_, prev, _ := s.preview(t, tok, id, map[string]string{"source": "relink"})
	if code, out, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID}); code != 200 || out.Request != nil {
		t.Fatalf("declined confirm = %d %+v\n%s", code, out.Request, raw)
	}

	addCard(thirdEngine)
	_, prev, raw = s.preview(t, tok, id, map[string]string{"source": "relink"})
	if !sameNames(prev.Request.NewCards, []string{secondEngine, thirdEngine}) || prev.Request.AlreadyRequested != 2 {
		t.Fatalf("preview = %+v\n%s", prev.Request, raw)
	}
	_, out, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID, "request_missing": true})
	if out.Request == nil || out.Request.Status != updateRequestJoined {
		t.Fatalf("request = %+v\n%s", out.Request, raw)
	}
	_, comments := s.filer.snapshot()
	body := comments[len(comments)-1].body
	for _, want := range []string{secondEngine, thirdEngine} {
		if !strings.Contains(body, "- [ ] "+want) {
			t.Errorf("the comment lacks %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "- [ ] "+s.tc.Manual) {
		t.Errorf("the comment repeats a listed card:\n%s", body)
	}
}

// A request filed before the cards were remembered has a row and no card
// rows. It keeps working by the older rule, and an update carries it.
func TestARequestFromBeforeCardsWereRememberedFallsBackAndIsCarried(t *testing.T) {
	s := newCardsStack(t)
	tok, _ := s.person(t, "111", "Alice")
	id := s.saveNeedy(t, tok)
	s.requestSaved(t, tok, id)
	if _, err := s.dbase.DB.Exec(`DELETE FROM deck_request_cards`); err != nil {
		t.Fatal(err)
	}

	_, prev, raw := s.preview(t, tok, id, map[string]string{"list": withSecond(s.tc, false)})
	if prev.Request.AlreadyRequested != 2 || !sameNames(prev.Request.NewCards, []string{secondEngine}) {
		t.Fatalf("fallback preview = %+v\n%s", prev.Request, raw)
	}
	// Declined: the list's key changes, and the old request's cards
	// move to the new key rather than being lost.
	if code, _, raw := s.confirm(t, tok, id, map[string]any{"preview_id": prev.PreviewID}); code != 200 {
		t.Fatalf("confirm: %d %s", code, raw)
	}
	_, prev, raw = s.preview(t, tok, id, map[string]string{"list": withSecond(s.tc, true)})
	if prev.Request.AlreadyRequested != 2 || !sameNames(prev.Request.NewCards, []string{secondEngine}) {
		t.Fatalf("carried preview = %+v\n%s", prev.Request, raw)
	}
}
