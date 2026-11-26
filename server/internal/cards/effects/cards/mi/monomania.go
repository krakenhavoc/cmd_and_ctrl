package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Monomania — Sorcery {3}{B}{B} (Magic 2012, S82 OLD batch 452, #5573):
//
//	"Target player chooses a card in their hand and discards the rest."
//
// Choosing one card and discarding the rest is the target player discarding
// every card but one of their choice: Breakthrough's "discard all but X"
// prompt for the target, sized as the spell resolves. A hand of one or no
// cards discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29c6935e-ef64-4298-af74-9b18eed93136",
		Name:         "Monomania",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.FirstLegalTargetPlayer(ctx)
			p := ctx.Game.PlayerByIDForEffect(victim)
			if p == nil {
				return nil
			}
			if n := p.Hand.Size() - 1; n > 0 {
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: p.ID, Source: ctx.Source(), N: n})
			}
			return nil
		},
	})
}
