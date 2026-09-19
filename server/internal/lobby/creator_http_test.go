package lobby

// creator_http_test.go covers issue #1098, closing out the two
// TODO(#1044) markers #1057 (rotate) and #1058 (/cc-end) left behind:
//
//   - the creator half of POST /games/{id}/invites/rotate's "host or
//     admin" rule — CanRotateInvites, creator.go — at the HTTP layer.
//   - GET /games/{id}/creator, the Discord bot's /cc-end host check,
//     which needs a real users/identities row so it exercises
//     against a database rather than the in-memory store the rest of
//     this package tests against.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// identifiedToken mints a RoleIdentified session carrying userID and
// nothing table-specific — the credential a signed-in creator holds
// before (or without ever) claiming a seat, per ADR 0051 decision 3.
func identifiedToken(t *testing.T, a auth.Authenticator, userID uuid.UUID) string {
	t.Helper()
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role:   auth.RoleIdentified,
		UserID: userID,
	}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	return tok
}

// --- POST /games/{id}/invites/rotate — the creator half -------------

func TestRotateInviteCreatorMayRotate(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	creator := uuid.New()
	meta, err := l.CreateBy("Creator's table", creator)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	tok := identifiedToken(t, a, creator)

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/invites/rotate", tok,
		rotateInviteRequest{Kind: "player"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("creator rotate: got %d, want 200", resp.StatusCode)
	}
	var rotated rotateInviteResponse
	if err := json.NewDecoder(resp.Body).Decode(&rotated); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if rotated.Token == "" || rotated.Token == meta.InviteToken {
		t.Fatalf("rotate returned an unusable token: %q (old %q)", rotated.Token, meta.InviteToken)
	}
}

func TestRotateInviteDifferentSignedInUserForbidden(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	creator := uuid.New()
	meta, err := l.CreateBy("Creator's table", creator)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	// A DIFFERENT signed-in user — not this game's creator, not admin.
	other := identifiedToken(t, a, uuid.New())

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/invites/rotate", other,
		rotateInviteRequest{Kind: "player"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("a different signed-in user's rotate: got %d, want 403", resp.StatusCode)
	}
}

func TestRotateInviteAdminStillWorksOnCreatorGame(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	creator := uuid.New()
	meta, err := l.CreateBy("Creator's table", creator)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var admin sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&admin)
	resp.Body.Close()

	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/invites/rotate", admin.Token,
		rotateInviteRequest{Kind: "player"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("admin rotate on a creator-owned game: got %d, want 200", resp.StatusCode)
	}
}

func TestRotateInviteGuestSeatedAtCreatorGameRefused(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	creator := uuid.New()
	meta, err := l.CreateBy("Creator's table", creator)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	// A guest who joins by name — a zero-UserID player session, even
	// though they are seated at the creator's own table.
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Guest"})
	var guestSess sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&guestSess)
	resp.Body.Close()

	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/invites/rotate", guestSess.Token,
		rotateInviteRequest{Kind: "player"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("a guest seated at the creator's own table: got %d, want 403", resp.StatusCode)
	}
}

// TestRotateInviteRequiresAdmin (http_test.go) already proves a
// seated-but-unrelated player is refused on an admin-created (no
// creator) game; the four cases above are additive, not a
// replacement.

// --- GameMeta.is_creator on GET /games — what the lobby UI's rotate
// buttons actually read (#1098) --------------------------------------

func TestListGamesIsCreatorPerViewer(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	creator := uuid.New()
	meta, err := l.CreateBy("Creator's table", creator)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}

	mine := identifiedToken(t, a, creator)
	resp := doGet(t, srv, "/games", mine)
	defer resp.Body.Close()
	var list listResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, g := range list.Games {
		if g.ID == meta.ID {
			found = true
			if !g.IsCreator {
				t.Errorf("creator's own view of their table: is_creator = false, want true")
			}
		}
	}
	if !found {
		t.Fatalf("game %s missing from GET /games", meta.ID)
	}

	other := identifiedToken(t, a, uuid.New())
	resp2 := doGet(t, srv, "/games", other)
	defer resp2.Body.Close()
	var list2 listResponse
	if err := json.NewDecoder(resp2.Body).Decode(&list2); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, g := range list2.Games {
		if g.ID == meta.ID && g.IsCreator {
			t.Error("a different signed-in user's view of someone else's table: is_creator = true, want false")
		}
	}
}

// --- GET /games/{id}/creator — the bot's /cc-end host check ---------

func adminOnlyToken(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	defer resp.Body.Close()
	var admin sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&admin); err != nil {
		t.Fatalf("decode admin login: %v", err)
	}
	return admin.Token
}

func TestGameCreatorEndpoint_Match(t *testing.T) {
	s := newUserStack(t, nil)
	u, err := s.users.UpsertFromDiscord(context.Background(), discord.User{ID: "d-100", Username: "alice"}, "", "identify")
	if err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	meta, err := s.lobby.CreateBy("Creator's table", u.ID)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	admin := adminOnlyToken(t, s.srv)

	resp := doGet(t, s.srv, "/games/"+meta.ID.String()+"/creator?discord_id=d-100", admin)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	var body gameCreatorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.IsCreator {
		t.Error("want is_creator: true for the actual creator's Discord id")
	}
}

func TestGameCreatorEndpoint_NoMatch(t *testing.T) {
	s := newUserStack(t, nil)
	u, err := s.users.UpsertFromDiscord(context.Background(), discord.User{ID: "d-100", Username: "alice"}, "", "identify")
	if err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	if _, err := s.users.UpsertFromDiscord(context.Background(), discord.User{ID: "d-200", Username: "bob"}, "", "identify"); err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	meta, err := s.lobby.CreateBy("Creator's table", u.ID)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	admin := adminOnlyToken(t, s.srv)

	resp := doGet(t, s.srv, "/games/"+meta.ID.String()+"/creator?discord_id=d-200", admin)
	defer resp.Body.Close()
	var body gameCreatorResponse
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.IsCreator {
		t.Error("want is_creator: false for a Discord id that isn't the creator")
	}
}

func TestGameCreatorEndpoint_UnknownDiscordID(t *testing.T) {
	s := newUserStack(t, nil)
	u, err := s.users.UpsertFromDiscord(context.Background(), discord.User{ID: "d-100", Username: "alice"}, "", "identify")
	if err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	meta, err := s.lobby.CreateBy("Creator's table", u.ID)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	admin := adminOnlyToken(t, s.srv)

	resp := doGet(t, s.srv, "/games/"+meta.ID.String()+"/creator?discord_id=never-signed-in", admin)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("an unknown discord_id should answer false, not error: got %d", resp.StatusCode)
	}
	var body gameCreatorResponse
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.IsCreator {
		t.Error("want is_creator: false for a discord_id with no linked identity")
	}
}

func TestGameCreatorEndpoint_NoCreatorAlwaysFalse(t *testing.T) {
	s := newUserStack(t, nil)
	if _, err := s.users.UpsertFromDiscord(context.Background(), discord.User{ID: "d-100", Username: "alice"}, "", "identify"); err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	// Plain Create: no creator, exactly like an admin-created or
	// file-imported table.
	meta, err := s.lobby.Create("Admin's table")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	admin := adminOnlyToken(t, s.srv)

	resp := doGet(t, s.srv, "/games/"+meta.ID.String()+"/creator?discord_id=d-100", admin)
	defer resp.Body.Close()
	var body gameCreatorResponse
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.IsCreator {
		t.Error("a game with no creator must answer false for every discord_id, never true")
	}
}

func TestGameCreatorEndpoint_NoDBPathAnswersFalseNotError(t *testing.T) {
	// newTestHTTPStack wires no users.Store (Config.Users is nil,
	// users.NoStore behind userStore()) — the shape of a deployment
	// with CMDCTRL_DATA_DIR unset. The rotate creator check doesn't
	// need a Users store at all, but this endpoint does, so this is
	// its dedicated no-DB coverage.
	srv, l, _ := newTestHTTPStack(t)
	creator := uuid.New()
	meta, err := l.CreateBy("Creator's table", creator)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var admin sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&admin)
	resp.Body.Close()

	resp = doGet(t, srv, "/games/"+meta.ID.String()+"/creator?discord_id=whatever", admin.Token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("no-DB path should answer 200/false, not error: got %d", resp.StatusCode)
	}
	var body gameCreatorResponse
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.IsCreator {
		t.Error("no-DB deployment can never resolve a discord_id to a user, so is_creator must be false")
	}
}

func TestGameCreatorEndpoint_RequiresAdmin(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	creator := uuid.New()
	meta, err := l.CreateBy("Creator's table", creator)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	// Even the game's own creator may not call this route — it is
	// admin-only by design (see the route registration comment).
	tok := identifiedToken(t, a, creator)

	resp := doGet(t, srv, "/games/"+meta.ID.String()+"/creator?discord_id=d-1", tok)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("non-admin caller: got %d, want 403", resp.StatusCode)
	}

	anon := doGet(t, srv, "/games/"+meta.ID.String()+"/creator?discord_id=d-1", "")
	defer anon.Body.Close()
	if anon.StatusCode != http.StatusUnauthorized {
		t.Errorf("anonymous caller: got %d, want 401", anon.StatusCode)
	}
}

func TestGameCreatorEndpoint_NotFound(t *testing.T) {
	srv, _, _ := newTestHTTPStack(t)
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var admin sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&admin)
	resp.Body.Close()

	resp = doGet(t, srv, "/games/"+uuid.New().String()+"/creator?discord_id=d-1", admin.Token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown game: got %d, want 404", resp.StatusCode)
	}
}

func TestGameCreatorEndpoint_MissingDiscordID(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	meta, err := l.CreateBy("Creator's table", uuid.New())
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var admin sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&admin)
	resp.Body.Close()

	resp = doGet(t, srv, "/games/"+meta.ID.String()+"/creator", admin.Token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("missing discord_id: got %d, want 400", resp.StatusCode)
	}
}
