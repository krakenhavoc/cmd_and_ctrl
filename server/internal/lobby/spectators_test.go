package lobby

// spectators_test.go: reporting, removing and blocking a spectator
// (#3113).

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// watcher is one spectator session with its socket open.
type watcher struct {
	sess sessionResponse
	conn *websocket.Conn
}

func (s *seatRemoveStack) spectate(t *testing.T, meta GameMeta, name string) watcher {
	t.Helper()
	sess := decodeSession(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: meta.SpectatorInvite, Name: name}), http.StatusOK)
	if sess.Principal.SpectatorID == uuid.Nil {
		t.Fatal("a spectator session carries no SpectatorID")
	}
	conn, err := dialAdmitted(s.srv, sess.Token)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return watcher{sess: sess, conn: conn}
}

func (s *seatRemoveStack) specsPath(gameID uuid.UUID, rest string) string {
	return "/games/" + gameID.String() + "/spectators" + rest
}

func (s *seatRemoveStack) listSpectators(t *testing.T, gameID uuid.UUID, token string, want int) []spectatorLine {
	t.Helper()
	resp := doJSON(t, s.srv, http.MethodGet, s.specsPath(gameID, ""), token, nil)
	defer resp.Body.Close()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET spectators = %d %s, want %d", resp.StatusCode, body, want)
	}
	if want != http.StatusOK {
		return nil
	}
	var out spectatorsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out.Spectators
}

func (s *seatRemoveStack) removeSpectator(t *testing.T, gameID, specID uuid.UUID, query, token string) *http.Response {
	t.Helper()
	return doJSON(t, s.srv, http.MethodDelete, s.specsPath(gameID, "/"+specID.String()+query), token, nil)
}

func (s *seatRemoveStack) report(t *testing.T, gameID uuid.UUID, token string, body reportRequest) *http.Response {
	t.Helper()
	return doJSON(t, s.srv, http.MethodPost, "/games/"+gameID.String()+"/reports", token, body)
}

// readsRemovedClose reads conn until it closes and checks it closed the
// way a removed seat's does.
func readsRemovedClose(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		_, _, err := conn.ReadMessage()
		if err == nil {
			continue // a broadcast that raced the close
		}
		ce, ok := err.(*websocket.CloseError)
		if !ok {
			t.Fatalf("the socket ended with %v, want a close frame", err)
		}
		if ce.Code != websocket.CloseNormalClosure || ce.Text != ws.SeatRemovedReason {
			t.Fatalf("close = %d %q, want 1000 %q", ce.Code, ce.Text, ws.SeatRemovedReason)
		}
		return
	}
}

func TestSpectatorListShowsWhoIsWatchingToTheTableOnly(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, _ := s.table(t)
	w := s.spectate(t, meta, "Watcher")
	w2 := s.spectate(t, meta, "")
	// Two tabs of the first session are still one spectator.
	extra, err := dialAdmitted(s.srv, w.sess.Token)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = extra.Close() })

	got := s.listSpectators(t, meta.ID, alice.Token, http.StatusOK)
	if len(got) != 2 {
		t.Fatalf("spectators = %+v, want two sessions", got)
	}
	names := map[string]string{}
	for _, g := range got {
		names[g.ID] = g.Name
	}
	if names[w.sess.Principal.SpectatorID.String()] != "Watcher" || names[w2.sess.Principal.SpectatorID.String()] != "Spectator" {
		t.Errorf("names = %v, want Watcher and the placeholder Spectator", names)
	}

	// A seat sees it; a spectator, a stranger and a bystander do not.
	s.listSpectators(t, meta.ID, "", http.StatusUnauthorized)
	s.listSpectators(t, meta.ID, w.sess.Token, http.StatusForbidden)
	s.listSpectators(t, meta.ID, identifiedToken(t, s.a, uuid.New()), http.StatusForbidden)
	s.listSpectators(t, meta.ID, adminToken(t, s.srv), http.StatusOK)
}

func TestAnAdminSocketIsNotASpectator(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, _ := s.table(t)
	conn, err := dialAdmitted(s.srv, adminToken(t, s.srv)+"&game="+meta.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if got := s.listSpectators(t, meta.ID, alice.Token, http.StatusOK); len(got) != 0 {
		t.Errorf("an admin's socket was listed as a spectator: %+v", got)
	}
}

func TestSpectatorCanBeReported(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, bob := s.table(t)
	w := s.spectate(t, meta, "Watcher")
	spec := w.sess.Principal.SpectatorID.String()

	// The same refusals as a seat's report, plus the new ones.
	s.wantStatus(t, s.report(t, meta.ID, alice.Token, reportRequest{SpectatorID: spec, Reason: "  "}), http.StatusBadRequest)
	s.wantStatus(t, s.report(t, meta.ID, alice.Token, reportRequest{SpectatorID: spec, PlayerID: bob.PlayerID.String(), Reason: "x"}), http.StatusBadRequest)
	s.wantStatus(t, s.report(t, meta.ID, alice.Token, reportRequest{SpectatorID: "nonsense", Reason: "x"}), http.StatusBadRequest)
	s.wantStatus(t, s.report(t, meta.ID, alice.Token, reportRequest{SpectatorID: uuid.NewString(), Reason: "x"}), http.StatusNotFound)
	s.wantStatus(t, s.report(t, meta.ID, w.sess.Token, reportRequest{SpectatorID: spec, Reason: "x"}), http.StatusForbidden)
	s.wantStatus(t, s.report(t, meta.ID, "", reportRequest{SpectatorID: spec, Reason: "x"}), http.StatusUnauthorized)

	s.wantStatus(t, s.report(t, meta.ID, alice.Token, reportRequest{SpectatorID: spec, Reason: "slurs in chat"}), http.StatusCreated)
	reports, err := s.l.Reports()
	if err != nil || len(reports) != 1 {
		t.Fatalf("reports = %+v, %v", reports, err)
	}
	r := reports[0]
	if r.ReportedName != "Watcher (spectator)" || r.ReportedPlayerID != w.sess.Principal.SpectatorID ||
		r.ReporterName != "Alice" || r.Reason != "slurs in chat" || r.ReportedUserID != "" {
		t.Errorf("report = %+v", r)
	}

	// A spectator who has gone cannot be reported afterwards.
	_ = w.conn.Close()
	deadline := time.Now().Add(3 * time.Second)
	for len(s.hub.Spectators(meta.ID)) != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	s.wantStatus(t, s.report(t, meta.ID, alice.Token, reportRequest{SpectatorID: spec, Reason: "again"}), http.StatusNotFound)
}

func TestRemoveSpectatorAuthMatrixAndSocketClose(t *testing.T) {
	s := newSeatRemoveStack(t)
	creator := uuid.New()
	meta, err := s.l.CreateBy("Spectator table", creator)
	if err != nil {
		t.Fatal(err)
	}
	alice := joinAs(t, s.srv, meta, "Alice")
	carol := joinAs(t, s.srv, meta, "Carol")
	w := s.spectate(t, meta, "Watcher")
	other := s.spectate(t, meta, "Other")
	id := w.sess.Principal.SpectatorID

	s.wantStatus(t, s.removeSpectator(t, meta.ID, id, "", ""), http.StatusUnauthorized)
	s.wantStatus(t, s.removeSpectator(t, meta.ID, id, "", carol.Token), http.StatusForbidden)
	s.wantStatus(t, s.removeSpectator(t, meta.ID, id, "", other.sess.Token), http.StatusForbidden)
	s.wantStatus(t, s.removeSpectator(t, meta.ID, id, "", w.sess.Token), http.StatusForbidden)
	s.wantStatus(t, s.removeSpectator(t, meta.ID, id, "", identifiedToken(t, s.a, uuid.New())), http.StatusForbidden)
	s.wantStatus(t, s.removeSpectator(t, meta.ID, uuid.New(), "", alice.Token), http.StatusNotFound)
	if n := len(s.hub.Spectators(meta.ID)); n != 2 {
		t.Fatalf("a refused request changed the table: %d watching", n)
	}

	// The host removes: 200, the socket closes with the seat reason, and
	// the session is refused from then on.
	resp := s.removeSpectator(t, meta.ID, id, "", alice.Token)
	var out removeSpectatorResponse
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&out) != nil {
		t.Fatalf("host removal = %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
	if out.Blocked || len(out.Spectators) != 1 || out.Spectators[0].Name != "Other" {
		t.Errorf("response = %+v, want unblocked and only Other left", out)
	}
	readsRemovedClose(t, w.conn)

	resp = doJSON(t, s.srv, http.MethodGet, "/games/"+meta.ID.String(), w.sess.Token, nil)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized || !strings.Contains(string(body), auth.SeatRemovedMessage) {
		t.Errorf("a removed spectator's token = %d %s, want 401 %q", resp.StatusCode, body, auth.SeatRemovedMessage)
	}
	if _, err := dialAdmitted(s.srv, w.sess.Token); err == nil {
		t.Error("a removed spectator's session opened a socket")
	}
	// The other spectator is untouched, on the same table.
	if resp := doJSON(t, s.srv, http.MethodGet, "/games/"+meta.ID.String(), other.sess.Token, nil); resp.StatusCode != http.StatusOK {
		t.Errorf("the other spectator's token = %d", resp.StatusCode)
	} else {
		_ = resp.Body.Close()
	}

	// Without a block the invite still works: a NEW session watches.
	s.spectate(t, meta, "Watcher")

	// The creator and an admin may remove too.
	s.wantStatus(t, s.removeSpectator(t, meta.ID, other.sess.Principal.SpectatorID, "", identifiedToken(t, s.a, creator)), http.StatusOK)
	again := s.spectate(t, meta, "Again")
	s.wantStatus(t, s.removeSpectator(t, meta.ID, again.sess.Principal.SpectatorID, "", adminToken(t, s.srv)), http.StatusOK)
}

func TestHostMayRemoveASpectatorFromAStartedGame(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, seats := s.startedTable(t, uuid.New(), "Alice", "Bob", "Carol")
	alice := seats[0]
	w := s.spectate(t, meta, "Loud")

	// The seat route stays the admin's alone on a started game (#3093)...
	s.wantStatus(t, s.del(t, meta.ID, seats[1].PlayerID, alice.Token), http.StatusConflict)
	// ...a spectator is not in the game, so the host does not need one.
	s.wantStatus(t, s.removeSpectator(t, meta.ID, w.sess.Principal.SpectatorID, "", alice.Token), http.StatusOK)
	readsRemovedClose(t, w.conn)
	if p := s.enginePlayer(t, meta.ID, seats[1].PlayerID); p.Removed {
		t.Error("removing a spectator touched a seat")
	}
}

func TestBlockedSpectatorIsRefusedByNameAndTheLimitsAreDocumented(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, _ := s.table(t)
	named := s.spectate(t, meta, "Loud Lou")
	nameless := s.spectate(t, meta, "")

	resp := s.removeSpectator(t, meta.ID, named.sess.Principal.SpectatorID, "?block=true", alice.Token)
	var out removeSpectatorResponse
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&out) != nil || !out.Blocked {
		t.Fatalf("block removal = %d %+v, want blocked", resp.StatusCode, out)
	}
	_ = resp.Body.Close()
	readsRemovedClose(t, named.conn)

	// The list shows the folded name, to the host.
	list := s.listBlocks(t, meta.ID, alice.Token, http.StatusOK)
	if len(list) != 1 || list[0].Label != "loudlou" || list[0].Kind != BlockEntryGuest {
		t.Fatalf("block list = %+v, want the folded name", list)
	}

	// Under that name (in any spelling), watching is refused with the
	// stable code; under another, it is not (the documented limit).
	for _, name := range []string{"Loud Lou", "L0UD L0U", "loudlou"} {
		resp := postJSON(t, s.srv, "/games/"+meta.ID.String()+"/spectate", "",
			spectateRequest{InviteToken: meta.SpectatorInvite, Name: name})
		var body struct{ Error, Code string }
		_ = json.NewDecoder(resp.Body).Decode(&body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden || body.Code != CodeBlockedFromTable || body.Error == "" {
			t.Errorf("blocked spectator %q = %d %+v, want 403 %s", name, resp.StatusCode, body, CodeBlockedFromTable)
		}
	}
	s.spectate(t, meta, "Quiet Quentin")

	// A guest who gave no name has nothing a block can hold: the removal
	// goes through, the answer says so, and nothing is listed.
	resp = s.removeSpectator(t, meta.ID, nameless.sess.Principal.SpectatorID, "?block=true", alice.Token)
	out = removeSpectatorResponse{}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&out) != nil || out.Blocked {
		t.Fatalf("nameless block removal = %d %+v, want blocked:false", resp.StatusCode, out)
	}
	_ = resp.Body.Close()
	readsRemovedClose(t, nameless.conn)
	if n := len(s.listBlocks(t, meta.ID, alice.Token, http.StatusOK)); n != 1 {
		t.Errorf("the nameless removal appeared in the block list: %d entries", n)
	}

	// Lifting the block lets the name watch again.
	s.wantStatus(t, s.unblock(t, meta.ID, list[0].ID, alice.Token), http.StatusOK)
	s.spectate(t, meta, "Loud Lou")
}

func TestUnblockRevivesTheSessionTheEntryKilled(t *testing.T) {
	// A session removed WITH a block carries its marker under the same
	// entry, so lifting the entry lets that very token back in.
	s := newSeatRemoveStack(t)
	meta, alice, _ := s.table(t)
	w := s.spectate(t, meta, "Revived")
	s.wantStatus(t, s.removeSpectator(t, meta.ID, w.sess.Principal.SpectatorID, "?block=true", alice.Token), http.StatusOK)
	readsRemovedClose(t, w.conn)
	if resp := doJSON(t, s.srv, http.MethodGet, "/games/"+meta.ID.String(), w.sess.Token, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("removed token = %d", resp.StatusCode)
	} else {
		_ = resp.Body.Close()
	}
	id := s.listBlocks(t, meta.ID, alice.Token, http.StatusOK)[0].ID
	s.wantStatus(t, s.unblock(t, meta.ID, id, alice.Token), http.StatusOK)
	if resp := doJSON(t, s.srv, http.MethodGet, "/games/"+meta.ID.String(), w.sess.Token, nil); resp.StatusCode != http.StatusOK {
		t.Errorf("the token after the unblock = %d, want 200", resp.StatusCode)
	} else {
		_ = resp.Body.Close()
	}
}

func TestSignedInSpectatorBlockHoldsByAccountAndShowsTheDisplayName(t *testing.T) {
	l := NewLobby(newTestRoomManager(t))
	meta, err := l.Create("Account spectators")
	if err != nil {
		t.Fatal(err)
	}
	user := uuid.New()
	who := ws.LiveSpectator{ID: uuid.New(), Name: "Eve", UserID: user, DiscordID: "5150"}
	blocked, err := l.RemoveSpectator(meta.ID, who, true)
	if err != nil || !blocked {
		t.Fatalf("RemoveSpectator = %v, %v", blocked, err)
	}
	if !l.SpectatorRemoved(meta.ID, who.ID) {
		t.Error("the removed session is not marked")
	}
	if l.SpectatorRemoved(meta.ID, uuid.New()) || l.SpectatorRemoved(uuid.New(), who.ID) {
		t.Error("an unrelated session or table reads as removed")
	}
	list, _ := l.BlockList(meta.ID)
	if len(list) != 1 || list[0].Label != "Eve" || list[0].Kind != BlockEntryAccount {
		t.Fatalf("list = %+v, want one account entry labelled Eve", list)
	}
	// The account is shut out of watching AND of taking a seat.
	if _, err := l.SpectateAs(meta.ID, meta.SpectatorInvite, user, "", "anything"); !errors.Is(err, ErrBlockedFromWatching) {
		t.Errorf("same account: %v, want ErrBlockedFromWatching", err)
	}
	if _, err := l.SpectateAs(meta.ID, meta.SpectatorInvite, uuid.Nil, "5150", ""); !errors.Is(err, ErrBlockedFromWatching) {
		t.Errorf("same Discord id: %v", err)
	}
	if _, _, err := l.JoinAs(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "5150", Username: "eve"}, user); !errors.Is(err, ErrBlockedFromTable) {
		t.Errorf("a blocked spectator taking a seat: %v", err)
	}
	if _, err := l.SpectateAs(meta.ID, meta.SpectatorInvite, uuid.New(), "6000", "Eve"); err != nil {
		t.Errorf("a different account called Eve: %v", err)
	}
}

func TestSpectatorRemovalAndBlockSurviveARestart(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, err := l.Create("Durable spectators")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatal(err)
	}
	blockedGuest := ws.LiveSpectator{ID: uuid.New(), Name: "Heckler"}
	kickedOnly := ws.LiveSpectator{ID: uuid.New(), Name: "Kicked"}
	if _, err := l.RemoveSpectator(meta.ID, blockedGuest, true); err != nil {
		t.Fatal(err)
	}
	if _, err := l.RemoveSpectator(meta.ID, kickedOnly, false); err != nil {
		t.Fatal(err)
	}

	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games", n)
	}
	for _, id := range []uuid.UUID{blockedGuest.ID, kickedOnly.ID} {
		if !l2.SpectatorRemoved(meta.ID, id) {
			t.Errorf("session %s is back after a restart", id)
		}
	}
	if _, err := l2.SpectateAs(meta.ID, meta.SpectatorInvite, uuid.Nil, "", "heckler"); !errors.Is(err, ErrBlockedFromWatching) {
		t.Errorf("blocked guest after a restart: %v", err)
	}
	if _, err := l2.SpectateAs(meta.ID, meta.SpectatorInvite, uuid.Nil, "", "Kicked"); err != nil {
		t.Errorf("a kick without a block must not outlive the session: %v", err)
	}
	list, _ := l2.BlockList(meta.ID)
	if len(list) != 1 || list[0].Label != "heckler" {
		t.Errorf("list after a restart = %+v, want only the blocked guest (a bare kick is not listed)", list)
	}
}
