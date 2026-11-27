package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Okina Nightwatch — Creature — Human Monk {4}{G}, 4/3 (Saviors of Kamigawa,
// #5601):
//
//	"As long as you have more cards in hand than each opponent, this creature
//	 gets +3/+3."
//
// Secretkeeper's comparison with a layer 7c +3/+3, read live over the hand
// sizes of you and every opponent still in the game; a tie is not "more".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1c222fe-0ca7-415b-a28a-3b704df53d66",
		Name:         "Okina Nightwatch",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.OldB588WhileYouHavePump(effects.OldB588YouHaveMoreCardsInHandThanEachOpponent, 3, 3,
				"As long as you have more cards in hand than each opponent, this creature gets +3/+3."),
		},
	})
}
