package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maggot Carrier — Creature — Zombie {B}, 1/1 (Eighth Edition, S82 OLD batch
// 609, #5607):
//
//	"When this creature enters, each player loses 1 life."
//
// Every live player, the controller included (EachPlayerLifeChanges); life
// loss, not damage (CR 119.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4ad4b81-3685-4f95-84c0-755263b9d3b1",
		Name:         "Maggot Carrier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Maggot Carrier — each player loses 1 life", effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.EachPlayerLifeChanges(ctx, -1)
			}))),
		},
	})
}
