package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Opal Gargoyle — Enchantment {1}{W} (Urza's Saga, S82 OLD b721):
//
//	"When an opponent casts a creature spell, if this permanent is an enchantment, it becomes a 2/2 Gargoyle creature with flying."
//
// A CR 603.4 intervening "if": the cast is checked as the ability triggers and the enchantment as it resolves. The effect has no duration (CR 611.2a) and does not say "in addition", so the permanent stops being an enchantment (CR 205.1a) and the ability cannot trigger again; it keeps its colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efef2d33-5354-4fce-880e-cb91fa60f730",
		Name:         "Opal Gargoyle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB721OpponentCastWhileEnchantment(effects.Creature()),
				"Opal Gargoyle — it becomes a 2/2 Gargoyle creature",
				effects.ThisEnchantmentBecomesACreature("Gargoyle", 2, 2, "Opal Gargoyle — a 2/2 Gargoyle creature", game.AddKeywordsMod("flying"))),
		},
	})
}
