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

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// burstFeed is a Feed that sends numbered states: n of them before
// Attach returns, and the rest from its own goroutine once the test
// asks for them.
type burstFeed struct {
	atAttach int
	out      chan FeedSink
}

type burstState struct {
	N int `json:"n"`
}

func (f *burstFeed) Attach(out FeedSink) (func(), error) {
	for i := range f.atAttach {
		out.SendFrame(protocol.KindDraft, burstState{N: i})
	}
	f.out <- out
	return func() {}, nil
}

func (f *burstFeed) Handle(protocol.Frame) *protocol.ErrorPayload { return nil }

// newFeedServer mounts one burstFeed at /feed on a fresh hub.
func newFeedServer(t *testing.T, feed *burstFeed) string {
	t.Helper()
	hub := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /feed", func(w http.ResponseWriter, r *http.Request) {
		hub.ServeFeed(w, r, feed, FeedBinding{UserID: uuid.New(), IssuedAt: time.Now(), Seated: true, Label: "test"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		hub.Shutdown(ctx)
		srv.Close()
	})
	return "ws" + strings.TrimPrefix(srv.URL, "http") + "/feed"
}

// readState reads frames until a feed state arrives, failing on a
// closed socket: the hub dropped the client.
func readState(t *testing.T, conn *websocket.Conn) int {
	t.Helper()
	for {
		f := readRawFrame(t, conn)
		if f.Kind != protocol.KindDraft {
			continue
		}
		var s burstState
		if err := json.Unmarshal(f.Payload, &s); err != nil {
			t.Fatalf("state payload: %v", err)
		}
		return s.N
	}
}

// A feed's frames are states, and only the newest matters (#5698). More
// of them than the send buffer holds, sent before the writer has run at
// all, used to fill the 16-frame buffer and disconnect the socket as a
// slow client. They are coalesced: the socket's first frame is the
// newest state, and the socket stays up.
func TestAFeedBurstBeforeTheWriterRunsIsCoalescedNotDropped(t *testing.T) {
	feed := &burstFeed{atAttach: 40, out: make(chan FeedSink, 1)}
	conn := dial(t, newFeedServer(t, feed))
	defer conn.Close()
	<-feed.out
	if n := readState(t, conn); n != 39 {
		t.Fatalf("the first frame is state %d, want the newest, 39", n)
	}
	// Still connected: a ping is answered.
	sendFrame(t, conn, protocol.Frame{V: protocol.Version, Kind: protocol.KindPing, ID: "p", Payload: json.RawMessage(`{"msg":"x"}`)})
	if f := readRawFrame(t, conn); f.Kind != protocol.KindPong {
		t.Fatalf("after the burst: %q, want a pong", f.Kind)
	}
}

// A feed that sends far faster than a slow reader reads keeps the
// socket: the reader sees states that only rise and ends on the newest.
func TestAFeedOutpacingASlowReaderKeepsTheSocket(t *testing.T) {
	feed := &burstFeed{atAttach: 1, out: make(chan FeedSink, 1)}
	conn := dial(t, newFeedServer(t, feed))
	defer conn.Close()
	out := <-feed.out
	if n := readState(t, conn); n != 0 {
		t.Fatalf("first state %d, want 0", n)
	}
	const last = 2000
	go func() {
		for i := 1; i <= last; i++ {
			out.SendFrame(protocol.KindDraft, burstState{N: i})
		}
	}()
	seen, prev := 0, 0
	for prev < last {
		time.Sleep(time.Millisecond)
		n := readState(t, conn)
		if n <= prev {
			t.Fatalf("state %d after %d: states must only rise", n, prev)
		}
		prev = n
		seen++
	}
	t.Logf("%d states sent, %d written to a reader that sleeps 1 ms a frame", last, seen)
}
