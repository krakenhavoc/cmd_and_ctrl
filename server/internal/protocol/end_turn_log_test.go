package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// end_turn_log_test.go — #2165: an effect ending the turn (CR 724.1) is
// one `turn_ended` line naming the active player and the card, ahead of
// the cleanup step's ordinary `step` line, so the jump in the turn bar
// is explained.
func TestTurnEndedIsOneLogLineBeforeTheCleanupStep(t *testing.T) {
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	active := g.Seats[g.Turn.ActiveSeat]
	source := uuid.New()
	g.WithWriteLock(func() {
		active.Graveyard.PushTop(game.Card{
			InstanceID: source, Name: "Time Stop", TypeLine: "Instant",
			Owner: active.ID, Controller: active.ID,
		})
		g.EndTheTurnForEffect(source)
	})
	g.SettleResolution()

	v := ViewOfGame(g)
	var ended []LogEvent
	cleanupAfter := false
	for _, e := range v.Log {
		if e.Kind == LogTurnEnded {
			ended = append(ended, e)
			continue
		}
		if len(ended) > 0 && e.Kind == LogStep && strings.Contains(strings.ToLower(e.Text), "cleanup") {
			cleanupAfter = true
		}
	}
	if len(ended) != 1 {
		t.Fatalf("want one turn_ended line, got %d", len(ended))
	}
	e := ended[0]
	if e.Seat != active.Seat {
		t.Errorf("seat %d, want the active seat %d", e.Seat, active.Seat)
	}
	if e.CardID != source.String() {
		t.Errorf("card_id %q, want the source %s", e.CardID, source)
	}
	if !strings.Contains(e.Text, "turn ends") || !strings.Contains(e.Text, "Time Stop") {
		t.Errorf("text = %q", e.Text)
	}
	if !cleanupAfter {
		t.Errorf("no cleanup step line after the turn_ended line")
	}
}
