package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Priest of Ancient Lore — Creature — Dwarf Cleric {2}{W}, 2/1 (Adventures in the Forgotten Realms):
//
//	When this creature enters, you gain 1 life and draw a card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a89b0fd5-84b2-487c-8b03-e78d97276fd3",
		Name:         "Priest of Ancient Lore",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 1, Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Priest of Ancient Lore — gain 1 life and draw a card", effects.Do(effects.GainLife{Amount: 1}, effects.DrawCards{N: 1})),
		},
	})
}
