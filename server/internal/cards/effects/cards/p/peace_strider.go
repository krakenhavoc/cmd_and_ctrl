package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peace Strider — Artifact Creature — Construct {4}, 3/3 (Mirrodin Besieged,
// #5576):
//
//	"When this creature enters, you gain 3 life."
//
// Angel of Mercy's enters trigger: it uses the stack, so it can be
// answered, and the life goes to the trigger's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0e96742-77c0-4de5-84b0-0d59a8bee3e0",
		Name:         "Peace Strider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Peace Strider — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
