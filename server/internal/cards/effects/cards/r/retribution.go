package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Retribution — Sorcery {2}{R}{R} (Homelands, S82 OLD b822, #5635):
//
//	"Choose two target creatures controlled by the same opponent. That player
//	chooses and sacrifices one of those creatures. Put a -1/-1 counter on the
//	other."
//
// Barrin's Spite's two-target shape on an opponent's creatures: the first clause
// must be an opponent's, the second is controlled by the same player
// (ControlledByTargetsController). Their controller picks the sacrifice; the
// other gets the counter (OldB822Retribution).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c7b0034-7eb5-4950-b6dc-8e1d07d26899",
		Name:         "Retribution",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			effects.Distinct(effects.ControlledByTargetsController(effects.TargetCreature("another target creature controlled by the same opponent"), 0)),
		),
		OnResolve: effects.OldB822Retribution,
	})
}
