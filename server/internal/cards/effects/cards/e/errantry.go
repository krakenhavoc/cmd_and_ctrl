package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Errantry — Enchantment — Aura {1}{R} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+0 and can only attack alone."
//
// A layer 7c pump and Master of Cruelties' attack limit, bound to the enchanted
// creature rather than the Aura (AttackLimit.OnAttached, CR 506.5, 508.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fee455f2-fcc4-47b1-a533-92fa1ae8ae63",
		Name:         "Errantry",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(3, 0)},
		AttackLimits: []game.AttackLimit{effects.CanOnlyAttackAloneAttached()},
	})
}
