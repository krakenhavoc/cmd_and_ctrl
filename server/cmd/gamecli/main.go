// Command gamecli is a thin WebSocket client for driving the
// cmd_and_ctrl server via the v0 action protocol. It connects, prints
// the initial snapshot, then either reads a scripted list of actions
// from a JSON file or accepts action JSON on stdin (one object per
// line). After each action it prints the resulting snapshot.
//
// This is a developer tool. The main consumer is the server's own
// end-to-end test in internal/ws/e2e_test.go, which dials directly
// rather than executing this binary, but the binary is useful for
// manual poking.
//
// Usage:
//
//	gamecli                          # interactive mode, read stdin
//	gamecli -script path/to/turn.json  # replay a scripted turn
//	gamecli -addr ws://host:8080/ws
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// script is the JSON shape the -script flag loads. Each entry becomes
// one action frame sent to the server, in order.
type script struct {
	Actions []scriptedAction `json:"actions"`
}

type scriptedAction struct {
	Type   string          `json:"type"`
	Player string          `json:"player,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
}

func main() {
	addr := flag.String("addr", "ws://localhost:8080/ws", "server WebSocket URL")
	scriptPath := flag.String("script", "", "path to a JSON file with an actions array (leave empty for stdin mode)")
	timeout := flag.Duration("timeout", 10*time.Second, "per-action response timeout")
	flag.Parse()

	conn, _, err := websocket.DefaultDialer.Dial(*addr, nil)
	if err != nil {
		log.Fatalf("dial %s: %v", *addr, err)
	}
	defer func() { _ = conn.Close() }()

	// Consume the initial snapshot the server sends on connect.
	initial, err := readSnapshot(conn, *timeout)
	if err != nil {
		log.Fatalf("read initial snapshot: %v", err)
	}
	fmt.Fprintln(os.Stderr, "--- initial snapshot ---")
	printSnapshot(initial)

	source, err := openActionSource(*scriptPath)
	if err != nil {
		log.Fatalf("open action source: %v", err)
	}
	defer func() { _ = source.Close() }()

	dec := json.NewDecoder(source)
	for {
		var a scriptedAction
		if err := dec.Decode(&a); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("decode action: %v", err)
		}
		if err := sendAction(conn, a); err != nil {
			log.Fatalf("send action %q: %v", a.Type, err)
		}
		snap, err := waitForSnapshot(conn, *timeout)
		if err != nil {
			log.Fatalf("wait for snapshot after %q: %v", a.Type, err)
		}
		fmt.Fprintf(os.Stderr, "--- snapshot after %s ---\n", a.Type)
		printSnapshot(snap)
	}
}

// openActionSource returns the reader the CLI pulls action JSON from.
// If scriptPath is non-empty and points at a file shaped like the
// `script` struct, openActionSource expands it into a sequence of
// action objects and returns a reader wrapping the buffer. Otherwise
// it returns stdin directly.
func openActionSource(scriptPath string) (io.ReadCloser, error) {
	if scriptPath == "" {
		return os.Stdin, nil
	}
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}
	var s script
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("parse script: %w", err)
	}
	// Re-serialise each action as its own line so the main loop can
	// decode them sequentially via a single json.Decoder.
	buf := &stringReadCloser{}
	for _, a := range s.Actions {
		line, err := json.Marshal(a)
		if err != nil {
			return nil, err
		}
		buf.data = append(buf.data, line...)
		buf.data = append(buf.data, '\n')
	}
	return buf, nil
}

// stringReadCloser is a tiny in-memory ReadCloser wrapping a byte
// slice. Just enough to back openActionSource without pulling in an
// external library.
type stringReadCloser struct {
	data []byte
	off  int
}

func (s *stringReadCloser) Read(p []byte) (int, error) {
	if s.off >= len(s.data) {
		return 0, io.EOF
	}
	n := copy(p, s.data[s.off:])
	s.off += n
	return n, nil
}

func (s *stringReadCloser) Close() error { return nil }

func sendAction(conn *websocket.Conn, a scriptedAction) error {
	payload, err := json.Marshal(protocol.ActionPayload{
		Type:   a.Type,
		Player: a.Player,
		Params: a.Params,
	})
	if err != nil {
		return err
	}
	frame, err := json.Marshal(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindAction,
		ID:      uuid.New().String(),
		Payload: payload,
	})
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, frame)
}

// waitForSnapshot loops through inbound frames, returning the next
// snapshot or failing fast if an error frame comes through first.
// Times out after d.
func waitForSnapshot(conn *websocket.Conn, d time.Duration) (protocol.SnapshotPayload, error) {
	deadline := time.Now().Add(d)
	if err := conn.SetReadDeadline(deadline); err != nil {
		return protocol.SnapshotPayload{}, err
	}
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return protocol.SnapshotPayload{}, err
		}
		var frame protocol.Frame
		if err := json.Unmarshal(raw, &frame); err != nil {
			return protocol.SnapshotPayload{}, fmt.Errorf("decode frame: %w", err)
		}
		switch frame.Kind {
		case protocol.KindSnapshot:
			var snap protocol.SnapshotPayload
			if err := json.Unmarshal(frame.Payload, &snap); err != nil {
				return protocol.SnapshotPayload{}, err
			}
			return snap, nil
		case protocol.KindError:
			var ep protocol.ErrorPayload
			_ = json.Unmarshal(frame.Payload, &ep)
			return protocol.SnapshotPayload{}, fmt.Errorf("server error: %s — %s", ep.Code, ep.Message)
		default:
			// Ignore other frame kinds (e.g. pongs in flight).
		}
	}
}

func readSnapshot(conn *websocket.Conn, d time.Duration) (protocol.SnapshotPayload, error) {
	return waitForSnapshot(conn, d)
}

// printSnapshot writes a human-readable one-pager of the snapshot to
// stdout. Not a stable format — the gamecli is a dev tool, not a
// programmatic consumer.
func printSnapshot(s protocol.SnapshotPayload) {
	fmt.Printf("seq=%d game=%s state=%s turn=%d seat=%d phase=%s step=%s\n",
		s.Seq, s.Game.ID, s.Game.State,
		s.Game.Turn.Number, s.Game.Turn.ActiveSeat, s.Game.Turn.Phase, s.Game.Turn.Step)
	for _, p := range s.Game.Seats {
		fmt.Printf("  seat %d %s life=%d hand=%d library=%d graveyard=%d command=%d\n",
			p.Seat, p.Name, p.Life, p.Hand.Count, p.Library.Count, p.Graveyard.Count, p.Command.Count)
	}
	fmt.Printf("  battlefield=%d stack=%d exile=%d\n",
		s.Game.Battlefield.Count, s.Game.Stack.Count, s.Game.Exile.Count)
	for _, choice := range s.Game.PendingChoices {
		if choice.Kind == "coin_call" {
			answers := "heads or tails"
			if choice.AllowStop {
				answers += " or stop"
			}
			fmt.Printf("  coin_call %s chooser=%s coins=%d: %s (resolve_choice with choice_id and call: %s)\n",
				choice.ID, choice.Chooser, choice.Coins, choice.Reason, answers)
		}
	}
}
