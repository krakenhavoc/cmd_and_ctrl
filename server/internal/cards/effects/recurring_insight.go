package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "c6815bbb-24a5-4c9f-bf30-6190bc766d05",
		Name:            "Recurring Insight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			opp, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			p := ctx.PlayerByID(opp)
			if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
				return nil
			}
			return DrawCards{Player: ctx.Controller(), N: p.Hand.Size()}.Apply(ctx)
		},
	})
}
