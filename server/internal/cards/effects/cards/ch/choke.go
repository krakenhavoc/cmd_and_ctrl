package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Choke — Enchantment {2}{G} (Eighth Edition, #5607):
//
//	"Islands don't untap during their controllers' untap steps."
//
// Every land with the Island subtype, basic or not, whoever controls it
// (CR 502.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "057fa60b-10b0-4612-be0d-157076c82241",
		Name:                  "Choke",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.Subtype("Island"))},
	})
}
