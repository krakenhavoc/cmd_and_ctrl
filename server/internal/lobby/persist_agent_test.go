package lobby

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// The agent badge's column (migration 0010, ADR 0124 §6): seatRecords
// writes it for an agent seat and leaves it NULL for every other, and
// loadEntry keeps the engine snapshot as the authority.

// seatAgentClients reads seats.agent_client for every seat of the
// game, keyed by player ID; a NULL is nil.
func seatAgentClients(t *testing.T, d *db.DB, gameID uuid.UUID) map[uuid.UUID]*string {
	t.Helper()
	rows, err := d.Query(`SELECT player_id, agent_client FROM seats WHERE game_id = ?`, gameID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	out := map[uuid.UUID]*string{}
	for rows.Next() {
		var (
			pid    string
			client sql.NullString
		)
		if err := rows.Scan(&pid, &client); err != nil {
			t.Fatal(err)
		}
		var v *string
		if client.Valid {
			v = &client.String
		}
		out[uuid.MustParse(pid)] = v
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestAgentSeatColumnIsWrittenForAgentsOnly(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, err := l.Create("FNM")
	if err != nil {
		t.Fatal(err)
	}
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	_, codex, err := l.JoinAgent(meta.ID, meta.InviteToken, "Codex", AgentDecl{Client: "Codex CLI"})
	if err != nil {
		t.Fatal(err)
	}
	_, nameless, err := l.JoinAgent(meta.ID, meta.InviteToken, "Agent", AgentDecl{Client: ""})
	if err != nil {
		t.Fatal(err)
	}
	_, bot, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20))
	if err != nil {
		t.Fatal(err)
	}

	got := seatAgentClients(t, openTestDB(t, dir), meta.ID)
	if len(got) != 4 {
		t.Fatalf("seats rows = %d, want 4", len(got))
	}
	for _, p := range []uuid.UUID{alice, bot} {
		if got[p] != nil {
			t.Errorf("seat %s has agent_client %q, want NULL: it is not an agent", p, *got[p])
		}
	}
	for p, want := range map[uuid.UUID]string{codex: "codex-cli", nameless: "unknown"} {
		if got[p] == nil || *got[p] != want {
			t.Errorf("agent seat %s has agent_client %v, want %q", p, got[p], want)
		}
	}

	// The column is rewritten with the seat list: a later write keeps
	// it, and a seat's departure takes its row with it.
	if _, err := l.SetDeck(meta.ID, alice, "deck", botDeck(20)); err != nil {
		t.Fatal(err)
	}
	if _, err := l.RemoveBot(meta.ID, bot); err != nil {
		t.Fatal(err)
	}
	got = seatAgentClients(t, openTestDB(t, dir), meta.ID)
	if c := got[codex]; c == nil || *c != "codex-cli" {
		t.Errorf("after a later seat write the agent's column is %v, want codex-cli", c)
	}
	if _, ok := got[bot]; ok {
		t.Error("the removed bot's row is still there")
	}
}

// A restored table takes the badge from the engine snapshot, whatever
// the column says. The column is read only for a seat whose engine
// player is missing.
func TestRestoreTakesTheAgentBadgeFromTheSnapshot(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, err := l.Create("FNM")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatal(err)
	}
	_, agent, err := l.JoinAgent(meta.ID, meta.InviteToken, "Claude", AgentDecl{Client: "claude-code"})
	if err != nil {
		t.Fatal(err)
	}

	d := openTestDB(t, dir)
	// The column disagrees with the snapshot: the snapshot wins.
	if _, err := d.Exec(`UPDATE seats SET agent_client = 'tampered' WHERE player_id = ?`, agent.String()); err != nil {
		t.Fatal(err)
	}
	// A seat row with no engine player behind it, as a table whose
	// restore point lost the seat would have: the column is all there
	// is.
	ghost := uuid.New()
	if _, err := d.Exec(`INSERT INTO seats (game_id, seat, player_id, guest_name, agent_client) VALUES (?, 9, ?, 'Ghost', 'codex')`,
		meta.ID.String(), ghost.String()); err != nil {
		t.Fatal(err)
	}

	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	back := mustGet(t, l2, meta.ID)
	found := map[uuid.UUID]SeatInfo{}
	for _, s := range back.Players {
		found[s.PlayerID] = s
	}
	if s := found[agent]; !s.IsAgent || s.AgentClient != "claude-code" {
		t.Errorf("restored agent seat: is_agent %v client %q, want the snapshot's claude-code", s.IsAgent, s.AgentClient)
	}
	if s, ok := found[ghost]; !ok || !s.IsAgent || s.AgentClient != "codex" {
		t.Errorf("seat with no engine player: %+v (present %v), want an agent seat from the column", s, ok)
	}
	for _, s := range back.Players {
		if s.PlayerID != agent && s.PlayerID != ghost && (s.IsAgent || s.AgentClient != "") {
			t.Errorf("seat %s came back as an agent: %+v", s.Name, s)
		}
	}
}

// The memory store carries the field like any other, so a lobby with
// no database round-trips it too.
func TestMemoryStoreCarriesAgentClient(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	id := uuid.New()
	if err := s.CreateGame(ctx, GameRecord{ID: id, Name: "G", State: "lobby"}, nil); err != nil {
		t.Fatal(err)
	}
	in := []SeatRecord{{Seat: 0, PlayerID: uuid.New(), GuestName: "A"}, {Seat: 1, PlayerID: uuid.New(), GuestName: "C", AgentClient: "codex"}}
	if err := s.ReplaceSeats(ctx, id, in); err != nil {
		t.Fatal(err)
	}
	_, out, err := s.LoadGame(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0].AgentClient != "" || out[1].AgentClient != "codex" {
		t.Errorf("LoadGame seats = %+v", out)
	}
}
