package lobby

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/google/uuid"
)

// ADR 0158 §1 item 4: GET /games carries the turn and each seat's life
// total for a game in play, and neither for a table in the lobby or a
// game that has ended.

func TestListCarriesTurnAndLifeForAnActiveGameOnly(t *testing.T) {
	l := newTestLobby(t)
	waiting, err := l.Create("Waiting")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := l.Join(waiting.ID, waiting.InviteToken, "Carol"); err != nil {
		t.Fatalf("Join Carol: %v", err)
	}
	meta, alice, bob := startTwoSeatGame(t, l, "Playing")
	if _, err := l.RoomOf(meta.ID).Game.ChangePlayerLife(alice, -5); err != nil {
		t.Fatalf("ChangePlayerLife: %v", err)
	}

	byID := map[uuid.UUID]GameMeta{}
	for _, g := range l.List() {
		byID[g.ID] = g
	}

	// While the opening roll is open no turn has begun, so none is
	// listed; the life totals already are.
	for _, g := range l.List() {
		if g.ID == meta.ID && g.Turn != 0 {
			t.Errorf("during the opening roll: turn = %d, want 0", g.Turn)
		}
	}
	finishOpeningRoll(t, l, meta.ID)
	byID = map[uuid.UUID]GameMeta{}
	for _, g := range l.List() {
		byID[g.ID] = g
	}

	active := byID[meta.ID]
	if active.Turn != 1 {
		t.Errorf("active game: turn = %d, want 1", active.Turn)
	}
	lives := map[uuid.UUID]int{}
	for _, p := range active.Players {
		if p.Life == nil {
			t.Fatalf("active game: seat %q has no life", p.Name)
		}
		lives[p.PlayerID] = *p.Life
	}
	if lives[alice] != 35 || lives[bob] != 40 {
		t.Errorf("active game lives: alice %d, bob %d; want 35 and 40", lives[alice], lives[bob])
	}

	lobbyTable := byID[waiting.ID]
	if lobbyTable.Turn != 0 {
		t.Errorf("lobby table: turn = %d, want 0", lobbyTable.Turn)
	}
	for _, p := range lobbyTable.Players {
		if p.Life != nil {
			t.Errorf("lobby table: seat %q has life %d, want none", p.Name, *p.Life)
		}
	}

	// The list's copy is the only one that carries them: the lobby's
	// own meta, which GET /games/{id} serves, never does.
	got, err := l.Get(meta.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Turn != 0 {
		t.Errorf("Get: turn = %d, want 0 (list only)", got.Turn)
	}
	for _, p := range got.Players {
		if p.Life != nil {
			t.Errorf("Get: seat %q has life, want none (list only)", p.Name)
		}
	}

	// Ended: neither field.
	l.RoomOf(meta.ID).Game.End()
	for _, g := range l.List() {
		if g.ID != meta.ID {
			continue
		}
		if g.State != "ended" {
			t.Fatalf("state = %q, want ended", g.State)
		}
		if g.Turn != 0 {
			t.Errorf("ended game: turn = %d, want 0", g.Turn)
		}
		for _, p := range g.Players {
			if p.Life != nil {
				t.Errorf("ended game: seat %q has life, want none", p.Name)
			}
		}
	}
}

func TestGetGamesWireCarriesTurnAndLife(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	waiting, err := l.Create("Waiting")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := l.Join(waiting.ID, waiting.InviteToken, "Carol"); err != nil {
		t.Fatalf("Join Carol: %v", err)
	}
	meta, _, _ := startTwoSeatGame(t, l, "Playing")
	finishOpeningRoll(t, l, meta.ID)

	resp := doGet(t, srv, "/games", identifiedToken(t, a, uuid.New()))
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var body struct {
		Games []map[string]json.RawMessage `json:"games"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("decode: %v (%s)", err, raw)
	}
	seen := 0
	for _, g := range body.Games {
		var id uuid.UUID
		_ = json.Unmarshal(g["id"], &id)
		var players []map[string]json.RawMessage
		_ = json.Unmarshal(g["players"], &players)
		switch id {
		case meta.ID:
			seen++
			if string(g["turn"]) != "1" {
				t.Errorf("active game: turn = %s, want 1", g["turn"])
			}
			for _, p := range players {
				if !bytes.Equal(p["life"], []byte("40")) {
					t.Errorf("active game seat %s: life = %s, want 40", p["name"], p["life"])
				}
			}
		case waiting.ID:
			seen++
			if _, ok := g["turn"]; ok {
				t.Errorf("lobby table carries turn %s, want none", g["turn"])
			}
			for _, p := range players {
				if _, ok := p["life"]; ok {
					t.Errorf("lobby table seat %s carries life, want none", p["name"])
				}
			}
		}
	}
	if seen != 2 {
		t.Fatalf("saw %d of the 2 games in GET /games: %s", seen, raw)
	}
}
