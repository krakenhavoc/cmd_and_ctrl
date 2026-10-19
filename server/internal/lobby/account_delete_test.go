package lobby

// Tests for DELETE /me/account (ADR 0149, #2983): who may call it, the
// typed confirmation, the active-table refusal, and that a finished
// game still restores and renders with the person's seat anonymised.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/accountdelete"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

type deletionStack struct {
	srv     *httptest.Server
	lobby   *Lobby
	mgr     *ws.RoomManager
	stub    *discordStub
	state   *discord.StateStore
	users   *users.SQLStore
	auth    auth.Authenticator
	db      *db.DB
	dir     string
	evictor *recordingEvictor
}

func newDeletionStack(t *testing.T) deletionStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	dir := t.TempDir()
	d := openTestDB(t, dir)
	us := users.NewSQLStore(d, nil)
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
	mgr := ws.NewRoomManager(quietLogger(), dir)
	srv, l, stub, state := newDiscordTestStackWith(t, func(c *Config) {
		c.Lobby = NewLobbyWithStore(mgr, NewSQLStore(d))
		c.Auth = a
		c.Users = us
		c.Revocations = rev
		c.SessionEvictor = ev
		c.Accounts = accountdelete.NewSQLStore(d)
	})
	return deletionStack{srv: srv, lobby: l, mgr: mgr, stub: stub, state: state, users: us, auth: a, db: d, dir: dir, evictor: ev}
}

// signIn walks a real Discord sign-in as the account the stub reports.
func (s deletionStack) signIn(t *testing.T, id, name, avatar string) (string, uuid.UUID) {
	t.Helper()
	s.stub.setUser(`{"id":"` + id + `","username":"` + strings.ToLower(name) + `","global_name":"` + name + `","avatar":"` + avatar + `"}`)
	tok := identityTokenFromCallback(t, s.srv, s.state)
	return tok, mustValidate(t, s.auth, tok).UserID
}

func (s deletionStack) deleteAccount(t *testing.T, tok, confirm string) *http.Response {
	t.Helper()
	return doJSON(t, s.srv, http.MethodDelete, "/me/account", tok, map[string]string{"confirm": confirm})
}

// sitAt seats a signed-in person (and a guest) at a fresh table and
// starts it. The deck names are neutral so no file can match a person's
// name through them.
func (s deletionStack) startTableWith(t *testing.T, userID uuid.UUID, ident DiscordIdentity, table string) (GameMeta, uuid.UUID, uuid.UUID) {
	t.Helper()
	meta, err := s.lobby.Create(table)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, signed, err := s.lobby.JoinAs(meta.ID, meta.InviteToken, "", ident, userID)
	if err != nil {
		t.Fatalf("JoinAs: %v", err)
	}
	_, guest, err := s.lobby.Join(meta.ID, meta.InviteToken, "Guesty")
	if err != nil {
		t.Fatalf("Join guest: %v", err)
	}
	for _, p := range []uuid.UUID{signed, guest} {
		if _, err := s.lobby.SetDeck(meta.ID, p, "Pile", botDeck(20)); err != nil {
			t.Fatalf("SetDeck: %v", err)
		}
	}
	if _, err := s.lobby.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return meta, signed, guest
}

func (s deletionStack) concede(t *testing.T, gameID, player uuid.UUID) {
	t.Helper()
	room := s.lobby.RoomOf(gameID)
	if _, _, err := room.Apply(player, func() error { return room.Game.Concede(player) }); err != nil {
		t.Fatalf("concede: %v", err)
	}
	if room.Game.CurrentState() != game.StateEnded {
		t.Fatalf("setup: state = %q, want ended", room.Game.CurrentState())
	}
}

func aliceIdentity() DiscordIdentity {
	return DiscordIdentity{ID: "discord-99", Username: "alice", GlobalName: "Alice", AvatarHash: "avhash"}
}

func TestDeleteAccountRouteAuth(t *testing.T) {
	s := newDeletionStack(t)
	tok, _ := s.signIn(t, "discord-99", "Alice", "avhash")

	// Anonymous: 401, and nothing happens.
	resp := s.deleteAccount(t, "", DeleteAccountPhrase)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("anonymous = %d, want 401", resp.StatusCode)
	}

	// A guest seat holds no account: a clear refusal, not a 401 (authFetch
	// signs the browser out on a 401).
	meta, _ := s.lobby.Create("Guest table")
	_, pid, err := s.lobby.Join(meta.ID, meta.InviteToken, "Guesty")
	if err != nil {
		t.Fatal(err)
	}
	guestTok := playerToken(t, s.auth, meta.ID, pid, "Guesty")
	resp = s.deleteAccount(t, guestTok, DeleteAccountPhrase)
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusForbidden || !strings.Contains(body, "no account to delete") {
		t.Errorf("guest = %d %s, want 403 explaining there is no account", resp.StatusCode, body)
	}

	// The shared admin token is not a person either.
	adminTok, _, err := s.auth.Issue(context.Background(), auth.Principal{Role: auth.RoleAdmin, AdminID: uuid.New()}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	resp = s.deleteAccount(t, adminTok, DeleteAccountPhrase)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("admin token = %d, want 403", resp.StatusCode)
	}

	// The confirmation is checked by the server.
	for _, bad := range []string{"", "yes", "delete"} {
		resp = s.deleteAccount(t, tok, bad)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("confirm %q = %d, want 400", bad, resp.StatusCode)
		}
	}
	if st, _ := statusOf(t, s.srv, tok); st != http.StatusOK {
		t.Errorf("a refused deletion must leave the session alone, /me = %d", st)
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return sb.String()
}

func TestDeleteAccountOnlyDeletesTheCaller(t *testing.T) {
	s := newDeletionStack(t)
	aliceTok, alice := s.signIn(t, "discord-99", "Alice", "avhash")
	beaTok, bea := s.signIn(t, "discord-77", "Bea", "beahash")
	exec := func(q string, args ...any) {
		if _, err := s.db.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	for _, u := range []uuid.UUID{alice, bea} {
		exec(`INSERT INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at)
			VALUES (?, ?, 'D', 'text', 'x', '[]', 1, 1, 1)`, uuid.NewString(), u.String())
	}

	resp := s.deleteAccount(t, aliceTok, DeleteAccountPhrase)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete = %d, want 204", resp.StatusCode)
	}

	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM decks WHERE owner_id = ?`, alice.String()).Scan(&n)
	if n != 0 {
		t.Errorf("Alice kept %d decks", n)
	}
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM decks WHERE owner_id = ?`, bea.String()).Scan(&n)
	if n != 1 {
		t.Errorf("Bea has %d decks, want 1: the deletion reached someone else", n)
	}
	if st, _ := statusOf(t, s.srv, beaTok); st != http.StatusOK {
		t.Errorf("Bea's session = %d, want 200", st)
	}
	if u, err := s.users.Get(context.Background(), bea); err != nil || u.DisplayName != "Bea" {
		t.Errorf("Bea's user row = %+v, %v", u, err)
	}
}

func TestDeleteAccountRevokesSessionsAndClearsTheCookie(t *testing.T) {
	s := newDeletionStack(t)
	tok, uid := s.signIn(t, "discord-99", "Alice", "avhash")
	// A second browser holds its own session.
	tok2 := identityTokenFromCallback(t, s.srv, s.state)

	resp := s.deleteAccount(t, tok, DeleteAccountPhrase)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete = %d", resp.StatusCode)
	}
	cleared := false
	for _, c := range resp.Cookies() {
		if c.Name == auth.SessionCookie && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Error("the session cookie was not cleared")
	}
	for name, tk := range map[string]string{"the deleting browser": tok, "another browser": tok2} {
		if st, _ := statusOf(t, s.srv, tk); st != http.StatusUnauthorized {
			t.Errorf("%s: /me = %d after the deletion, want 401", name, st)
		}
	}
	if len(s.evictor.calls) == 0 || s.evictor.calls[0].user != uid {
		t.Errorf("open sockets were not evicted: %+v", s.evictor.calls)
	}

	// The refusal survives a restart: a fresh Revocations over the same
	// database still knows the watermark, because the tombstone row kept
	// it. Without that, the token would work again after the next deploy.
	rev2, err := users.NewRevocations(context.Background(), s.users)
	if err != nil {
		t.Fatal(err)
	}
	p := mustValidate(t, h2(t), tok) // signature-only validation
	if !rev2.Revoked(uid, p.IssuedAt) {
		t.Error("after a restart the pre-deletion token is accepted again")
	}

	// Signing in with the same Discord account afterwards is a new account.
	tokNew, uidNew := s.signIn(t, "discord-99", "Alice", "avhash")
	if uidNew == uid {
		t.Error("signing in again resurrected the deleted account")
	}
	if st, _ := statusOf(t, s.srv, tokNew); st != http.StatusOK {
		t.Errorf("the fresh account's session = %d, want 200", st)
	}
}

// h2 is the HMAC authenticator without the revocation wrapper, to read
// a token's claims after it was revoked.
func h2(t *testing.T) auth.Authenticator {
	t.Helper()
	h, err := auth.NewHMACAuthenticator([]byte("lobby-test-session-key-0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestDeleteAccountRefusedWhileSeatedAtAnUnendedTable(t *testing.T) {
	s := newDeletionStack(t)
	tok, uid := s.signIn(t, "discord-99", "Alice", "avhash")
	meta, signed, _ := s.startTableWith(t, uid, aliceIdentity(), "Live pod")

	resp := s.deleteAccount(t, tok, DeleteAccountPhrase)
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusConflict || !strings.Contains(body, "Live pod") || !strings.Contains(body, "has not ended") {
		t.Fatalf("delete at a live table = %d %s, want 409 naming the table", resp.StatusCode, body)
	}
	if st, _ := statusOf(t, s.srv, tok); st != http.StatusOK {
		t.Errorf("a refused deletion revoked the session: /me = %d", st)
	}
	if u, err := s.users.Get(context.Background(), uid); err != nil || u.DisplayName != "Alice" {
		t.Errorf("a refused deletion changed the user: %+v %v", u, err)
	}

	// Once the table ends the same request goes through.
	s.concede(t, meta.ID, signed)
	resp = s.deleteAccount(t, tok, DeleteAccountPhrase)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete after the table ended = %d, want 204", resp.StatusCode)
	}
}

func TestDeleteAccountRefusedWhileSeatedInALobby(t *testing.T) {
	s := newDeletionStack(t)
	tok, uid := s.signIn(t, "discord-99", "Alice", "avhash")
	meta, _ := s.lobby.Create("Waiting room")
	if _, _, err := s.lobby.JoinAs(meta.ID, meta.InviteToken, "", aliceIdentity(), uid); err != nil {
		t.Fatal(err)
	}
	resp := s.deleteAccount(t, tok, DeleteAccountPhrase)
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusConflict || !strings.Contains(body, "has not started") {
		t.Errorf("delete from a lobby = %d %s, want 409", resp.StatusCode, body)
	}
}

// TestDeletedPlayersPastGameStillRestoresAndRenders is the check that
// matters for everyone else at the table: after the deletion the
// finished game is still there, with the same players and the same
// result, and the deleted seat reads "Deleted player" everywhere it can
// be read: the lobby, the live room, the files on disk and a restart.
func TestDeletedPlayersPastGameStillRestoresAndRenders(t *testing.T) {
	s := newDeletionStack(t)
	tok, uid := s.signIn(t, "discord-99", "Alice", "avhash")
	meta, signed, guest := s.startTableWith(t, uid, aliceIdentity(), "Friday pod")
	s.concede(t, meta.ID, signed)

	// Make the files carry the identity first, so the test can tell a
	// scrub from a file that never held it.
	mustContain := func(when string) {
		t.Helper()
		found := false
		for _, f := range s.gameFiles(meta.ID) {
			raw, err := os.ReadFile(f)
			if err == nil && strings.Contains(string(raw), "avhash") {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s: no game file carries the Discord avatar hash, so the scrub below proves nothing", when)
		}
	}
	mustContain("before the deletion")

	resp := s.deleteAccount(t, tok, DeleteAccountPhrase)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete = %d", resp.StatusCode)
	}

	// The lobby's copy.
	got, err := s.lobby.Get(meta.ID)
	if err != nil {
		t.Fatalf("the finished game is gone from the lobby: %v", err)
	}
	for _, p := range got.Players {
		if p.PlayerID == signed {
			if p.Name != accountdelete.DeletedPlayerName || p.DiscordID != "" || p.DiscordAvatarHash != "" || p.DisplayName != "" {
				t.Errorf("deleted seat in the lobby = %+v", p)
			}
		}
		if p.PlayerID == guest && p.Name != "Guesty" {
			t.Errorf("the other player's seat changed: %+v", p)
		}
	}
	if got.State != string(game.StateEnded) {
		t.Errorf("state = %q, want ended", got.State)
	}

	// The live room.
	view, _, err := s.lobby.RoomOf(meta.ID).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	for _, seat := range view.Seats {
		if seat.ID == signed.String() && (seat.Name != accountdelete.DeletedPlayerName || seat.DiscordID != "" || seat.DiscordAvatarHash != "") {
			t.Errorf("deleted seat in the live view = %+v", seat)
		}
	}

	// The files on disk: nothing that names the person is left, and the
	// other player's name is.
	for _, f := range s.gameFiles(meta.ID) {
		raw, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		text := string(raw)
		for _, secret := range []string{"avhash", "discord-99", `"Alice"`} {
			if strings.Contains(text, secret) {
				t.Errorf("%s still contains %q", filepath.Base(filepath.Dir(f))+"/"+filepath.Base(f), secret)
			}
		}
		if !strings.Contains(text, "Guesty") {
			t.Errorf("%s lost the other player's name", filepath.Base(f))
		}
		if !strings.Contains(text, accountdelete.DeletedPlayerName) {
			t.Errorf("%s does not carry %q", filepath.Base(f), accountdelete.DeletedPlayerName)
		}
	}

	// A restart brings the finished table back with the same two seats.
	l2, _ := newDurableLobby(t, s.dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	back, err := l2.Get(meta.ID)
	if err != nil {
		t.Fatalf("the finished game did not restore: %v", err)
	}
	names := map[string]bool{}
	for _, p := range back.Players {
		names[p.Name] = true
	}
	if len(back.Players) != 2 || !names["Guesty"] || !names[accountdelete.DeletedPlayerName] || names["Alice"] {
		t.Errorf("restored seats = %+v", back.Players)
	}
	rv, _, err := l2.RoomOf(meta.ID).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if rv.State != string(game.StateEnded) {
		t.Errorf("restored view state = %q, want ended", rv.State)
	}
	for _, seat := range rv.Seats {
		if seat.Name == "Alice" || seat.DiscordID != "" {
			t.Errorf("restored view still names the person: %+v", seat)
		}
	}
}

// A game that ended over the WebSocket reads "active" in the lobby and in
// the games row until watchEnd catches up. A person whose table is over
// must not be told it has not ended.
func TestDeleteAccountAfterATableEndedButTheRowStillSaysActive(t *testing.T) {
	s := newDeletionStack(t)
	tok, uid := s.signIn(t, "discord-99", "Alice", "avhash")
	meta, signed, _ := s.startTableWith(t, uid, aliceIdentity(), "Just ended")
	s.concede(t, meta.ID, signed)

	// watchEnd writes the end to the lobby entry and the games row in
	// its own goroutine. Put the stale state back only after it has
	// finished (#3986): a write-back that races it can be overwritten
	// in memory but not in the row, or the reverse, and then the row
	// says "active" while the entry says "ended". endedAt only climbs
	// from nil, so it is a monotone latch for "watchEnd is done".
	deadline := time.Now().Add(10 * time.Second)
	for {
		s.lobby.mu.Lock()
		done := s.lobby.games[meta.ID].endedAt != nil
		s.lobby.mu.Unlock()
		if done {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("setup: watchEnd never recorded the end of the conceded table")
		}
		time.Sleep(time.Millisecond)
	}

	// Put back the state watchEnd had written. watchEnd has returned, so
	// nothing writes the row again until the deletion check syncs it.
	s.lobby.mu.Lock()
	e := s.lobby.games[meta.ID]
	e.meta.State = string(game.StateActive)
	e.endedAt = nil
	_, err := s.db.Exec(`UPDATE games SET state = 'active', ended_at = NULL WHERE id = ?`, meta.ID.String())
	s.lobby.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}

	resp := s.deleteAccount(t, tok, DeleteAccountPhrase)
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete = %d %s, want 204: the table is over", resp.StatusCode, body)
	}
}

func (s deletionStack) gameFiles(id uuid.UUID) []string {
	n := id.String()
	return []string{
		filepath.Join(s.dir, "games", n+".json"),
		filepath.Join(s.dir, "replays", n+".jsonl"),
		filepath.Join(s.dir, "restore", n+".json"),
	}
}
