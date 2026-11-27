package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witchstalker — Creature — Wolf {1}{G}{G}, 3/3 (Magic 2014, S82 old sets):
//
//	"Hexproof
//	 Whenever an opponent casts a blue or black spell during your turn, put a
//	 +1/+1 counter on this creature."
//
// Mold Adder's cast trigger (an opponent's spell that is blue, black or both,
// read through the card's text), narrowed to the Witchstalker controller's own
// turn, and not optional. One counter for each such spell, written with the Do
// primitive that declares a footprint.
//
// No simplification.
func init() {
	opponentBlueOrBlack := effects.AnOpponentCast(effects.Or(effects.OfColor("U"), effects.OfColor("B")))
	effects.Register(effects.Spec{
		OracleID:        "a5d3af88-2ca6-4e3d-8b6e-67482136a5c9",
		Name:            "Witchstalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.IsActivePlayer(g, source.Controller) && opponentBlueOrBlack(ev, source, lki, g)
			}, "Witchstalker — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
