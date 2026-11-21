package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eidolon of Rhetoric — Enchantment Creature — Spirit {2}{W}, 1/4 (Journey into Nyx, #5558):
//
//	"Each player can't cast more than one spell each turn."
//
// Rule of Law's cast restriction on a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33017203-233e-498e-806e-94b1dccf5c35",
		Name:         "Eidolon of Rhetoric",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.EachPlayerMaxSpellsPerTurn(1,
				"Eidolon of Rhetoric — each player can't cast more than one spell each turn."),
		},
	})
}
