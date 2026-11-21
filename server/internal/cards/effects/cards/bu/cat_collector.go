package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cat Collector — Creature — Human Citizen {2}{W}, 3/2 (Foundations,
// #4465):
//
//	"When this creature enters, create a Food token.
//	 Whenever you gain life for the first time during each of your turns,
//	 create a 1/1 white Cat creature token."
//
// The entry trigger is Bakersbane Duo's Food. The second trigger is
// Leech Collector's first-gain-of-the-turn reader, narrowed to a turn of
// yours: a gain on an opponent's turn neither triggers it nor uses up the
// first gain of your next turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52f1b5a4-62ff-44de-80da-503d4e52d312",
		Name:         "Cat Collector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cat Collector — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
			effects.On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.IsYourTurn(g, source.Controller) && effects.SocB28FirstLifeGainThisTurn(ev, source, g)
			}, "Cat Collector — create a 1/1 white Cat creature token",
				effects.Do(effects.CreateToken{Template: effects.B36WhiteCatToken(), N: 1})),
		},
	})
}
