package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2529: a skipped extra turn is one log line naming the player and the
// card that made it, so the earlier "will take an extra turn" line does
// not just evaporate.
func TestSkippedExtraTurnIsInTheLog(t *testing.T) {
	g := buildActiveGame(t)
	source := uuid.New()
	g.WithWriteLock(func() {
		g.Seats[0].Graveyard.PushTop(game.Card{
			InstanceID: source, Name: "Time Warp", TypeLine: "Sorcery",
			Owner: g.Seats[0].ID, Controller: g.Seats[0].ID,
		})
		g.EmitEvent(game.Event{Kind: game.EventExtraTurnSkipped, Actor: g.Seats[0].ID, Source: source, Amount: 1})
	})
	var got *LogEvent
	for _, e := range ViewOfGame(g).Log {
		if e.Kind == LogExtraTurnSkipped {
			e := e
			got = &e
		}
	}
	if got == nil {
		t.Fatal("no extra_turn_skipped line")
	}
	if got.Seat != 0 || got.CardID != source.String() || !strings.Contains(got.Text, "skips their extra turn") || !strings.Contains(got.Text, "Time Warp") {
		t.Errorf("line = %+v", got)
	}
}
