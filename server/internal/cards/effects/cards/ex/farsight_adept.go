package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Farsight Adept — Creature — Kor Wizard {2}{W}, 3/3 (Zendikar Rising,
// #5508):
//
//	"When this creature enters, you and target opponent each draw a
//	 card."
//
// Flumph's draw on an enters trigger: the opponent is chosen as the
// trigger goes on the stack; you draw, then the opponent does if still a
// legal target (youAndTargetOpponentEachDraw).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72b10cba-0f46-4b99-9634-e7173ac4c259",
		Name:         "Farsight Adept",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Farsight Adept — you and target opponent each draw a card", effects.YouAndTargetOpponentEachDraw()),
				effects.TargetPlayer("target opponent", effects.Opponent()),
			),
		},
	})
}
