package lobby

// Tests for personal_blocks.go: a host's personal block keeps the blocked
// person out of the host's tables (ADR 0213 §3, owner decision 6 (a)).

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// blockAnswer is a status and a body, the two things a refused person sees.
type blockAnswer struct {
	status int
	body   string
}

func (s friendsStack) join(t *testing.T, tok, gameID, invite string) blockAnswer {
	t.Helper()
	code, raw := call(t, s, "POST", "/games/"+gameID+"/join", tok, map[string]string{"invite_token": invite})
	return blockAnswer{code, raw}
}

func (s friendsStack) joinCode(t *testing.T, tok, invite string) blockAnswer {
	t.Helper()
	code, raw := call(t, s, "POST", "/join", tok, map[string]string{"invite_token": invite})
	return blockAnswer{code, raw}
}

func (s friendsStack) preview(t *testing.T, tok, gameID, invite string) blockAnswer {
	t.Helper()
	code, raw := call(t, s, "GET", "/games/"+gameID+"/preview?t="+url.QueryEscape(invite), tok, nil)
	return blockAnswer{code, raw}
}

func (s friendsStack) spectate(t *testing.T, tok, gameID, invite string) blockAnswer {
	t.Helper()
	code, raw := call(t, s, "POST", "/games/"+gameID+"/spectate", tok, map[string]string{"invite_token": invite})
	return blockAnswer{code, raw}
}

func seatedUsers(t *testing.T, s friendsStack) map[string]bool {
	t.Helper()
	m, err := s.lobby.Get(s.meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]bool{}
	for _, seat := range m.Players {
		out[seat.UserID] = true
	}
	return out
}

// Alice hosts "FNM" and blocks Bob. Bob is refused at the invite preview,
// at the join and at the join by code, in the words a wrong token gets;
// Carol, whom Alice has not blocked, is not. A guest is not.
func TestPersonalBlockKeepsTheBlockedPersonOutOfTheHostsTable(t *testing.T) {
	s := newFriendsStack(t)
	gid, invite := s.meta.ID.String(), s.meta.InviteToken
	bob := s.bob.ID.String()

	// Before the block, Bob can see the table.
	if a := s.preview(t, s.bobTok, gid, invite); a.status != http.StatusOK {
		t.Fatalf("before the block, Bob's preview: %+v", a)
	}

	if code, raw := call(t, s, "PUT", "/me/blocks/"+bob, s.aliceTok, nil); code != http.StatusOK {
		t.Fatalf("block: %d %s", code, raw)
	}

	// What a person who holds a wrong token is told, at each entry.
	wrongPreview := s.preview(t, s.bobTok, gid, "not-the-invite")
	wrongJoin := s.join(t, s.bobTok, gid, "not-the-invite")
	wrongCode := s.joinCode(t, s.bobTok, "not-the-invite")
	if wrongPreview.status != http.StatusUnauthorized || wrongJoin.status != http.StatusUnauthorized || wrongCode.status != http.StatusUnauthorized {
		t.Fatalf("a wrong token is not a 401: %+v %+v %+v", wrongPreview, wrongJoin, wrongCode)
	}

	for name, got := range map[string]struct{ got, want blockAnswer }{
		"preview":      {s.preview(t, s.bobTok, gid, invite), wrongPreview},
		"join":         {s.join(t, s.bobTok, gid, invite), wrongJoin},
		"join by code": {s.joinCode(t, s.bobTok, invite), wrongCode},
	} {
		if got.got != got.want {
			t.Errorf("%s for the blocked person = %+v; want the wrong-token blockAnswer %+v", name, got.got, got.want)
		}
		if strings.Contains(strings.ToLower(got.got.body), "block") || strings.Contains(strings.ToLower(got.got.body), "removed") {
			t.Errorf("%s names the block: %s", name, got.got.body)
		}
	}
	if seatedUsers(t, s)[bob] {
		t.Fatal("the blocked person got a seat")
	}

	// Not blocked: Carol previews, and sits by code.
	if a := s.preview(t, s.carolTok, gid, invite); a.status != http.StatusOK {
		t.Errorf("Carol's preview: %+v", a)
	}
	if a := s.joinCode(t, s.carolTok, invite); a.status != http.StatusOK {
		t.Errorf("Carol joining by code: %+v", a)
	}
	// A guest has no account to block.
	if a := s.preview(t, "", gid, invite); a.status != http.StatusOK {
		t.Errorf("a guest's preview: %+v", a)
	}
	if code, raw := call(t, s, "POST", "/games/"+gid+"/join", "", map[string]string{"invite_token": invite, "name": "Guesty"}); code != http.StatusOK {
		t.Errorf("a guest joining: %d %s", code, raw)
	}

	// Lifting the block lifts the refusal.
	if code, raw := call(t, s, "DELETE", "/me/blocks/"+bob, s.aliceTok, nil); code != http.StatusNoContent {
		t.Fatalf("unblock: %d %s", code, raw)
	}
	if a := s.join(t, s.bobTok, gid, invite); a.status != http.StatusOK {
		t.Errorf("Bob after the unblock: %+v", a)
	}
}

// The block is the HOST's. Bob blocking Alice does not keep Bob out of
// Alice's table, and Alice blocking Bob does not keep Alice out of
// Bob's.
func TestPersonalBlockIsTheHostsOneWay(t *testing.T) {
	s := newFriendsStack(t)
	alice := s.alice.String()
	gid, invite := s.meta.ID.String(), s.meta.InviteToken

	if code, raw := call(t, s, "PUT", "/me/blocks/"+alice, s.bobTok, nil); code != http.StatusOK {
		t.Fatalf("Bob blocks Alice: %d %s", code, raw)
	}
	if a := s.preview(t, s.bobTok, gid, invite); a.status != http.StatusOK {
		t.Errorf("Bob's preview of Alice's table: %+v", a)
	}
	if a := s.join(t, s.bobTok, gid, invite); a.status != http.StatusOK {
		t.Errorf("Bob joining Alice's table, though he blocked her: %+v", a)
	}

	// A table Bob hosts is another matter: his block of Alice reaches it.
	bobTable, err := s.lobby.CreateBy("Bob's", s.bob.ID)
	if err != nil {
		t.Fatal(err)
	}
	if a := s.join(t, s.aliceTok, bobTable.ID.String(), bobTable.InviteToken); a.status == http.StatusOK {
		t.Errorf("Alice joining Bob's table, though Bob blocked her: %+v", a)
	}
	// Alice's own table, she hosts: no block of her own reaches it.
	if a := s.preview(t, s.aliceTok, gid, invite); a.status != http.StatusOK {
		t.Errorf("Alice's preview of her own table: %+v", a)
	}
}

// A block made while the person already sits leaves them seated: the
// host's Remove control is the tool for that (seat_remove.go). They are
// told they are seated, not that the invite is gone.
func TestPersonalBlockLeavesASeatedPersonSeated(t *testing.T) {
	s := newFriendsStack(t)
	gid, invite := s.meta.ID.String(), s.meta.InviteToken
	bob := s.bob.ID.String()

	if a := s.join(t, s.bobTok, gid, invite); a.status != http.StatusOK {
		t.Fatalf("Bob sits: %+v", a)
	}
	if code, raw := call(t, s, "PUT", "/me/blocks/"+bob, s.aliceTok, nil); code != http.StatusOK {
		t.Fatalf("block: %d %s", code, raw)
	}
	if !seatedUsers(t, s)[bob] {
		t.Fatal("the block unseated Bob")
	}
	if a := s.preview(t, s.bobTok, gid, invite); a.status != http.StatusOK {
		t.Errorf("a seated Bob's preview: %+v", a)
	}
	a := s.join(t, s.bobTok, gid, invite)
	if a.status != http.StatusConflict || !strings.Contains(a.body, ErrAlreadySeated.Error()) {
		t.Errorf("a seated Bob joining again: %+v; want the already-seated blockAnswer", a)
	}
}

// Watching is not a seat: decision 6 names a seat, so a blocked person's
// spectator invite still works.
func TestPersonalBlockDoesNotReachWatching(t *testing.T) {
	s := newFriendsStack(t)
	gid := s.meta.ID.String()
	if code, raw := call(t, s, "PUT", "/me/blocks/"+s.bob.ID.String(), s.aliceTok, nil); code != http.StatusOK {
		t.Fatalf("block: %d %s", code, raw)
	}
	if a := s.preview(t, s.bobTok, gid, s.meta.SpectatorInvite); a.status != http.StatusOK {
		t.Errorf("the blocked person's spectator preview: %+v", a)
	}
	if a := s.spectate(t, s.bobTok, gid, s.meta.SpectatorInvite); a.status != http.StatusOK {
		t.Errorf("the blocked person watching: %+v", a)
	}
}

// fakeBlocks is a PersonalBlocks over a set of "blocker>target" pairs.
type fakeBlocks struct {
	pairs map[string]bool
	err   error
}

func (f fakeBlocks) IsBlocking(_ context.Context, user, target string) (bool, error) {
	return f.pairs[user+">"+target], f.err
}

// The host is whoever hosts now: a table handed to someone else answers
// to that person's list, not the first host's.
func TestPersonalBlockFollowsTheHost(t *testing.T) {
	l := newTestLobby(t)
	host, carol, bob := uuid.New(), uuid.New(), uuid.New()
	meta, err := l.CreateBy("T", host)
	if err != nil {
		t.Fatal(err)
	}
	ident := func(id string) DiscordIdentity { return DiscordIdentity{ID: id, Username: id, GlobalName: id} }
	if _, _, err := l.JoinAs(meta.ID, meta.InviteToken, "", ident("d-host"), host); err != nil {
		t.Fatal(err)
	}
	_, carolSeat, err := l.JoinAs(meta.ID, meta.InviteToken, "", ident("d-carol"), carol)
	if err != nil {
		t.Fatal(err)
	}
	l.SetPersonalBlocks(fakeBlocks{pairs: map[string]bool{host.String() + ">" + bob.String(): true}})

	if _, _, err := l.JoinAs(meta.ID, meta.InviteToken, "", ident("d-bob"), bob); !errors.Is(err, ErrInvalidInvite) {
		t.Fatalf("Bob joining the first host's table: %v; want ErrInvalidInvite", err)
	}
	if _, err := l.TransferHost(meta.ID, carolSeat); err != nil {
		t.Fatal(err)
	}
	// Carol hosts, and has blocked nobody.
	if _, _, err := l.JoinAs(meta.ID, meta.InviteToken, "", ident("d-bob"), bob); err != nil {
		t.Fatalf("Bob joining after the table changed hands: %v", err)
	}
}

// A lookup that fails does not seat the person the host may have
// blocked, and does not pretend the invite is wrong.
func TestPersonalBlockLookupFailureRefusesTheSeat(t *testing.T) {
	l := newTestLobby(t)
	host, bob := uuid.New(), uuid.New()
	meta, err := l.CreateBy("T", host)
	if err != nil {
		t.Fatal(err)
	}
	// Nobody has sat yet: the creator hosts.
	boom := errors.New("db down")
	l.SetPersonalBlocks(fakeBlocks{err: boom})
	_, _, err = l.JoinAs(meta.ID, meta.InviteToken, "", DiscordIdentity{ID: "d-bob", Username: "b", GlobalName: "b"}, bob)
	if !errors.Is(err, boom) || errors.Is(err, ErrInvalidInvite) {
		t.Fatalf("JoinAs with a failing lookup: %v", err)
	}
	if got, _ := l.Get(meta.ID); len(got.Players) != 0 {
		t.Errorf("a seat was taken: %+v", got.Players)
	}
	// A guest never asks the list.
	if _, _, err := l.JoinAs(meta.ID, meta.InviteToken, "Guesty", DiscordIdentity{}, uuid.Nil); err != nil {
		t.Errorf("a guest joining: %v", err)
	}
}
