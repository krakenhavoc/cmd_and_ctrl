package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Secretkeeper — Creature — Spirit {3}{U}, 2/2 (Saviors of Kamigawa, #5601):
//
//	"As long as you have more cards in hand than each opponent, this creature
//	 gets +2/+2 and has flying."
//
// Two statics read live through the layers, a layer 7c +2/+2 and a layer 6
// flying, both over the hand sizes of you and every opponent still in the game
// (DependsOnHandSize, so a draw, a discard or a play changes them at once). A
// tie is not "more".
//
// No simplification.
func init() {
	const when = "As long as you have more cards in hand than each opponent, this creature "
	effects.Register(effects.Spec{
		OracleID:     "709a8bf0-bd0b-4ee0-9c06-98b7551447b5",
		Name:         "Secretkeeper",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.OldB588WhileYouHavePump(effects.OldB588YouHaveMoreCardsInHandThanEachOpponent, 2, 2, when+"gets +2/+2."),
			effects.OldB588WhileYouHaveKeyword(effects.OldB588YouHaveMoreCardsInHandThanEachOpponent, "flying", when+"has flying."),
		},
	})
}
