package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	WakeTheReflectionsOracle  = "211868ca-662b-4054-bf50-c9e16bb58c49"
	RootbornDefensesOracle    = "04046ae4-5c51-434b-930c-f3b1d348bf4b"
	GrowingRanksOracle        = "bec6fb31-60a3-432f-ba36-aebf172f8b27"
	SunderingGrowthOracle     = "a2a380d8-4df7-4357-862c-ed3fb795db6c"
	VituGhaziGuildmageOracle  = "da6f79ec-5f7e-4099-89c5-6d2bfc14d957"
	DeterminedIterationOracle = "e4ed7935-0263-4cde-8118-ac18fffce3ef"
	DruidsDeliveranceOracle   = "fde7645a-5f02-4d5f-b38c-8390f325899e"
)

func PopulateTokens(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && c.IsToken() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
