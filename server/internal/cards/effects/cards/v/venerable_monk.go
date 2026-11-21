package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venerable Monk — Creature — Human Monk Cleric {2}{W}, 2/2 (Portal, #5630):
//
//	"When this creature enters, you gain 2 life."
//
// An enters trigger on the stack, like Mandroid Squadron's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c5cfa6d-857f-44a9-80f7-46f016ef71e4",
		Name:         "Venerable Monk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Venerable Monk — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
