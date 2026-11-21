package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lion Vulture — Creature — Cat Bird {3}{B}, 2/2 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Flying
//	 At the beginning of your end step, if an opponent lost life this
//	 turn, put a +1/+1 counter on this creature and draw a card."
//
// An intervening "if" (CR 603.4): checked as the end step begins, off the
// engine's per-turn life-loss tally of every opponent, and again as the
// trigger resolves. The counter goes on only while the Vulture is still
// on the battlefield; the draw happens either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "169d24f3-a308-4a2e-ae25-85dae448994e",
		Name:            "Lion Vulture",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ByYou(ev, source, lki, g) && effects.OpponentLifeLostThisTurn(g, source.Controller) > 0
			}, "Lion Vulture — put a +1/+1 counter on this creature and draw a card", effects.Do(effects.If{
				Cond: effects.AvB34CondAnOpponentLostLife{},
				Then: []effects.Applier{effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}, effects.DrawCards{N: 1}},
			})),
		},
	})
}
