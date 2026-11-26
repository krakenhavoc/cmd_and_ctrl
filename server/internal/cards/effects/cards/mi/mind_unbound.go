package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Unbound — Enchantment {4}{U}{U} (Magic 2012, S82 OLD batch 452,
// #5573):
//
//	"At the beginning of your upkeep, put a lore counter on this
//	 enchantment, then draw a card for each lore counter on this
//	 enchantment."
//
// The counter goes on while it is still on the battlefield, and the draw
// counts the lore counters on it as it is now, or as it last was (CR
// 608.2h) if it has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f12aec9-5898-4261-bebd-1b545e0aea68",
		Name:         "Mind Unbound",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Mind Unbound — put a lore counter on this enchantment, then draw a card for each lore counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.CounterOnThis{Kind: game.CounterLore, N: 1}).Apply(ctx); err != nil {
						return err
					}
					info, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					return effects.DrawCards{N: info.Counters[game.CounterLore]}.Apply(ctx)
				}),
		},
	})
}
