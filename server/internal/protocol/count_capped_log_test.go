package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3040: a clamp that bites is one public log line naming the quantity,
// the ceiling and the object, so a capped table is told it is capped.
func TestCountCappedIsInTheLog(t *testing.T) {
	g := buildActiveGame(t)
	source := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: source, Name: "Ouroboroid", TypeLine: "Creature — Plant Wurm",
			Owner: g.Seats[0].ID, Controller: g.Seats[0].ID,
		})
		g.EmitEvent(game.Event{Kind: game.EventCountCapped, Actor: g.Seats[0].ID, Source: source, Label: "+1/+1 counters", Amount: game.MaxCount})
	})
	var got *LogEvent
	for _, e := range ViewOfGame(g).Log {
		if e.Kind == LogCountCapped {
			e := e
			got = &e
		}
	}
	if got == nil {
		t.Fatal("no count_capped line")
	}
	if got.CardID != source.String() || got.Amount != game.MaxCount || got.Label != "+1/+1 counters" ||
		!strings.Contains(got.Text, "Ouroboroid") || !strings.Contains(got.Text, "capped at 1073741824") {
		t.Errorf("line = %+v", got)
	}
}
