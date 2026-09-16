package lobby

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// Tests for the login-page half of Discord sign-in (ADR 0050): an
// OAuth round-trip that starts with no invite in hand, mints an
// identity-only session, and is traded for a seat by POST /join.
//
// The stub identity in newDiscordTestStack is Alice / discord-99 /
// avhash; these tests lean on that rather than restating it.

// followOneRedirect issues a GET that stops AT the 302 rather than
// following it, and returns the Location header.
func followOneRedirect(t *testing.T, srv *httptest.Server, path string) string {
	t.Helper()
	client := *srv.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Get(srv.URL + path)
	if err != nil {
		t.Fatalf("get %s: %v", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("status for %s: got %d, want 302", path, resp.StatusCode)
	}
	return resp.Header.Get("Location")
}

// fragmentParams parses the query string the SPA hand-off packs into
// the URL fragment ("/#/oauth-complete?token=…").
func fragmentParams(t *testing.T, loc string) url.Values {
	t.Helper()
	_, q, ok := strings.Cut(loc, "?")
	if !ok {
		t.Fatalf("no query in fragment: %q", loc)
	}
	v, err := url.ParseQuery(q)
	if err != nil {
		t.Fatalf("parse fragment %q: %v", loc, err)
	}
	return v
}

// identityTokenFromCallback walks the login-page half of the flow and
// returns the identity-only session token from the fragment.
func identityTokenFromCallback(t *testing.T, srv *httptest.Server, store *discord.StateStore) string {
	t.Helper()
	state, _, err := store.Start(uuid.Nil, "")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	frag := fragmentParams(t, followOneRedirect(t, srv, "/auth/discord/callback?state="+state+"&code=code-1"))
	tok := frag.Get("token")
	if tok == "" {
		t.Fatal("callback fragment carried no token")
	}
	return tok
}

func TestDiscordStartWithoutAnInviteIsUnbound(t *testing.T) {
	srv, _, _, _ := newDiscordTestStack(t)

	loc := followOneRedirect(t, srv, "/auth/discord/start")
	if !strings.Contains(loc, "discord.com/oauth2/authorize") {
		t.Errorf("Location: %q does not point at Discord authorize", loc)
	}
	// The unbound flow is not a weaker flow: it carries the same
	// PKCE challenge the invite-link one does.
	if !strings.Contains(loc, "code_challenge_method=S256") {
		t.Errorf("unbound start must still use PKCE S256: %q", loc)
	}
}

func TestDiscordStartRejectsHalfAnInvite(t *testing.T) {
	srv, l, _, _ := newDiscordTestStack(t)
	meta, _ := l.Create("FNM")

	// Either half alone is malformed. Guessing would mean silently
	// dropping a seat claim (treat as unbound) or inventing a table.
	for _, path := range []string{
		"/auth/discord/start?game=" + meta.ID.String(),
		"/auth/discord/start?t=" + meta.InviteToken,
	} {
		resp := doGet(t, srv, path, "")
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", path, resp.StatusCode)
		}
	}
}

func TestDiscordCallbackWithoutAnInviteMintsAnIdentitySession(t *testing.T) {
	srv, _, _, store := newDiscordTestStack(t)

	state, _, err := store.Start(uuid.Nil, "")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	loc := followOneRedirect(t, srv, "/auth/discord/callback?state="+state+"&code=code-1")
	if !strings.HasPrefix(loc, "/#/oauth-complete?") {
		t.Fatalf("Location should redirect to the SPA hand-off: %q", loc)
	}

	frag := fragmentParams(t, loc)
	if frag.Get("token") == "" {
		t.Error("fragment carried no session token")
	}
	if got := frag.Get("name"); got != "Alice" {
		t.Errorf("name: got %q, want Alice (shown while they type a code)", got)
	}
	// The ABSENCE of these two is how the client tells this flow from
	// a claimed seat, so it is the assertion that matters most here.
	if frag.Get("game") != "" || frag.Get("player_id") != "" {
		t.Errorf("identity fragment must carry no seat: %q", loc)
	}

	resp := doGet(t, srv, "/me", frag.Get("token"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/me status: got %d, want 200", resp.StatusCode)
	}
	var probe map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&probe); err != nil {
		t.Fatalf("decode /me: %v", err)
	}
	role, _ := probe["role"].(string)
	if role == "" {
		if p, ok := probe["principal"].(map[string]any); ok {
			role, _ = p["role"].(string)
		}
	}
	if role != string(auth.RoleIdentified) {
		t.Errorf("role: got %q, want %q", role, auth.RoleIdentified)
	}
}

func TestJoinByCodeWithIdentitySessionSeatsTheDiscordUser(t *testing.T) {
	srv, l, _, store := newDiscordTestStack(t)
	meta, _ := l.Create("FNM")
	tok := identityTokenFromCallback(t, srv, store)

	resp := postJSON(t, srv, "/join", tok, map[string]string{"invite_token": meta.InviteToken})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}

	var body struct {
		Principal struct {
			Role      string `json:"role"`
			GameID    string `json:"game_id"`
			Name      string `json:"name"`
			DiscordID string `json:"discord_id"`
		} `json:"principal"`
		PlayerID string `json:"player_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Principal.Role != string(auth.RolePlayer) {
		t.Errorf("role: got %q, want player — the identity session should be traded in", body.Principal.Role)
	}
	if body.Principal.GameID != meta.ID.String() {
		t.Errorf("game_id: got %q, want %s", body.Principal.GameID, meta.ID)
	}
	if body.Principal.DiscordID != "discord-99" {
		t.Errorf("discord_id: got %q — identity did not survive the swap", body.Principal.DiscordID)
	}
	if body.Principal.Name != "Alice" {
		t.Errorf("name: got %q, want Alice", body.Principal.Name)
	}
	if body.PlayerID == "" {
		t.Error("player_id missing from join response")
	}

	// Opponents see the seat through GameMeta, so assert there too.
	updated, err := l.Get(meta.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(updated.Players) != 1 {
		t.Fatalf("seat count: got %d, want 1", len(updated.Players))
	}
	seat := updated.Players[0]
	if seat.DisplayName != "Alice" || seat.DiscordAvatarHash != "avhash" {
		t.Errorf("seat: got display=%q avatar=%q, want Alice/avhash", seat.DisplayName, seat.DiscordAvatarHash)
	}
}

func TestJoinByCodeIgnoresABodyNameWhenSignedIn(t *testing.T) {
	srv, l, _, store := newDiscordTestStack(t)
	meta, _ := l.Create("FNM")
	tok := identityTokenFromCallback(t, srv, store)

	// A verified identity must not be relabelable by the request
	// body — otherwise the avatar says Alice and the name says
	// anything the caller liked.
	resp := postJSON(t, srv, "/join", tok, map[string]string{
		"invite_token": meta.InviteToken,
		"name":         "Mallory",
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}

	updated, err := l.Get(meta.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got := updated.Players[0].DisplayName; got != "Alice" {
		t.Errorf("display name: got %q, want Alice", got)
	}
}

func TestJoinByCodeAnonymouslyStillNeedsAName(t *testing.T) {
	srv, l, _, _ := newDiscordTestStack(t)
	meta, _ := l.Create("FNM")

	resp := postJSON(t, srv, "/join", "", map[string]string{"invite_token": meta.InviteToken})
	resp.Body.Close()
	// ErrEmptyName maps to 400, not 409 — see the error switch in
	// http.go. The docs table said otherwise until this change.
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("nameless anonymous join: got %d, want 400", resp.StatusCode)
	}

	// …and with a name it behaves exactly like the classic join, so a
	// deploy with Discord unconfigured keeps working.
	resp2 := postJSON(t, srv, "/join", "", map[string]string{
		"invite_token": meta.InviteToken,
		"name":         "Bob",
	})
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("named anonymous join: got %d, want 200", resp2.StatusCode)
	}
}

func TestJoinByCodeRejectsAnUnknownCode(t *testing.T) {
	srv, _, _, _ := newDiscordTestStack(t)

	resp := postJSON(t, srv, "/join", "", map[string]string{
		"invite_token": "no-such-code",
		"name":         "Bob",
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", resp.StatusCode)
	}
}

func TestJoinByCodeRefusesAnArchivedTable(t *testing.T) {
	srv, l, _, _ := newDiscordTestStack(t)
	meta, _ := l.Create("FNM")
	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatalf("SetArchived: %v", err)
	}

	// A stale code in an old chat message should read as expired
	// rather than quietly reopening a retired table.
	resp := postJSON(t, srv, "/join", "", map[string]string{
		"invite_token": meta.InviteToken,
		"name":         "Bob",
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", resp.StatusCode)
	}
}

func TestJoinByCodeRefusesASessionThatAlreadyHasASeat(t *testing.T) {
	srv, l, _, _ := newDiscordTestStack(t)
	first, _ := l.Create("FNM")
	second, _ := l.Create("Second table")

	resp := postJSON(t, srv, "/join", "", map[string]string{
		"invite_token": first.InviteToken,
		"name":         "Bob",
	})
	var seated struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&seated); err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp.Body.Close()
	if seated.Token == "" {
		t.Fatal("first join returned no session")
	}

	// Reusing a seated session to grab another seat is a client bug;
	// minting a second seat silently would be worse than an error.
	resp2 := postJSON(t, srv, "/join", seated.Token, map[string]string{
		"invite_token": second.InviteToken,
		"name":         "Bob",
	})
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("status: got %d, want 409", resp2.StatusCode)
	}
}

func TestFindByInviteResolvesOnlyLivePlayerInvites(t *testing.T) {
	_, l, _, _ := newDiscordTestStack(t)
	meta, _ := l.Create("FNM")

	got, err := l.FindByInvite(meta.InviteToken)
	if err != nil {
		t.Fatalf("FindByInvite: %v", err)
	}
	if got != meta.ID {
		t.Errorf("game: got %s, want %s", got, meta.ID)
	}

	if _, err := l.FindByInvite(""); err == nil {
		t.Error("an empty code must not resolve")
	}
	if _, err := l.FindByInvite("nope"); err == nil {
		t.Error("an unknown code must not resolve")
	}
	// The spectator invite is a different credential: letting it
	// resolve here would let a read-only link claim a seat.
	if meta.SpectatorInvite != "" {
		if _, err := l.FindByInvite(meta.SpectatorInvite); err == nil {
			t.Error("a spectator invite must not resolve to a joinable table")
		}
	}
}

func TestWSAuthorizerRefusesAnIdentitySession(t *testing.T) {
	a := auth.NewMemoryAuthenticator()
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role:      auth.RoleIdentified,
		DiscordID: "discord-99",
	}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	// An identity principal has no GameID; reaching the hub it would
	// bind to the zero game. It must be turned away at the door.
	az := &WSAuthorizer{Auth: a}
	req := httptest.NewRequest(http.MethodGet, "/ws?token="+tok, nil)
	if _, err := az.AuthorizeUpgrade(req); err == nil {
		t.Fatal("an identity session must not be able to open a socket")
	}
}
