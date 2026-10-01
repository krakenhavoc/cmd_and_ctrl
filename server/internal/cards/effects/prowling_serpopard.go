package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prowling Serpopard — Creature — Cat Snake {1}{G}{G}, 4/3:
//
//	"This spell can't be countered.
//	 Creature spells you control can't be countered."
//
// Two different statements. The first is the spell's own rider
// (Spec.CantBeCountered), read while the Serpopard is on the stack; the
// second is ADR 0106 §4's battlefield static (#1806), read while it is
// on the battlefield. Neither does the other's job.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4bdfd718-f474-4223-88d8-7fa9fb0c86b4",
		Name:            "Prowling Serpopard",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsYouControlCantBeCountered("Creature spells you control can't be countered.", Creature()),
		},
	})
}
