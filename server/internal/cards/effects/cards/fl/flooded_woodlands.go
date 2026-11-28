package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flooded Woodlands — Enchantment {2}{U}{B} (Ice Age, S82 OLD batch 848, #5636):
//
//	"Green creatures can't attack unless their controller sacrifices a land of
//	 their choice for each green creature they control that's attacking. (This
//	 cost is paid as attackers are declared.)"
//
// The #6190 attack cost (CreaturesCantAttackUnlessControllerSacrifices): an
// AttackTax on every green attacker at the table, whoever controls it, owed by
// that creature's controller whatever it attacks (OnAttacker), one land of their
// own choice per taxed attacker, paid as attackers are declared (CR 508.1h).
// Colour is read off the attacker through the layers at declaration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "413172aa-b508-4e67-b68e-6a749d2e8130",
		Name:         "Flooded Woodlands",
		Completeness: effects.CompletenessFull,
		AttackTaxes:  []game.AttackTax{effects.OldB848CreaturesCantAttackUnlessLandSacrificed("G")},
	})
}
