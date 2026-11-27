package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Staunch Defenders — Creature — Human Soldier {3}{W}{W}, 3/4 (Eighth Edition,
// S82 OLD batch 609, #5607):
//
//	"When this creature enters, you gain 4 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec8bf245-4b2c-432b-9b2a-8d7b9224258c",
		Name:         "Staunch Defenders",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Staunch Defenders — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
