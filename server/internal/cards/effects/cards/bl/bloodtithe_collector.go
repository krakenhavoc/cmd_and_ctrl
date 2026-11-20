package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodtithe Collector — Creature — Vampire Noble {4}{B}, 3/4
// (Foundations, #4465):
//
//	"Flying
//	 When this creature enters, if an opponent lost life this turn, each
//	 opponent discards a card."
//
// Lion Vulture's "if an opponent lost life this turn" on Burglar Rat's
// enters discard. The condition is an intervening "if" (CR 603.4):
// checked as the trigger would fire and again as it resolves. Each
// opponent chooses the card they discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "93088b0c-d158-473a-9c06-aa2f344a9868",
		Name:            "Bloodtithe Collector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OpponentLifeLostThisTurn(g, source.Controller) > 0
			}), "Bloodtithe Collector — each opponent discards a card (an opponent lost life this turn)",
				effects.Do(effects.If{
					Cond: effects.AvB34CondAnOpponentLostLife{},
					Then: []effects.Applier{effects.Instruction(func(ctx *effects.Context) error { return effects.EachOpponentDiscardsOne(ctx.Game, ctx.Item) })},
				})),
		},
	})
}
