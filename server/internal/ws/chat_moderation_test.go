package ws

// chat_moderation_test.go: the chat filter and the recent-chat ring a
// player report reads (#2984).

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func sendChat(t *testing.T, conn *websocket.Conn, text string) {
	t.Helper()
	payload, _ := json.Marshal(protocol.ChatPayload{Text: text})
	frame, _ := json.Marshal(protocol.Frame{V: protocol.Version, Kind: protocol.KindChat, ID: uuid.NewString(), Payload: payload})
	if err := conn.WriteMessage(websocket.TextMessage, frame); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestChatMasksAListedWordForEveryoneAndTheRingHoldsWhatTheTableSaw(t *testing.T) {
	wsURL, g, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()
	a := dialAs(t, wsURL, g.Seats[0].ID)
	defer a.Close()
	b := dialAs(t, wsURL, g.Seats[1].ID)
	defer b.Close()
	_ = readSnapshotFrame(t, a)
	_ = readSnapshotFrame(t, b)

	sendChat(t, a, "you absolute r.e.t.a.r.d.e.d player, nice Wrath of God")
	want := "you absolute *************** player, nice Wrath of God"
	for label, conn := range map[string]*websocket.Conn{"sender": a, "other": b} {
		if got := readChatFrame(t, conn); got.Text != want {
			t.Errorf("%s saw %q, want %q", label, got.Text, want)
		}
	}

	sendChat(t, b, "gg")
	_ = readChatFrame(t, a)
	_ = readChatFrame(t, b)

	got := room.RecentChat(10)
	if len(got) != 2 || got[0].Text != want || got[0].AuthorName != g.Seats[0].Name || got[1].Text != "gg" {
		t.Fatalf("ring = %+v", got)
	}
}

func TestRecentChatKeepsTheLastLinesOnly(t *testing.T) {
	_, _, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()
	for i := 0; i < RecentChatLimit+5; i++ {
		room.RecordChat(protocol.ChatPayload{Text: fmt.Sprintf("m%d", i)})
	}
	all := room.RecentChat(1000)
	if len(all) != RecentChatLimit || all[0].Text != "m5" || all[len(all)-1].Text != fmt.Sprintf("m%d", RecentChatLimit+4) {
		t.Fatalf("ring = %d lines, first %q last %q", len(all), all[0].Text, all[len(all)-1].Text)
	}
	if last3 := room.RecentChat(3); len(last3) != 3 || last3[2].Text != all[len(all)-1].Text {
		t.Errorf("RecentChat(3) = %+v", last3)
	}
	// A copy: changing the answer does not change the ring.
	all[0].Text = "tampered"
	if room.RecentChat(1000)[0].Text == "tampered" {
		t.Error("RecentChat returned the ring itself")
	}
}
