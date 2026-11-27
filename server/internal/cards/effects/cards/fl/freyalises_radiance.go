package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Freyalise's Radiance — Enchantment {1}{G} (Coldsnap, S82 OLD b559):
//
//	"Cumulative upkeep {2}
//	 Snow permanents don't untap during their controllers' untap steps."
//
// Back to Basics' untap restriction over permanents with the snow supertype
// (CR 205.4g), whoever controls them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "d7f28d00-2ec3-4d4e-b3a6-89d33c66da28",
		Name:                  "Freyalise's Radiance",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.Snow())},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Freyalise's Radiance — cumulative upkeep {2}", "{2}"),
		},
	})
}
