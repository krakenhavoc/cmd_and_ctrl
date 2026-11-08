package lobby

// Tests for ADR 0163 in the lobby: the Linked Roles verification URL,
// the callback's role-connection branch over a real database and the
// real linkedroles service, with Discord's token, user and
// role-connection endpoints all fakes; /community's flag; and the push
// the end of a game asks for.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/linkedroles"
)

const linkedRoleApp = "777777777777777777"

// roleConnectionAPI is Discord's role-connection PUT: what was written,
// with which token.
type roleConnectionAPI struct {
	mu     sync.Mutex
	auths  []string
	bodies []map[string]any
}

func newRoleConnectionAPI(t *testing.T) *roleConnectionAPI {
	t.Helper()
	api := &roleConnectionAPI{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/users/@me/applications/"+linkedRoleApp+"/role-connection" {
			http.NotFound(w, r)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		api.mu.Lock()
		api.auths = append(api.auths, r.Header.Get("Authorization"))
		api.bodies = append(api.bodies, body)
		api.mu.Unlock()
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(discord.SwapBotAPIBaseForTesting(srv.URL))
	return api
}

func (a *roleConnectionAPI) snapshot() ([]string, []map[string]any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string(nil), a.auths...), append([]map[string]any(nil), a.bodies...)
}

// linkedRoleStack is the sign-up stack (real database, sealed tokens)
// with Linked Roles on when on is true.
func linkedRoleStack(t *testing.T, on bool) (signupStack, *roleConnectionAPI) {
	t.Helper()
	api := newRoleConnectionAPI(t)
	// The service needs the stack's database, which exists only once the
	// stack does; the Config holds this pointer and it is filled after.
	late := &lateLinkedRoles{}
	var oauth discord.Config
	var client *http.Client
	s := newSignupStack(t, false, func(c *Config) {
		if on {
			c.LinkedRoles = late
			oauth, client = c.Discord, c.DiscordHTTPClient
		}
	})
	if on {
		lr := discord.LinkedRoles{AppID: linkedRoleApp, Token: "bot-token", Requested: true}
		late.svc = linkedroles.New(linkedroles.NewDiscord(oauth, lr, client), s.users, linkedroles.NewSQLStats(s.db), nil)
	}
	// The role-connection grant: its own scopes and refresh token.
	s.stub.mu.Lock()
	s.stub.tokenBody = `{"access_token":"at-rc","token_type":"Bearer","expires_in":604800,"refresh_token":"rt-rc","scope":"identify role_connections.write"}`
	s.stub.mu.Unlock()
	return s, api
}

// lateLinkedRoles delegates to a service set after the stack is built.
type lateLinkedRoles struct{ svc *linkedroles.Service }

func (l *lateLinkedRoles) Link(ctx context.Context, userID uuid.UUID, subject string, tok discord.TokenResponse) error {
	return l.svc.Link(ctx, userID, subject, tok)
}

func (l *lateLinkedRoles) Forget(subject string, sealed []byte) { l.svc.Forget(subject, sealed) }

func TestLinkedRoleRouteOffSaysUnavailable(t *testing.T) {
	s, _ := linkedRoleStack(t, false)
	resp := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/linked-role", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || resp.Header.Get("Location") != linkedRoleUnavailable {
		t.Errorf("off: %d %q, want 302 to %s", resp.StatusCode, resp.Header.Get("Location"), linkedRoleUnavailable)
	}
	if body := communityBody(t, s); body["linked_roles"] != false {
		t.Errorf("off: /community linked_roles = %v", body["linked_roles"])
	}
}

func communityBody(t *testing.T, s signupStack) map[string]any {
	t.Helper()
	resp := doGet(t, s.srv, "/community", "")
	defer resp.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return body
}

// startLinkedRole follows the verification URL and returns the state it
// parked.
func startLinkedRole(t *testing.T, s signupStack) string {
	t.Helper()
	resp := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/linked-role", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("linked-role: %d", resp.StatusCode)
	}
	loc, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	q := loc.Query()
	if q.Get("scope") != "identify role_connections.write" || q.Get("prompt") != "consent" ||
		q.Get("redirect_uri") != "http://localhost/cb" || q.Get("code_challenge_method") != "S256" {
		t.Fatalf("authorize URL = %s", loc)
	}
	return q.Get("state")
}

func TestLinkedRoleLinksAnAccountWithoutASession(t *testing.T) {
	s, api := linkedRoleStack(t, true)
	if body := communityBody(t, s); body["linked_roles"] != true {
		t.Errorf("on: /community linked_roles = %v", body["linked_roles"])
	}
	ctx := context.Background()
	if _, err := s.users.UpsertFromDiscord(ctx, discord.User{ID: "discord-99", Username: "alice", GlobalName: "Alice"}, "rt-signin", "identify"); err != nil {
		t.Fatal(err)
	}

	state := startLinkedRole(t, s)
	resp := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/callback?state="+state+"&code=c1", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || resp.Header.Get("Location") != linkedRoleLinked {
		t.Fatalf("callback: %d %q, want 302 to %s", resp.StatusCode, resp.Header.Get("Location"), linkedRoleLinked)
	}
	// Linking is not signing in.
	for _, c := range resp.Cookies() {
		if c.Name == auth.SessionCookie {
			t.Errorf("the callback set a session cookie")
		}
	}
	if rt, ok, err := s.users.RoleConnectionToken(ctx, "discord-99"); err != nil || !ok || rt != "rt-rc" {
		t.Errorf("stored grant = %q %v %v, want rt-rc", rt, ok, err)
	}
	if rt, _, _ := s.users.RefreshToken(ctx, "discord-99"); rt != "rt-signin" {
		t.Errorf("the sign-in grant changed: %q", rt)
	}
	auths, bodies := api.snapshot()
	if len(auths) != 1 || auths[0] != "Bearer at-rc" {
		t.Fatalf("role-connection PUTs = %v, want one with the grant's access token", auths)
	}
	m, _ := bodies[0]["metadata"].(map[string]any)
	if bodies[0]["platform_username"] != "Alice" || m["games_played"] != "0" || m["games_won"] != "0" || m["joined"] == "" {
		t.Errorf("pushed %v", bodies[0])
	}
}

func TestLinkedRoleWithNoAccountStoresNothing(t *testing.T) {
	s, api := linkedRoleStack(t, true)
	state := startLinkedRole(t, s)
	resp := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/callback?state="+state+"&code=c1", "")
	resp.Body.Close()
	if resp.Header.Get("Location") != linkedRoleSignIn {
		t.Fatalf("no account: %q, want %s", resp.Header.Get("Location"), linkedRoleSignIn)
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil || n != 0 {
		t.Errorf("%d users rows (%v) after linking with no account, want 0", n, err)
	}
	if auths, _ := api.snapshot(); len(auths) != 0 {
		t.Errorf("%d role-connection PUTs with no account", len(auths))
	}
}

func TestLinkedRoleRefusedOnDiscordGoesToTheFailedPage(t *testing.T) {
	s, _ := linkedRoleStack(t, true)
	state := startLinkedRole(t, s)
	resp := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/callback?state="+state+"&error=access_denied", "")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || resp.Header.Get("Location") != linkedRoleFailed {
		t.Errorf("refused: %d %q %s, want 302 to %s", resp.StatusCode, resp.Header.Get("Location"), body, linkedRoleFailed)
	}
	// The state is spent: a replay is the plain 400.
	again := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/callback?state="+state+"&code=c1", "")
	again.Body.Close()
	if again.StatusCode != http.StatusBadRequest {
		t.Errorf("replayed state: %d, want 400", again.StatusCode)
	}
}

func TestLinkedRoleWithoutTheScopeIsNotKept(t *testing.T) {
	s, api := linkedRoleStack(t, true)
	if _, err := s.users.UpsertFromDiscord(context.Background(), discord.User{ID: "discord-99", Username: "alice"}, "", "identify"); err != nil {
		t.Fatal(err)
	}
	s.stub.mu.Lock()
	s.stub.tokenBody = `{"access_token":"at-rc","token_type":"Bearer","refresh_token":"rt-rc","scope":"identify"}`
	s.stub.mu.Unlock()
	state := startLinkedRole(t, s)
	resp := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/callback?state="+state+"&code=c1", "")
	resp.Body.Close()
	if resp.Header.Get("Location") != linkedRoleFailed {
		t.Errorf("no role_connections.write: %q, want %s", resp.Header.Get("Location"), linkedRoleFailed)
	}
	if _, ok, _ := s.users.RoleConnectionToken(context.Background(), "discord-99"); ok {
		t.Error("a grant without the scope was kept")
	}
	if auths, _ := api.snapshot(); len(auths) != 0 {
		t.Errorf("%d PUTs", len(auths))
	}
	if strings.Contains(s.logs.String(), "rt-rc") || strings.Contains(s.logs.String(), "at-rc") {
		t.Error("a token is in the log")
	}
}

// recordingQueue is the push queue: who was asked for.
type recordingQueue struct {
	mu  sync.Mutex
	ids []uuid.UUID
}

func (q *recordingQueue) Enqueue(ids ...uuid.UUID) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.ids = append(q.ids, ids...)
}

func (q *recordingQueue) snapshot() []uuid.UUID {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]uuid.UUID(nil), q.ids...)
}

func TestAGameThatEndsAsksForAPushForItsAccounts(t *testing.T) {
	l := newTestLobby(t)
	q := &recordingQueue{}
	l.SetLinkedRoles(q)
	meta, err := l.Create("Linked night")
	if err != nil {
		t.Fatal(err)
	}
	aliceUser, bobUser := uuid.New(), uuid.New()
	_, alice, err := l.JoinAs(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "1", Username: "alice"}, aliceUser)
	if err != nil {
		t.Fatal(err)
	}
	_, bob, err := l.JoinAs(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "2", Username: "bob"}, bobUser)
	if err != nil {
		t.Fatal(err)
	}
	_, guest, err := l.Join(meta.ID, meta.InviteToken, "Guest")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []uuid.UUID{alice, bob, guest} {
		if _, err := l.SetDeck(meta.ID, id, "Deck", botDeck(20)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	if got := q.snapshot(); len(got) != 0 {
		t.Fatalf("a push before the game ended: %v", got)
	}
	concedeSeat(t, l, meta.ID, alice)
	concedeSeat(t, l, meta.ID, guest)
	awaitEnded(t, l, meta.ID)
	got := q.snapshot()
	if len(got) != 2 || !((got[0] == aliceUser && got[1] == bobUser) || (got[0] == bobUser && got[1] == aliceUser)) {
		t.Errorf("pushes asked for %v, want Alice's and Bob's accounts once each (the guest has none)", got)
	}
	// Reading the ended game again asks for nothing more.
	awaitEnded(t, l, meta.ID)
	if n := len(q.snapshot()); n != 2 {
		t.Errorf("%d pushes after reading the ended game again, want 2", n)
	}
}

func TestAnAbandonedOrPracticeGameAsksForNoPush(t *testing.T) {
	l := newTestLobby(t)
	q := &recordingQueue{}
	l.SetLinkedRoles(q)
	user := uuid.New()
	l.mu.Lock()
	for _, e := range []*gameEntry{
		{outcome: "", meta: GameMeta{Players: []SeatInfo{{UserID: user.String()}}}},
		{outcome: "win", practice: true, meta: GameMeta{Players: []SeatInfo{{UserID: user.String()}}}},
		{outcome: "win", meta: GameMeta{Players: []SeatInfo{{UserID: user.String(), IsBot: true}, {UserID: user.String(), IsAgent: true}}}},
	} {
		l.linkedRolesEndedLocked(e)
	}
	l.mu.Unlock()
	if got := q.snapshot(); len(got) != 0 {
		t.Errorf("pushes asked for %v, want none", got)
	}
}

// recordingLinkedRoles records what account deletion asks of it.
type recordingLinkedRoles struct {
	mu     sync.Mutex
	forgot []string
	sealed [][]byte
}

func (r *recordingLinkedRoles) Link(context.Context, uuid.UUID, string, discord.TokenResponse) error {
	return nil
}

func (r *recordingLinkedRoles) Forget(subject string, sealed []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.forgot = append(r.forgot, subject)
	r.sealed = append(r.sealed, sealed)
}

func TestDeletingALinkedAccountEmptiesItsDiscordConnection(t *testing.T) {
	rec := &recordingLinkedRoles{}
	s := newDeletionStackWith(t, func(c *Config) { c.LinkedRoles = rec })
	tok, alice := s.signIn(t, "discord-99", "Alice", "avhash")
	// Another account, linked too, is left alone.
	_, bea := s.signIn(t, "discord-77", "Bea", "beahash")
	if _, err := s.db.Exec(`UPDATE identities SET role_connection_token = x'0c' WHERE user_id = ?`, bea.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`UPDATE identities SET role_connection_token = x'0a0b' WHERE user_id = ?`, alice.String()); err != nil {
		t.Fatal(err)
	}
	resp := s.deleteAccount(t, tok, DeleteAccountPhrase)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete = %d, want 204", resp.StatusCode)
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.forgot) != 1 || rec.forgot[0] != "discord-99" || string(rec.sealed[0]) != "\x0a\x0b" {
		t.Errorf("Forget calls = %v %x, want one for discord-99 with the sealed grant", rec.forgot, rec.sealed)
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM identities WHERE role_connection_token IS NOT NULL`).Scan(&n); err != nil || n != 1 {
		t.Errorf("%d grants left (%v), want only Bea's", n, err)
	}
}

func TestDeletingAnUnlinkedAccountCallsNoDiscord(t *testing.T) {
	rec := &recordingLinkedRoles{}
	s := newDeletionStackWith(t, func(c *Config) { c.LinkedRoles = rec })
	tok, _ := s.signIn(t, "discord-99", "Alice", "avhash")
	resp := s.deleteAccount(t, tok, DeleteAccountPhrase)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete = %d, want 204", resp.StatusCode)
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.forgot) != 0 {
		t.Errorf("an unlinked account's deletion asked Discord: %v", rec.forgot)
	}
}
