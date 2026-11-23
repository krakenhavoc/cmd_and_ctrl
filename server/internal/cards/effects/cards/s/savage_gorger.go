package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savage Gorger — Creature — Vampire {1}{B}{B}, 1/1 (Core Set 2020, #5518):
//
//	"Flying
//	 At the beginning of your end step, if an opponent lost life this turn,
//	 put a +1/+1 counter on this creature. (Damage causes loss of life.)"
//
// Lion Vulture's intervening "if" (CR 603.4), read off every opponent's
// per-turn life-loss tally as the end step begins and again as the ability
// resolves, with the read declared (Arrogant Outlaw's condition) so the row
// carries a footprint.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "307d20fd-a9c9-4da9-b990-26c653054640",
		Name:            "Savage Gorger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ByYou(ev, source, lki, g) && effects.OpponentLifeLostThisTurn(g, source.Controller) > 0
			}, "Savage Gorger — put a +1/+1 counter on this creature", effects.Do(effects.If{
				Cond: effects.S82MidB06CondAnOpponentLostLife{},
				Then: []effects.Applier{effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}},
			})),
		},
	})
}
