package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// world_enchantments_1997_test.go — #1997: the world enchantments ADR
// 0109 PR 7 left out, each with the one shape of its own it needed.

const (
	TombstoneStairwellOracle = "9aaf1149-eccb-434a-a1a7-14b93fa50141"
	FieldOfDreamsOracle      = "6dcc62f8-2fd9-473b-9e6a-2dd7e610f855"
	KoskunFallsOracle        = "a9b9e798-debc-4f2c-9770-e16f670082f7"
	ElkinLairOracle          = "b3488947-7142-4e6c-9c92-2728469b3ad9"
	WintersNightOracle       = "e4f73802-b639-47ab-a6c6-f761afa4e00e"
	HallOfGemstoneOracle     = "3da9fecc-064a-44e6-b88e-4b10194e367e"
)

// --- Koskun Falls ---------------------------------------------------

// koskunPrompt advances to the controller's upkeep, resolves the trigger
// and returns its pay-unless prompt.
func KoskunPrompt(t *testing.T, g *game.Game, seat int) *game.PendingChoice {
	t.Helper()
	AdvanceToUpkeepOf(t, g, seat)
	PassPriorityAroundTable(t, g)
	return RemoraUpkeepPrompt(t, g, g.Seats[seat].ID)
}
