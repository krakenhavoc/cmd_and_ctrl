package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Vitality — Enchantment — Aura Curse {2}{W} (Commander 2017,
// #5533):
//
//	"Enchant player
//	 Whenever enchanted player is attacked, you gain 2 life. Each opponent
//	 attacking that player does the same."
//
// Curse of Verbosity's shape with 2 life in place of a card
// (c17CurseWhenAttacked): one declaration is one "is attacked", and each
// attacking opponent gains 2 life once however many creatures they send.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d2eb412-e09b-4ab3-8653-b458e9f38eff",
		Name:         "Curse of Vitality",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: effects.C17CurseWhenAttacked("Curse of Vitality — you gain 2 life",
			"Curse of Vitality — the attacking opponent gains 2 life", effects.Do(effects.GainLife{Amount: 2})),
	})
}
