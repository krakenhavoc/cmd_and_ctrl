package lobby

// Tests for POST /games/{id}/invites/dm — ADR 0051 decision 5's
// direct-message invite (S34 sub-PR 6, tracking #607). Discord is an
// httptest.Server throughout; nothing here reaches the network.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

const dmInviteBase = "https://cmd.example"

// dmStub stands in for Discord's REST API and records what the server
// sent it.
type dmStub struct {
	mu sync.Mutex

	openStatus, msgStatus int
	openBody              string

	opens      int
	messages   int
	recipients []string
	contents   []string
}

func (s *dmStub) snapshot() (opens, messages int, recipients, contents []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.opens, s.messages,
		append([]string(nil), s.recipients...), append([]string(nil), s.contents...)
}

func (s *dmStub) setStatuses(open, msg int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.openStatus, s.msgStatus = open, msg
}

// newDMStack wires a database-backed stack whose Discord REST calls
// land on a stub. token is CMDCTRL_DISCORD_BOT_TOKEN's value — pass
// "" for the unconfigured deployment. base is the invite origin — ""
// for a deployment that has none. relax turns the rate limiters off,
// which every test but the rate-limit one wants.
func newDMStack(t *testing.T, token, base string, relax bool) (userStack, *dmStub) {
	t.Helper()
	if relax {
		t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	}
	stub := &dmStub{openStatus: http.StatusOK, msgStatus: http.StatusOK, openBody: `{"id":"dm-1"}`}
	mux := http.NewServeMux()
	mux.HandleFunc("/users/@me/channels", func(w http.ResponseWriter, r *http.Request) {
		stub.mu.Lock()
		defer stub.mu.Unlock()
		stub.opens++
		var body struct {
			RecipientID string `json:"recipient_id"`
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		stub.recipients = append(stub.recipients, body.RecipientID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(stub.openStatus)
		_, _ = w.Write([]byte(stub.openBody))
	})
	mux.HandleFunc("/channels/", func(w http.ResponseWriter, r *http.Request) {
		stub.mu.Lock()
		defer stub.mu.Unlock()
		stub.messages++
		var body struct {
			Content string `json:"content"`
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		stub.contents = append(stub.contents, body.Content)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(stub.msgStatus)
		_, _ = w.Write([]byte(`{"id":"m-1"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Cleanup(discord.SwapBotAPIBaseForTesting(srv.URL))

	s := newUserStackWith(t, nil, func(c *Config) {
		c.DiscordBot = discord.Bot{Token: token}
		c.InviteBaseURL = base
	})
	return s, stub
}

// playerTokenWithUser mints a seat session that also carries a person
// — what a signed-in player actually holds after JoinAs.
func playerTokenWithUser(t *testing.T, a auth.Authenticator, gameID, playerID, userID uuid.UUID, name string) string {
	t.Helper()
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RolePlayer, GameID: gameID, PlayerID: playerID, UserID: userID, Name: name,
	}, time.Hour)
	if err != nil {
		t.Fatalf("issue player session: %v", err)
	}
	return tok
}

// identityToken mints a signed-in-but-unseated session for a user,
// without walking the whole OAuth callback again.
func identityToken(t *testing.T, a auth.Authenticator, userID uuid.UUID, discordID, name string) string {
	t.Helper()
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RoleIdentified, UserID: userID,
		DiscordID: discordID, DiscordGlobalName: name, Name: name,
	}, time.Hour)
	if err != nil {
		t.Fatalf("issue identity session: %v", err)
	}
	return tok
}

// dmTarget adds a user with a Discord identity but no seat: somebody
// to invite.
func dmTarget(t *testing.T, s userStack, snowflake, name string) users.User {
	t.Helper()
	u, err := s.users.UpsertFromDiscord(context.Background(),
		discord.User{ID: snowflake, Username: strings.ToLower(name), GlobalName: name}, "", "identify")
	if err != nil {
		t.Fatalf("UpsertFromDiscord %s: %v", name, err)
	}
	return u
}

// postDM calls the route and returns the status and the raw body.
func postDM(t *testing.T, s userStack, gameID uuid.UUID, tok string, body any) (int, string) {
	t.Helper()
	resp := postJSON(t, s.srv, "/games/"+gameID.String()+"/invites/dm", tok, body)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

// seatedStack is the common arrangement: a signed-in Alice holding a
// seat at a table she created, and a Bob to invite.
type seatedStack struct {
	userStack
	stub  *dmStub
	meta  GameMeta
	alice uuid.UUID
	// aliceSeat is a player session for Alice's seat (what the client
	// actually holds once she has sat down).
	aliceSeat string
	bob       users.User
}

func newSeatedStack(t *testing.T, token string) seatedStack {
	t.Helper()
	s, stub := newDMStack(t, token, dmInviteBase, true)
	tok := identityTokenFromCallback(t, s.srv, s.state)
	alice := mustValidate(t, s.auth, tok).UserID

	meta, err := s.lobby.CreateBy("FNM", alice)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	_, playerID, err := s.lobby.JoinAs(meta.ID, meta.InviteToken, "",
		DiscordIdentity{ID: "discord-99", Username: "alice", GlobalName: "Alice"}, alice)
	if err != nil {
		t.Fatalf("JoinAs: %v", err)
	}
	seat := playerTokenWithUser(t, s.auth, meta.ID, playerID, alice, "Alice")
	return seatedStack{
		userStack: s, stub: stub, meta: meta, alice: alice,
		aliceSeat: seat, bob: dmTarget(t, s, "d-bob", "Bob"),
	}
}

func TestInviteDMSendsTheGamesExistingInviteLink(t *testing.T) {
	s := newSeatedStack(t, "bot-token-shhh")

	status, raw := postDM(t, s.userStack, s.meta.ID, s.aliceSeat,
		map[string]string{"user_id": s.bob.ID.String()})
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, raw)
	}
	var body dmInviteResponse
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatalf("decode: %v (%s)", err, raw)
	}
	if !body.Sent || body.UserID != s.bob.ID.String() || body.DisplayName != "Bob" {
		t.Errorf("body = %+v", body)
	}

	opens, msgs, recipients, contents := s.stub.snapshot()
	if opens != 1 || msgs != 1 {
		t.Fatalf("Discord calls: %d opens, %d messages; want 1 each", opens, msgs)
	}
	if recipients[0] != "d-bob" {
		t.Errorf("recipient_id = %q, want Bob's snowflake resolved from our user id", recipients[0])
	}
	// Nothing new is minted: the DM carries the game's CURRENT player
	// invite, which still works for everyone who already has it.
	wantURL := dmInviteBase + "/#/games/" + s.meta.ID.String() + "/join?t=" + s.meta.InviteToken
	if !strings.Contains(contents[0], wantURL) {
		t.Errorf("DM content %q does not carry %q", contents[0], wantURL)
	}
	if !strings.Contains(contents[0], "Alice") || !strings.Contains(contents[0], "FNM") {
		t.Errorf("DM content %q names neither the inviter nor the table", contents[0])
	}
	after, err := s.lobby.Get(s.meta.ID)
	if err != nil || after.InviteToken != s.meta.InviteToken {
		t.Errorf("the invite moved: %q -> %q (%v)", s.meta.InviteToken, after.InviteToken, err)
	}
	// The response never echoes the token or the snowflake.
	if strings.Contains(raw, s.meta.InviteToken) || strings.Contains(raw, "d-bob") {
		t.Errorf("the response leaks the invite or the snowflake: %s", raw)
	}
}

func TestInviteDMAuthorization(t *testing.T) {
	s := newSeatedStack(t, "bot-token-shhh")
	body := map[string]string{"user_id": s.bob.ID.String()}

	t.Run("a seated player may", func(t *testing.T) {
		if status, raw := postDM(t, s.userStack, s.meta.ID, s.aliceSeat, body); status != http.StatusOK {
			t.Errorf("status %d: %s", status, raw)
		}
	})

	t.Run("the admin may", func(t *testing.T) {
		if status, raw := postDM(t, s.userStack, s.meta.ID, adminToken(t, s.srv), body); status != http.StatusOK {
			t.Errorf("status %d: %s", status, raw)
		}
	})

	t.Run("the creator may, without a seat", func(t *testing.T) {
		// Carol creates a table and never sits down. Her identity
		// session is not bound to the game at all.
		carol := dmTarget(t, s.userStack, "d-carol", "Carol")
		meta, err := s.lobby.CreateBy("Carol's table", carol.ID)
		if err != nil {
			t.Fatalf("CreateBy: %v", err)
		}
		tok := identityToken(t, s.auth, carol.ID, "d-carol", "Carol")
		if status, raw := postDM(t, s.userStack, meta.ID, tok, body); status != http.StatusOK {
			t.Errorf("creator: status %d: %s", status, raw)
		}
	})

	t.Run("a signed-in stranger may not", func(t *testing.T) {
		dave := dmTarget(t, s.userStack, "d-dave", "Dave")
		tok := identityToken(t, s.auth, dave.ID, "d-dave", "Dave")
		status, raw := postDM(t, s.userStack, s.meta.ID, tok, body)
		if status != http.StatusForbidden {
			t.Errorf("stranger: status %d, want 403: %s", status, raw)
		}
	})

	t.Run("a player at another table may not", func(t *testing.T) {
		other, err := s.lobby.Create("somewhere else")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		_, pid, err := s.lobby.Join(other.ID, other.InviteToken, "Elsewhere")
		if err != nil {
			t.Fatalf("Join: %v", err)
		}
		tok := playerToken(t, s.auth, other.ID, pid, "Elsewhere")
		if status, raw := postDM(t, s.userStack, s.meta.ID, tok, body); status != http.StatusForbidden {
			t.Errorf("other table: status %d, want 403: %s", status, raw)
		}
	})

	t.Run("no session at all", func(t *testing.T) {
		if status, _ := postDM(t, s.userStack, s.meta.ID, "", body); status != http.StatusUnauthorized {
			t.Errorf("status %d, want 401", status)
		}
	})

	t.Run("an unknown game is a 404", func(t *testing.T) {
		if status, _ := postDM(t, s.userStack, uuid.New(), s.aliceSeat, body); status != http.StatusNotFound {
			t.Errorf("status %d, want 404", status)
		}
	})
}

// TestInviteDMDiscordIDIsAdminOnly pins the decision recorded in the
// route's comment: an ordinary caller names a person by OUR user id,
// and only the admin credential (which #613's slash command already
// holds) may hand over a raw snowflake.
func TestInviteDMDiscordIDIsAdminOnly(t *testing.T) {
	s := newSeatedStack(t, "bot-token-shhh")
	body := map[string]string{"discord_id": "d-someone-else"}

	status, raw := postDM(t, s.userStack, s.meta.ID, s.aliceSeat, body)
	if status != http.StatusForbidden {
		t.Errorf("seated player with discord_id: status %d, want 403: %s", status, raw)
	}
	if opens, _, _, _ := s.stub.snapshot(); opens != 0 {
		t.Errorf("a refused request still reached Discord (%d opens)", opens)
	}

	if status, raw := postDM(t, s.userStack, s.meta.ID, adminToken(t, s.srv), body); status != http.StatusOK {
		t.Errorf("admin with discord_id: status %d, want 200: %s", status, raw)
	}
	if _, _, recipients, _ := s.stub.snapshot(); len(recipients) != 1 || recipients[0] != "d-someone-else" {
		t.Errorf("recipients = %v, want the snowflake straight through", recipients)
	}
}

func TestInviteDMBodyValidation(t *testing.T) {
	s := newSeatedStack(t, "bot-token-shhh")
	cases := []struct {
		name string
		body map[string]string
		want int
	}{
		{"empty", map[string]string{}, http.StatusBadRequest},
		{"both", map[string]string{"user_id": s.bob.ID.String(), "discord_id": "d-x"}, http.StatusBadRequest},
		{"not a uuid", map[string]string{"user_id": "bob"}, http.StatusBadRequest},
		{"unknown user", map[string]string{"user_id": uuid.NewString()}, http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if status, raw := postDM(t, s.userStack, s.meta.ID, s.aliceSeat, tc.body); status != tc.want {
				t.Errorf("status %d, want %d: %s", status, tc.want, raw)
			}
		})
	}
}

// TestInviteDMTargetWithNoDiscordIdentity: a users row that no Discord
// account points at has nowhere to receive a DM. It cannot happen
// while Discord is the only provider, so the store is arranged by hand.
func TestInviteDMTargetWithNoDiscordIdentity(t *testing.T) {
	s := newSeatedStack(t, "bot-token-shhh")
	if _, err := s.db.Exec(`DELETE FROM identities WHERE subject = 'd-bob'`); err != nil {
		t.Fatalf("delete identity: %v", err)
	}
	status, raw := postDM(t, s.userStack, s.meta.ID, s.aliceSeat,
		map[string]string{"user_id": s.bob.ID.String()})
	if status != http.StatusUnprocessableEntity {
		t.Errorf("status %d, want 422: %s", status, raw)
	}
}

func TestInviteDMSurfacesDiscordFailures(t *testing.T) {
	cases := []struct {
		name       string
		openStatus int
		want       int
		wantText   string
	}{
		{"403 — no shared guild or DMs closed", http.StatusForbidden, http.StatusUnprocessableEntity, "direct messages closed"},
		{"429 — Discord throttled the bot", http.StatusTooManyRequests, http.StatusTooManyRequests, "rate-limited"},
		{"401 — the bot token was rejected", http.StatusUnauthorized, http.StatusBadGateway, "bot credentials"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newSeatedStack(t, "bot-token-shhh")
			s.stub.setStatuses(tc.openStatus, http.StatusOK)

			status, raw := postDM(t, s.userStack, s.meta.ID, s.aliceSeat,
				map[string]string{"user_id": s.bob.ID.String()})
			if status != tc.want {
				t.Fatalf("status %d, want %d: %s", status, tc.want, raw)
			}
			if !strings.Contains(raw, tc.wantText) {
				t.Errorf("body %q does not explain the failure (want %q)", raw, tc.wantText)
			}
			if strings.Contains(raw, "bot-token-shhh") {
				t.Fatal("the bot token reached the caller")
			}
		})
	}
}

// TestInviteDMWithoutABotTokenIsUnavailable is the "never fails open"
// case: an unconfigured deployment refuses, names the variable, and
// makes no outbound call — while every other route keeps working.
func TestInviteDMWithoutABotTokenIsUnavailable(t *testing.T) {
	s := newSeatedStack(t, "")

	status, raw := postDM(t, s.userStack, s.meta.ID, s.aliceSeat,
		map[string]string{"user_id": s.bob.ID.String()})
	if status != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503: %s", status, raw)
	}
	if !strings.Contains(raw, discord.BotTokenEnv) {
		t.Errorf("the 503 does not name %s: %s", discord.BotTokenEnv, raw)
	}
	if opens, msgs, _, _ := s.stub.snapshot(); opens != 0 || msgs != 0 {
		t.Errorf("an unconfigured server called Discord: %d opens, %d messages", opens, msgs)
	}
	// Everything else on the same deployment is unaffected.
	if status, _, _ := getTablemates(t, s.userStack, s.aliceSeat); status != http.StatusOK {
		t.Errorf("GET /me/tablemates broke with no bot token: %d", status)
	}
	resp := doGet(t, s.srv, "/games/"+s.meta.ID.String(), s.aliceSeat)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /games/{id} broke with no bot token: %d", resp.StatusCode)
	}
}

// TestInviteDMWithoutAnInviteOriginIsUnavailable: a bot token with
// nowhere to point the link is still "not configured".
func TestInviteDMWithoutAnInviteOriginIsUnavailable(t *testing.T) {
	s, stub := newDMStack(t, "bot-token-shhh", "", true)
	meta, err := s.lobby.Create("no origin")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, pid, err := s.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("Join: %v", err)
	}
	bob := dmTarget(t, s, "d-bob", "Bob")
	tok := playerToken(t, s.auth, meta.ID, pid, "Alice")

	status, raw := postDM(t, s, meta.ID, tok, map[string]string{"user_id": bob.ID.String()})
	if status != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503: %s", status, raw)
	}
	if !strings.Contains(raw, "CMDCTRL_PUBLIC_BASE_URL") {
		t.Errorf("the 503 does not name the missing variable: %s", raw)
	}
	if opens, _, _, _ := stub.snapshot(); opens != 0 {
		t.Errorf("called Discord with no link to send (%d opens)", opens)
	}
}

// TestInviteDMWithoutTheInvitePlaintextIs409 covers the restart gap:
// only the process that minted an invite holds its plaintext, so a
// table recovered from the database has a hash and no link. We refuse
// and point at rotation rather than rotating silently — rotating would
// revoke the link the table has already shared.
func TestInviteDMWithoutTheInvitePlaintextIs409(t *testing.T) {
	s := newSeatedStack(t, "bot-token-shhh")
	// Stand in for a restart: the row keeps the hash, this process
	// forgets the plaintext.
	s.lobby.mu.Lock()
	s.lobby.games[s.meta.ID].meta.InviteToken = ""
	s.lobby.mu.Unlock()

	status, raw := postDM(t, s.userStack, s.meta.ID, s.aliceSeat,
		map[string]string{"user_id": s.bob.ID.String()})
	if status != http.StatusConflict {
		t.Fatalf("status %d, want 409: %s", status, raw)
	}
	if !strings.Contains(raw, "invites/rotate") {
		t.Errorf("the 409 does not say how to recover: %s", raw)
	}
	if opens, _, _, _ := s.stub.snapshot(); opens != 0 {
		t.Errorf("DMed a table with no link (%d opens)", opens)
	}
}

// TestInviteDMIsRateLimitedPerCaller runs with the limiters ON: the
// per-caller bucket is 1 DM / 10 s with a burst of 3, so the fourth
// call in a row is refused even though the IP bucket would still
// allow it.
func TestInviteDMIsRateLimitedPerCaller(t *testing.T) {
	s, stub := newDMStack(t, "bot-token-shhh", dmInviteBase, false)
	meta, err := s.lobby.Create("rate me")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, pid, err := s.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("Join: %v", err)
	}
	tok := playerToken(t, s.auth, meta.ID, pid, "Alice")
	bob := dmTarget(t, s, "d-bob", "Bob")
	body := map[string]string{"user_id": bob.ID.String()}

	for i := 1; i <= 3; i++ {
		if status, raw := postDM(t, s, meta.ID, tok, body); status != http.StatusOK {
			t.Fatalf("DM %d: status %d, want 200: %s", i, status, raw)
		}
	}
	status, raw := postDM(t, s, meta.ID, tok, body)
	if status != http.StatusTooManyRequests {
		t.Fatalf("the fourth DM in a row: status %d, want 429: %s", status, raw)
	}
	if opens, _, _, _ := stub.snapshot(); opens != 3 {
		t.Errorf("Discord saw %d opens, want the 3 that were allowed", opens)
	}
}

// TestInviteDMOnADeploymentWithNoDatabase: users.NoStore knows
// nobody, so there is no user id that resolves and the route says so
// rather than erroring. The seat authorisation still works, which is
// the part that must not depend on a database.
func TestInviteDMOnADeploymentWithNoDatabase(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	stub := &dmStub{openStatus: http.StatusOK, msgStatus: http.StatusOK, openBody: `{"id":"dm-1"}`}
	dsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		stub.mu.Lock()
		defer stub.mu.Unlock()
		stub.opens++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"dm-1"}`))
	}))
	t.Cleanup(dsrv.Close)
	t.Cleanup(discord.SwapBotAPIBaseForTesting(dsrv.URL))

	srv, l, _, _ := newDiscordTestStackWith(t, func(c *Config) {
		c.DiscordBot = discord.Bot{Token: "bot-token-shhh"}
		c.InviteBaseURL = dmInviteBase
	})
	meta, err := l.Create("no db")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatalf("Join: %v", err)
	}
	// The admin is the one caller a no-database deployment can still
	// authorise, so it isolates the users lookup as the thing missing.
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/invites/dm", adminToken(t, srv),
		map[string]string{"user_id": uuid.NewString()})
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d, want 404 (there are no users to resolve): %s", resp.StatusCode, raw)
	}
	stub.mu.Lock()
	defer stub.mu.Unlock()
	if stub.opens != 0 {
		t.Errorf("called Discord for a user that cannot exist (%d opens)", stub.opens)
	}
}

func TestDMInviteMessageFlattensNames(t *testing.T) {
	got := dmInviteMessage("Al*ice\nEveryone @here", "Friday `Night`", "https://x/#/games/1/join?t=abc")
	if strings.Count(got, "\n") != 1 {
		t.Errorf("a name forged an extra line: %q", got)
	}
	for _, bad := range []string{"*", "`", "@"} {
		if strings.Contains(got, bad) {
			t.Errorf("%q survived flattening: %q", bad, got)
		}
	}
	if !strings.HasSuffix(got, "https://x/#/games/1/join?t=abc") {
		t.Errorf("the link is not the last line: %q", got)
	}
}

func TestInviteURLMatchesTheBotsShape(t *testing.T) {
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	got := inviteURL("https://cmd.example/", id, "tok-1")
	want := "https://cmd.example/#/games/11111111-2222-3333-4444-555555555555/join?t=tok-1"
	if got != want {
		t.Errorf("inviteURL = %q, want %q", got, want)
	}
}
