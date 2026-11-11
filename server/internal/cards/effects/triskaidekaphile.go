package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Triskaidekaphile — Creature — Human Wizard {1}{U}, 1/3 (Bloomburrow
// Commander, #5368):
//
//	"You have no maximum hand size.
//	 At the beginning of your upkeep, if you have exactly thirteen cards
//	 in your hand, you win the game.
//	 {3}{U}: Draw a card."
//
// The upkeep trigger is Felidar Sovereign's shape with a hand count: an
// intervening "if" (CR 603.4) that triggers only when its controller
// holds exactly thirteen as the upkeep begins, and is checked again as
// it resolves, so a discard in response leaves it doing nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "ef238ba8-acf0-4e62-9abd-5291886a32e8",
		Name:          "Triskaidekaphile",
		Completeness:  CompletenessFull,
		NoMaxHandSize: true,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && cmdr2S22HandHas(g, source.Controller, 13)
			}, "Triskaidekaphile — if you have exactly thirteen cards in your hand, you win the game", cmdr2S22WinIfHandHas(13)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Draw a card.",
			Cost:    ManaCost("{3}{U}"),
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
