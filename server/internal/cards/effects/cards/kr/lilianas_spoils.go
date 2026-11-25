package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana's Spoils — Sorcery {3}{B} (Core Set 2019, S82 OLD batch 200, #5527):
//
//	"Target opponent discards a card.
//	 Look at the top five cards of your library. You may reveal a black card
//	 from among them and put it into your hand. Put the rest on the bottom of
//	 your library in a random order."
//
// Mind Peel's discard (the opponent picks the card, CR 701.9b) and then
// Skalla Wolf's look-at-five for a black card. With its one target illegal the
// whole spell does nothing (CR 608.2b), the look included.
//
// No simplification.
func init() {
	discard := effects.TargetOpponentDiscards(1)
	look := effects.LookAtTopThenMayTakeToHand(5, effects.OfColor("B"), 1,
		"Liliana's Spoils — you may reveal a black card and put it into your hand")
	effects.Register(effects.Spec{
		OracleID:     "751da0fa-0ab1-4b8c-81bb-d86a0a280ac4",
		Name:         "Liliana's Spoils",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.FirstLegalPlayerTarget(ctx); !ok {
				return nil
			}
			if err := discard(ctx.Game, item); err != nil {
				return err
			}
			return look(ctx.Game, item)
		},
	})
}
