package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leyline of the Meek — Enchantment {2}{W}{W} (Guildpact):
//
//	"If this card is in your opening hand, you may begin the game with it
//	 on the battlefield.
//	 Creature tokens get +1/+1."
//
// The opening-hand clause is Leyline of Vitality's. The anthem is a layer 7c
// modify on every creature token on the battlefield, whoever controls it
// (not only yours), read on every layer pass. A token creature is a creature
// token as it enters and as long as it is one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "971f7b30-0088-4e12-bc1a-262acc7f5f90",
		Name:         "Leyline of the Meek",
		Completeness: effects.CompletenessFull,
		OpeningHand:  effects.BeginTheGameOnTheBattlefield(),
		Static: []game.StaticAbility{{
			Label:    "Creature tokens get +1/+1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature() && target.IsToken()
			},
			Apply: effects.PumpBy(1, 1),
		}},
	})
}
