package lobby

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// Tests for S34 sub-PR 2 at the HTTP layer (ADR 0051 decisions 2, 3
// and 5): a Discord sign-in writes a users + identities row, the
// session it mints carries the user's id, a seat session minted from
// it carries the same id, and admin / guest sessions carry none.

const signinTestKey = "lobby-test-identity-key-0123456789abcdef"

// userStack is the Discord test stack over a real database: an SQL
// lobby store, an SQL user store (sealing refresh tokens under
// sealer, or discarding them when sealer is nil) and HMAC sessions, so
// UserID has to survive a token round trip rather than sit in a map.
type userStack struct {
	srv   *httptest.Server
	lobby *Lobby
	stub  *discordStub
	state *discord.StateStore
	users *users.SQLStore
	auth  auth.Authenticator
	db    *db.DB
}

func newUserStack(t *testing.T, sealer *users.Sealer) userStack {
	t.Helper()
	return newUserStackWith(t, sealer, nil)
}

// newUserStackWith is newUserStack with a hook that can adjust the
// Config after the database-backed pieces are wired — what S34 sub-PR
// 6's DM-invite tests need to add a bot token and an invite origin.
func newUserStackWith(t *testing.T, sealer *users.Sealer, configure func(*Config)) userStack {
	t.Helper()
	dir := t.TempDir()
	d := openTestDB(t, dir)
	us := users.NewSQLStore(d, sealer)
	a, err := auth.NewHMACAuthenticator([]byte("lobby-test-session-key-0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewHMACAuthenticator: %v", err)
	}
	srv, l, stub, state := newDiscordTestStackWith(t, func(c *Config) {
		c.Lobby = NewLobbyWithStore(ws.NewRoomManager(quietLogger(), ""), NewSQLStore(d))
		c.Auth = a
		c.Users = us
		if configure != nil {
			configure(c)
		}
	})
	stub.tokenBody = `{"access_token":"tok-stub","token_type":"Bearer","expires_in":3600,"refresh_token":"rt-stub-secret","scope":"identify"}`
	return userStack{srv: srv, lobby: l, stub: stub, state: state, users: us, auth: a, db: d}
}

func mustValidate(t *testing.T, a auth.Authenticator, tok string) auth.Principal {
	t.Helper()
	p, err := a.Validate(context.Background(), tok)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	return p
}

func TestDiscordSignInRecordsTheUserAndStampsTheSession(t *testing.T) {
	s := newUserStack(t, mustSealer(t))

	tok := identityTokenFromCallback(t, s.srv, s.state)
	p := mustValidate(t, s.auth, tok)
	if p.Role != auth.RoleIdentified {
		t.Fatalf("role %q", p.Role)
	}
	if p.UserID == uuid.Nil {
		t.Fatal("identity session carries no UserID")
	}
	// Decision 3 keeps the cached Discord fields through S34.
	if p.DiscordID != "discord-99" || p.DiscordGlobalName != "Alice" || p.DiscordAvatarHash != "avhash" {
		t.Errorf("cached Discord fields lost: %+v", p)
	}

	u, err := s.users.Get(context.Background(), p.UserID)
	if err != nil {
		t.Fatalf("users.Get: %v", err)
	}
	if u.DisplayName != "Alice" || u.AvatarURL != "/avatars/discord-99/avhash.png" {
		t.Errorf("user row = %+v", u)
	}

	// The refresh token Discord returned is stored, sealed.
	var raw []byte
	if err := s.db.QueryRow(`SELECT refresh_token FROM identities WHERE subject = 'discord-99'`).Scan(&raw); err != nil {
		t.Fatalf("read refresh_token: %v", err)
	}
	if raw == nil || bytes.Contains(raw, []byte("rt-stub-secret")) {
		t.Errorf("refresh_token column = %q, want a sealed, non-plaintext value", raw)
	}
	if rt, ok, err := s.users.RefreshToken(context.Background(), "discord-99"); err != nil || !ok || rt != "rt-stub-secret" {
		t.Errorf("RefreshToken = %q %v %v", rt, ok, err)
	}

	// Signing in again is the same person, not a second row.
	again := mustValidate(t, s.auth, identityTokenFromCallback(t, s.srv, s.state))
	if again.UserID != p.UserID {
		t.Errorf("second sign-in: UserID %s, want %s", again.UserID, p.UserID)
	}
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	if n != 1 {
		t.Errorf("users rows = %d, want 1", n)
	}
}

func TestDiscordInviteFlowSeatSessionCarriesTheUser(t *testing.T) {
	s := newUserStack(t, mustSealer(t))
	meta, err := s.lobby.Create("FNM")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	state, _, err := s.state.Start(meta.ID, meta.InviteToken)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	frag := fragmentParams(t, followOneRedirect(t, s.srv, "/auth/discord/callback?state="+state+"&code=code-1"))
	p := mustValidate(t, s.auth, frag.Get("token"))
	if p.Role != auth.RolePlayer || p.UserID == uuid.Nil {
		t.Fatalf("seat session = role %q user %s, want a player with a UserID", p.Role, p.UserID)
	}
	if _, err := s.users.Get(context.Background(), p.UserID); err != nil {
		t.Errorf("seat session's UserID has no users row: %v", err)
	}
}

func TestJoinByCodeCarriesTheIdentitySessionsUser(t *testing.T) {
	s := newUserStack(t, mustSealer(t))
	meta, err := s.lobby.Create("FNM")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	idTok := identityTokenFromCallback(t, s.srv, s.state)
	identity := mustValidate(t, s.auth, idTok)

	resp := postJSON(t, s.srv, "/join", idTok, map[string]string{"invite_token": meta.InviteToken})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /join: %d", resp.StatusCode)
	}
	var body struct {
		Token     string `json:"token"`
		Principal struct {
			UserID string `json:"user_id"`
		} `json:"principal"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Principal.UserID != identity.UserID.String() {
		t.Errorf("response principal user_id = %q, want %s", body.Principal.UserID, identity.UserID)
	}
	seat := mustValidate(t, s.auth, body.Token)
	if seat.Role != auth.RolePlayer || seat.UserID != identity.UserID {
		t.Errorf("seat session = role %q user %s, want player / %s", seat.Role, seat.UserID, identity.UserID)
	}
}

func TestGuestAndAdminSessionsCarryNoUser(t *testing.T) {
	s := newUserStack(t, mustSealer(t))
	meta, err := s.lobby.Create("FNM")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	resp := postJSON(t, s.srv, "/join", "", map[string]string{"invite_token": meta.InviteToken, "name": "Guest"})
	var guest struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&guest)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("guest join: %d", resp.StatusCode)
	}
	if p := mustValidate(t, s.auth, guest.Token); p.UserID != uuid.Nil {
		t.Errorf("guest seat session carries UserID %s", p.UserID)
	}

	resp = postJSON(t, s.srv, "/admin/login", "", map[string]string{"token": "shared-admin-token"})
	var admin struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&admin)
	resp.Body.Close()
	if p := mustValidate(t, s.auth, admin.Token); p.Role != auth.RoleAdmin || p.UserID != uuid.Nil {
		t.Errorf("admin session = role %q user %s, want admin with no UserID", p.Role, p.UserID)
	}

	// An admin-created game records no creator.
	resp = postJSON(t, s.srv, "/games", admin.Token, map[string]string{"name": "Admin table"})
	var created GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /games: %d", resp.StatusCode)
	}
	var createdBy *string
	if err := s.db.QueryRow(`SELECT created_by FROM games WHERE id = ?`, created.ID.String()).Scan(&createdBy); err != nil {
		t.Fatalf("read created_by: %v", err)
	}
	if createdBy != nil {
		t.Errorf("admin-created game has created_by %q, want NULL", *createdBy)
	}
}

func TestCreateByRecordsTheCreator(t *testing.T) {
	s := newUserStack(t, nil)
	p := mustValidate(t, s.auth, identityTokenFromCallback(t, s.srv, s.state))

	meta, err := s.lobby.CreateBy("Alice's table", p.UserID)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	var gameCreator, inviteCreator string
	if err := s.db.QueryRow(`SELECT created_by FROM games WHERE id = ?`, meta.ID.String()).Scan(&gameCreator); err != nil {
		t.Fatalf("read games.created_by: %v", err)
	}
	if gameCreator != p.UserID.String() {
		t.Errorf("games.created_by = %q, want %s", gameCreator, p.UserID)
	}
	if err := s.db.QueryRow(`SELECT DISTINCT created_by FROM invites WHERE game_id = ?`, meta.ID.String()).Scan(&inviteCreator); err != nil {
		t.Fatalf("read invites.created_by: %v", err)
	}
	if inviteCreator != p.UserID.String() {
		t.Errorf("invites.created_by = %q, want %s", inviteCreator, p.UserID)
	}

	// The column is a foreign key now: a creator with no users row is
	// refused, and the half-made game is not left in the lobby.
	before := len(s.lobby.List())
	if _, err := s.lobby.CreateBy("ghost", uuid.New()); err == nil {
		t.Error("CreateBy accepted a user that does not exist")
	}
	if after := len(s.lobby.List()); after != before {
		t.Errorf("lobby has %d games after a refused create, want %d", after, before)
	}
}

func TestSignInWithoutIdentityKeyStoresNoRefreshToken(t *testing.T) {
	s := newUserStack(t, nil)
	p := mustValidate(t, s.auth, identityTokenFromCallback(t, s.srv, s.state))
	if p.UserID == uuid.Nil {
		t.Fatal("sign-in without an identity key must still record the user")
	}
	var raw []byte
	if err := s.db.QueryRow(`SELECT refresh_token FROM identities WHERE subject = 'discord-99'`).Scan(&raw); err != nil {
		t.Fatalf("read refresh_token: %v", err)
	}
	if raw != nil {
		t.Errorf("refresh_token = %q with no key, want NULL", raw)
	}
}

// TestDiscordSignInWithoutADatabase is the CMDCTRL_DATA_DIR="" path:
// no user store is wired, sign-in works, and the session carries a
// zero UserID — exactly the pre-S34 behaviour.
func TestDiscordSignInWithoutADatabase(t *testing.T) {
	srv, l, _, state := newDiscordTestStack(t)

	tok := identityTokenFromCallback(t, srv, state)
	resp := doGet(t, srv, "/me", tok)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/me: %d", resp.StatusCode)
	}
	var me map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&me)
	principal := me
	if inner, ok := me["principal"].(map[string]any); ok {
		principal = inner
	}
	if uid, ok := principal["user_id"]; ok && uid != "" && uid != uuid.Nil.String() {
		t.Errorf("user_id = %v with no database, want absent", uid)
	}
	if principal["discord_id"] != "discord-99" {
		t.Errorf("discord_id = %v, sign-in did not complete", principal["discord_id"])
	}

	// And the seat still claims through the invite flow.
	meta, _ := l.Create("FNM")
	st, _, err := state.Start(meta.ID, meta.InviteToken)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	frag := fragmentParams(t, followOneRedirect(t, srv, "/auth/discord/callback?state="+st+"&code=code-2"))
	if frag.Get("player_id") == "" {
		t.Error("invite-flow sign-in without a database claimed no seat")
	}
}

func mustSealer(t *testing.T) *users.Sealer {
	t.Helper()
	s, err := users.NewSealer(signinTestKey)
	if err != nil {
		t.Fatalf("NewSealer: %v", err)
	}
	return s
}
