package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea Gate Oracle — Creature — Human Wizard {2}{U}, 1/3 (Rise of the
// Eldrazi, #5581):
//
//	"When this creature enters, look at the top two cards of your library.
//	 Put one of them into your hand and the other on the bottom of your
//	 library."
//
// Teferi, Temporal Archmage's +1 on an enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b333c194-d285-4dfb-984c-57b7e88393af",
		Name:         "Sea Gate Oracle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sea Gate Oracle — look at the top two cards of your library",
				effects.LookAtTopThenTakeOneRestOnBottom(2, "Sea Gate Oracle — put one into your hand and the other on the bottom")),
		},
	})
}
