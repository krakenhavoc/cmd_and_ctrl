package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "307d20fd-a9c9-4da9-b990-26c653054640",
		Name:            "Savage Gorger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ByYou(ev, source, lki, g) && opponentLifeLostThisTurn(g, source.Controller) > 0
			}, "Savage Gorger — put a +1/+1 counter on this creature", Do(If{
				Cond: s82MidB06CondAnOpponentLostLife{},
				Then: []Applier{CounterOnThis{Kind: game.CounterPlusOne, N: 1}},
			})),
		},
	})
}
