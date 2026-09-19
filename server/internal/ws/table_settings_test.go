package ws

// table_settings_test.go — ADR 0075 §2.3, sub-PR 3. The WebSocket
// half of the settings surface: who may send `set_table_settings`
// (and its deprecated alias `set_undo_limit`), and what the change
// does NOT do, which is mint an undo entry.
//
// The second property is the one worth a test of its own. A settings
// change that went through Room.Apply would be undoable, and then
// lowering the undo limit could be taken back with the undo it was
// meant to stop — the exact loop the ADR names.

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// settingsAction builds a set_table_settings frame body from raw JSON,
// which is how a client sends it: the params object IS the patch.
func settingsAction(patch string) protocol.ActionPayload {
	return protocol.ActionPayload{Type: "set_table_settings", Params: json.RawMessage(patch)}
}

// expectActionError sends a, reads one frame and fails unless it is an
// error frame. Returns the decoded payload so a caller can assert on
// the message.
func expectActionError(t *testing.T, conn *websocket.Conn, a protocol.ActionPayload) protocol.ErrorPayload {
	t.Helper()
	sendActionFrame(t, conn, a)
	frame := readNextFrame(t, conn)
	if frame.Kind != protocol.KindError {
		t.Fatalf("expected an error frame for %q, got %q", a.Type, frame.Kind)
	}
	var ep protocol.ErrorPayload
	if err := json.Unmarshal(frame.Payload, &ep); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	return ep
}

func allowSpawn(g *game.Game) bool {
	var on bool
	g.ReadSnapshot(func() { on = g.Settings.AllowSpawn })
	return on
}

// TestSetTableSettingsAcceptedFromHost — the happy path. The host
// sends a two-field patch; both land, and the fields the patch did not
// name are untouched.
func TestSetTableSettingsAcceptedFromHost(t *testing.T) {
	wsURL, g, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()

	host := g.Seats[0].ID
	room.SetHost(host)
	conn := dialAs(t, wsURL, host)
	defer conn.Close()
	readNextFrame(t, conn) // initial

	snap := sendActionAndWait(t, conn, settingsAction(`{"undo_limit":-1,"allow_spawn":true}`))
	if n := undoLimit(g); n != game.UndoUnlimited {
		t.Errorf("UndoLimit: got %d, want %d", n, game.UndoUnlimited)
	}
	if !allowSpawn(g) {
		t.Error("AllowSpawn: got false, want true")
	}
	// Absent fields are left alone — that is the whole point of a
	// patch, and the reason the route is a PATCH and the action takes
	// a partial rather than the whole struct.
	var life int
	g.ReadSnapshot(func() { life = g.Settings.StartingLife })
	if life != game.StartingLife {
		t.Errorf("StartingLife: got %d, want %d (the patch did not name it)", life, game.StartingLife)
	}
	if snap.Game.Settings == nil || snap.Game.Settings.UndoLimit != game.UndoUnlimited {
		t.Errorf("broadcast settings: got %+v, want undo_limit=%d", snap.Game.Settings, game.UndoUnlimited)
	}
	// The table hears about it (ADR 0075 §2.3).
	if !hasSettingsLine(snap.Game.Log, "undo_limit") {
		t.Errorf("no settings line in the log: %v", logKinds(snap.Game.Log))
	}
}

// TestSetTableSettingsRefusedForNonHostSeat — a seated player who is
// not the host is refused, and nothing changes. This is the behaviour
// change ADR 0075 exists for: before S35 any seat could dial the undo
// limit in its own favour mid-game.
func TestSetTableSettingsRefusedForNonHostSeat(t *testing.T) {
	wsURL, g, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()

	room.SetHost(g.Seats[0].ID)
	other := g.Seats[1].ID
	conn := dialAs(t, wsURL, other)
	defer conn.Close()
	readNextFrame(t, conn) // initial

	ep := expectActionError(t, conn, settingsAction(`{"undo_limit":9,"allow_spawn":true}`))
	if ep.Code != protocol.CodeBadRequest {
		t.Errorf("code: got %q, want %q", ep.Code, protocol.CodeBadRequest)
	}
	if !strings.Contains(ep.Message, "host") {
		t.Errorf("message %q does not say who may do this", ep.Message)
	}
	if n := undoLimit(g); n != game.DefaultUndoLimit {
		t.Errorf("UndoLimit moved on a refused patch: got %d, want %d", n, game.DefaultUndoLimit)
	}
	if allowSpawn(g) {
		t.Error("AllowSpawn moved on a refused patch")
	}
}

// TestSetUndoLimitRefusedForNonHost — the deprecated alias carries the
// same gate. If it did not, it would be the documented way around the
// new one.
func TestSetUndoLimitRefusedForNonHost(t *testing.T) {
	wsURL, g, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()

	room.SetHost(g.Seats[0].ID)
	conn := dialAs(t, wsURL, g.Seats[1].ID)
	defer conn.Close()
	readNextFrame(t, conn) // initial

	expectActionError(t, conn, protocol.ActionPayload{
		Type: "set_undo_limit", Params: json.RawMessage(`{"limit":7}`),
	})
	if n := undoLimit(g); n != game.DefaultUndoLimit {
		t.Errorf("UndoLimit: got %d, want %d", n, game.DefaultUndoLimit)
	}
}

// TestSetTableSettingsAcceptedFromAdmin — an admin connection manages
// any table, host or no host. It is also the case Binding.PlayerID
// cannot answer on its own, which is why Binding.Admin exists and why
// Client carries it.
func TestSetTableSettingsAcceptedFromAdmin(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Deliberately no SetHost: an admin needs no host to exist.
	_, g, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()

	// A second server over the SAME room, whose authorizer marks the
	// connection as the admin. The e2e harness's default binding
	// resolver has no way to say so.
	hub := NewHub(log)
	hub.SetRoom(room)
	hub.SetAuthorizer(UpgradeAuthorizerFunc(func(r *http.Request) (Binding, error) {
		return Binding{GameID: g.ID, Admin: true}, nil
	}))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conn := dial(t, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws")
	defer conn.Close()
	readNextFrame(t, conn) // initial

	sendActionAndWait(t, conn, settingsAction(`{"allow_spawn":true}`))
	if !allowSpawn(g) {
		t.Error("an admin connection could not change the table settings")
	}
}

// TestSettingsChangeIsNotUndoable — the change mints no undo entry, so
// the undo that follows it takes back the PLAY before it, and the
// setting stays where the host put it.
func TestSettingsChangeIsNotUndoable(t *testing.T) {
	wsURL, g, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()

	host := g.Seats[0].ID
	room.SetHost(host)
	conn := dialAs(t, wsURL, host)
	defer conn.Close()
	readNextFrame(t, conn) // initial

	before := handSize(g, 0)
	sendActionAndWait(t, conn, protocol.ActionPayload{Type: "draw_card", Player: host.String()})
	if n := handSize(g, 0); n != before+1 {
		t.Fatalf("setup draw: hand=%d, want %d", n, before+1)
	}
	sendActionAndWait(t, conn, settingsAction(`{"undo_limit":5}`))

	// One undo. If the settings change had pushed an entry this would
	// pop THAT one and the draw would survive.
	sendActionAndWait(t, conn, protocol.ActionPayload{Type: "undo"})
	if n := handSize(g, 0); n != before {
		t.Errorf("after undo: hand=%d, want %d — the undo popped the settings change, not the draw", n, before)
	}
	if n := undoLimit(g); n != 5 {
		t.Errorf("undo rolled the setting back: UndoLimit=%d, want 5", n)
	}
}

// TestStartingLifeRefusedOnALiveTable — ADR 0075 §2.3's documented
// rejection, over the wire. The other fields in the same patch are NOT
// applied: UpdateSettings validates and state-checks the whole patch
// before it writes any of it.
func TestStartingLifeRefusedOnALiveTable(t *testing.T) {
	wsURL, g, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()

	host := g.Seats[0].ID
	room.SetHost(host)
	conn := dialAs(t, wsURL, host)
	defer conn.Close()
	readNextFrame(t, conn) // initial

	ep := expectActionError(t, conn, settingsAction(`{"starting_life":60,"allow_spawn":true}`))
	if !strings.Contains(ep.Message, "starting life") {
		t.Errorf("message %q does not name the field that was refused", ep.Message)
	}
	if allowSpawn(g) {
		t.Error("a refused patch applied one of its other fields")
	}
}

func hasSettingsLine(entries []protocol.LogEvent, label string) bool {
	for _, e := range entries {
		if e.Kind == protocol.LogSettings && e.Label == label {
			return true
		}
	}
	return false
}

func logKinds(entries []protocol.LogEvent) []protocol.LogKind {
	out := make([]protocol.LogKind, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Kind)
	}
	return out
}
