package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cumulative_upkeep_test.go — #567 / CR 702.24. The counter goes on,
// then the printed cost is charged once per counter, and declining
// sacrifices.

const OracleMysticRemora = "8a52f3c0-2552-4425-b2e3-5496eb2232a7"

func RemoraUpkeepPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	t.Fatalf("no cumulative-upkeep prompt for %s", chooser)
	return nil
}

func AgeCountersOn(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		return 0
	}
	return c.Counters[game.CounterAge]
}
