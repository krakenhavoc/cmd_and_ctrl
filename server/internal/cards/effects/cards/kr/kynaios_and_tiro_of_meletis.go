package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kynaios and Tiro of Meletis — Legendary Creature — Human Soldier
// {R}{G}{W}{U}, 2/8 (Secret Lair Drop):
//
//	"At the beginning of your end step, draw a card. Each player may put a
//	 land card from their hand onto the battlefield, then each opponent who
//	 didn't draws a card."
//
// Anax and Cymede & Kynaios and Tiro's body on the end step: the controller
// draws, then each player in turn order from the controller is asked "may put
// a land card" (a put, not a play: no land drop is spent), and an opponent who
// didn't draws as soon as they have answered; a draw changes only that
// player's own hand, so this is not told apart from drawing after everyone has
// answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53ee4254-fef7-49ec-aafc-0320987764e6",
		Name:         "Kynaios and Tiro of Meletis",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Lands: 1, LandsUntapped: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Kynaios and Tiro of Meletis — draw a card; each player may put a land from hand, then each opponent who didn't draws a card",
				effects.DrawThenEachMayPutALand),
		},
	})
}
