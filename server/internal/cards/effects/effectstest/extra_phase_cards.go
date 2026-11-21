package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extra_phase_cards_test.go — the extra-combat half of ADR 0059's first
// wave (sub-PR 2b, #753). Each card is driven through the real cast,
// activation or attack, and the added phase is asserted by where the
// cursor goes next, never by reading the plan alone.

const (
	RelentlessAssaultOracle     = "dc1c0e8c-d0c1-445c-968f-7dec91e5d5fc"
	SeizeTheDayOracle           = "2a6abd59-e448-46f1-9af8-bb9040645971"
	AggravatedAssaultOracle     = "20129459-a386-41eb-899d-1aede3427300"
	FullThrottleOracle          = "3c2f87eb-b0f4-4578-b143-6fba313a82ec"
	KarlachOracle               = "037355be-71e7-4866-80a6-80352c304970"
	HellkiteChargerOracle       = "d4f28a4b-d821-4132-bdec-4c528318f8e2"
	AureliaWarleaderOracle      = "0f5a3a09-2f07-4774-9e0f-e99d9a444166"
	SphinxOfTheSecondSunOracle  = "516101be-be39-4d84-8fee-d8a79930dd0a"
	EomerMarshalOfRohanOracle   = "b2d95950-18b3-463f-94f4-299e420751dc"
	extraPhaseTestCreatureType  = "Creature — Test"
	ExtraPhaseTestLegendaryType = "Legendary Creature — Test"
)

// stepOnce advances one step and returns where the cursor landed.
func StepOnce(t *testing.T, g *game.Game) game.Turn {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	return g.Turn
}

// countResolved counts EventResolve entries with `label` since `from`.
func CountResolved(g *game.Game, from int, label string) int {
	n := 0
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventResolve && ev.Label == label {
			n++
		}
	}
	return n
}
