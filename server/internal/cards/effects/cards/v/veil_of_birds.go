package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Veil of Birds — Enchantment {U} (Urza's Saga, S82 OLD b721):
//
//	"When an opponent casts a spell, if this permanent is an enchantment, it becomes a 1/1 Bird creature with flying."
//
// A CR 603.4 intervening "if": the cast is checked as the ability triggers and the enchantment as it resolves. The effect has no duration (CR 611.2a) and does not say "in addition", so the permanent stops being an enchantment (CR 205.1a) and the ability cannot trigger again; it keeps its colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc6278e5-7eb1-4590-880e-fe9efa0ec4d9",
		Name:         "Veil of Birds",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB721OpponentCastWhileEnchantment(nil),
				"Veil of Birds — it becomes a 1/1 Bird creature",
				effects.ThisEnchantmentBecomesACreature("Bird", 1, 1, "Veil of Birds — a 1/1 Bird creature", game.AddKeywordsMod("flying"))),
		},
	})
}
