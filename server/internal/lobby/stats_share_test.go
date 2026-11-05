package lobby

// Tests for stats_share.go (ADR 0217 §6, slice 6): minting, rotating
// and revoking a share link, and what the shared view shows and never
// shows.

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// newShareStack is the advanced stack with a user store (the shared view
// names its owner) and the basic record wired in.
func newShareStack(t *testing.T) advancedStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	d := openTestDB(t, t.TempDir())
	idx, oracle := advancedCards()
	stats, library := gamestats.NewSQLStore(d), decklibrary.NewSQLStore(d)
	s := newAdminStackWith(t, idx, func(c *Config) {
		c.StatsPrefs = gamestats.NewSQLPrefsStore(d)
		c.AdvancedStats = stats
		c.DeckLibrary = library
		c.Users = users.NewSQLStore(d, nil)
		c.Now = func() time.Time { return advancedNow }
	}, statsAdminDiscordID)
	s.lobby.SetGameStats(stats)
	return advancedStack{statsPrefsStack: statsPrefsStack{adminStack: s, db: d}, stats: stats, library: library, oracle: oracle}
}

func (s advancedStack) mintShare(t *testing.T, tok string) shareLinkResponse {
	t.Helper()
	code, out := decodeInto[shareLinkResponse](t, s.statsPrefsStack, "POST", "/me/stats/share-link", tok, nil)
	if code != http.StatusCreated {
		t.Fatalf("POST /me/stats/share-link: %d", code)
	}
	return out
}

// shared reads a shared view with no credential, as a Discord click does.
func (s advancedStack) shared(t *testing.T, token, query string) (int, string) {
	t.Helper()
	resp := do(t, s.srv, "GET", "/stats/shared/"+token+query, "", nil)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func TestShareLinkMintRotateRevoke(t *testing.T) {
	s := newShareStack(t)
	tok, _ := s.person(t, "222", "Bea")
	guest := s.issue(t, auth.Principal{Role: auth.RolePlayer, GameID: uuid.New(), PlayerID: uuid.New()})
	if code := status(t, s.srv, "POST", "/me/stats/share-link", "", nil); code != http.StatusUnauthorized {
		t.Errorf("signed out: %d, want 401", code)
	}
	if code := status(t, s.srv, "POST", "/me/stats/share-link", guest, nil); code != http.StatusForbidden {
		t.Errorf("a guest: %d, want 403", code)
	}

	first := s.mintShare(t, tok)
	if !shareTokenRe.MatchString(first.Token) || first.Path != "#/stats/shared/"+first.Token || first.CreatedAt != advancedNow.UnixMilli() {
		t.Fatalf("minted %+v", first)
	}
	// Only the hash is stored.
	var stored string
	if err := s.db.QueryRow(`SELECT share_link_hash FROM stats_prefs`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == first.Token || stored != shareLinkHash(first.Token) {
		t.Errorf("stored %q; want the token's SHA-256, never the token", stored)
	}
	if p := s.prefs(t, tok); p.ShareLinkAt != advancedNow.UnixMilli() {
		t.Errorf("GET /me/stats/prefs share_link_at = %d", p.ShareLinkAt)
	}
	if code, _ := s.shared(t, first.Token, ""); code != http.StatusOK {
		t.Fatalf("the new link: %d, want 200", code)
	}

	second := s.mintShare(t, tok)
	if second.Token == first.Token {
		t.Fatal("minting again gave the same token")
	}
	if code, _ := s.shared(t, first.Token, ""); code != http.StatusNotFound {
		t.Errorf("the rotated-out link: %d, want 404", code)
	}
	if code, _ := s.shared(t, second.Token, ""); code != http.StatusOK {
		t.Errorf("the rotated-in link: %d, want 200", code)
	}

	if code := status(t, s.srv, "DELETE", "/me/stats/share-link", tok, nil); code != http.StatusNoContent {
		t.Fatalf("DELETE: %d, want 204", code)
	}
	if code, _ := s.shared(t, second.Token, ""); code != http.StatusNotFound {
		t.Errorf("a revoked link: %d, want 404", code)
	}
	if p := s.prefs(t, tok); p.ShareLinkAt != 0 {
		t.Errorf("after revoking, share_link_at = %d", p.ShareLinkAt)
	}
	// Revoking twice is fine.
	if code := status(t, s.srv, "DELETE", "/me/stats/share-link", tok, nil); code != http.StatusNoContent {
		t.Errorf("DELETE again: %d, want 204", code)
	}
	for _, bad := range []string{"short", strings.Repeat("A", 22), strings.Repeat("A", 23)} {
		if code, _ := s.shared(t, bad, ""); code != http.StatusNotFound {
			t.Errorf("token %q: %d, want 404", bad, code)
		}
	}
}

// A minted link does not touch the advanced switch.
func TestShareLinkKeepsTheSwitch(t *testing.T) {
	s := newShareStack(t)
	tok, _ := s.person(t, "222", "Bea")
	s.on(t, tok)
	s.mintShare(t, tok)
	if p := s.prefs(t, tok); !p.Advanced || p.Choice != statsChoiceOn {
		t.Errorf("after minting: %+v; want the switch kept on", p)
	}
}

func TestSharedViewShowsTheOwnerAndNoOneElse(t *testing.T) {
	s := newShareStack(t)
	tok, user := s.person(t, "222", "Bea")
	// An opponent who is a person with a name of their own.
	_, other := s.person(t, "333", "Zed the Opponent")
	d := s.saveDeck(t, user, "Atraxa Superfriends", atraxaList, "Atraxa, Praetors' Voice")
	for i := 0; i < 4; i++ {
		s.seed(t, user, seededGame{daysAgo: i, deckID: d.ID.String(), hash: s.currentHash(), result: gamestats.ResultWin,
			others: []string{"person", "bot", "person"}})
	}
	if _, err := s.db.Exec(`UPDATE seats SET user_id = ? WHERE seat = 1`, other.String()); err != nil {
		t.Fatal(err)
	}
	link := s.mintShare(t, tok)

	// Advanced off: the basic record, no overview.
	code, body := s.shared(t, link.Token, "")
	if code != http.StatusOK {
		t.Fatalf("shared: %d %s", code, body)
	}
	var got sharedStatsResponse
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatal(err)
	}
	if got.DisplayName != "Bea" || got.Advanced || got.Overview != nil || got.Basic.Games != 4 {
		t.Errorf("advanced off: %+v", got)
	}

	s.on(t, tok)
	code, body = s.shared(t, link.Token, "?makeup=all&range=30d")
	if code != http.StatusOK {
		t.Fatalf("shared: %d %s", code, body)
	}
	got = sharedStatsResponse{}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatal(err)
	}
	if !got.Advanced || got.Overview == nil || got.Overview.Record.Games != 4 || got.Overview.Range != gamestats.Range30 {
		t.Fatalf("advanced on: %+v", got)
	}
	if len(got.Overview.Decks) == 0 || got.Overview.Decks[0].Name != "Atraxa" {
		t.Errorf("the owner's deck names are part of the view: %+v", got.Overview.Decks)
	}
	// Never: another person's name, a table name, a game id, a deck id.
	for _, never := range []string{"Zed the Opponent", "Friday", "game_id", "replay", d.ID.String(), other.String(), user.String()} {
		if strings.Contains(body, never) {
			t.Errorf("the shared view carries %q", never)
		}
	}
	// A bad filter is a 400, as on the owner's page.
	if code, _ := s.shared(t, link.Token, "?makeup=everyone"); code != http.StatusBadRequest {
		t.Errorf("bad makeup: %d, want 400", code)
	}

	// A deleted account's link stops working.
	if _, err := s.db.Exec(`UPDATE users SET deleted_at = 9 WHERE id = ?`, user.String()); err != nil {
		t.Fatal(err)
	}
	if code, _ := s.shared(t, link.Token, ""); code != http.StatusNotFound {
		t.Errorf("a deleted account's link: %d, want 404", code)
	}
}

func TestSharedViewWithNoStore(t *testing.T) {
	s := newAdminStackWith(t, nil, nil, statsAdminDiscordID)
	if code := status(t, s.srv, "GET", "/stats/shared/"+strings.Repeat("A", 22), "", nil); code != http.StatusNotFound {
		t.Errorf("no store: %d, want 404", code)
	}
}
