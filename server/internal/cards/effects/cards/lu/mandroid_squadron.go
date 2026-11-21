package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mandroid Squadron — Artifact Creature — Human Soldier {1}{W}, 0/4
// (Marvel Super Heroes Commander, #4455): "When this creature enters,
// you gain 2 life." The ETB gain-life shape of Doom's Servo-Guards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95ab1529-0de0-416a-82b4-c5ed9276a9b4",
		Name:         "Mandroid Squadron",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mandroid Squadron — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
