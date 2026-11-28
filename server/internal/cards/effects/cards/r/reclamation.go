package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reclamation — Enchantment {2}{G}{W} (Ice Age, S82 OLD batch 848, #5636):
//
//	"Black creatures can't attack unless their controller sacrifices a land of
//	 their choice for each black creature they control that's attacking. (This
//	 cost is paid as attackers are declared.)"
//
// The #6190 attack cost (CreaturesCantAttackUnlessControllerSacrifices): an
// AttackTax on every black attacker at the table, whoever controls it, owed by
// that creature's controller whatever it attacks (OnAttacker), one land of their
// own choice per taxed attacker, paid as attackers are declared (CR 508.1h).
// Colour is read off the attacker through the layers at declaration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff942899-a70d-4fca-b125-9712aa635fff",
		Name:         "Reclamation",
		Completeness: effects.CompletenessFull,
		AttackTaxes:  []game.AttackTax{effects.OldB848CreaturesCantAttackUnlessLandSacrificed("B")},
	})
}
