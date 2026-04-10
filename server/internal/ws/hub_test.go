package ws

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// newTestServer stands up an httptest server with a hub mounted at /ws
// and returns the ws:// URL and a cleanup function. The hub is silenced
// via a discard logger to keep test output clean.
func newTestServer(t *testing.T) (string, func()) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(log)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	return wsURL, srv.Close
}

func dial(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	return conn
}

func sendFrame(t *testing.T, conn *websocket.Conn, f protocol.Frame) {
	t.Helper()
	raw, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal frame: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, raw); err != nil {
		t.Fatalf("ws write: %v", err)
	}
}

func readFrame(t *testing.T, conn *websocket.Conn) protocol.Frame {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ws read: %v", err)
	}
	var f protocol.Frame
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}
	return f
}

// TestPingRoundTrip is the S01 exit criteria, codified: client sends a
// ping frame, server replies with a pong frame carrying the same id and
// echoed msg. If this test passes, the seam works.
func TestPingRoundTrip(t *testing.T) {
	wsURL, cleanup := newTestServer(t)
	defer cleanup()

	conn := dial(t, wsURL)
	defer conn.Close()

	payload, err := json.Marshal(protocol.PingPayload{Msg: "hello"})
	if err != nil {
		t.Fatalf("marshal ping payload: %v", err)
	}
	sendFrame(t, conn, protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindPing,
		ID:      "test-id-1",
		Payload: payload,
	})

	got := readFrame(t, conn)
	if got.Kind != protocol.KindPong {
		t.Errorf("kind: got %q, want %q", got.Kind, protocol.KindPong)
	}
	if got.ID != "test-id-1" {
		t.Errorf("id: got %q, want %q", got.ID, "test-id-1")
	}
	if got.V != protocol.Version {
		t.Errorf("v: got %d, want %d", got.V, protocol.Version)
	}
	var pong protocol.PongPayload
	if err := json.Unmarshal(got.Payload, &pong); err != nil {
		t.Fatalf("unmarshal pong payload: %v", err)
	}
	if pong.Msg != "hello" {
		t.Errorf("pong msg: got %q, want %q", pong.Msg, "hello")
	}
	if pong.ServerTime == "" {
		t.Error("pong server_time is empty")
	}
	if _, err := time.Parse(time.RFC3339, pong.ServerTime); err != nil {
		t.Errorf("pong server_time %q is not RFC3339: %v", pong.ServerTime, err)
	}
}

func TestBadVersionIsRejected(t *testing.T) {
	wsURL, cleanup := newTestServer(t)
	defer cleanup()

	conn := dial(t, wsURL)
	defer conn.Close()

	sendFrame(t, conn, protocol.Frame{
		V:    999,
		Kind: protocol.KindPing,
		ID:   "test-id-2",
	})

	got := readFrame(t, conn)
	if got.Kind != protocol.KindError {
		t.Fatalf("kind: got %q, want %q", got.Kind, protocol.KindError)
	}
	var errPayload protocol.ErrorPayload
	if err := json.Unmarshal(got.Payload, &errPayload); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	if errPayload.Code != protocol.CodeBadVersion {
		t.Errorf("error code: got %q, want %q", errPayload.Code, protocol.CodeBadVersion)
	}
}

func TestUnknownKindIsRejected(t *testing.T) {
	wsURL, cleanup := newTestServer(t)
	defer cleanup()

	conn := dial(t, wsURL)
	defer conn.Close()

	sendFrame(t, conn, protocol.Frame{
		V:    protocol.Version,
		Kind: "explode",
		ID:   "test-id-3",
	})

	got := readFrame(t, conn)
	if got.Kind != protocol.KindError {
		t.Fatalf("kind: got %q, want %q", got.Kind, protocol.KindError)
	}
	var errPayload protocol.ErrorPayload
	if err := json.Unmarshal(got.Payload, &errPayload); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	if errPayload.Code != protocol.CodeBadRequest {
		t.Errorf("error code: got %q, want %q", errPayload.Code, protocol.CodeBadRequest)
	}
}

func TestBadJSONIsRejected(t *testing.T) {
	wsURL, cleanup := newTestServer(t)
	defer cleanup()

	conn := dial(t, wsURL)
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("{not json")); err != nil {
		t.Fatalf("ws write: %v", err)
	}

	got := readFrame(t, conn)
	if got.Kind != protocol.KindError {
		t.Fatalf("kind: got %q, want %q", got.Kind, protocol.KindError)
	}
	var errPayload protocol.ErrorPayload
	if err := json.Unmarshal(got.Payload, &errPayload); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	if errPayload.Code != protocol.CodeBadJSON {
		t.Errorf("error code: got %q, want %q", errPayload.Code, protocol.CodeBadJSON)
	}
}

// TestShutdownEmptyHubIsImmediate regresses a bug where Hub.Shutdown
// unconditionally blocked until its context deadline even when no
// clients were connected, making the dev loop feel sluggish.
func TestShutdownEmptyHubIsImmediate(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(log)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	hub.Shutdown(ctx)
	elapsed := time.Since(start)

	if elapsed > 200*time.Millisecond {
		t.Errorf("empty Shutdown took %s, want <200ms", elapsed)
	}
}

// TestServeWSRejectsAfterShutdown guards the closed-flag regression:
// once Shutdown has been called, subsequent WebSocket upgrade attempts
// must be refused with 503 rather than joining a tearing-down hub.
func TestServeWSRejectsAfterShutdown(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(log)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	hub.Shutdown(ctx)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatalf("dial after shutdown: expected error, got nil")
	}
	if resp == nil {
		t.Fatalf("dial after shutdown: no HTTP response on the failed upgrade")
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
}

// TestShutdownWithClientsWaitsForPumps verifies Shutdown waits for the
// read/write pumps of connected clients to exit, then returns.
func TestShutdownWithClientsWaitsForPumps(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(log)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	// Connect three clients.
	var conns []*websocket.Conn
	for range 3 {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("ws dial: %v", err)
		}
		conns = append(conns, conn)
	}
	defer func() {
		for _, c := range conns {
			_ = c.Close()
		}
	}()

	// Give the server a moment to register all three.
	deadline := time.Now().Add(time.Second)
	for hub.Count() < 3 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := hub.Count(); got != 3 {
		t.Fatalf("hub count: got %d, want 3", got)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	hub.Shutdown(ctx)
	elapsed := time.Since(start)

	// Should be well under the 2s timeout — the clients are local and
	// shutting down is fast.
	if elapsed > 1*time.Second {
		t.Errorf("Shutdown took %s, want <1s", elapsed)
	}
}
