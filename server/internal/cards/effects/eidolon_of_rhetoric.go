package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eidolon of Rhetoric — Enchantment Creature — Spirit {2}{W}, 1/4 (Journey into Nyx, #5558):
//
//	"Each player can't cast more than one spell each turn."
//
// Rule of Law's cast restriction on a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33017203-233e-498e-806e-94b1dccf5c35",
		Name:         "Eidolon of Rhetoric",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			EachPlayerMaxSpellsPerTurn(1,
				"Eidolon of Rhetoric — each player can't cast more than one spell each turn."),
		},
	})
}
