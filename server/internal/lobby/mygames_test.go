package lobby

// Tests for S34 sub-PR 4 (ADR 0051, tracking #607): seats.user_id is
// written when a signed-in person claims a seat, pending Discord seats
// are linked at sign-in, GET /me/games lists a user's seats, a user can
// reclaim their own seat without a ticket, and a seated player can link
// Discord to the seat they hold — mid-game included.

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// newMyGamesStack is newUserStack with the rate limiters relaxed: the
// link flow is two requests on the join bucket per round-trip, and
// these tests take several.
func newMyGamesStack(t *testing.T) userStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	return newUserStack(t, nil)
}

// seatRow reads one seat's person columns straight from the database.
func seatRow(t *testing.T, s userStack, gameID, playerID uuid.UUID) (userID, pending sql.NullString) {
	t.Helper()
	err := s.db.QueryRow(`SELECT user_id, pending_discord_id FROM seats WHERE game_id = ? AND player_id = ?`,
		gameID.String(), playerID.String()).Scan(&userID, &pending)
	if err != nil {
		t.Fatalf("read seat row: %v", err)
	}
	return userID, pending
}

// decodeSession reads a sessionResponse body and closes it.
func decodeSession(t *testing.T, resp *http.Response, want int) sessionResponse {
	t.Helper()
	defer resp.Body.Close()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d, want %d: %s", resp.StatusCode, want, body)
	}
	var out sessionResponse
	if want == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			t.Fatalf("decode session: %v", err)
		}
	}
	return out
}

// getMyGames calls GET /me/games and returns the status, the decoded
// list and the raw body.
func getMyGames(t *testing.T, s userStack, tok string) (int, []MyGame, string) {
	t.Helper()
	resp := doGet(t, s.srv, "/me/games", tok)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var body myGamesResponse
	if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("decode /me/games: %v (%s)", err, raw)
		}
	}
	return resp.StatusCode, body.Games, string(raw)
}

// getWithCookie issues a GET carrying tok as the session COOKIE only
// (the shape of a browser navigation) and does not follow a redirect.
func getWithCookie(t *testing.T, srv interface{ Client() *http.Client }, base, path, tok string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, base+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if tok != "" {
		req.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: tok})
	}
	client := *srv.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("get %s: %v", path, err)
	}
	return resp
}

// playerToken mints a guest player session for a seat directly.
func playerToken(t *testing.T, a auth.Authenticator, gameID, playerID uuid.UUID, name string) string {
	t.Helper()
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RolePlayer, GameID: gameID, PlayerID: playerID, Name: name,
	}, time.Hour)
	if err != nil {
		t.Fatalf("issue player session: %v", err)
	}
	return tok
}

// linkDiscord walks GET /auth/discord/link and the callback with tok as
// the browser's cookie, and returns the callback's response.
func linkDiscord(t *testing.T, s userStack, gameID uuid.UUID, tok string) *http.Response {
	t.Helper()
	start := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/link?game="+gameID.String(), tok)
	start.Body.Close()
	if start.StatusCode != http.StatusFound {
		t.Fatalf("link start: %d", start.StatusCode)
	}
	loc, err := url.Parse(start.Header.Get("Location"))
	if err != nil {
		t.Fatalf("parse Location: %v", err)
	}
	state := loc.Query().Get("state")
	if state == "" {
		t.Fatalf("link start did not redirect to Discord with a state: %s", loc)
	}
	return getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/callback?state="+state+"&code=code-link", tok)
}

// viewRecorder keeps the last broadcast view per game.
type viewRecorder struct {
	mu   sync.Mutex
	last map[uuid.UUID]protocol.GameView
}

func (r *viewRecorder) BroadcastState(gameID uuid.UUID, _ uint64, view protocol.GameView) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.last == nil {
		r.last = map[uuid.UUID]protocol.GameView{}
	}
	r.last[gameID] = view
}

func (r *viewRecorder) seat(gameID uuid.UUID, playerID uuid.UUID) (protocol.PlayerView, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.last[gameID].Seats {
		if p.ID == playerID.String() {
			return p, true
		}
	}
	return protocol.PlayerView{}, false
}

// --- seats.user_id on every claim ---------------------------------

func TestSignedInClaimsWriteSeatUserID(t *testing.T) {
	s := newMyGamesStack(t)
	meta, err := s.lobby.Create("FNM")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	idTok := identityTokenFromCallback(t, s.srv, s.state)
	who := mustValidate(t, s.auth, idTok)

	// POST /join with an identity session.
	sess := decodeSession(t, postJSON(t, s.srv, "/join", idTok, map[string]string{"invite_token": meta.InviteToken}), http.StatusOK)
	uid, pending := seatRow(t, s, meta.ID, sess.PlayerID)
	if uid.String != who.UserID.String() {
		t.Errorf("POST /join: seats.user_id = %q, want %s", uid.String, who.UserID)
	}
	if pending.Valid {
		t.Errorf("POST /join: a linked seat also waits on pending_discord_id %q", pending.String)
	}

	// A guest seat has no user.
	guest := decodeSession(t, postJSON(t, s.srv, "/join", "", map[string]string{"invite_token": meta.InviteToken, "name": "Guest"}), http.StatusOK)
	if uid, _ := seatRow(t, s, meta.ID, guest.PlayerID); uid.Valid {
		t.Errorf("guest seat has user_id %q", uid.String)
	}

	// The OAuth invite flow.
	other, err := s.lobby.Create("Other")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	st, _, err := s.state.Start(other.ID, other.InviteToken)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	frag := fragmentParams(t, followOneRedirect(t, s.srv, "/auth/discord/callback?state="+st+"&code=code-2"))
	if frag.Get("user_id") != who.UserID.String() {
		t.Errorf("invite-flow fragment user_id = %q, want %s", frag.Get("user_id"), who.UserID)
	}
	pid := uuid.MustParse(frag.Get("player_id"))
	if uid, _ := seatRow(t, s, other.ID, pid); uid.String != who.UserID.String() {
		t.Errorf("OAuth claim: seats.user_id = %q, want %s", uid.String, who.UserID)
	}
}

func TestInviteLinkJoinSeatsASignedInPersonAsThemselves(t *testing.T) {
	s := newMyGamesStack(t)
	meta, _ := s.lobby.Create("FNM")
	idTok := identityTokenFromCallback(t, s.srv, s.state)
	who := mustValidate(t, s.auth, idTok)

	// The Join page posts a name; a signed-in person's seat takes the
	// Discord identity instead.
	sess := decodeSession(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/join", idTok,
		map[string]string{"invite_token": meta.InviteToken, "name": "Typed"}), http.StatusOK)
	if sess.Principal.UserID != who.UserID || sess.Principal.DiscordID != "discord-99" {
		t.Errorf("seat session = %+v, want the signed-in user", sess.Principal)
	}
	uid, _ := seatRow(t, s, meta.ID, sess.PlayerID)
	if uid.String != who.UserID.String() {
		t.Errorf("seats.user_id = %q, want %s", uid.String, who.UserID)
	}
	got, _ := s.lobby.Get(meta.ID)
	if seat, _ := findSeat(got.Players, sess.PlayerID); seat.DisplayName != "Alice" || seat.DiscordID != "discord-99" {
		t.Errorf("seat = %+v, want Alice's Discord identity", seat)
	}
}

func TestAPersonHoldsOneSeatPerTable(t *testing.T) {
	s := newMyGamesStack(t)
	meta, _ := s.lobby.Create("FNM")
	idTok := identityTokenFromCallback(t, s.srv, s.state)
	decodeSession(t, postJSON(t, s.srv, "/join", idTok, map[string]string{"invite_token": meta.InviteToken}), http.StatusOK)
	resp := postJSON(t, s.srv, "/join", idTok, map[string]string{"invite_token": meta.InviteToken})
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("second seat for one person: %d, want 409", resp.StatusCode)
	}
}

// --- pending seats link at sign-in ----------------------------------

func TestSignInLinksPendingSeats(t *testing.T) {
	s := newMyGamesStack(t)
	ctx := context.Background()

	// A live table where discord-99 sat before it had a users row.
	live, _ := s.lobby.Create("Live")
	_, livePID, err := s.lobby.JoinWithIdentity(live.ID, live.InviteToken, "", DiscordIdentity{ID: "discord-99", GlobalName: "Alice"})
	if err != nil {
		t.Fatalf("JoinWithIdentity: %v", err)
	}
	if uid, pending := seatRow(t, s, live.ID, livePID); uid.Valid || pending.String != "discord-99" {
		t.Fatalf("before sign-in: user_id %v pending %v, want NULL / discord-99", uid, pending)
	}

	// A finished game that exists only as rows, as an import leaves it.
	store := NewSQLStore(s.db)
	old := GameRecord{ID: uuid.New(), Name: "Last week", State: "ended", CreatedAt: time.Now().Add(-7 * 24 * time.Hour).UTC()}
	if err := store.CreateGame(ctx, old, nil); err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	oldPID := uuid.New()
	if err := store.ReplaceSeats(ctx, old.ID, []SeatRecord{
		{Seat: 0, PlayerID: oldPID, GuestName: "Alice", PendingDiscordID: "discord-99"},
		{Seat: 1, PlayerID: uuid.New(), GuestName: "Bob"},
	}); err != nil {
		t.Fatalf("ReplaceSeats: %v", err)
	}

	who := mustValidate(t, s.auth, identityTokenFromCallback(t, s.srv, s.state))
	for _, seat := range []struct{ game, player uuid.UUID }{{live.ID, livePID}, {old.ID, oldPID}} {
		uid, pending := seatRow(t, s, seat.game, seat.player)
		if uid.String != who.UserID.String() || pending.Valid {
			t.Errorf("after sign-in, game %s: user_id %v pending %v, want %s / NULL", seat.game, uid, pending, who.UserID)
		}
	}

	// The live table's memory was linked too: its next seat write
	// (a deck upload rewrites every seat row) keeps the link.
	if _, err := s.lobby.SetDeck(live.ID, livePID, "Deck", botDeck(5)); err != nil {
		t.Fatalf("SetDeck: %v", err)
	}
	if uid, pending := seatRow(t, s, live.ID, livePID); uid.String != who.UserID.String() || pending.Valid {
		t.Errorf("after a seat write: user_id %v pending %v; the link was undone", uid, pending)
	}

	// Idempotent: signing in again links nothing new and moves nothing.
	again := mustValidate(t, s.auth, identityTokenFromCallback(t, s.srv, s.state))
	if again.UserID != who.UserID {
		t.Fatalf("second sign-in is a different user")
	}
	if n, err := s.lobby.LinkPendingSeats("discord-99", who.UserID); err != nil || n != 0 {
		t.Errorf("LinkPendingSeats after linking = %d, %v; want 0, nil", n, err)
	}

	// And both games are now "my games", newest first.
	_, games, _ := getMyGames(t, s, identityTokenFromCallback(t, s.srv, s.state))
	if len(games) != 2 || games[0].ID != live.ID || games[1].ID != old.ID {
		t.Fatalf("my games = %+v, want Live then Last week", games)
	}
	if games[1].Rejoin != "" {
		t.Errorf("a game with no live table offers a rejoin: %q", games[1].Rejoin)
	}
	if len(games[1].Others) != 1 || games[1].Others[0].Name != "Bob" || games[1].Others[0].Seat != 1 {
		t.Errorf("Last week's other seats = %+v, want Bob at seat 1", games[1].Others)
	}
}

func TestLinkPendingSeatsIsIdempotentInBothStores(t *testing.T) {
	for name, store := range map[string]Store{
		"memory": NewMemoryStore(),
		"sql":    NewSQLStore(openTestDB(t, t.TempDir())),
	} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			var userID string
			if sq, ok := store.(*SQLStore); ok {
				u, err := users.NewSQLStore(sq.db, nil).UpsertFromDiscord(ctx, discord.User{ID: "d-1", Username: "a"}, "", "identify")
				if err != nil {
					t.Fatalf("UpsertFromDiscord: %v", err)
				}
				userID = u.ID.String()
			} else {
				userID = uuid.NewString()
			}
			g := GameRecord{ID: uuid.New(), Name: "g", State: "lobby", CreatedAt: time.Now().UTC()}
			if err := store.CreateGame(ctx, g, nil); err != nil {
				t.Fatalf("CreateGame: %v", err)
			}
			if err := store.ReplaceSeats(ctx, g.ID, []SeatRecord{
				{Seat: 0, PlayerID: uuid.New(), GuestName: "A", PendingDiscordID: "d-1"},
				{Seat: 1, PlayerID: uuid.New(), GuestName: "B", PendingDiscordID: "d-2"},
			}); err != nil {
				t.Fatalf("ReplaceSeats: %v", err)
			}
			if n, err := store.LinkPendingSeats(ctx, "d-1", userID); err != nil || n != 1 {
				t.Fatalf("first link = %d, %v; want 1", n, err)
			}
			if n, err := store.LinkPendingSeats(ctx, "d-1", userID); err != nil || n != 0 {
				t.Errorf("second link = %d, %v; want 0", n, err)
			}
			_, seats, err := store.LoadGame(ctx, g.ID)
			if err != nil {
				t.Fatalf("LoadGame: %v", err)
			}
			if seats[0].UserID != userID || seats[0].PendingDiscordID != "" {
				t.Errorf("linked seat = %+v", seats[0])
			}
			if seats[1].UserID != "" || seats[1].PendingDiscordID != "d-2" {
				t.Errorf("someone else's pending seat was touched: %+v", seats[1])
			}
			mine, err := store.SeatsOfUser(ctx, userID)
			if err != nil || len(mine) != 1 || mine[0].Seat != 0 || len(mine[0].Others) != 1 || mine[0].Others[0].Name != "B" {
				t.Errorf("SeatsOfUser = %+v, %v", mine, err)
			}
		})
	}
}

// --- GET /me/games --------------------------------------------------

func TestMyGamesNeedsASignedInUser(t *testing.T) {
	s := newMyGamesStack(t)
	meta, _ := s.lobby.Create("FNM")
	guest := decodeSession(t, postJSON(t, s.srv, "/join", "", map[string]string{"invite_token": meta.InviteToken, "name": "Guest"}), http.StatusOK)

	noUser, _, err := s.auth.Issue(context.Background(), auth.Principal{Role: auth.RoleIdentified, DiscordID: "d"}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	// #1154: an authenticated session that is not a person is 403, not
	// 401. Only the missing credential is 401 — the client clears the
	// session on any 401, so calling a valid admin session "expired"
	// logged the admin out of a page they were merely looking at.
	for name, tc := range map[string]struct {
		tok  string
		want int
	}{
		"no session":            {"", http.StatusUnauthorized},
		"guest player":          {guest.Token, http.StatusForbidden},
		"admin":                 {adminSession(t, s.auth), http.StatusForbidden},
		"identity with no user": {noUser, http.StatusForbidden},
	} {
		if status, _, _ := getMyGames(t, s, tc.tok); status != tc.want {
			t.Errorf("%s: GET /me/games = %d, want %d", name, status, tc.want)
		}
	}
}

func TestMyGamesListsSeatsNewestFirstWithoutInvites(t *testing.T) {
	s := newMyGamesStack(t)
	idTok := identityTokenFromCallback(t, s.srv, s.state)

	first, _ := s.lobby.Create("First")
	decodeSession(t, postJSON(t, s.srv, "/join", "", map[string]string{"invite_token": first.InviteToken, "name": "Bob"}), http.StatusOK)
	decodeSession(t, postJSON(t, s.srv, "/join", idTok, map[string]string{"invite_token": first.InviteToken}), http.StatusOK)
	time.Sleep(2 * time.Millisecond) // created_at is stored in ms
	second, _ := s.lobby.Create("Second")
	decodeSession(t, postJSON(t, s.srv, "/join", idTok, map[string]string{"invite_token": second.InviteToken}), http.StatusOK)
	notMine, _ := s.lobby.Create("Not mine")

	status, games, raw := getMyGames(t, s, idTok)
	if status != http.StatusOK {
		t.Fatalf("GET /me/games = %d", status)
	}
	if len(games) != 2 || games[0].ID != second.ID || games[1].ID != first.ID {
		t.Fatalf("games = %+v, want Second then First", games)
	}
	if g := games[1]; g.Seat != 1 || g.State != "lobby" || g.CreatedAt != first.CreatedAt.UnixMilli() ||
		len(g.Others) != 1 || g.Others[0].Name != "Bob" || g.Others[0].Seat != 0 {
		t.Errorf("First = %+v", g)
	}
	if games[0].Rejoin != "/me/games/"+second.ID.String()+"/session" {
		t.Errorf("rejoin = %q", games[0].Rejoin)
	}
	for _, m := range []GameMeta{first, second, notMine} {
		if strings.Contains(raw, m.InviteToken) || strings.Contains(raw, m.SpectatorInvite) {
			t.Fatalf("GET /me/games leaked an invite token: %s", raw)
		}
	}
	if strings.Contains(raw, notMine.ID.String()) {
		t.Error("GET /me/games lists a table the user never sat at")
	}

	// Archived tables stay listed, as history, with no way back in.
	if _, err := s.lobby.SetArchived(first.ID, true); err != nil {
		t.Fatalf("SetArchived: %v", err)
	}
	_, games, _ = getMyGames(t, s, idTok)
	if len(games) != 2 || games[1].ArchivedAt == nil || games[1].Rejoin != "" {
		t.Errorf("archived game = %+v, want archived_at set and no rejoin", games[1])
	}
}

// --- seat reclaim by user -------------------------------------------

func TestUserReclaimsTheirSeatInAStartedGame(t *testing.T) {
	s := newMyGamesStack(t)
	meta, _ := s.lobby.Create("FNM")
	idTok := identityTokenFromCallback(t, s.srv, s.state)
	who := mustValidate(t, s.auth, idTok)
	mine := decodeSession(t, postJSON(t, s.srv, "/join", idTok, map[string]string{"invite_token": meta.InviteToken}), http.StatusOK)
	bob := decodeSession(t, postJSON(t, s.srv, "/join", "", map[string]string{"invite_token": meta.InviteToken, "name": "Bob"}), http.StatusOK)
	for _, p := range []uuid.UUID{mine.PlayerID, bob.PlayerID} {
		if _, err := s.lobby.SetDeck(meta.ID, p, "Deck", botDeck(20)); err != nil {
			t.Fatalf("SetDeck: %v", err)
		}
	}
	if _, err := s.lobby.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// The seat's own session is gone; the identity session gets it back.
	back := decodeSession(t, postJSON(t, s.srv, "/me/games/"+meta.ID.String()+"/session", idTok, nil), http.StatusOK)
	if back.PlayerID != mine.PlayerID || back.Principal.Role != auth.RolePlayer || back.Principal.UserID != who.UserID {
		t.Errorf("reclaimed session = %+v, want the user's own seat", back.Principal)
	}
	if back.Principal.DiscordID != "discord-99" {
		t.Errorf("reclaimed session lost the seat's Discord identity: %+v", back.Principal)
	}
	if back.Game == nil || back.Game.InviteToken != "" || back.Game.SpectatorInvite != "" {
		t.Error("reclaim by user leaked a table invite")
	}
	if p := mustValidate(t, s.auth, back.Token); p.GameID != meta.ID || p.PlayerID != mine.PlayerID {
		t.Errorf("token binds %v/%v", p.GameID, p.PlayerID)
	}

	// A player session with the user works too (another table's seat).
	decodeSession(t, postJSON(t, s.srv, "/me/games/"+meta.ID.String()+"/session", back.Token, nil), http.StatusOK)

	// Guests and other people cannot.
	for name, tc := range map[string]struct {
		tok  string
		want int
	}{
		"guest":   {bob.Token, http.StatusForbidden},
		"no auth": {"", http.StatusUnauthorized},
	} {
		resp := postJSON(t, s.srv, "/me/games/"+meta.ID.String()+"/session", tc.tok, nil)
		resp.Body.Close()
		if resp.StatusCode != tc.want {
			t.Errorf("%s: %d, want %d", name, resp.StatusCode, tc.want)
		}
	}
	s.stub.setUser(`{"id":"discord-42","username":"carol","global_name":"Carol","avatar":"c"}`)
	carol := identityTokenFromCallback(t, s.srv, s.state)
	resp := postJSON(t, s.srv, "/me/games/"+meta.ID.String()+"/session", carol, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("a user with no seat there: %d, want 403", resp.StatusCode)
	}
	resp = postJSON(t, s.srv, "/me/games/"+uuid.NewString()+"/session", idTok, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown game: %d, want 404", resp.StatusCode)
	}
	if _, err := s.lobby.SetArchived(meta.ID, true); err != nil {
		t.Fatalf("SetArchived: %v", err)
	}
	resp = postJSON(t, s.srv, "/me/games/"+meta.ID.String()+"/session", idTok, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("archived table: %d, want 409", resp.StatusCode)
	}
}

func TestSeatUserSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	d := openTestDB(t, dir)
	u, err := users.NewSQLStore(d, nil).UpsertFromDiscord(ctx, discord.User{ID: "d-7", Username: "al"}, "", "identify")
	if err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	l := NewLobbyWithStore(ws.NewRoomManager(quietLogger(), dir), NewSQLStore(d))
	meta, _ := l.Create("FNM")
	_, pid, err := l.JoinAs(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "d-7", Username: "al"}, u.ID)
	if err != nil {
		t.Fatalf("JoinAs: %v", err)
	}

	l2 := NewLobbyWithStore(ws.NewRoomManager(quietLogger(), dir), NewSQLStore(openTestDB(t, dir)))
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games", n)
	}
	_, seat, err := l2.ReclaimByUser(meta.ID, u.ID)
	if err != nil || seat.PlayerID != pid {
		t.Fatalf("ReclaimByUser after restart = %v, %v; want seat %v", seat.PlayerID, err, pid)
	}
}

// --- GET /auth/discord/link -----------------------------------------

// The headline for the open question in ADR 0051: a guest at a live
// table signs in with Discord and becomes that seat's user, without
// leaving the game, and the table sees the new name at once.
func TestGuestLinksDiscordMidGame(t *testing.T) {
	s := newMyGamesStack(t)
	rec := &viewRecorder{}
	s.lobby.SetStateBroadcaster(rec)
	meta, alice, _ := startTwoSeatGame(t, s.lobby, "FNM")
	aliceTok := playerToken(t, s.auth, meta.ID, alice, "Alice")
	s.stub.setUser(`{"id":"discord-77","username":"al","global_name":"Alice on Discord","avatar":"h77"}`)

	resp := linkDiscord(t, s, meta.ID, aliceTok)
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("callback: %d", resp.StatusCode)
	}
	frag := fragmentParams(t, resp.Header.Get("Location"))
	if frag.Get("game") != meta.ID.String() || frag.Get("player_id") != alice.String() || frag.Get("user_id") == "" {
		t.Fatalf("fragment = %v, want the same seat and a user", frag)
	}
	p := mustValidate(t, s.auth, frag.Get("token"))
	if p.Role != auth.RolePlayer || p.PlayerID != alice || p.UserID.String() != frag.Get("user_id") || p.DiscordID != "discord-77" {
		t.Errorf("new session = %+v", p)
	}

	// The seat is the user's, in the row and in memory.
	uid, pending := seatRow(t, s, meta.ID, alice)
	if uid.String != p.UserID.String() || pending.Valid {
		t.Errorf("seat row: user_id %v pending %v", uid, pending)
	}
	got, _ := s.lobby.Get(meta.ID)
	seat, _ := findSeat(got.Players, alice)
	if seat.DisplayName != "Alice on Discord" || seat.DiscordAvatarHash != "h77" || seat.Name != "Alice" {
		t.Errorf("seat = %+v", seat)
	}
	// Opponents see it at once: the link broadcast a view carrying it.
	if pv, ok := rec.seat(meta.ID, alice); !ok || pv.DisplayName != "Alice on Discord" || pv.DiscordID != "discord-77" {
		t.Errorf("broadcast view of the seat = %+v (found %v)", pv, ok)
	}
	// And it is now in the user's games, with the way back in.
	_, games, _ := getMyGames(t, s, frag.Get("token"))
	if len(games) != 1 || games[0].ID != meta.ID || games[0].Rejoin == "" {
		t.Errorf("my games after linking = %+v", games)
	}
}

func TestDiscordLinkSwapsTheSeatsIdentity(t *testing.T) {
	s := newMyGamesStack(t)
	meta, alice, _ := startTwoSeatGame(t, s.lobby, "FNM")
	resp := linkDiscord(t, s, meta.ID, playerToken(t, s.auth, meta.ID, alice, "Alice"))
	resp.Body.Close()
	firstTok := fragmentParams(t, resp.Header.Get("Location")).Get("token")
	first := mustValidate(t, s.auth, firstTok)

	s.stub.setUser(`{"id":"discord-88","username":"al2","global_name":"Alice Alt","avatar":"h88"}`)
	resp = linkDiscord(t, s, meta.ID, firstTok)
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("swap callback: %d", resp.StatusCode)
	}
	second := mustValidate(t, s.auth, fragmentParams(t, resp.Header.Get("Location")).Get("token"))
	if second.UserID == first.UserID {
		t.Fatal("swap kept the old user")
	}
	if uid, _ := seatRow(t, s, meta.ID, alice); uid.String != second.UserID.String() {
		t.Errorf("seat user after swap = %q, want %s", uid.String, second.UserID)
	}
	if _, games, _ := getMyGames(t, s, firstTok); len(games) != 0 {
		t.Errorf("the old account still lists the seat: %+v", games)
	}
}

func TestDiscordLinkIsBoundToTheSeatsOwnBrowser(t *testing.T) {
	s := newMyGamesStack(t)
	meta, alice, bob := startTwoSeatGame(t, s.lobby, "FNM")
	aliceTok := playerToken(t, s.auth, meta.ID, alice, "Alice")
	bobTok := playerToken(t, s.auth, meta.ID, bob, "Bob")

	start := func(tok, game string) (int, string) {
		resp := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/link?game="+game, tok)
		defer resp.Body.Close()
		loc, _ := url.Parse(resp.Header.Get("Location"))
		if loc == nil {
			return resp.StatusCode, ""
		}
		return resp.StatusCode, loc.Query().Get("state")
	}

	// Alice starts a link; the consent URL is finished in a browser
	// with no session, then in Bob's. Neither touches her seat, and
	// neither costs a token exchange.
	for name, finisher := range map[string]string{"no session": "", "another seat": bobTok} {
		status, state := start(aliceTok, meta.ID.String())
		if status != http.StatusFound || state == "" {
			t.Fatalf("start: %d", status)
		}
		s.stub.tokenWasCalled()
		resp := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/callback?state="+state+"&code=c", finisher)
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s finishing Alice's link: %d, want 403", name, resp.StatusCode)
		}
		if s.stub.tokenWasCalled() {
			t.Errorf("%s: the code was exchanged before the session was checked", name)
		}
	}
	if uid, _ := seatRow(t, s, meta.ID, alice); uid.Valid {
		t.Errorf("Alice's seat was linked to %q by someone else", uid.String)
	}
	got, _ := s.lobby.Get(meta.ID)
	if seat, _ := findSeat(got.Players, alice); seat.DiscordID != "" {
		t.Errorf("Alice's seat took a Discord identity: %+v", seat)
	}

	// The start itself refuses a session for another table, an
	// admin, and no session at all.
	if status, _ := start(aliceTok, uuid.NewString()); status != http.StatusConflict {
		t.Errorf("?game= for another table: %d, want 409", status)
	}
	if status, _ := start(adminSession(t, s.auth), meta.ID.String()); status != http.StatusForbidden {
		t.Errorf("admin: %d, want 403", status)
	}
	if status, _ := start("", meta.ID.String()); status != http.StatusUnauthorized {
		t.Errorf("no session: %d, want 401", status)
	}
}

func TestDiscordLinkRefusesASecondSeatForOnePerson(t *testing.T) {
	s := newMyGamesStack(t)
	meta, alice, bob := startTwoSeatGame(t, s.lobby, "FNM")
	resp := linkDiscord(t, s, meta.ID, playerToken(t, s.auth, meta.ID, alice, "Alice"))
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("Alice's link: %d", resp.StatusCode)
	}
	// Same Discord account (the stub's), Bob's seat.
	resp = linkDiscord(t, s, meta.ID, playerToken(t, s.auth, meta.ID, bob, "Bob"))
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("linking one account to a second seat: %d, want 409", resp.StatusCode)
	}
	if uid, _ := seatRow(t, s, meta.ID, bob); uid.Valid {
		t.Errorf("Bob's seat was linked: %q", uid.String)
	}
}
