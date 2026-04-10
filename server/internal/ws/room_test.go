package ws

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// newRoomTestServer stands up a hub with a seeded demo room and a
// crash-recovery directory in t.TempDir(). Returns the ws:// URL, the
// seeded game (for test assertions), the temp dir, and a cleanup
// function.
func newRoomTestServer(t *testing.T) (string, *game.Game, string, func()) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	g := seedTestGame(t)
	tmpDir := t.TempDir()
	room := NewRoom(g, log, tmpDir)
	hub := NewHub(log)
	hub.SetRoom(room)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	return wsURL, g, tmpDir, srv.Close
}

func seedTestGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	// Deterministic shuffle so golden file tests are reproducible.
	if err := g.Start(rand.New(rand.NewPCG(42, 42))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g
}

func readSnapshotFrame(t *testing.T, conn *websocket.Conn) protocol.SnapshotPayload {
	t.Helper()
	f := readFrame(t, conn)
	if f.Kind != protocol.KindSnapshot {
		t.Fatalf("kind: got %q, want %q", f.Kind, protocol.KindSnapshot)
	}
	var snap protocol.SnapshotPayload
	if err := json.Unmarshal(f.Payload, &snap); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	return snap
}

func sendActionFrame(t *testing.T, conn *websocket.Conn, a protocol.ActionPayload) string {
	t.Helper()
	id := uuid.New().String()
	payload, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal action payload: %v", err)
	}
	sendFrame(t, conn, protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindAction,
		ID:      id,
		Payload: payload,
	})
	return id
}

func TestRoomInitialSnapshotOnConnect(t *testing.T) {
	wsURL, g, _, cleanup := newRoomTestServer(t)
	defer cleanup()

	conn := dial(t, wsURL)
	defer conn.Close()

	snap := readSnapshotFrame(t, conn)
	// Snapshot reuses the current seq (no bump) — a freshly-created
	// room has seen no actions yet, so seq=0 is the expected initial
	// value. The first Apply will bump to seq=1.
	if snap.Seq != 0 {
		t.Errorf("initial snapshot seq: got %d, want 0", snap.Seq)
	}
	if snap.Game.ID != g.ID.String() {
		t.Errorf("game id: got %q, want %q", snap.Game.ID, g.ID.String())
	}
	if len(snap.Game.Seats) != 2 {
		t.Errorf("seats: got %d, want 2", len(snap.Game.Seats))
	}
	if snap.Game.Turn.Step != "untap" {
		t.Errorf("turn step: got %q, want untap", snap.Game.Turn.Step)
	}
}

func TestRoomActionDrawCardBroadcasts(t *testing.T) {
	wsURL, g, _, cleanup := newRoomTestServer(t)
	defer cleanup()

	conn := dial(t, wsURL)
	defer conn.Close()

	// Consume initial snapshot. seq=0 (no actions yet, Snapshot doesn't bump).
	initial := readSnapshotFrame(t, conn)
	if initial.Seq != 0 {
		t.Fatalf("initial seq: got %d, want 0", initial.Seq)
	}

	// Send a draw_card action for seat 0.
	seat0ID := g.Seats[0].ID.String()
	sendActionFrame(t, conn, protocol.ActionPayload{
		Type:   "draw_card",
		Player: seat0ID,
	})

	// First Apply bumps seq to exactly 1.
	after := readSnapshotFrame(t, conn)
	if after.Seq != 1 {
		t.Errorf("seq after first action: got %d, want 1", after.Seq)
	}
	if after.Game.Seats[0].Hand.Count != 1 {
		t.Errorf("hand count: got %d, want 1", after.Game.Seats[0].Hand.Count)
	}
}

func TestRoomActionBroadcastsToAllClients(t *testing.T) {
	wsURL, g, _, cleanup := newRoomTestServer(t)
	defer cleanup()

	// Two clients. c2's connect sends its initial snapshot only to
	// c2 (not a broadcast), so c1 does not receive an extra frame on
	// c2's connect.
	c1 := dial(t, wsURL)
	defer c1.Close()
	readSnapshotFrame(t, c1) // c1 initial

	c2 := dial(t, wsURL)
	defer c2.Close()
	readSnapshotFrame(t, c2) // c2 initial (targeted, c1 sees nothing)

	// Action from c1 should be broadcast to BOTH c1 and c2.
	sendActionFrame(t, c1, protocol.ActionPayload{
		Type:   "draw_card",
		Player: g.Seats[0].ID.String(),
	})
	snap1 := readSnapshotFrame(t, c1)
	snap2 := readSnapshotFrame(t, c2)

	if snap1.Seq != snap2.Seq {
		t.Errorf("seq mismatch across clients: c1=%d c2=%d", snap1.Seq, snap2.Seq)
	}
	if snap1.Game.Seats[0].Hand.Count != 1 || snap2.Game.Seats[0].Hand.Count != 1 {
		t.Errorf("hand count not reflected in both snapshots")
	}
}

func TestRoomActionUnknownTypeReturnsError(t *testing.T) {
	wsURL, _, _, cleanup := newRoomTestServer(t)
	defer cleanup()

	conn := dial(t, wsURL)
	defer conn.Close()
	readSnapshotFrame(t, conn) // initial

	id := sendActionFrame(t, conn, protocol.ActionPayload{
		Type: "explode",
	})

	f := readFrame(t, conn)
	if f.Kind != protocol.KindError {
		t.Fatalf("kind: got %q, want %q", f.Kind, protocol.KindError)
	}
	if f.ID != id {
		t.Errorf("error id: got %q, want %q", f.ID, id)
	}
	var ep protocol.ErrorPayload
	if err := json.Unmarshal(f.Payload, &ep); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if ep.Code != protocol.CodeBadRequest {
		t.Errorf("code: got %q, want %q", ep.Code, protocol.CodeBadRequest)
	}
}

func TestRoomActionMissingPayloadRejected(t *testing.T) {
	wsURL, _, _, cleanup := newRoomTestServer(t)
	defer cleanup()

	conn := dial(t, wsURL)
	defer conn.Close()
	readSnapshotFrame(t, conn) // initial

	// Action frame with no payload at all.
	id := uuid.New().String()
	sendFrame(t, conn, protocol.Frame{
		V:    protocol.Version,
		Kind: protocol.KindAction,
		ID:   id,
	})

	f := readFrame(t, conn)
	if f.Kind != protocol.KindError {
		t.Fatalf("kind: got %q, want %q", f.Kind, protocol.KindError)
	}
	var ep protocol.ErrorPayload
	_ = json.Unmarshal(f.Payload, &ep)
	if ep.Code != protocol.CodeBadRequest {
		t.Errorf("code: got %q, want %q", ep.Code, protocol.CodeBadRequest)
	}
}

func TestRoomActionFailsOnHubWithoutRoom(t *testing.T) {
	// A hub without a room attached should reject action frames.
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(log)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	conn := dial(t, wsURL)
	defer conn.Close()

	id := sendActionFrame(t, conn, protocol.ActionPayload{Type: "draw_card"})
	f := readFrame(t, conn)
	if f.Kind != protocol.KindError {
		t.Fatalf("kind: got %q, want %q", f.Kind, protocol.KindError)
	}
	if f.ID != id {
		t.Errorf("id: got %q, want %q", f.ID, id)
	}
}

func TestRoomCrashRecoveryDumpsToDisk(t *testing.T) {
	wsURL, g, tmpDir, cleanup := newRoomTestServer(t)
	defer cleanup()

	conn := dial(t, wsURL)
	defer conn.Close()

	// Initial snapshot on connect should have written a dump already.
	readSnapshotFrame(t, conn)

	// Trigger an action so we have a second snapshot written.
	sendActionFrame(t, conn, protocol.ActionPayload{
		Type:   "draw_card",
		Player: g.Seats[0].ID.String(),
	})
	readSnapshotFrame(t, conn)

	path := filepath.Join(tmpDir, "games", g.ID.String()+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read dump file: %v", err)
	}
	var snap protocol.SnapshotPayload
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("unmarshal dump: %v", err)
	}
	if snap.Game.ID != g.ID.String() {
		t.Errorf("dumped game id: got %q, want %q", snap.Game.ID, g.ID.String())
	}
	// After 1 draw, the latest snapshot should show hand count 1.
	if snap.Game.Seats[0].Hand.Count != 1 {
		t.Errorf("dumped hand count: got %d, want 1", snap.Game.Seats[0].Hand.Count)
	}
}

// TestRoomInitialSnapshotOnlyGoesToJoiner regresses the C1 fix:
// the initial snapshot sent on connect must not be broadcast to
// already-connected clients. A regression that replaced the targeted
// client.sendRaw with hub.broadcastAll would double-bump the seq
// counter for every connection and make existing clients re-render
// on every new join.
func TestRoomInitialSnapshotOnlyGoesToJoiner(t *testing.T) {
	wsURL, _, _, cleanup := newRoomTestServer(t)
	defer cleanup()

	c1 := dial(t, wsURL)
	defer c1.Close()
	readSnapshotFrame(t, c1) // c1's initial

	// Connect a second client. c1 must NOT receive a second snapshot
	// just because c2 joined.
	c2 := dial(t, wsURL)
	defer c2.Close()
	readSnapshotFrame(t, c2) // c2's initial

	// Try to read from c1 with a short deadline. Nothing should be
	// sitting in its buffer: the new-join did not broadcast.
	_ = c1.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, raw, err := c1.ReadMessage()
	if err == nil {
		t.Errorf("c1 received unexpected frame after c2 joined: %s", raw)
	}
}

// TestRoomActionSeqIsMonotonicAcrossActions regresses the Room.Apply
// seq discipline: four successive actions must produce snapshots with
// seq 1, 2, 3, 4 — no gaps, no duplicates, no regressions.
func TestRoomActionSeqIsMonotonicAcrossActions(t *testing.T) {
	wsURL, g, _, cleanup := newRoomTestServer(t)
	defer cleanup()

	conn := dial(t, wsURL)
	defer conn.Close()

	initial := readSnapshotFrame(t, conn)
	if initial.Seq != 0 {
		t.Fatalf("initial seq: got %d, want 0", initial.Seq)
	}

	seat0 := g.Seats[0].ID.String()
	for i := 1; i <= 4; i++ {
		sendActionFrame(t, conn, protocol.ActionPayload{
			Type:   "draw_card",
			Player: seat0,
		})
		got := readSnapshotFrame(t, conn)
		if got.Seq != uint64(i) {
			t.Errorf("action %d: seq got %d, want %d", i, got.Seq, i)
		}
	}
}

// TestRoomActionFrameVersionRejected regresses the same frame-version
// guard as TestBadVersionIsRejected but for an action frame instead
// of a ping. A refactor that moved the version check after the Kind
// switch would let a v=999 action land on the game.
func TestRoomActionFrameVersionRejected(t *testing.T) {
	wsURL, g, _, cleanup := newRoomTestServer(t)
	defer cleanup()

	conn := dial(t, wsURL)
	defer conn.Close()
	readSnapshotFrame(t, conn) // initial

	payload, err := json.Marshal(protocol.ActionPayload{
		Type:   "draw_card",
		Player: g.Seats[0].ID.String(),
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	// Send a v=999 action frame — must be rejected.
	sendFrame(t, conn, protocol.Frame{
		V:       999,
		Kind:    protocol.KindAction,
		ID:      "test-bad-v",
		Payload: payload,
	})
	f := readFrame(t, conn)
	if f.Kind != protocol.KindError {
		t.Fatalf("kind: got %q, want %q", f.Kind, protocol.KindError)
	}
	var ep protocol.ErrorPayload
	_ = json.Unmarshal(f.Payload, &ep)
	if ep.Code != protocol.CodeBadVersion {
		t.Errorf("code: got %q, want %q", ep.Code, protocol.CodeBadVersion)
	}
}

// TestClassifyActionErrorGameNotActive regresses the wire-code
// classification policy: an action against a non-active game reports
// CodeInternal (server state problem), not CodeBadRequest (client
// fault).
func TestClassifyActionErrorGameNotActive(t *testing.T) {
	// Use a Room whose game is in Ended state.
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	g := seedTestGame(t)
	g.End()

	room := NewRoom(g, log, "")
	hub := NewHub(log)
	hub.SetRoom(room)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	conn := dial(t, wsURL)
	defer conn.Close()
	readSnapshotFrame(t, conn) // initial (game is Ended but Snapshot still works)

	// Any action should fail with ErrGameNotActive → CodeInternal.
	sendActionFrame(t, conn, protocol.ActionPayload{
		Type:   "draw_card",
		Player: g.Seats[0].ID.String(),
	})
	f := readFrame(t, conn)
	if f.Kind != protocol.KindError {
		t.Fatalf("kind: got %q, want %q", f.Kind, protocol.KindError)
	}
	var ep protocol.ErrorPayload
	_ = json.Unmarshal(f.Payload, &ep)
	if ep.Code != protocol.CodeInternal {
		t.Errorf("code: got %q, want %q (ErrGameNotActive should classify as internal)", ep.Code, protocol.CodeInternal)
	}
}

func TestRoomDisabledDumpDir(t *testing.T) {
	// With dumpDir == "", crash recovery is disabled and no file is written.
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	g := seedTestGame(t)
	room := NewRoom(g, log, "")
	hub := NewHub(log)
	hub.SetRoom(room)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	conn := dial(t, wsURL)
	defer conn.Close()
	readSnapshotFrame(t, conn)

	// Nothing to assert on disk — success means no crash. The goal is
	// that BuildSnapshotFrame doesn't fail when dumpDir is empty.
}
