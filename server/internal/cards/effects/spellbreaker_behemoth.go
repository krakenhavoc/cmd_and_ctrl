package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spellbreaker Behemoth — Creature — Beast {1}{R}{G}{G}, 5/5:
//
//	"This spell can't be countered.
//	 Creature spells you control with power 5 or greater can't be
//	 countered."
//
// The rider is Spec.CantBeCountered; the second line is ADR 0106 §4's
// battlefield static (#1806). A creature spell's power is read as the
// spell has it on the stack, the way every other spell-power read in
// the catalog reads it (The Fantastic Four).
//
// ONE SIMPLIFICATION. The engine does not run a characteristic-defining
// ability on a card on the stack, though CR 604.3 says one works there,
// so a */* creature spell (Tarmogoyf, Multani) is read at the 0 its
// printed box holds and is not covered. Strictly weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "cba07472-7212-4411-a9f9-38a48870ad69",
		Name:         "Spellbreaker Behemoth",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A creature spell whose power is set by its own text, such as a */* creature, isn't protected — it can still be countered.",
		},
		CantBeCountered: true,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsYouControlCantBeCountered("Creature spells you control with power 5 or greater can't be countered.", Creature(), PowerGE(5)),
		},
	})
}
