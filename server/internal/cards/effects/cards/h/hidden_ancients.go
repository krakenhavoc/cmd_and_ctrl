package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidden Ancients — Enchantment {1}{G} (Urza's Saga, S82 OLD b721):
//
//	"When an opponent casts an enchantment spell, if this permanent is an enchantment, it becomes a 5/5 Treefolk creature."
//
// A CR 603.4 intervening "if": the cast is checked as the ability triggers and the enchantment as it resolves. The effect has no duration (CR 611.2a) and does not say "in addition", so the permanent stops being an enchantment (CR 205.1a) and the ability cannot trigger again; it keeps its colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc430e1b-d57f-4035-b518-9bc2e26a6c04",
		Name:         "Hidden Ancients",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB721OpponentCastWhileEnchantment(effects.Enchantment()),
				"Hidden Ancients — it becomes a 5/5 Treefolk creature",
				effects.ThisEnchantmentBecomesACreature("Treefolk", 5, 5, "Hidden Ancients — a 5/5 Treefolk creature")),
		},
	})
}
