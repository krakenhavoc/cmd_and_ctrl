package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sultai Ascendancy — Enchantment {B}{G}{U} (Khans of Tarkir):
//
//	"At the beginning of your upkeep, surveil 2."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06e92ab3-c781-4f32-9ada-09a388047067",
		Name:         "Sultai Ascendancy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Sultai Ascendancy — surveil 2", effects.Do(effects.Surveil{N: 2})),
		},
	})
}
