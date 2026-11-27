package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidden Gibbons — Enchantment {G}:
//
//	"When an opponent casts an instant spell, if this permanent is an
//	 enchantment, it becomes a 4/4 Ape creature."
//
// A trigger with an intervening "if" (CR 603.4), like Opal Champion's. The
// creature is no longer an enchantment (CR 205.1a), so it triggers once.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2315b044-bfd2-40dd-a73e-4cf064b36999",
		Name:         "Hidden Gibbons",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB717AnOpponentCastIfEnchantment(effects.Instant()),
				"Hidden Gibbons — it becomes a 4/4 Ape creature",
				effects.ThisEnchantmentBecomesACreature("Ape", 4, 4, "Hidden Gibbons — a 4/4 Ape creature")),
		},
	})
}
