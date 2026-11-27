package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Opal Champion — Enchantment {2}{W}:
//
//	"When an opponent casts a creature spell, if this permanent is an
//	 enchantment, it becomes a 3/3 Knight creature with first strike."
//
// A trigger with an intervening "if" (CR 603.4): it checks that this is still
// an enchantment as it resolves. The creature it becomes is not an enchantment
// (CR 205.1a), so it never triggers again. The first strike is part of the
// same effect and lasts as long as the permanent is this object.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d3617c3-1526-40e6-bc90-6139b5d3e1f1",
		Name:         "Opal Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB717AnOpponentCastIfEnchantment(effects.Creature()),
				"Opal Champion — it becomes a 3/3 Knight creature with first strike",
				effects.ThisEnchantmentBecomesACreature("Knight", 3, 3, "Opal Champion — a 3/3 Knight creature with first strike",
					game.AddKeywordsMod("first strike"))),
		},
	})
}
