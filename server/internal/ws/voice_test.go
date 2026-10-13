package ws

import (
	"io"
	"log/slog"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Tests for voice.go (ADR 0159 §6): the room keeps its own copy of the
// voice state and stamps it on every capture, and the per-seat filter
// reduces it.

func TestRoomStampsVoiceOnEveryCapture(t *testing.T) {
	room := NewRoom(seedTestGame(t), slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	if room.Voice() != nil {
		t.Fatal("a new room has a voice stamp")
	}
	in := &protocol.VoiceView{
		Enabled: true, Status: protocol.VoiceStatusReady, URL: "https://discord.com/channels/1/2",
		Access: map[string]string{"alice": protocol.VoiceAccessGranted},
	}
	room.SetVoice(in)
	in.Access["alice"] = "changed" // the room keeps its own copy
	in.Status = "changed"

	view, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if view.Voice == nil || view.Voice.Status != protocol.VoiceStatusReady || view.Voice.Access["alice"] != protocol.VoiceAccessGranted {
		t.Errorf("snapshot stamp = %+v", view.Voice)
	}
	if got := room.ViewFor("alice").Voice; got == nil || got.You != protocol.VoiceAccessGranted || got.URL == "" {
		t.Errorf("alice's view = %+v", got)
	}
	if got := room.ViewFor("bob").Voice; got == nil || got.You != protocol.VoiceAccessSpectator || got.URL != "" {
		t.Errorf("an unseated viewer's view = %+v", got)
	}

	room.SetVoice(nil)
	if view, _, _ := room.Snapshot(); view.Voice != nil {
		t.Errorf("cleared, but the snapshot still has %+v", view.Voice)
	}
}
