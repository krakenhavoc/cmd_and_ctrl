package ws

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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

// ADR 0211 §6: the lobby's warning reaches every captured view as
// idle_close_at, and the human move that answers it clears it in that
// move's own frame. Nothing that is not a person's own move clears it.
func TestIdleCloseAtIsStampedUntilAPersonMoves(t *testing.T) {
	room, human, bot, _ := seatKindsRoom(t)
	closes := time.Date(2026, 11, 13, 21, 14, 0, 0, time.FixedZone("CET", 3600))
	want := "2026-11-13T20:14:00Z"

	view, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if view.IdleCloseAt != "" {
		t.Fatalf("an unwarned game's view has idle_close_at %q", view.IdleCloseAt)
	}
	room.SetIdleCloseAt(closes)
	if view, _, _ = room.Snapshot(); view.IdleCloseAt != want {
		t.Fatalf("idle_close_at = %q, want %q (RFC 3339, UTC)", view.IdleCloseAt, want)
	}
	if got := room.ViewFor(human.String()).IdleCloseAt; got != want {
		t.Errorf("a seat's view has idle_close_at %q, want %q", got, want)
	}

	notAnswers := []struct {
		name   string
		commit func() (string, error)
	}{
		{"a bot's action", func() (string, error) {
			v, _, err := room.ApplyAction(ActionTag{Actor: bot}, ok)
			return v.IdleCloseAt, err
		}},
		{"the admin's action", func() (string, error) {
			v, _, err := room.ApplyAction(ActionTag{Actor: uuid.Nil}, ok)
			return v.IdleCloseAt, err
		}},
		{"an admin playing a seat", func() (string, error) {
			v, _, err := room.ApplyAction(ActionTag{Actor: human, Admin: true}, ok)
			return v.IdleCloseAt, err
		}},
		{"a lobby step", func() (string, error) {
			v, _, err := room.ApplyExternal(ok)
			return v.IdleCloseAt, err
		}},
	}
	for _, c := range notAnswers {
		got, err := c.commit()
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got != want {
			t.Fatalf("%s: the frame's idle_close_at is %q, want the warning kept", c.name, got)
		}
	}
	if _, _, err := room.ApplyAction(ActionTag{Actor: human}, refuse); err == nil {
		t.Fatal("refuse did not refuse")
	}
	if room.IdleCloseAt().IsZero() {
		t.Fatal("a refused move took the warning back")
	}

	v, _, err := room.ApplyAction(ActionTag{Actor: human}, ok)
	if err != nil {
		t.Fatal(err)
	}
	if v.IdleCloseAt != "" {
		t.Errorf("the answering move's own frame still has idle_close_at %q", v.IdleCloseAt)
	}
	if !room.IdleCloseAt().IsZero() {
		t.Error("a person's move left the warning standing")
	}

	// A bundle of their own answers it too.
	room.SetIdleCloseAt(closes)
	if v, _, err = room.ApplyBundle(Bundle{Actor: human, Steps: []func() error{ok}}); err != nil {
		t.Fatalf("bundle: %v", err)
	}
	if v.IdleCloseAt != "" {
		t.Errorf("a person's bundle kept idle_close_at %q", v.IdleCloseAt)
	}
}

// A take-back is a move (ADR 0211 §1): a person's own undo answers the
// warning in its own frame; the admin's does not.
func TestIdleCloseAtIsClearedByAPersonsUndo(t *testing.T) {
	room, _ := newBundleRoom(t)
	a := room.Game.Seats[0].ID
	closes := time.Date(2026, 11, 13, 21, 14, 0, 0, time.UTC)
	unlimited := game.UndoUnlimited
	if err := room.Game.UpdateSettings(uuid.Nil, game.SettingsPatch{UndoLimit: &unlimited}); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, _, err := room.ApplyAction(ActionTag{Actor: a}, ok); err != nil {
			t.Fatal(err)
		}
	}
	room.SetIdleCloseAt(closes)
	v, _, err := room.UndoBy(a, true)
	if err != nil {
		t.Fatalf("the admin's undo: %v", err)
	}
	if v.IdleCloseAt == "" {
		t.Fatal("an admin's undo, bound to the seat, answered the warning")
	}
	if v, _, err = room.UndoBy(a, false); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if v.IdleCloseAt != "" {
		t.Errorf("a person's undo kept idle_close_at %q", v.IdleCloseAt)
	}
}

func TestPlayOutIsTheRoomsAndStartsOff(t *testing.T) {
	room, _, _, _ := seatKindsRoom(t)
	if room.PlayingOut() {
		t.Fatal("a new room plays out")
	}
	room.SetPlayOut(true)
	if !room.PlayingOut() {
		t.Fatal("SetPlayOut(true) did not stick")
	}
	view, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if view.Settings != nil && view.Settings.BotPace == "fast" {
		t.Error("the play-out reached the table's settings")
	}
	room.SetPlayOut(false)
	if room.PlayingOut() {
		t.Fatal("SetPlayOut(false) did not stick")
	}
}

// ADR 0211 §2 (sub-PR 5): a person's socket on their seat ends the
// play-out at once; a spectator's and an admin's do not, since neither
// is at the table.
func TestAPersonsSocketEndsThePlayOut(t *testing.T) {
	g := seedTestGame(t)
	room := NewRoom(g, quietLog(), "")
	seat := g.Seats[0].ID
	room.SetPlayOut(true)

	for _, b := range []Binding{
		{GameID: g.ID, ReadOnly: true},
		{GameID: g.ID, Admin: true},
		{GameID: g.ID, PlayerID: seat, Admin: true},
		{GameID: g.ID, PlayerID: seat, ReadOnly: true},
	} {
		readFrame(t, serveBound(t, room, b, nil))
		if !room.PlayingOut() {
			t.Fatalf("a socket bound %+v ended the play-out", b)
		}
	}
	readFrame(t, serveBound(t, room, Binding{GameID: g.ID, PlayerID: seat}, nil))
	if room.PlayingOut() {
		t.Fatal("a person's socket on their seat left the bots playing out fast")
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
