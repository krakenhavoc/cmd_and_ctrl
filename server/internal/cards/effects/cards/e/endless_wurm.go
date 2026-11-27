package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Endless Wurm — Creature — Wurm {3}{G}{G} 9/9 (Urza's Saga, #5624):
//
//	"Trample
//	 At the beginning of your upkeep, sacrifice this creature unless you
//	 sacrifice an enchantment."
//
// A controller with an enchantment chooses between sacrificing it and the
// Wurm; one without sacrifices the Wurm (Bog Elemental's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce4d714c-c22b-47be-a372-a3139c4e5b1c",
		Name:            "Endless Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Endless Wurm — sacrifice it unless you sacrifice an enchantment",
				effects.SacrificeThisUnlessYouSacrificeA("Endless Wurm", "an enchantment", effects.Enchantment(), effects.S82OldB727SacrificeAnEnchantmentOrThisAnswer)),
		},
	})
}
