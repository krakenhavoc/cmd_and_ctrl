package lobby

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// Tests for ADR 0156 §1, the sign-up gate: the role check at sign-in
// in both callback flows, the re-check at renewal, the admin exemption,
// and GET /community (§6). The off-gate case is every other Discord
// test in this package, unchanged; TestGateOffSignInMakesNoCall pins
// that it makes no Discord call and writes nothing new.

const (
	gateGuild = "111111111111111111"
	gateRole  = "222222222222222222"
	// The member record Discord returns, with and without the role.
	memberWithRole    = `{"roles":["999","` + gateRole + `"],"nick":"Al"}`
	memberWithoutRole = `{"roles":["999"],"nick":"Al"}`
)

// lockedBuffer is a log sink the handler goroutine writes and the test
// reads.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

// signupStack is the Discord test stack over a real database, with
// sealed refresh tokens, per-user revocation, a clock the test moves,
// and the logs captured.
type signupStack struct {
	revocationStack
	users   *users.SQLStore
	db      *db.DB
	clock   *testClock
	logs    *lockedBuffer
	evictor *recordingEvictor
}

// newSignupStack builds the stack with the gate on (gate true) or off.
func newSignupStack(t *testing.T, gate bool, configure func(*Config)) signupStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	d := openTestDB(t, t.TempDir())
	us := users.NewSQLStore(d, mustSealer(t))
	rev, err := users.NewRevocations(context.Background(), us)
	if err != nil {
		t.Fatalf("NewRevocations: %v", err)
	}
	h, err := auth.NewHMACAuthenticator([]byte("lobby-test-session-key-0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewHMACAuthenticator: %v", err)
	}
	a := auth.WithRevocation(h, rev)
	ev := &recordingEvictor{}
	clock := &testClock{}
	logs := &lockedBuffer{}
	srv, l, stub, state := newDiscordTestStackWith(t, func(c *Config) {
		c.Lobby = NewLobbyWithStore(ws.NewRoomManager(quietLogger(), ""), NewSQLStore(d))
		c.Auth = a
		c.Users = us
		c.Revocations = rev
		c.SessionEvictor = ev
		c.Now = clock.Now
		c.Log = slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelInfo}))
		if gate {
			c.Discord.Gate = discord.Gate{GuildID: gateGuild, RoleID: gateRole}
		}
		if configure != nil {
			configure(c)
		}
	})
	// A gated sign-in grants both scopes and a refresh token.
	stub.tokenBody = `{"access_token":"at-signin","token_type":"Bearer","expires_in":3600,"refresh_token":"rt-signin","scope":"identify guilds.members.read"}`
	return signupStack{
		revocationStack: revocationStack{srv: srv, lobby: l, stub: stub, state: state, auth: a, evictor: ev},
		users:           us, db: d, clock: clock, logs: logs, evictor: ev,
	}
}

// loginCallback walks the login-page callback and returns the response.
func (s signupStack) loginCallback(t *testing.T) *http.Response {
	t.Helper()
	state, _, err := s.state.Start(uuid.Nil, "")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	return getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/callback?state="+state+"&code=code-1", "")
}

// inviteCallback walks the invite-link callback for meta.
func (s signupStack) inviteCallback(t *testing.T, meta GameMeta) *http.Response {
	t.Helper()
	state, _, err := s.state.Start(meta.ID, meta.InviteToken)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	return getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/callback?state="+state+"&code=code-1", "")
}

func (s signupStack) count(t *testing.T, q string) int {
	t.Helper()
	var n int
	if err := s.db.QueryRow(q).Scan(&n); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return n
}

// access reads the stored access check for a user.
func (s signupStack) access(t *testing.T, id uuid.UUID) (sql.NullInt64, sql.NullString) {
	t.Helper()
	var (
		at  sql.NullInt64
		res sql.NullString
	)
	if err := s.db.QueryRow(`SELECT access_checked_at, access_result FROM users WHERE id = ?`, id.String()).Scan(&at, &res); err != nil {
		t.Fatalf("read access: %v", err)
	}
	return at, res
}

// signInAdmitted signs in with the member check answering "has the
// role" and returns the identity token.
func (s signupStack) signInAdmitted(t *testing.T) string {
	t.Helper()
	s.stub.setMember(http.StatusOK, memberWithRole)
	tok := identityTokenFromCallback(t, s.srv, s.state)
	s.stub.calls()
	return tok
}

// location reads a 302's Location and checks no session was handed out.
func refusedTo(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("callback: %d %s, want 302", resp.StatusCode, body)
	}
	if c, ok := sessionCookie(resp); ok && c.Value != "" {
		t.Errorf("a refused sign-in set a session cookie")
	}
	loc := resp.Header.Get("Location")
	if strings.Contains(loc, "token=") {
		t.Errorf("a refused sign-in handed out a token: %s", loc)
	}
	return loc
}

// --- the gate off --------------------------------------------------------

// TestGateOffSignInMakesNoCall: with the gate off, sign-in and renewal
// are what they were before ADR 0156. No member call, no refresh grant,
// `identify` only, and nothing written to the access columns.
func TestGateOffSignInMakesNoCall(t *testing.T) {
	s := newSignupStack(t, false, nil)
	loc := followOneRedirect(t, s.srv, "/auth/discord/start")
	if u, _ := url.Parse(loc); u.Query().Get("scope") != "identify" {
		t.Errorf("gate off: authorize scope %q, want identify", u.Query().Get("scope"))
	}
	tok := identityTokenFromCallback(t, s.srv, s.state)
	p := mustValidate(t, s.auth, tok)
	if p.UserID == uuid.Nil {
		t.Fatal("gate off: the sign-in carries no user")
	}
	meta, _ := s.lobby.Create("FNM")
	frag := fragmentParams(t, followOneRedirect(t, s.srv, "/auth/discord/callback?state="+mustStart(t, s.state, meta)+"&code=c"))
	if seat := mustValidate(t, s.auth, frag.Get("token")); seat.UserID != p.UserID || seat.Role != auth.RolePlayer {
		t.Errorf("gate off: the invite flow seated %+v", seat)
	}

	s.clock.advance(16 * 24 * time.Hour)
	afterAMillisecond()
	if got := minted(t, renew(t, s.revocationStack, tok), http.StatusOK); got.Token == tok {
		t.Error("gate off: a half-spent session was not renewed")
	}
	if m, r := s.stub.calls(); m != 0 || r != 0 {
		t.Errorf("gate off: %d member checks and %d refresh grants, want none", m, r)
	}
	if at, res := s.access(t, p.UserID); at.Valid || res.Valid {
		t.Errorf("gate off: access columns written: %v %v", at, res)
	}
}

func mustStart(t *testing.T, store *discord.StateStore, meta GameMeta) string {
	t.Helper()
	state, _, err := store.Start(meta.ID, meta.InviteToken)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	return state
}

// --- sign-in -------------------------------------------------------------

func TestGateOnAsksForGuildMembersRead(t *testing.T) {
	s := newSignupStack(t, true, nil)
	loc := followOneRedirect(t, s.srv, "/auth/discord/start")
	u, _ := url.Parse(loc)
	if got := u.Query().Get("scope"); got != "identify guilds.members.read" {
		t.Errorf("gate on: scope %q", got)
	}
}

func TestGateAdmitsAMemberWithTheRole(t *testing.T) {
	s := newSignupStack(t, true, nil)
	s.stub.setMember(http.StatusOK, memberWithRole)
	p := mustValidate(t, s.auth, identityTokenFromCallback(t, s.srv, s.state))
	if p.UserID == uuid.Nil || p.Role != auth.RoleIdentified {
		t.Fatalf("admitted sign-in = %+v", p)
	}
	if m, _ := s.stub.calls(); m != 1 {
		t.Errorf("member checks = %d, want 1", m)
	}
	at, res := s.access(t, p.UserID)
	if !at.Valid || res.String != "admitted" {
		t.Errorf("access after an admitted sign-in: %v %v", at, res)
	}
	var scopes string
	_ = s.db.QueryRow(`SELECT scopes FROM identities WHERE subject = 'discord-99'`).Scan(&scopes)
	if scopes != "identify guilds.members.read" {
		t.Errorf("identity scopes = %q", scopes)
	}
}

func TestGateExemptsAnAllowlistedAdmin(t *testing.T) {
	s := newSignupStack(t, true, func(c *Config) {
		c.Admins = NewAdmins(NewAdminList("discord-99"), nil)
	})
	s.stub.setMember(http.StatusNotFound, `{}`)
	p := mustValidate(t, s.auth, identityTokenFromCallback(t, s.srv, s.state))
	if p.UserID == uuid.Nil {
		t.Fatal("an allowlisted admin who is not a member was refused")
	}
	if m, _ := s.stub.calls(); m != 0 {
		t.Errorf("an exempt admin cost %d member checks, want 0", m)
	}
	if at, res := s.access(t, p.UserID); at.Valid || res.Valid {
		t.Errorf("an exempt admin has an access check recorded: %v %v", at, res)
	}
}

// TestGateRefusesTheLoginFlow: not a member, or a member without the
// role, is sent to the gate page, gets no session, and leaves no row.
func TestGateRefusesTheLoginFlow(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"not a member", http.StatusNotFound, `{"message":"Unknown Guild","code":10004}`},
		{"no role", http.StatusOK, memberWithoutRole},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newSignupStack(t, true, nil)
			s.stub.setMember(tc.status, tc.body)
			if loc := refusedTo(t, s.loginCallback(t)); loc != "/#/request-access" {
				t.Errorf("Location %q, want /#/request-access", loc)
			}
			if n := s.count(t, `SELECT COUNT(*) FROM users`); n != 0 {
				t.Errorf("a refused sign-in left %d users rows", n)
			}
			if n := s.count(t, `SELECT COUNT(*) FROM identities`); n != 0 {
				t.Errorf("a refused sign-in left %d identities rows", n)
			}
		})
	}
}

// An account that exists and has since lost the role is refused at
// sign-in without its row being touched: no new name, no new refresh
// token, no access write (ADR 0156 §1 item 4).
func TestGateRefusalLeavesAnExistingAccountUntouched(t *testing.T) {
	s := newSignupStack(t, true, nil)
	p := mustValidate(t, s.auth, s.signInAdmitted(t))
	var before string
	_ = s.db.QueryRow(`SELECT display_name || '|' || last_seen_at || '|' || access_checked_at FROM users WHERE id = ?`, p.UserID.String()).Scan(&before)
	rtBefore, _, _ := s.users.RefreshToken(context.Background(), "discord-99")

	s.stub.setMember(http.StatusOK, memberWithoutRole)
	s.stub.setUser(`{"id":"discord-99","username":"alice","global_name":"Renamed","avatar":"newhash"}`)
	s.stub.tokenBody = `{"access_token":"at-2","refresh_token":"rt-second","scope":"identify guilds.members.read"}`
	if loc := refusedTo(t, s.loginCallback(t)); loc != "/#/request-access" {
		t.Errorf("Location %q", loc)
	}
	var after string
	_ = s.db.QueryRow(`SELECT display_name || '|' || last_seen_at || '|' || access_checked_at FROM users WHERE id = ?`, p.UserID.String()).Scan(&after)
	if after != before {
		t.Errorf("the refused sign-in updated the row: %q -> %q", before, after)
	}
	if rt, _, _ := s.users.RefreshToken(context.Background(), "discord-99"); rt != rtBefore {
		t.Errorf("the refused sign-in replaced the refresh token")
	}
}

// TestGateOutageAtSignInFailsClosed: Discord not answering refuses the
// sign-in, with the retry page, and writes nothing, in either flow.
func TestGateOutageAtSignInFailsClosed(t *testing.T) {
	s := newSignupStack(t, true, nil)
	s.stub.setMember(http.StatusServiceUnavailable, ``)
	if loc := refusedTo(t, s.loginCallback(t)); loc != "/#/request-access?retry=1" {
		t.Errorf("login flow: Location %q, want /#/request-access?retry=1", loc)
	}
	meta, _ := s.lobby.Create("FNM")
	if loc := refusedTo(t, s.inviteCallback(t, meta)); loc != "/#/request-access?retry=1" {
		t.Errorf("invite flow: Location %q, want /#/request-access?retry=1", loc)
	}
	if got, _ := s.lobby.Get(meta.ID); len(got.Players) != 0 {
		t.Errorf("an outage took a seat: %+v", got.Players)
	}
	if n := s.count(t, `SELECT COUNT(*) FROM users`); n != 0 {
		t.Errorf("an outage left %d users rows", n)
	}
	if !strings.Contains(s.logs.String(), "level=WARN") {
		t.Errorf("the outage was not logged at WARN:\n%s", s.logs.String())
	}
}

// TestGateRefusedInviteSeatsAGuest: the host's link vouches for the
// table. A refused person is seated as a guest named from Discord, with
// no user and no Discord identity, and still no users row.
func TestGateRefusedInviteSeatsAGuest(t *testing.T) {
	s := newSignupStack(t, true, nil)
	s.stub.setMember(http.StatusNotFound, `{}`)
	meta, _ := s.lobby.Create("FNM")
	resp := s.inviteCallback(t, meta)
	resp.Body.Close()
	frag := fragmentParams(t, resp.Header.Get("Location"))
	if frag.Get("user_id") != "" {
		t.Errorf("the guest's fragment names a user: %v", frag)
	}
	p := mustValidate(t, s.auth, frag.Get("token"))
	if p.Role != auth.RolePlayer || p.UserID != uuid.Nil || p.DiscordID != "" || p.Name != "Alice" || p.GameID != meta.ID {
		t.Errorf("guest session = %+v", p)
	}
	if d := p.ExpiresAt.Sub(p.IssuedAt); d != sessionTTL {
		t.Errorf("guest session lives %v, want the guest lifetime %v", d, sessionTTL)
	}
	got, _ := s.lobby.Get(meta.ID)
	if len(got.Players) != 1 || got.Players[0].Name != "Alice" || got.Players[0].DiscordID != "" || got.Players[0].PlayerID != p.PlayerID {
		t.Errorf("seats = %+v", got.Players)
	}
	if n := s.count(t, `SELECT COUNT(*) FROM users`); n != 0 {
		t.Errorf("a refused invite sign-in left %d users rows", n)
	}
	if n := s.count(t, `SELECT COUNT(*) FROM seats WHERE user_id IS NOT NULL OR pending_discord_id IS NOT NULL`); n != 0 {
		t.Errorf("the guest seat carries an account: %d", n)
	}
}

// An admitted invite sign-in is today's: a seat with the user.
func TestGateAdmittedInviteSeatsTheUser(t *testing.T) {
	s := newSignupStack(t, true, nil)
	s.stub.setMember(http.StatusOK, memberWithRole)
	meta, _ := s.lobby.Create("FNM")
	resp := s.inviteCallback(t, meta)
	resp.Body.Close()
	p := mustValidate(t, s.auth, fragmentParams(t, resp.Header.Get("Location")).Get("token"))
	if p.UserID == uuid.Nil || p.DiscordID != "discord-99" {
		t.Errorf("admitted invite seat = %+v", p)
	}
}

// --- renewal -------------------------------------------------------------

// halfSpentRenew moves past half-life and renews tok.
func (s signupStack) halfSpentRenew(t *testing.T, tok string) *http.Response {
	t.Helper()
	s.clock.advance(15*24*time.Hour + time.Minute)
	afterAMillisecond()
	return renew(t, s.revocationStack, tok)
}

func TestRenewalReCheckAdmitted(t *testing.T) {
	s := newSignupStack(t, true, nil)
	tok := s.signInAdmitted(t)
	user := mustValidate(t, s.auth, tok).UserID

	got := minted(t, s.halfSpentRenew(t, tok), http.StatusOK)
	if got.Token == tok {
		t.Fatal("an admitted re-check did not renew")
	}
	if m, r := s.stub.calls(); m != 1 || r != 1 {
		t.Errorf("re-check made %d member checks and %d refresh grants, want 1 and 1", m, r)
	}
	s.stub.mu.Lock()
	sent := s.stub.lastRefresh
	s.stub.mu.Unlock()
	if sent != "rt-signin" {
		t.Errorf("the refresh grant carried %q, want the stored token", sent)
	}
	if rt, ok, err := s.users.RefreshToken(context.Background(), "discord-99"); err != nil || !ok || rt != "rt-rotated" {
		t.Errorf("stored refresh token after the re-check = %q %v %v, want the rotated one", rt, ok, err)
	}
	at, res := s.access(t, user)
	if res.String != "admitted" || at.Int64 < s.clock.Now().Add(-time.Minute).UnixMilli() {
		t.Errorf("access after the re-check: %v %v, want admitted at the renewal", at, res)
	}
}

func TestRenewalReCheckRefusedRevokesEverySession(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"role removed", http.StatusOK, memberWithoutRole, "no_role"},
		{"left the server", http.StatusNotFound, `{}`, "not_member"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newSignupStack(t, true, nil)
			tok := s.signInAdmitted(t)
			other := s.signInAdmitted(t) // a second browser
			user := mustValidate(t, s.auth, tok).UserID

			s.stub.setMember(tc.status, tc.body)
			resp := s.halfSpentRenew(t, tok)
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("refused re-check: %d %s, want 401", resp.StatusCode, body)
			}
			if strings.Contains(string(body), `"token"`) {
				t.Errorf("a refused re-check handed out a token: %s", body)
			}
			for name, tk := range map[string]string{"this session": tok, "another browser": other} {
				if code, _ := statusOf(t, s.srv, tk); code != http.StatusUnauthorized {
					t.Errorf("%s after a refused re-check: %d, want 401 (watermark moved)", name, code)
				}
			}
			u, err := s.users.Get(context.Background(), user)
			if err != nil || u.SessionsInvalidBefore.IsZero() {
				t.Errorf("watermark not moved: %+v %v", u, err)
			}
			if _, res := s.access(t, user); res.String != tc.want {
				t.Errorf("access result %q, want %q", res.String, tc.want)
			}
			s.evictor.mu.Lock()
			evicted := len(s.evictor.calls)
			s.evictor.mu.Unlock()
			if evicted != 1 {
				t.Errorf("sockets evicted %d times, want 1", evicted)
			}
			// A fresh sign-in is the way back, and it is checked too.
			s.stub.setMember(http.StatusOK, memberWithRole)
			if fresh := s.signInAdmitted(t); mustValidate(t, s.auth, fresh).UserID != user {
				t.Error("signing in again after the role came back is a different person")
			}
		})
	}
}

func TestRenewalWithNoRefreshTokenIsNotRenewed(t *testing.T) {
	s := newSignupStack(t, true, nil)
	s.stub.tokenBody = `{"access_token":"at-signin","scope":"identify guilds.members.read"}`
	tok := s.signInAdmitted(t)
	old := mustValidate(t, s.auth, tok)

	got := minted(t, s.halfSpentRenew(t, tok), http.StatusOK)
	if got.Token != tok || !got.ExpiresAt.Equal(old.ExpiresAt) {
		t.Errorf("no refresh token: renewed to %s, want the same session back", got.ExpiresAt)
	}
	if m, r := s.stub.calls(); m != 0 || r != 0 {
		t.Errorf("no refresh token: %d member checks, %d refresh grants; want none", m, r)
	}
	if code, _ := statusOf(t, s.srv, tok); code != http.StatusOK {
		t.Errorf("the session after a held renewal: %d, want it still valid", code)
	}
}

func TestRenewalWithARejectedRefreshTokenIsNotRenewed(t *testing.T) {
	s := newSignupStack(t, true, nil)
	tok := s.signInAdmitted(t)
	s.stub.setRefresh(http.StatusBadRequest, `{"error":"invalid_grant"}`)

	got := minted(t, s.halfSpentRenew(t, tok), http.StatusOK)
	if got.Token != tok {
		t.Error("a rejected refresh token renewed the session")
	}
	if m, r := s.stub.calls(); m != 0 || r != 1 {
		t.Errorf("%d member checks, %d refresh grants; want 0 and 1", m, r)
	}
	if code, _ := statusOf(t, s.srv, tok); code != http.StatusOK {
		t.Errorf("the session after a held renewal: %d, want it still valid", code)
	}
}

func TestRenewalOutageRenewsAndWarns(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*discordStub)
	}{
		{"member check down", func(st *discordStub) { st.setMember(http.StatusBadGateway, ``) }},
		{"token endpoint down", func(st *discordStub) { st.setRefresh(http.StatusServiceUnavailable, ``) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newSignupStack(t, true, nil)
			tok := s.signInAdmitted(t)
			user := mustValidate(t, s.auth, tok).UserID
			atBefore, _ := s.access(t, user)
			tc.setup(s.stub)

			got := minted(t, s.halfSpentRenew(t, tok), http.StatusOK)
			if got.Token == tok {
				t.Error("an outage stopped the renewal; it must not sign players out")
			}
			if !strings.Contains(s.logs.String(), "level=WARN") || !strings.Contains(s.logs.String(), "renewing anyway") {
				t.Errorf("the outage was not logged at WARN:\n%s", s.logs.String())
			}
			if at, res := s.access(t, user); at != atBefore || res.String != "admitted" {
				t.Errorf("an outage overwrote the last answer: %v %v", at, res)
			}
		})
	}
}

func TestRenewalExemptsAnAllowlistedAdmin(t *testing.T) {
	s := newSignupStack(t, true, func(c *Config) {
		c.Admins = NewAdmins(NewAdminList("discord-99"), nil)
	})
	tok := identityTokenFromCallback(t, s.srv, s.state)
	s.stub.setMember(http.StatusNotFound, `{}`)
	got := minted(t, s.halfSpentRenew(t, tok), http.StatusOK)
	if got.Token == tok {
		t.Error("an exempt admin was not renewed")
	}
	if m, r := s.stub.calls(); m != 0 || r != 0 {
		t.Errorf("an exempt admin's renewal made %d member checks, %d refresh grants", m, r)
	}
}

// Below half-life a renewal is a reinstall and asks Discord nothing.
func TestRenewalBelowHalfLifeAsksNothing(t *testing.T) {
	s := newSignupStack(t, true, nil)
	tok := s.signInAdmitted(t)
	s.stub.setMember(http.StatusNotFound, `{}`)
	if got := minted(t, renew(t, s.revocationStack, tok), http.StatusOK); got.Token != tok {
		t.Error("a fresh session was re-issued")
	}
	if m, r := s.stub.calls(); m != 0 || r != 0 {
		t.Errorf("%d member checks, %d refresh grants below half-life; want none", m, r)
	}
}

// --- GET /community ------------------------------------------------------

func TestCommunityIsPublicCacheableAndNamesNoIDs(t *testing.T) {
	const voiceCategory = "424242424242424242"
	s := newSignupStack(t, true, func(c *Config) {
		c.CommunityInviteURL = "https://discord.gg/c2table"
		c.DiscordVoice = discord.Voice{Token: "bot-token", GuildID: gateGuild, CategoryID: voiceCategory}
	})
	resp := doGet(t, s.srv, "/community", "")
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /community with no session: %d %s", resp.StatusCode, raw)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "public, max-age=300" {
		t.Errorf("Cache-Control %q", cc)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 3 || body["invite_url"] != "https://discord.gg/c2table" || body["gate"] != true || body["voice"] != true {
		t.Errorf("body = %s", raw)
	}
	if strings.Contains(string(raw), gateGuild) || strings.Contains(string(raw), gateRole) ||
		strings.Contains(string(raw), voiceCategory) {
		t.Errorf("GET /community serves an id: %s", raw)
	}

	off := newSignupStack(t, false, nil)
	resp2 := doGet(t, off.srv, "/community", "")
	defer resp2.Body.Close()
	var offBody map[string]any
	_ = json.NewDecoder(resp2.Body).Decode(&offBody)
	if offBody["invite_url"] != "" || offBody["gate"] != false || offBody["voice"] != false {
		t.Errorf("gate off, no invite, no voice: %v", offBody)
	}
}

// Linking Discord to a guest seat would make an account, so a refused
// link goes to the gate page and the seat stays the guest's.
func TestGateRefusedLinkLeavesTheGuestSeat(t *testing.T) {
	s := newSignupStack(t, true, nil)
	s.stub.setMember(http.StatusOK, memberWithoutRole)
	guest := guestSeatToken(t, s.srv, s.lobby)
	p := mustValidate(t, s.auth, guest)

	start := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/link?game="+p.GameID.String(), guest)
	start.Body.Close()
	loc, err := url.Parse(start.Header.Get("Location"))
	if err != nil || loc.Query().Get("state") == "" {
		t.Fatalf("link start: %d %v", start.StatusCode, start.Header.Get("Location"))
	}
	resp := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/callback?state="+loc.Query().Get("state")+"&code=c", guest)
	if got := refusedTo(t, resp); got != "/#/request-access" {
		t.Errorf("Location %q, want /#/request-access", got)
	}
	meta, _ := s.lobby.Get(p.GameID)
	if len(meta.Players) != 1 || meta.Players[0].DiscordID != "" || meta.Players[0].Name != "Guest" {
		t.Errorf("the seat after a refused link: %+v", meta.Players)
	}
	if n := s.count(t, `SELECT COUNT(*) FROM users`); n != 0 {
		t.Errorf("a refused link left %d users rows", n)
	}
	if code, _ := statusOf(t, s.srv, guest); code != http.StatusOK {
		t.Errorf("the guest's own session after a refused link: %d", code)
	}
}
