package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// retrace_followups_test.go — #2550. The retrace cards #2528 left out
// (Deeproot Historian's two-type grant and four token makers) and the
// Cackling Counterpart caveat that had gone stale.

const (
	HistorianOracle = "4b8d82fe-571d-4fd6-8d0c-77b7e5ef3e23"
	CennsOracle     = "49fe9f5a-5821-4586-b913-7d8aef1f8669"
	SkybreakOracle  = "a7213bef-0e9c-44e1-97cb-d1da0ef370bf"
	WormOracle      = "b2223f1c-e607-43e0-86dd-5e3225330066"
	FormlessOracle  = "b978bdfe-a73e-4e99-90d7-4f99182c45ee"
)

// ---------------------------------------------------------- Deeproot Historian

func HistorianTable(t *testing.T) (g *game.Game, me *game.Player, historian, land uuid.UUID) {
	t.Helper()
	g, me, land, _ = RetraceTable(t)
	historian = PushCatalogPermanent(g, me.ID, "Deeproot Historian", "Creature — Merfolk Druid", HistorianOracle, false)
	return
}
