package effectstest

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- helpers --------------------------------------------------------------

// evidenceEvents is every collection in the log.
func EvidenceEvents(g *game.Game) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == game.EventCollectEvidence {
			out = append(out, ev)
		}
	}
	return out
}

// wantOneCollection fails unless the log holds exactly one collection, by
// `who`, of exactly `cards` (in that order) worth `total`.
func WantOneCollection(t *testing.T, g *game.Game, who uuid.UUID, total int, cards ...uuid.UUID) {
	t.Helper()
	evs := EvidenceEvents(g)
	if len(evs) != 1 {
		t.Fatalf("%d collect evidence events, want 1: %+v", len(evs), evs)
	}
	if ev := evs[0]; ev.Actor != who || ev.Amount != total || !slices.Equal(ev.CardIDs, cards) {
		t.Errorf("event actor %s, amount %d, cards %v; want %s, %d, %v", ev.Actor, ev.Amount, ev.CardIDs, who, total, cards)
	}
}
