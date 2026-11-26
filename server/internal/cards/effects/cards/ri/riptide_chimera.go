package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riptide Chimera — Enchantment Creature — Chimera {2}{U}, 3/4 (Journey into
// Nyx, S82 OLD, #5558):
//
//	"Flying
//	 At the beginning of your upkeep, return an enchantment you control to its
//	 owner's hand."
//
// Roaring Primadox's mandatory upkeep bounce over enchantments: the controller
// picks as the trigger resolves and the Chimera itself, an enchantment, is a
// legal pick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "718d72c4-ad37-49b6-860a-0b921c0de978",
		Name:            "Riptide Chimera",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Riptide Chimera — return an enchantment you control", effects.Do(effects.ReturnOneYouControl{
				Match:    game.Card.IsEnchantment,
				Question: "Riptide Chimera — return an enchantment you control to its owner's hand",
			})),
		},
	})
}
