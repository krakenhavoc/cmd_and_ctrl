package lobby

// Tests for friends_http.go and the blocks GET /me/tablemates and the DM
// invite route apply (ADR 0213 §1 to §3, slice 3 of #3313).

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// friendsStack is a seatedStack with a session for Bob too: Alice and Bob
// shared "Last week", Alice sits at "FNM", and Carol has an account and
// has never sat with either.
type friendsStack struct {
	seatedStack
	aliceTok string
	bobTok   string
	carol    users.User
	carolTok string
}

func newFriendsStack(t *testing.T) friendsStack {
	t.Helper()
	s := newSeatedStack(t, "bot-token-shhh")
	carol := dmTarget(t, s.userStack, "d-carol", "Carol")
	return friendsStack{
		seatedStack: s,
		aliceTok:    identityToken(t, s.auth, s.alice, "discord-99", "Alice"),
		bobTok:      identityToken(t, s.auth, s.bob.ID, "d-bob", "Bob"),
		carol:       carol,
		carolTok:    identityToken(t, s.auth, carol.ID, "d-carol", "Carol"),
	}
}

func call(t *testing.T, s friendsStack, method, path, tok string, body any) (int, string) {
	t.Helper()
	resp := do(t, s.srv, method, path, tok, body)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func (s friendsStack) ask(t *testing.T, tok string, body any) (int, friendRequestResponse, string) {
	t.Helper()
	code, raw := call(t, s, "POST", "/me/friends/requests", tok, body)
	var out friendRequestResponse
	_ = json.Unmarshal([]byte(raw), &out)
	return code, out, raw
}

func (s friendsStack) friendsOf(t *testing.T, tok string) friendsResponse {
	t.Helper()
	code, raw := call(t, s, "GET", "/me/friends", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("GET /me/friends: %d %s", code, raw)
	}
	var out friendsResponse
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func entryIDs(es []FriendEntry) []string {
	out := []string{}
	for _, e := range es {
		out = append(out, e.UserID)
	}
	return out
}

func mateByID(ms []Tablemate, id uuid.UUID) (Tablemate, bool) {
	for _, m := range ms {
		if m.UserID == id {
			return m, true
		}
	}
	return Tablemate{}, false
}

func TestFriendRequestAskAcceptAndUnfriend(t *testing.T) {
	s := newFriendsStack(t)
	bob := s.bob.ID.String()
	alice := s.alice.String()

	code, out, raw := s.ask(t, s.aliceTok, map[string]string{"user_id": bob})
	if code != http.StatusOK || out.Outcome != "sent" || out.UserID != bob || out.DisplayName != "Bob" {
		t.Fatalf("ask: %d %s", code, raw)
	}
	if got := s.friendsOf(t, s.aliceTok); strings.Join(entryIDs(got.Outgoing), ",") != bob ||
		got.MaxFriends != 200 || got.MaxOutgoing != 20 {
		t.Errorf("Alice's overview: %+v", got)
	}
	got := s.friendsOf(t, s.bobTok)
	if len(got.Incoming) != 1 || got.Incoming[0].UserID != alice || got.Incoming[0].DisplayName != "Alice" {
		t.Fatalf("Bob's incoming: %+v", got.Incoming)
	}
	// Played with marks the request on both sides.
	_, mates, _ := getTablemates(t, s.userStack, s.bobTok)
	if m, ok := mateByID(mates, s.alice); !ok || m.Request != RequestIncoming || m.Friend {
		t.Errorf("Bob's row for Alice: %+v %v", m, ok)
	}
	_, mates, _ = getTablemates(t, s.userStack, s.aliceTok)
	if m, _ := mateByID(mates, s.bob.ID); m.Request != RequestOutgoing {
		t.Errorf("Alice's row for Bob: %+v", m)
	}

	if code, raw := call(t, s, "POST", "/me/friends/requests/"+alice+"/accept", s.bobTok, nil); code != http.StatusOK ||
		!strings.Contains(raw, `"outcome":"friends"`) {
		t.Fatalf("accept: %d %s", code, raw)
	}
	if got := s.friendsOf(t, s.aliceTok); strings.Join(entryIDs(got.Friends), ",") != bob || len(got.Outgoing) != 0 {
		t.Errorf("Alice after the accept: %+v", got)
	}
	_, mates, _ = getTablemates(t, s.userStack, s.aliceTok)
	if m, _ := mateByID(mates, s.bob.ID); !m.Friend || m.Request != "" {
		t.Errorf("Alice's row for her friend Bob: %+v", m)
	}
	if code, _, raw := s.ask(t, s.aliceTok, map[string]string{"user_id": bob}); code != http.StatusOK ||
		!strings.Contains(raw, "already_friends") {
		t.Errorf("asking a friend: %d %s", code, raw)
	}

	if code, raw := call(t, s, "DELETE", "/me/friends/"+alice, s.bobTok, nil); code != http.StatusNoContent {
		t.Fatalf("unfriend: %d %s", code, raw)
	}
	if got := s.friendsOf(t, s.aliceTok); len(got.Friends) != 0 {
		t.Errorf("still friends after Bob removed Alice: %+v", got.Friends)
	}
}

func TestFriendRequestIgnoreAndCancel(t *testing.T) {
	s := newFriendsStack(t)
	bob, alice := s.bob.ID.String(), s.alice.String()
	s.ask(t, s.aliceTok, map[string]string{"user_id": bob})
	if code, raw := call(t, s, "POST", "/me/friends/requests/"+alice+"/ignore", s.bobTok, nil); code != http.StatusNoContent {
		t.Fatalf("ignore: %d %s", code, raw)
	}
	if got := s.friendsOf(t, s.bobTok); len(got.Incoming) != 0 {
		t.Errorf("an ignored request is still Bob's to answer: %+v", got.Incoming)
	}
	if got := s.friendsOf(t, s.aliceTok); len(got.Outgoing) != 1 {
		t.Errorf("Alice was told Bob ignored her: %+v", got.Outgoing)
	}
	if code, raw := call(t, s, "DELETE", "/me/friends/requests/"+bob, s.aliceTok, nil); code != http.StatusNoContent {
		t.Fatalf("cancel: %d %s", code, raw)
	}
	if code, raw := call(t, s, "DELETE", "/me/friends/requests/"+bob, s.aliceTok, nil); code != http.StatusNotFound {
		t.Errorf("cancelling twice: %d %s", code, raw)
	}
	if code, raw := call(t, s, "POST", "/me/friends/requests/"+alice+"/accept", s.bobTok, nil); code != http.StatusNotFound {
		t.Errorf("accepting a cancelled request: %d %s", code, raw)
	}
}

func TestFriendRequestNamesOnlySomeoneYouPlayedWith(t *testing.T) {
	s := newFriendsStack(t)
	code, _, stranger := s.ask(t, s.aliceTok, map[string]string{"user_id": s.carol.ID.String()})
	if code != http.StatusNotFound || !strings.Contains(stranger, "played with") {
		t.Fatalf("a stranger: %d %s", code, stranger)
	}
	if _, _, nobody := s.ask(t, s.aliceTok, map[string]string{"user_id": uuid.NewString()}); nobody != stranger {
		t.Errorf("an id that is nobody's answered %q, a stranger %q", nobody, stranger)
	}
	for _, bad := range []map[string]string{
		{},
		{"user_id": "not-a-uuid"},
		{"user_id": s.alice.String()},
		{"user_id": s.bob.ID.String(), "game_id": s.meta.ID.String()},
		{"game_id": s.meta.ID.String()},
	} {
		if code, _, raw := s.ask(t, s.aliceTok, bad); code != http.StatusBadRequest {
			t.Errorf("body %v: %d %s, want 400", bad, code, raw)
		}
	}
}

func TestFriendRequestFromASeat(t *testing.T) {
	s := newFriendsStack(t)
	_, bobPlayer, err := s.lobby.JoinAs(s.meta.ID, s.meta.InviteToken, "",
		DiscordIdentity{ID: "d-bob", Username: "bob", GlobalName: "Bob"}, s.bob.ID)
	if err != nil {
		t.Fatalf("JoinAs Bob: %v", err)
	}
	_, guest, err := s.lobby.Join(s.meta.ID, s.meta.InviteToken, "Guesty")
	if err != nil {
		t.Fatalf("Join guest: %v", err)
	}
	seat := func(pid uuid.UUID) map[string]string {
		return map[string]string{"game_id": s.meta.ID.String(), "player_id": pid.String()}
	}

	code, out, raw := s.ask(t, s.aliceTok, seat(bobPlayer))
	if code != http.StatusOK || out.Outcome != "sent" || out.DisplayName != "Bob" {
		t.Fatalf("from Bob's seat: %d %s", code, raw)
	}
	if strings.Contains(raw, s.bob.ID.String()) {
		t.Errorf("the seat's answer carries the account id: %s", raw)
	}
	if code, _, raw := s.ask(t, s.aliceTok, seat(guest)); code != http.StatusUnprocessableEntity {
		t.Errorf("a guest seat: %d %s, want 422", code, raw)
	}
	if code, _, raw := s.ask(t, s.aliceTok, seat(uuid.New())); code != http.StatusNotFound {
		t.Errorf("no such seat: %d %s, want 404", code, raw)
	}
	// Carol does not sit at FNM, so its seats are not hers to name.
	if code, _, raw := s.ask(t, s.carolTok, seat(bobPlayer)); code != http.StatusForbidden {
		t.Errorf("a caller not at the table: %d %s, want 403", code, raw)
	}
}

func TestBlockingHidesBothWaysAndRefusesInvites(t *testing.T) {
	s := newFriendsStack(t)
	bob, alice := s.bob.ID.String(), s.alice.String()
	// Friends first, so the block has something to end.
	s.ask(t, s.aliceTok, map[string]string{"user_id": bob})
	s.ask(t, s.bobTok, map[string]string{"user_id": alice})
	if got := s.friendsOf(t, s.aliceTok); len(got.Friends) != 1 {
		t.Fatalf("not friends: %+v", got)
	}

	if code, raw := call(t, s, "PUT", "/me/blocks/"+bob, s.aliceTok, nil); code != http.StatusOK || !strings.Contains(raw, `"blocked":true`) {
		t.Fatalf("block: %d %s", code, raw)
	}
	if got := s.friendsOf(t, s.bobTok); len(got.Friends) != 0 {
		t.Errorf("the block left the friendship: %+v", got.Friends)
	}
	for who, tok := range map[string]string{"Alice": s.aliceTok, "Bob": s.bobTok} {
		_, mates, _ := getTablemates(t, s.userStack, tok)
		if len(mates) != 0 {
			t.Errorf("%s's Played with still lists the other: %+v", who, mates)
		}
	}
	if code, raw := call(t, s, "GET", "/me/blocks/"+bob, s.aliceTok, nil); code != http.StatusOK || !strings.Contains(raw, "true") {
		t.Errorf("Alice's block of Bob: %d %s", code, raw)
	}
	if code, raw := call(t, s, "GET", "/me/blocks/"+alice, s.bobTok, nil); code != http.StatusOK || !strings.Contains(raw, "false") {
		t.Errorf("Bob is told about Alice's block: %d %s", code, raw)
	}
	code, raw := call(t, s, "GET", "/me/blocks", s.aliceTok, nil)
	if code != http.StatusOK || !strings.Contains(raw, bob) || !strings.Contains(raw, "Bob") {
		t.Errorf("Alice's list: %d %s", code, raw)
	}

	// Requests: the blocker is told why, the blocked person is not.
	if code, _, raw := s.ask(t, s.aliceTok, map[string]string{"user_id": bob}); code != http.StatusForbidden ||
		!strings.Contains(raw, "you blocked") {
		t.Errorf("the blocker asking: %d %s", code, raw)
	}
	code, _, blockedAsk := s.ask(t, s.bobTok, map[string]string{"user_id": alice})
	if code != http.StatusForbidden || strings.Contains(blockedAsk, "block") {
		t.Errorf("the blocked person asking: %d %s", code, blockedAsk)
	}

	// DM invites: Alice to Bob says she blocked him; Bob to Alice reads
	// like a stranger.
	if status, raw := postDM(t, s.userStack, s.meta.ID, s.aliceSeat, map[string]string{"user_id": bob}); status != http.StatusForbidden ||
		!strings.Contains(raw, "you blocked") {
		t.Errorf("Alice inviting Bob: %d %s", status, raw)
	}
	bobTable, err := s.lobby.CreateBy("Bob's", s.bob.ID)
	if err != nil {
		t.Fatal(err)
	}
	if status, raw := postDM(t, s.userStack, bobTable.ID, s.bobTok, map[string]string{"user_id": alice}); status != http.StatusNotFound ||
		!strings.Contains(raw, "no such tablemate") {
		t.Errorf("Bob inviting Alice: %d %s", status, raw)
	}
	if opens, _, _, _ := s.stub.snapshot(); opens != 0 {
		t.Errorf("a refused invite reached Discord (%d opens)", opens)
	}

	// Unblocking brings back the list, not the friendship.
	if code, raw := call(t, s, "DELETE", "/me/blocks/"+bob, s.aliceTok, nil); code != http.StatusNoContent {
		t.Fatalf("unblock: %d %s", code, raw)
	}
	_, mates, _ := getTablemates(t, s.userStack, s.aliceTok)
	if m, ok := mateByID(mates, s.bob.ID); !ok || m.Friend {
		t.Errorf("after the unblock: %+v %v; want Bob back, not a friend", m, ok)
	}
}

func TestBlockNeedsAnAccount(t *testing.T) {
	s := newFriendsStack(t)
	if code, raw := call(t, s, "PUT", "/me/blocks/"+uuid.NewString(), s.aliceTok, nil); code != http.StatusNotFound {
		t.Errorf("blocking nobody: %d %s", code, raw)
	}
	if code, raw := call(t, s, "PUT", "/me/blocks/"+s.alice.String(), s.aliceTok, nil); code != http.StatusBadRequest {
		t.Errorf("blocking yourself: %d %s", code, raw)
	}
	if code, raw := call(t, s, "PUT", "/me/blocks/nope", s.aliceTok, nil); code != http.StatusBadRequest {
		t.Errorf("a bad id: %d %s", code, raw)
	}
	// Anyone with an account can be blocked, played with or not.
	if code, raw := call(t, s, "PUT", "/me/blocks/"+s.carol.ID.String(), s.aliceTok, nil); code != http.StatusOK {
		t.Errorf("blocking Carol: %d %s", code, raw)
	}
}

func TestDMInviteReachesAFriendWhoIsNotATablemate(t *testing.T) {
	s := newFriendsStack(t)
	carol := s.carol.ID.String()
	if status, _ := postDM(t, s.userStack, s.meta.ID, s.aliceSeat, map[string]string{"user_id": carol}); status != http.StatusNotFound {
		t.Fatalf("a stranger: %d", status)
	}
	a, b := s.alice.String(), carol
	if a > b {
		a, b = b, a
	}
	if _, err := s.db.Exec(`INSERT INTO friendships (user_a, user_b, created_at) VALUES (?, ?, 1)`, a, b); err != nil {
		t.Fatal(err)
	}
	if status, raw := postDM(t, s.userStack, s.meta.ID, s.aliceSeat, map[string]string{"user_id": carol}); status != http.StatusOK {
		t.Fatalf("a friend: %d %s", status, raw)
	}
	if _, _, recipients, _ := s.stub.snapshot(); len(recipients) != 1 || recipients[0] != "d-carol" {
		t.Errorf("recipients = %v", recipients)
	}
}

func TestSocialSettings(t *testing.T) {
	s := newFriendsStack(t)
	code, raw := call(t, s, "GET", "/me/social", s.bobTok, nil)
	if code != http.StatusOK || raw != `{"presence":"friends","requests":"played_with"}`+"\n" {
		t.Fatalf("defaults: %d %q", code, raw)
	}
	if code, raw := call(t, s, "PUT", "/me/social", s.bobTok, map[string]string{"requests": "nobody"}); code != http.StatusOK ||
		!strings.Contains(raw, `"requests":"nobody"`) || !strings.Contains(raw, `"presence":"friends"`) {
		t.Fatalf("PUT requests: %d %s", code, raw)
	}
	// A refusal for "nobody" reads the same as a block's.
	code, _, nobody := s.ask(t, s.aliceTok, map[string]string{"user_id": s.bob.ID.String()})
	if code != http.StatusForbidden || !strings.Contains(nobody, errRequestRefused) {
		t.Errorf("asking someone who takes none: %d %s", code, nobody)
	}
	for _, bad := range []any{
		map[string]string{},
		map[string]string{"presence": "everyone"},
		map[string]string{"requests": "friends"},
	} {
		if code, raw := call(t, s, "PUT", "/me/social", s.bobTok, bad); code != http.StatusBadRequest {
			t.Errorf("PUT %v: %d %s", bad, code, raw)
		}
	}
	if code, raw := call(t, s, "PUT", "/me/social", s.bobTok, map[string]string{"presence": "nobody"}); code != http.StatusOK ||
		!strings.Contains(raw, `"requests":"nobody"`) {
		t.Errorf("the second PUT lost the first: %d %s", code, raw)
	}
}

func TestMyGamesMarksFriendsAtTheTable(t *testing.T) {
	s := newFriendsStack(t)
	s.ask(t, s.aliceTok, map[string]string{"user_id": s.bob.ID.String()})
	s.ask(t, s.bobTok, map[string]string{"user_id": s.alice.String()})
	code, raw := call(t, s, "GET", "/me/games", s.aliceTok, nil)
	if code != http.StatusOK {
		t.Fatalf("GET /me/games: %d %s", code, raw)
	}
	var body myGamesResponse
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, g := range body.Games {
		if g.ID != s.earlier.ID {
			continue
		}
		for _, o := range g.Others {
			if o.Name == "Bob" {
				found = true
				if !o.Friend {
					t.Errorf("Bob at %q is not marked a friend: %+v", g.Name, o)
				}
			}
		}
	}
	if !found {
		t.Fatalf("no Bob at the earlier table: %s", raw)
	}
	if strings.Contains(raw, s.bob.ID.String()) {
		t.Errorf("GET /me/games carries an account id: %s", raw)
	}
}

func TestFriendsRoutesNeedAPersonAndADatabase(t *testing.T) {
	s := newFriendsStack(t)
	_, guest, err := s.lobby.Join(s.meta.ID, s.meta.InviteToken, "Guesty")
	if err != nil {
		t.Fatal(err)
	}
	guestTok := playerToken(t, s.auth, s.meta.ID, guest, "Guesty")
	for _, tok := range []string{guestTok, adminToken(t, s.srv)} {
		if code, raw := call(t, s, "GET", "/me/friends", tok, nil); code != http.StatusForbidden {
			t.Errorf("GET /me/friends for a credential: %d %s", code, raw)
		}
	}

	off := newUserStackWith(t, nil, func(c *Config) { c.Social = nil })
	tok := identityTokenFromCallback(t, off.srv, off.state)
	for _, r := range []struct{ method, path string }{
		{"GET", "/me/friends"}, {"GET", "/me/blocks"}, {"GET", "/me/social"},
		{"PUT", "/me/blocks/" + uuid.NewString()},
	} {
		resp := do(t, off.srv, r.method, r.path, tok, nil)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("%s %s with no social store: %d, want 503", r.method, r.path, resp.StatusCode)
		}
	}
	// Played with still works, with nothing to mark.
	if code, _, raw := getTablemates(t, off, tok); code != http.StatusOK {
		t.Errorf("GET /me/tablemates with no social store: %d %s", code, raw)
	}
}
