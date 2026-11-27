package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Opal Caryatid — Enchantment {W} (Urza's Saga, S82 OLD b721):
//
//	"When an opponent casts a creature spell, if this permanent is an enchantment, it becomes a 2/2 Soldier creature."
//
// A CR 603.4 intervening "if": the cast is checked as the ability triggers and the enchantment as it resolves. The effect has no duration (CR 611.2a) and does not say "in addition", so the permanent stops being an enchantment (CR 205.1a) and the ability cannot trigger again; it keeps its colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eabcb1bd-5317-4e01-8f18-94f01c12c852",
		Name:         "Opal Caryatid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB721OpponentCastWhileEnchantment(effects.Creature()),
				"Opal Caryatid — it becomes a 2/2 Soldier creature",
				effects.ThisEnchantmentBecomesACreature("Soldier", 2, 2, "Opal Caryatid — a 2/2 Soldier creature")),
		},
	})
}
