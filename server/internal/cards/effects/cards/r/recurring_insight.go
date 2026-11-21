package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Recurring Insight — Sorcery {4}{U}{U}:
//
//	"Draw cards equal to the number of cards in target opponent's hand.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The hand is counted as the spell resolves. Rebound is the engine's
// keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c6815bbb-24a5-4c9f-bf30-6190bc766d05",
		Name:            "Recurring Insight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			opp, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			p := ctx.PlayerByID(opp)
			if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
				return nil
			}
			return effects.DrawCards{Player: ctx.Controller(), N: p.Hand.Size()}.Apply(ctx)
		},
	})
}
