package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// saturating_counts_test.go — #3040. A doubling board must not be able
// to overflow `int` or ask Go for a terabyte string. The ceilings live
// in game/saturate.go; these tests drive the real cards into them.

// countCappedEvents counts the EventCountCapped announcements whose
// label contains `what`, read off the whole event log.
func CountCappedEvents(g *game.Game, what string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventCountCapped && strings.Contains(ev.Label, what) {
				n++
			}
		}
	})
	return n
}

// effectivePowerWithCounters is the counter-aware power a combat reads.
func EffectivePowerWithCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	p, found := 0, false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				p, found = c.PowerForComparison(), true
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return p
}
