package ws

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// idle_test.go pins the room's half of ADR 0211 §1: which commits
// reset a started game's idle clock. A person's and an agent's own move
// do; a bot's, a stand-in's, an admin's and the lobby's do not.

func TestLastHumanAtCountsOnlyAPersonsOwnMove(t *testing.T) {
	room, human, bot, agent := seatKindsRoom(t)
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	room.SetClock(func() time.Time { return now })

	notCounted := []struct {
		name   string
		commit func() error
	}{
		{"a bot's action", func() error { _, _, err := room.ApplyAction(ActionTag{Actor: bot}, ok); return err }},
		{"a bot's no-undo action", func() error { _, _, err := room.ApplyExternalAction(ActionTag{Actor: bot}, ok); return err }},
		{"a bot's bundle", func() error {
			_, _, err := room.ApplyBundle(Bundle{Actor: bot, Steps: []func() error{ok}})
			return err
		}},
		{"the admin's action", func() error { _, _, err := room.ApplyAction(ActionTag{Actor: uuid.Nil}, ok); return err }},
		{"an admin playing a seat", func() error {
			_, _, err := room.ApplyAction(ActionTag{Actor: human, Admin: true}, ok)
			return err
		}},
		{"a lobby step", func() error { _, _, err := room.ApplyExternal(ok); return err }},
		{"a refused move", func() error { _, _, _ = room.ApplyAction(ActionTag{Actor: human}, refuse); return nil }},
	}
	for _, c := range notCounted {
		if err := c.commit(); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got := room.LastHumanAt(); !got.IsZero() {
			t.Fatalf("%s stamped LastHumanAt %v; only a person's own move may", c.name, got)
		}
	}

	if _, _, err := room.ApplyAction(ActionTag{Actor: human}, ok); err != nil {
		t.Fatal(err)
	}
	if got := room.LastHumanAt(); !got.Equal(now) {
		t.Fatalf("a person's move: LastHumanAt %v, want %v", got, now)
	}

	now = now.Add(time.Hour)
	if _, _, err := room.ApplyExternalAction(ActionTag{Actor: agent}, ok); err != nil {
		t.Fatal(err)
	}
	if got := room.LastHumanAt(); !got.Equal(now) {
		t.Fatalf("an agent's move counts as a person's (ADR 0211 owner decision 5): LastHumanAt %v, want %v", got, now)
	}

	// A stand-in runner commits as the person's seat; that is the bot's
	// move, not theirs.
	room.SetStandIn(human, "heuristic", now)
	later := now.Add(time.Hour)
	now = later
	if _, _, err := room.ApplyAction(ActionTag{Actor: human}, ok); err != nil {
		t.Fatal(err)
	}
	if got := room.LastHumanAt(); !got.Equal(later.Add(-time.Hour)) {
		t.Fatalf("a stand-in's commit moved LastHumanAt to %v", got)
	}
}

func TestLastSeatedChatAtIgnoresBots(t *testing.T) {
	room, human, bot, _ := seatKindsRoom(t)
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	room.SetClock(func() time.Time { return now })
	room.noteSeatedChat(bot)
	if got := room.LastSeatedChatAt(); !got.IsZero() {
		t.Fatalf("a bot's chat stamped %v", got)
	}
	room.noteSeatedChat(human)
	if got := room.LastSeatedChatAt(); !got.Equal(now) {
		t.Fatalf("a seated person's chat: %v, want %v", got, now)
	}
	if got := room.LastHumanAt(); !got.IsZero() {
		t.Fatalf("chat is not a move, but LastHumanAt is %v", got)
	}
}
