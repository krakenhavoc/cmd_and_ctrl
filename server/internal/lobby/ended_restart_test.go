package lobby

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// #2919: players stay on an ended table until they leave it. A deploy
// restarts the server several times a day, and every connected client
// redials after one. If the ended table did not come back, every seat
// still looking at the final board would lose it on the next deploy,
// and a page reload afterwards would find "game not found".

// TestEndedTableSurvivesARestart ends a game, restarts the process over
// the same data directory, and dials the table as each seat that was on
// it: the one who conceded (eliminated) and the winner.
func TestEndedTableSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, alice, bob := startTwoSeatGame(t, l, "Ended")

	room := l.RoomOf(meta.ID)
	if _, _, err := room.Apply(alice, func() error { return room.Game.Concede(alice) }); err != nil {
		t.Fatalf("concede: %v", err)
	}
	if room.Game.CurrentState() != game.StateEnded {
		t.Fatalf("setup: state = %q, want ended", room.Game.CurrentState())
	}
	seq := room.Seq()

	// --- the deploy -----------------------------------------------
	l2, mgr2 := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1: the ended table did not come back", n)
	}
	back, err := l2.Get(meta.ID)
	if err != nil {
		t.Fatalf("restored lobby does not know the ended table: %v", err)
	}
	if back.State != string(game.StateEnded) {
		t.Errorf("state = %q, want ended", back.State)
	}
	room2 := l2.RoomOf(meta.ID)
	if room2 == nil {
		t.Fatal("no room for the ended table after the restart")
	}
	if got := room2.Seq(); got != seq {
		t.Errorf("seq = %d, want %d: a client already holding the final frame must not be rewound", got, seq)
	}

	hub := ws.NewHub(quietLogger())
	hub.SetManager(mgr2)
	srv := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { hub.Shutdown(context.Background()) })

	seats := []struct{ name, id string }{
		{"alice (conceded)", alice.String()},
		{"bob (won)", bob.String()},
	}
	for _, s := range seats {
		url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/?game=" + meta.ID.String() + "&player=" + s.id
		conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			status := 0
			if resp != nil {
				status = resp.StatusCode
			}
			t.Fatalf("%s: dial after restart failed (status %d): %v", s.name, status, err)
		}
		var f protocol.Frame
		if err := conn.ReadJSON(&f); err != nil {
			t.Fatalf("%s: read first frame: %v", s.name, err)
		}
		if f.Kind != protocol.KindSnapshot {
			t.Fatalf("%s: first frame kind = %q, want snapshot", s.name, f.Kind)
		}
		var snap struct {
			Game struct {
				State string `json:"state"`
			} `json:"game"`
		}
		if err := json.Unmarshal(f.Payload, &snap); err != nil {
			t.Fatalf("%s: decode snapshot: %v", s.name, err)
		}
		if snap.Game.State != "ended" {
			t.Errorf("%s: snapshot state = %q, want ended", s.name, snap.Game.State)
		}
		_ = conn.Close()
	}
}
