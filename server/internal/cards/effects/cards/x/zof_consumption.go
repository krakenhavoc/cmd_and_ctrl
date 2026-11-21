package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zof Consumption // Zof Bloodbog — Modal DFC. Front: Sorcery {4}{B}{B}:
// "Each opponent loses 4 life and you gain 4 life." Back: Land
// (mdfc_lands.go).
//
// Two steps (CR 608.2c): the life loss, which can stop on a CR 616 ordering
// prompt, then the gain, so the gain waits for the answer. No
// simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9f11985-e460-425d-b083-9cb0edf1983a",
		Name:         "Zof Consumption",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.EachOpponentLosesLife{N: 4}, effects.GainLife{Amount: 4}),
	})
}
