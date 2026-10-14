package lobby

// creator_seat_test.go: #3494. A signed-in creator sits down at the
// table they make, in seat 1; and a host may leave their own seat
// before the start, keeping the table when it is theirs.

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

func TestCreatingATableSeatsItsCreator(t *testing.T) {
	s := newTableStack(t)
	alice, aliceID := s.signIn(t, "discord-99", "Alice")

	created := s.mustCreate(t, alice, "Alice's table")
	if created.Seat == nil || created.Seat.Token == "" || created.Seat.PlayerID == uuid.Nil {
		t.Fatalf("no seat came back with the create: %+v", created.Seat)
	}
	// The plaintext invites are still the creator's to share.
	if created.InviteToken == "" || created.SpectatorInvite == "" {
		t.Errorf("seating the creator cost them the invites: %+v", created.GameMeta)
	}
	if len(created.Players) != 1 || created.Players[0].PlayerID != created.Seat.PlayerID {
		t.Fatalf("players = %+v, want the creator alone", created.Players)
	}
	if !created.Players[0].IsHost || created.Players[0].DiscordID != "discord-99" {
		t.Errorf("the creator's seat = %+v, want the host, as themselves", created.Players[0])
	}

	// The seat session is theirs: their user, bound to this seat.
	p := mustValidate(t, s.auth, created.Seat.Token)
	if p.Role != auth.RolePlayer || p.UserID != aliceID || p.GameID != created.ID || p.PlayerID != created.Seat.PlayerID {
		t.Errorf("seat session principal = %+v", p)
	}
}

func TestTheAdminTokenCreatesATableWithNoSeat(t *testing.T) {
	s := newTableStack(t)
	created := s.mustCreate(t, s.admin(t), "for the pod")
	if created.Seat != nil {
		t.Errorf("the admin token was seated: %+v", created.Seat)
	}
	if meta, _ := s.lobby.Get(created.ID); len(meta.Players) != 0 {
		t.Errorf("players = %+v, want none", meta.Players)
	}
}

// deleteSeat calls DELETE /games/{id}/seats/{player} and decodes the
// answer whatever its status.
func deleteSeat(t *testing.T, s *tableStack, tok string, gameID, playerID uuid.UUID) (int, leaveSeatResponse) {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, s.srv.URL+"/games/"+gameID.String()+"/seats/"+playerID.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := s.srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out leaveSeatResponse
	if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("decode: %v (%s)", err, raw)
		}
	}
	return resp.StatusCode, out
}

func TestTheCreatorLeavesTheirSeatAndKeepsTheTable(t *testing.T) {
	s := newTableStack(t)
	alice, aliceID := s.signIn(t, "discord-99", "Alice")
	created := s.mustCreate(t, alice, "Alice's table")
	bob := s.join(t, "", created.GameMeta, "Bob")

	code, out := deleteSeat(t, s, created.Seat.Token, created.ID, created.Seat.PlayerID)
	if code != http.StatusOK {
		t.Fatalf("leaving: %d", code)
	}
	// Signed in still: an identity session for the same person.
	if out.Session == nil {
		t.Fatal("no session to carry on with")
	}
	p := mustValidate(t, s.auth, out.Session.Token)
	if p.Role != auth.RoleIdentified || p.UserID != aliceID || p.DiscordID != "discord-99" {
		t.Errorf("next session = %+v", p)
	}
	if !out.IsCreator {
		t.Error("the answer does not say the table is still theirs")
	}

	// The table is held for her: Bob is not made host.
	meta, _ := s.lobby.Get(created.ID)
	if len(meta.Players) != 1 || meta.Players[0].PlayerID != bob.PlayerID {
		t.Fatalf("players = %+v, want Bob alone", meta.Players)
	}
	if meta.HostPlayerID != uuid.Nil || meta.Players[0].IsHost {
		t.Errorf("hosting passed to Bob: host %s", meta.HostPlayerID)
	}
	// Her old seat session is refused from now on (auth.WithSeatCheck).
	if !s.lobby.SeatRemoved(created.ID, created.Seat.PlayerID) {
		t.Error("the old seat is not marked removed")
	}

	// As creator she still manages the table, and sitting back down
	// makes her host again.
	back := s.join(t, out.Session.Token, meta, "")
	meta, _ = s.lobby.Get(created.ID)
	if meta.HostPlayerID != back.PlayerID {
		t.Errorf("host after sitting back down = %s, want %s", meta.HostPlayerID, back.PlayerID)
	}
}

func TestOnlyTheHostRemovesTheHostsSeat(t *testing.T) {
	s := newTableStack(t)
	alice, _ := s.signIn(t, "discord-99", "Alice")
	created := s.mustCreate(t, alice, "Alice's table")
	bob := s.join(t, "", created.GameMeta, "Bob")

	// Not Bob, and not an admin either: the host's seat is the host's.
	for name, tok := range map[string]string{"another seat": bob.Token, "admin": s.admin(t)} {
		if code, _ := deleteSeat(t, s, tok, created.ID, created.Seat.PlayerID); code == http.StatusOK {
			t.Errorf("%s removed the host's seat", name)
		}
	}
	if meta, _ := s.lobby.Get(created.ID); len(meta.Players) != 2 {
		t.Errorf("players = %+v, want both", meta.Players)
	}
}

// A host who is not the creator passes hosting on when they leave.
func TestAnotherHostLeavingPassesHostingOn(t *testing.T) {
	s := newTableStack(t)
	table := s.mustCreate(t, s.admin(t), "the pod")
	first := s.join(t, "", table.GameMeta, "First")
	second := s.join(t, "", table.GameMeta, "Second")
	if meta, _ := s.lobby.Get(table.ID); meta.HostPlayerID != first.PlayerID {
		t.Fatalf("host = %s, want the first seat", meta.HostPlayerID)
	}

	code, out := deleteSeat(t, s, first.Token, table.ID, first.PlayerID)
	if code != http.StatusOK {
		t.Fatalf("leaving: %d", code)
	}
	if out.Session != nil {
		t.Errorf("a guest was handed a session: %+v", out.Session)
	}
	if meta, _ := s.lobby.Get(table.ID); meta.HostPlayerID != second.PlayerID {
		t.Errorf("host = %s, want the next seat %s", meta.HostPlayerID, second.PlayerID)
	}
}
