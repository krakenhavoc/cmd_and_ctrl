package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thryx, the Sudden Storm — Legendary Creature — Elemental Giant
// {3}{U}{U}, 4/5:
//
//	"Flash
//	 Flying
//	 Spells you cast with mana value 5 or greater cost {1} less to cast
//	 and can't be countered."
//
// One printed sentence, two statics. The discount is a CostModifier
// (CR 601.2f); "can't be countered" is ADR 0106 §4's battlefield static
// (#1806) in its "you cast" form, so it follows the caster through a
// change of control (ADR 0104) and never covers a copy, which is not
// cast (CR 707.10). Both read the mana value the spell has on the
// stack, X included (CR 202.3e), and neither moves it: mana value is
// counted from the mana cost (CR 202.3), and a reduction changes only
// what is paid (CR 118.7).
//
// No simplification.
func init() {
	const clause = "Spells you cast with mana value 5 or greater cost {1} less to cast and can't be countered."
	effects.Register(effects.Spec{
		OracleID:        "20004398-bd04-4583-8378-9df743246a32",
		Name:            "Thryx, the Sudden Storm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, clause, effects.YourSpell(), effects.SpellManaValueAtLeast(5)),
		},
		SpellsCantBeCountered: []game.CounterShieldStatic{
			effects.SpellsYouCastCantBeCountered(clause, effects.ManaValueGE(5)),
		},
	})
}
