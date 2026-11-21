package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scurry Oak — Creature — Treefolk {2}{G}, 1/2:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 Whenever one or more +1/+1 counters are put on this creature, you
//	 may create a 1/1 green Squirrel creature token."
//
// Herd Baloth's trigger with a Squirrel: "one or more … are put" is one
// trigger per placement event, read off the log by
// b33CountersPlacedDelta so a removal never triggers it. Any placement
// counts — its own evolve, and anything else that puts a +1/+1 counter
// on it. Evolve is the engine's keyword trigger (game/evolve.go,
// #1805). A Hardened Scales turns one evolve counter into two, which is
// still one placement and one Squirrel.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eee6c02b-7d6a-4445-ad27-8b03176147d4",
		Name:            "Scurry Oak",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Target == source.InstanceID && effects.B33CountersPlacedDelta(ev, game.CounterPlusOne, g) > 0
			}, "Scurry Oak — you may create a 1/1 green Squirrel", effects.Do(effects.MayChoice{
				Question: "Scurry Oak — create a 1/1 green Squirrel creature token?",
				Then:     effects.MayCreateTokens,
				Word:     effects.TokenWord(1, "1/1 green Squirrel"),
			})),
		},
	})
}
