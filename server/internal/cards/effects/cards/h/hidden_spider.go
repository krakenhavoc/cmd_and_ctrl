package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidden Spider — Enchantment {G} (Urza's Saga, S82 OLD b721):
//
//	"When an opponent casts a creature spell with flying, if this permanent is an enchantment, it becomes a 3/5 Spider creature with reach."
//
// The spell's flying is read through the layers as it is cast. A CR 603.4 intervening "if": the cast is checked as the ability triggers and the enchantment as it resolves. The effect has no duration (CR 611.2a) and does not say "in addition", so the permanent stops being an enchantment (CR 205.1a) and the ability cannot trigger again; it keeps its colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0709570-a145-4c3c-ba56-f06d25753855",
		Name:         "Hidden Spider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB721OpponentCastWhileEnchantment(effects.And(effects.Creature(), effects.HasKeyword("flying"))),
				"Hidden Spider — it becomes a 3/5 Spider creature",
				effects.ThisEnchantmentBecomesACreature("Spider", 3, 5, "Hidden Spider — a 3/5 Spider creature", game.AddKeywordsMod("reach"))),
		},
	})
}
