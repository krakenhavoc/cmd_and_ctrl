package lobby

// Tests for ADR 0051 decision 8 — tablemates as a query (S34 sub-PR
// 6, tracking #607): the self-join itself in both stores, and
// GET /me/tablemates over it.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

// tablemateFixture seats four kinds of thing across three games so
// every exclusion the decision names is exercised at once:
//
//	old   (3 days ago): me, Bob, a guest, a bot
//	fresh (1 day ago):  me, Cara
//	other (2 days ago): Bob and Cara, without me
//
// My tablemates are therefore Cara (fresh) then Bob (old) — recency
// order — and never the guest, the bot, myself, or anyone I only
// share `other` with.
func tablemateFixture(t *testing.T, store Store, me, bob, cara string) {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	games := []struct {
		name  string
		at    time.Time
		seats []SeatRecord
	}{
		{"old", now.Add(-72 * time.Hour), []SeatRecord{
			{Seat: 0, PlayerID: uuid.New(), UserID: me, GuestName: "Me"},
			{Seat: 1, PlayerID: uuid.New(), UserID: bob, GuestName: "Bob"},
			{Seat: 2, PlayerID: uuid.New(), GuestName: "A Guest"},
			{Seat: 3, PlayerID: uuid.New(), GuestName: "Ghoul", BotTier: "heuristic"},
		}},
		{"fresh", now.Add(-24 * time.Hour), []SeatRecord{
			{Seat: 0, PlayerID: uuid.New(), UserID: me, GuestName: "Me"},
			{Seat: 1, PlayerID: uuid.New(), UserID: cara, GuestName: "Cara"},
		}},
		{"other", now.Add(-48 * time.Hour), []SeatRecord{
			{Seat: 0, PlayerID: uuid.New(), UserID: bob, GuestName: "Bob"},
			{Seat: 1, PlayerID: uuid.New(), UserID: cara, GuestName: "Cara"},
		}},
	}
	for _, g := range games {
		rec := GameRecord{ID: uuid.New(), Name: g.name, State: "ended", CreatedAt: g.at}
		if err := store.CreateGame(ctx, rec, nil); err != nil {
			t.Fatalf("CreateGame %s: %v", g.name, err)
		}
		if err := store.ReplaceSeats(ctx, rec.ID, g.seats); err != nil {
			t.Fatalf("ReplaceSeats %s: %v", g.name, err)
		}
	}
}

func TestTablematesInBothStores(t *testing.T) {
	for name, newStore := range map[string]func(t *testing.T) (Store, [3]string){
		"memory": func(*testing.T) (Store, [3]string) {
			return NewMemoryStore(), [3]string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
		},
		"sql": func(t *testing.T) (Store, [3]string) {
			d := openTestDB(t, t.TempDir())
			us := users.NewSQLStore(d, nil)
			var ids [3]string
			for i, who := range []discord.User{
				{ID: "d-me", Username: "me", GlobalName: "Me"},
				{ID: "d-bob", Username: "bob", GlobalName: "Bob"},
				{ID: "d-cara", Username: "cara", GlobalName: "Cara"},
			} {
				u, err := us.UpsertFromDiscord(context.Background(), who, "", "identify")
				if err != nil {
					t.Fatalf("UpsertFromDiscord: %v", err)
				}
				ids[i] = u.ID.String()
			}
			return NewSQLStore(d), ids
		},
	} {
		t.Run(name, func(t *testing.T) {
			store, ids := newStore(t)
			me, bob, cara := ids[0], ids[1], ids[2]
			tablemateFixture(t, store, me, bob, cara)

			got, err := store.Tablemates(context.Background(), me)
			if err != nil {
				t.Fatalf("Tablemates: %v", err)
			}
			if len(got) != 2 {
				t.Fatalf("Tablemates = %+v, want exactly Cara and Bob", got)
			}
			// Recency: the table we shared most recently first.
			if got[0].UserID != cara || got[1].UserID != bob {
				t.Errorf("order = %s, %s; want Cara (fresh) then Bob (old)", got[0].UserID, got[1].UserID)
			}
			if got[0].DisplayName != "Cara" || got[1].DisplayName != "Bob" {
				t.Errorf("display names = %q, %q", got[0].DisplayName, got[1].DisplayName)
			}
			if !got[0].LastPlayedAt.After(got[1].LastPlayedAt) {
				t.Errorf("last-played times are not in recency order: %v then %v",
					got[0].LastPlayedAt, got[1].LastPlayedAt)
			}
			for _, tm := range got {
				if tm.UserID == me {
					t.Error("the caller is in their own tablemates")
				}
				if tm.DisplayName == "A Guest" || tm.DisplayName == "Ghoul" {
					t.Errorf("a seat with no user is offered as a person: %+v", tm)
				}
			}

			// Someone who has never sat down has nobody.
			if got, err := store.Tablemates(context.Background(), uuid.NewString()); err != nil || len(got) != 0 {
				t.Errorf("a stranger's tablemates = %+v, %v; want none", got, err)
			}
			if got, err := store.Tablemates(context.Background(), ""); err != nil || len(got) != 0 {
				t.Errorf("empty user id = %+v, %v; want none", got, err)
			}
		})
	}
}

// TestTablematesCountsOnlyGamesTheCallerSatIn pins the half of the
// join that is easy to get wrong: sharing a game with somebody is not
// the same as them having played, and a game I am not in must not
// contribute either its people or its recency.
func TestTablematesCountsOnlyGamesTheCallerSatIn(t *testing.T) {
	store := NewMemoryStore()
	me, bob, cara := uuid.NewString(), uuid.NewString(), uuid.NewString()
	tablemateFixture(t, store, me, bob, cara)

	// `other` is the most recent game Bob and Cara share, but I am not
	// in it, so it must not pull Bob's recency ahead of Cara's.
	got, err := store.Tablemates(context.Background(), me)
	if err != nil {
		t.Fatalf("Tablemates: %v", err)
	}
	if got[0].UserID != cara {
		t.Fatalf("a game the caller never sat in changed the order: %+v", got)
	}
}

func TestLobbyTablematesWithNoUser(t *testing.T) {
	l := newTestLobby(t)
	got, err := l.Tablemates(uuid.Nil)
	if err != nil || len(got) != 0 {
		t.Errorf("Tablemates(uuid.Nil) = %+v, %v; want an empty list", got, err)
	}
}

// --- GET /me/tablemates -------------------------------------------

// getTablemates calls the route and returns the status and the list.
func getTablemates(t *testing.T, s userStack, tok string) (int, []Tablemate, string) {
	t.Helper()
	resp := doGet(t, s.srv, "/me/tablemates", tok)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var body tablematesResponse
	if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("decode /me/tablemates: %v (%s)", err, raw)
		}
	}
	return resp.StatusCode, body.Tablemates, string(raw)
}

func TestMeTablematesListsPeopleYouHavePlayedWith(t *testing.T) {
	s := newMyGamesStack(t)
	tok := identityTokenFromCallback(t, s.srv, s.state)
	me := mustValidate(t, s.auth, tok).UserID

	bob, err := s.users.UpsertFromDiscord(context.Background(),
		discord.User{ID: "d-bob", Username: "bob", GlobalName: "Bob", Avatar: "bobhash"}, "", "identify")
	if err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}

	meta, err := s.lobby.CreateBy("FNM", me)
	if err != nil {
		t.Fatalf("CreateBy: %v", err)
	}
	if _, _, err := s.lobby.JoinAs(meta.ID, meta.InviteToken, "",
		DiscordIdentity{ID: "discord-99", Username: "alice", GlobalName: "Alice"}, me); err != nil {
		t.Fatalf("JoinAs me: %v", err)
	}
	if _, _, err := s.lobby.JoinAs(meta.ID, meta.InviteToken, "",
		DiscordIdentity{ID: "d-bob", Username: "bob", GlobalName: "Bob", AvatarHash: "bobhash"}, bob.ID); err != nil {
		t.Fatalf("JoinAs Bob: %v", err)
	}

	status, mates, raw := getTablemates(t, s, tok)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, raw)
	}
	if len(mates) != 1 || mates[0].UserID != bob.ID {
		t.Fatalf("tablemates = %+v, want just Bob (%s)", mates, bob.ID)
	}
	if mates[0].DisplayName != "Bob" {
		t.Errorf("display_name = %q", mates[0].DisplayName)
	}
	if mates[0].AvatarURL != "/avatars/d-bob/bobhash.png" {
		t.Errorf("avatar_url = %q", mates[0].AvatarURL)
	}
	if mates[0].LastPlayedAt <= 0 {
		t.Errorf("last_played_at = %d, want the shared table's creation time", mates[0].LastPlayedAt)
	}
	// A snowflake is never on the wire as an identifier. (The avatar
	// path contains one, which is the path the client already loads
	// every seat's avatar from — so the check is on the JSON fields,
	// not on the whole body.)
	if mates[0].UserID.String() == "d-bob" {
		t.Errorf("user_id is a Discord snowflake: %s", raw)
	}
}

func TestMeTablematesNeedsASignedInPerson(t *testing.T) {
	s := newMyGamesStack(t)

	if status, _, _ := getTablemates(t, s, ""); status != http.StatusUnauthorized {
		t.Errorf("no session: status %d, want 401", status)
	}
	if status, _, _ := getTablemates(t, s, adminToken(t, s.srv)); status != http.StatusForbidden {
		t.Errorf("admin: status %d, want 403 (an admin is a credential, not a person — but it IS authenticated, #1154)", status)
	}

	// A guest's seat session has a seat but no person.
	meta, err := s.lobby.Create("guests only")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, playerID, err := s.lobby.Join(meta.ID, meta.InviteToken, "Guest")
	if err != nil {
		t.Fatalf("Join: %v", err)
	}
	guest := playerToken(t, s.auth, meta.ID, playerID, "Guest")
	if status, _, _ := getTablemates(t, s, guest); status != http.StatusForbidden {
		t.Errorf("guest seat: status %d, want 403", status)
	}
}

// TestMeTablematesOnADeploymentWithNoDatabase: with no users table
// there is nobody to be, so the route answers 403 for every session,
// exactly as GET /me/games and GET /me/decks do — and the memory
// store's own Tablemates still answers an empty list rather than
// erroring, so the handler is never the thing that breaks.
func TestMeTablematesOnADeploymentWithNoDatabase(t *testing.T) {
	srv, l, _, state := newDiscordTestStack(t)
	meta, err := l.Create("no db")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Guest"); err != nil {
		t.Fatalf("Join: %v", err)
	}
	// A Discord sign-in on this deployment mints an identity session
	// with a zero UserID (users.NoStore), which is the case that
	// matters: it LOOKS signed in and still has no person behind it.
	tok := identityTokenFromCallback(t, srv, state)
	resp := doGet(t, srv, "/me/tablemates", tok)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status %d, want 403", resp.StatusCode)
	}

	got, err := NewMemoryStore().Tablemates(context.Background(), uuid.NewString())
	if err != nil || len(got) != 0 {
		t.Errorf("memory Tablemates = %+v, %v; want an empty list and no error", got, err)
	}
}
