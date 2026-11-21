package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Koala-Sheep — Creature — Bear Sheep {2}{W}, 3/2 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"When this creature enters, you gain 3 life."
//
// Angel of Mercy's enters trigger on a ground body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e0226eb-20a3-488c-8f2b-f99a5b4d0cf4",
		Name:         "Koala-Sheep",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Koala-Sheep — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
