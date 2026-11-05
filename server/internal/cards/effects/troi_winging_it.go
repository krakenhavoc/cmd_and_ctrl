package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Troi, Winging It — Legendary Creature — Betazoid Human Advisor {3}{W},
// 3/2 (Star Trek Commander, #4974):
//
//	"Whenever an opponent draws a card during your turn, create a 1/1
//	 blue Faerie creature token with flying."
//
// One trigger per card an opponent draws while Troi's controller is the
// active player (trekB47OpponentDrewDuringYourTurn); a draw on any other
// turn, or one of your own, triggers nothing. Each makes one 1/1 blue
// Faerie token with flying under Troi's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38561191-ecd1-4028-a0fd-4bca6a7eab58",
		Name:         "Troi, Winging It",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, trekB47OpponentDrewDuringYourTurn,
				"Troi, Winging It — create a 1/1 blue Faerie creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 blue Faerie with flying"), N: 1})),
		},
	})
}
