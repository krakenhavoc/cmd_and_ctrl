package ws

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestShutdownDeliversCloseFrameBeforeDroppingConnection is the
// server-side half of #518's remaining acceptance criteria. Before the
// fix, hub.Shutdown wrote the close control frame with WriteControl,
// discarded any error, and called conn.Close() on the very next line —
// racing the close frame's delivery against the FIN/RST the immediate
// Close sends on the same socket. Whether a client observed a clean
// 1001 (CloseGoingAway) or an abnormal 1006 depended on that race, so
// the same binary behaved differently deploy to deploy.
//
// This test pins the contract Shutdown must now uphold: every client
// still connected when Shutdown runs gets a clean close(1001), never a
// bare abnormal closure, even with several clients closing
// concurrently.
func TestShutdownDeliversCloseFrameBeforeDroppingConnection(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(log)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	const numClients = 5
	var conns []*websocket.Conn
	for range numClients {
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

	deadline := time.Now().Add(time.Second)
	for hub.Count() < numClients && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := hub.Count(); got != numClients {
		t.Fatalf("hub count: got %d, want %d", got, numClients)
	}

	// Each client reads on its own goroutine so a blocked read on one
	// connection cannot delay — or mask a failure on — another.
	results := make([]error, numClients)
	var wg sync.WaitGroup
	for i, c := range conns {
		wg.Add(1)
		go func(i int, c *websocket.Conn) {
			defer wg.Done()
			_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
			_, _, err := c.ReadMessage()
			results[i] = err
		}(i, c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	hub.Shutdown(ctx)

	wg.Wait()

	for i, err := range results {
		if err == nil {
			t.Errorf("client %d: ReadMessage returned nil error, want a close error", i)
			continue
		}
		if !websocket.IsCloseError(err, websocket.CloseGoingAway) {
			t.Errorf("client %d: got %v, want a clean close (1001 CloseGoingAway), not an abnormal closure", i, err)
		}
	}
}

// TestShutdownStillHonoursCtxWhenWriteBlocks guards the ctx-bound
// escape hatch closeAndDrop relies on: if a client's own
// closeGracePeriod wait would outlast the caller's deadline, Shutdown
// must still return promptly rather than waiting out the full grace
// period on every connection.
func TestShutdownStillHonoursCtxWhenWriteBlocks(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(log)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(time.Second)
	for hub.Count() < 1 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	// A ctx that is already effectively expired: closeAndDrop's grace
	// wait must select on ctx.Done() immediately rather than blocking
	// for the full closeGracePeriod.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	time.Sleep(5 * time.Millisecond)

	start := time.Now()
	hub.Shutdown(ctx)
	elapsed := time.Since(start)

	if elapsed > closeGracePeriod {
		t.Errorf("Shutdown with an expired ctx took %s, want well under closeGracePeriod (%s)", elapsed, closeGracePeriod)
	}
}
