package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anax and Cymede & Kynaios and Tiro — Legendary Creature — Human Soldier
// {1}{R}{G}{W}{U}, 3/8 (Mystery Booster 2):
//
//	"First strike, vigilance
//	 Heroic — Whenever you cast a spell that targets Anax and Cymede &
//	 Kynaios and Tiro, draw a card. Each player may put a land card from
//	 their hand onto the battlefield, then each opponent who didn't draws
//	 a card."
//
// The shared heroic trigger (CR 207.2c). The controller draws, then each
// player in turn order from the controller is asked "may put a land card"
// (a put, not a play: no land drop is spent). An opponent who didn't put
// one draws as soon as they have answered; that is not told apart from
// drawing after everyone has answered, since a draw changes only that
// player's own hand and no other player's choice reads it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68de3f03-2ae2-4644-acc8-299a57480e97",
		Name:            "Anax and Cymede & Kynaios and Tiro",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Draws: 1, Lands: 1, LandsUntapped: 1},
		PrintedKeywords: []string{"first strike", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Anax and Cymede & Kynaios and Tiro — draw a card; each player may put a land from hand, then each opponent who didn't draws (heroic)",
				effects.DrawThenEachMayPutALand),
		},
	})
}
