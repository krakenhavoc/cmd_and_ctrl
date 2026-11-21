package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krosan Druid — Creature — Centaur Druid {2}{G}, 2/3 (Dominaria, #5530):
//
//	"Kicker {4}{G} (You may pay an additional {4}{G} as you cast this spell.)
//	 When this creature enters, if it was kicked, you gain 10 life."
//
// A kicked enters trigger (CR 603.4): 10 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "1fbed9e4-86f6-4115-8468-8593e340701f",
		Name:          "Krosan Druid",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{4}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Krosan Druid — kicked: gain 10 life", effects.Do(effects.GainLife{Amount: 10})),
		},
	})
}
