package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sage's Row Savant — Creature — Vedalken Wizard {1}{U}, 2/1 (Ravnica: Clue
// Edition):
//
//	"When this creature enters, scry 2."
//
// Aether Theorist's scry primitive, declared with its Purpose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6eaf5cdc-2843-4167-bd18-df4c2ac73ac0",
		Name:         "Sage's Row Savant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sage's Row Savant — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		Purpose: game.Purpose{Scry: 2},
	})
}
