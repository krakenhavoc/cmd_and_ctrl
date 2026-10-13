package lobby

// blocks_list_test.go: the block list and the unblock (#3113).

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func (s *seatRemoveStack) blocksPath(gameID uuid.UUID, rest ...string) string {
	return "/games/" + gameID.String() + "/blocks" + strings.Join(rest, "")
}

func (s *seatRemoveStack) listBlocks(t *testing.T, gameID uuid.UUID, token string, want int) []BlockEntry {
	t.Helper()
	resp := doJSON(t, s.srv, http.MethodGet, s.blocksPath(gameID), token, nil)
	defer resp.Body.Close()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET blocks = %d %s, want %d", resp.StatusCode, body, want)
	}
	if want != http.StatusOK {
		return nil
	}
	var out blockListResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out.Blocks
}

func (s *seatRemoveStack) unblock(t *testing.T, gameID uuid.UUID, blockID, token string) *http.Response {
	t.Helper()
	return doJSON(t, s.srv, http.MethodDelete, s.blocksPath(gameID, "/", blockID), token, nil)
}

// blockedBob is a table whose creator, host and a bystander can all be
// told apart: Alice hosts, Bob was removed and blocked, Carol sits.
func (s *seatRemoveStack) blockedBob(t *testing.T) (meta GameMeta, creator uuid.UUID, alice, carol sessionResponse) {
	t.Helper()
	creator = uuid.New()
	meta, err := s.l.CreateBy("Block list table", creator)
	if err != nil {
		t.Fatal(err)
	}
	alice = joinAs(t, s.srv, meta, "Alice")
	bob := joinAs(t, s.srv, meta, "Bob")
	carol = joinAs(t, s.srv, meta, "Carol")
	s.wantStatus(t,
		doJSON(t, s.srv, http.MethodDelete, "/games/"+meta.ID.String()+"/seats/"+bob.PlayerID.String()+"?block=true", alice.Token, nil),
		http.StatusOK)
	return meta, creator, alice, carol
}

func TestBlockListAuthorityOnAnUnstartedTable(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, creator, alice, carol := s.blockedBob(t)
	spec := decodeSession(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/spectate", "",
		map[string]string{"invite_token": meta.SpectatorInvite}), http.StatusOK)

	// Everyone who may not manage the table is refused, anonymous with 401.
	s.listBlocks(t, meta.ID, "", http.StatusUnauthorized)
	s.listBlocks(t, meta.ID, carol.Token, http.StatusForbidden)
	s.listBlocks(t, meta.ID, spec.Token, http.StatusForbidden)
	s.listBlocks(t, meta.ID, identifiedToken(t, s.a, uuid.New()), http.StatusForbidden)

	// The host, the creator and an admin see Bob, by his folded name.
	for name, tok := range map[string]string{
		"host":    alice.Token,
		"creator": identifiedToken(t, s.a, creator),
		"admin":   adminToken(t, s.srv),
	} {
		list := s.listBlocks(t, meta.ID, tok, http.StatusOK)
		if len(list) != 1 || list[0].Label != "bob" || list[0].Kind != BlockEntryGuest || list[0].ID == "" {
			t.Errorf("%s sees %+v, want one guest entry labelled bob", name, list)
		}
	}

	// A refused unblock lifted nothing.
	id := s.listBlocks(t, meta.ID, alice.Token, http.StatusOK)[0].ID
	s.wantStatus(t, s.unblock(t, meta.ID, id, carol.Token), http.StatusForbidden)
	s.wantStatus(t, s.unblock(t, meta.ID, id, spec.Token), http.StatusForbidden)
	s.wantStatus(t, s.unblock(t, meta.ID, id, ""), http.StatusUnauthorized)
	if n := len(s.listBlocks(t, meta.ID, alice.Token, http.StatusOK)); n != 1 {
		t.Errorf("a refused unblock changed the list: %d entries", n)
	}
}

func TestUnblockLetsThePersonSitBackDown(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, _, alice, _ := s.blockedBob(t)

	resp := postJSON(t, s.srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Bob"})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("blocked rejoin = %d, want 403", resp.StatusCode)
	}
	_ = resp.Body.Close()

	id := s.listBlocks(t, meta.ID, alice.Token, http.StatusOK)[0].ID
	resp = s.unblock(t, meta.ID, id, alice.Token)
	var out blockListResponse
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&out) != nil || len(out.Blocks) != 0 {
		t.Fatalf("unblock = %d %+v, want 200 and an empty list", resp.StatusCode, out)
	}
	_ = resp.Body.Close()

	joinAs(t, s.srv, meta, "Bob")

	// The entry is gone: lifting it again is a 404, not a second success.
	s.wantStatus(t, s.unblock(t, meta.ID, id, alice.Token), http.StatusNotFound)
	s.wantStatus(t, s.unblock(t, meta.ID, "no-such-entry", alice.Token), http.StatusNotFound)
}

func TestBlockListOnAStartedTableIsTheAdminsAlone(t *testing.T) {
	s := newSeatRemoveStack(t)
	creator := uuid.New()
	meta, seats := s.startedTable(t, creator, "Alice", "Bob", "Carol")
	alice, bob := seats[0], seats[1]
	admin := adminToken(t, s.srv)
	s.wantStatus(t,
		doJSON(t, s.srv, http.MethodDelete, "/games/"+meta.ID.String()+"/seats/"+bob.PlayerID.String()+"?block=true", admin, nil),
		http.StatusOK)

	// The host and the creator are answered 409, as for removing a seat.
	s.listBlocks(t, meta.ID, alice.Token, http.StatusConflict)
	s.listBlocks(t, meta.ID, identifiedToken(t, s.a, creator), http.StatusConflict)
	s.listBlocks(t, meta.ID, seats[2].Token, http.StatusForbidden)

	list := s.listBlocks(t, meta.ID, admin, http.StatusOK)
	if len(list) != 1 || list[0].Label != "bob" {
		t.Fatalf("admin sees %+v on a started table, want Bob's entry", list)
	}
	s.wantStatus(t, s.unblock(t, meta.ID, list[0].ID, alice.Token), http.StatusConflict)
	if n := s.l.BlockCount(meta.ID); n != 1 {
		t.Fatalf("a refused unblock changed the table: %d blocks", n)
	}
	s.wantStatus(t, s.unblock(t, meta.ID, list[0].ID, admin), http.StatusOK)
	if n := s.l.BlockCount(meta.ID); n != 0 {
		t.Errorf("blocks after the admin's unblock = %d", n)
	}
}

func TestBlockListNamesAnAccountByDisplayNameAndNeverShowsADiscordID(t *testing.T) {
	l := NewLobby(newTestRoomManager(t))
	meta, err := l.Create("Account list")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatal(err)
	}
	user := uuid.New()
	disc := DiscordIdentity{ID: "90019001", Username: "dave_d", GlobalName: "Dave"}
	_, dave, err := l.JoinAs(meta.ID, meta.InviteToken, "", disc, user)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.RemoveSeatBlocking(meta.ID, dave, true); err != nil {
		t.Fatal(err)
	}
	list, err := l.BlockList(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	// An account and its Discord id are two rows but one line.
	if n := l.BlockCount(meta.ID); n != 2 {
		t.Fatalf("blocks = %d, want the account and the Discord id", n)
	}
	if len(list) != 1 || list[0].Label != "Dave" || list[0].Kind != BlockEntryAccount {
		t.Fatalf("list = %+v, want one account entry labelled Dave", list)
	}
	raw, _ := json.Marshal(list)
	for _, secret := range []string{"90019001", user.String(), "dave_d"} {
		if strings.Contains(string(raw), secret) {
			t.Errorf("the block list leaks %q: %s", secret, raw)
		}
	}

	// One click lifts both rows.
	if err := l.Unblock(meta.ID, list[0].ID); err != nil {
		t.Fatal(err)
	}
	if n := l.BlockCount(meta.ID); n != 0 {
		t.Errorf("blocks after the unblock = %d, want 0", n)
	}
	if _, _, err := l.JoinAs(meta.ID, meta.InviteToken, "", disc, user); err != nil {
		t.Errorf("rejoin after the unblock: %v", err)
	}
}

func TestBlockListAndUnblockSurviveARestart(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, err := l.Create("Durable list")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatal(err)
	}
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	_, carol, _ := l.Join(meta.ID, meta.InviteToken, "Carol")
	for _, id := range []uuid.UUID{bob, carol} {
		if _, err := l.RemoveSeatBlocking(meta.ID, id, true); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := l.BlockList(meta.ID)
	if len(list) != 2 {
		t.Fatalf("list = %+v", list)
	}
	var bobEntry BlockEntry
	for _, e := range list {
		if e.Label == "bob" {
			bobEntry = e
		}
	}
	if err := l.Unblock(meta.ID, bobEntry.ID); err != nil {
		t.Fatal(err)
	}

	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games", n)
	}
	after, err := l2.BlockList(meta.ID)
	if err != nil || len(after) != 1 || after[0].Label != "carol" {
		t.Fatalf("list after a restart = %+v, %v; want only Carol (Bob's unblock must not come back)", after, err)
	}
	if _, _, err := l2.Join(meta.ID, meta.InviteToken, "Bob"); err != nil {
		t.Errorf("Bob after a restart: %v", err)
	}
	if _, _, err := l2.Join(meta.ID, meta.InviteToken, "Carol"); !errors.Is(err, ErrBlockedFromTable) {
		t.Errorf("Carol after a restart: %v, want still blocked", err)
	}
}
