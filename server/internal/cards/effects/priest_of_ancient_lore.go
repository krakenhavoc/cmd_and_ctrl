package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Priest of Ancient Lore — Creature — Dwarf Cleric {2}{W}, 2/1 (Adventures in the Forgotten Realms):
//
//	When this creature enters, you gain 1 life and draw a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a89b0fd5-84b2-487c-8b03-e78d97276fd3",
		Name:         "Priest of Ancient Lore",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 1, Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Priest of Ancient Lore — gain 1 life and draw a card", Do(GainLife{Amount: 1}, DrawCards{N: 1})),
		},
	})
}
