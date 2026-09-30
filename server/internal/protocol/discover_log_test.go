package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0099: a finished discover is a log line naming N and the card,
// or saying nothing was found.
func TestDiscoverLogLine(t *testing.T) {
	card := uuid.New()
	turn, step := 1, ""
	var sacrificed uuid.UUID
	e, ok := projectEvent(game.Event{Kind: game.EventDiscover, Actor: uuid.New(), Amount: 4, CardID: card},
		func(uuid.UUID) int { return 0 }, &turn, &step, &sacrificed)
	if !ok || e.Kind != LogDiscover || e.Amount != 4 || e.CardID != card.String() {
		t.Fatalf("projected %+v, %v", e, ok)
	}
	e.actorName = "Alice"
	if got := renderLogText(e, "Lightning Helix", ""); got != "Alice discovered 4 — Lightning Helix" {
		t.Errorf("line = %q", got)
	}
	e.CardID = ""
	if got := renderLogText(e, "", ""); got != "Alice discovered 4 and found nothing" {
		t.Errorf("empty line = %q", got)
	}
}
